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
 * Thunder MFA Setup Utilities
 *
 * Automated setup for MFA testing prerequisites:
 * - Admin authentication
 * - Notification sender creation
 * - MFA flow creation
 * - Test user creation
 * - Application configuration
 */

import { APIRequestContext } from "@playwright/test";

export interface SetupConfig {
  thunderUrl: string;
  mockSmsUrl: string;
  adminUsername: string;
  adminPassword: string;
  applicationId: string;
  testUser: {
    username: string;
    password: string;
    email: string;
    mobileNumber: string;
    firstName: string;
  };
}

export interface SetupResult {
  adminToken: string;
  notificationSenderId: string;
  flowId: string;
  userId: string;
  applicationId: string;
  cleanupFunctions: Array<() => Promise<void>>;
  resourcesCreated: {
    notificationSender: boolean;
    flow: boolean;
    user: boolean;
  };
}

export class ThunderMFASetup {
  constructor(
    private request: APIRequestContext,
    private config: SetupConfig
  ) {}

  /**
   * Perform complete MFA setup
   */
  async setup(): Promise<SetupResult> {
    console.log("\n=== Thunder MFA Setup Started ===");

    const cleanupFunctions: Array<() => Promise<void>> = [];
    const resourcesCreated = {
      notificationSender: false,
      flow: false,
      user: false,
    };

    try {
      // Step 1: Get admin token
      const adminToken = await this.getAdminToken();
      console.log("✓ Admin authentication successful");

      // Step 2: Create notification sender
      const notificationSenderId = await this.createOrGetNotificationSender(adminToken);
      if (notificationSenderId.startsWith("created:")) {
        const id = notificationSenderId.replace("created:", "");
        console.log(`✓ Notification sender created: ${id}`);
        cleanupFunctions.push(() => this.deleteNotificationSender(adminToken, id));
        resourcesCreated.notificationSender = true;
      } else {
        console.log(`✓ Using existing notification sender: ${notificationSenderId}`);
      }
      const senderId = notificationSenderId.replace("created:", "");

      // Step 3: Create MFA flow
      const flowId = await this.createOrGetMFAFlow(adminToken, senderId);
      if (flowId.startsWith("created:")) {
        const id = flowId.replace("created:", "");
        console.log(`✓ MFA flow created: ${id}`);
        cleanupFunctions.push(() => this.deleteFlow(adminToken, id));
        resourcesCreated.flow = true;
      } else {
        console.log(`✓ Using existing MFA flow: ${flowId}`);
      }
      const actualFlowId = flowId.replace("created:", "");

      // Step 4: Create test user
      const userResult = await this.createOrGetTestUser(adminToken);
      if (userResult.startsWith("created:")) {
        const id = userResult.replace("created:", "");
        console.log(`✓ Test user created: ${id}`);
        cleanupFunctions.push(() => this.deleteUser(adminToken, id));
        resourcesCreated.user = true;
      } else {
        console.log(`✓ Using existing test user: ${userResult}`);
      }
      const userId = userResult.replace("created:", "");

      // Step 5: Update application with MFA flow
      await this.updateApplicationFlow(adminToken, this.config.applicationId, actualFlowId);
      console.log(`✓ Application updated with MFA flow`);

      console.log("=== Thunder MFA Setup Completed ===\n");

      return {
        adminToken,
        notificationSenderId: senderId,
        flowId: actualFlowId,
        userId,
        applicationId: this.config.applicationId,
        cleanupFunctions,
        resourcesCreated,
      };
    } catch (error) {
      console.error("✗ MFA Setup failed:", error);
      // Run cleanup for any resources created before failure
      await this.cleanup(cleanupFunctions);
      throw error;
    }
  }

  /**
   * Cleanup all created resources
   */
  async cleanup(cleanupFunctions: Array<() => Promise<void>>): Promise<void> {
    console.log("\n=== Thunder MFA Cleanup Started ===");

    for (const cleanup of cleanupFunctions.reverse()) {
      try {
        await cleanup();
      } catch (error) {
        console.error("⚠️  Cleanup error (non-fatal):", error);
      }
    }

    console.log("=== Thunder MFA Cleanup Completed ===\n");
  }

  /**
   * Get admin authentication token
   */
  private async getAdminToken(): Promise<string> {
    // Step 1: Start authentication flow
    const flowResponse = await this.request.post(`${this.config.thunderUrl}/flow/execute`, {
      data: {
        applicationId: this.config.applicationId,
        flowType: "AUTHENTICATION",
      },
      ignoreHTTPSErrors: true,
    });

    if (!flowResponse.ok()) {
      throw new Error(`Failed to start authentication flow: ${await flowResponse.text()}`);
    }

    const flowData = await flowResponse.json();
    const flowId = flowData.flowId;

    // Step 2: Submit credentials
    const authResponse = await this.request.post(`${this.config.thunderUrl}/flow/execute`, {
      data: {
        flowId,
        inputs: {
          username: this.config.adminUsername,
          password: this.config.adminPassword,
          requested_permissions: "system",
        },
        action: "action_001",
      },
      ignoreHTTPSErrors: true,
    });

    if (!authResponse.ok()) {
      throw new Error(`Admin authentication failed: ${await authResponse.text()}`);
    }

    const authData = await authResponse.json();
    return authData.assertion;
  }

  /**
   * Create or get existing notification sender for SMS
   */
  private async createOrGetNotificationSender(adminToken: string): Promise<string> {
    const senderName = "E2E Mock SMS Sender";

    // Try to create the notification sender
    const response = await this.request.post(`${this.config.thunderUrl}/notification-senders/message`, {
      data: {
        name: senderName,
        description: "Mock SMS sender for e2e MFA testing",
        provider: "custom",
        properties: [
          {
            name: "url",
            value: this.config.mockSmsUrl,
            isSecret: false,
          },
          {
            name: "http_method",
            value: "POST",
            isSecret: false,
          },
          {
            name: "content_type",
            value: "JSON",
            isSecret: false,
          },
        ],
      },
      headers: {
        Authorization: `Bearer ${adminToken}`,
        "Content-Type": "application/json",
        Accept: "application/json",
      },
      ignoreHTTPSErrors: true,
    });

    if (response.ok()) {
      const data = await response.json();
      return `created:${data.id}`;
    }

    // Check if it's a duplicate error
    const errorText = await response.text();
    if (errorText.includes("MNS-1005") || errorText.includes("Duplicate sender name")) {
      const existingId = await this.getExistingNotificationSender(adminToken, senderName);
      return existingId; // Return without "created:" prefix
    }

    throw new Error(`Failed to create notification sender: ${errorText}`);
  }

  /**
   * Get existing notification sender by name
   */
  private async getExistingNotificationSender(adminToken: string, name: string): Promise<string> {
    const response = await this.request.get(`${this.config.thunderUrl}/notification-senders/message`, {
      headers: {
        Authorization: `Bearer ${adminToken}`,
      },
      ignoreHTTPSErrors: true,
    });

    if (!response.ok()) {
      throw new Error(`Failed to fetch notification senders: ${await response.text()}`);
    }

    const data = await response.json();
    const sender = data?.find((s: any) => s.name == name);

    if (!sender) {
      console.log(data);
      throw new Error(`Notification sender '${name}' exists but could not be found in the list`);
    }

    return sender.id;
  }

  /**
   * Create or get existing MFA authentication flow
   */
  private async createOrGetMFAFlow(adminToken: string, senderId: string): Promise<string> {
    const flowHandle = "e2e-mfa-auth-flow";

    const response = await this.request.post(`${this.config.thunderUrl}/flows`, {
      data: {
        handle: flowHandle,
        name: "E2E MFA Authentication Flow",
        flowType: "AUTHENTICATION",
        activeVersion: 3,
        nodes: this.getMFAFlowNodes(senderId),
      },
      headers: {
        Authorization: `Bearer ${adminToken}`,
        "Content-Type": "application/json",
      },
      ignoreHTTPSErrors: true,
    });

    if (response.ok()) {
      const data = await response.json();
      return `created:${data.id}`;
    }

    // Check if it's a duplicate error
    const errorText = await response.text();
    if (errorText.includes("duplicate") || errorText.includes("already exists") || response.status() === 409) {
      const existingId = await this.getExistingFlow(adminToken, flowHandle);
      return existingId; // Return without "created:" prefix
    }

    throw new Error(`Failed to create MFA flow: ${errorText}`);
  }

  /**
   * Get existing flow by handle
   */
  private async getExistingFlow(adminToken: string, handle: string): Promise<string> {
    const response = await this.request.get(`${this.config.thunderUrl}/flows?filter=handle eq "${handle}"`, {
      headers: {
        Authorization: `Bearer ${adminToken}`,
      },
      ignoreHTTPSErrors: true,
    });

    if (!response.ok()) {
      throw new Error(`Failed to fetch flows: ${await response.text()}`);
    }

    const data = await response.json();
    const flow = data.flows?.find((f: any) => f.handle === handle);

    if (!flow) {
      throw new Error(`Flow '${handle}' exists but could not be found in the list`);
    }

    return flow.id;
  }

  /**
   * Create or get existing test user with mobile number
   */
  private async createOrGetTestUser(adminToken: string): Promise<string> {
    // Get organization unit from Person user schema
    const schemasResponse = await this.request.get(`${this.config.thunderUrl}/user-schemas`, {
      headers: {
        Authorization: `Bearer ${adminToken}`,
      },
      ignoreHTTPSErrors: true,
    });

    if (!schemasResponse.ok()) {
      throw new Error(`Failed to fetch user schemas: ${await schemasResponse.text()}`);
    }

    const schemasData = await schemasResponse.json();
    const personSchema = schemasData.schemas?.find((s: any) => s.name === "Person");

    if (!personSchema || !personSchema.ouId) {
      throw new Error("Person user schema not found or missing organization unit");
    }

    const defaultOuId = personSchema.ouId;

    // Create user
    const response = await this.request.post(`${this.config.thunderUrl}/users`, {
      data: {
        type: "Person",
        organizationUnit: defaultOuId,
        attributes: {
          username: this.config.testUser.username,
          password: this.config.testUser.password,
          firstName: this.config.testUser.firstName,
          email: this.config.testUser.email,
          mobileNumber: this.config.testUser.mobileNumber,
        },
      },
      headers: {
        Authorization: `Bearer ${adminToken}`,
        "Content-Type": "application/json",
      },
      ignoreHTTPSErrors: true,
    });

    if (response.ok()) {
      const data = await response.json();
      return `created:${data.id}`;
    }

    const errorText = await response.text();
    // User might already exist, try to get existing user
    if (response.status() === 409 || errorText.includes("already exists")) {
      const existingId = await this.getExistingUser(adminToken);
      return existingId; // Return without "created:" prefix
    }

    throw new Error(`Failed to create test user: ${errorText}`);
  }

  /**
   * Get existing user by username
   */
  private async getExistingUser(adminToken: string): Promise<string> {
    const response = await this.request.get(
      `${this.config.thunderUrl}/users?filter=username eq "${this.config.testUser.username}"`,
      {
        headers: {
          Authorization: `Bearer ${adminToken}`,
        },
        ignoreHTTPSErrors: true,
      }
    );

    if (!response.ok()) {
      throw new Error(`Failed to fetch existing user: ${await response.text()}`);
    }

    const data = await response.json();
    if (!data.users || data.users.length === 0) {
      throw new Error("User exists but could not be found");
    }

    return data.users[0].id;
  }

  /**
   * Update application with MFA flow
   */
  private async updateApplicationFlow(adminToken: string, applicationId: string, flowId: string): Promise<void> {
    // Get current application details
    const getResponse = await this.request.get(`${this.config.thunderUrl}/applications/${applicationId}`, {
      headers: {
        Authorization: `Bearer ${adminToken}`,
      },
      ignoreHTTPSErrors: true,
    });

    if (!getResponse.ok()) {
      throw new Error(`Failed to fetch application: ${await getResponse.text()}`);
    }

    const appData = await getResponse.json();

    // Update with new flow ID
    const updatedApp = {
      ...appData,
      auth_flow_id: flowId,
    };

    const updateResponse = await this.request.put(`${this.config.thunderUrl}/applications/${applicationId}`, {
      data: updatedApp,
      headers: {
        Authorization: `Bearer ${adminToken}`,
        "Content-Type": "application/json",
      },
      ignoreHTTPSErrors: true,
    });

    if (!updateResponse.ok()) {
      throw new Error(`Failed to update application: ${await updateResponse.text()}`);
    }
  }

  /**
   * Delete notification sender
   */
  private async deleteNotificationSender(adminToken: string, senderId: string): Promise<void> {
    try {
      const response = await this.request.delete(`${this.config.thunderUrl}/notification-senders/message/${senderId}`, {
        headers: {
          Authorization: `Bearer ${adminToken}`,
        },
        ignoreHTTPSErrors: true,
      });

      if (response.ok()) {
        console.log(`✓ Notification sender deleted: ${senderId}`);
      } else {
        console.log(`⚠️  Could not delete notification sender: ${await response.text()}`);
      }
    } catch (error) {
      console.log(`⚠️  Error deleting notification sender: ${error}`);
    }
  }

  /**
   * Delete flow
   */
  private async deleteFlow(adminToken: string, flowId: string): Promise<void> {
    try {
      const response = await this.request.delete(`${this.config.thunderUrl}/flows/${flowId}`, {
        headers: {
          Authorization: `Bearer ${adminToken}`,
        },
        ignoreHTTPSErrors: true,
      });

      if (response.ok()) {
        console.log(`✓ Flow deleted: ${flowId}`);
      } else {
        console.log(`⚠️  Could not delete flow: ${await response.text()}`);
      }
    } catch (error) {
      console.log(`⚠️  Error deleting flow: ${error}`);
    }
  }

  /**
   * Delete user
   */
  private async deleteUser(adminToken: string, userId: string): Promise<void> {
    try {
      const response = await this.request.delete(`${this.config.thunderUrl}/users/${userId}`, {
        headers: {
          Authorization: `Bearer ${adminToken}`,
        },
        ignoreHTTPSErrors: true,
      });

      if (response.ok()) {
        console.log(`✓ User deleted: ${userId}`);
      } else {
        console.log(`⚠️  Could not delete user: ${await response.text()}`);
      }
    } catch (error) {
      console.log(`⚠️  Error deleting user: ${error}`);
    }
  }

  /**
   * Get MFA flow node definitions
   */
  private getMFAFlowNodes(senderId: string): any[] {
    return [
      {
        id: "start",
        type: "START",
        layout: { size: { width: 101, height: 34 }, position: { x: 62, y: 87 } },
        onSuccess: "prompt_credentials",
      },
      {
        id: "prompt_credentials",
        type: "PROMPT",
        layout: { size: { width: 350, height: 560 }, position: { x: 562, y: 62 } },
        meta: {
          components: [
            {
              category: "DISPLAY",
              id: "text_001",
              label: "{{ t(signin:heading) }}",
              resourceType: "ELEMENT",
              type: "TEXT",
              variant: "HEADING_1",
            },
            {
              category: "BLOCK",
              id: "block_001",
              resourceType: "ELEMENT",
              type: "BLOCK",
              components: [
                {
                  category: "FIELD",
                  hint: "",
                  id: "input_001",
                  inputType: "text",
                  label: "{{ t(elements:fields.username.label) }}",
                  placeholder: "{{ t(elements:fields.username.placeholder) }}",
                  ref: "username",
                  required: true,
                  resourceType: "ELEMENT",
                  type: "TEXT_INPUT",
                },
                {
                  category: "FIELD",
                  hint: "",
                  id: "input_002",
                  inputType: "text",
                  label: "{{ t(elements:fields.password.label) }}",
                  placeholder: "{{ t(elements:fields.password.placeholder) }}",
                  ref: "password",
                  required: true,
                  resourceType: "ELEMENT",
                  type: "PASSWORD_INPUT",
                },
                {
                  category: "ACTION",
                  eventType: "SUBMIT",
                  id: "action_001",
                  label: "{{ t(elements:buttons.submit.text) }}",
                  resourceType: "ELEMENT",
                  type: "ACTION",
                  variant: "PRIMARY",
                },
              ],
            },
          ],
        },
        inputs: [
          { ref: "input_001", type: "TEXT_INPUT", identifier: "username", required: true },
          { ref: "input_002", type: "PASSWORD_INPUT", identifier: "password", required: true },
        ],
        actions: [{ ref: "action_001", nextNode: "basic_auth" }],
      },
      {
        id: "basic_auth",
        type: "TASK_EXECUTION",
        layout: { size: { width: 217, height: 113 }, position: { x: 1062, y: 62 } },
        inputs: [
          { ref: "input_001", type: "TEXT_INPUT", identifier: "username", required: true },
          { ref: "input_002", type: "PASSWORD_INPUT", identifier: "password", required: true },
        ],
        executor: { name: "BasicAuthExecutor" },
        onSuccess: "authorization_check",
      },
      {
        id: "authorization_check",
        type: "TASK_EXECUTION",
        layout: { size: { width: 200, height: 113 }, position: { x: 1562, y: 62 } },
        executor: { name: "AuthorizationExecutor" },
        onSuccess: "send_otp",
      },
      {
        id: "send_otp",
        type: "TASK_EXECUTION",
        layout: { size: { width: 200, height: 113 }, position: { x: 2062, y: 62 } },
        inputs: [{ ref: "otp_input_24ux", type: "OTP_INPUT", identifier: "otp", required: false }],
        properties: { senderId },
        executor: { name: "SMSOTPAuthExecutor", mode: "send" },
        onSuccess: "view_s2t2",
      },
      {
        id: "verify_otp",
        type: "TASK_EXECUTION",
        layout: { size: { width: 200, height: 113 }, position: { x: 3062, y: 62 } },
        inputs: [{ ref: "otp_input_24ux", type: "OTP_INPUT", identifier: "otp", required: false }],
        properties: { senderId },
        executor: { name: "SMSOTPAuthExecutor", mode: "verify" },
        onSuccess: "auth_assert",
      },
      {
        id: "auth_assert",
        type: "TASK_EXECUTION",
        layout: { size: { width: 244, height: 113 }, position: { x: 3562, y: 62 } },
        executor: { name: "AuthAssertExecutor" },
        onSuccess: "end",
      },
      {
        id: "end",
        type: "END",
        layout: { size: { width: 85, height: 34 }, position: { x: 4062, y: 87 } },
      },
      {
        id: "view_s2t2",
        type: "PROMPT",
        layout: { size: { width: 350, height: 522 }, position: { x: 2591, y: 37 } },
        meta: {
          components: [
            {
              category: "DISPLAY",
              id: "text_nwu6",
              label: "Verify OTP",
              resourceType: "ELEMENT",
              type: "TEXT",
              variant: "HEADING_3",
            },
            {
              category: "BLOCK",
              id: "block_gwme",
              resourceType: "ELEMENT",
              type: "BLOCK",
              components: [
                {
                  category: "FIELD",
                  hint: "",
                  id: "otp_input_24ux",
                  inputType: "text",
                  label: "Enter the code sent to your mobile",
                  placeholder: "",
                  ref: "otp",
                  required: false,
                  resourceType: "ELEMENT",
                  type: "OTP_INPUT",
                },
                {
                  category: "ACTION",
                  eventType: "TRIGGER",
                  id: "action_s76e",
                  label: "Verify",
                  resourceType: "ELEMENT",
                  type: "ACTION",
                  variant: "PRIMARY",
                },
                {
                  category: "ACTION",
                  eventType: "SUBMIT",
                  id: "resend_6o42",
                  label: "Resend OTP",
                  resourceType: "ELEMENT",
                  type: "RESEND",
                },
              ],
            },
          ],
        },
        inputs: [{ ref: "otp_input_24ux", type: "OTP_INPUT", identifier: "otp", required: false }],
        actions: [
          { ref: "action_s76e", nextNode: "verify_otp" },
          { ref: "resend_6o42", nextNode: "send_otp" },
        ],
      },
    ];
  }
}
