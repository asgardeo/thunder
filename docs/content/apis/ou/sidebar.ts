import type { SidebarsConfig } from "@docusaurus/plugin-content-docs";

const sidebar: SidebarsConfig = {
  apisidebar: [
    {
      type: "doc",
      id: "apis/ou/organizationunit-management-api",
    },
    {
      type: "category",
      label: "organization-units",
      items: [
        {
          type: "doc",
          id: "apis/ou/list-root-organization-units",
          label: "List root organization units",
          className: "api-method get",
        },
        {
          type: "doc",
          id: "apis/ou/create-a-new-organization-unit",
          label: "Create a new organization unit",
          className: "api-method post",
        },
        {
          type: "doc",
          id: "apis/ou/get-an-organization-unit-by-id",
          label: "Get an organization unit by id",
          className: "api-method get",
        },
        {
          type: "doc",
          id: "apis/ou/update-an-organization-unit-by-id",
          label: "Update an organization unit by id",
          className: "api-method put",
        },
        {
          type: "doc",
          id: "apis/ou/delete-an-organization-unit-by-id",
          label: "Delete an organization unit by id",
          className: "api-method delete",
        },
        {
          type: "doc",
          id: "apis/ou/list-child-organization-units",
          label: "List child organization units",
          className: "api-method get",
        },
        {
          type: "doc",
          id: "apis/ou/list-users-in-organization-unit",
          label: "List users in organization unit",
          className: "api-method get",
        },
        {
          type: "doc",
          id: "apis/ou/list-groups-in-organization-unit",
          label: "List groups in organization unit",
          className: "api-method get",
        },
      ],
    },
    {
      type: "category",
      label: "organization-units-by-path",
      items: [
        {
          type: "doc",
          id: "apis/ou/get-an-organization-unit-by-hierarchical-handle-path",
          label: "Get an organization unit by hierarchical handle path",
          className: "api-method get",
        },
        {
          type: "doc",
          id: "apis/ou/update-an-organization-unit-by-hierarchical-handle-path",
          label: "Update an organization unit by hierarchical handle path",
          className: "api-method put",
        },
        {
          type: "doc",
          id: "apis/ou/delete-an-organization-unit-by-hierarchical-handle-path",
          label: "Delete an organization unit by hierarchical handle path",
          className: "api-method delete",
        },
        {
          type: "doc",
          id: "apis/ou/create-a-new-organization-unit-under-the-specified-handle-path",
          label: "Create a new organization unit under the specified handle path",
          className: "api-method post",
        },
        {
          type: "doc",
          id: "apis/ou/list-child-organization-units-by-handle-path",
          label: "List child organization units by handle path",
          className: "api-method get",
        },
        {
          type: "doc",
          id: "apis/ou/list-users-in-organization-unit-by-handle-path",
          label: "List users in organization unit by handle path",
          className: "api-method get",
        },
        {
          type: "doc",
          id: "apis/ou/list-groups-in-organization-unit-by-handle-path",
          label: "List groups in organization unit by handle path",
          className: "api-method get",
        },
      ],
    },
  ],
};

export default sidebar.apisidebar;
