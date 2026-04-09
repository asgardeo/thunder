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
 * Wildcard Redirect URI E2E Tests
 *
 * API-only Playwright tests (no browser) verifying that wildcard redirect URI
 * support works end-to-end against a live Thunder instance.
 *
 * Acceptance criteria covered:
 *   Registration validation (POST /applications):
 *     AC-01  Wildcard in scheme               → 400 Bad Request
 *     AC-02  Wildcard in host                 → 400 Bad Request
 *     AC-03  Wildcard in query string         → 400 Bad Request
 *     AC-04  Wildcard only in path            → 201 Created, stored as-is
 *     AC-05  Regex syntax in path             → 400 Bad Request
 *     AC-13  Deeplink wildcard in path        → 201 Created
 *     AC-14  Wildcard in deeplink scheme      → 400 Bad Request
 *
 *   Authorization request matching (GET /oauth2/authorize):
 *     AC-06  * matches one segment only       → redirected with ?code=
 *     AC-07  ** matches zero / many segments  → redirected with ?code=
 *     AC-08  Exact match (no wildcard)        → redirected with ?code=
 *     AC-09  No match                         → invalid_request
 *     AC-10  Query param mismatch             → invalid_request
 *     AC-11  Multiple URIs, first match wins  → redirected with ?code=
 *     AC-12  Wildcard registered, URI omitted → invalid_request
 *
 * Prerequisites:
 *   - Thunder running at THUNDER_URL (default https://localhost:8090)
 *   - Admin credentials in ADMIN_USERNAME / ADMIN_PASSWORD
 *   - Admin application ID in ADMIN_APP_ID (used for the /flow/execute auth)
 *
 * Note: AC-15 and AC-16 (token exact-match) are covered by the existing
 * authorization_code grant handler, which has no wildcard path at all. Those
 * scenarios are better validated at the unit level; they are documented here
 * for completeness but not exercised with a full browser-based token exchange.
 */

import { test, expect, request as playwrightRequest, APIRequestContext } from "@playwright/test";
import fs from "fs";
import path from "path";

const thunderUrl = process.env.THUNDER_URL ?? "https://localhost:8090";

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

/**
 * Load the admin auth token from the stored session state.
 * Playwright's global setup saves authentication to playwright/.auth/console-admin.json
 */
function getAdminTokenFromStoredSession(): string {
  const authFilePath = path.join(__dirname, "../../playwright/.auth/console-admin.json");
  if (!fs.existsSync(authFilePath)) {
    throw new Error(`Auth state file not found: ${authFilePath}. Run global setup first.`);
  }
  
  const authState = JSON.parse(fs.readFileSync(authFilePath, "utf-8"));
  const sessionData = authState.origins?.[0]?.sessionStorage?.find(
    (item: { name: string }) => item.name === "session_data-instance_0-CONSOLE"
  );
  
  if (!sessionData) {
    throw new Error("No session data found in auth state");
  }
  
  const sessionObj = JSON.parse(sessionData.value);
  const token = sessionObj.access_token;
  
  if (!token) {
    throw new Error("No access_token found in session data");
  }
  
  return token;
}

interface CreatedApp {
  appId: string;
  clientId: string;
  clientSecret: string;
}

/**
 * Create a minimal OAuth2 application with the given redirect URIs.
 * Returns the created app's IDs for later cleanup.
 */
async function createApp(
  ctx: APIRequestContext,
  token: string,
  name: string,
  redirectUris: string[]
): Promise<CreatedApp> {
  const resp = await ctx.post(`${thunderUrl}/applications`, {
    data: {
      name,
      inboundAuthConfig: [
        {
          type: "oauth2",
          config: {
            redirectUris,
            grantTypes: ["authorization_code"],
            responseTypes: ["code"],
            tokenEndpointAuthMethod: "client_secret_basic",
            pkceRequired: false,
            publicClient: false,
          },
        },
      ],
    },
    headers: { Authorization: `Bearer ${token}`, "Content-Type": "application/json" },
    ignoreHTTPSErrors: true,
  });
  if (!resp.ok()) {
    throw new Error(`createApp failed (${resp.status()}): ${await resp.text()}`);
  }
  const body = await resp.json();
  const oauth = body.inboundAuthConfig?.[0]?.config;
  return {
    appId: body.id as string,
    clientId: oauth?.clientId as string,
    clientSecret: oauth?.clientSecret as string,
  };
}

/** Delete an application by ID. Failures are non-fatal (logged only). */
async function deleteApp(ctx: APIRequestContext, token: string, appId: string): Promise<void> {
  const resp = await ctx.delete(`${thunderUrl}/applications/${appId}`, {
    headers: { Authorization: `Bearer ${token}` },
    ignoreHTTPSErrors: true,
  });
  if (!resp.ok()) {
    console.warn(`[cleanup] DELETE /applications/${appId} returned ${resp.status()}`);
  }
}

/**
 * Issue a GET /oauth2/authorize and stop at the first redirect so we can
 * inspect the Location header before the browser would follow it.
 *
 * Thunder's behaviour:
 *   - Valid   redirect_uri  → 302 to login page   (no errorCode param)
 *   - Invalid redirect_uri  → 302 to error page   (has ?errorCode= param)
 */
async function authorizeAndGetRedirect(
  ctx: APIRequestContext,
  clientId: string,
  redirectUri: string,
  state = "test_state"
): Promise<{ location: string | null; status: number }> {
  const params = new URLSearchParams({
    client_id: clientId,
    response_type: "code",
    state,
  });
  if (redirectUri) {
    params.set("redirect_uri", redirectUri);
  }

  const resp = await ctx.get(`${thunderUrl}/oauth2/authorize?${params}`, {
    maxRedirects: 0,
    ignoreHTTPSErrors: true,
  });

  return {
    status: resp.status(),
    // location is present on 302 responses
    location: resp.headers()["location"] ?? null,
  };
}

/**
 * Returns true when the Location header indicates the server routed the
 * request to its own error page (invalid redirect URI case).
 */
function isErrorPageRedirect(location: string | null): boolean {
  if (!location) return false;
  // Thunder's error page redirect carries ?errorCode= in the URL.
  return location.includes("errorCode=");
}

/**
 * Returns true when the Location header indicates the server accepted the
 * redirect URI and routed to the login / flow page.
 */
function isLoginPageRedirect(location: string | null): boolean {
  if (!location) return false;
  return !location.includes("errorCode=");
}

// ---------------------------------------------------------------------------
// Test suites
// ---------------------------------------------------------------------------

test.describe("Wildcard Redirect URI — Registration Validation", () => {
  let ctx: APIRequestContext;
  let adminToken: string;
  const createdAppIds: string[] = [];

  test.beforeAll(async () => {
    ctx = await playwrightRequest.newContext({ ignoreHTTPSErrors: true });
    adminToken = getAdminTokenFromStoredSession();
  });

  test.afterAll(async () => {
    for (const id of createdAppIds) {
      await deleteApp(ctx, adminToken, id);
    }
    await ctx.dispose();
  });

  test("AC-01 — wildcard in scheme is rejected with 400", async () => {
    const resp = await ctx.post(`${thunderUrl}/applications`, {
      data: {
        name: "e2e-wc-ac01",
        inboundAuthConfig: [
          {
            type: "oauth2",
            config: {
              redirectUris: ["http*://example.com/callback"],
              grantTypes: ["authorization_code"],
              responseTypes: ["code"],
              tokenEndpointAuthMethod: "client_secret_basic",
            },
          },
        ],
      },
      headers: { Authorization: `Bearer ${adminToken}`, "Content-Type": "application/json" },
      ignoreHTTPSErrors: true,
    });
    expect(resp.status()).toBe(400);
  });

  test("AC-02 — wildcard in host is rejected with 400", async () => {
    const resp = await ctx.post(`${thunderUrl}/applications`, {
      data: {
        name: "e2e-wc-ac02",
        inboundAuthConfig: [
          {
            type: "oauth2",
            config: {
              redirectUris: ["https://*.example.com/callback"],
              grantTypes: ["authorization_code"],
              responseTypes: ["code"],
              tokenEndpointAuthMethod: "client_secret_basic",
            },
          },
        ],
      },
      headers: { Authorization: `Bearer ${adminToken}`, "Content-Type": "application/json" },
      ignoreHTTPSErrors: true,
    });
    expect(resp.status()).toBe(400);
  });

  test("AC-03 — wildcard in query string is rejected with 400", async () => {
    const resp = await ctx.post(`${thunderUrl}/applications`, {
      data: {
        name: "e2e-wc-ac03",
        inboundAuthConfig: [
          {
            type: "oauth2",
            config: {
              redirectUris: ["https://example.com/callback?foo=*"],
              grantTypes: ["authorization_code"],
              responseTypes: ["code"],
              tokenEndpointAuthMethod: "client_secret_basic",
            },
          },
        ],
      },
      headers: { Authorization: `Bearer ${adminToken}`, "Content-Type": "application/json" },
      ignoreHTTPSErrors: true,
    });
    expect(resp.status()).toBe(400);
  });

  test("AC-04 — wildcard only in path is accepted and stored as-is", async () => {
    const resp = await ctx.post(`${thunderUrl}/applications`, {
      data: {
        name: "e2e-wc-ac04",
        inboundAuthConfig: [
          {
            type: "oauth2",
            config: {
              redirectUris: ["https://example.com/callback/*"],
              grantTypes: ["authorization_code"],
              responseTypes: ["code"],
              tokenEndpointAuthMethod: "client_secret_basic",
            },
          },
        ],
      },
      headers: { Authorization: `Bearer ${adminToken}`, "Content-Type": "application/json" },
      ignoreHTTPSErrors: true,
    });
    expect(resp.status()).toBe(201);
    const body = await resp.json();
    createdAppIds.push(body.id);
    const storedUris: string[] = body.inboundAuthConfig?.[0]?.config?.redirectUris ?? [];
    expect(storedUris).toContain("https://example.com/callback/*");
  });

  test("AC-05 — regex syntax in path is rejected with 400", async () => {
    const resp = await ctx.post(`${thunderUrl}/applications`, {
      data: {
        name: "e2e-wc-ac05",
        inboundAuthConfig: [
          {
            type: "oauth2",
            config: {
              redirectUris: ["https://example.com/callback/[a-z]+"],
              grantTypes: ["authorization_code"],
              responseTypes: ["code"],
              tokenEndpointAuthMethod: "client_secret_basic",
            },
          },
        ],
      },
      headers: { Authorization: `Bearer ${adminToken}`, "Content-Type": "application/json" },
      ignoreHTTPSErrors: true,
    });
    expect(resp.status()).toBe(400);
  });

  test("AC-13 — deeplink with wildcard in path is accepted", async () => {
    const resp = await ctx.post(`${thunderUrl}/applications`, {
      data: {
        name: "e2e-wc-ac13",
        inboundAuthConfig: [
          {
            type: "oauth2",
            config: {
              redirectUris: ["myapp://callback/*"],
              grantTypes: ["authorization_code"],
              responseTypes: ["code"],
              tokenEndpointAuthMethod: "client_secret_basic",
            },
          },
        ],
      },
      headers: { Authorization: `Bearer ${adminToken}`, "Content-Type": "application/json" },
      ignoreHTTPSErrors: true,
    });
    expect(resp.status()).toBe(201);
    const body = await resp.json();
    createdAppIds.push(body.id);
  });

  test("AC-14 — wildcard in deeplink scheme is rejected with 400", async () => {
    const resp = await ctx.post(`${thunderUrl}/applications`, {
      data: {
        name: "e2e-wc-ac14",
        inboundAuthConfig: [
          {
            type: "oauth2",
            config: {
              redirectUris: ["my*app://callback"],
              grantTypes: ["authorization_code"],
              responseTypes: ["code"],
              tokenEndpointAuthMethod: "client_secret_basic",
            },
          },
        ],
      },
      headers: { Authorization: `Bearer ${adminToken}`, "Content-Type": "application/json" },
      ignoreHTTPSErrors: true,
    });
    expect(resp.status()).toBe(400);
  });
});

// ---------------------------------------------------------------------------

test.describe("Wildcard Redirect URI — Authorization Request Matching", () => {
  let ctx: APIRequestContext;
  let adminToken: string;

  // One app with a mix of wildcard patterns for multi-URI tests (AC-11).
  let multiApp: CreatedApp;
  // One app with a single wildcard URI for single-star/double-star/omit tests.
  let singleStarApp: CreatedApp;
  // One app with a double-star pattern.
  let doubleStarApp: CreatedApp;
  // One app with an exact URI (no wildcard) for AC-08.
  let exactApp: CreatedApp;
  // One app with query-constrained URI for AC-10.
  let queryApp: CreatedApp;

  test.beforeAll(async () => {
    ctx = await playwrightRequest.newContext({ ignoreHTTPSErrors: true });
    adminToken = getAdminTokenFromStoredSession();

    const currentTimestamp = Date.now();
    // Create apps sequentially to avoid SQLite database lock contention
    singleStarApp = await createApp(ctx, adminToken, `e2e-wc-authz-single-${currentTimestamp}`, ["https://client.example.com/cb/*"]);
    doubleStarApp = await createApp(ctx, adminToken, `e2e-wc-authz-double-${currentTimestamp}`, ["https://client.example.com/app/**/cb"]);
    exactApp = await createApp(ctx, adminToken, `e2e-wc-authz-exact-${currentTimestamp}`, ["https://client.example.com/callback"]);
    queryApp = await createApp(ctx, adminToken, `e2e-wc-authz-query-${currentTimestamp}`, ["https://client.example.com/cb?foo=bar"]);
    multiApp = await createApp(ctx, adminToken, `e2e-wc-authz-multi-${currentTimestamp}`, [
      "https://client.example.com/a/*",
      "https://client.example.com/b/*",
    ]);
  });

  test.afterAll(async () => {
    // Delete apps sequentially to avoid SQLite database lock contention
    await deleteApp(ctx, adminToken, singleStarApp.appId);
    await deleteApp(ctx, adminToken, doubleStarApp.appId);
    await deleteApp(ctx, adminToken, exactApp.appId);
    await deleteApp(ctx, adminToken, queryApp.appId);
    await deleteApp(ctx, adminToken, multiApp.appId);
    await ctx.dispose();
  });

  test("AC-06 — single * matches exactly one path segment", async () => {
    // Valid URI: Thunder redirects to login page (no errorCode in Location).
    const { status, location } = await authorizeAndGetRedirect(
      ctx,
      singleStarApp.clientId,
      "https://client.example.com/cb/v1"
    );
    expect(status).toBe(302);
    expect(isLoginPageRedirect(location)).toBe(true);
  });

  test("AC-06 — single * rejects two path segments", async () => {
    // Invalid URI: Thunder redirects to its own error page (errorCode in Location).
    const { status, location } = await authorizeAndGetRedirect(
      ctx,
      singleStarApp.clientId,
      "https://client.example.com/cb/v1/extra"
    );
    expect(status).toBe(302);
    expect(isErrorPageRedirect(location)).toBe(true);
  });

  test("AC-07 — ** matches multiple path segments", async () => {
    const { status, location } = await authorizeAndGetRedirect(
      ctx,
      doubleStarApp.clientId,
      "https://client.example.com/app/tenant/region/cb"
    );
    expect(status).toBe(302);
    expect(isLoginPageRedirect(location)).toBe(true);
  });

  test("AC-07 — ** matches zero path segments", async () => {
    const { status, location } = await authorizeAndGetRedirect(
      ctx,
      doubleStarApp.clientId,
      "https://client.example.com/app/cb"
    );
    expect(status).toBe(302);
    expect(isLoginPageRedirect(location)).toBe(true);
  });

  test("AC-08 — exact match (no wildcard) is accepted", async () => {
    const { status, location } = await authorizeAndGetRedirect(
      ctx,
      exactApp.clientId,
      "https://client.example.com/callback"
    );
    expect(status).toBe(302);
    expect(isLoginPageRedirect(location)).toBe(true);
  });

  test("AC-09 — no matching URI redirects to error page", async () => {
    const { status, location } = await authorizeAndGetRedirect(
      ctx,
      singleStarApp.clientId,
      "https://client.example.com/other"
    );
    expect(status).toBe(302);
    expect(isErrorPageRedirect(location)).toBe(true);
  });

  test("AC-10 — query param mismatch redirects to error page", async () => {
    const { status, location } = await authorizeAndGetRedirect(
      ctx,
      queryApp.clientId,
      "https://client.example.com/cb?foo=baz"
    );
    expect(status).toBe(302);
    expect(isErrorPageRedirect(location)).toBe(true);
  });

  test("AC-11 — multiple URIs registered, first match wins", async () => {
    const { status, location } = await authorizeAndGetRedirect(
      ctx,
      multiApp.clientId,
      "https://client.example.com/b/x"
    );
    expect(status).toBe(302);
    expect(isLoginPageRedirect(location)).toBe(true);
  });

  test("AC-12 — wildcard registered, redirect_uri omitted redirects to error page", async () => {
    // Empty string → omits redirect_uri from the request.
    const { status, location } = await authorizeAndGetRedirect(ctx, singleStarApp.clientId, "");
    expect(status).toBe(302);
    expect(isErrorPageRedirect(location)).toBe(true);
  });
});
