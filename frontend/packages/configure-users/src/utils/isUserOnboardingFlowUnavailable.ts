/**
 * Copyright (c) 2026, WSO2 LLC. (https://www.wso2.com).
 *
 * WSO2 LLC. licenses this file to you under the Apache License,
 * Version 2.0 (the "License"); you may not use this file except
 * in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */

const FLOW_NOT_FOUND_CODE = 'FLM-1003';

const getString = (value: unknown): string | null => {
  if (typeof value !== 'string') {
    return null;
  }

  const normalized = value.trim().toLowerCase();

  return normalized.length > 0 ? normalized : null;
};

export default function isUserOnboardingFlowUnavailable(value: unknown): boolean {
  if (!value || typeof value !== 'object') {
    return false;
  }

  const payload = value as {
    code?: unknown;
    message?: unknown;
    description?: unknown;
    failureReason?: unknown;
    response?: {
      data?: {
        code?: unknown;
        message?: unknown;
        description?: unknown;
        failureReason?: unknown;
      };
    };
  };

  const code = payload.code ?? payload.response?.data?.code;

  if (code === FLOW_NOT_FOUND_CODE) {
    return true;
  }

  const candidates = [
    getString(payload.message),
    getString(payload.description),
    getString(payload.failureReason),
    getString(payload.response?.data?.message),
    getString(payload.response?.data?.description),
    getString(payload.response?.data?.failureReason),
  ].filter((item): item is string => item !== null);

  return candidates.some(
    (candidate) => candidate.includes('flow') && (candidate.includes('not found') || candidate.includes('does not exist')),
  );
}
