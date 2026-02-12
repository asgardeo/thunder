# Authorization Endpoint

The Authorization endpoint (`/oauth2/authorize`) initiates the OAuth 2.0 authorization process.

## Endpoint

```
GET /oauth2/authorize
```

## Request Parameters

- `response_type`: Must be `"code"` for authorization code grant type
- `client_id`: OAuth client identifier
- `redirect_uri`: Redirect URI (must match registered URI)
- `scope`: Space-separated scopes (URL-encoded)
- `state`: CSRF protection token (recommended)
- `code_challenge`: PKCE code challenge (if PKCE is used)
- `code_challenge_method`: `"S256"` or `"plain"` (if PKCE is used)
- `resource`: Target resource/audience (RFC 8707, optional)

## Example

```
https://localhost:8090/oauth2/authorize?
  response_type=code&
  client_id=my_client&
  redirect_uri=https://localhost:3000/callback&
  scope=openid%20profile%20email&
  state=random_state_value
```

## Response

Redirects to `redirect_uri` with authorization code:

```
https://localhost:3000/callback?code=AUTHORIZATION_CODE&state=random_state_value
```

## Related Documentation

- [Authorization Code](../grant-types/authorization-code.md) - Complete grant type guide
- [PKCE](../features/pkce.md) - Using PKCE with authorization

