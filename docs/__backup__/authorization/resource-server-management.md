# Resource Server Management

Resource servers are the foundation of Thunder's permission system. A resource server represents an application, service, or API that defines the permissions that can be granted to users through roles.

## 🎯 Overview

Before assigning permissions to users via roles, you must first define:

1. **Resource Servers** - Applications or services (e.g., "Booking API", "Payment Service")
2. **Resources** (Optional) - Hierarchical organization of resources (e.g., "reservations", "reservations:online")
3. **Actions** - Operations that can be performed (e.g., "create", "view", "update", "delete")

These components combine to create **permissions** - the actual permission strings that get assigned to roles.

## 📋 Understanding the Permission Model

Thunder's permission system uses a hierarchical structure with flexible permission derivation:

```
Resource Server: "Booking API" (delimiter: ":")
├── Action: "create" → Permission: "create"
├── Action: "list" → Permission: "list"
└── Resource: "reservations"
    ├── Permission: "reservations" (derived from resource handle)
    └── Actions:
        ├── "create" → Permission: "reservations:create"
        ├── "view" → Permission: "reservations:view"
        ├── "update" → Permission: "reservations:update"
        └── Resource: "online-booking"
            └── Actions:
                └── "create" → Permission: "reservations:online-booking:create"
```

### Key Concepts

- **Resource Server** - Container for all resources and actions
  - Has a unique ID and optional identifier
  - Defines the delimiter character (default: `:`) used in permission strings
  - Scoped to an organization unit

- **Resources** - Named entities that can be nested hierarchically
  - Each has an immutable `handle` used for permission derivation
  - Can have a parent resource (immutable after creation)
  - Permission string auto-derived from hierarchy: `parent:child`

- **Actions** - Operations that can be performed
  - Can be defined at resource server level (global actions)
  - Can be defined at resource level (scoped actions)
  - Each has an immutable `handle` used for permission derivation

- **Permissions** - Auto-generated strings from the hierarchy
  - Server-level action: `{action_handle}`
  - Resource-level action: `{resource_handle}:{action_handle}`
  - Nested resource action: `{parent}:{child}:{action_handle}`

## 🏗️ Creating Resource Servers

### Step 1: Create Organization Unit

Resource servers belong to an organization unit. Create one if you don't have it:

```bash
curl -kL -X POST -H 'Content-Type: application/json' https://localhost:8090/organization-units \
-H 'Authorization: Bearer <token>' \
-d '{
    "name": "Engineering",
    "description": "Engineering department",
    "handle": "engineering"
}'
```

Save the organization unit `id` from the response.

### Step 2: Create Resource Server

Create a resource server:

```bash
curl -kL -X POST -H 'Content-Type: application/json' https://localhost:8090/resource-servers \
-H 'Authorization: Bearer <token>' \
-d '{
    "name": "Booking System",
    "description": "Handles all booking operations",
    "identifier": "booking-system",
    "ouId": "<organization-unit-id>",
    "delimiter": ":"
}'
```

**Parameters:**
- `name` (required) - Display name of the resource server
- `description` (optional) - Description of what this resource server manages
- `identifier` (optional) - Unique identifier used in authorization requests
- `ouId` (required) - Organization unit ID
- `delimiter` (optional) - Character to separate hierarchy levels (default: `:`, immutable)

**Response:**

```json
{
    "id": "3fa85f64-5717-4562-b3fc-2c963f66afa6",
    "name": "Booking System",
    "description": "Handles all booking operations",
    "identifier": "booking-system",
    "ouId": "a839f4bd-39dc-4eaa-b5cc-210d8ecaee87",
    "delimiter": ":"
}
```

Save the resource server `id` - you'll need it for creating resources and actions.

## 🔧 Defining Permissions

You can define permissions in two ways, depending on your needs:

### Option A: Flat Permissions (Server-Level Actions)

For simple, non-hierarchical permissions, create actions directly on the resource server:

```bash
curl -kL -X POST -H 'Content-Type: application/json' https://localhost:8090/resource-servers/<resource-server-id>/actions \
-H 'Authorization: Bearer <token>' \
-d '{
    "name": "Create Reservation",
    "description": "Permission to create reservations",
    "handle": "create_reservation"
}'
```

**Response:**

```json
{
    "id": "9c6d3g0g-7e8b-4d82-b454-4d2ccg87gh5e",
    "name": "Create Reservation",
    "description": "Permission to create reservations",
    "handle": "create_reservation",
    "permission": "create_reservation"
}
```

**Resulting permission:** `create_reservation`

**When to use:** Simple APIs with flat permission structure (e.g., microservices with basic CRUD operations).

### Option B: Hierarchical Permissions (Resources + Actions)

For complex, hierarchical permissions, create resources first, then actions on those resources:

#### 1. Create a Resource

```bash
curl -kL -X POST -H 'Content-Type: application/json' https://localhost:8090/resource-servers/<resource-server-id>/resources \
-H 'Authorization: Bearer <token>' \
-d '{
    "name": "Reservations",
    "description": "Reservation management",
    "handle": "reservations",
    "parent": null
}'
```

**Parameters:**
- `name` (required) - Display name
- `handle` (required) - Immutable identifier used in permissions
- `description` (optional) - Description
- `parent` (optional) - Parent resource ID for nesting (null for top-level, immutable)

**Response:**

```json
{
    "id": "7a4b1f8e-5c69-4b60-9232-2b0aaf65ef3c",
    "name": "Reservations",
    "description": "Reservation management",
    "handle": "reservations",
    "parent": null,
    "permission": "reservations"
}
```

#### 2. Create Actions on the Resource

```bash
curl -kL -X POST -H 'Content-Type: application/json' https://localhost:8090/resource-servers/<resource-server-id>/resources/<resource-id>/actions \
-H 'Authorization: Bearer <token>' \
-d '{
    "name": "Create",
    "description": "Create a new reservation",
    "handle": "create"
}'
```

**Response:**

```json
{
    "id": "be8f5i2i-9g0d-5fa4-d676-6f4eeg09ij7g",
    "name": "Create",
    "description": "Create a new reservation",
    "handle": "create",
    "permission": "reservations:create"
}
```

**Resulting permission:** `reservations:create`

#### 3. Create Nested Resources (Optional)

You can nest resources for deeper hierarchies:

```bash
curl -kL -X POST -H 'Content-Type: application/json' https://localhost:8090/resource-servers/<resource-server-id>/resources \
-H 'Authorization: Bearer <token>' \
-d '{
    "name": "Online Booking",
    "description": "Online booking subsystem",
    "handle": "online-booking",
    "parent": "<reservations-resource-id>"
}'
```

Then create actions on the nested resource:

```bash
curl -kL -X POST -H 'Content-Type: application/json' https://localhost:8090/resource-servers/<resource-server-id>/resources/<online-booking-resource-id>/actions \
-H 'Authorization: Bearer <token>' \
-d '{
    "name": "Create",
    "handle": "create"
}'
```

**Resulting permission:** `reservations:online-booking:create`

**When to use:** Complex applications with nested resources and fine-grained access control.

## 📚 Complete Examples

### Example 1: Simple Microservice (Flat Structure)

Setting up permissions for a simple payment microservice:

```bash
# Create resource server
PAYMENT_RS_ID=$(curl -kL -X POST -H 'Content-Type: application/json' https://localhost:8090/resource-servers \
-H 'Authorization: Bearer <token>' \
-d '{
    "name": "Payment Service",
    "description": "Payment processing microservice",
    "identifier": "payment-service",
    "ouId": "<organization-unit-id>"
}' | jq -r '.id')

# Create actions at resource server level
curl -kL -X POST https://localhost:8090/resource-servers/$PAYMENT_RS_ID/actions \
-H 'Authorization: Bearer <token>' \
-H 'Content-Type: application/json' \
-d '{"name": "Process Payment", "handle": "process_payment"}'

curl -kL -X POST https://localhost:8090/resource-servers/$PAYMENT_RS_ID/actions \
-H 'Authorization: Bearer <token>' \
-H 'Content-Type: application/json' \
-d '{"name": "View Payment", "handle": "view_payment"}'

curl -kL -X POST https://localhost:8090/resource-servers/$PAYMENT_RS_ID/actions \
-H 'Authorization: Bearer <token>' \
-H 'Content-Type: application/json' \
-d '{"name": "Refund Payment", "handle": "refund_payment"}'
```

**Resulting permissions:**
- `process_payment`
- `view_payment`
- `refund_payment`

### Example 2: Complex Application (Hierarchical Structure)

Setting up permissions for a hotel management system:

```bash
# Create resource server
HOTEL_RS_ID=$(curl -kL -X POST -H 'Content-Type: application/json' https://localhost:8090/resource-servers \
-H 'Authorization: Bearer <token>' \
-d '{
    "name": "Hotel Management API",
    "description": "Complete hotel management system",
    "identifier": "hotel-api",
    "ouId": "<organization-unit-id>"
}' | jq -r '.id')

# Create top-level resources
RESERVATIONS_ID=$(curl -kL -X POST https://localhost:8090/resource-servers/$HOTEL_RS_ID/resources \
-H 'Authorization: Bearer <token>' \
-H 'Content-Type: application/json' \
-d '{"name": "Reservations", "handle": "reservations"}' | jq -r '.id')

GUESTS_ID=$(curl -kL -X POST https://localhost:8090/resource-servers/$HOTEL_RS_ID/resources \
-H 'Authorization: Bearer <token>' \
-H 'Content-Type: application/json' \
-d '{"name": "Guests", "handle": "guests"}' | jq -r '.id')

# Create actions on reservations
for action in create view update cancel check-in check-out; do
  curl -kL -X POST https://localhost:8090/resource-servers/$HOTEL_RS_ID/resources/$RESERVATIONS_ID/actions \
  -H 'Authorization: Bearer <token>' \
  -H 'Content-Type: application/json' \
  -d "{\"name\": \"${action^}\", \"handle\": \"$action\"}"
done

# Create actions on guests
for action in create view update delete; do
  curl -kL -X POST https://localhost:8090/resource-servers/$HOTEL_RS_ID/resources/$GUESTS_ID/actions \
  -H 'Authorization: Bearer <token>' \
  -H 'Content-Type: application/json' \
  -d "{\"name\": \"${action^}\", \"handle\": \"$action\"}"
done

# Create nested resource under reservations
ONLINE_BOOKING_ID=$(curl -kL -X POST https://localhost:8090/resource-servers/$HOTEL_RS_ID/resources \
-H 'Authorization: Bearer <token>' \
-H 'Content-Type: application/json' \
-d "{\"name\": \"Online Booking\", \"handle\": \"online-booking\", \"parent\": \"$RESERVATIONS_ID\"}" | jq -r '.id')

# Create action on nested resource
curl -kL -X POST https://localhost:8090/resource-servers/$HOTEL_RS_ID/resources/$ONLINE_BOOKING_ID/actions \
-H 'Authorization: Bearer <token>' \
-H 'Content-Type: application/json' \
-d '{"name": "Create", "handle": "create"}'
```

**Resulting permissions:**
- `reservations:create`
- `reservations:view`
- `reservations:update`
- `reservations:cancel`
- `reservations:check-in`
- `reservations:check-out`
- `reservations:online-booking:create`
- `guests:create`
- `guests:view`
- `guests:update`
- `guests:delete`

## 📋 Managing Resource Servers

### List Resource Servers

```bash
curl -kL -H 'Accept: application/json' -H 'Authorization: Bearer <token>' \
https://localhost:8090/resource-servers?limit=10&offset=0
```

### Get Resource Server Details

```bash
curl -kL -H 'Accept: application/json' -H 'Authorization: Bearer <token>' \
https://localhost:8090/resource-servers/<resource-server-id>
```

### Update Resource Server

Update name, description, identifier, or organization unit:

```bash
curl -kL -X PUT -H 'Content-Type: application/json' https://localhost:8090/resource-servers/<resource-server-id> \
-H 'Authorization: Bearer <token>' \
-d '{
    "name": "Hotel Management API v2",
    "description": "Updated hotel management system",
    "identifier": "hotel-api-v2",
    "ouId": "<organization-unit-id>"
}'
```

**Note:** The delimiter is immutable and cannot be changed after creation.

### Delete Resource Server

```bash
curl -kL -X DELETE -H 'Authorization: Bearer <token>' \
https://localhost:8090/resource-servers/<resource-server-id>
```

**Important:** Cannot delete a resource server that has resources or actions. Delete all child resources and actions first.

## 📋 Managing Resources

### List Resources

List top-level resources (no parent):

```bash
curl -kL -H 'Accept: application/json' -H 'Authorization: Bearer <token>' \
https://localhost:8090/resource-servers/<resource-server-id>/resources
```

List child resources of a parent:

```bash
curl -kL -H 'Accept: application/json' -H 'Authorization: Bearer <token>' \
https://localhost:8090/resource-servers/<resource-server-id>/resources?parentId=<parent-resource-id>
```

### Update Resource

Update name and description (handle and parent are immutable):

```bash
curl -kL -X PUT -H 'Content-Type: application/json' \
https://localhost:8090/resource-servers/<resource-server-id>/resources/<resource-id> \
-H 'Authorization: Bearer <token>' \
-d '{
    "name": "Reservations Updated",
    "description": "Updated description"
}'
```

### Delete Resource

```bash
curl -kL -X DELETE -H 'Authorization: Bearer <token>' \
https://localhost:8090/resource-servers/<resource-server-id>/resources/<resource-id>
```

**Important:** Cannot delete a resource that has sub-resources or actions.

## 📋 Managing Actions

### List Actions

List resource server level actions:

```bash
curl -kL -H 'Accept: application/json' -H 'Authorization: Bearer <token>' \
https://localhost:8090/resource-servers/<resource-server-id>/actions
```

List resource level actions:

```bash
curl -kL -H 'Accept: application/json' -H 'Authorization: Bearer <token>' \
https://localhost:8090/resource-servers/<resource-server-id>/resources/<resource-id>/actions
```

### Update Action

Update name and description (handle is immutable):

```bash
curl -kL -X PUT -H 'Content-Type: application/json' \
https://localhost:8090/resource-servers/<resource-server-id>/actions/<action-id> \
-H 'Authorization: Bearer <token>' \
-d '{
    "name": "Create Reservation Updated",
    "description": "Updated description"
}'
```

### Delete Action

```bash
curl -kL -X DELETE -H 'Authorization: Bearer <token>' \
https://localhost:8090/resource-servers/<resource-server-id>/actions/<action-id>
```

## 💡 Best Practices

### 1. Plan Your Permission Model

Before creating resource servers, design your permission model:
- Identify your applications/services
- Decide on flat vs hierarchical structure
- Choose meaningful, consistent handle names
- Consider future extensibility

### 2. Use Consistent Naming

- **Handles**: Use lowercase, hyphenated or underscored (e.g., `check-in`, `online_booking`)
- **Names**: Use clear, descriptive names (e.g., "Check In Guest", "Online Booking")
- **Actions**: Use standard verbs (create, read, update, delete, view, list)

### 3. Choose the Right Delimiter

Common delimiters:
- `:` (default) - Most common, widely used (e.g., `reservations:create`)
- `.` - Dot notation style (e.g., `reservations.create`)
- `/` - Path-like style (e.g., `reservations/create`)

**Important:** Delimiter is immutable - choose carefully before creating the resource server.

### 4. Organize by Service Boundaries

Create separate resource servers for:
- Different microservices
- Different applications
- Different security domains

### 5. Immutability Considerations

These fields are **immutable** after creation:
- Resource server `delimiter`
- Resource `handle`
- Resource `parent`
- Action `handle`

Plan these carefully as they cannot be changed later.

### 6. Handle Conflicts

- Resource server identifiers must be globally unique
- Resource handles must be unique within their parent (or resource server if top-level)
- Action handles must be unique within their resource (or resource server)

## 🔗 Next Steps

Now that you've set up resource servers and permissions:

- [Role-Based Access Control](./role-based-access-control.md) - Create roles and assign these permissions to users
- [OAuth Scopes and Permissions](./oauth-scopes-permissions.md) - Use permissions with OAuth 2.0
- [Flow-Based Authorization](./flow-based-authorization.md) - Integrate permissions into authentication flows
