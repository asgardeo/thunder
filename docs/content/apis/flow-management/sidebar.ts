import type { SidebarsConfig } from "@docusaurus/plugin-content-docs";

const sidebar: SidebarsConfig = {
  apisidebar: [
    {
      type: "doc",
      id: "apis/flow-management/flow-management-api",
    },
    {
      type: "category",
      label: "Flow Management",
      items: [
        {
          type: "doc",
          id: "apis/flow-management/list-flows",
          label: "List all flows",
          className: "api-method get",
        },
        {
          type: "doc",
          id: "apis/flow-management/create-flow",
          label: "Create a new flow",
          className: "api-method post",
        },
        {
          type: "doc",
          id: "apis/flow-management/get-flow",
          label: "Get flow by ID",
          className: "api-method get",
        },
        {
          type: "doc",
          id: "apis/flow-management/update-flow",
          label: "Update flow",
          className: "api-method put",
        },
        {
          type: "doc",
          id: "apis/flow-management/delete-flow",
          label: "Delete flow",
          className: "api-method delete",
        },
      ],
    },
    {
      type: "category",
      label: "Flow Versioning",
      items: [
        {
          type: "doc",
          id: "apis/flow-management/get-flow-versions",
          label: "Get flow version history",
          className: "api-method get",
        },
        {
          type: "doc",
          id: "apis/flow-management/get-flow-version",
          label: "Get specific flow version",
          className: "api-method get",
        },
        {
          type: "doc",
          id: "apis/flow-management/restore-flow-version",
          label: "Restore flow to a previous version",
          className: "api-method post",
        },
      ],
    },
  ],
};

export default sidebar.apisidebar;
