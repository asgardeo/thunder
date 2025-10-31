# Configuring Message Senders

This guide provides instructions on how to configure message senders for sending OTPs and notifications in WSO2 Thunder. You can manage message notification senders through the REST API, allowing you to create, update, retrieve, and delete message senders.

---

## Managing Message Senders

You can manage message notification senders using the following REST API endpoints.

### List Message Notification Senders

Retrieve a list of all configured message notification senders.

```bash
curl -kL -H 'Accept: application/json' https://localhost:8090/notification-senders/message
```

### Create a Message Notification Sender

Create a new message notification sender. See sender-specific examples below for the request body.

```bash
curl -kL -H 'Content-Type: application/json' -H 'Accept: application/json' https://localhost:8090/notification-senders/message \
-d '{ ... }'
```

Refer [Message Sender Configuration](#message-sender-configuration) for details on the supported senders and their required properties.

### Get a Message Notification Sender by ID

Retrieve details of a specific message notification sender by its unique identifier.

```bash
curl -kL -H 'Accept: application/json' https://localhost:8090/notification-senders/message/<sender_id>
```

### Update a Message Notification Sender

Update an existing message notification sender by its unique identifier.

```bash
curl -kL -X PUT -H 'Content-Type: application/json' -H 'Accept: application/json' https://localhost:8090/notification-senders/message/<sender_id> \
-d '{ ... }'
```

### Delete a Message Notification Sender

Delete a message notification sender by its unique identifier.

```bash
curl -kL -X DELETE https://localhost:8090/notification-senders/message/<sender_id>
```

---

## Message Sender Configuration

To configure a message sender, you need to specify the `provider` type and the required properties for that provider. The `provider` can be one of the following: `twilio`, `vonage`, or `custom`. Each provider has its own set of required properties.

### Twilio

**Required Properties:**

| Property      | Description                                    | Example                            |
|---------------|------------------------------------------------|------------------------------------|
| `account_sid` | Twilio Account SID                             | ACXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX |
| `auth_token`  | Twilio Auth Token                              | your_auth_token                    |
| `sender_id`   | Sender ID or phone number                      | +1234567890                        |

**Example cURL:**

```bash
curl -kL -H 'Content-Type: application/json' -H 'Accept: application/json' https://localhost:8090/notification-senders/message \
-d '{
  "name": "Twilio SMS Sender",
  "description": "Sender for sending SMS messages using Twilio",
  "provider": "twilio",
  "properties": [
    {
      "name": "account_sid",
      "value": "ACXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX",
      "is_secret": true
    },
    {
      "name": "auth_token",
      "value": "your_auth_token",
      "is_secret": true
    },
    {
      "name": "sender_id",
      "value": "+1234567890"
    }
  ]
}'
```

### Vonage

**Required Properties:**

| Property    | Description                                    | Example          |
|-------------|------------------------------------------------|------------------|
| api_key     | Vonage API Key                                 | 12345678         |
| api_secret  | Vonage API Secret                              | your_api_secret  |
| sender_id   | Sender ID or phone number                      | VonageSender     |

**Example cURL:**

```bash
curl -kL -H 'Content-Type: application/json' -H 'Accept: application/json' https://localhost:8090/notification-senders/message \
-d '{
  "name": "Vonage SMS Sender",
  "description": "Sender for sending SMS messages using Vonage",
  "provider": "vonage",
  "properties": [
    {
      "name": "api_key",
      "value": "12345678",
      "is_secret": true
    },
    {
      "name": "api_secret",
      "value": "your_api_secret",
      "is_secret": true
    },
    {
      "name": "sender_id",
      "value": "VonageSender"
    }
  ]
}'
```

### Custom HTTP Notification Sender

**Required Properties:**

| Property      | Description                                                           | Example Value                                           |
|---------------|-----------------------------------------------------------------------|---------------------------------------------------------|
| url           | Endpoint URL for sending messages                                     | https://api.example.com/send-sms                        |
| http_method   | HTTP method to use. Ex: `POST`, `GET`                                 | POST                                                    |
| http_headers  | HTTP headers (string, comma-separated key:value pairs.                | Authorization:Bearer 1234,Content-Type:application/json |
| content_type  | Content type of the request payload. Can be one of `JSON` or `FORM`   | JSON                                                    |

**Example cURL:**

```bash
curl -kL -H 'Content-Type: application/json' -H 'Accept: application/json' https://localhost:8090/notification-senders/message \
-d '{
  "name": "Custom SMS Sender",
  "description": "Sender for sending SMS messages via custom provider",
  "provider": "custom",
  "properties": [
    {
      "name": "url",
      "value": "https://api.example.com/send-sms"
    },
    {
      "name": "http_method",
      "value": "POST"
    },
    {
      "name": "http_headers",
      "value": "Authorization:BearerToken,Content-Type:application/json"
    },
    {
      "name": "content_type",
      "value": "JSON"
    }
  ]
}'
```

---

For more details on the API, refer to the [OpenAPI specification](../../../api/notification-sender.yaml).
