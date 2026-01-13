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

/**
 * Global Setup
 *
 * Runs once before all tests. Use for expensive operations
 * that only need to happen once per test run.
 */

import dotenv from "dotenv";
import path from "path";
import fs from "fs";

async function globalSetup() {
  console.log("🚀 Running global setup...");

  // Load environment variables
  const envPath = path.resolve(__dirname, "../.env");
  dotenv.config({ path: envPath });

  // Verify required environment variables
  const requiredVars = ["BASE_URL", "ADMIN_USERNAME", "ADMIN_PASSWORD"];
  const missingVars = requiredVars.filter(varName => !process.env[varName]);

  if (missingVars.length > 0) {
    console.error("❌ Missing required environment variables:", missingVars.join(", "));
    console.error("Please create a .env file based on .env.example");
    process.exit(1);
  }

  // Ensure auth directory exists
  const authDir = path.join(__dirname, "../playwright/.auth");
  if (!fs.existsSync(authDir)) {
    fs.mkdirSync(authDir, { recursive: true });
  }

  console.log("✅ Global setup complete");
  console.log(`   Base URL: ${process.env.BASE_URL}`);
  console.log(`   Admin User: ${process.env.ADMIN_USERNAME}`);
}

export default globalSetup;
