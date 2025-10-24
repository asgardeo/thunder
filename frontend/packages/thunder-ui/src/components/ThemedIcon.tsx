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

import Box, {BoxProps} from '@mui/material/Box';
import {useColorScheme} from '@mui/material/styles';
import useMediaQuery from '@mui/material/useMediaQuery';
import {JSX} from 'react';

/**
 * Represents attributes for different color schemes (light and dark modes).
 */
export interface ColorSchemeAttr {
  /** The value for light mode */
  light: string;

  /** The value for dark mode */
  dark: string;
}

/**
 * Props for the ThemedIcon component.
 * Extends BoxProps to inherit all Box component properties.
 */
export interface ThemedIconProps extends BoxProps {
  /** Source URLs for the icon in different color schemes */
  src: ColorSchemeAttr;

  /** Alt text for the icon, can be a string or different text for each color scheme */
  alt: ColorSchemeAttr | string;

  /** Width of the icon in pixels or CSS units */
  width?: number | string;

  /** Height of the icon in pixels or CSS units */
  height?: number | string;
}

/**
 * A themed icon component that automatically switches between light and dark mode variants
 * based on the current color scheme. Supports both automatic system preference detection
 * and manual theme selection.
 *
 * @param props - The component props
 * @param props.src - Object containing light and dark mode image sources
 * @param props.alt - Alt text for accessibility, can be a string or object with light/dark variants
 * @param props.width - Icon width, defaults to 16px
 * @param props.height - Icon height, defaults to 16px
 * @param props.rest - Additional props passed to the underlying Box component
 * @returns A themed icon component that adapts to the current color scheme
 *
 * @example
 * ```tsx
 * <ThemedIcon
 *   src={{
 *     light: '/icons/logo-light.svg',
 *     dark: '/icons/logo-dark.svg'
 *   }}
 *   alt="Company Logo"
 *   width={24}
 *   height={24}
 * />
 * ```
 *
 * @example
 * ```tsx
 * <ThemedIcon
 *   src={{
 *     light: '/icons/home-light.png',
 *     dark: '/icons/home-dark.png'
 *   }}
 *   alt={{
 *     light: 'Home icon (light mode)',
 *     dark: 'Home icon (dark mode)'
 *   }}
 * />
 * ```
 */
export default function ThemedIcon({src, alt, width = 16, height = 16, ...rest}: ThemedIconProps): JSX.Element {
  const {mode} = useColorScheme();

  const prefersDark = useMediaQuery('(prefers-color-scheme: dark)');

  let actualMode: string | undefined;

  if (mode === 'system') {
    actualMode = prefersDark ? 'dark' : 'light';
  } else {
    actualMode = mode;
  }

  const imageSrc: string = actualMode === 'dark' ? src.dark : src.light;
  let imageAlt: string;

  if (typeof alt === 'string') {
    imageAlt = alt;
  } else {
    imageAlt = actualMode === 'dark' ? alt.dark : alt.light;
  }

  return <Box component="img" src={imageSrc} alt={imageAlt} width={width} height={height} {...rest} />;
}
