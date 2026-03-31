# Wildcard Redirect URI Support

## Problem statement

Thunder currently validates redirect URIs with exact string matching only. This feature adds wildcard support (`*` for single path segment, `**` for multi-segment) in the path component of redirect URIs. Protocol, domain, and query string must never contain wildcards. Token requests retain exact matching per RFC 6749 §4.1.3.

## Goals

- Support pattern for the URL path only.
- Protocol and domain name cannot be given as pattern.
- In the OIDC authorization call, redirect_uri must be absolute. It must match with at least one of the registered redirect uris
- Only simple wildcard syntax (e.g., * for a single path segment and ** for multiple path segment) is supported. Full regular expression syntax is out of scope.
- The redirect_uri in token requests is still validated with exact match logic per RFC 6749 §4.1.3; pattern matching is not applied there.
- Invalid redirect_uri should inform the resource owner of the error and MUST NOT automatically redirect the user-agent to the invalid redirection URI
- Support deeplinks for the mobile app redirections. wildcards are not allowed in the scheme part of the deeplink.
- At least one redirection uri MUST be registrated per client.

## Non-Goals

- Query params / fragments — patterns are path-only, so these are never wildcard-matched.
- Unregistered dynamic URIs — enforces the RFC 6749 requirement that all redirect URIs be pre-registered.

## Tasks

- Add support in the backend to registration and authorize endpoints
- Add display of "Invalid Request" error message in the UI when the request is rejected.


## Acceptance Criteria

| ID | Area | Criteria |
|----|------|----------|
| AC-01 | Registration | Wildcard in protocol (e.g., `http*://example.com/callback`) is rejected with `400 Bad Request` |
| AC-02 | Registration | Wildcard in host/domain (e.g., `https://*.example.com/callback`) is rejected with `400 Bad Request` |
| AC-03 | Registration | Wildcard in query string (e.g., `https://example.com/callback?foo=*`) is rejected with `400 Bad Request` |
| AC-04 | Registration | Wildcard only in path (e.g., `https://example.com/callback/*`) is accepted and stored as-is |
| AC-05 | Registration | Full regex syntax (e.g., `https://example.com/callback/[a-z]+`) is rejected; only `*` and `**` are supported |
| AC-06 | Authorization | Single-segment wildcard `*` matches exactly one path segment (e.g., `callback/*` matches `callback/v1` but not `callback/v1/extra`) |
| AC-07 | Authorization | Multi-segment wildcard `**` matches zero or more path segments (e.g., `app/**/callback` matches both `app/callback` and `app/tenant/region/callback`) |
| AC-08 | Authorization | Exact redirect URI match succeeds when no wildcard is registered |
| AC-09 | Authorization | Authorization request whose `redirect_uri` matches no registered URI returns `invalid_request` per RFC 6749 |
| AC-10 | Authorization | Query strings must match exactly between the incoming `redirect_uri` and the registered uri. A mismatch in query parameters (present on one side but not the other, or differing values) causes the request to be rejected |
| AC-11 | Authorization | Multiple registered URIs (mix of exact and wildcard) are each evaluated in registration order; first match wins |
| AC-12 | Authorization | If `redirect_uri` is omitted in the authorization request and the single registered URI contains a wildcard, the request is rejected — a wildcard pattern cannot serve as a concrete redirect target |
| AC-13 | Registration | Deeplink URI with wildcard in path (e.g., `myapp://callback/*`) is accepted and stored as-is |
| AC-14 | Registration | Wildcard in the scheme of a deeplink (e.g., `my*app://callback`) is rejected with `400 Bad Request` |
| AC-15 | Token request | `redirect_uri` in the token request is validated with exact match only per RFC 6749 §4.1.3; pattern matching is not applied |
| AC-16 | Token request | `redirect_uri` in the token request must exactly match `redirect_uri` passed in the Authorization request. If redirect_uri does not match, token request is rejected with `400 Bad Request` |


## Technical Notes:

Add `MatchRedirectURIPattern(pattern, incoming string) (bool, error)` utility function to `http_util.go`.

Method expectation:
- Returns `(false, error)` if either URI is malformed (unparseable, missing scheme or host).
- Returns `(false, nil)` when the URIs are well-formed but do not match.
- Returns `(true, nil)` when the incoming URI matches the pattern.
- Scheme and host must match exactly (case-insensitive per RFC 3986).
- Query string must match exactly; a mismatch in query parameters returns `(false, nil)`.
- Wildcards (`*` and `**`) are supported only in the path component:
    - `*` matches exactly one path segment (no slashes).
    - `**` matches zero or more path segments.

### Files to Modify

| File | Purpose |
|------|---------|
| `backend/internal/application/service.go` | Add wildcard-aware validation at registration |
| `backend/internal/application/model/oauth_app.go` | Replace exact match (`slices.Contains`) with pattern match in `validateRedirectURI` |
| `backend/internal/system/utils/http_util.go` | Add reusable path wildcard matching utility |

Token validation in `backend/internal/oauth/oauth2/granthandlers/authorization_code.go` requires **no change** — it already performs exact match against the concrete URI stored in the auth code (correct per RFC 6749 §4.1.3).

---