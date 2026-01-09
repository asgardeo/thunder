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
 * Test Data Generators
 *
 * Provides factories for generating test data with unique identifiers.
 * Helps avoid data conflicts in parallel test execution.
 *
 * @example
 * import { TestDataFactory } from '../../utils/test-data';
 *
 * const userData = TestDataFactory.createUser({ firstName: 'John' });
 * const appData = TestDataFactory.createApplication({ name: 'MyApp' });
 */

import { randomBytes } from "crypto";

export class TestDataFactory {
  /**
   * Generate unique identifier
   */
  static generateUniqueId(prefix: string = ""): string {
    const timestamp = Date.now();
    const random = randomBytes(4).toString("hex");
    return prefix ? `${prefix}_${timestamp}_${random}` : `${timestamp}_${random}`;
  }

  /**
   * Generate unique email
   */
  static generateEmail(username?: string): string {
    const user = username || `user_${this.generateUniqueId()}`;
    return `${user}@test.wso2.com`;
  }

  /**
   * Generate user data
   */
  static createUser(overrides?: Partial<UserData>): UserData {
    const uniqueId = this.generateUniqueId();
    return {
      username: `testuser_${uniqueId}`,
      email: this.generateEmail(`testuser_${uniqueId}`),
      firstName: `TestFirst_${uniqueId}`,
      lastName: `TestLast_${uniqueId}`,
      password: process.env.TEST_USER_PASSWORD || "TestPassword@123",
      ...overrides,
    };
  }

  /**
   * Generate application data
   */
  static createApplication(overrides?: Partial<ApplicationData>): ApplicationData {
    const uniqueId = this.generateUniqueId();
    return {
      name: `TestApp_${uniqueId}`,
      description: `Test application created at ${new Date().toISOString()}`,
      callbackUrls: ["http://localhost:3000/callback"],
      ...overrides,
    };
  }

  /**
   * Generate multiple test data items
   */
  static createBulkUsers(count: number, baseData?: Partial<UserData>): UserData[] {
    return Array.from({ length: count }, () => this.createUser(baseData));
  }
}

export interface UserData {
  username: string;
  email: string;
  firstName: string;
  lastName: string;
  password?: string;
  [key: string]: any;
}

export interface ApplicationData {
  name: string;
  description: string;
  callbackUrls: string[];
  [key: string]: any;
}
