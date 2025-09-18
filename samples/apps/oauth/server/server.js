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

const express = require('express');
const path = require('path');
const https = require('https');
const fs = require('fs');

const app = express();
const PORT = process.env.PORT || 3000;

// Use actual working directory to access certs
const isRunningAsExecutable = process.pkg !== undefined;
const certDir = isRunningAsExecutable
  ? path.dirname(process.execPath)
  : path.join(process.cwd());

const keyPath = path.join(certDir, 'server.key');
const certPath = path.join(certDir, 'server.cert');

// Serve static files from the 'dist' directory
const appDir = path.join(certDir, 'app');
app.use(express.static(appDir));

// Handle SPA routing
app.get('*', (req, res) => {
  res.sendFile(path.join(appDir, 'index.html'));
});

if (fs.existsSync(keyPath) && fs.existsSync(certPath)) {
  const sslOptions = {
    key: fs.readFileSync(keyPath),
    cert: fs.readFileSync(certPath),
  };

  https.createServer(sslOptions, app).listen(PORT, () => {
    console.log(`✅ HTTPS server running at https://localhost:${PORT}`);
    console.log('Press Ctrl+C to stop the server.');
  });
} else {
  app.listen(PORT, () => {
    console.log(`⚠️  HTTPS certs missing. Falling back to HTTP at http://localhost:${PORT}`);
    console.log('Press Ctrl+C to stop the server.');
  });
}
