/*
 * Copyright (c) 2025, WSO2 LLC. (http://www.wso2.com).
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
import { fileURLToPath } from 'url';

const __dirname = path.dirname(fileURLToPath(import.meta.url));

// paths
const projectRoot = path.resolve(__dirname, '../');
const frontendRootFolder = path.resolve(__dirname, '../', '../', '../');
const distFolder = path.join(projectRoot, 'dist');
const packageJsonPath = path.join(distFolder, 'package.json');

const filesToCopy = [
  { src: path.join(projectRoot, 'package.json'), dest: path.join(distFolder, 'package.json') },
  { src: path.join(projectRoot, 'tsconfig.json'), dest: path.join(distFolder, 'tsconfig.json') },
  { src: path.join(projectRoot, 'next.config.ts'), dest: path.join(distFolder, 'next.config.ts') },
  { src: path.join(projectRoot, 'server.js'), dest: path.join(distFolder, 'server.js') },
  { src: path.join(projectRoot, 'server.key'), dest: path.join(distFolder, 'server.key') },
  { src: path.join(projectRoot, 'server.cert'), dest: path.join(distFolder, 'server.cert') },
  { src: path.join(projectRoot, 'public'), dest: path.join(distFolder, 'public') },
  { src: path.join(projectRoot, 'src'), dest: path.join(distFolder, 'src') },
  { src: path.join(frontendRootFolder, 'packages'), dest: path.join(distFolder, 'packages') },
];

if (fs.existsSync(distFolder)) {
  // Remove the folder and its contents
  fs.rmSync(distFolder, { recursive: true, force: true });
}

// Recreate the folder
fs.mkdirSync(distFolder, { recursive: true });

// Copy files and folders
for (const { src, dest } of filesToCopy) {
  if (fs.existsSync(src)) {
    const stats = fs.statSync(src);
    if (stats.isDirectory()) {
      // Copy directory (Node.js 16+)
      fs.cpSync(src, dest, {
        recursive: true,
        filter: (srcPath) => {
          // Exclude the "node_modules" folder
          return !srcPath.includes('node_modules');
        },
      });
      console.log(`Copied directory: ${src} -> ${dest}`);
    } else {
      // Copy file
      fs.copyFileSync(src, dest);
      console.log(`Copied file: ${src} -> ${dest}`);
    }
  } else {
    console.error(`Source not found: ${src}`);
  }
}

// Edit package.json in the dist folder. and update "file:../../packages" to "file:./packages" path
if (fs.existsSync(packageJsonPath)) {
  const packageJson = JSON.parse(fs.readFileSync(packageJsonPath, 'utf-8'));

  // Modify all dependencies starting with "file:../../packages"
  if (packageJson.dependencies) {
    for (const [key, value] of Object.entries(packageJson.dependencies)) {
      if (value.startsWith('file:../../packages')) {
        // Replace "file:../../packages" with "file:./packages"
        packageJson.dependencies[key] = value.replace('file:../../packages', 'file:./packages');
      }
    }
  }

  // Write the updated package.json back to the dist folder
  fs.writeFileSync(packageJsonPath, JSON.stringify(packageJson, null, 2), 'utf-8');
  console.log(`Updated package.json in ${distFolder}`);
} else {
  console.error(`package.json not found in ${distFolder}`);
}
