import type { SidebarsConfig } from "@docusaurus/plugin-content-docs";

const sidebar: SidebarsConfig = {
  apisidebar: [
    {
      type: "doc",
      id: "apis/branding/branding-management-api",
    },
    {
      type: "category",
      label: "branding",
      items: [
        {
          type: "doc",
          id: "apis/branding/list-branding-configurations",
          label: "List branding configurations",
          className: "api-method get",
        },
        {
          type: "doc",
          id: "apis/branding/create-a-branding-configuration",
          label: "Create a branding configuration",
          className: "api-method post",
        },
        {
          type: "doc",
          id: "apis/branding/get-a-branding-configuration-by-id",
          label: "Get a branding configuration by ID",
          className: "api-method get",
        },
        {
          type: "doc",
          id: "apis/branding/update-branding-configuration",
          label: "Update branding configuration",
          className: "api-method put",
        },
        {
          type: "doc",
          id: "apis/branding/delete-a-branding-configuration-by-id",
          label: "Delete a branding configuration by ID",
          className: "api-method delete",
        },
        {
          type: "doc",
          id: "apis/branding/resolve-a-branding-configuration",
          label: "Resolve a branding configuration",
          className: "api-method get",
        },
      ],
    },
  ],
};

export default sidebar.apisidebar;
