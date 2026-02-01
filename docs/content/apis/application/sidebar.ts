import type { SidebarsConfig } from "@docusaurus/plugin-content-docs";

const sidebar: SidebarsConfig = {
  apisidebar: [
    {
      type: "doc",
      id: "apis/application/application-management-api",
    },
    {
      type: "category",
      label: "applications",
      items: [
        {
          type: "doc",
          id: "apis/application/list-applications",
          label: "List applications",
          className: "api-method get",
        },
        {
          type: "doc",
          id: "apis/application/create-an-application",
          label: "Create an application",
          className: "api-method post",
        },
        {
          type: "doc",
          id: "apis/application/get-an-application-by-id",
          label: "Get an application by ID",
          className: "api-method get",
        },
        {
          type: "doc",
          id: "apis/application/update-an-application",
          label: "Update an application",
          className: "api-method put",
        },
        {
          type: "doc",
          id: "apis/application/delete-an-application",
          label: "Delete an application",
          className: "api-method delete",
        },
      ],
    },
  ],
};

export default sidebar.apisidebar;
