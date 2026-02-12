import type { SidebarsConfig } from "@docusaurus/plugin-content-docs";

const sidebar: SidebarsConfig = {
  apisidebar: [
    {
      type: "doc",
      id: "apis/i18n/thunder-i-18-n-api",
    },
    {
      type: "category",
      label: "languages",
      items: [
        {
          type: "doc",
          id: "apis/i18n/list-languages",
          label: "List all languages",
          className: "api-method get",
        },
      ],
    },
    {
      type: "category",
      label: "resolve",
      items: [
        {
          type: "doc",
          id: "apis/i18n/resolve-translations",
          label: "Resolve all translations for a language",
          className: "api-method get",
        },
        {
          type: "doc",
          id: "apis/i18n/resolve-translation",
          label: "Resolve a single translation",
          className: "api-method get",
        },
      ],
    },
    {
      type: "category",
      label: "management",
      items: [
        {
          type: "doc",
          id: "apis/i18n/set-translations",
          label: "Set translations for a language",
          className: "api-method post",
        },
        {
          type: "doc",
          id: "apis/i18n/delete-translations",
          label: "Clear all translations for a language",
          className: "api-method delete",
        },
        {
          type: "doc",
          id: "apis/i18n/set-translation",
          label: "Set a single translation",
          className: "api-method post",
        },
        {
          type: "doc",
          id: "apis/i18n/delete-translation",
          label: "Clear a single translation",
          className: "api-method delete",
        },
      ],
    },
  ],
};

export default sidebar.apisidebar;
