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

const base64UrlDecode = (base64UrlString: string): string => {
  let base64 = base64UrlString.replace(/-/g, "+").replace(/_/g, "/");

  while (base64.length % 4 !== 0) {
    base64 += "=";
  }

  return atob(base64);
};

export interface DecodedToken {
  header: Record<string, unknown>;
  payload: Record<string, unknown>;
  signature: string;
}

export const decodeJwt = (token: string): DecodedToken | null => {
  try {
    const parts = token.split(".");

    if (parts.length !== 3) {
      console.error("Invalid JWT format: token must have exactly 3 parts");
      return null;
    }

    const [header, payload, signature] = parts;

    const decodedHeader = JSON.parse(base64UrlDecode(header));
    const decodedPayload = JSON.parse(base64UrlDecode(payload));

    return {
      header: decodedHeader,
      payload: decodedPayload,
      signature,
    };
  } catch (error) {
    console.error("Failed to decode token:", error);
    return null;
  }
};
