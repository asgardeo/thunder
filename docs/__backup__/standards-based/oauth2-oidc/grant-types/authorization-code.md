# Authorization Code

The Authorization Code grant type involves redirecting users to an authorization server where they authenticate and authorize the application, then exchanging an authorization code for an access token.

## Overview

The Authorization Code grant type is the recommended OAuth 2.0 grant type for applications that require user authentication. It can be used with both confidential clients (with client secrets) and public clients (with PKCE).

**Process Steps:**
1. Client redirects user to authorization endpoint
2. User authenticates and authorizes the application
3. Authorization server redirects back with an authorization code
4. Client exchanges authorization code for access token

## Use Cases

- Web applications (both server-side and client-side)
- Mobile applications (with PKCE)
- Single Page Applications (SPAs) with PKCE

## Prerequisites

1. **Create an OAuth Application** with `authorization_code` grant type
2. **Configure Gate Client** (optional for custom user authentication UI)
3. **Create Users** in Thunder (for testing)

## Step-by-Step Guide

### Step 1: Obtain Admin Token

First, obtain an admin token to create applications.

#### 1.1: Initiate the Authentication Flow

Run the following command, replacing `<application_id>` with your sample app ID (created during Thunder setup).

```bash
FLOW_RESPONSE=$(curl -k -s -X POST 'https://localhost:8090/flow/execute' \
  -d '{"applicationId":"<application_id>","flowType":"AUTHENTICATION"}')

FLOW_ID=$(echo $FLOW_RESPONSE | jq -r '.flowId')
```

#### 1.2: Submit Admin Credentials

Run the following command with the extracted `flowId`.

```bash
ADMIN_TOKEN_RESPONSE=$(curl -k -s -X POST 'https://localhost:8090/flow/execute' \
  -d '{"flowId":"'$FLOW_ID'","inputs":{"username":"admin","password":"admin","requested_permissions":"system"},"action":"action_001"}')

ADMIN_TOKEN=$(echo $ADMIN_TOKEN_RESPONSE | jq -r '.assertion')
```

**Response:**

```json
{
  "flowId": "2c6d4c45-3de9-4a70-ae6b-ba1d034af6bc",
  "flowStatus": "COMPLETE",
  "data": {},
  "assertion": "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

### Step 2: Create an OAuth Application

Create an application with OAuth 2.0 configuration supporting the authorization code grant type:

```bash
curl -kL -X POST https://localhost:8090/applications \
  -H 'Content-Type: application/json' \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -d '{
    "name": "My Web Application",
    "description": "Web application using authorization code grant type",
    "auth_flow_id": "<auth-flow-uuid>",
    "inbound_auth_config": [
      {
        "type": "oauth2",
        "config": {
          "client_id": "my_web_app_client",
          "client_secret": "my_web_app_secret",
          "redirect_uris": [
            "https://localhost:3000/callback"
          ],
          "grant_types": [
            "authorization_code",
            "refresh_token"
          ],
          "response_types": [
            "code"
          ],
          "token_endpoint_auth_method": "client_secret_basic",
          "pkce_required": false,
          "public_client": false,
          "scopes": [
            "openid",
            "profile",
            "email"
          ],
          "token": {
            "issuer": "thunder",
            "access_token": {
              "validity_period": 3600,
              "user_attributes": [
                "email",
                "username",
                "given_name",
                "family_name"
              ]
            },
            "id_token": {
              "validity_period": 3600,
              "user_attributes": [
                "sub",
                "email",
                "name",
                "given_name",
                "family_name"
              ],
              "scope_claims": {
                "profile": [
                  "name",
                  "given_name",
                  "family_name",
                  "picture"
                ],
                "email": [
                  "email",
                  "email_verified"
                ]
              }
            }
          }
        }
      }
    ]
  }'
```

**Response:**

```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "name": "My Web Application",
  "description": "Web application using authorization code grant type",
  "client_id": "my_web_app_client",
  "client_secret": "my_web_app_secret",
  "auth_flow_id": "<auth-flow-uuid>",
  "inbound_auth_config": [
    {
      "type": "oauth2",
      "config": {
        "client_id": "my_web_app_client",
        "client_secret": "my_web_app_secret",
        "redirect_uris": [
          "https://localhost:3000/callback"
        ],
        "grant_types": [
          "authorization_code",
          "refresh_token"
        ],
        "response_types": [
          "code"
        ],
        "token_endpoint_auth_method": "client_secret_basic",
        "pkce_required": false,
        "public_client": false,
        "scopes": [
          "openid",
          "profile",
          "email"
        ]
      }
    }
  ]
}
```

**Key Configuration:**
- `grant_types`: Must include `"authorization_code"`
- `response_types`: Must include `"code"`
- `redirect_uris`: Valid redirect URIs for your application
- `token_endpoint_auth_method`:
  - `"client_secret_basic"` or `"client_secret_post"` for confidential clients
  - `"none"` for public clients (must use PKCE)
- `public_client`: Set to `true` for public clients (mobile apps, SPAs)
- `pkce_required`: Set to `true` for enhanced security (required for public clients, recommended for all)
- `scopes`: OIDC scopes like `"openid"`, `"profile"`, `"email"`

### Step 3: Configure Gate Client (Optional)

If you want to have your custom login UI, configure the gate client in `deployment.yaml`:

```yaml
gate_client:
  hostname: "localhost"
  port: 9090
  scheme: "https"
  login_path: "/login"
  error_path: "/error"
```

Restart Thunder after making this change.

### Step 4: Initiate Authorization Request

Redirect the user to the authorization endpoint:

```
https://localhost:8090/oauth2/authorize?
  response_type=code&
  client_id=my_web_app_client&
  redirect_uri=https://localhost:3000/callback&
  scope=openid%20profile%20email&
  state=random_state_value_12345
```

**Query Parameters:**
- `response_type`: Must be `"code"`
- `client_id`: Your OAuth client ID
- `redirect_uri`: Must match one of the registered redirect URIs
- `scope`: Space-separated list of scopes (URL-encoded)
- `state`: Random value for CSRF protection (recommended)

**Optional Parameters:**
- `code_challenge`: PKCE code challenge (if PKCE is required)
- `code_challenge_method`: `"S256"` (if PKCE is required)
- `resource`: Target resource/audience (RFC 8707)

### Step 5: User Authentication

The user will be redirected to Thunder's authentication page (or your custom gate client). After successful authentication, Thunder redirects back to your application:

```
https://localhost:3000/callback?code=AUTHORIZATION_CODE&state=random_state_value_12345
```

**Extract the authorization code** from the `code` query parameter.

### Step 6: Exchange Code for Token

Exchange the authorization code for an access token.

**For Confidential Clients** (with client secret):

```bash
curl -kL -X POST https://localhost:8090/oauth2/token \
  -H 'Content-Type: application/x-www-form-urlencoded' \
  -u 'my_web_app_client:my_web_app_secret' \
  -d 'grant_type=authorization_code' \
  -d 'code=AUTHORIZATION_CODE' \
  -d 'redirect_uri=https://localhost:3000/callback'
```

**For Public Clients** (without client secret, PKCE required):

```bash
curl -kL -X POST https://localhost:8090/oauth2/token \
  -H 'Content-Type: application/x-www-form-urlencoded' \
  -d 'grant_type=authorization_code' \
  -d 'code=AUTHORIZATION_CODE' \
  -d 'redirect_uri=https://localhost:3000/callback' \
  -d 'client_id=my_web_app_client' \
  -d 'code_verifier=CODE_VERIFIER_VALUE'
```

**Response (with `openid` scope):**

```json
{
  "access_token": "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9...",
  "token_type": "Bearer",
  "expires_in": 3600,
  "refresh_token": "refresh_token_xyz123",
  "scope": "openid profile email",
  "id_token": "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```



## PKCE Support

Thunder supports PKCE (Proof Key for Code Exchange) for enhanced security, mandatory for public clients adhering to OAuth 2.0.

### Using PKCE

1. **Generate Code Verifier and Challenge** (client-side):

```javascript
// Generate a random code verifier (43-128 characters)
const codeVerifier = generateRandomString(43);

// Generate code challenge using S256
const codeChallenge = await sha256(codeVerifier);
const codeChallengeMethod = "S256";
```

2. **Include PKCE Parameters in Authorization Request**:

```
https://localhost:8090/oauth2/authorize?
  response_type=code&
  client_id=my_web_app_client&
  redirect_uri=https://localhost:3000/callback&
  scope=openid%20profile%20email&
  state=random_state_value_12345&
  code_challenge=CODE_CHALLENGE&
  code_challenge_method=S256
```

3. **Include Code Verifier in Token Request**:

```bash
curl -kL -X POST https://localhost:8090/oauth2/token \
  -H 'Content-Type: application/x-www-form-urlencoded' \
  -u 'my_web_app_client:my_web_app_secret' \
  -d 'grant_type=authorization_code' \
  -d 'code=AUTHORIZATION_CODE' \
  -d 'redirect_uri=https://localhost:3000/callback' \
  -d 'code_verifier=CODE_VERIFIER'
```

**PKCE Code Challenge Methods:**
- `S256`: SHA256 hash (recommended)
- `plain`: Plain text (less secure)

## Resource Parameter (RFC 8707)

You can specify a target resource/audience using the `resource` parameter:

```
https://localhost:8090/oauth2/authorize?
  response_type=code&
  client_id=my_web_app_client&
  redirect_uri=https://localhost:3000/callback&
  scope=openid%20profile%20email&
  resource=https://api.example.com/resource
```

The resource parameter must be:
- An absolute URI
- Without a fragment component

## Token Response

### Access Token

The access token is a JWT containing:
- `sub`: Subject (user ID)
- `aud`: Audience (client ID or resource)
- `scope`: Granted scopes
- `exp`: Expiration time
- `iat`: Issued at time
- User attributes (as configured)

### ID Token

The ID token is only returned when the `openid` scope is requested. If `openid` is not included in the requested scopes, no ID token will be returned.

When `openid` scope is requested, an ID token is included with:
- `sub`: Subject (user ID)
- `aud`: Audience (client ID)
- `iss`: Issuer
- `exp`: Expiration time
- `iat`: Issued at time
- `auth_time`: Authentication time
- User claims based on requested scopes

### Refresh Token

A refresh token is issued if the application supports `refresh_token` grant type. Use it to obtain new access tokens without re-authentication.

## Error Handling

Common errors and solutions:

| Error | Description | Solution |
|-------|-------------|----------|
| `invalid_request` | Missing required parameter | Include all required parameters |
| `invalid_client` | Invalid client credentials | Verify client ID and secret |
| `invalid_grant` | Invalid authorization code | Code may be expired or already used |
| `invalid_scope` | Invalid scope requested | Check allowed scopes for the application |
| `unauthorized_client` | Client not authorized for grant type | Verify grant type is enabled |



## Complete Example

See the [React Vanilla Sample Application](../../../../../samples/apps/react-vanilla-sample/) for a complete implementation example.

## Related Documentation

- [PKCE Implementation](../features/pkce.md) - Detailed PKCE guide
- [Refresh Token](refresh-token.md) - Using refresh tokens
- [Token Endpoint](../endpoints/token-endpoint.md) - Token endpoint details

