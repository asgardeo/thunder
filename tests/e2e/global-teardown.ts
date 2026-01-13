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
 * Global Teardown
 *
 * Runs once after all tests. Use for cleanup operations.
 */

import fs from "fs";
import path from "path";

async function globalTeardown() {
  console.log("🧹 Running global teardown...");

  // Optional: Clean up old auth files (keep the latest)
  const authDir = path.join(__dirname, "../playwright/.auth");
  if (fs.existsSync(authDir)) {
    const files = fs.readdirSync(authDir);
    const oldFiles = files.filter(f => f.startsWith("working-login") && f !== "working-login.json");

    oldFiles.forEach(file => {
      try {
        fs.unlinkSync(path.join(authDir, file));
      } catch (err) {
        console.warn(`Could not delete ${file}:`, err);
      }
    });
  }

  console.log("✅ Global teardown complete");
}

export default globalTeardown;
