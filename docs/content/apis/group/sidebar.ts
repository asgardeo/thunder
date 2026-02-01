import type { SidebarsConfig } from "@docusaurus/plugin-content-docs";

const sidebar: SidebarsConfig = {
  apisidebar: [
    {
      type: "doc",
      id: "apis/group/group-management-api",
    },
    {
      type: "category",
      label: "groups",
      items: [
        {
          type: "doc",
          id: "apis/group/list-groups",
          label: "List groups",
          className: "api-method get",
        },
        {
          type: "doc",
          id: "apis/group/create-a-new-group",
          label: "Create a new group",
          className: "api-method post",
        },
        {
          type: "doc",
          id: "apis/group/get-a-group-by-id",
          label: "Get a group by id",
          className: "api-method get",
        },
        {
          type: "doc",
          id: "apis/group/update-a-group-by-id",
          label: "Update a group by id",
          className: "api-method put",
        },
        {
          type: "doc",
          id: "apis/group/delete-a-group-by-id",
          label: "Delete a group by id",
          className: "api-method delete",
        },
        {
          type: "doc",
          id: "apis/group/list-members-in-group",
          label: "List members in group",
          className: "api-method get",
        },
      ],
    },
    {
      type: "category",
      label: "groups-by-path",
      items: [
        {
          type: "doc",
          id: "apis/group/list-groups-in-organization-unit-specified-by-handle-path",
          label: "List groups in organization unit specified by handle path",
          className: "api-method get",
        },
        {
          type: "doc",
          id: "apis/group/create-a-new-group-under-the-organization-unit-specified-by-the-handle-path",
          label: "Create a new group under the organization unit specified by the handle path",
          className: "api-method post",
        },
      ],
    },
  ],
};

export default sidebar.apisidebar;
