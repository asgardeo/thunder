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

import {useMemo, PropsWithChildren} from 'react';
import {useConfig} from '@thunder/shared-contexts';
import BrandingContext, {type BrandingContextType} from './BrandingContext';
import useGetBrandingResolve from '../../api/useGetBrandingResolve';
import {BrandingType} from '../../models/branding';
import createThemeFromBranding from '../../utils/createThemeFromBranding';
import extractLayoutFromBranding from '../../utils/extractLayoutFromBranding';
import extractImagesFromBranding from '../../utils/extractImagesFromBranding';

/**
 * Props for the BrandingProvider component.
 *
 * @public
 */
export type BrandingProviderProps = PropsWithChildren;

/**
 * React context provider component that provides Thunder branding configuration
 * to all child components.
 *
 * This component loads branding data from the server using the client UUID
 * and provides it through React context. It creates utility methods for common
 * branding operations such as getting the theme, layout, images, and checking
 * branding status.
 *
 * @param props - The component props
 * @param props.children - React children to be wrapped with the branding context
 *
 * @returns JSX element that provides branding context to children
 *
 * @example
 * ```tsx
 * import BrandingProvider from './BrandingProvider';
 * import App from './App';
 *
 * function Root() {
 *   return (
 *     <BrandingProvider>
 *       <App />
 *     </BrandingProvider>
 *   );
 * }
 * ```
 *
 * @public
 */
export default function BrandingProvider({children}: BrandingProviderProps) {
  const {getClientUuid} = useConfig();
  const clientUuid = getClientUuid();

  // Skip branding resolution when no client UUID is available
  const shouldLoadBranding = Boolean(clientUuid && clientUuid.trim().length > 0);

  const {
    data: branding,
    isLoading,
    error,
  } = useGetBrandingResolve(
    {
      id: clientUuid ?? '',
      type: BrandingType.APP,
    },
    {
      enabled: shouldLoadBranding,
    },
  );

  const contextValue: BrandingContextType = useMemo(() => {
    const theme = branding ? createThemeFromBranding(branding) : undefined;
    const images = extractImagesFromBranding(branding);
    const layout = extractLayoutFromBranding(branding);

    return {
      branding,
      isBrandingEnabled: Boolean(branding),
      isLoading,
      error,
      theme,
      images,
      layout,
    };
  }, [branding, isLoading, error]);

  return <BrandingContext.Provider value={contextValue}>{children}</BrandingContext.Provider>;
}
