/*
 * Copyright (c) 2026, WSO2 LLC. (https://www.wso2.com).
 *
 * WSO2 LLC. licenses this file to you under the Apache License,
 * Version 2.0 (the "License"); you may not use this file except
 * in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied.  See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */

/**
 * Application Edit E2E Tests
 *
 * Covers the application edit page tabs and the delete flow.
 * A test application is created via API in beforeAll and deleted in afterAll,
 * avoiding wizard dependency for edit tests.
 *
 * Required environment variables:
 * - BASE_URL: Console base URL
 * - THUNDER_URL: Thunder API base URL (defaults to https://localhost:8090)
 * - ADMIN_USERNAME: Admin username (defaults to admin)
 * - ADMIN_PASSWORD: Admin password (defaults to admin)
 * - SAMPLE_APP_ID: Application ID used to obtain admin token
 * - DEFAULT_OU_ID: (optional) OU ID for the test application; fetched from the API when not set
 */

import { test, expect } from "../../fixtures/console";
import { TestDataFactory } from "../../utils/test-data";
import { getAdminToken } from "../../utils/authentication";

const thunderUrl = process.env.THUNDER_URL || "https://localhost:8090";
const defaultOuId = process.env.DEFAULT_OU_ID || "";

test.describe("Application Edit", () => {
  let testAppId: string;
  let testAppName: string;
  let suiteOuId: string;

  test.beforeAll(async ({ request }) => {
    if (!process.env.SAMPLE_APP_ID) throw new Error("SAMPLE_APP_ID env var is required for Application Edit tests");

    console.log("\n=== Application Edit Suite Setup ===");
    const appData = TestDataFactory.createApplication({ name: `TestApp_EDIT_${Date.now()}` });
    testAppName = appData.name;

    const adminToken = await getAdminToken(request);

    let resolvedOuId = defaultOuId;
    if (!resolvedOuId) {
      const ouResponse = await request.get(`${thunderUrl}/organization-units?limit=1`, {
        headers: { Authorization: `Bearer ${adminToken}` },
        ignoreHTTPSErrors: true,
      });
      if (ouResponse.ok()) {
        const ouData = await ouResponse.json();
        resolvedOuId = ouData.organizationUnits?.[0]?.id ?? "";
      }
    }
    if (!resolvedOuId) throw new Error("Could not determine an OU ID for Application Edit tests");
    suiteOuId = resolvedOuId;

    const createResponse = await request.post(`${thunderUrl}/applications`, {
      data: {
        name: appData.name,
        description: appData.description,
        ouId: resolvedOuId,
        inboundAuthConfig: [
          {
            type: "oauth2",
            config: { clientType: "PUBLIC", redirectUris: ["http://localhost:3000/callback"] },
          },
        ],
      },
      headers: {
        Authorization: `Bearer ${adminToken}`,
        "Content-Type": "application/json",
      },
      ignoreHTTPSErrors: true,
    });

    if (!createResponse.ok()) {
      throw new Error(`Failed to create test application: ${await createResponse.text()}`);
    }

    const created = await createResponse.json();
    testAppId = created.id;
    console.log(`Test application created: ${testAppName} (${testAppId})`);
  });

  test.afterAll(async ({ request }) => {
    console.log("\n=== Application Edit Suite Cleanup ===");
    if (!testAppId) return;

    const adminToken = await getAdminToken(request);

    const deleteResponse = await request.delete(`${thunderUrl}/applications/${testAppId}`, {
      headers: {
        Authorization: `Bearer ${adminToken}`,
      },
      ignoreHTTPSErrors: true,
    });

    if (deleteResponse.ok()) {
      console.log(`Test application deleted: ${testAppId}`);
    } else {
      console.warn(`Failed to delete test application ${testAppId}: ${await deleteResponse.text()}`);
    }
  });

  /** TC006: Edit page - General tab is default and shows Quick Copy fields */
  test("TC006: Edit page - General tab is default and shows Quick Copy fields", async ({ applicationsPage }) => {
    await test.step("Navigate to application edit page", async () => {
      console.log(`Navigating to edit page for app: ${testAppId}`);
      await applicationsPage.gotoEdit(testAppId);
      await applicationsPage.screenshot("tc006-edit-page");
    });

    await test.step("Verify General tab is selected by default", async () => {
      const generalTab = applicationsPage.page.getByRole("tab", { name: /general/i });
      await expect(generalTab).toBeVisible();
      await expect(generalTab).toHaveAttribute("aria-selected", "true");
      console.log("General tab is active by default");
    });

    await test.step("Verify Application ID field is visible in Quick Copy section", async () => {
      await expect(applicationsPage.applicationIdField).toBeVisible();
      console.log("Application ID field visible");
      await applicationsPage.screenshot("tc006-quick-copy-fields");
    });
  });

  /** TC007: Edit page - Add and save a redirect URI */
  test("TC007: Edit page - Add and save a redirect URI", async ({ applicationsPage }) => {
    const testUri = "https://example.com/callback";

    await test.step("Navigate to application edit page", async () => {
      await applicationsPage.gotoEdit(testAppId);
      await applicationsPage.screenshot("tc007-edit-page");
    });

    await test.step("Add a redirect URI", async () => {
      console.log("Adding redirect URI:", testUri);
      await applicationsPage.addRedirectUri(testUri);
      await applicationsPage.screenshot("tc007-uri-added");
    });

    await test.step("Save changes and verify no error alert", async () => {
      await applicationsPage.saveChanges();
      await expect(applicationsPage.errorAlert).not.toBeVisible();
      console.log("Saved successfully with no error");
      await applicationsPage.screenshot("tc007-saved");
    });

    await test.step("Verify URI remains after save", async () => {
      const uriInput = applicationsPage.page.locator('input[placeholder*="callback" i]').last();
      await expect(uriInput).toHaveValue(testUri);
      console.log("URI still present after save:", testUri);
    });
  });

  /** TC008: Edit page - Application URL rejects invalid URL */
  test("TC008: Edit page - Application URL rejects invalid URL", async ({ applicationsPage }) => {
    await test.step("Navigate to application edit page", async () => {
      await applicationsPage.gotoEdit(testAppId);
    });

    await test.step("Type an invalid URL into Application URL field", async () => {
      await applicationsPage.applicationUrlInput.waitFor({ state: "visible" });
      await applicationsPage.applicationUrlInput.fill("not-a-valid-url");
      await applicationsPage.applicationUrlInput.blur();
      console.log("Entered invalid URL value");
      await applicationsPage.screenshot("tc008-invalid-url");
    });

    await test.step("Verify inline validation error is shown for invalid URL", async () => {
      const validationError = applicationsPage.page
        .getByText(/please enter a valid url/i)
        .or(applicationsPage.page.getByText(/invalid url/i));
      await expect(validationError.first()).toBeVisible();
      console.log("Inline validation error shown for invalid URL — correct");
      await applicationsPage.screenshot("tc008-validation-error");
    });
  });

  /** TC009: Edit page - Flows tab renders flow selectors */
  test("TC009: Edit page - Flows tab renders flow selectors", async ({ applicationsPage }) => {
    await test.step("Navigate to application edit page", async () => {
      await applicationsPage.gotoEdit(testAppId);
    });

    await test.step("Click Flows tab", async () => {
      await applicationsPage.clickTab("Flows");
      console.log("Clicked Flows tab");
      await applicationsPage.screenshot("tc009-flows-tab");
    });

    await test.step("Verify authentication and registration flow selectors are visible", async () => {
      const authFlowSelector = applicationsPage.page
        .getByText(/authentication flow/i)
        .or(applicationsPage.page.getByLabel(/authentication flow/i));
      const registrationFlowSelector = applicationsPage.page
        .getByText(/registration flow/i)
        .or(applicationsPage.page.getByLabel(/registration flow/i));

      await expect(authFlowSelector.first()).toBeVisible();
      await expect(registrationFlowSelector.first()).toBeVisible();
      console.log("Authentication and Registration flow selectors visible");
      await applicationsPage.screenshot("tc009-flow-selectors");
    });
  });

  /** TC010: Edit page - Customization tab renders contact fields */
  test("TC010: Edit page - Customization tab renders contact fields", async ({ applicationsPage }) => {
    await test.step("Navigate to application edit page", async () => {
      await applicationsPage.gotoEdit(testAppId);
    });

    await test.step("Click Customization tab", async () => {
      await applicationsPage.clickTab("Customization");
      console.log("Clicked Customization tab");
      await applicationsPage.screenshot("tc010-customization-tab");
    });

    await test.step("Verify Terms of Service and Privacy Policy URI fields are visible", async () => {
      const tosField = applicationsPage.page
        .locator('input[placeholder*="terms" i]')
        .or(applicationsPage.page.locator('input[placeholder*="example.com/terms" i]'));
      const privacyField = applicationsPage.page
        .locator('input[placeholder*="privacy" i]')
        .or(applicationsPage.page.locator('input[placeholder*="example.com/privacy" i]'));

      await expect(tosField.first()).toBeVisible();
      await expect(privacyField.first()).toBeVisible();
      console.log("Terms of Service and Privacy Policy URI fields visible");
      await applicationsPage.screenshot("tc010-fields-visible");
    });
  });

  /** TC011: Edit page - Inline name edit persists */
  test("TC011: Edit page - Inline name edit persists", async ({ applicationsPage }) => {
    const newName = `TestApp_RENAMED_${Date.now()}`;

    await test.step("Navigate to application edit page", async () => {
      await applicationsPage.gotoEdit(testAppId);
      await applicationsPage.screenshot("tc011-edit-page");
    });

    await test.step("Click the edit icon next to the application name and rename it", async () => {
      const nameHeading = applicationsPage.page.locator("h3").filter({ hasText: testAppName }).first();
      await nameHeading.waitFor({ state: "visible" });

      // Click the Edit pencil IconButton that sits next to the h3 heading
      const editNameButton = nameHeading.locator("..").getByRole("button").first();
      await editNameButton.click();

      const nameInput = applicationsPage.page.locator(`input[value*="TestApp"]`).first();
      await nameInput.waitFor({ state: "visible" });
      await nameInput.fill(newName);
      await nameInput.press("Enter");
      console.log("Renamed application to:", newName);
      await applicationsPage.screenshot("tc011-renamed");
    });

    await test.step("Verify heading shows new name", async () => {
      const updatedHeading = applicationsPage.page.locator("h3").filter({ hasText: newName });
      await expect(updatedHeading.first()).toBeVisible();
      console.log("Heading updated to new name");
    });

    await test.step("Save changes", async () => {
      await applicationsPage.saveChanges();
      console.log("Saved name change");
      await applicationsPage.screenshot("tc011-saved");
    });

    await test.step("Reload page and verify new name persists", async () => {
      await applicationsPage.page.reload({ waitUntil: "networkidle" });
      const persistedHeading = applicationsPage.page.locator("h3").filter({ hasText: newName });
      await expect(persistedHeading.first()).toBeVisible();
      console.log("New name persists after page reload:", newName);
      await applicationsPage.screenshot("tc011-name-persisted");
      testAppName = newName;
    });
  });

  /** TC012: Delete application from Danger Zone */
  test("TC012: Delete application from Danger Zone", async ({ applicationsPage, request }) => {
    let deleteTestAppId: string | undefined;
    let deleteTestAppName: string = "";

    try {
      await test.step("Create a dedicated application for delete test", async () => {
        const appData = TestDataFactory.createApplication({ name: `TestApp_DELETE_${Date.now()}` });
        deleteTestAppName = appData.name;
        const adminToken = await getAdminToken(request);

        const createResponse = await request.post(`${thunderUrl}/applications`, {
          data: {
            name: appData.name,
            description: appData.description,
            ouId: suiteOuId,
          },
          headers: {
            Authorization: `Bearer ${adminToken}`,
            "Content-Type": "application/json",
          },
          ignoreHTTPSErrors: true,
        });

        expect(createResponse.ok()).toBeTruthy();

        const created = await createResponse.json();
        deleteTestAppId = created.id;
        if (!deleteTestAppId) throw new Error("Delete test app creation failed — cannot proceed with delete test");
        console.log(`Created delete test app: ${deleteTestAppId}`);
      });

      await test.step("Navigate to application edit page", async () => {
        await applicationsPage.gotoEdit(deleteTestAppId!);
        await applicationsPage.screenshot("tc012-before-delete");
      });

      await test.step("Click Delete Application and confirm", async () => {
        await applicationsPage.clickDeleteApplication();
        await applicationsPage.confirmDelete();
        console.log("Delete confirmed");
        deleteTestAppId = undefined;
      });

      await test.step("Verify redirected to applications list", async () => {
        await applicationsPage.page.waitForURL(/\/console\/applications$/, { timeout: 15000 });
        await applicationsPage.verifyPageLoaded();
        console.log("Redirected to applications list after delete");
        await applicationsPage.screenshot("tc012-after-delete");
      });

      await test.step("Verify deleted app no longer appears in list", async () => {
        const deletedRow = applicationsPage.page
          .locator('[data-testid="applications-list"]')
          .getByText(deleteTestAppName);
        await expect(deletedRow).not.toBeVisible();
        console.log("Deleted app not found in list — correct");
      });
    } finally {
      if (deleteTestAppId) {
        const adminToken = await getAdminToken(request).catch(() => null);
        if (adminToken) {
          await request
            .delete(`${thunderUrl}/applications/${deleteTestAppId}`, {
              headers: { Authorization: `Bearer ${adminToken}` },
              ignoreHTTPSErrors: true,
            })
            .catch(() => {});
        }
      }
    }
  });
});
