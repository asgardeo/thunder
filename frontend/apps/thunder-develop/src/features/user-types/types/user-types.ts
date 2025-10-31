/**
 * Copyright (c) 2025, WSO2 LLC. (https://www.wso2.com).
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
 * TypeScript types and interfaces for User Types (User Schemas) feature
 * Based on the OpenAPI specification for UserSchema endpoints
 */

/**
 * Base property definition types for user schema
 */
interface BasePropertyDefinition {
  required?: boolean;
  unique?: boolean;
}

/**
 * String property definition
 */
export interface StringPropertyDefinition extends BasePropertyDefinition {
  type: 'string';
  enum?: string[];
  regex?: string;
}

/**
 * Number property definition
 */
export interface NumberPropertyDefinition extends BasePropertyDefinition {
  type: 'number';
}

/**
 * Boolean property definition
 */
export interface BooleanPropertyDefinition extends BasePropertyDefinition {
  type: 'boolean';
}

/**
 * Object property definition with nested properties
 */
export interface ObjectPropertyDefinition extends BasePropertyDefinition {
  type: 'object';
  properties: Record<string, PropertyDefinition>;
}

/**
 * Array item definition (can be primitive or object)
 */
export type ArrayItemDefinition =
  | {
      type: 'string' | 'number' | 'boolean';
      enum?: string[];
    }
  | {
      type: 'object';
      properties: Record<string, PropertyDefinition>;
    };

/**
 * Array property definition
 */
export interface ArrayPropertyDefinition extends BasePropertyDefinition {
  type: 'array';
  items: ArrayItemDefinition;
}

/**
 * Discriminated union of all property definition types
 */
export type PropertyDefinition =
  | StringPropertyDefinition
  | NumberPropertyDefinition
  | BooleanPropertyDefinition
  | ObjectPropertyDefinition
  | ArrayPropertyDefinition;

/**
 * User schema definition (key-value pairs of property definitions)
 */
export type UserSchemaDefinition = Record<string, PropertyDefinition>;

/**
 * Complete User Schema object as returned by API
 */
export interface ApiUserSchema {
  id: string;
  name: string;
  schema: UserSchemaDefinition;
}

/**
 * User Schema list item (minimal representation)
 */
export interface UserSchemaListItem {
  id: string;
  name: string;
}

/**
 * Link object for pagination
 */
export interface Link {
  rel: string;
  href: string;
}

/**
 * Response for GET /user-schemas (list with pagination)
 */
export interface UserSchemaListResponse {
  totalResults: number;
  startIndex: number;
  count: number;
  schemas: UserSchemaListItem[];
  links?: Link[];
}

/**
 * Request body for POST /user-schemas (create)
 */
export interface CreateUserSchemaRequest {
  name: string;
  schema: UserSchemaDefinition;
}

/**
 * Request body for PUT /user-schemas/{id} (update)
 */
export interface UpdateUserSchemaRequest {
  name: string;
  schema: UserSchemaDefinition;
}

/**
 * Query parameters for listing user schemas
 */
export interface UserSchemaListParams {
  limit?: number;
  offset?: number;
}

/**
 * API Error structure
 */
export interface ApiError {
  code: string;
  message: string;
  description: string;
}

/**
 * Property type union for form inputs
 */
export type PropertyType = 'string' | 'number' | 'boolean' | 'array' | 'object';

/**
 * Schema property input type for create/edit forms
 */
export interface SchemaPropertyInput {
  id: string;
  name: string;
  type: PropertyType;
  required: boolean;
  unique: boolean;
  enum: string[];
  regex: string;
}
