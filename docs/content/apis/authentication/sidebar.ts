import type { SidebarsConfig } from "@docusaurus/plugin-content-docs";

const sidebar: SidebarsConfig = {
  apisidebar: [
    {
      type: "doc",
      id: "apis/authentication/authentication-api",
    },
    {
      type: "category",
      label: "Credentials",
      items: [
        {
          type: "doc",
          id: "apis/authentication/authenticate-with-credentials",
          label: "Authenticate with credentials",
          className: "api-method post",
        },
      ],
    },
    {
      type: "category",
      label: "SMS OTP",
      items: [
        {
          type: "doc",
          id: "apis/authentication/send-a-sms-otp",
          label: "Send a SMS OTP",
          className: "api-method post",
        },
        {
          type: "doc",
          id: "apis/authentication/verify-a-sms-otp",
          label: "Verify a SMS OTP",
          className: "api-method post",
        },
      ],
    },
    {
      type: "category",
      label: "Google",
      items: [
        {
          type: "doc",
          id: "apis/authentication/start-google-authentication",
          label: "Start Google authentication",
          className: "api-method post",
        },
        {
          type: "doc",
          id: "apis/authentication/finish-google-authentication",
          label: "Finish Google authentication",
          className: "api-method post",
        },
      ],
    },
    {
      type: "category",
      label: "Github",
      items: [
        {
          type: "doc",
          id: "apis/authentication/start-github-authentication",
          label: "Start Github authentication",
          className: "api-method post",
        },
        {
          type: "doc",
          id: "apis/authentication/finish-github-authentication",
          label: "Finish Github authentication",
          className: "api-method post",
        },
      ],
    },
    {
      type: "category",
      label: "Standard",
      items: [
        {
          type: "doc",
          id: "apis/authentication/start-standard-o-auth-authentication",
          label: "Start standard OAuth authentication",
          className: "api-method post",
        },
        {
          type: "doc",
          id: "apis/authentication/finish-standard-o-auth-authentication",
          label: "Finish standard OAuth authentication",
          className: "api-method post",
        },
      ],
    },
    {
      type: "category",
      label: "WebAuthn / Passkey",
      items: [
        {
          type: "doc",
          id: "apis/authentication/start-web-authn-authentication",
          label: "Start WebAuthn authentication",
          className: "api-method post",
        },
        {
          type: "doc",
          id: "apis/authentication/finish-web-authn-authentication",
          label: "Finish WebAuthn authentication",
          className: "api-method post",
        },
      ],
    },
  ],
};

export default sidebar.apisidebar;
