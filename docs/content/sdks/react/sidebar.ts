import type { SidebarsConfig } from "@docusaurus/plugin-content-docs";

const sidebar: SidebarsConfig = {
  reactSdkSidebar: [
    {
      type: "doc",
      id: "sdks/react/overview",
      label: "Overview",
    },
    {
      type: "category",
      label: "APIs",
      collapsed: false,
      items: [
        {
          type: "category",
          label: "Contexts",
          collapsed: false,
          items: [
            {
              type: "doc",
              id: "sdks/react/contexts/asgardeo-provider",
              label: "<AsgardeoProvider />",
            },
          ],
        },
        {
          type: "category",
          label: "Components",
          collapsed: false,
          items: [
            {
              type: "category",
              label: "Action Components",
              collapsed: false,
              items: [
                {
                  type: "doc",
                  id: "sdks/react/components/sign-in-button",
                  label: "<SignInButton />",
                },
                {
                  type: "doc",
                  id: "sdks/react/components/sign-out-button",
                  label: "<SignOutButton />",
                },
                {
                  type: "doc",
                  id: "sdks/react/components/sign-up-button",
                  label: "<SignUpButton />",
                },
              ],
            },
            {
              type: "category",
              label: "Control Components",
              collapsed: false,
              items: [
                {
                  type: "doc",
                  id: "sdks/react/components/signed-in",
                  label: "<SignedIn />",
                },
                {
                  type: "doc",
                  id: "sdks/react/components/signed-out",
                  label: "<SignedOut />",
                },
                {
                  type: "doc",
                  id: "sdks/react/components/loading",
                  label: "<Loading />",
                },
              ],
            },
            {
              type: "category",
              label: "User Self-care Components",
              collapsed: false,
              items: [
                {
                  type: "doc",
                  id: "sdks/react/components/user-dropdown",
                  label: "<UserDropdown />",
                },
                {
                  type: "doc",
                  id: "sdks/react/components/user-profile",
                  label: "<UserProfile />",
                },
                {
                  type: "doc",
                  id: "sdks/react/components/user",
                  label: "<User />",
                },
              ],
            },
            {
              type: "category",
              label: "Organization Components (B2B)",
              collapsed: false,
              items: [
                {
                  type: "doc",
                  id: "sdks/react/components/create-organization",
                  label: "<CreateOrganization />",
                },
                {
                  type: "doc",
                  id: "sdks/react/components/organization-profile",
                  label: "<OrganizationProfile />",
                },
                {
                  type: "doc",
                  id: "sdks/react/components/organization-switcher",
                  label: "<OrganizationSwitcher />",
                },
                {
                  type: "doc",
                  id: "sdks/react/components/organization-list",
                  label: "<OrganizationList />",
                },
                {
                  type: "doc",
                  id: "sdks/react/components/organization",
                  label: "<Organization />",
                },
              ],
            },
          ],
        },
        {
          type: "category",
          label: "Hooks",
          collapsed: false,
          items: [
            {
              type: "doc",
              id: "sdks/react/hooks/use-asgardeo",
              label: "useAsgardeo()",
            },
          ],
        },
      ],
    },
  ],
};

export default sidebar.reactSdkSidebar;
