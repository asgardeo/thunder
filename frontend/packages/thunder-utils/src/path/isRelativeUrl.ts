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

/**
 * Returns `true` if the given URL is relative.
 *
 * A URL is considered relative if it is not absolute — i.e. it does not
 * start with a URL scheme (e.g. `http://`, `https://`) or `//`.
 *
 * @param url - The URL string to check.
 * @returns `true` if `url` is a relative URL, `false` otherwise.
 */
export default function isRelativeUrl(url: string): boolean {
  return !url.startsWith('//') && !/^[a-zA-Z][a-zA-Z\d+\-.]*:\/\//.test(url);
}
