# UserInfo Endpoint

The UserInfo endpoint (`/oauth2/userinfo`) is an OpenID Connect endpoint that returns claims about the authenticated user based on the access token.

## Endpoint

```
GET /oauth2/userinfo
POST /oauth2/userinfo
```

## Authentication

The UserInfo endpoint requires Bearer token authentication:

```
Authorization: Bearer <access_token>
```

## Request

**GET Request:**

```bash
curl -kL -X GET https://localhost:8090/oauth2/userinfo \
  -H 'Authorization: Bearer ACCESS_TOKEN'
```

**POST Request:**

```bash
curl -kL -X POST https://localhost:8090/oauth2/userinfo \
  -H 'Authorization: Bearer ACCESS_TOKEN' \
  -H 'Content-Type: application/x-www-form-urlencoded'
```

## Response

**Success Response (200 OK):**

```json
{
  "sub": "user-123",
  "email": "user@example.com",
  "email_verified": true,
  "name": "John Doe",
  "given_name": "John",
  "family_name": "Doe",
  "picture": "https://example.com/avatar.jpg"
}
```

The response includes claims based on:
- Scopes requested (e.g., `profile`, `email`)
- User attributes configured in the application
- Scope claims mapping

## Response Headers

- `Content-Type: application/json`
- `Cache-Control: no-store`
- `Pragma: no-cache`

## Error Responses

**Invalid Token (401 Unauthorized):**

```json
{
  "error": "invalid_token",
  "error_description": "The access token provided is expired, revoked, malformed, or invalid"
}
```

## Related Documentation

- [Authorization Code](../grant-types/authorization-code.md) - How to obtain access token

