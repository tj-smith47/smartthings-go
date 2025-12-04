package smartthings

import (
	"bufio"
	"context"
	"crypto/rand"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestNewHubLocalClient(t *testing.T) {
	tests := []struct {
		name    string
		cfg     *HubLocalConfig
		wantErr bool
		errMsg  string
	}{
		{
			name:    "nil config",
			cfg:     nil,
			wantErr: true,
			errMsg:  "config is required",
		},
		{
			name:    "empty hub IP",
			cfg:     &HubLocalConfig{Token: "test-token"},
			wantErr: true,
			errMsg:  "hub IP is required",
		},
		{
			name:    "empty token",
			cfg:     &HubLocalConfig{HubIP: "192.168.1.100"},
			wantErr: true,
			errMsg:  "token is required",
		},
		{
			name: "valid config with defaults",
			cfg: &HubLocalConfig{
				HubIP: "192.168.1.100",
				Token: "test-token",
			},
			wantErr: false,
		},
		{
			name: "valid config with custom port",
			cfg: &HubLocalConfig{
				HubIP:   "192.168.1.100",
				HubPort: 8080,
				Token:   "test-token",
			},
			wantErr: false,
		},
		{
			name: "valid config with buffer size",
			cfg: &HubLocalConfig{
				HubIP:           "192.168.1.100",
				Token:           "test-token",
				EventBufferSize: 500,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewHubLocalClient(tt.cfg)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("error = %q, want containing %q", err.Error(), tt.errMsg)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if client == nil {
				t.Fatal("expected client, got nil")
			}
		})
	}
}

func TestHubLocalClient_DefaultPort(t *testing.T) {
	client, err := NewHubLocalClient(&HubLocalConfig{
		HubIP: "192.168.1.100",
		Token: "test-token",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if client.hubPort != HubLocalDefaultPort {
		t.Errorf("hubPort = %d, want %d", client.hubPort, HubLocalDefaultPort)
	}
}

func TestHubLocalClient_CustomPort(t *testing.T) {
	client, err := NewHubLocalClient(&HubLocalConfig{
		HubIP:   "192.168.1.100",
		HubPort: 8080,
		Token:   "test-token",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if client.hubPort != 8080 {
		t.Errorf("hubPort = %d, want 8080", client.hubPort)
	}
}

func TestComputeWebSocketAccept(t *testing.T) {
	// Test vector from RFC 6455
	key := "dGhlIHNhbXBsZSBub25jZQ=="
	expected := "s3pPLMBiTxaQ9kYGzzhZRbK+xOo="

	result := computeWebSocketAccept(key)
	if result != expected {
		t.Errorf("computeWebSocketAccept(%q) = %q, want %q", key, result, expected)
	}
}

// mockWebSocketServer creates a mock WebSocket server for testing.
type mockWebSocketServer struct {
	listener net.Listener
	conns    []net.Conn
	mu       sync.Mutex
	t        *testing.T

	// Handlers
	onConnect    func(conn net.Conn, reader *bufio.Reader)
	onMessage    func(conn net.Conn, opcode byte, payload []byte)
	acceptUpgrade bool
}

func newMockWebSocketServer(t *testing.T) *mockWebSocketServer {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to create listener: %v", err)
	}

	s := &mockWebSocketServer{
		listener:      listener,
		t:             t,
		acceptUpgrade: true,
	}

	go s.serve()
	return s
}

func (s *mockWebSocketServer) serve() {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			return
		}

		s.mu.Lock()
		s.conns = append(s.conns, conn)
		s.mu.Unlock()

		go s.handleConn(conn)
	}
}

func (s *mockWebSocketServer) handleConn(conn net.Conn) {
	reader := bufio.NewReader(conn)

	// Read HTTP upgrade request
	req, err := http.ReadRequest(reader)
	if err != nil {
		return
	}

	// Verify upgrade headers
	if req.Header.Get("Upgrade") != "websocket" {
		http.Error(nil, "not a websocket upgrade", http.StatusBadRequest)
		conn.Close()
		return
	}

	wsKey := req.Header.Get("Sec-WebSocket-Key")
	if wsKey == "" {
		conn.Close()
		return
	}

	if !s.acceptUpgrade {
		conn.Write([]byte("HTTP/1.1 403 Forbidden\r\n\r\n"))
		conn.Close()
		return
	}

	// Send upgrade response
	accept := computeWebSocketAccept(wsKey)
	resp := fmt.Sprintf(
		"HTTP/1.1 101 Switching Protocols\r\n"+
			"Upgrade: websocket\r\n"+
			"Connection: Upgrade\r\n"+
			"Sec-WebSocket-Accept: %s\r\n"+
			"\r\n",
		accept,
	)
	conn.Write([]byte(resp))

	if s.onConnect != nil {
		s.onConnect(conn, reader)
	}

	// Read messages
	for {
		opcode, payload, err := s.readClientFrame(reader)
		if err != nil {
			return
		}

		if s.onMessage != nil {
			s.onMessage(conn, opcode, payload)
		}

		if opcode == wsOpcodeClose {
			return
		}
	}
}

func (s *mockWebSocketServer) readClientFrame(reader *bufio.Reader) (byte, []byte, error) {
	header := make([]byte, 2)
	if _, err := reader.Read(header); err != nil {
		return 0, nil, err
	}

	opcode := header[0] & 0x0F
	masked := header[1]&0x80 != 0
	payloadLen := int(header[1] & 0x7F)

	if payloadLen == 126 {
		ext := make([]byte, 2)
		if _, err := reader.Read(ext); err != nil {
			return 0, nil, err
		}
		payloadLen = int(ext[0])<<8 | int(ext[1])
	} else if payloadLen == 127 {
		ext := make([]byte, 8)
		if _, err := reader.Read(ext); err != nil {
			return 0, nil, err
		}
		// Simplified - assume payload fits in int
		payloadLen = int(ext[6])<<8 | int(ext[7])
	}

	var maskKey [4]byte
	if masked {
		if _, err := reader.Read(maskKey[:]); err != nil {
			return 0, nil, err
		}
	}

	payload := make([]byte, payloadLen)
	if _, err := reader.Read(payload); err != nil {
		return 0, nil, err
	}

	if masked {
		for i := range payload {
			payload[i] ^= maskKey[i%4]
		}
	}

	return opcode, payload, nil
}

func (s *mockWebSocketServer) sendFrame(conn net.Conn, opcode byte, payload []byte) error {
	var frame []byte
	frame = append(frame, 0x80|opcode) // FIN + opcode

	payloadLen := len(payload)
	if payloadLen < 126 {
		frame = append(frame, byte(payloadLen)) // No mask for server
	} else if payloadLen < 65536 {
		frame = append(frame, 126)
		frame = append(frame, byte(payloadLen>>8), byte(payloadLen))
	} else {
		frame = append(frame, 127)
		frame = append(frame, 0, 0, 0, 0,
			byte(payloadLen>>24), byte(payloadLen>>16), byte(payloadLen>>8), byte(payloadLen))
	}

	frame = append(frame, payload...)
	_, err := conn.Write(frame)
	return err
}

func (s *mockWebSocketServer) addr() string {
	return s.listener.Addr().String()
}

func (s *mockWebSocketServer) close() {
	s.listener.Close()
	s.mu.Lock()
	for _, conn := range s.conns {
		conn.Close()
	}
	s.mu.Unlock()
}

func TestHubLocalClient_Connect(t *testing.T) {
	server := newMockWebSocketServer(t)
	defer server.close()

	parts := strings.Split(server.addr(), ":")
	ip := parts[0]
	port := 0
	fmt.Sscanf(parts[1], "%d", &port)

	client, err := NewHubLocalClient(&HubLocalConfig{
		HubIP:   ip,
		HubPort: port,
		Token:   "test-token",
	})
	if err != nil {
		t.Fatalf("NewHubLocalClient: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Connect(ctx); err != nil {
		t.Fatalf("Connect: %v", err)
	}
	defer client.Close()

	if !client.IsConnected() {
		t.Error("expected IsConnected() = true")
	}
}

func TestHubLocalClient_ConnectAlreadyConnected(t *testing.T) {
	server := newMockWebSocketServer(t)
	defer server.close()

	parts := strings.Split(server.addr(), ":")
	ip := parts[0]
	port := 0
	fmt.Sscanf(parts[1], "%d", &port)

	client, _ := NewHubLocalClient(&HubLocalConfig{
		HubIP:   ip,
		HubPort: port,
		Token:   "test-token",
	})

	ctx := context.Background()
	_ = client.Connect(ctx)
	defer client.Close()

	err := client.Connect(ctx)
	if err == nil {
		t.Fatal("expected error for double connect")
	}
	if !strings.Contains(err.Error(), "already connected") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestHubLocalClient_ReceiveEvent(t *testing.T) {
	var serverConn net.Conn
	connChan := make(chan struct{})

	server := newMockWebSocketServer(t)
	server.onConnect = func(conn net.Conn, reader *bufio.Reader) {
		serverConn = conn
		close(connChan)
	}
	defer server.close()

	parts := strings.Split(server.addr(), ":")
	ip := parts[0]
	port := 0
	fmt.Sscanf(parts[1], "%d", &port)

	client, _ := NewHubLocalClient(&HubLocalConfig{
		HubIP:   ip,
		HubPort: port,
		Token:   "test-token",
	})

	ctx := context.Background()
	if err := client.Connect(ctx); err != nil {
		t.Fatalf("Connect: %v", err)
	}
	defer client.Close()

	// Wait for connection
	select {
	case <-connChan:
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for connection")
	}

	// Send a device event from server
	event := map[string]any{
		"messageType": "deviceEvent",
		"deviceEvent": map[string]any{
			"deviceId":    "device-123",
			"component":   "main",
			"capability":  "switch",
			"attribute":   "switch",
			"value":       "on",
			"stateChange": true,
		},
	}
	eventJSON, _ := json.Marshal(event)
	server.sendFrame(serverConn, wsOpcodeText, eventJSON)

	// Wait for event
	select {
	case ev := <-client.Events():
		if ev.DeviceID != "device-123" {
			t.Errorf("DeviceID = %q, want %q", ev.DeviceID, "device-123")
		}
		if ev.Capability != "switch" {
			t.Errorf("Capability = %q, want %q", ev.Capability, "switch")
		}
		if ev.Value != "on" {
			t.Errorf("Value = %v, want %q", ev.Value, "on")
		}
		if !ev.StateChange {
			t.Error("expected StateChange = true")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for event")
	}
}

func TestHubLocalClient_Subscribe(t *testing.T) {
	var receivedMsg []byte
	msgChan := make(chan struct{})

	server := newMockWebSocketServer(t)
	server.onMessage = func(conn net.Conn, opcode byte, payload []byte) {
		if opcode == wsOpcodeText {
			receivedMsg = payload
			close(msgChan)
		}
	}
	defer server.close()

	parts := strings.Split(server.addr(), ":")
	ip := parts[0]
	port := 0
	fmt.Sscanf(parts[1], "%d", &port)

	client, _ := NewHubLocalClient(&HubLocalConfig{
		HubIP:   ip,
		HubPort: port,
		Token:   "test-token",
	})

	ctx := context.Background()
	_ = client.Connect(ctx)
	defer client.Close()

	// Give connection time to establish
	time.Sleep(100 * time.Millisecond)

	if err := client.Subscribe(ctx, "device-1", "device-2"); err != nil {
		t.Fatalf("Subscribe: %v", err)
	}

	select {
	case <-msgChan:
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for subscribe message")
	}

	var msg map[string]any
	if err := json.Unmarshal(receivedMsg, &msg); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if msg["messageType"] != "subscribe" {
		t.Errorf("messageType = %v, want %q", msg["messageType"], "subscribe")
	}

	deviceIDs, ok := msg["deviceIds"].([]any)
	if !ok || len(deviceIDs) != 2 {
		t.Errorf("deviceIds = %v, want 2 elements", msg["deviceIds"])
	}
}

func TestHubLocalClient_SubscribeEmpty(t *testing.T) {
	client, _ := NewHubLocalClient(&HubLocalConfig{
		HubIP: "192.168.1.1",
		Token: "test-token",
	})

	err := client.Subscribe(context.Background())
	if err == nil {
		t.Fatal("expected error for empty subscribe")
	}
	if !strings.Contains(err.Error(), "at least one device ID") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestHubLocalClient_Unsubscribe(t *testing.T) {
	var receivedMsg []byte
	msgChan := make(chan struct{})

	server := newMockWebSocketServer(t)
	server.onMessage = func(conn net.Conn, opcode byte, payload []byte) {
		if opcode == wsOpcodeText {
			receivedMsg = payload
			close(msgChan)
		}
	}
	defer server.close()

	parts := strings.Split(server.addr(), ":")
	ip := parts[0]
	port := 0
	fmt.Sscanf(parts[1], "%d", &port)

	client, _ := NewHubLocalClient(&HubLocalConfig{
		HubIP:   ip,
		HubPort: port,
		Token:   "test-token",
	})

	ctx := context.Background()
	_ = client.Connect(ctx)
	defer client.Close()

	time.Sleep(100 * time.Millisecond)

	if err := client.Unsubscribe(ctx, "device-1"); err != nil {
		t.Fatalf("Unsubscribe: %v", err)
	}

	select {
	case <-msgChan:
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for unsubscribe message")
	}

	var msg map[string]any
	json.Unmarshal(receivedMsg, &msg)

	if msg["messageType"] != "unsubscribe" {
		t.Errorf("messageType = %v, want %q", msg["messageType"], "unsubscribe")
	}
}

func TestHubLocalClient_SubscribeAll(t *testing.T) {
	var receivedMsg []byte
	msgChan := make(chan struct{})

	server := newMockWebSocketServer(t)
	server.onMessage = func(conn net.Conn, opcode byte, payload []byte) {
		if opcode == wsOpcodeText {
			receivedMsg = payload
			close(msgChan)
		}
	}
	defer server.close()

	parts := strings.Split(server.addr(), ":")
	ip := parts[0]
	port := 0
	fmt.Sscanf(parts[1], "%d", &port)

	client, _ := NewHubLocalClient(&HubLocalConfig{
		HubIP:   ip,
		HubPort: port,
		Token:   "test-token",
	})

	ctx := context.Background()
	_ = client.Connect(ctx)
	defer client.Close()

	time.Sleep(100 * time.Millisecond)

	if err := client.SubscribeAll(ctx); err != nil {
		t.Fatalf("SubscribeAll: %v", err)
	}

	select {
	case <-msgChan:
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for subscribeAll message")
	}

	var msg map[string]any
	json.Unmarshal(receivedMsg, &msg)

	if msg["messageType"] != "subscribeAll" {
		t.Errorf("messageType = %v, want %q", msg["messageType"], "subscribeAll")
	}
}

func TestHubLocalClient_Close(t *testing.T) {
	server := newMockWebSocketServer(t)
	defer server.close()

	parts := strings.Split(server.addr(), ":")
	ip := parts[0]
	port := 0
	fmt.Sscanf(parts[1], "%d", &port)

	client, _ := NewHubLocalClient(&HubLocalConfig{
		HubIP:   ip,
		HubPort: port,
		Token:   "test-token",
	})

	ctx := context.Background()
	_ = client.Connect(ctx)

	if err := client.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	if client.IsConnected() {
		t.Error("expected IsConnected() = false after close")
	}

	// Double close should be safe
	if err := client.Close(); err != nil {
		t.Errorf("double Close: %v", err)
	}
}

func TestHubLocalClient_PingPong(t *testing.T) {
	pongReceived := make(chan struct{})

	server := newMockWebSocketServer(t)
	server.onConnect = func(conn net.Conn, reader *bufio.Reader) {
		// Small delay to allow client's readLoop goroutine to start
		// This prevents a race where ping is sent before client is listening
		time.Sleep(50 * time.Millisecond)
		// Send ping to client
		server.sendFrame(conn, wsOpcodePing, []byte("ping"))
	}
	server.onMessage = func(conn net.Conn, opcode byte, payload []byte) {
		if opcode == wsOpcodePong {
			close(pongReceived)
		}
	}
	defer server.close()

	parts := strings.Split(server.addr(), ":")
	ip := parts[0]
	port := 0
	fmt.Sscanf(parts[1], "%d", &port)

	client, _ := NewHubLocalClient(&HubLocalConfig{
		HubIP:   ip,
		HubPort: port,
		Token:   "test-token",
	})

	ctx := context.Background()
	_ = client.Connect(ctx)
	defer client.Close()

	select {
	case <-pongReceived:
		// Success
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for pong")
	}
}

func TestHubLocalClient_ConnectFailure(t *testing.T) {
	// Try to connect to a non-existent server
	client, _ := NewHubLocalClient(&HubLocalConfig{
		HubIP:   "127.0.0.1",
		HubPort: 59999, // Unlikely to be in use
		Token:   "test-token",
	})

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	err := client.Connect(ctx)
	if err == nil {
		client.Close()
		t.Fatal("expected error for connection to non-existent server")
	}
}

func TestHubLocalClient_UpgradeRejected(t *testing.T) {
	server := newMockWebSocketServer(t)
	server.acceptUpgrade = false
	defer server.close()

	parts := strings.Split(server.addr(), ":")
	ip := parts[0]
	port := 0
	fmt.Sscanf(parts[1], "%d", &port)

	client, _ := NewHubLocalClient(&HubLocalConfig{
		HubIP:   ip,
		HubPort: port,
		Token:   "test-token",
	})

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err := client.Connect(ctx)
	if err == nil {
		client.Close()
		t.Fatal("expected error for rejected upgrade")
	}
	if !strings.Contains(err.Error(), "403") && !strings.Contains(err.Error(), "unexpected status") {
		t.Logf("error: %v", err)
	}
}

func TestHubLocalEvent_JSON(t *testing.T) {
	jsonData := `{
		"deviceId": "abc-123",
		"component": "main",
		"capability": "temperatureMeasurement",
		"attribute": "temperature",
		"value": 72.5,
		"unit": "F",
		"timestamp": "2024-01-15T10:30:00Z",
		"stateChange": true
	}`

	var event HubLocalEvent
	if err := json.Unmarshal([]byte(jsonData), &event); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if event.DeviceID != "abc-123" {
		t.Errorf("DeviceID = %q, want %q", event.DeviceID, "abc-123")
	}
	if event.Capability != "temperatureMeasurement" {
		t.Errorf("Capability = %q, want %q", event.Capability, "temperatureMeasurement")
	}
	if event.Value != 72.5 {
		t.Errorf("Value = %v, want 72.5", event.Value)
	}
	if event.Unit != "F" {
		t.Errorf("Unit = %q, want %q", event.Unit, "F")
	}
}

func TestHubLocalMessage_JSON(t *testing.T) {
	jsonData := `{
		"messageType": "deviceEvent",
		"deviceEvent": {
			"deviceId": "device-123",
			"component": "main",
			"capability": "switch",
			"attribute": "switch",
			"value": "on"
		}
	}`

	var msg HubLocalMessage
	if err := json.Unmarshal([]byte(jsonData), &msg); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if msg.MessageType != "deviceEvent" {
		t.Errorf("MessageType = %q, want %q", msg.MessageType, "deviceEvent")
	}
	if msg.DeviceEvent == nil {
		t.Fatal("DeviceEvent is nil")
	}
	if msg.DeviceEvent.DeviceID != "device-123" {
		t.Errorf("DeviceEvent.DeviceID = %q, want %q", msg.DeviceEvent.DeviceID, "device-123")
	}
}

func TestHubLocalError_JSON(t *testing.T) {
	jsonData := `{
		"messageType": "error",
		"error": {
			"code": "UNAUTHORIZED",
			"message": "Invalid token"
		}
	}`

	var msg HubLocalMessage
	if err := json.Unmarshal([]byte(jsonData), &msg); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if msg.Error == nil {
		t.Fatal("Error is nil")
	}
	if msg.Error.Code != "UNAUTHORIZED" {
		t.Errorf("Error.Code = %q, want %q", msg.Error.Code, "UNAUTHORIZED")
	}
	if msg.Error.Message != "Invalid token" {
		t.Errorf("Error.Message = %q, want %q", msg.Error.Message, "Invalid token")
	}
}

// Helper to generate random WebSocket key
func generateWebSocketKey() string {
	key := make([]byte, 16)
	rand.Read(key)
	return base64.StdEncoding.EncodeToString(key)
}

// Helper to compute SHA1 hash for WebSocket accept
func computeAcceptKey(key string) string {
	h := sha1.New()
	h.Write([]byte(key + "258EAFA5-E914-47DA-95CA-C5AB0DC85B11"))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}
