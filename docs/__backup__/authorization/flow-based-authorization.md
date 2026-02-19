# Flow-Based Authorization

Flow-based authorization integrates permission evaluation directly into Thunder's authentication flows. When users authenticate, the system evaluates which permissions they're authorized to access and includes this information in the authentication assertion (JWT).

This approach is ideal for native mobile applications and single-page applications (SPAs) that use Thunder's app-native authentication flows.

> Note: Refer to [Role-Based Access Control](./role-based-access-control.md) for setting up roles and permissions before using flow-based authorization.

## 🎯 How It Works

1. **Request Permissions**: Application initiates authentication and requests specific permissions
2. **Authenticate User**: User completes authentication flow (username/password, SMS OTP, etc.)
3. **Evaluate Authorization**: Thunder checks which requested permissions the user has through their roles
4. **Return Assertion**: Authentication completes with a JWT containing authorized permissions

## 📋 Prerequisites

Before implementing flow-based authorization, ensure you have:

1. **Organization Unit** - Created and noted the ID
2. **Application** - Configured with a flow that includes authorization (see [Flow Guides](/docs/guides/flows/))
3. **Roles and Permissions** - Defined roles with permissions
4. **Users** - Created users with role assignments

## 🚀 Complete Example: Document Management System

This example demonstrates authorization for a document management system with read and write permissions.

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
    "name": "DocManagement",
    "description": "Document management organization",
    "handle": "docmanagement"
}'
```

Save the organization unit `id` from the response.

### Step 3: Create Application

Create an application that uses the basic authentication flow (which includes authorization):

```bash
curl -kL -X POST -H 'Content-Type: application/json' https://localhost:8090/applications \
-H 'Authorization: Bearer <token>' \
-d '{
    "name": "Document Manager App",
    "description": "Application for managing documents",
    "auth_flow_id": "<auth-flow-uuid>"
}'
```

Save the application `id` from the response.

### Step 4: Create User Type

Before creating users, define a user type:

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

### Step 5: Create User

Create a user in your organization unit:

```bash
curl -kL -X POST -H 'Content-Type: application/json' https://localhost:8090/users \
-H 'Authorization: Bearer <token>' \
-d '{
    "organizationUnit": "<organization-unit-id>",
    "type": "test-person",
    "attributes": {
        "username": "alice",
        "password": "<password>",
        "email": "alice@example.com",
        "given_name": "Alice",
        "family_name": "Smith"
    }
}'
```

Save the user `id` from the response.

### Step 6: Create Role with Permissions

Create a role with document permissions and assign it to the user:

```bash
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

### Step 7: Initiate Authentication with Requested Permissions

Start the authentication flow and request specific permissions:

```bash
curl -kL -H 'Content-Type: application/json' https://localhost:8090/flow/execute \
-d '{
    "applicationId": "<application-id>",
    "flowType": "AUTHENTICATION",
    "inputs": {
        "requested_permissions": "read:documents write:documents delete:documents"
    }
}'
```

**Response:**

```json
{
    "flowId": "flow-123",
    "flowStatus": "INCOMPLETE",
    "type": "VIEW",
    "data": {
        "inputs": [
            {
                "name": "username",
                "type": "string",
                "required": true
            },
            {
                "name": "password",
                "type": "string",
                "required": true
            }
        ]
    }
}
```

Note the `flowId` from the response.

### Step 8: Complete Authentication

Complete the authentication by providing credentials:

```bash
curl -kL -H 'Content-Type: application/json' https://localhost:8090/flow/execute \
-d '{
    "flowId": "<flow-id>",
    "inputs": {
        "username": "alice",
        "password": "<password>"
    }
}'
```

**Response:**

```json
{
    "flowId": "flow-123",
    "flowStatus": "COMPLETE",
    "assertion": "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

### Step 9: Decode JWT and Extract Permissions

The JWT assertion contains the authorized permissions. Decode the JWT to extract them:

**JWT Claims:**

```json
{
    "sub": "user-123",
    "username": "alice",
    "email": "alice@example.com",
    "authorized_permissions": "read:documents write:documents",
    "iat": 1704067200,
    "exp": 1704070800
}
```

**Key Points:**
- User requested: `read:documents write:documents delete:documents`
- User has role with: `read:documents write:documents`
- User received: `read:documents write:documents` (delete was filtered out)
- Format: Space-separated string in `authorized_permissions` claim
