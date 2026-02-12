import type { SidebarsConfig } from "@docusaurus/plugin-content-docs";

const sidebar: SidebarsConfig = {
  apisidebar: [
    {
      type: "doc",
      id: "apis/user/user-management-api",
    },
    {
      type: "category",
      label: "users",
      items: [
        {
          type: "doc",
          id: "apis/user/list-users",
          label: "List users",
          className: "api-method get",
        },
        {
          type: "doc",
          id: "apis/user/create-a-new-user",
          label: "Create a new user",
          className: "api-method post",
        },
        {
          type: "doc",
          id: "apis/user/get-a-user-by-id",
          label: "Get a user by id",
          className: "api-method get",
        },
        {
          type: "doc",
          id: "apis/user/update-a-user-by-id",
          label: "Update a user by id",
          className: "api-method put",
        },
        {
          type: "doc",
          id: "apis/user/delete-a-user-by-id",
          label: "Delete a user by id",
          className: "api-method delete",
        },
        {
          type: "doc",
          id: "apis/user/list-groups-that-the-user-belongs-to",
          label: "List groups that the user belongs to",
          className: "api-method get",
        },
      ],
    },
    {
      type: "category",
      label: "users-by-path",
      items: [
        {
          type: "doc",
          id: "apis/user/list-users-in-organization-unit-specified-by-handle-path",
          label: "List users in organization unit specified by handle path",
          className: "api-method get",
        },
        {
          type: "doc",
          id: "apis/user/create-a-new-user-under-the-organization-unit-specified-by-the-handle-path",
          label: "Create a new user under the organization unit specified by the handle path",
          className: "api-method post",
        },
      ],
    },
    {
      type: "category",
      label: "self",
      items: [
        {
          type: "doc",
          id: "apis/user/get-self-user-profile",
          label: "Get self user profile",
          className: "api-method get",
        },
        {
          type: "doc",
          id: "apis/user/update-self-user-profile",
          label: "Update self user profile",
          className: "api-method put",
        },
        {
          type: "doc",
          id: "apis/user/update-self-user-credentials",
          label: "Update self user credentials",
          className: "api-method post",
        },
      ],
    },
    {
      type: "category",
      label: "user-schemas",
      items: [
        {
          type: "doc",
          id: "apis/user/list-user-type-schemas",
          label: "List user type schemas",
          className: "api-method get",
        },
        {
          type: "doc",
          id: "apis/user/create-a-new-user-type-schema",
          label: "Create a new user type schema",
          className: "api-method post",
        },
        {
          type: "doc",
          id: "apis/user/get-a-user-type-schema-by-id",
          label: "Get a user type schema by ID",
          className: "api-method get",
        },
        {
          type: "doc",
          id: "apis/user/update-a-user-type-schema-by-id",
          label: "Update a user type schema by ID",
          className: "api-method put",
        },
        {
          type: "doc",
          id: "apis/user/delete-a-user-type-schema-by-id",
          label: "Delete a user type schema by ID",
          className: "api-method delete",
        },
      ],
    },
  ],
};

export default sidebar.apisidebar;
