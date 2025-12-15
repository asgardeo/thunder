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

/**
 * Global timeouts for Playwright E2E tests
 */
export const Timeouts = {
    
    /** Default timeout for UI actions (clicks, fills) */
    DEFAULT_ACTION: 15000,

    /** Timeout for checking element visibility */
    ELEMENT_VISIBILITY: 5000,

    /** Timeout for loading large forms */
    FORM_LOAD: 10000,

    /** Timeout for full page loads */
    PAGE_LOAD: 30000,

    /** Global test timeout */
    GLOBAL_TEST: 60000,

    /** Wait for network idle state */
    NETWORK_IDLE: 5000,

    /** Search debounce wait */
    SEARCH_DEBOUNCE: 500,

    /** Auth initialization wait */
    AUTH_INITIALIZATION: 500,

    /** Post auth wait */
    POST_AUTH: 2000,

    /** Timeout for login redirects */
    REDIRECT: 20000,
} as const;
