# Role-Based Access Control (RBAC)

Thunder's Role-Based Access Control (RBAC) system provides flexible permission management through roles. Roles are collections of permissions that can be assigned to users or groups, enabling fine-grained access control across your applications.

## 🎯 How It Works

1. **Define Resource Servers and Permissions** - Create resource servers representing your applications/services and define the actions (permissions) available
2. **Create Roles** - Group permissions by resource server into named roles
3. **Assign Roles to Users/Groups** - Associate roles with users or groups
4. **Users Inherit Permissions** - Users automatically get all permissions from their assigned roles

## 📋 Prerequisites

Before implementing RBAC, ensure you understand:

- **Organization Units** - Containers for scoping roles and users
- **Permission Model** - Which applications/services need access control and what operations they support
- **Users and Groups** - Who will be assigned roles

## 🚀 Complete Example: Hotel Booking System

This example demonstrates setting up RBAC for a hotel booking system with front desk agents who can manage reservations and process payments.

### Step 1: Set Up Resource Servers and Permissions

Create resource servers and define the permissions that will be assigned to roles:

```bash
# Create Booking API resource server
BOOKING_RS_ID=$(curl -kL -X POST -H 'Content-Type: application/json' https://localhost:8090/resource-servers \
-H 'Authorization: Bearer <token>' \
-d '{
    "name": "Booking API",
    "description": "Hotel booking and reservation system",
    "identifier": "booking-api",
    "ouId": "<organization-unit-id>"
}' | jq -r '.id')

# Create reservations resource
RESERVATION_RES_ID=$(curl -kL -X POST https://localhost:8090/resource-servers/$BOOKING_RS_ID/resources \
-H 'Authorization: Bearer <token>' \
-H 'Content-Type: application/json' \
-d '{
    "name": "Reservations",
    "description": "Reservation management",
    "handle": "reservations"
}' | jq -r '.id')

# Create actions for reservations
curl -kL -X POST https://localhost:8090/resource-servers/$BOOKING_RS_ID/resources/$RESERVATION_RES_ID/actions \
-H 'Authorization: Bearer <token>' -H 'Content-Type: application/json' \
-d '{"name": "Create", "description": "Create new reservation", "handle": "create"}'

curl -kL -X POST https://localhost:8090/resource-servers/$BOOKING_RS_ID/resources/$RESERVATION_RES_ID/actions \
-H 'Authorization: Bearer <token>' -H 'Content-Type: application/json' \
-d '{"name": "View", "description": "View reservation details", "handle": "view"}'

curl -kL -X POST https://localhost:8090/resource-servers/$BOOKING_RS_ID/resources/$RESERVATION_RES_ID/actions \
-H 'Authorization: Bearer <token>' -H 'Content-Type: application/json' \
-d '{"name": "Update", "description": "Update reservation", "handle": "update"}'

curl -kL -X POST https://localhost:8090/resource-servers/$BOOKING_RS_ID/resources/$RESERVATION_RES_ID/actions \
-H 'Authorization: Bearer <token>' -H 'Content-Type: application/json' \
-d '{"name": "Check In", "description": "Check in guest", "handle": "check_in"}'

curl -kL -X POST https://localhost:8090/resource-servers/$BOOKING_RS_ID/resources/$RESERVATION_RES_ID/actions \
-H 'Authorization: Bearer <token>' -H 'Content-Type: application/json' \
-d '{"name": "Check Out", "description": "Check out guest", "handle": "check_out"}'

# Create Payment API resource server
PAYMENT_RS_ID=$(curl -kL -X POST -H 'Content-Type: application/json' https://localhost:8090/resource-servers \
-H 'Authorization: Bearer <token>' \
-d '{
    "name": "Payment API",
    "description": "Payment processing system",
    "identifier": "payment-api",
    "ouId": "<organization-unit-id>"
}' | jq -r '.id')

# Create payment actions at resource server level (flat permissions)
curl -kL -X POST https://localhost:8090/resource-servers/$PAYMENT_RS_ID/actions \
-H 'Authorization: Bearer <token>' -H 'Content-Type: application/json' \
-d '{"name": "Process Payment", "description": "Process customer payment", "handle": "process_payment"}'

curl -kL -X POST https://localhost:8090/resource-servers/$PAYMENT_RS_ID/actions \
-H 'Authorization: Bearer <token>' -H 'Content-Type: application/json' \
-d '{"name": "View Payment", "description": "View payment details", "handle": "view_payment"}'
```

This creates the following permissions:
- **Booking API**: `reservations:create`, `reservations:view`, `reservations:update`, `reservations:check_in`, `reservations:check_out`
- **Payment API**: `process_payment`, `view_payment`

> **📖 For complete details on resource servers, hierarchical vs. flat permissions, and best practices, see:**
>
> **[Resource Server Management →](./resource-server-management.md)**

### Step 2: Create Organization Unit

Create an organization unit to contain your users and roles:

```bash
OU_ID=$(curl -kL -X POST -H 'Content-Type: application/json' https://localhost:8090/organization-units \
-H 'Authorization: Bearer <token>' \
-d '{
    "name": "Hotel Operations",
    "description": "Hotel front desk and operations staff",
    "handle": "hotel-ops"
}' | jq -r '.id')
```

Save the organization unit `id` from the response.

### Step 3: Create User Type

Before creating users, define a user type. This is mandatory and only needs to be done once:

```bash
curl -kL -X POST -H 'Content-Type: application/json' https://localhost:8090/user-types \
-H 'Authorization: Bearer <token>' \
-d '{
    "name": "hotel-employee",
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
        },
        "employee_id": {
            "type": "string"
        }
    }
}'
```

### Step 4: Create User

Create a user in your organization unit:

```bash
USER_ID=$(curl -kL -X POST -H 'Content-Type: application/json' https://localhost:8090/users \
-H 'Authorization: Bearer <token>' \
-d '{
    "organizationUnit": "'"$OU_ID"'",
    "type": "hotel-employee",
    "attributes": {
        "username": "alice.smith",
        "password": "<password>",
        "email": "alice.smith@hotel.com",
        "given_name": "Alice",
        "family_name": "Smith",
        "employee_id": "EMP001"
    }
}' | jq -r '.id')
```

Save the user `id` from the response.

### Step 5: Create Role with Permissions

Now create a role using the permissions from Step 1 and assign it to the user from Step 4:

```bash
curl -kL -X POST -H 'Content-Type: application/json' https://localhost:8090/roles \
-H 'Authorization: Bearer <token>' \
-d '{
    "name": "FrontDeskAgent",
    "description": "Front desk agent with booking and payment permissions",
    "ouId": "'"$OU_ID"'",
    "permissions": [
        {
            "resourceServerId": "'"$BOOKING_RS_ID"'",
            "permissions": [
                "reservations:create",
                "reservations:view",
                "reservations:update",
                "reservations:check_in",
                "reservations:check_out"
            ]
        },
        {
            "resourceServerId": "'"$PAYMENT_RS_ID"'",
            "permissions": [
                "process_payment",
                "view_payment"
            ]
        }
    ],
    "assignments": [
        {
            "id": "'"$USER_ID"'",
            "type": "user"
        }
    ]
}'
```

**Response:**

```json
{
    "id": "role-abc123",
    "name": "FrontDeskAgent",
    "description": "Front desk agent with booking and payment permissions",
    "ouId": "ou-xyz789",
    "permissions": [
        {
            "resourceServerId": "rs-booking-123",
            "permissions": [
                "reservations:create",
                "reservations:view",
                "reservations:update",
                "reservations:check_in",
                "reservations:check_out"
            ]
        },
        {
            "resourceServerId": "rs-payment-456",
            "permissions": [
                "process_payment",
                "view_payment"
            ]
        }
    ],
    "assignments": [
        {
            "id": "user-alice-789",
            "type": "user",
            "display": "alice.smith"
        }
    ]
}
```

**The user `alice.smith` now has the `FrontDeskAgent` role with permissions to:**
- Create, view, update reservations
- Check in and check out guests
- Process and view payments

## 👥 Assigning Roles to Groups

Roles can also be assigned to groups, making it easier to manage permissions for multiple users.

### Step 1: Create a Group with users

Create a group in your organization unit:

```bash
GROUP_ID=$(curl -kL -X POST -H 'Content-Type: application/json' https://localhost:8090/groups \
-H 'Authorization: Bearer <token>' \
-d '{
    "name": "Front Desk Team",
    "description": "All front desk agents",
    "organizationUnitId": "'"$OU_ID"'"
    "members": [
        {
            "id": ""'"$USER_ID"'"",
            "type": "user"
        }
    ]
}' | jq -r '.id')
```

### Step 2: Assign Role to Group

Assign the role to the group (all group members inherit the role's permissions):

```bash
curl -kL -X POST -H 'Content-Type: application/json' https://localhost:8090/roles/<role-id>/assignments/add \
-H 'Authorization: Bearer <token>' \
-d '{
    "assignments": [
        {
            "id": "'"$GROUP_ID"'",
            "type": "group"
        }
    ]
}'
```

Now all users in the "Front Desk Team" group automatically have the FrontDeskAgent role's permissions.

## 📋 Managing Roles

### List All Roles

Retrieve all roles in an organization unit:

```bash
curl -kL -H 'Accept: application/json' -H 'Authorization: Bearer <token>' \
https://localhost:8090/roles?ouId=$OU_ID
```

### Get Role Details

Retrieve details of a specific role:

```bash
curl -kL -H 'Accept: application/json' -H 'Authorization: Bearer <token>' \
https://localhost:8090/roles/<role-id>
```

### Get Role Assignments

List all users and groups assigned to a role:

```bash
curl -kL -H 'Accept: application/json' -H 'Authorization: Bearer <token>' \
https://localhost:8090/roles/<role-id>/assignments
```

To include display names, add the `include=display` parameter:

```bash
curl -kL -H 'Accept: application/json' -H 'Authorization: Bearer <token>' \
'https://localhost:8090/roles/<role-id>/assignments?include=display'
```

**Response:**

```json
{
    "totalResults": 2,
    "startIndex": 1,
    "count": 2,
    "assignments": [
        {
            "id": "user-123",
            "type": "user",
            "display": "alice.smith"
        },
        {
            "id": "group-456",
            "type": "group",
            "display": "Front Desk Team"
        }
    ]
}
```

### Update Role Permissions

Update the permissions associated with a role. **Note:** The update replaces all existing permissions, so include all permissions you want to keep:

```bash
curl -kL -X PUT -H 'Content-Type: application/json' https://localhost:8090/roles/<role-id> \
-H 'Authorization: Bearer <token>' \
-d '{
    "name": "SeniorFrontDeskAgent",
    "description": "Senior front desk agent with additional permissions",
    "ouId": "'"$OU_ID"'",
    "permissions": [
        {
            "resourceServerId": "'"$BOOKING_RS_ID"'",
            "permissions": [
                "reservations:create",
                "reservations:view",
                "reservations:update",
                "reservations:cancel"
            ]
        },
        {
            "resourceServerId": "'"$PAYMENT_RS_ID"'",
            "permissions": [
                "process_payment",
                "view_payment",
                "refund_payment"
            ]
        }
    ]
}'
```

### Add Assignments to Existing Role

Add more users or groups to an existing role:

```bash
curl -kL -X POST -H 'Content-Type: application/json' https://localhost:8090/roles/<role-id>/assignments/add \
-H 'Authorization: Bearer <token>' \
-d '{
    "assignments": [
        {
            "id": "<user-id>",
            "type": "user"
        },
        {
            "id": "<group-id>",
            "type": "group"
        }
    ]
}'
```

### Remove Role Assignments

Remove users or groups from a role:

```bash
curl -kL -X POST -H 'Content-Type: application/json' https://localhost:8090/roles/<role-id>/assignments/remove \
-H 'Authorization: Bearer <token>' \
-d '{
    "assignments": [
        {
            "id": "<user-id>",
            "type": "user"
        }
    ]
}'
```

### Delete a Role

Delete a role. **Note:** All assignments must be removed first:

```bash
curl -kL -X DELETE -H 'Authorization: Bearer <token>' \
https://localhost:8090/roles/<role-id>
```

## ⚠️ Error Handling

When working with roles, you may encounter these common errors:

### ROL-1012: Invalid Permissions

This error occurs when you try to assign permissions that don't exist in the resource management system.

```json
{
    "code": "ROL-1012",
    "message": "Invalid permissions",
    "description": "One or more permissions do not exist in the resource management system"
}
```

**How to debug:**

```bash
# List all resources in a resource server
curl -kL -H 'Accept: application/json' -H 'Authorization: Bearer <token>' \
https://localhost:8090/resource-servers/$BOOKING_RS_ID/resources

# List all actions for a specific resource
curl -kL -H 'Accept: application/json' -H 'Authorization: Bearer <token>' \
https://localhost:8090/resource-servers/$BOOKING_RS_ID/resources/$RESERVATION_RES_ID/actions

# Example response shows the exact permission strings:
{
    "actions": [
        {
            "id": "action-123",
            "name": "Create",
            "handle": "create",
            "permission": "reservations:create"
        }
    ]
}
```

**Resolution:**
1. Verify the `resourceServerId` corresponds to a valid resource server
2. Use the exact `permission` value from the actions API response
3. Ensure the permission format matches: `resource:action` for hierarchical, or `action` for server-level actions

### ROL-1005: Organization Unit Not Found

This error occurs when the specified organization unit doesn't exist.

```json
{
    "code": "ROL-1005",
    "message": "Organization unit not found",
    "description": "Organization unit not found"
}
```

**Resolution:** Verify the organization unit ID exists by listing organization units.

### ROL-1004: Role Name Conflict

A role with the same name already exists in the organization unit. Role names must be unique within an organization unit.

```json
{
    "code": "ROL-1004",
    "message": "Role name conflict",
    "description": "A role with the same name exists under the same organization unit"
}
```

**Resolution:** Use a different role name or update the existing role instead.

### ROL-1006: Cannot Delete Role

Cannot delete a role that still has active assignments.

**Resolution:** Remove all user and group assignments before deleting the role:

```bash
# Get current assignments
curl -kL https://localhost:8090/roles/<role-id>/assignments

# Remove each assignment
curl -kL -X POST https://localhost:8090/roles/<role-id>/assignments/remove \
-d '{"assignments": [{"id": "<assignment-id>", "type": "user"}]}'

# Then delete the role
curl -kL -X DELETE https://localhost:8090/roles/<role-id>
```

### ROL-1007: Invalid Assignment ID

One or more user or group IDs in the assignments don't exist in the system.

**Resolution:** Verify the user/group IDs exist before adding them to a role.

## 🔗 Next Steps

Now that you've set up RBAC, learn how to use it in your applications:

- **[Flow-Based Authorization →](./flow-based-authorization.md)** - Integrate authorization into authentication flows for native mobile apps and SPAs
- **[OAuth Scopes and Permissions →](./oauth-scopes-permissions.md)** - Use permissions with OAuth 2.0 for third-party integrations
- **[Resource Server Management →](./resource-server-management.md)** - Advanced resource server configuration and best practices
