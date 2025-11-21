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

import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import "./index.css";
import App from "./App.tsx";
import { AsgardeoProvider } from "@asgardeo/react";
import { ConfigurationError } from "./ConfigurationError.tsx";

const baseUrl = import.meta.env.VITE_THUNDER_BASE_URL as string;
const clientId = import.meta.env.VITE_REACT_APP_CLIENT_ID as string;

// Validate required configuration
const missingConfig: string[] = [];
if (!baseUrl) {
  missingConfig.push("VITE_THUNDER_BASE_URL");
}
if (!clientId) {
  missingConfig.push("VITE_REACT_APP_CLIENT_ID");
}

if (missingConfig.length > 0) {
  console.error(
    "⚠️ Missing required environment variables:",
    missingConfig.join(", ")
  );
  console.error(
    "Please configure these variables in your .env file. See .env.example for reference."
  );
}

createRoot(document.getElementById("root")!).render(
  <StrictMode>
    {missingConfig.length > 0 ? (
      <ConfigurationError missingConfig={missingConfig} />
    ) : (
      <AsgardeoProvider
        baseUrl={baseUrl}
        clientId={clientId}
        platform="AsgardeoV2"
      >
        <App />
      </AsgardeoProvider>
    )}
  </StrictMode>
);
