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
 * Validates a name for feature or package creation, throwing an error if invalid.
 *
 * @param value - The name to validate
 * @param type - The type of entity (e.g., 'Feature', 'Package')
 * @throws Error if the name is invalid
 *
 * @example
 * validateName('user-management', 'Feature');
 * // Throws if name is not valid
 *
 * @public
 */
export default function validateName(name: string, type = 'Feature'): void {
  if (!name || name.trim().length === 0) {
    throw new Error(`${type} name cannot be empty`);
  }

  const trimmed = name.trim();

  // Check for valid characters (must start with letter, end with letter/number, and contain only letters, numbers, underscores, and hyphens)
  if (!/^[a-zA-Z][a-zA-Z0-9_-]*[a-zA-Z0-9]$/.test(trimmed) && !/^[a-zA-Z]$/.test(trimmed)) {
    throw new Error(
      `${type} name must start with a letter, end with a letter or number, and contain only letters, numbers, underscores, and hyphens`,
    );
  }

  // Check length
  if (trimmed.length > 50) {
    throw new Error(`${type} name must be 50 characters or less`);
  }

  // Check for reserved words
  const reserved = ['index', 'src', 'dist', 'build', 'node_modules', 'package', 'test', '__tests__'];
  if (reserved.includes(trimmed.toLowerCase())) {
    throw new Error(`${type} name '${trimmed}' is reserved`);
  }
}
