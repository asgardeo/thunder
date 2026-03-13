/**
 * Copyright (c) 2026, WSO2 LLC. (https://www.wso2.com).
 *
 * WSO2 LLC. licenses this file to you under the Apache License,
 * Version 2.0 (the "License"); you may not use this file except
 * in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */

import {Monitor, Moon, Sun} from '@wso2/oxygen-ui-icons-react';
import type {JSX} from 'react';
import type {ColorSchemeOption} from '@thunder/shared-design';

export interface ColorSchemeOptionItem {
  id: ColorSchemeOption;
  label: string;
  icon: JSX.Element;
}

/**
 * Display options for color scheme selection (includes icons for UI rendering).
 */
const ColorSchemeOptions: ColorSchemeOptionItem[] = [
  {id: 'light', label: 'Light', icon: <Sun size={14} />},
  {id: 'dark', label: 'Dark', icon: <Moon size={14} />},
  {id: 'system', label: 'System', icon: <Monitor size={14} />},
];

export default ColorSchemeOptions;
