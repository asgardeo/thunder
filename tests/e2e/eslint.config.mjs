import thunderPlugin from "@thunder/eslint-plugin-thunder";
import playwright from "eslint-plugin-playwright";

export default [
  ...thunderPlugin.configs.base,
  {
    ignores: ["node_modules/**", "playwright-report/**", "test-results/**", "blob-report/**", "playwright/.auth/**"],
  },
  {
    files: ["**/*.ts"],
    plugins: {
      playwright: playwright,
    },
    rules: {
      "playwright/no-conditional-in-test": "error",
      "playwright/no-wait-for-timeout": "warn",
      "playwright/no-element-handle": "error",
      "playwright/no-eval": "error",
      "playwright/no-focused-test": "error",
      "playwright/no-skipped-test": "warn",
      "playwright/valid-expect": "error",
      "playwright/prefer-web-first-assertions": "error",
      "playwright/no-useless-await": "error",
      "playwright/require-top-level-describe": "error",
    },
  },
];
