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

import { getConfig } from "../config";

export interface OrganizationUnit {
  id: string;
  handle: string;
  name: string;
  description?: string;
}

interface OrganizationUnitListResponse {
  totalResults: number;
  startIndex: number;
  count: number;
  organizationUnits: OrganizationUnit[];
}

const cachedOrganizationUnits: Record<string, string> = {};

async function fetchOrganizationUnits(): Promise<OrganizationUnit[]> {
  const { baseUrl } = getConfig();

  const response = await fetch(`${baseUrl}/organization-units`, {
    method: "GET",
    headers: {
      "Content-Type": "application/json",
    },
  });

  if (!response.ok) {
    throw new Error("Failed to fetch organization units");
  }

  const data: OrganizationUnitListResponse = await response.json();
  return data.organizationUnits;
}

export async function getOrganizationUnitId(handle: string): Promise<string> {
  if (cachedOrganizationUnits[handle]) {
    return cachedOrganizationUnits[handle];
  }

  const organizationUnits = await fetchOrganizationUnits();

  if (organizationUnits.length === 0) {
    throw new Error("No organization units found");
  }

  const ou = organizationUnits.find((ou) => ou.handle === handle);

  if (!ou) {
    throw new Error(`Organization unit "${handle}" not found`);
  }

  cachedOrganizationUnits[handle] = ou.id;
  return ou.id;
}

export async function getDefaultOrganizationUnitId(): Promise<string> {
  return getOrganizationUnitId("default");
}

export async function getCustomersOrganizationUnitId(): Promise<string> {
  return getOrganizationUnitId("customers");
}

export interface User {
  id: string;
  organizationUnit: string;
  type: string;
  attributes: {
    username?: string;
    firstName?: string;
    lastName?: string;
    email?: string;
    [key: string]: unknown;
  };
}

interface UserListResponse {
  totalResults: number;
  startIndex: number;
  count: number;
  users: User[];
}

export async function fetchUsers(filter?: string): Promise<User[]> {
  const { baseUrl } = getConfig();

  let url = `${baseUrl}/users?limit=100`;
  if (filter) {
    url += `&filter=${encodeURIComponent(filter)}`;
  }

  const response = await fetch(url, {
    method: "GET",
    headers: {
      "Content-Type": "application/json",
    },
  });

  if (!response.ok) {
    throw new Error("Failed to fetch users");
  }

  const data: UserListResponse = await response.json();
  return data.users;
}
