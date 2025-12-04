package smartthings

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	// SmartThings Hub URN patterns
	smartThingsURN = "urn:SmartThingsCommunity"
)

// ssdpMulticastAddr is the SSDP multicast address (variable to allow testing)
var ssdpMulticastAddr = "239.255.255.250:1900"

// DiscoveredHub represents a SmartThings Hub found via SSDP discovery.
type DiscoveredHub struct {
	// IP is the hub's local IP address.
	IP string

	// Port is the hub's local API port (typically 39500).
	Port int

	// Location is the SSDP location URL for the hub description.
	Location string

	// Server is the server header from the SSDP response.
	Server string

	// USN is the unique service name.
	USN string

	// ST is the search target that matched.
	ST string
}

// DiscoveredTV represents a Samsung TV found via mDNS or SSDP discovery.
type DiscoveredTV struct {
	// IP is the TV's local IP address.
	IP string

	// Port is the TV's control port.
	Port int

	// Name is the friendly name of the TV.
	Name string

	// Model is the TV model name.
	Model string

	// UUID is the unique device identifier.
	UUID string
}

// Discovery provides local network device discovery functionality.
type Discovery struct {
	// Timeout is the maximum time to wait for discovery responses.
	// Defaults to 3 seconds if zero.
	Timeout time.Duration
}

// NewDiscovery creates a new Discovery instance.
func NewDiscovery(timeout time.Duration) *Discovery {
	if timeout == 0 {
		timeout = 3 * time.Second
	}
	return &Discovery{Timeout: timeout}
}

// FindHubs discovers SmartThings Hubs on the local network using SSDP.
// Returns all hubs that respond within the timeout period.
func (d *Discovery) FindHubs(ctx context.Context) ([]DiscoveredHub, error) {
	return d.ssdpDiscover(ctx, "ssdp:all", func(headers map[string]string) *DiscoveredHub {
		// Look for SmartThings-related responses
		st := headers["ST"]
		usn := headers["USN"]
		server := headers["SERVER"]

		// Check if this looks like a SmartThings hub
		if !strings.Contains(st, "SmartThings") &&
			!strings.Contains(usn, "SmartThings") &&
			!strings.Contains(server, "SmartThings") &&
			!strings.Contains(st, smartThingsURN) {
			return nil
		}

		location := headers["LOCATION"]
		ip, port := parseLocationURL(location)
		if ip == "" {
			return nil
		}

		return &DiscoveredHub{
			IP:       ip,
			Port:     port,
			Location: location,
			Server:   server,
			USN:      usn,
			ST:       st,
		}
	})
}

// FindTVs discovers Samsung TVs on the local network using SSDP.
// Samsung TVs respond to SSDP with urn:samsung.com:device:RemoteControlReceiver:1.
func (d *Discovery) FindTVs(ctx context.Context) ([]DiscoveredTV, error) {
	tvs := []DiscoveredTV{}

	// Samsung TVs respond to multiple URNs
	searchTargets := []string{
		"urn:samsung.com:device:RemoteControlReceiver:1",
		"urn:dial-multiscreen-org:service:dial:1",
		"ssdp:all",
	}

	seen := make(map[string]bool)

	for _, st := range searchTargets {
		results, err := d.ssdpDiscoverTV(ctx, st, seen)
		if err != nil {
			continue // Try next search target
		}
		tvs = append(tvs, results...)
	}

	return tvs, nil
}

// ssdpDiscover performs SSDP M-SEARCH and returns matching hubs.
func (d *Discovery) ssdpDiscover(ctx context.Context, searchTarget string, matcher func(map[string]string) *DiscoveredHub) ([]DiscoveredHub, error) {
	conn, err := net.ListenPacket("udp4", ":0")
	if err != nil {
		return nil, fmt.Errorf("Discovery: listen: %w", err)
	}
	defer conn.Close()

	// Resolve multicast address
	addr, err := net.ResolveUDPAddr("udp4", ssdpMulticastAddr)
	if err != nil {
		return nil, fmt.Errorf("Discovery: resolve multicast: %w", err)
	}

	// Build M-SEARCH request
	request := fmt.Sprintf(
		"M-SEARCH * HTTP/1.1\r\n"+
			"HOST: %s\r\n"+
			"MAN: \"ssdp:discover\"\r\n"+
			"MX: %d\r\n"+
			"ST: %s\r\n"+
			"\r\n",
		ssdpMulticastAddr,
		int(d.Timeout.Seconds()),
		searchTarget,
	)

	// Send discovery request
	if _, err := conn.WriteTo([]byte(request), addr); err != nil {
		return nil, fmt.Errorf("Discovery: send: %w", err)
	}

	// Collect responses
	hubs := []DiscoveredHub{}
	seen := make(map[string]bool)

	deadline := time.Now().Add(d.Timeout)
	_ = conn.SetReadDeadline(deadline)

	buf := make([]byte, 2048)
	for {
		select {
		case <-ctx.Done():
			return hubs, ctx.Err()
		default:
		}

		n, remoteAddr, err := conn.ReadFrom(buf)
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				break // Normal timeout, we're done
			}
			continue // Other errors, keep trying
		}

		headers := parseSSDPResponse(string(buf[:n]))
		if headers == nil {
			continue
		}

		// Use remote address if location parsing fails
		if headers["_REMOTE_IP"] == "" {
			if udpAddr, ok := remoteAddr.(*net.UDPAddr); ok {
				headers["_REMOTE_IP"] = udpAddr.IP.String()
			}
		}

		hub := matcher(headers)
		if hub == nil {
			continue
		}

		// Deduplicate by IP
		if seen[hub.IP] {
			continue
		}
		seen[hub.IP] = true
		hubs = append(hubs, *hub)
	}

	return hubs, nil
}

// ssdpDiscoverTV performs SSDP M-SEARCH for Samsung TVs.
func (d *Discovery) ssdpDiscoverTV(ctx context.Context, searchTarget string, seen map[string]bool) ([]DiscoveredTV, error) {
	conn, err := net.ListenPacket("udp4", ":0")
	if err != nil {
		return nil, fmt.Errorf("Discovery: listen: %w", err)
	}
	defer conn.Close()

	addr, err := net.ResolveUDPAddr("udp4", ssdpMulticastAddr)
	if err != nil {
		return nil, fmt.Errorf("Discovery: resolve multicast: %w", err)
	}

	request := fmt.Sprintf(
		"M-SEARCH * HTTP/1.1\r\n"+
			"HOST: %s\r\n"+
			"MAN: \"ssdp:discover\"\r\n"+
			"MX: %d\r\n"+
			"ST: %s\r\n"+
			"\r\n",
		ssdpMulticastAddr,
		int(d.Timeout.Seconds()),
		searchTarget,
	)

	if _, err := conn.WriteTo([]byte(request), addr); err != nil {
		return nil, fmt.Errorf("Discovery: send: %w", err)
	}

	tvs := []DiscoveredTV{}
	deadline := time.Now().Add(d.Timeout)
	_ = conn.SetReadDeadline(deadline)

	buf := make([]byte, 2048)
	for {
		select {
		case <-ctx.Done():
			return tvs, ctx.Err()
		default:
		}

		n, remoteAddr, err := conn.ReadFrom(buf)
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				break
			}
			continue
		}

		headers := parseSSDPResponse(string(buf[:n]))
		if headers == nil {
			continue
		}

		// Check if this is a Samsung TV
		st := strings.ToLower(headers["ST"])
		server := strings.ToLower(headers["SERVER"])
		usn := headers["USN"]

		isSamsungTV := strings.Contains(st, "samsung") ||
			strings.Contains(server, "samsung") ||
			strings.Contains(strings.ToLower(usn), "samsung")

		if !isSamsungTV {
			continue
		}

		ip := ""
		port := 8001 // Default Samsung TV port
		location := headers["LOCATION"]

		if location != "" {
			ip, port = parseLocationURL(location)
		}
		if ip == "" {
			if udpAddr, ok := remoteAddr.(*net.UDPAddr); ok {
				ip = udpAddr.IP.String()
			}
		}

		if ip == "" || seen[ip] {
			continue
		}
		seen[ip] = true

		// Extract UUID from USN
		uuid := extractUUID(usn)

		// Extract model from server header
		model := extractModel(server)

		tvs = append(tvs, DiscoveredTV{
			IP:    ip,
			Port:  port,
			Name:  headers["FRIENDLY_NAME"],
			Model: model,
			UUID:  uuid,
		})
	}

	return tvs, nil
}

// parseSSDPResponse parses an SSDP response into a header map.
func parseSSDPResponse(response string) map[string]string {
	headers := make(map[string]string)
	scanner := bufio.NewScanner(strings.NewReader(response))

	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		// Skip the HTTP status line
		if strings.HasPrefix(line, "HTTP/") {
			continue
		}

		idx := strings.Index(line, ":")
		if idx == -1 {
			continue
		}

		key := strings.ToUpper(strings.TrimSpace(line[:idx]))
		value := strings.TrimSpace(line[idx+1:])
		headers[key] = value
	}

	return headers
}

// parseLocationURL extracts IP and port from a location URL.
func parseLocationURL(location string) (string, int) {
	// Pattern: http://IP:PORT/...
	re := regexp.MustCompile(`https?://([^:/]+):?(\d*)`)
	matches := re.FindStringSubmatch(location)
	if len(matches) < 2 {
		return "", 0
	}

	ip := matches[1]
	port := 80
	if len(matches) > 2 && matches[2] != "" {
		if p, err := strconv.Atoi(matches[2]); err == nil {
			port = p
		}
	}

	return ip, port
}

// extractUUID extracts a UUID from a USN string.
// Format: uuid:XXXXXXXX-XXXX-XXXX-XXXX-XXXXXXXXXXXX::urn:...
func extractUUID(usn string) string {
	re := regexp.MustCompile(`uuid:([a-fA-F0-9-]+)`)
	matches := re.FindStringSubmatch(usn)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}

// extractModel tries to extract a model name from the server header.
func extractModel(server string) string {
	// Common patterns: "Samsung/1.0" or "Model: UN65TU8000"
	parts := strings.Split(server, "/")
	if len(parts) > 0 && strings.Contains(strings.ToLower(parts[0]), "samsung") {
		return strings.TrimSpace(parts[0])
	}
	return ""
}

// DiscoverAll performs discovery for all supported device types.
// Returns hubs and TVs found on the local network.
func (d *Discovery) DiscoverAll(ctx context.Context) ([]DiscoveredHub, []DiscoveredTV, error) {
	var (
		wg         sync.WaitGroup
		hubs       []DiscoveredHub
		tvs        []DiscoveredTV
		hubsErr    error
		tvsErr     error
		mu         sync.Mutex
	)

	wg.Go(func() {
		result, err := d.FindHubs(ctx)
		mu.Lock()
		hubs = result
		hubsErr = err
		mu.Unlock()
	})

	wg.Go(func() {
		result, err := d.FindTVs(ctx)
		mu.Lock()
		tvs = result
		tvsErr = err
		mu.Unlock()
	})

	wg.Wait()

	// Return first error if any
	if hubsErr != nil {
		return hubs, tvs, hubsErr
	}
	return hubs, tvs, tvsErr
}
