/**
 * Copyright (c) 2025, WSO2 LLC. (https://www.wso2.com).
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

import type {ThemeConfig} from '@/features/themes/models/theme';
import type {LayoutConfig} from '@/features/layouts/models/layout';

/**
 * Device types for responsive preview
 */
export type DeviceType = 'desktop' | 'tablet' | 'mobile';

/**
 * Device viewport dimensions
 */
export interface DeviceViewport {
  width: number;
  height: number;
  label: string;
}

/**
 * Predefined device viewports
 */
export const DEVICE_VIEWPORTS: Record<DeviceType, DeviceViewport> = {
  desktop: {
    width: 1440,
    height: 900,
    label: 'Desktop',
  },
  tablet: {
    width: 768,
    height: 1024,
    label: 'Tablet',
  },
  mobile: {
    width: 375,
    height: 667,
    label: 'Mobile',
  },
};

/**
 * Design mode - visual or code
 * Code mode will be implemented in the future for Auth0 AUCL-like functionality
 */
export type DesignMode = 'visual' | 'code';

/**
 * Combined design configuration
 * Includes both theme and layout, and can be extended with code-based customization
 */
export interface DesignConfig {
  id?: string;
  displayName: string;
  description?: string;
  mode: DesignMode;
  theme: ThemeConfig;
  layout: LayoutConfig;
  // Future: code-based customization
  customCode?: {
    html?: string;
    css?: string;
    javascript?: string;
  };
  createdAt?: string;
  updatedAt?: string;
}

/**
 * Design editor state
 */
export interface DesignEditorState {
  selectedDevice: DeviceType;
  zoom: number;
  showRulers: boolean;
  showGrid: boolean;
  selectedElement?: string;
  mode: DesignMode;
}

/**
 * Default design editor state
 */
export const DEFAULT_EDITOR_STATE: DesignEditorState = {
  selectedDevice: 'desktop',
  zoom: 100,
  showRulers: true,
  showGrid: false,
  mode: 'visual',
};
