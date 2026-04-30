# Token Introspection Endpoint

The Token Introspection endpoint (`/oauth2/introspect`) validates and returns information about an access token (RFC 7662).

## Endpoint

```
POST /oauth2/introspect
```

## Request

```bash
curl -kL -X POST https://localhost:8090/oauth2/introspect \
  -H 'Content-Type: application/x-www-form-urlencoded' \
  -u 'client_id:client_secret' \
  -d 'token=ACCESS_TOKEN'
```

## Response

**Active Token:**

```json
{
  "active": true,
  "scope": "openid profile email",
  "client_id": "client_id",
  "username": "user@example.com",
  "exp": 1234567890,
  "iat": 1234564290
}
```

**Inactive Token:**

```json
{
  "active": false
}
```

## Related Documentation

- [Token Endpoint](token-endpoint.md) - How tokens are issued

