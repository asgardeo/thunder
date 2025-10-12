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

const fs = require('fs');
const path = require('path');

const rootDir = path.resolve(__dirname, '..');
const componentsDir = path.join(rootDir, 'src', 'components');
const packageJsonPath = path.join(rootDir, 'package.json');

const packageJson = require(packageJsonPath);

const exportsField = {
  '.': {
    import: './dist/index.js',
    require: './dist/index.js',
  },
};

fs.readdirSync(componentsDir).forEach((file) => {
  const ext = path.extname(file);
  const name = path.basename(file, ext);
  if (ext === '.ts' || ext === '.tsx') {
    exportsField[`./${name}`] = {
      import: `./dist/components/${name}.js`,
      require: `./dist/components/${name}.js`,
    };
  }
});

packageJson.exports = exportsField;

fs.writeFileSync(packageJsonPath, JSON.stringify(packageJson, null, 2));
