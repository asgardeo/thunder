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

import {useContext} from 'react';
import BrandingContext, {BrandingContextType} from './BrandingContext';

/**
 * React hook for accessing Thunder branding configuration throughout the application.
 *
 * This hook provides access to the branding data loaded from the server, resolved theme,
 * layout configuration, and images. It must be used within a component tree wrapped by
 * `BrandingProvider`, otherwise it will throw an error.
 *
 * The hook returns a context object containing the complete branding data and utility
 * methods for common operations like getting theme, layout, images, and checking
 * branding status.
 *
 * @returns The branding context containing branding data and utility methods
 *
 * @throws {Error} Throws an error if used outside of BrandingProvider
 *
 * @example
 * Basic usage:
 * ```tsx
 * import useBranding from './useBranding';
 *
 * function MyComponent() {
 *   const {
 *     isBrandingEnabled,
 *     theme,
 *     images,
 *     layout
 *   } = useBranding();
 *
 *   const isLeftAligned = layout?.type === 'LEFT_ALIGNED';
 *
 *   return (
 *     <div>
 *       <p>Branding enabled: {isBrandingEnabled}</p>
 *       <p>Layout type: {layout?.type}</p>
 *       <p>Is left aligned: {isLeftAligned}</p>
 *     </div>
 *   );
 * }
 * ```
 *
 * @example
 * Using theme, images, and layout directly:
 * ```tsx
 * function ThemedComponent() {
 *   const { theme, images, layout, isBrandingEnabled } = useBranding();
 *
 *   return (
 *     <div
 *       style={{
 *         backgroundColor: theme?.palette?.primary?.main,
 *         color: theme?.palette?.primary?.contrastText,
 *       }}
 *     >
 *       Custom themed content
 *     </div>
 *   );
 * }
 * ```
 *
 * @public
 */
export default function useBranding(): BrandingContextType {
  const context = useContext(BrandingContext);
  if (context === undefined) {
    throw new Error('useBranding must be used within a BrandingProvider');
  }
  return context;
}
