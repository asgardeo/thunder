# OAuth Scopes and Permissions

Thunder's OAuth 2.0 implementation supports custom permission scopes alongside standard OpenID Connect (OIDC) scopes. When clients request access with custom scopes, Thunder evaluates the user's permissions through their role assignments and issues access tokens containing only the authorized scopes.

This approach provides standards-compliant OAuth 2.0 authorization while maintaining fine-grained access control through Thunder's Role-Based Access Control (RBAC) system.

> Note: Refer to [Role-Based Access Control](./role-based-access-control.md) for setting up roles and permissions, and [OAuth Authentication](../authentication/standards-based/oauth-authentication.md) for basic OAuth setup.

## 🎯 How It Works

1. **Client Requests Scopes**: OAuth client initiates authorization with both OIDC and custom permission scopes
2. **User Authenticates**: User completes authentication through Thunder's authentication flow
3. **Evaluate Permissions**: Thunder checks which custom scopes the user is authorized for
4. **Issue Token**: Access token contains OIDC scopes plus authorized custom scopes

## 🚀 Complete Example: Document Management OAuth App

This example demonstrates OAuth 2.0 authorization code flow with custom permission scopes.

### Step 1: Set Up Resource Server and Permissions

Before creating roles, set up your permission model. You'll need to create a resource server and define the permissions (actions) that can be granted.

**Quick setup for this example:**

```bash
# Create document API resource server
DOCS_RS_ID=$(curl -kL -X POST -H 'Content-Type: application/json' https://localhost:8090/resource-servers \
-H 'Authorization: Bearer <token>' \
-d '{"name": "Document API", "identifier": "document-api", "ouId": "<organization-unit-id>"}' | jq -r '.id')

# Create documents resource
DOC_RESOURCE_ID=$(curl -kL -X POST https://localhost:8090/resource-servers/$DOCS_RS_ID/resources \
-H 'Authorization: Bearer <token>' -H 'Content-Type: application/json' \
-d '{"name": "Documents", "handle": "documents"}' | jq -r '.id')

# Create read action
curl -kL -X POST https://localhost:8090/resource-servers/$DOCS_RS_ID/resources/$DOC_RESOURCE_ID/actions \
-H 'Authorization: Bearer <token>' -H 'Content-Type: application/json' \
-d '{"name": "Read", "handle": "read"}'

# Create write action
curl -kL -X POST https://localhost:8090/resource-servers/$DOCS_RS_ID/resources/$DOC_RESOURCE_ID/actions \
-H 'Authorization: Bearer <token>' -H 'Content-Type: application/json' \
-d '{"name": "Write", "handle": "write"}'

# Create delete action
curl -kL -X POST https://localhost:8090/resource-servers/$DOCS_RS_ID/resources/$DOC_RESOURCE_ID/actions \
-H 'Authorization: Bearer <token>' -H 'Content-Type: application/json' \
-d '{"name": "Delete", "handle": "delete"}'
```

This creates permissions: `documents:read`, `documents:write`, `documents:delete`

> **📖 For complete details on resource servers and permission setup, see:**
>
> **[Resource Server Management →](./resource-server-management.md)**

### Step 2: Create Organization Unit

```bash
curl -kL -X POST -H 'Content-Type: application/json' https://localhost:8090/organization-units \
-H 'Authorization: Bearer <token>' \
-d '{
    "name": "OAuthDocManagement",
    "description": "OAuth document management organization",
    "handle": "oauth-docmanagement"
}'
```

Save the organization unit `id` from the response.

### Step 3: Create OAuth Application

Create an application with OAuth 2.0 inbound authentication:

```bash
curl -kL -X POST -H 'Content-Type: application/json' https://localhost:8090/applications \
-H 'Authorization: Bearer <token>' \
-d '{
    "name": "Document Manager OAuth App",
    "description": "OAuth application for document management",
    "auth_flow_id": "<auth-flow-uuid>",
    "inbound_auth_config": [
        {
            "type": "oauth2",
            "config": {
                "client_id": "doc_manager_client",
                "client_secret": "doc_manager_secret",
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
                "token_endpoint_auth_method": "client_secret_basic"
            }
        }
    ]
}'
```

### Step 4: Create User Type

Before creating users, define a user type. 

```bash
curl -kL -X POST -H 'Content-Type: application/json' https://localhost:8090/user-types \
-H 'Authorization: Bearer <token>' \
-d '{
    "name": "test-person",
    "schema": {
        "username": {
            "type": "string"
        },
        "password": {
            "type": "string"
        },
        "email": {
            "type": "string"
        },
        "given_name": {
            "type": "string"
        },
        "family_name": {
            "type": "string"
        }
    }
}'
```

> **Note:** User type creation is mandatory. You only need to create a schema once.

### Step 5: Create User with Role

Create a user and assign a role with document permissions:

```bash
# Create user
curl -kL -X POST -H 'Content-Type: application/json' https://localhost:8090/users \
-H 'Authorization: Bearer <token>' \
-d '{
    "organizationUnit": "<organization-unit-id>",
    "type": "test-person",
    "attributes": {
        "username": "bob",
        "password": "<password>",
        "email": "bob@example.com",
        "given_name": "Bob",
        "family_name": "Johnson"
    }
}'
```

Save the user `id`, then create a role:

```bash
# Create role with permissions and assign to user
curl -kL -X POST -H 'Content-Type: application/json' https://localhost:8090/roles \
-H 'Authorization: Bearer <token>' \
-d '{
    "name": "DocumentEditor",
    "description": "Can read and write documents",
    "ouId": "<organization-unit-id>",
    "permissions": [
        {
            "resourceServerId": "<document-api-resource-server-id>",
            "permissions": [
                "read:documents",
                "write:documents"
            ]
        }
    ],
    "assignments": [
        {
            "id": "<user-id>",
            "type": "user"
        }
    ]
}'
```

### Step 6: Initiate OAuth Authorization

Open a browser and start the OAuth authorization code flow with custom scopes:

```bash
https://localhost:8090/oauth2/authorize?
client_id=doc_manager_client&
redirect_uri=https://localhost:3000/callback&
response_type=code&
scope=openid%20profile%20email%20read:documents%20write:documents%20delete:documents&
state=random_state_value
```

**Query Parameters:**
- `client_id` - OAuth client identifier
- `redirect_uri` - Where to redirect after authentication
- `response_type` - Use `code` for authorization code flow
- `scope` - Space-separated list of OIDC and custom scopes
- `state` - Random value for CSRF protection

**Response:**

The server will redirect to the gate client for authentication. Follow the authentication flow (typically username/password).

### Step 7: Complete Authentication

After successful authentication, Thunder redirects back to your application with an authorization code:

```
https://localhost:3000/callback?code=<auth-code>&state=random_state_value
```

### Step 8: Exchange Code for Token

Exchange the authorization code for an access token:

```bash
curl -k -X POST https://localhost:8090/oauth2/token \
  -d 'grant_type=authorization_code' \
  -d 'code=<auth-code>' \
  -d 'redirect_uri=https://localhost:3000/callback' \
  -u 'doc_manager_client:doc_manager_secret'
```

**Response:**

```json
{
    "access_token": "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9...",
    "token_type": "Bearer",
    "expires_in": 3600,
    "refresh_token": "refresh_token_here",
    "scope": "openid profile email read:documents write:documents"
}
```

**Key Points:**
- Client requested: `openid profile email read:documents write:documents delete:documents`
- User has permissions: `read:documents write:documents`
- Token contains: `openid profile email read:documents write:documents`
- `delete:documents` was filtered out (user not authorized)

### Step 9: Decode Access Token

Decode the JWT access token to verify the scopes:

**Token Claims:**

```json
{
    "sub": "user-456",
    "aud": "doc_manager_client",
    "scope": "openid profile email read:documents write:documents",
    "username": "bob",
    "email": "bob@example.com",
    "iat": 1704067200,
    "exp": 1704070800
}
```

The `scope` claim contains all authorized scopes as a space-separated string.
