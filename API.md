# API Documentation

Base URL: `http://localhost:4010`

All responses follow a consistent JSON envelope format.

## Response Format

### Success

```json
{
  "success": true,
  "data": { ... },
  "message": "Description of the result",
  "meta": {
    "timestamp": "2026-02-17T10:30:00Z",
    "requestId": "550e8400-e29b-41d4-a716-446655440000"
  }
}
```

### Error

```json
{
  "success": false,
  "error": {
    "code": "ERROR_CODE",
    "message": "Human-readable error message"
  },
  "meta": {
    "timestamp": "2026-02-17T10:30:00Z",
    "requestId": "550e8400-e29b-41d4-a716-446655440000"
  }
}
```

### Error Codes

| Code | HTTP Status | Description |
|---|---|---|
| `UNAUTHORIZED` | 401 | Missing or invalid API key |
| `RATE_LIMITED` | 429 | Too many requests |
| `INVALID_REQUEST` | 400 | Malformed request body |
| `MISSING_TOKEN` | 400 | Token (phone number) not provided |
| `INVALID_METHOD` | 400 | Connection method must be `qr` or `code` |
| `INVALID_PHONE` | 400 | Phone number failed validation |
| `INVALID_MESSAGE` | 400 | Message text is empty |
| `DEVICE_NOT_FOUND` | 404 | No session for the given token |
| `DEVICE_NOT_CONNECTED` | 500 | Device exists but is not connected |
| `CONNECTION_FAILED` | 500 | Failed to establish WhatsApp connection |
| `SEND_FAILED` | 500 | Failed to send message |
| `VALIDATION_FAILED` | 500 | Phone validation request failed |
| `INTERNAL_ERROR` | 500 | Unexpected server error |

---

## Endpoints

### Health Check

#### `GET /health`

Basic health check. No authentication required.

**Response:**

```json
{
  "success": true,
  "message": "ok",
  "meta": { "timestamp": "...", "requestId": "..." }
}
```

---

#### `GET /health/detailed`

Detailed health check with memory usage, goroutine count, and connected devices.

**Headers:**

```
Authorization: Bearer YOUR_API_KEY
```

**Response:**

```json
{
  "success": true,
  "data": {
    "uptime": "2h30m15s",
    "goroutines": 12,
    "memoryAlloc": 15728640,
    "memorySys": 25165824,
    "deviceCount": 2,
    "devices": [
      { "token": "60123456789", "status": "connected" },
      { "token": "60198765432", "status": "disconnected" }
    ]
  },
  "message": "ok",
  "meta": { "timestamp": "...", "requestId": "..." }
}
```

---

### Devices

#### `POST /devices`

Connect a WhatsApp device. The QR code or pairing code is delivered via WebSocket, not in the HTTP response.

**Headers:**

```
Authorization: Bearer YOUR_API_KEY
Content-Type: application/json
```

**Request Body:**

```json
{
  "token": "60123456789",
  "method": "qr"
}
```

| Field | Type | Required | Description |
|---|---|---|---|
| `token` | string | Yes | Phone number (used as device identifier) |
| `method` | string | No | `qr` (default) or `code` |

**Response:**

```json
{
  "success": true,
  "data": {
    "token": "60123456789",
    "method": "qr"
  },
  "message": "QR code sent via WebSocket",
  "meta": { "timestamp": "...", "requestId": "..." }
}
```

**Connection Flow:**

1. Call `POST /devices` with the phone number and method
2. Listen on the WebSocket (`/ws`) for events matching your token
3. For `qr` method: receive `qrcode` events containing a QR string — render as QR image
4. For `code` method: receive `pairing-code` event containing the code — display to user
5. Once the user scans/enters the code, receive `connection-success` event

---

#### `DELETE /devices/:token`

Disconnect a device and logout from WhatsApp. The device will need to be re-paired to connect again. The session database is deleted.

**Headers:**

```
Authorization: Bearer YOUR_API_KEY
```

**Response:**

```json
{
  "success": true,
  "data": {
    "token": "60123456789"
  },
  "message": "Device disconnected and logged out",
  "meta": { "timestamp": "...", "requestId": "..." }
}
```

---

### Messages

#### `POST /messages`

Send a text message through a connected device. Includes a simulated typing delay (configurable via `TYPING_DELAY_MS`).

**Headers:**

```
Authorization: Bearer YOUR_API_KEY
Content-Type: application/json
```

**Request Body:**

```json
{
  "token": "60123456789",
  "to": "60198765432",
  "text": "Hello from wa-gateway-go!"
}
```

| Field | Type | Required | Description |
|---|---|---|---|
| `token` | string | Yes | Sender device token |
| `to` | string | Yes | Recipient phone number |
| `text` | string | Yes | Message text |

The `to` field is validated against the configured phone rules. Leading `0` is automatically replaced with the country code (e.g., `0123456789` becomes `60123456789`).

**Response:**

```json
{
  "success": true,
  "data": {
    "messageId": "3EB0ABC123456789",
    "timestamp": "2026-02-17T10:30:00Z"
  },
  "message": "Message sent",
  "meta": { "timestamp": "...", "requestId": "..." }
}
```

---

### Phone Validation

#### `POST /validate/phone`

Check if a phone number is registered on WhatsApp. Results are cached for the configured TTL.

**Headers:**

```
Authorization: Bearer YOUR_API_KEY
Content-Type: application/json
```

**Request Body:**

```json
{
  "token": "60123456789",
  "phone": "60198765432"
}
```

| Field | Type | Required | Description |
|---|---|---|---|
| `token` | string | Yes | Device token to use for the check |
| `phone` | string | Yes | Phone number to validate |

**Response:**

```json
{
  "success": true,
  "data": {
    "phone": "60198765432",
    "isOnWhatsApp": true,
    "jid": "60198765432@s.whatsapp.net",
    "cached": false
  },
  "message": "Phone validation result",
  "meta": { "timestamp": "...", "requestId": "..." }
}
```

---

### Contacts

#### `GET /contacts/:token`

List contacts captured from incoming messages and history sync for a specific device.

**Headers:**

```
Authorization: Bearer YOUR_API_KEY
```

**Query Parameters:**

| Parameter | Default | Description |
|---|---|---|
| `limit` | `100` | Number of contacts to return (max 1000) |
| `offset` | `0` | Pagination offset |
| `format` | `json` | Response format: `json` or `csv` |

**Response (JSON):**

```json
{
  "success": true,
  "data": {
    "contacts": [
      {
        "phone": "60198765432",
        "name": "Ali",
        "source": "message",
        "firstSeen": "2026-02-17T08:00:00Z",
        "lastSeen": "2026-02-17T10:30:00Z"
      }
    ],
    "total": 42,
    "limit": 100,
    "offset": 0
  },
  "message": "Contacts retrieved",
  "meta": { "timestamp": "...", "requestId": "..." }
}
```

**Response (CSV):**

When `?format=csv`, returns a downloadable CSV file:

```csv
phone,name,source,first_seen,last_seen
60198765432,Ali,message,2026-02-17T08:00:00Z,2026-02-17T10:30:00Z
```

---

### Cache

#### `DELETE /cache`

Clear the phone validation cache.

**Headers:**

```
Authorization: Bearer YOUR_API_KEY
```

**Response:**

```json
{
  "success": true,
  "data": {
    "cleared": 156
  },
  "message": "Cache cleared",
  "meta": { "timestamp": "...", "requestId": "..." }
}
```

---

## WebSocket

### `GET /ws`

Upgrade to WebSocket connection. No authentication required (browsers cannot set custom headers on WebSocket connections).

**URL:** `ws://localhost:4010/ws`

### Events

All messages are JSON objects with this structure:

```json
{
  "event": "event_name",
  "token": "60123456789",
  "data": ...,
  "message": "optional message"
}
```

#### `qrcode`

Sent when a QR code is generated for device pairing. The `data` field contains the raw QR code string. Render it as a QR image using any QR library.

```json
{
  "event": "qrcode",
  "token": "60123456789",
  "data": "2@ABC123DEF456..."
}
```

Multiple `qrcode` events may be sent as each QR code has a short expiry. Always render the latest one.

#### `pairing-code`

Sent when a pairing code is generated for phone-based linking.

```json
{
  "event": "pairing-code",
  "token": "60123456789",
  "data": {
    "code": "ABCD-EFGH"
  }
}
```

#### `connection-success`

Device connected to WhatsApp successfully.

```json
{
  "event": "connection-success",
  "token": "60123456789"
}
```

#### `connection-error`

Connection failed, device disconnected, or was logged out.

```json
{
  "event": "connection-error",
  "token": "60123456789",
  "message": "Disconnected"
}
```

---

## Webhooks

### Configuration

Set `WEBHOOK_URL` in your `.env` to enable webhooks. Optionally set `WEBHOOK_SECRET` for request signing.

### Request Format

```http
POST /your/webhook/endpoint
Content-Type: application/json
User-Agent: wa-gateway-go/1.0
X-Webhook-Signature: <hmac-sha256-hex>
```

```json
{
  "event": "device.connected",
  "token": "60123456789",
  "data": null,
  "timestamp": "2026-02-17T10:30:00Z"
}
```

### Signature Verification

If `WEBHOOK_SECRET` is configured, verify the request signature:

```python
# Python example
import hmac, hashlib

expected = hmac.new(
    webhook_secret.encode(),
    request.body,
    hashlib.sha256
).hexdigest()

assert request.headers['X-Webhook-Signature'] == expected
```

```php
// PHP example
$expected = hash_hmac('sha256', $request->getContent(), $webhookSecret);
$valid = hash_equals($expected, $request->header('X-Webhook-Signature'));
```

```go
// Go example
mac := hmac.New(sha256.New, []byte(webhookSecret))
mac.Write(body)
expected := hex.EncodeToString(mac.Sum(nil))
valid := hmac.Equal([]byte(expected), []byte(signature))
```

### Events

#### `device.connected`

```json
{
  "event": "device.connected",
  "token": "60123456789",
  "data": null,
  "timestamp": "2026-02-17T10:30:00Z"
}
```

#### `device.disconnected`

```json
{
  "event": "device.disconnected",
  "token": "60123456789",
  "data": null,
  "timestamp": "2026-02-17T10:30:00Z"
}
```

#### `device.logged_out`

Device was removed from WhatsApp (user unlinked from phone). Must re-pair.

```json
{
  "event": "device.logged_out",
  "token": "60123456789",
  "data": null,
  "timestamp": "2026-02-17T10:30:00Z"
}
```

#### `message.receipt`

Delivery, read, or played receipt for a sent message.

```json
{
  "event": "message.receipt",
  "token": "60123456789",
  "data": {
    "type": "read",
    "messageIds": ["3EB0ABC123456789"],
    "from": "60198765432@s.whatsapp.net",
    "timestamp": "2026-02-17T10:31:00Z"
  },
  "timestamp": "2026-02-17T10:31:00Z"
}
```

Receipt types: `delivered`, `read`, `played` (for voice messages).

#### `contacts.new`

New contact captured from an incoming message.

```json
{
  "event": "contacts.new",
  "token": "60123456789",
  "data": {
    "phone": "60198765432",
    "name": "Ali"
  },
  "timestamp": "2026-02-17T10:30:00Z"
}
```

#### `contacts.sync`

Batch of contacts discovered from WhatsApp history sync (happens on first connect).

```json
{
  "event": "contacts.sync",
  "token": "60123456789",
  "data": [
    {
      "phone": "60198765432",
      "name": "Ali",
      "source": "history",
      "firstSeen": "2026-02-17T08:00:00Z",
      "lastSeen": "2026-02-17T08:00:00Z"
    }
  ],
  "timestamp": "2026-02-17T10:30:00Z"
}
```
