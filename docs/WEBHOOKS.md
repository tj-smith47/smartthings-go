# SmartThings Webhooks Integration

This document outlines webhook-based real-time event handling as an alternative to polling.

## Current Implementation: Polling

**How it works:**
- Scheduled task runs every 15 seconds (`/db/appdata/api/tasks/`)
- Fetches all device states from SmartThings API
- Updates PostgreSQL database with latest states
- Gome UI polls database every 15 seconds

**Pros:**
- Simple implementation
- No webhook registration needed
- Works with PAT authentication
- Reliable (no missed events if webhook endpoint is down)

**Cons:**
- 15-second latency for state changes
- Unnecessary API calls when no changes occur
- Higher bandwidth usage
- **240 API requests/hour** (15s polling × 4 = 240/hour)

---

## Webhooks Overview

SmartThings supports webhooks for real-time device event notifications via **SmartApps**.

### How It Works
```
Device State Change
  ↓
SmartThings Cloud detects change
  ↓
HTTP POST to your webhook endpoint
  ↓
Process event + update database
  ↓
Push to UI via WebSocket/SSE (optional)
```

### Latency Comparison
- **Polling**: 0-15 seconds (average 7.5s)
- **Webhooks**: <1 second (near real-time)

---

## Implementation: SmartApp Webhooks

### Requirements
1. **HTTPS endpoint** with valid SSL certificate (✅ you have `api.jarvispro.io`)
2. **Public accessibility** (check if cloudflare-tunnel allows inbound webhooks)
3. **SmartApp** registered at https://smartthings.developer.samsung.com/
4. **Webhook handler** to process POST requests

### Step 1: Register SmartApp

**Create Automation:**
1. Go to https://smartthings.developer.samsung.com/workspace
2. Click "Create New Automation"
3. Select "Automation for the SmartThings App"
4. Choose "Webhook" as app type

**Configuration:**
```json
{
  "appName": "JarvisPro Device Sync",
  "displayName": "JarvisPro",
  "description": "Real-time device state synchronization",
  "appType": "WEBHOOK_SMART_APP",
  "webhookSmartApp": {
    "targetUrl": "https://api.jarvispro.io/webhooks/smartthings",
    "targetStatus": "READY"
  },
  "permissions": [
    "r:devices:*",
    "x:devices:*",
    "r:locations:*"
  ]
}
```

### Step 2: Implement Webhook Handler

**Add to `/db/appdata/api/handlers/webhooks.go`:**
```go
package handlers

import (
    "encoding/json"
    "net/http"
    "github.com/gin-gonic/gin"
    "github.com/rs/zerolog/log"
)

type SmartThingsWebhook struct {
    Lifecycle string                 `json:"lifecycle"`
    EventData map[string]interface{} `json:"eventData"`
}

type DeviceEvent struct {
    DeviceID    string      `json:"deviceId"`
    ComponentID string      `json:"componentId"`
    Capability  string      `json:"capability"`
    Attribute   string      `json:"attribute"`
    Value       interface{} `json:"value"`
    StateChange bool        `json:"stateChange"`
}

func HandleSmartThingsWebhook(c *gin.Context) {
    var webhook SmartThingsWebhook
    if err := c.ShouldBindJSON(&webhook); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    switch webhook.Lifecycle {
    case "PING":
        // Health check from SmartThings
        c.JSON(200, gin.H{"challenge": webhook.EventData["challenge"]})

    case "CONFIGURATION":
        // Initial setup - subscribe to all device events
        handleConfiguration(c, webhook.EventData)

    case "EVENT":
        // Process device state change
        handleDeviceEvent(c, webhook.EventData)

    case "UNINSTALL":
        // Cleanup subscriptions
        log.Info().Msg("SmartApp uninstalled")
        c.JSON(200, gin.H{"status": "ok"})

    default:
        log.Warn().Str("lifecycle", webhook.Lifecycle).Msg("Unknown lifecycle event")
        c.JSON(200, gin.H{"status": "ok"})
    }
}

func handleDeviceEvent(c *gin.Context, eventData map[string]interface{}) {
    events, ok := eventData["events"].([]interface{})
    if !ok {
        c.JSON(400, gin.H{"error": "invalid event data"})
        return
    }

    for _, e := range events {
        eventMap := e.(map[string]interface{})
        deviceID := eventMap["deviceId"].(string)
        capability := eventMap["capability"].(string)
        attribute := eventMap["attribute"].(string)
        value := eventMap["value"]

        log.Info().
            Str("deviceId", deviceID).
            Str("capability", capability).
            Str("attribute", attribute).
            Interface("value", value).
            Msg("Device event received")

        // Update device state in database
        UpdateDeviceState(deviceID, capability, attribute, value)
    }

    c.JSON(200, gin.H{"status": "ok"})
}

func handleConfiguration(c *gin.Context, configData map[string]interface{}) {
    installedApp := configData["installedApp"].(map[string]interface{})
    installedAppID := installedApp["installedAppId"].(string)

    // Subscribe to all device events
    SubscribeToAllDevices(installedAppID)

    c.JSON(200, gin.H{
        "configurationData": map[string]interface{}{
            "initialize": map[string]interface{}{
                "name":        "JarvisPro Device Sync",
                "description": "Real-time device state synchronization",
                "permissions": []string{"r:devices:*", "x:devices:*"},
                "firstPageId": "1",
            },
        },
    })
}
```

**Add route in `main.go`:**
```go
api.POST("/webhooks/smartthings", handlers.HandleSmartThingsWebhook)
```

### Step 3: Subscribe to Device Events

```go
func SubscribeToAllDevices(installedAppID string) error {
    client := smartthings.NewClient(os.Getenv("SMARTTHINGS_TOKEN"))
    devices, err := client.ListDevices()
    if err != nil {
        return err
    }

    for _, device := range devices {
        // Subscribe to all capabilities for each device
        for _, cap := range device.Capabilities {
            err := SubscribeToCapability(installedAppID, device.DeviceID, cap.ID)
            if err != nil {
                log.Error().Err(err).
                    Str("deviceId", device.DeviceID).
                    Str("capability", cap.ID).
                    Msg("Failed to subscribe")
            }
        }
    }

    return nil
}

func SubscribeToCapability(appID, deviceID, capabilityID string) error {
    // Use SmartThings API to create subscription
    // POST /installedapps/{installedAppId}/subscriptions
    // Body: {"sourceType": "DEVICE", "device": {"deviceId": "...", "componentId": "main", "capability": "..."}}
    return nil
}
```

---

## Hybrid Approach: Webhooks + Polling Fallback (Recommended)

**Best of both worlds:**
1. Use webhooks for instant updates (priority)
2. Keep polling at **5-minute interval** as fallback/safety net
3. If webhook fails or is delayed, polling ensures eventual consistency

### Benefits
- **Instant updates** when webhooks work (< 1 second latency)
- **Resilience** to webhook failures or downtime
- **Lower API usage**: Polling reduced from 240/hour to 12/hour
- **Better UX**: Real-time updates for switches, instant feedback

### Configuration
```go
// Webhook handler - immediate update
func handleDeviceEvent(...) {
    UpdateDeviceState(...)
    BroadcastToClients(...)  // Optional: WebSocket push to Gome UI
}

// Polling task - now 5 minutes instead of 15 seconds
scheduler.Every(5).Minutes().Do(func() {
    RefreshAllDevices()
})
```

---

## Infrastructure Requirements

### 1. HTTPS Endpoint
✅ **Already have:** `https://api.jarvispro.io` with Let's Encrypt cert (via Traefik)

### 2. Public Accessibility
⚠️ **Check if cloudflare-tunnel allows inbound webhooks**

If using Cloudflare Tunnel:
- May need to whitelist SmartThings IP ranges
- Or create separate public endpoint for webhooks only

**Test accessibility:**
```bash
curl -X POST https://api.jarvispro.io/webhooks/smartthings \
  -H "Content-Type: application/json" \
  -d '{"lifecycle":"PING","eventData":{"challenge":"test"}}'
```

### 3. Webhook Signature Verification
SmartThings signs webhook payloads with HMAC-SHA256 for security.

**Add verification:**
```go
func verifySignature(req *http.Request, signature string) bool {
    // Get public key from SmartThings
    // Verify HMAC-SHA256 signature
    // Return true if valid
    return true  // TODO: Implement
}
```

---

## Implementation Effort

| Component | Time |
|-----------|------|
| SmartApp registration | 1 hour |
| Webhook handler implementation | 3 hours |
| Event processing & DB updates | 2 hours |
| Subscription logic | 1 hour |
| Testing & debugging | 2 hours |
| **Total** | **~9 hours** |

**Hybrid approach:**
- Add above + reduce polling: **+1 hour**
- **Total: ~10 hours**

---

## Recommendation

**For your use case** (~10 devices, homelab):
1. **Short term**: Keep polling at 15s - works fine, simple
2. **Medium term**: Investigate if rate limiting is causing "daily token rotation" issue
3. **If you need instant updates** (e.g., light switches feel sluggish):
   - Implement webhooks for critical devices only
   - Keep polling for everything else
4. **Long term**: Hybrid approach (webhooks + 5min polling) for production-grade reliability

**Priority**: **Low** - Polling works well for current scale. Implement webhooks only if:
- Need < 1s latency for UI updates
- Scaling to 50+ devices (polling becomes expensive)
- Building public-facing app with multiple users

---

## Alternative: Faster Polling

**Instead of webhooks**, you could:
- Reduce polling to **5 seconds** for near-real-time feel
- Use **conditional requests** (ETag/If-Modified-Since headers) to reduce bandwidth
- Still uses 720 requests/hour (within rate limits)

**Pros:**
- Much simpler than webhooks (no changes needed)
- Lower latency (5s average instead of 7.5s)
- No infrastructure changes

**Cons:**
- More API calls (720/hour vs 240/hour)
- Still not truly real-time
- Slightly higher bandwidth

---

## Resources
- [SmartThings SmartApp Webhooks](https://smartthings.developer.samsung.com/docs/smartapps/webhooks.html)
- [SmartThings Subscriptions API](https://smartthings.developer.samsung.com/docs/api-ref/st-api.html#tag/Subscriptions)
- [Webhook Security Best Practices](https://docs.github.com/en/webhooks/using-webhooks/best-practices-for-using-webhooks)
