# Client Orchestrated Login Flows (Individual Authentication APIs)

Thunder provides standalone authentication APIs that can be used to authenticate users independently without a server orchestrated flow. These APIs are useful for simple authentication scenarios or when you want to build custom authentication flows in your application.

## Credentials Authentication

The credentials authentication API allows you to authenticate users by providing any user identifying attribute (such as username, email, or mobile number) along with any credentials attribute (such as password or PIN).

1. **Create a User**

    First, create an organization unit and a user with the desired identifying and credentials attributes:

    ```bash
    curl -kL -X POST -H 'Content-Type: application/json' -H 'Accept: application/json' https://localhost:8090/organization-units \
    -d '{
        "name": "Asgard",
        "description": "Realm of the gods",
        "handle": "asgard"
    }'
    ```

    Note the `id` from the response above, then use it to create a user:

    ```bash
    curl -kL -H 'Content-Type: application/json' https://localhost:8090/users \
    -d '{
        "organizationUnit": "{ou-id-from-above}",
        "type": "superhuman",
        "attributes": {
            "username": "thor",
            "password": "<password>",
            "email": "thor@thunder.sky",
            "given_name": "Thor",
            "family_name": "Odinson"
        }
    }'
    ```

2. **Authenticate with Credentials**

    Use the credentials authentication API to authenticate the user. You can provide any user identifying attribute along with any credentials attribute.

    ```bash
    curl -kL -X POST -H 'Content-Type: application/json' -H 'Accept: application/json' https://localhost:8090/auth/credentials/authenticate \
    -d '{
        "username": "thor",
        "password": "<password>"
    }'
    ```

    If authentication is successful, you'll receive a response with the user details and an assertion token:

    ```json
    {
        "id": "550e8400-e29b-41d4-a716-446655440000",
        "type": "superhuman",
        "organization_unit": "660e8400-e29b-41d4-a716-446655440000",
        "assertion": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
    }
    ```

    **For step-up authentication**, you can pass an existing assertion token to enrich it with credentials authentication:

    ```bash
    curl -kL -X POST -H 'Content-Type: application/json' -H 'Accept: application/json' https://localhost:8090/auth/credentials/authenticate \
    -d '{
        "username": "thor",
        "password": "<password>",
        "assertion": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
    }'
    ```

    If authentication is successful, you'll receive a response with the user details and an assertion token (enriched with credentials authentication if an existing assertion was provided):

    ```json
    {
        "id": "550e8400-e29b-41d4-a716-446655440000",
        "type": "superhuman",
        "organization_unit": "660e8400-e29b-41d4-a716-446655440000",
        "assertion": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
    }
    ```

## SMS OTP Authentication

The SMS OTP authentication API allows you to authenticate users by sending a One Time Password to their mobile number via SMS.

1. **Configure a Message Sender**

    Configure a message sender to send SMS messages. You can use services like Twilio, Vonage, or a custom service.

2. **Create a Message Sender**

    Create a message sender using the following cURL command:

    ```bash
    curl -kL -H 'Content-Type: application/json' -H 'Accept: application/json' https://localhost:8090/notification-senders/message \
    -d '{
      "name": "Custom SMS Sender",
      "description": "Sender for sending SMS messages",
      "provider": "custom",
      "properties": [
        {
          "name": "url",
          "value": "<custom_sms_provider_url>"
        },
        {
          "name": "http_method",
          "value": "POST"
        },
        {
          "name": "content_type",
          "value": "JSON"
        }
      ]
    }'
    ```

    Note the `id` from the response - you'll need it as the `sender_id` in the next steps.

    > Note: Refer [Configuring Message Senders](/../../notification-sender/configure-message-senders.md) for more details on configuring message senders.

3. **Send SMS OTP**

    Send an OTP to a user's mobile number:

    ```bash
    curl -kL -X POST -H 'Content-Type: application/json' -H 'Accept: application/json' https://localhost:8090/auth/otp/sms/send \
    -d '{
        "sender_id": "550e8400-e29b-41d4-a716-446655440000",
        "recipient": "+1234567890"
    }'
    ```

    You'll receive a response with a session token:

    ```json
    {
        "status": "SUCCESS",
        "session_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
    }
    ```

4. **Verify SMS OTP**

    Verify the OTP received by the user:

    ```bash
    curl -kL -X POST -H 'Content-Type: application/json' -H 'Accept: application/json' https://localhost:8090/auth/otp/sms/verify \
    -d '{
        "session_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
        "otp": "123456"
    }'
    ```

    **For step-up authentication**, you can pass an existing assertion token to enrich it with OTP authentication:

    ```bash
    curl -kL -X POST -H 'Content-Type: application/json' -H 'Accept: application/json' https://localhost:8090/auth/otp/sms/verify \
    -d '{
        "session_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
        "otp": "123456",
        "assertion": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
    }'
    ```

    If the OTP is valid, you'll receive a response with the user details and an assertion token (enriched with OTP authentication if an existing assertion was provided):

    ```json
    {
        "id": "550e8400-e29b-41d4-a716-446655440000",
        "type": "superhuman",
        "organization_unit": "660e8400-e29b-41d4-a716-446655440000",
        "assertion": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
    }
    ```

## Google OAuth Authentication

The Google OAuth authentication API allows you to authenticate users using their Google account through the OAuth 2.0 flow.

1. **Create a Google OAuth Application**

    - Go to the [Google Cloud Console](https://console.cloud.google.com/)
    - Create a new OAuth 2.0 Client ID
    - Set the authorized redirect URI to `https://localhost:8090/auth/oauth/google/callback`
    - Copy the **Client ID** and **Client Secret**

2. **Configure the Google Identity Provider**

    Create a Google identity provider in Thunder:

    ```bash
    curl -kL -H 'Content-Type: application/json' -H 'Accept: application/json' https://localhost:8090/identity-providers \
    -d '{
      "name": "Google",
      "description": "Login with Google",
      "type": "GOOGLE",
      "properties": [
        {
          "name": "client_id",
          "value": "<google_client_id>",
          "is_secret": false
        },
        {
          "name": "client_secret",
          "value": "<google_client_secret>",
          "is_secret": true
        },
        {
          "name": "redirect_uri",
          "value": "https://localhost:8090/auth/oauth/google/callback",
          "is_secret": false
        },
        {
          "name": "scopes",
          "value": "openid,email,profile",
          "is_secret": false
        }
      ]
    }'
    ```

    Note the `id` from the response - you'll need it as the `idp_id` in the next steps.

    > Note: Refer [Configuring Identity Providers](/../../identity-provider/configure-identity-providers.md) for more details on configuring identity providers.

3. **Start Google Authentication**

    Initiate the Google OAuth flow:

    ```bash
    curl -kL -X POST -H 'Content-Type: application/json' -H 'Accept: application/json' https://localhost:8090/auth/oauth/google/start \
    -d '{
        "idp_id": "550e8400-e29b-41d4-a716-446655440000"
    }'
    ```

    You'll receive a response with the redirect URL and session token:

    ```json
    {
        "redirect_url": "https://accounts.google.com/o/oauth2/auth?client_id=...",
        "session_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
    }
    ```

4. **Complete Google Authentication**

    After the user authenticates with Google and is redirected back, complete the authentication:

    ```bash
    curl -kL -X POST -H 'Content-Type: application/json' -H 'Accept: application/json' https://localhost:8090/auth/oauth/google/finish \
    -d '{
        "session_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
        "code": "<authorization_code_from_google>"
    }'
    ```

    **For step-up authentication**, you can pass an existing assertion token to enrich it with Google authentication:

    ```bash
    curl -kL -X POST -H 'Content-Type: application/json' -H 'Accept: application/json' https://localhost:8090/auth/oauth/google/finish \
    -d '{
        "session_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
        "code": "<authorization_code_from_google>",
        "assertion": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
    }'
    ```

    If authentication is successful, you'll receive a response with the user details and an assertion token (enriched with Google authentication if an existing assertion was provided):

    ```json
    {
        "id": "550e8400-e29b-41d4-a716-446655440000",
        "type": "superhuman",
        "organization_unit": "660e8400-e29b-41d4-a716-446655440000",
        "assertion": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
    }
    ```

## GitHub OAuth Authentication

The GitHub OAuth authentication API allows you to authenticate users using their GitHub account through the OAuth 2.0 flow.

1. **Create a GitHub OAuth Application**

    - Go to GitHub Settings > Developer settings > OAuth Apps
    - Create a new OAuth App
    - Set the Authorization callback URL to `https://localhost:8090/auth/oauth/github/callback`
    - Copy the **Client ID** and **Client Secret**

2. **Configure the GitHub Identity Provider**

    Create a GitHub identity provider in Thunder:

    ```bash
    curl -kL -H 'Content-Type: application/json' -H 'Accept: application/json' https://localhost:8090/identity-providers \
    -d '{
      "name": "Github",
      "description": "Login with Github",
      "type": "GITHUB",
      "properties": [
        {
          "name": "client_id",
          "value": "<github_client_id>",
          "is_secret": false
        },
        {
          "name": "client_secret",
          "value": "<github_client_secret>",
          "is_secret": true
        },
        {
          "name": "redirect_uri",
          "value": "https://localhost:8090/auth/oauth/github/callback",
          "is_secret": false
        },
        {
          "name": "scopes",
          "value": "user:email,read:user",
          "is_secret": false
        }
      ]
    }'
    ```

    Note the `id` from the response - you'll need it as the `idp_id` in the next steps.

    > Note: Refer [Configuring Identity Providers](/../../identity-provider/configure-identity-providers.md) for more details on configuring identity providers.

3. **Start GitHub Authentication**

    Initiate the GitHub OAuth flow:

    ```bash
    curl -kL -X POST -H 'Content-Type: application/json' -H 'Accept: application/json' https://localhost:8090/auth/oauth/github/start \
    -d '{
        "idp_id": "550e8400-e29b-41d4-a716-446655440000"
    }'
    ```

    You'll receive a response with the redirect URL and session token:

    ```json
    {
        "redirect_url": "https://github.com/login/oauth/authorize?client_id=...",
        "session_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
    }
    ```

4. **Complete GitHub Authentication**

    After the user authenticates with GitHub and is redirected back, complete the authentication:

    ```bash
    curl -kL -X POST -H 'Content-Type: application/json' -H 'Accept: application/json' https://localhost:8090/auth/oauth/github/finish \
    -d '{
        "session_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
        "code": "<authorization_code_from_github>"
    }'
    ```

    **For step-up authentication**, you can pass an existing assertion token to enrich it with GitHub authentication:

    ```bash
    curl -kL -X POST -H 'Content-Type: application/json' -H 'Accept: application/json' https://localhost:8090/auth/oauth/github/finish \
    -d '{
        "session_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
        "code": "<authorization_code_from_github>",
        "assertion": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
    }'
    ```

    If authentication is successful, you'll receive a response with the user details and an assertion token (enriched with Github authentication if an existing assertion was provided):

    ```json
    {
        "id": "550e8400-e29b-41d4-a716-446655440000",
        "type": "superhuman",
        "organization_unit": "660e8400-e29b-41d4-a716-446655440000",
        "assertion": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
    }
    ```

## Standard OAuth IDP Authentication

The Standard OAuth IDP authentication API allows you to authenticate users using any standard OAuth 2.0 / OpenID Connect compliant identity provider.

1. **Create an OAuth Application in Your Identity Provider**

    - Configure an OAuth 2.0 application in your identity provider
    - Set the redirect URI to `https://localhost:8090/auth/oauth/standard/callback`
    - Copy the **Client ID**, **Client Secret**, **Authorization Endpoint**, and **Token Endpoint**

2. **Configure the Standard OAuth Identity Provider**

    Create a standard OAuth identity provider in Thunder. The exact properties required may vary depending on your OAuth provider:

    ```bash
    curl -kL -X POST -H 'Content-Type: application/json' -H 'Accept: application/json' https://localhost:8090/identity-providers \
    -d '{
      "name": "Custom OAuth IDP",
      "description": "Standard OAuth identity provider",
      "type": "OAUTH",
      "properties": [
        {
          "name": "client_id",
          "value": "<oauth_client_id>",
          "is_secret": false
        },
        {
          "name": "client_secret",
          "value": "<oauth_client_secret>",
          "is_secret": true
        },
        {
          "name": "authorization_endpoint",
          "value": "https://provider.com/oauth2/authorize",
          "is_secret": false
        },
        {
          "name": "token_endpoint",
          "value": "https://provider.com/oauth2/token",
          "is_secret": false
        },
        {
          "name": "user_info_endpoint",
          "value": "https://provider.com/oauth2/userinfo",
          "is_secret": false
        },
        {
          "name": "redirect_uri",
          "value": "https://localhost:8090/auth/oauth/standard/callback",
          "is_secret": false
        },
        {
          "name": "scopes",
          "value": "openid,profile,email",
          "is_secret": false
        }
      ]
    }'
    ```

    Note the `id` from the response - you'll need it as the `idp_id` in the next steps.

    > Note: Refer [Configuring Identity Providers](/../../identity-provider/configure-identity-providers.md) for more details on configuring identity providers.

3. **Start Standard OAuth Authentication**

    Initiate the OAuth flow:

    ```bash
    curl -kL -X POST -H 'Content-Type: application/json' -H 'Accept: application/json' https://localhost:8090/auth/oauth/standard/start \
    -d '{
        "idp_id": "550e8400-e29b-41d4-a716-446655440000"
    }'
    ```

    You'll receive a response with the redirect URL and session token:

    ```json
    {
        "redirect_url": "https://provider.com/oauth2/authorize?client_id=...",
        "session_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
    }
    ```

4. **Complete Standard OAuth Authentication**

    After the user authenticates with the identity provider and is redirected back, complete the authentication:

    ```bash
    curl -kL -X POST -H 'Content-Type: application/json' -H 'Accept: application/json' https://localhost:8090/auth/oauth/standard/finish \
    -d '{
        "session_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
        "code": "<authorization_code_from_provider>"
    }'
    ```

    **For step-up authentication**, you can pass an existing assertion token to enrich it with OAuth authentication:

    ```bash
    curl -kL -X POST -H 'Content-Type: application/json' -H 'Accept: application/json' https://localhost:8090/auth/oauth/standard/finish \
    -d '{
        "session_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
        "code": "<authorization_code_from_provider>",
        "assertion": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
    }'
    ```

    If authentication is successful, you'll receive a response with the user details and an assertion token (enriched with OAuth IDP authentication if an existing assertion was provided):

    ```json
    {
        "id": "550e8400-e29b-41d4-a716-446655440000",
        "type": "superhuman",
        "organization_unit": "660e8400-e29b-41d4-a716-446655440000",
        "assertion": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
    }
    ```

## Working with Assertion Tokens and Step-Up Authentication

### Assertion Tokens

When you successfully authenticate using the individual authentication APIs, Thunder returns an assertion token (JWT) in the response. This token contains authentication assurance information including:

- Authentication methods used: Which authentication mechanisms were completed (e.g., password, OTP, social login)
- Assurance level: The strength/confidence level of the authentication
- User identity: Information about the authenticated user

You can control whether the assertion token is returned by using the `skip_assertion` parameter in your API requests:

- `skip_assertion: false` (default) - The API returns the assertion token in the response
- `skip_assertion: true` - The API skips generating and returning the assertion token

Example:

```bash
curl -kL -X POST -H 'Content-Type: application/json' -H 'Accept: application/json' https://localhost:8090/auth/credentials/authenticate \
-d '{
    "username": "thor",
    "password": "<password>",
    "skip_assertion": true
}'
```

### Step-Up Authentication

Thunder supports step-up authentication, allowing you to progressively increase the authentication assurance level by combining multiple authentication methods. This is particularly useful for scenarios requiring stronger authentication for sensitive operations.

To perform step-up authentication, pass an existing assertion token when calling supported authentication APIs. The API will enrich the existing assertion with the new authentication method, creating a stronger authentication context.

Supported APIs for step-up authentication:
- Credentials Authentication (`/auth/credentials/authenticate`)
- SMS OTP Verification (`/auth/otp/sms/verify`)
- Google OAuth (`/auth/oauth/google/finish`)
- GitHub OAuth (`/auth/oauth/github/finish`)
- Standard OAuth (`/auth/oauth/standard/finish`)

For example, consider a scenario where a user first authenticates with a password and later needs to perform a sensitive operation requiring OTP verification. The user can first authenticate with their password to receive an assertion token, then verify the OTP while passing the existing assertion token to receive an enhanced assertion token that includes both authentication methods.

Example of step-up authentication with SMS OTP verification:

```bash
curl -kL -X POST -H 'Content-Type: application/json' -H 'Accept: application/json' https://localhost:8090/auth/otp/sms/verify \
-d '{
    "session_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "otp": "123456",
    "assertion": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}'
```
