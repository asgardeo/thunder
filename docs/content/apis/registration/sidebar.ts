import type { SidebarsConfig } from "@docusaurus/plugin-content-docs";

const sidebar: SidebarsConfig = {
  apisidebar: [
    {
      type: "doc",
      id: "apis/registration/registration-api",
    },
    {
      type: "category",
      label: "Passkey / WebAuthn / FIDO2 Registration",
      items: [
        {
          type: "doc",
          id: "apis/registration/start-passkey-credential-registration",
          label: "Start Passkey credential registration",
          className: "api-method post",
        },
        {
          type: "doc",
          id: "apis/registration/finish-passkey-credential-registration",
          label: "Finish Passkey credential registration",
          className: "api-method post",
        },
      ],
    },
  ],
};

export default sidebar.apisidebar;
