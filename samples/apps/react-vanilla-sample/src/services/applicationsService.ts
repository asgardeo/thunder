/*
 * Copyright (c) 2025, WSO2 LLC. (https://www.wso2.com).
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

import config from '../config';

const { applicationsEndpoint } = config;

/**
 * Get applications list from the server.
 *
 * @returns A promise that resolves to the list of applications.
 */
export const getApplications = async () => {
    const headers = {
        'Content-Type': 'application/json',
        'Accept': 'application/json',
    };

    const response = await fetch(applicationsEndpoint, { headers });

    if (!response.ok) {
        const error = await response.json().catch(() => ({}));
        console.error('Error retrieving applications list:', error);
        throw new Error('Failed to retrieve applications list.');
    }

    return response.json();
};
