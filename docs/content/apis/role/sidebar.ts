import type { SidebarsConfig } from "@docusaurus/plugin-content-docs";

const sidebar: SidebarsConfig = {
  apisidebar: [
    {
      type: "doc",
      id: "apis/role/role-management-api",
    },
    {
      type: "category",
      label: "roles",
      items: [
        {
          type: "doc",
          id: "apis/role/list-roles",
          label: "List roles",
          className: "api-method get",
        },
        {
          type: "doc",
          id: "apis/role/create-a-new-role",
          label: "Create a new role",
          className: "api-method post",
        },
        {
          type: "doc",
          id: "apis/role/get-role-details",
          label: "Get role details",
          className: "api-method get",
        },
        {
          type: "doc",
          id: "apis/role/update-role",
          label: "Update role",
          className: "api-method put",
        },
        {
          type: "doc",
          id: "apis/role/delete-a-role-by-id",
          label: "Delete a role by id",
          className: "api-method delete",
        },
        {
          type: "doc",
          id: "apis/role/get-role-assignments",
          label: "Get role assignments",
          className: "api-method get",
        },
        {
          type: "doc",
          id: "apis/role/add-assignments-to-a-role",
          label: "Add assignments to a role",
          className: "api-method post",
        },
        {
          type: "doc",
          id: "apis/role/remove-assignments-from-a-role",
          label: "Remove assignments from a role",
          className: "api-method post",
        },
      ],
    },
  ],
};

export default sidebar.apisidebar;
