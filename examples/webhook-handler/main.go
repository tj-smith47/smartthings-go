// Example webhook handler demonstrating SmartThings webhook processing
//
// This example shows how to:
// - Receive and validate webhook requests from SmartThings
// - Handle CONFIRMATION lifecycle for webhook registration
// - Process device events
// - Execute device commands in response to events
//
// Usage:
//
//	export SMARTTHINGS_WEBHOOK_SECRET=your-webhook-secret
//	export SMARTTHINGS_TOKEN=your-api-token  # For device commands
//	go run main.go
//
// Then configure your SmartApp to send webhooks to this server's /webhook endpoint.
// For local development, use a tunnel like ngrok: ngrok http 8080
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	st "github.com/tj-smith47/smartthings-go"
)

var (
	webhookSecret string
	apiClient     *st.Client
)

func main() {
	// Load configuration
	webhookSecret = os.Getenv("SMARTTHINGS_WEBHOOK_SECRET")
	if webhookSecret == "" {
		log.Println("Warning: SMARTTHINGS_WEBHOOK_SECRET not set, signature validation disabled")
	}

	// Optional: Create API client for device commands
	apiToken := os.Getenv("SMARTTHINGS_TOKEN")
	if apiToken != "" {
		var err error
		apiClient, err = st.NewClient(apiToken)
		if err != nil {
			log.Fatalf("Failed to create API client: %v", err)
		}
		log.Println("API client configured for device commands")
	}

	// Set up routes
	http.HandleFunc("/", handleHome)
	http.HandleFunc("/webhook", handleWebhook)
	http.HandleFunc("/health", handleHealth)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Starting webhook handler on http://localhost:%s", port)
	log.Printf("Webhook endpoint: http://localhost:%s/webhook", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func handleHome(w http.ResponseWriter, r *http.Request) {
	html := `<!DOCTYPE html>
<html>
<head><title>SmartThings Webhook Handler</title></head>
<body>
<h1>SmartThings Webhook Handler</h1>
<p>This server handles SmartThings webhook events.</p>
<h2>Endpoints</h2>
<ul>
  <li><code>POST /webhook</code> - Webhook receiver</li>
  <li><code>GET /health</code> - Health check</li>
</ul>
<h2>Configuration</h2>
<ul>
  <li>SMARTTHINGS_WEBHOOK_SECRET: ` + (map[bool]string{true: "Set", false: "Not set"})[webhookSecret != ""] + `</li>
  <li>SMARTTHINGS_TOKEN: ` + (map[bool]string{true: "Set", false: "Not set"})[apiClient != nil] + `</li>
</ul>
</body>
</html>`
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprint(w, html)
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "healthy",
		"time":   time.Now().Format(time.RFC3339),
	})
}

func handleWebhook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse and validate the webhook
	event, err := st.ParseWebhookRequest(r, webhookSecret)
	if err != nil {
		log.Printf("Webhook parse error: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	log.Printf("Received webhook: lifecycle=%s", event.Lifecycle)

	// Handle different lifecycle types
	switch event.Lifecycle {
	case st.LifecycleConfirmation:
		handleConfirmation(w, event)
	case st.LifecycleEvent:
		handleEvent(w, event)
	case st.LifecyclePing:
		handlePing(w, event)
	default:
		log.Printf("Unknown lifecycle: %s", event.Lifecycle)
		w.WriteHeader(http.StatusOK)
	}
}

func handleConfirmation(w http.ResponseWriter, event *st.WebhookEvent) {
	// SmartThings sends a confirmation URL that must be fetched to complete registration
	if event.ConfirmationData == nil || event.ConfirmationData.ConfirmationURL == "" {
		log.Println("Missing confirmation URL")
		http.Error(w, "Missing confirmation URL", http.StatusBadRequest)
		return
	}

	log.Printf("Confirming webhook registration: %s", event.ConfirmationData.ConfirmationURL)

	// Fetch the confirmation URL
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, event.ConfirmationData.ConfirmationURL, nil)
	if err != nil {
		log.Printf("Failed to create confirmation request: %v", err)
		http.Error(w, "Confirmation failed", http.StatusInternalServerError)
		return
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("Failed to confirm webhook: %v", err)
		http.Error(w, "Confirmation failed", http.StatusInternalServerError)
		return
	}
	resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("Confirmation returned status %d", resp.StatusCode)
		http.Error(w, "Confirmation failed", http.StatusInternalServerError)
		return
	}

	log.Println("Webhook registration confirmed")
	w.WriteHeader(http.StatusOK)
}

func handleEvent(w http.ResponseWriter, event *st.WebhookEvent) {
	if event.EventData == nil {
		w.WriteHeader(http.StatusOK)
		return
	}

	for _, deviceEvent := range event.EventData.Events {
		log.Printf("Device event: type=%s device=%s capability=%s attribute=%s value=%v",
			deviceEvent.EventType,
			deviceEvent.DeviceEvent.DeviceID,
			deviceEvent.DeviceEvent.Capability,
			deviceEvent.DeviceEvent.Attribute,
			deviceEvent.DeviceEvent.Value,
		)

		// Example: React to motion sensor events
		if deviceEvent.DeviceEvent.Capability == "motionSensor" &&
			deviceEvent.DeviceEvent.Attribute == "motion" &&
			deviceEvent.DeviceEvent.Value == "active" {
			handleMotionDetected(deviceEvent.DeviceEvent.DeviceID)
		}

		// Example: React to door/window sensor events
		if deviceEvent.DeviceEvent.Capability == "contactSensor" &&
			deviceEvent.DeviceEvent.Attribute == "contact" &&
			deviceEvent.DeviceEvent.Value == "open" {
			handleDoorOpened(deviceEvent.DeviceEvent.DeviceID)
		}
	}

	w.WriteHeader(http.StatusOK)
}

func handlePing(w http.ResponseWriter, event *st.WebhookEvent) {
	// Respond to ping with the challenge
	response := map[string]any{
		"pingData": event.PingData,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func handleMotionDetected(deviceID string) {
	log.Printf("Motion detected on device %s", deviceID)

	// Example: Turn on a light when motion is detected
	if apiClient != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// Replace with your actual light device ID
		lightDeviceID := os.Getenv("LIGHT_DEVICE_ID")
		if lightDeviceID != "" {
			err := apiClient.ExecuteCommands(ctx, lightDeviceID, []st.Command{
				{
					Component:  "main",
					Capability: "switch",
					Command:    "on",
				},
			})
			if err != nil {
				log.Printf("Failed to turn on light: %v", err)
			} else {
				log.Printf("Turned on light %s", lightDeviceID)
			}
		}
	}
}

func handleDoorOpened(deviceID string) {
	log.Printf("Door opened on device %s", deviceID)

	// Example: Send a notification, log the event, etc.
	// You could integrate with notification services here
}
