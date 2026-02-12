# Scopes and Permissions

Thunder supports both standard OpenID Connect (OIDC) scopes and custom permission scopes, providing fine-grained access control through integration with Thunder's Role-Based Access Control (RBAC) system.

## OAuth Scopes

**OAuth scopes** are strings that represent the access being requested. In Thunder, there are two types of scopes:

### 1. Standard OIDC Scopes

Standard OIDC scopes control what user identity information is included in **ID tokens**. These scopes are **not permission-based** - they're always granted if requested (assuming the user has the corresponding attributes).



**Important:** For user attributes to be included in the ID token, they must be listed in `token.id_token.user_attributes`, even if the corresponding scope is requested and the user has that attribute.

**Example Configuration:**
```json
{
  "inbound_auth_config": [{
    "type": "oauth2",
    "config": {
      "token": {
        "id_token": {
          "user_attributes": [
            "name",
            "email",
            "given_name",
            "family_name"
          ],
          "scope_claims": {
            "profile": ["name", "given_name", "family_name"],
            "email": ["email"]
          }
        }
      }
    }
  }]
}
```

In this example, if `scope=profile email` is requested:
- **ID Token** will include: `name`, `given_name`, `family_name`, `email` (only attributes in `id_token.user_attributes` that match scope claims)

#### OIDC Scope-to-Claims Mapping

OIDC scopes map to specific user claims in ID tokens. Thunder supports standard OIDC scope-to-claims mappings:

| Scope | Claims Included By Default |
|-------|----------------|
| `openid` | `sub` (subject/user ID) |
| `profile` | `name`, `given_name`, `family_name`, `middle_name`, `nickname`, `preferred_username`, `profile`, `picture`, `website`, `gender`, `birthdate`, `zoneinfo`, `locale`, `updated_at` |
| `email` | `email`, `email_verified` |
| `phone` | `phone_number`, `phone_number_verified` |
| `address` | `address` |

**Note:** The `groups` claim can be included in ID tokens through custom scope claims mapping in the ID token configuration, but `group` is not a standard OIDC scope.

You can customize the scope-to-claims mapping per application in the ID token configuration using `scope_claims`.

**ID Token Claim Filtering Logic:**
1. For each requested OIDC scope, Thunder determines which claims should be included based on:
   - App-specific `scope_claims` mapping (if configured), or
   - Standard OIDC scope-to-claims mapping
2. Only claims that are:
   - Listed in `token.id_token.user_attributes`
   - Present in the user's actual attributes
   - Associated with the requested scope
   are included in the ID token.

**Example:** If `scope=profile` is requested but `name` is not in `token.id_token.user_attributes`, the `name` claim will not be included in the ID token, even though it's part of the standard `profile` scope.

### 2. Custom Permission Scopes

Custom permission scopes are application-specific scopes that map to Thunder permissions. These scopes **are permission-based** and require evaluation through RBAC. They are included in **access tokens**.

> For detailed information on how permission scopes work, including how to set up roles, permissions, and use them with OAuth applications, see [OAuth Scopes and Permissions](../../../authorization/oauth-scopes-permissions.md).

**How Custom Permission Scopes Work:**
1. Client requests scopes: `scope=openid profile read:documents write:documents`
2. Thunder separates scopes:
   - **OIDC scopes**: `openid`, `profile` (always granted if requested, control ID token claims)
   - **Permission scopes**: `read:documents`, `write:documents` (evaluated against user permissions)
3. Permission evaluation:
   - Custom scopes are passed to the authentication process as `requested_permissions`
   - Thunder's authorization engine evaluates which permissions the user has through their role assignments
   - Only authorized permissions are returned in `authorized_permissions`
4. Token issuance:
   - **Access token** contains: OIDC scopes + authorized permission scopes in the `scope` claim
   - **ID token** contains: user attribute claims based on OIDC scopes (if `openid` scope is requested). ID tokens do not include a `scope` claim.
   - Unauthorized permission scopes are filtered out

**Example:**
- Client requests: `scope=openid profile read:documents write:documents delete:documents`
- User has permissions: `read:documents`, `write:documents` (through role assignments)
- **Access token** `scope` claim contains: `openid profile read:documents write:documents`
- `delete:documents` is filtered out (user not authorized)

**Note:** For the `client_credentials` grant type, currently all scopes requested are returned in the token without any permission evaluation.

## Examples

### Example 1: OIDC Scopes Only

**Request:**
```
scope=openid profile email
```

**Result:**
- All requested scopes are granted (no permission evaluation)
- **ID Token** includes: claims `sub`, `name`, `email`, etc. (based on `id_token.user_attributes` configuration and requested OIDC scopes). ID tokens do not include a `scope` claim.
- **Access Token** includes: `openid`, `profile`, `email` in `scope` claim, and user attributes (based on `access_token.user_attributes` configuration)

### Example 2: Mixed OIDC and Permission Scopes

**Request:**
```
scope=openid profile read:documents write:documents delete:documents
```

**User Permissions:** `read:documents`, `write:documents` (through role assignments)

**Result:**
- OIDC scopes granted: `openid`, `profile` (control ID token claims)
- Permission scopes granted: `read:documents`, `write:documents`
- Permission scope filtered: `delete:documents` (user not authorized)
- **ID Token** includes: user attribute claims based on `openid` and `profile` scopes (based on `id_token.user_attributes` configuration). ID tokens do not include a `scope` claim.
- **Access Token** `scope` claim: `openid profile read:documents write:documents` (includes OIDC scopes and authorized permission scopes)

### Example 3: Permission Scopes Only

**Request:**
```
scope=api:read api:write
```

**User Permissions:** `api:read` (through role assignments)

**Result:**
- Permission scope granted: `api:read`
- Permission scope filtered: `api:write` (user not authorized)
- **Access Token** `scope` claim: `api:read`
- **ID Token**: Not issued (requires `openid` scope)

## Access Token User Attributes

Access tokens can contain user attributes, but these are **independent of OIDC scopes**. Access token user attributes are controlled by the `token.access_token.user_attributes` configuration.

**How Access Token User Attributes Work:**
- If `token.access_token.user_attributes` is specified (non-empty array), only those attributes are included in the access token
- If `token.access_token.user_attributes` is empty or not specified, **all user attributes** are included in the access token
- OIDC scopes do not control which attributes appear in access tokens

**Example Configuration:**
```json
{
  "inbound_auth_config": [{
    "type": "oauth2",
    "config": {
      "token": {
        "access_token": {
          "user_attributes": [
            "email",
            "username"
          ]
        }
      }
    }
  }]
}
```

In this example, the access token will include only `email` and `username` user attributes, regardless of which OIDC scopes are requested.

## Related Documentation

- [Authorization Code](../grant-types/authorization-code.md) - How to request scopes
- [Role-Based Access Control](../../../authorization/role-based-access-control.md) - Setting up roles and permissions
- [OAuth Scopes and Permissions](../../../authorization/oauth-scopes-permissions.md) - Complete example with RBAC setup
