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

import fs from 'fs';
import path from 'path';
import https from 'https';
import { fileURLToPath, parse } from 'url';
import next from 'next';

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

const requiredServerFilesConfig = path.resolve(__dirname, '.next/required-server-files.json');

// Check if the required server files configuration exists
if (fs.existsSync(requiredServerFilesConfig)) {
  const jsonData = JSON.parse(fs.readFileSync(requiredServerFilesConfig, 'utf-8'));
  const nextConfig = jsonData.config || {};

  process.env.__NEXT_PRIVATE_STANDALONE_CONFIG = JSON.stringify(nextConfig);
}

const PORT = 9090;
const HOST = 'localhost';
const keyPath = path.resolve(__dirname, 'server.key');
const certPath = path.resolve(__dirname, 'server.cert');
const dev = process.env.NODE_ENV === 'development';
const app = next({ dev: dev, dir: __dirname });
const httpsOptions = {
  key: fs.readFileSync(keyPath),
  cert: fs.readFileSync(certPath),
};

const handle = app.getRequestHandler();

function getTimestampWithOffset() {
  const now = new Date();

  const pad = (n) => n.toString().padStart(2, '0');

  const isoDate = now.getFullYear() + '-' +
                  pad(now.getMonth() + 1) + '-' +
                  pad(now.getDate()) + 'T' +
                  pad(now.getHours()) + ':' +
                  pad(now.getMinutes()) + ':' +
                  pad(now.getSeconds()) + '.' +
                  now.getMilliseconds().toString().padStart(3, '0');

  const offsetMin = now.getTimezoneOffset();
  const offsetSign = offsetMin <= 0 ? '+' : '-';
  const offsetHours = pad(Math.floor(Math.abs(offsetMin) / 60));
  const offsetMinutes = pad(Math.abs(offsetMin) % 60);
  const tzOffset = `${offsetSign}${offsetHours}:${offsetMinutes}`;

  return `${isoDate}${tzOffset}`;
}

console.log(`Starting WSO2 Thunder gate app in ${dev ? 'development' : 'production'} mode...`);

app.prepare().then(() => {
  https.createServer(httpsOptions, (req, res) => {
    const parsedUrl = parse(req.url, true);
    handle(req, res, parsedUrl);
  }).listen(PORT, () => {
    const isoWithOffset = getTimestampWithOffset();
    console.log(
      `time=${isoWithOffset} level=INFO msg="WSO2 Thunder gate app started..." address=${HOST}:${PORT}`
    );
  });
});
