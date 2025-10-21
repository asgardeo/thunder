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

// Load runtime configuration asynchronously
const loadRuntimeConfig = async (): Promise<RuntimeConfig> => {
    if (import.meta.env.DEV) {
        // In development mode, skip fetching runtime.json
        return {};
    }

    try {
        const response = await fetch('/runtime.json');
        if (!response.ok) {
            console.warn(`Failed to fetch runtime.json: ${response.status} ${response.statusText}`);
            return {};
        }
        return await response.json();
    } catch (error) {
        console.error('Error loading runtime configuration:', error);
        // Return empty config to fallback to environment variables
        return {};
    }
};

// Initialize runtime config with promise-based guard to prevent race conditions
let runtimeConfig: RuntimeConfig = {};
let configInitializationPromise: Promise<void> | null = null;

const initializeConfig = async (): Promise<void> => {
    // If already initializing or initialized, return the existing promise
    if (configInitializationPromise) {
        return configInitializationPromise;
    }

    // Create and store the initialization promise
    // Wrap in try-catch to ensure the promise never rejects
    configInitializationPromise = (async () => {
        try {
            runtimeConfig = await loadRuntimeConfig();
        } catch (error) {
            // This should never happen since loadRuntimeConfig has its own error handling,
            // but we catch it here as a safety measure to ensure the promise always resolves
            console.error('Unexpected error during config initialization:', error);
            runtimeConfig = {};
        }
    })();

    return configInitializationPromise;
};

// Start initialization immediately
const configReady = initializeConfig();

/**
 * Ensures the configuration is fully loaded before proceeding.
 *
 * IMPORTANT: You must call this function and await it before accessing config values
 * in production mode to ensure runtime.json has been loaded.
 *
 * @example
 * ```tsx
 * await ensureConfigReady();
 * const endpoint = config.applicationID; // Now safe to access
 * ```
 */
export const ensureConfigReady = (): Promise<void> => configReady;

const config = {
    get applicationID() {
        return getConfigValue(runtimeConfig.applicationID, import.meta.env.VITE_REACT_APP_AUTH_APP_ID) || '';
    },
    get applicationsEndpoint() {
        return getConfigValue(runtimeConfig.applicationsEndpoint, import.meta.env.VITE_REACT_APPLICATIONS_ENDPOINT) || '';
    },
    get flowEndpoint() {
        return getConfigValue(runtimeConfig.flowEndpoint, import.meta.env.VITE_REACT_APP_SERVER_FLOW_ENDPOINT) || '';
    },
    get authorizationEndpoint() {
        return getConfigValue(runtimeConfig.authorizationEndpoint, import.meta.env.VITE_REACT_APP_SERVER_AUTHORIZATION_ENDPOINT) || '';
    },
    get tokenEndpoint() {
        return getConfigValue(runtimeConfig.tokenEndpoint, import.meta.env.VITE_REACT_APP_SERVER_TOKEN_ENDPOINT) || '';
    },
    get clientId() {
        return import.meta.env.VITE_REACT_APP_CLIENT_ID || '';
    },
    get clientSecret() {
        return import.meta.env.VITE_REACT_APP_CLIENT_SECRET || '';
    },
    get redirectUri() {
        return getConfigValue(runtimeConfig.redirectUri, import.meta.env.VITE_REACT_APP_REDIRECT_URI) || '';
    },
    get scope() {
        return import.meta.env.VITE_REACT_APP_SCOPE || '';
    }
};

export default config;
