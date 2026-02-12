import type { SidebarsConfig } from "@docusaurus/plugin-content-docs";

const sidebar: SidebarsConfig = {
  apisidebar: [
    {
      type: "doc",
      id: "apis/health-check/health-check-api",
    },
    {
      type: "category",
      label: "health",
      items: [
        {
          type: "doc",
          id: "apis/health-check/liveness-check",
          label: "Liveness Check",
          className: "api-method get",
        },
        {
          type: "doc",
          id: "apis/health-check/readiness-check",
          label: "Readiness check",
          className: "api-method get",
        },
      ],
    },
  ],
};

export default sidebar.apisidebar;
