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
 * API Testing Utilities
 *
 * Provides helpers for API request testing alongside UI tests.
 *
 * @example
 * import { createApiClient } from '../../utils/api-testing';
 *
 * test('verify user via API', async ({ page, request }) => {
 *   const api = createApiClient(request, baseUrl);
 *   const users = await api.get('/api/v1/users');
 *   expect(users.status).toBe(200);
 * });
 */

import { APIRequestContext, expect } from "@playwright/test";

export interface ApiResponse<T = any> {
  status: number;
  data: T;
  headers: Record<string, string>;
}

export class ApiClient {
  constructor(
    private request: APIRequestContext,
    private baseUrl: string
  ) {}

  async get<T = any>(path: string, options?: any): Promise<ApiResponse<T>> {
    const response = await this.request.get(`${this.baseUrl}${path}`, options);
    return this.parseResponse<T>(response);
  }

  async post<T = any>(path: string, data?: any, options?: any): Promise<ApiResponse<T>> {
    const response = await this.request.post(`${this.baseUrl}${path}`, {
      data,
      ...options,
    });
    return this.parseResponse<T>(response);
  }

  async put<T = any>(path: string, data?: any, options?: any): Promise<ApiResponse<T>> {
    const response = await this.request.put(`${this.baseUrl}${path}`, {
      data,
      ...options,
    });
    return this.parseResponse<T>(response);
  }

  async patch<T = any>(path: string, data?: any, options?: any): Promise<ApiResponse<T>> {
    const response = await this.request.patch(`${this.baseUrl}${path}`, {
      data,
      ...options,
    });
    return this.parseResponse<T>(response);
  }

  async delete<T = any>(path: string, options?: any): Promise<ApiResponse<T>> {
    const response = await this.request.delete(`${this.baseUrl}${path}`, options);
    return this.parseResponse<T>(response);
  }

  private async parseResponse<T>(response: any): Promise<ApiResponse<T>> {
    const headers: Record<string, string> = {};
    response.headers().forEach((value: string, key: string) => {
      headers[key] = value;
    });

    return {
      status: response.status(),
      data: await response.json().catch(() => null),
      headers,
    };
  }

  /**
   * Assert common response patterns
   */
  async expectSuccess(response: ApiResponse): Promise<void> {
    expect(response.status).toBeGreaterThanOrEqual(200);
    expect(response.status).toBeLessThan(300);
  }

  async expectError(response: ApiResponse, expectedStatus?: number): Promise<void> {
    if (expectedStatus) {
      expect(response.status).toBe(expectedStatus);
    } else {
      expect(response.status).toBeGreaterThanOrEqual(400);
    }
  }
}

export function createApiClient(request: APIRequestContext, baseUrl: string): ApiClient {
  return new ApiClient(request, baseUrl);
}
