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

import axios from 'axios';
import config from '../config';

export const NativeAuthSubmitType = {
    INPUT: 'INPUT',
    SOCIAL: 'SOCIAL',
    OTP: 'OTP',
} as const;

export type NativeAuthSubmitType = (typeof NativeAuthSubmitType)[keyof typeof NativeAuthSubmitType];

type NativeAuthSubmitPayload =
  | { type: typeof NativeAuthSubmitType.INPUT; [key: string]: string }
  | { type: typeof NativeAuthSubmitType.SOCIAL; code: string }
  | { type: typeof NativeAuthSubmitType.OTP; otp: string };

const { applicationID, clientId, flowEndpoint, redirectUri, tokenEndpoint } = config;

/**
 * Generates a cryptographically secure random code verifier for PKCE (RFC 7636).
 * 
 * Generates a 43-character base64url-encoded string from 32 random bytes,
 * which complies with RFC 7636 requirements (minimum 43 characters, maximum 128 characters).
 * 
 * @returns {string} - PKCE code verifier.
 */
const generateCodeVerifier = (): string => {
    const array = new Uint8Array(32);
    crypto.getRandomValues(array);
    return base64UrlEncode(array);
};

/**
 * Base64 URL encodes a byte array.
 * 
 * @param {Uint8Array} buffer - The byte array to encode.
 * @returns {string} - The base64 URL encoded string.
 */
const base64UrlEncode = (buffer: Uint8Array): string => {
    let binary = '';
    for (let i = 0; i < buffer.length; i++) {
        binary += String.fromCharCode(buffer[i]);
    }
    const base64 = btoa(binary);
    return base64.replace(/\+/g, '-').replace(/\//g, '_').replace(/=/g, '');
};

/**
 * Generates a code challenge from a code verifier using SHA-256.
 * 
 * @param {string} verifier - The code verifier.
 * @returns {Promise<string>} - A promise that resolves to the code challenge.
 */
const generateCodeChallenge = async (verifier: string): Promise<string> => {
    const encoder = new TextEncoder();
    const data = encoder.encode(verifier);
    const hash = await crypto.subtle.digest('SHA-256', data);
    return base64UrlEncode(new Uint8Array(hash));
};

/**
 * Initiates the OAuth 2.0 authorization code flow by redirecting the user to the authorization endpoint.
 * 
 * @returns {void}
 */
export const initiateRedirectAuth = () => {
    const state = Math.random().toString(36).substring(2, 15); // Generate a random state.
    const url = `${flowEndpoint}/authn?response_type=code&client_id=${clientId}&redirect_uri=${redirectUri}&scope=openid&state=${state}`;
    window.location.href = url;
};

/**
 * Initiates the native authentication or registration flow by sending a POST request to the flow endpoint.
 * 
 * @param {string} flowType - The type of flow to initiate. Defaults to 'LOGIN'.
 * @returns {Promise<object>} - A promise that resolves to the response data from the server.
 */
export const initiateNativeAuthFlow = async (flowType: 'LOGIN' | 'REGISTRATION' = 'LOGIN') => {
    const headers = {
        'Content-Type': 'application/json'
    };

    const data: Record<string, string> = {
        "applicationId": applicationID
    };

    if (flowType === 'REGISTRATION') {
        data.flowType = 'REGISTRATION';
    } else {
        data.flowType = 'AUTHENTICATION';
    }

    try {
        const response = await axios.post(`${flowEndpoint}/execute`, data, {
            headers,
        });

        return { data: response.data };
    } catch (error) {
        if (axios.isAxiosError(error)) {
            const flowTypeName = flowType === 'REGISTRATION' ? 'registration' : 'authentication';
            const message = error.response?.status === 400
              ? `Error initiating native ${flowTypeName} request.`
              : error.response?.data?.message || 'Server error occurred.';
            throw new Error(message);
        } else {
            throw new Error('Unexpected error occurred.');
        }
    }
};

/**
 * Initiates the native authentication or registration flow with additional data.
 * 
 * @param {string} flowType - The type of flow to initiate. Defaults to 'LOGIN'.
 * @param {string} actionId - The ID of the action to execute.
 * @param {object} inputs - Optional input data to include in the request.
 * @returns {Promise<object>} - A promise that resolves to the response data from the server.
 */
export const initiateNativeAuthFlowWithData = async (flowType: 'LOGIN' | 'REGISTRATION' = 'LOGIN', 
    actionId: string | null, inputs?: Record<string, unknown>) => {
    const headers = {
        'Content-Type': 'application/json'
    };

    const data: Record<string, unknown> = {
        "applicationId": applicationID,
    };

    if (actionId) {
        data.action = actionId;
    }

    if (flowType === 'REGISTRATION') {
        data.flowType = 'REGISTRATION';
    } else {
        data.flowType = 'AUTHENTICATION';
    }

    // Include inputs if provided
    if (inputs && Object.keys(inputs).length > 0) {
        data.inputs = inputs;
    }

    try {
        const response = await axios.post(`${flowEndpoint}/execute`, data, {
            headers,
        });

        return { data: response.data };
    } catch (error) {
        if (axios.isAxiosError(error)) {
            const flowTypeName = flowType === 'REGISTRATION' ? 'registration' : 'authentication';
            const message = error.response?.status === 400
              ? `Error initiating native ${flowTypeName} request.`
              : error.response?.data?.message || 'Server error occurred.';
            throw new Error(message);
        } else {
            throw new Error('Unexpected error occurred.');
        }
    }
};

/**
 * Submits the user's selected authentication option when multiple options are available.
 * 
 * @param {string} flowId - The flow ID received from the initiateNativeAuth response.
 * @param {string} actionId - The ID of the selected authentication action.
 * @param {object} inputs - Optional input data to submit with the decision.
 * @returns {Promise<object>} - A promise that resolves to the response data from the server.
 */
export const submitAuthDecision = async (flowId: string, actionId: string, inputs?: Record<string, unknown>) => {
    const headers = {
        'Content-Type': 'application/json'
    };

    const data: Record<string, unknown> = {
        flowId: flowId,
        action: actionId
    };

    // Include inputs if provided
    if (inputs && Object.keys(inputs).length > 0) {
        data.inputs = inputs;
    }

    try {
        const response = await axios.post(`${flowEndpoint}/execute`, data, {
            headers,
        });

        return { data: response.data };
    } catch (error) {
        if (axios.isAxiosError(error)) {
            const message = error.response?.status === 400
              ? 'Error processing authentication option.'
              : error.response?.data?.message || 'Server error occurred.';
            throw new Error(message);
        } else {
            throw new Error('Unexpected error occurred.');
        }
    }
};

/**
 * Submits the native authentication form data to the server.
 * 
 * @param {string} flowId - The flow ID received from the initiateNativeAuth response.
 * @param {object} payload - The payload containing the form data or other required information.
 * @param {string} action - Optional action ref to include in the request.
 * @returns {Promise<object>} - A promise that resolves to the response data from the server.
 */
export const submitNativeAuth = async (
    flowId: string,
    payload: Record<string, unknown> | NativeAuthSubmitPayload,
    action?: string
) => {
    const headers = {
        'Content-Type': 'application/json'
    };

    const data: Record<string, unknown> = {
        flowId: flowId
    };

    // Include action if provided
    if (action) {
        data.action = action;
    }

    if ('type' in payload) {
        if (payload.type === NativeAuthSubmitType.INPUT) {
            // For input type, include all fields except 'type'
            const { ...inputValues } = payload;
            data.inputs = inputValues;
        } else if (payload.type === NativeAuthSubmitType.SOCIAL) {
            data.inputs = {
                code: payload.code
            };
        } else if (payload.type === NativeAuthSubmitType.OTP) {
            data.inputs = {
                otp: payload.otp
            };
        }
    } else {
        // Handle as generic payload
        data.inputs = payload;
    }

    try {
        const response = await axios.post(`${flowEndpoint}/execute`, data, {
            headers,
        });

        return { data: response.data };
    } catch (error) {
        if (axios.isAxiosError(error)) {
            const message = error.response?.status === 400
              ? 'Login failed. Please check your credentials.'
              : error.response?.data?.message || 'Server error occurred.';
            throw new Error(message);
        } else {
            throw new Error('Unexpected error occurred.');
        }
    }
}

/**
 * Exchanges the authorization code for an access token.
 * 
 * @param {string} code - The authorization code received from the OAuth server.
 * @param {string | null} [codeVerifier] - Optional code verifier for PKCE.
 * @returns {Promise<object>} - A promise that resolves to the access token data.
 */
export const exchangeCodeForToken = async (code: string, codeVerifier?: string | null) => {
    const headers = {
        'Content-Type': 'application/x-www-form-urlencoded',
    };

    const data = new URLSearchParams();
    data.append('grant_type', 'authorization_code');
    data.append('redirect_uri', redirectUri);
    data.append('code', code);
    data.append('client_id', clientId);
    
    // Include code_verifier if provided (PKCE)
    if (codeVerifier) {
        data.append('code_verifier', codeVerifier);
    }

    try {
        const response = await axios.post(tokenEndpoint, data, {
            headers,
        });
        return response.data; // This will contain the access token
    } catch (error) {
        console.error('Error exchanging code for token:', error);
        throw error;
    }
};

/**
 * Generates a cryptographically secure random state parameter.
 * 
 * @returns {string} - A secure random state string for CSRF protection.
 */
const generateSecureState = (): string => {
    const array = new Uint8Array(16);
    crypto.getRandomValues(array);
    return base64UrlEncode(array);
};

/**
 * Initiates the OAuth 2.0 authorization code flow with PKCE support.
 * Generates code verifier and challenge, stores the verifier and state, and redirects to authorization endpoint.
 * 
 * @param {string} authorizationEndpoint - The OAuth2 authorization endpoint URL.
 * @param {string} clientId - The OAuth2 client ID.
 * @param {string} redirectUri - The redirect URI.
 * @param {string} scope - The OAuth2 scope.
 * @returns {Promise<void>}
 */
export const initiateAuthWithPKCE = async (
    authorizationEndpoint: string,
    clientId: string,
    redirectUri: string,
    scope: string
): Promise<void> => {
    const state = generateSecureState();
    const codeVerifier = generateCodeVerifier();
    const codeChallenge = await generateCodeChallenge(codeVerifier);
    
    // Store code verifier and state in session storage for later use
    sessionStorage.setItem('pkce_code_verifier', codeVerifier);
    sessionStorage.setItem('pkce_state', state);
    
    const params = new URLSearchParams({
        response_type: 'code',
        client_id: clientId,
        redirect_uri: redirectUri,
        scope: scope,
        state: state,
        code_challenge: codeChallenge,
        code_challenge_method: 'S256'
    });
    
    window.location.href = `${authorizationEndpoint}?${params.toString()}`;
};

/**
 * Validates the state parameter for CSRF protection.
 * 
 * @param {string} returnedState - The state parameter returned from the authorization server.
 * @returns {boolean} - True if the state is valid, false otherwise.
 */
export const validateState = (returnedState: string): boolean => {
    const storedState = sessionStorage.getItem('pkce_state');
    return storedState !== null && storedState === returnedState;
};

/**
 * Retrieves the stored code verifier from session storage and cleans up.
 * 
 * @returns {string | null} - The stored code verifier or null if not found.
 */
export const getAndClearCodeVerifier = (): string | null => {
    const codeVerifier = sessionStorage.getItem('pkce_code_verifier');
    if (codeVerifier) {
        sessionStorage.removeItem('pkce_code_verifier');
        sessionStorage.removeItem('pkce_state');
    }
    return codeVerifier;
};
