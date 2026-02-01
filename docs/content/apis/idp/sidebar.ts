import type { SidebarsConfig } from "@docusaurus/plugin-content-docs";

const sidebar: SidebarsConfig = {
  apisidebar: [
    {
      type: "doc",
      id: "apis/idp/identity-provider-management-api",
    },
    {
      type: "category",
      label: "Identity Providers",
      items: [
        {
          type: "doc",
          id: "apis/idp/list-identity-providers",
          label: "List identity providers",
          className: "api-method get",
        },
        {
          type: "doc",
          id: "apis/idp/create-an-identity-provider",
          label: "Create an identity provider",
          className: "api-method post",
        },
        {
          type: "doc",
          id: "apis/idp/get-an-identity-provider-by-id",
          label: "Get an identity provider by ID",
          className: "api-method get",
        },
        {
          type: "doc",
          id: "apis/idp/update-an-identity-provider",
          label: "Update an identity provider",
          className: "api-method put",
        },
        {
          type: "doc",
          id: "apis/idp/delete-an-identity-provider",
          label: "Delete an identity provider",
          className: "api-method delete",
        },
      ],
    },
  ],
};

export default sidebar.apisidebar;
