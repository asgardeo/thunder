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

interface RuntimeConfig {
    applicationID?: string;
    applicationsEndpoint?: string;
    flowEndpoint?: string;
    authorizationEndpoint?: string;
    tokenEndpoint?: string;
    redirectUri?: string;
}

let runtimeConfig: RuntimeConfig = {};
if (!import.meta.env.DEV) {
    const response = await fetch('/runtime.json');
    runtimeConfig = await response.json();
}

// Helper function to get config value, preferring env vars in development mode
// and filtering out placeholder values from both runtime and env sources
const getConfigValue = (runtimeValue: string | undefined, envValue: string | undefined): string | undefined => {
    const isPlaceholder = (value: string | undefined): boolean => {
        return !value || (value.startsWith('{') && value.endsWith('}'));
    };

    // In development mode, prefer env variables, filtering out placeholders from both
    const filteredEnvValue = isPlaceholder(envValue) ? undefined : envValue;
    const filteredRuntimeValue = isPlaceholder(runtimeValue) ? undefined : runtimeValue;

    if (import.meta.env.DEV) {
        return filteredEnvValue || filteredRuntimeValue;
    }

    // In production mode, prefer runtime config but filter out placeholders from both
    return filteredRuntimeValue || filteredEnvValue;
};

const config = {
    applicationID: getConfigValue(runtimeConfig.applicationID, import.meta.env.VITE_REACT_APP_AUTH_APP_ID) || '',
    applicationsEndpoint: getConfigValue(runtimeConfig.applicationsEndpoint, import.meta.env.VITE_REACT_APPLICATIONS_ENDPOINT) || '',
    flowEndpoint: getConfigValue(runtimeConfig.flowEndpoint, import.meta.env.VITE_REACT_APP_SERVER_FLOW_ENDPOINT) || '',
    authorizationEndpoint: getConfigValue(runtimeConfig.authorizationEndpoint, import.meta.env.VITE_REACT_APP_SERVER_AUTHORIZATION_ENDPOINT) || '',
    tokenEndpoint: getConfigValue(runtimeConfig.tokenEndpoint, import.meta.env.VITE_REACT_APP_SERVER_TOKEN_ENDPOINT) || '',
    clientId: import.meta.env.VITE_REACT_APP_CLIENT_ID || '',
    clientSecret: import.meta.env.VITE_REACT_APP_CLIENT_SECRET || '',
    redirectUri: getConfigValue(runtimeConfig.redirectUri, import.meta.env.VITE_REACT_APP_REDIRECT_URI) || '',
    scope: import.meta.env.VITE_REACT_APP_SCOPE || ''
};

export default config;
