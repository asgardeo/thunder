# JWKS Endpoint

The JWKS (JSON Web Key Set) endpoint (`/oauth2/jwks`) provides public keys for validating JWT tokens issued by Thunder.

## Endpoint

```
GET /oauth2/jwks
```

## Request

```bash
curl -kL https://localhost:8090/oauth2/jwks
```

## Response

```json
{
  "keys": [
    {
      "kty": "RSA",
      "kid": "key-id",
      "use": "sig",
      "n": "modulus...",
      "e": "AQAB",
      "alg": "RS256"
    }
  ]
}
```

## Usage

Use the JWKS to validate JWT tokens:

1. Extract `kid` from token header
2. Find matching key in JWKS
3. Use key to verify token signature

## Related Documentation

- [MCP Server Securing](../features/mcp-server-securing.md) - Token validation example

