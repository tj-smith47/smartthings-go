package smartthings

import (
	"bufio"
	"context"
	"crypto/rand"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

const (
	// HubLocalDefaultPort is the default port for SmartThings Hub local API.
	HubLocalDefaultPort = 39500

	// WebSocket magic GUID for Sec-WebSocket-Accept calculation.
	websocketMagicGUID = "258EAFA5-E914-47DA-95CA-C5AB0DC85B11"

	// WebSocket opcodes.
	wsOpcodeContinuation = 0x0
	wsOpcodeText         = 0x1
	wsOpcodeBinary       = 0x2
	wsOpcodeClose        = 0x8
	wsOpcodePing         = 0x9
	wsOpcodePong         = 0xA
)

// HubLocalEvent represents a real-time device event from the hub's local API.
type HubLocalEvent struct {
	// DeviceID is the UUID of the device that generated the event.
	DeviceID string `json:"deviceId"`

	// Component is the device component (e.g., "main").
	Component string `json:"component"`

	// Capability is the capability that changed (e.g., "switch", "temperatureMeasurement").
	Capability string `json:"capability"`

	// Attribute is the specific attribute (e.g., "switch", "temperature").
	Attribute string `json:"attribute"`

	// Value is the new value of the attribute.
	Value any `json:"value"`

	// Unit is the unit of measurement (if applicable).
	Unit string `json:"unit,omitempty"`

	// Timestamp is when the event occurred.
	Timestamp time.Time `json:"timestamp"`

	// StateChange indicates if this is a state change (true) or just a report (false).
	StateChange bool `json:"stateChange"`
}

// HubLocalMessage represents a message from the hub's local WebSocket API.
type HubLocalMessage struct {
	MessageType string          `json:"messageType"`
	DeviceEvent *HubLocalEvent  `json:"deviceEvent,omitempty"`
	Error       *HubLocalError  `json:"error,omitempty"`
	Raw         json.RawMessage `json:"-"`
}

// HubLocalError represents an error from the hub's local API.
type HubLocalError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// HubLocalClient provides real-time device events via WebSocket connection
// to a SmartThings Hub's local API.
type HubLocalClient struct {
	hubIP   string
	hubPort int
	token   string

	conn     net.Conn
	connMu   sync.Mutex
	reader   *bufio.Reader
	events   chan HubLocalEvent
	errors   chan error
	done     chan struct{}
	closeErr error

	// Ping/pong tracking
	lastPong time.Time
	pongMu   sync.RWMutex

	// Reconnect configuration
	reconnectEnabled  bool
	reconnectDelay    time.Duration
	reconnectMaxDelay time.Duration
	onReconnect       func()
	onDisconnect      func(error)

	// Subscription tracking for resubscription after reconnect
	subscriptions   []string // Device IDs subscribed to
	subscribeAll    bool     // Whether SubscribeAll was called
	subscriptionsMu sync.RWMutex

	// Reconnect state
	reconnecting   bool
	reconnectMu    sync.Mutex
	manualClose    bool // True if Close() was called explicitly
	eventBufferSz  int
}

// HubLocalConfig configures the HubLocalClient.
type HubLocalConfig struct {
	// HubIP is the local IP address of the SmartThings hub.
	HubIP string

	// HubPort is the local API port (default: 39500).
	HubPort int

	// Token is the local API bearer token from the hub.
	Token string

	// EventBufferSize is the size of the events channel buffer (default: 100).
	EventBufferSize int

	// ReconnectEnabled enables automatic reconnection on disconnect (default: true).
	ReconnectEnabled *bool

	// ReconnectDelay is the initial delay before reconnecting (default: 5s).
	ReconnectDelay time.Duration

	// ReconnectMaxDelay is the maximum delay between reconnection attempts (default: 5m).
	ReconnectMaxDelay time.Duration

	// OnReconnect is called when a reconnection occurs.
	OnReconnect func()

	// OnDisconnect is called when the connection is lost.
	OnDisconnect func(error)
}

// NewHubLocalClient creates a new client for connecting to a SmartThings Hub's local API.
// The token should be obtained from the hub's local API authentication.
func NewHubLocalClient(cfg *HubLocalConfig) (*HubLocalClient, error) {
	if cfg == nil {
		return nil, errors.New("HubLocalClient: config is required")
	}
	if cfg.HubIP == "" {
		return nil, errors.New("HubLocalClient: hub IP is required")
	}
	if cfg.Token == "" {
		return nil, errors.New("HubLocalClient: token is required")
	}

	port := cfg.HubPort
	if port == 0 {
		port = HubLocalDefaultPort
	}

	bufSize := cfg.EventBufferSize
	if bufSize == 0 {
		bufSize = 100
	}

	// Default reconnect enabled to true
	reconnectEnabled := true
	if cfg.ReconnectEnabled != nil {
		reconnectEnabled = *cfg.ReconnectEnabled
	}

	reconnectDelay := cfg.ReconnectDelay
	if reconnectDelay == 0 {
		reconnectDelay = 5 * time.Second
	}

	reconnectMaxDelay := cfg.ReconnectMaxDelay
	if reconnectMaxDelay == 0 {
		reconnectMaxDelay = 5 * time.Minute
	}

	return &HubLocalClient{
		hubIP:             cfg.HubIP,
		hubPort:           port,
		token:             cfg.Token,
		events:            make(chan HubLocalEvent, bufSize),
		errors:            make(chan error, 10),
		done:              make(chan struct{}),
		eventBufferSz:     bufSize,
		reconnectEnabled:  reconnectEnabled,
		reconnectDelay:    reconnectDelay,
		reconnectMaxDelay: reconnectMaxDelay,
		onReconnect:       cfg.OnReconnect,
		onDisconnect:      cfg.OnDisconnect,
	}, nil
}

// Connect establishes a WebSocket connection to the hub's local API.
// The connection will remain open until Close is called or an error occurs.
func (c *HubLocalClient) Connect(ctx context.Context) error {
	c.connMu.Lock()
	defer c.connMu.Unlock()

	if c.conn != nil {
		return errors.New("HubLocalClient: already connected")
	}

	// Build WebSocket URL
	wsURL := fmt.Sprintf("ws://%s:%d/events", c.hubIP, c.hubPort)

	// Perform WebSocket handshake
	conn, err := c.dialWebSocket(ctx, wsURL)
	if err != nil {
		return fmt.Errorf("HubLocalClient: connect failed: %w", err)
	}

	c.conn = conn
	c.reader = bufio.NewReader(conn)
	c.lastPong = time.Now()

	// Start read loop in background
	go c.readLoop()

	// Start ping loop in background
	go c.pingLoop()

	return nil
}

// dialWebSocket performs the WebSocket handshake using stdlib.
func (c *HubLocalClient) dialWebSocket(ctx context.Context, wsURL string) (net.Conn, error) {
	u, err := url.Parse(wsURL)
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}

	host := u.Host
	if !strings.Contains(host, ":") {
		if u.Scheme == "wss" {
			host += ":443"
		} else {
			host += ":80"
		}
	}

	// Create connection with context deadline
	var d net.Dialer
	conn, err := d.DialContext(ctx, "tcp", host)
	if err != nil {
		return nil, fmt.Errorf("dial failed: %w", err)
	}

	// Generate WebSocket key
	key := make([]byte, 16)
	if _, err := rand.Read(key); err != nil {
		conn.Close()
		return nil, fmt.Errorf("generate key failed: %w", err)
	}
	wsKey := base64.StdEncoding.EncodeToString(key)

	// Build upgrade request
	path := u.Path
	if path == "" {
		path = "/"
	}
	if u.RawQuery != "" {
		path += "?" + u.RawQuery
	}

	req := fmt.Sprintf(
		"GET %s HTTP/1.1\r\n"+
			"Host: %s\r\n"+
			"Upgrade: websocket\r\n"+
			"Connection: Upgrade\r\n"+
			"Sec-WebSocket-Key: %s\r\n"+
			"Sec-WebSocket-Version: 13\r\n"+
			"Authorization: Bearer %s\r\n"+
			"\r\n",
		path, u.Host, wsKey, c.token,
	)

	// Send upgrade request
	if _, err := conn.Write([]byte(req)); err != nil {
		conn.Close()
		return nil, fmt.Errorf("write request failed: %w", err)
	}

	// Read response
	reader := bufio.NewReader(conn)
	resp, err := http.ReadResponse(reader, nil)
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("read response failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusSwitchingProtocols {
		conn.Close()
		return nil, fmt.Errorf("unexpected status: %d %s", resp.StatusCode, resp.Status)
	}

	// Verify Sec-WebSocket-Accept
	expectedAccept := computeWebSocketAccept(wsKey)
	if resp.Header.Get("Sec-WebSocket-Accept") != expectedAccept {
		conn.Close()
		return nil, errors.New("invalid Sec-WebSocket-Accept header")
	}

	return conn, nil
}

// computeWebSocketAccept calculates the expected Sec-WebSocket-Accept value.
func computeWebSocketAccept(key string) string {
	h := sha1.New()
	h.Write([]byte(key + websocketMagicGUID))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

// readLoop continuously reads WebSocket frames from the connection.
func (c *HubLocalClient) readLoop() {
	var disconnectErr error

	defer func() {
		// Only close channels if this is a manual close or reconnect is disabled
		c.reconnectMu.Lock()
		shouldReconnect := c.reconnectEnabled && !c.manualClose
		c.reconnectMu.Unlock()

		// Call disconnect callback
		if c.onDisconnect != nil && disconnectErr != nil {
			c.onDisconnect(disconnectErr)
		}

		if shouldReconnect {
			// Trigger reconnect in background
			go c.reconnectLoop()
		} else {
			// Close channels only when not reconnecting
			close(c.events)
			close(c.errors)
		}
	}()

	for {
		select {
		case <-c.done:
			return
		default:
		}

		// Set read deadline
		c.connMu.Lock()
		if c.conn != nil {
			c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		}
		c.connMu.Unlock()

		// Read WebSocket frame
		opcode, payload, err := c.readFrame()
		if err != nil {
			select {
			case <-c.done:
				return
			default:
				if !errors.Is(err, io.EOF) && !errors.Is(err, net.ErrClosed) {
					disconnectErr = fmt.Errorf("read frame: %w", err)
					select {
					case c.errors <- disconnectErr:
					default:
					}
				} else {
					disconnectErr = err
				}
				return
			}
		}

		switch opcode {
		case wsOpcodeText:
			c.handleTextMessage(payload)
		case wsOpcodeBinary:
			c.handleTextMessage(payload) // Treat as JSON
		case wsOpcodePing:
			c.sendPong(payload)
		case wsOpcodePong:
			c.pongMu.Lock()
			c.lastPong = time.Now()
			c.pongMu.Unlock()
		case wsOpcodeClose:
			disconnectErr = errors.New("server closed connection")
			return
		}
	}
}

// readFrame reads a single WebSocket frame.
func (c *HubLocalClient) readFrame() (opcode byte, payload []byte, err error) {
	c.connMu.Lock()
	reader := c.reader
	c.connMu.Unlock()

	if reader == nil {
		return 0, nil, errors.New("not connected")
	}

	// Read first two bytes (FIN, opcode, mask flag, payload length)
	header := make([]byte, 2)
	if _, err := io.ReadFull(reader, header); err != nil {
		return 0, nil, err
	}

	// fin := header[0]&0x80 != 0
	opcode = header[0] & 0x0F
	masked := header[1]&0x80 != 0
	payloadLen := uint64(header[1] & 0x7F)

	// Extended payload length
	if payloadLen == 126 {
		ext := make([]byte, 2)
		if _, err := io.ReadFull(reader, ext); err != nil {
			return 0, nil, err
		}
		payloadLen = uint64(ext[0])<<8 | uint64(ext[1])
	} else if payloadLen == 127 {
		ext := make([]byte, 8)
		if _, err := io.ReadFull(reader, ext); err != nil {
			return 0, nil, err
		}
		payloadLen = uint64(ext[0])<<56 | uint64(ext[1])<<48 | uint64(ext[2])<<40 | uint64(ext[3])<<32 |
			uint64(ext[4])<<24 | uint64(ext[5])<<16 | uint64(ext[6])<<8 | uint64(ext[7])
	}

	// Read masking key if present (server frames should not be masked)
	var maskKey [4]byte
	if masked {
		if _, err := io.ReadFull(reader, maskKey[:]); err != nil {
			return 0, nil, err
		}
	}

	// Read payload
	payload = make([]byte, payloadLen)
	if _, err := io.ReadFull(reader, payload); err != nil {
		return 0, nil, err
	}

	// Unmask if needed
	if masked {
		for i := range payload {
			payload[i] ^= maskKey[i%4]
		}
	}

	return opcode, payload, nil
}

// handleTextMessage processes a text WebSocket message.
func (c *HubLocalClient) handleTextMessage(payload []byte) {
	var msg HubLocalMessage
	msg.Raw = payload

	if err := json.Unmarshal(payload, &msg); err != nil {
		c.errors <- fmt.Errorf("parse message: %w", err)
		return
	}

	if msg.DeviceEvent != nil {
		select {
		case c.events <- *msg.DeviceEvent:
		default:
			// Buffer full, drop event
		}
	}

	if msg.Error != nil {
		c.errors <- fmt.Errorf("hub error [%s]: %s", msg.Error.Code, msg.Error.Message)
	}
}

// sendFrame sends a WebSocket frame.
func (c *HubLocalClient) sendFrame(opcode byte, payload []byte) error {
	c.connMu.Lock()
	defer c.connMu.Unlock()

	if c.conn == nil {
		return errors.New("not connected")
	}

	// Client frames must be masked
	maskKey := make([]byte, 4)
	if _, err := rand.Read(maskKey); err != nil {
		return err
	}

	// Build frame
	var frame []byte
	frame = append(frame, 0x80|opcode) // FIN + opcode

	payloadLen := len(payload)
	if payloadLen < 126 {
		frame = append(frame, 0x80|byte(payloadLen)) // Mask + length
	} else if payloadLen < 65536 {
		frame = append(frame, 0x80|126)
		frame = append(frame, byte(payloadLen>>8), byte(payloadLen))
	} else {
		frame = append(frame, 0x80|127)
		frame = append(frame, 0, 0, 0, 0,
			byte(payloadLen>>24), byte(payloadLen>>16), byte(payloadLen>>8), byte(payloadLen))
	}

	// Add mask key
	frame = append(frame, maskKey...)

	// Mask and add payload
	maskedPayload := make([]byte, len(payload))
	for i := range payload {
		maskedPayload[i] = payload[i] ^ maskKey[i%4]
	}
	frame = append(frame, maskedPayload...)

	_, err := c.conn.Write(frame)
	return err
}

// sendPong sends a pong frame in response to a ping.
func (c *HubLocalClient) sendPong(payload []byte) {
	if err := c.sendFrame(wsOpcodePong, payload); err != nil {
		c.errors <- fmt.Errorf("send pong: %w", err)
	}
}

// pingLoop sends periodic ping frames to keep the connection alive.
func (c *HubLocalClient) pingLoop() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-c.done:
			return
		case <-ticker.C:
			// Check if we've received a pong recently
			c.pongMu.RLock()
			lastPong := c.lastPong
			c.pongMu.RUnlock()

			if time.Since(lastPong) > 90*time.Second {
				c.errors <- errors.New("connection timeout: no pong received")
				c.Close()
				return
			}

			// Send ping
			if err := c.sendFrame(wsOpcodePing, nil); err != nil {
				c.errors <- fmt.Errorf("send ping: %w", err)
				return
			}
		}
	}
}

// Events returns a channel that receives device events.
// The channel is closed when the connection is closed.
func (c *HubLocalClient) Events() <-chan HubLocalEvent {
	return c.events
}

// Errors returns a channel that receives connection errors.
// The channel is closed when the connection is closed.
func (c *HubLocalClient) Errors() <-chan error {
	return c.errors
}

// Close closes the WebSocket connection and prevents automatic reconnection.
func (c *HubLocalClient) Close() error {
	// Mark as manual close to prevent reconnection
	c.reconnectMu.Lock()
	c.manualClose = true
	c.reconnectMu.Unlock()

	c.connMu.Lock()

	if c.conn == nil {
		c.connMu.Unlock()
		return nil
	}

	// Signal done to stop goroutines
	select {
	case <-c.done:
		// Already closed
	default:
		close(c.done)
	}

	// Send close frame (inline to avoid deadlock with sendFrame)
	closeFrame := []byte{0x80 | wsOpcodeClose, 0x80, 0, 0, 0, 0} // Empty masked close frame
	_, _ = c.conn.Write(closeFrame)

	// Close connection
	err := c.conn.Close()
	c.conn = nil
	c.reader = nil
	c.closeErr = err
	c.connMu.Unlock()

	return err
}

// IsConnected returns true if the client is currently connected.
func (c *HubLocalClient) IsConnected() bool {
	c.connMu.Lock()
	defer c.connMu.Unlock()
	return c.conn != nil
}

// Subscribe sends a subscription request for device events.
// Call this after Connect to start receiving events for specific devices.
// Subscriptions are tracked and automatically restored after reconnection.
func (c *HubLocalClient) Subscribe(ctx context.Context, deviceIDs ...string) error {
	if len(deviceIDs) == 0 {
		return errors.New("at least one device ID is required")
	}

	// Track subscriptions for reconnect
	c.subscriptionsMu.Lock()
	for _, id := range deviceIDs {
		// Add if not already present
		found := false
		for _, existing := range c.subscriptions {
			if existing == id {
				found = true
				break
			}
		}
		if !found {
			c.subscriptions = append(c.subscriptions, id)
		}
	}
	c.subscriptionsMu.Unlock()

	msg := map[string]any{
		"messageType": "subscribe",
		"deviceIds":   deviceIDs,
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("marshal subscribe: %w", err)
	}

	return c.sendFrame(wsOpcodeText, data)
}

// Unsubscribe removes subscription for device events.
func (c *HubLocalClient) Unsubscribe(ctx context.Context, deviceIDs ...string) error {
	if len(deviceIDs) == 0 {
		return errors.New("at least one device ID is required")
	}

	msg := map[string]any{
		"messageType": "unsubscribe",
		"deviceIds":   deviceIDs,
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("marshal unsubscribe: %w", err)
	}

	return c.sendFrame(wsOpcodeText, data)
}

// SubscribeAll subscribes to events from all devices on the hub.
// This subscription is tracked and automatically restored after reconnection.
func (c *HubLocalClient) SubscribeAll(ctx context.Context) error {
	// Track that we want all subscriptions
	c.subscriptionsMu.Lock()
	c.subscribeAll = true
	c.subscriptionsMu.Unlock()

	msg := map[string]any{
		"messageType": "subscribeAll",
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("marshal subscribeAll: %w", err)
	}

	return c.sendFrame(wsOpcodeText, data)
}

// reconnectLoop handles automatic reconnection with exponential backoff.
func (c *HubLocalClient) reconnectLoop() {
	c.reconnectMu.Lock()
	if c.reconnecting {
		c.reconnectMu.Unlock()
		return
	}
	c.reconnecting = true
	c.reconnectMu.Unlock()

	defer func() {
		c.reconnectMu.Lock()
		c.reconnecting = false
		c.reconnectMu.Unlock()
	}()

	delay := c.reconnectDelay
	ctx := context.Background()

	for {
		// Check if manually closed
		c.reconnectMu.Lock()
		if c.manualClose {
			c.reconnectMu.Unlock()
			return
		}
		c.reconnectMu.Unlock()

		// Wait before attempting reconnect
		time.Sleep(delay)

		// Check again after sleep
		c.reconnectMu.Lock()
		if c.manualClose {
			c.reconnectMu.Unlock()
			return
		}
		c.reconnectMu.Unlock()

		// Reset connection state for reconnect
		c.connMu.Lock()
		c.conn = nil
		c.reader = nil
		c.done = make(chan struct{})
		c.connMu.Unlock()

		// Attempt to connect
		wsURL := fmt.Sprintf("ws://%s:%d/events", c.hubIP, c.hubPort)
		conn, err := c.dialWebSocket(ctx, wsURL)
		if err != nil {
			// Increase delay with exponential backoff
			delay = time.Duration(float64(delay) * 1.5)
			if delay > c.reconnectMaxDelay {
				delay = c.reconnectMaxDelay
			}

			select {
			case c.errors <- fmt.Errorf("reconnect failed: %w", err):
			default:
			}
			continue
		}

		// Connection successful
		c.connMu.Lock()
		c.conn = conn
		c.reader = bufio.NewReader(conn)
		c.lastPong = time.Now()
		c.connMu.Unlock()

		// Start read and ping loops
		go c.readLoop()
		go c.pingLoop()

		// Restore subscriptions
		c.subscriptionsMu.RLock()
		subscribeAll := c.subscribeAll
		subscriptions := make([]string, len(c.subscriptions))
		copy(subscriptions, c.subscriptions)
		c.subscriptionsMu.RUnlock()

		if subscribeAll {
			if err := c.SubscribeAll(ctx); err != nil {
				select {
				case c.errors <- fmt.Errorf("resubscribe all failed: %w", err):
				default:
				}
			}
		} else if len(subscriptions) > 0 {
			if err := c.Subscribe(ctx, subscriptions...); err != nil {
				select {
				case c.errors <- fmt.Errorf("resubscribe failed: %w", err):
				default:
				}
			}
		}

		// Call reconnect callback
		if c.onReconnect != nil {
			c.onReconnect()
		}

		return
	}
}

// SetReconnectEnabled enables or disables automatic reconnection.
func (c *HubLocalClient) SetReconnectEnabled(enabled bool) {
	c.reconnectMu.Lock()
	c.reconnectEnabled = enabled
	c.reconnectMu.Unlock()
}

// IsReconnecting returns true if the client is currently attempting to reconnect.
func (c *HubLocalClient) IsReconnecting() bool {
	c.reconnectMu.Lock()
	defer c.reconnectMu.Unlock()
	return c.reconnecting
}
