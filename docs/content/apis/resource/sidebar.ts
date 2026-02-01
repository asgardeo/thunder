import type { SidebarsConfig } from "@docusaurus/plugin-content-docs";

const sidebar: SidebarsConfig = {
  apisidebar: [
    {
      type: "doc",
      id: "apis/resource/resource-server-management-api",
    },
    {
      type: "category",
      label: "resource-servers",
      items: [
        {
          type: "doc",
          id: "apis/resource/list-resource-servers",
          label: "List resource servers",
          className: "api-method get",
        },
        {
          type: "doc",
          id: "apis/resource/create-a-new-resource-server",
          label: "Create a new resource server",
          className: "api-method post",
        },
        {
          type: "doc",
          id: "apis/resource/get-resource-server-details",
          label: "Get resource server details",
          className: "api-method get",
        },
        {
          type: "doc",
          id: "apis/resource/update-resource-server",
          label: "Update resource server",
          className: "api-method put",
        },
        {
          type: "doc",
          id: "apis/resource/delete-a-resource-server-by-id",
          label: "Delete a resource server by id",
          className: "api-method delete",
        },
      ],
    },
    {
      type: "category",
      label: "resources",
      items: [
        {
          type: "doc",
          id: "apis/resource/list-resources-in-resource-server",
          label: "List resources in resource server",
          className: "api-method get",
        },
        {
          type: "doc",
          id: "apis/resource/create-a-new-resource",
          label: "Create a new resource",
          className: "api-method post",
        },
        {
          type: "doc",
          id: "apis/resource/get-resource-details",
          label: "Get resource details",
          className: "api-method get",
        },
        {
          type: "doc",
          id: "apis/resource/update-resource",
          label: "Update resource",
          className: "api-method put",
        },
        {
          type: "doc",
          id: "apis/resource/delete-a-resource-by-id",
          label: "Delete a resource by id",
          className: "api-method delete",
        },
      ],
    },
    {
      type: "category",
      label: "actions",
      items: [
        {
          type: "doc",
          id: "apis/resource/list-actions-at-resource-server-level",
          label: "List actions at resource server level",
          className: "api-method get",
        },
        {
          type: "doc",
          id: "apis/resource/create-an-action-at-resource-server-level",
          label: "Create an action at resource server level",
          className: "api-method post",
        },
        {
          type: "doc",
          id: "apis/resource/get-action-details-at-resource-server-level",
          label: "Get action details at resource server level",
          className: "api-method get",
        },
        {
          type: "doc",
          id: "apis/resource/update-action-at-resource-server-level",
          label: "Update action at resource server level",
          className: "api-method put",
        },
        {
          type: "doc",
          id: "apis/resource/delete-an-action-at-resource-server-level",
          label: "Delete an action at resource server level",
          className: "api-method delete",
        },
        {
          type: "doc",
          id: "apis/resource/list-actions-at-resource-level",
          label: "List actions at resource level",
          className: "api-method get",
        },
        {
          type: "doc",
          id: "apis/resource/create-an-action-at-resource-level",
          label: "Create an action at resource level",
          className: "api-method post",
        },
        {
          type: "doc",
          id: "apis/resource/get-action-details-at-resource-level",
          label: "Get action details at resource level",
          className: "api-method get",
        },
        {
          type: "doc",
          id: "apis/resource/update-action-at-resource-level",
          label: "Update action at resource level",
          className: "api-method put",
        },
        {
          type: "doc",
          id: "apis/resource/delete-an-action-at-resource-level",
          label: "Delete an action at resource level",
          className: "api-method delete",
        },
      ],
    },
  ],
};

export default sidebar.apisidebar;
