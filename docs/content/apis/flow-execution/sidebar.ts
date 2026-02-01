import type { SidebarsConfig } from "@docusaurus/plugin-content-docs";

const sidebar: SidebarsConfig = {
  apisidebar: [
    {
      type: "doc",
      id: "apis/flow-execution/flow-execution-api-for-app-native-authentication",
    },
    {
      type: "category",
      label: "Execute a flow step",
      items: [
        {
          type: "doc",
          id: "apis/flow-execution/execute-a-flow-step",
          label: "Execute a flow step",
          className: "api-method post",
        },
      ],
    },
  ],
};

export default sidebar.apisidebar;
