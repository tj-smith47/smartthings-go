package smartthings

import (
	"bufio"
	"context"
	"crypto/rand"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"sync"
	"time"
)

const (
	// TVLocalDefaultPort is the default port for Samsung TV local API (non-secure).
	TVLocalDefaultPort = 8001

	// TVLocalSecurePort is the secure port for Samsung TV local API.
	TVLocalSecurePort = 8002

	// TVLocalAPIPath is the WebSocket path for Samsung TV remote control.
	TVLocalAPIPath = "/api/v2/channels/samsung.remote.control"
)

// TVLocalConfig configures the TVLocalClient.
type TVLocalConfig struct {
	// TVIP is the local IP address of the Samsung TV.
	TVIP string

	// Port is the TV API port (default: 8002 for secure, 8001 for non-secure).
	Port int

	// Token is the authentication token (required for secure connections on port 8002).
	// This is obtained on first connection when the TV prompts the user.
	Token string

	// AppName is the name shown on the TV when connecting (default: "SmartThings-Go").
	AppName string

	// Secure enables TLS for the WebSocket connection (default: true if port is 8002).
	Secure *bool

	// OnTokenReceived is called when the TV provides an authentication token.
	// This allows persisting the token for future connections.
	OnTokenReceived func(token string)
}

// TVLocalClient provides direct control of Samsung TVs via their local WebSocket API.
// This is separate from the SmartThings API and provides lower latency control.
type TVLocalClient struct {
	tvIP    string
	port    int
	token   string
	appName string
	secure  bool

	conn         net.Conn
	connMu       sync.Mutex
	reader       *bufio.Reader
	done         chan struct{}
	responses    chan tvLocalResponse
	errors       chan error
	onTokenRecvd func(token string)

	// Ping/pong tracking
	lastPong time.Time
	pongMu   sync.RWMutex
}

// tvLocalResponse represents a response from the TV's WebSocket API.
type tvLocalResponse struct {
	Event string         `json:"event"`
	Data  map[string]any `json:"data,omitempty"`
}

// TVDeviceInfo contains information about a Samsung TV.
type TVDeviceInfo struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	Model          string `json:"model"`
	ModelName      string `json:"modelName"`
	Type           string `json:"type"`
	NetworkType    string `json:"networkType"`
	SSID           string `json:"ssid,omitempty"`
	WifiMac        string `json:"wifiMac,omitempty"`
	IP             string `json:"ip"`
	Device         string `json:"device"`
	TokenAuth      string `json:"tokenAuthSupport,omitempty"`
	PowerState     string `json:"PowerState,omitempty"`
	FrameTVSupport string `json:"FrameTVSupport,omitempty"`
	OS             string `json:"OS,omitempty"`
}

// NewTVLocalClient creates a new client for connecting to a Samsung TV's local API.
func NewTVLocalClient(cfg *TVLocalConfig) (*TVLocalClient, error) {
	if cfg == nil {
		return nil, errors.New("TVLocalClient: config is required")
	}
	if cfg.TVIP == "" {
		return nil, errors.New("TVLocalClient: TV IP is required")
	}

	port := cfg.Port
	if port == 0 {
		port = TVLocalSecurePort
	}

	appName := cfg.AppName
	if appName == "" {
		appName = "SmartThings-Go"
	}

	// Default to secure if port is 8002
	secure := port == TVLocalSecurePort
	if cfg.Secure != nil {
		secure = *cfg.Secure
	}

	return &TVLocalClient{
		tvIP:         cfg.TVIP,
		port:         port,
		token:        cfg.Token,
		appName:      appName,
		secure:       secure,
		done:         make(chan struct{}),
		responses:    make(chan tvLocalResponse, 10),
		errors:       make(chan error, 10),
		onTokenRecvd: cfg.OnTokenReceived,
	}, nil
}

// Connect establishes a WebSocket connection to the TV's local API.
func (c *TVLocalClient) Connect(ctx context.Context) error {
	c.connMu.Lock()
	defer c.connMu.Unlock()

	if c.conn != nil {
		return errors.New("TVLocalClient: already connected")
	}

	// Build WebSocket URL
	scheme := "ws"
	if c.secure {
		scheme = "wss"
	}

	// Encode app name for URL
	encodedName := base64.StdEncoding.EncodeToString([]byte(c.appName))

	wsURL := fmt.Sprintf("%s://%s:%d%s?name=%s",
		scheme, c.tvIP, c.port, TVLocalAPIPath, url.QueryEscape(encodedName))

	if c.token != "" {
		wsURL += "&token=" + url.QueryEscape(c.token)
	}

	// Perform WebSocket handshake
	conn, err := c.dialWebSocket(ctx, wsURL)
	if err != nil {
		return fmt.Errorf("TVLocalClient: connect failed: %w", err)
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

// dialWebSocket performs the WebSocket handshake.
func (c *TVLocalClient) dialWebSocket(ctx context.Context, wsURL string) (net.Conn, error) {
	u, err := url.Parse(wsURL)
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}

	host := u.Host
	if !hasPort(host) {
		if u.Scheme == "wss" {
			host += ":443"
		} else {
			host += ":80"
		}
	}

	// Create connection with context deadline
	var conn net.Conn
	var d net.Dialer

	if c.secure {
		// Use TLS for secure connections
		tlsConfig := &tls.Config{
			InsecureSkipVerify: true, // Samsung TVs use self-signed certs
		}
		conn, err = tls.DialWithDialer(&d, "tcp", host, tlsConfig)
	} else {
		conn, err = d.DialContext(ctx, "tcp", host)
	}
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
			"\r\n",
		path, u.Host, wsKey,
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

// hasPort checks if a host string includes a port.
func hasPort(host string) bool {
	for i := len(host) - 1; i >= 0; i-- {
		if host[i] == ':' {
			return true
		}
		if host[i] < '0' || host[i] > '9' {
			return false
		}
	}
	return false
}

// readLoop continuously reads WebSocket frames from the connection.
func (c *TVLocalClient) readLoop() {
	defer func() {
		close(c.responses)
		close(c.errors)
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
					c.errors <- fmt.Errorf("read frame: %w", err)
				}
				return
			}
		}

		switch opcode {
		case wsOpcodeText:
			c.handleTextMessage(payload)
		case wsOpcodeBinary:
			c.handleTextMessage(payload)
		case wsOpcodePing:
			c.sendPong(payload)
		case wsOpcodePong:
			c.pongMu.Lock()
			c.lastPong = time.Now()
			c.pongMu.Unlock()
		case wsOpcodeClose:
			return
		}
	}
}

// readFrame reads a single WebSocket frame.
func (c *TVLocalClient) readFrame() (opcode byte, payload []byte, err error) {
	c.connMu.Lock()
	reader := c.reader
	c.connMu.Unlock()

	if reader == nil {
		return 0, nil, errors.New("not connected")
	}

	// Read first two bytes
	header := make([]byte, 2)
	if _, err := io.ReadFull(reader, header); err != nil {
		return 0, nil, err
	}

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

	// Read masking key if present
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
func (c *TVLocalClient) handleTextMessage(payload []byte) {
	var resp tvLocalResponse
	if err := json.Unmarshal(payload, &resp); err != nil {
		c.errors <- fmt.Errorf("parse message: %w", err)
		return
	}

	// Check for token in response
	if resp.Event == "ms.channel.connect" {
		if data := resp.Data; data != nil {
			if token, ok := data["token"].(string); ok && token != "" && c.onTokenRecvd != nil {
				c.onTokenRecvd(token)
			}
		}
	}

	select {
	case c.responses <- resp:
	default:
		// Buffer full, drop response
	}
}

// sendFrame sends a WebSocket frame.
func (c *TVLocalClient) sendFrame(opcode byte, payload []byte) error {
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
		frame = append(frame, 0x80|byte(payloadLen))
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
func (c *TVLocalClient) sendPong(payload []byte) {
	if err := c.sendFrame(wsOpcodePong, payload); err != nil {
		c.errors <- fmt.Errorf("send pong: %w", err)
	}
}

// pingLoop sends periodic ping frames to keep the connection alive.
func (c *TVLocalClient) pingLoop() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-c.done:
			return
		case <-ticker.C:
			c.pongMu.RLock()
			lastPong := c.lastPong
			c.pongMu.RUnlock()

			if time.Since(lastPong) > 90*time.Second {
				c.errors <- errors.New("connection timeout: no pong received")
				c.Close()
				return
			}

			if err := c.sendFrame(wsOpcodePing, nil); err != nil {
				c.errors <- fmt.Errorf("send ping: %w", err)
				return
			}
		}
	}
}

// Responses returns a channel that receives TV responses.
func (c *TVLocalClient) Responses() <-chan tvLocalResponse {
	return c.responses
}

// Errors returns a channel that receives connection errors.
func (c *TVLocalClient) Errors() <-chan error {
	return c.errors
}

// Close closes the WebSocket connection.
func (c *TVLocalClient) Close() error {
	c.connMu.Lock()
	defer c.connMu.Unlock()

	if c.conn == nil {
		return nil
	}

	// Signal done
	select {
	case <-c.done:
	default:
		close(c.done)
	}

	// Send close frame
	closeFrame := []byte{0x80 | wsOpcodeClose, 0x80, 0, 0, 0, 0}
	_, _ = c.conn.Write(closeFrame)

	err := c.conn.Close()
	c.conn = nil
	c.reader = nil

	return err
}

// IsConnected returns true if the client is connected.
func (c *TVLocalClient) IsConnected() bool {
	c.connMu.Lock()
	defer c.connMu.Unlock()
	return c.conn != nil
}

// SendKey sends a remote control key press to the TV.
// Common keys: KEY_POWER, KEY_HOME, KEY_MENU, KEY_UP, KEY_DOWN, KEY_LEFT, KEY_RIGHT,
// KEY_ENTER, KEY_RETURN, KEY_EXIT, KEY_VOLUP, KEY_VOLDOWN, KEY_MUTE,
// KEY_CHUP, KEY_CHDOWN, KEY_0 through KEY_9, KEY_PLAY, KEY_PAUSE, KEY_STOP
func (c *TVLocalClient) SendKey(key string) error {
	msg := map[string]any{
		"method": "ms.remote.control",
		"params": map[string]any{
			"Cmd":          "Click",
			"DataOfCmd":    key,
			"Option":       "false",
			"TypeOfRemote": "SendRemoteKey",
		},
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("marshal key command: %w", err)
	}

	return c.sendFrame(wsOpcodeText, data)
}

// SendText sends text input to the TV.
// This is used when the TV has a text input field active (e.g., search box).
func (c *TVLocalClient) SendText(text string) error {
	// Encode text as base64
	encoded := base64.StdEncoding.EncodeToString([]byte(text))

	msg := map[string]any{
		"method": "ms.remote.control",
		"params": map[string]any{
			"Cmd":          encoded,
			"TypeOfRemote": "SendInputString",
			"DataOfCmd":    "base64",
		},
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("marshal text command: %w", err)
	}

	return c.sendFrame(wsOpcodeText, data)
}

// LaunchApp launches an app on the TV.
func (c *TVLocalClient) LaunchApp(appID string, actionType string, metaTag string) error {
	if actionType == "" {
		actionType = "DEEP_LINK"
	}

	msg := map[string]any{
		"method": "ms.channel.emit",
		"params": map[string]any{
			"event": "ed.apps.launch",
			"to":    "host",
			"data": map[string]any{
				"appId":       appID,
				"action_type": actionType,
				"metaTag":     metaTag,
			},
		},
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("marshal app launch: %w", err)
	}

	return c.sendFrame(wsOpcodeText, data)
}

// OpenBrowser opens a URL in the TV's browser.
func (c *TVLocalClient) OpenBrowser(urlStr string) error {
	return c.LaunchApp("org.tizen.browser", "NATIVE_LAUNCH", urlStr)
}

// GetDeviceInfo fetches TV device information via HTTP REST API.
func (c *TVLocalClient) GetDeviceInfo(ctx context.Context) (*TVDeviceInfo, error) {
	scheme := "http"
	if c.secure {
		scheme = "https"
	}

	url := fmt.Sprintf("%s://%s:%d/api/v2/", scheme, c.tvIP, c.port)

	client := &http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read body: %w", err)
	}

	// Response wraps device info in a "device" field
	var wrapper struct {
		Device TVDeviceInfo `json:"device"`
	}
	if err := json.Unmarshal(body, &wrapper); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}

	return &wrapper.Device, nil
}

// Common TV key codes
const (
	TVKeyPower     = "KEY_POWER"
	TVKeyPowerOff  = "KEY_POWEROFF"
	TVKeyHome      = "KEY_HOME"
	TVKeyMenu      = "KEY_MENU"
	TVKeyUp        = "KEY_UP"
	TVKeyDown      = "KEY_DOWN"
	TVKeyLeft      = "KEY_LEFT"
	TVKeyRight     = "KEY_RIGHT"
	TVKeyEnter     = "KEY_ENTER"
	TVKeyReturn    = "KEY_RETURN"
	TVKeyBack      = "KEY_RETURN"
	TVKeyExit      = "KEY_EXIT"
	TVKeyVolumeUp  = "KEY_VOLUP"
	TVKeyVolumeDn  = "KEY_VOLDOWN"
	TVKeyMute      = "KEY_MUTE"
	TVKeyChannelUp = "KEY_CHUP"
	TVKeyChannelDn = "KEY_CHDOWN"
	TVKeyPlay      = "KEY_PLAY"
	TVKeyPause     = "KEY_PAUSE"
	TVKeyStop      = "KEY_STOP"
	TVKeyRewind    = "KEY_REWIND"
	TVKeyFF        = "KEY_FF"
	TVKeySource    = "KEY_SOURCE"
	TVKeyInfo      = "KEY_INFO"
	TVKeyGuide     = "KEY_GUIDE"
	TVKeyTools     = "KEY_TOOLS"
	TVKey0         = "KEY_0"
	TVKey1         = "KEY_1"
	TVKey2         = "KEY_2"
	TVKey3         = "KEY_3"
	TVKey4         = "KEY_4"
	TVKey5         = "KEY_5"
	TVKey6         = "KEY_6"
	TVKey7         = "KEY_7"
	TVKey8         = "KEY_8"
	TVKey9         = "KEY_9"
)
