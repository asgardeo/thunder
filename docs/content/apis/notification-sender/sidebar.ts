import type { SidebarsConfig } from "@docusaurus/plugin-content-docs";

const sidebar: SidebarsConfig = {
  apisidebar: [
    {
      type: "doc",
      id: "apis/notification-sender/notification-sender-api",
    },
    {
      type: "category",
      label: "Message Senders",
      items: [
        {
          type: "doc",
          id: "apis/notification-sender/list-message-notification-senders",
          label: "List message notification senders",
          className: "api-method get",
        },
        {
          type: "doc",
          id: "apis/notification-sender/create-a-new-message-notification-sender",
          label: "Create a new message notification sender",
          className: "api-method post",
        },
        {
          type: "doc",
          id: "apis/notification-sender/get-a-message-notification-sender-by-id",
          label: "Get a message notification sender by ID",
          className: "api-method get",
        },
        {
          type: "doc",
          id: "apis/notification-sender/update-a-message-notification-sender",
          label: "Update a message notification sender",
          className: "api-method put",
        },
        {
          type: "doc",
          id: "apis/notification-sender/delete-a-message-notification-sender",
          label: "Delete a message notification sender",
          className: "api-method delete",
        },
      ],
    },
    {
      type: "category",
      label: "One Time Password (OTP)",
      items: [
        {
          type: "doc",
          id: "apis/notification-sender/send-a-one-time-password-otp",
          label: "Send a One Time Password (OTP)",
          className: "api-method post",
        },
        {
          type: "doc",
          id: "apis/notification-sender/verify-a-one-time-password-otp",
          label: "Verify a One Time Password (OTP)",
          className: "api-method post",
        },
      ],
    },
  ],
};

export default sidebar.apisidebar;
