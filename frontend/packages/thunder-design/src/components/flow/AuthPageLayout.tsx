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

import {cn} from '@thunder/utils';
import {CircularProgress, Stack} from '@wso2/oxygen-ui';
import type {JSX, PropsWithChildren} from 'react';

export interface AuthPageLayoutProps extends PropsWithChildren {
  /** Class name prefix for Thunder CSS (e.g. 'SignIn' → 'ThunderSignIn--root'). */
  variant?: string;
  /** Optional background CSS value (color, gradient, image). Falls back to the theme's background.default. */
  background?: string;
  /** Indicates whether the design is still loading. When true, the layout will show a spinner. */
  isLoading?: boolean;
}

/**
 * Shared page-level layout for authentication screens (sign-in, sign-up, accept-invite).
 *
 * Provides the full-page centering structure used by the Gate app. Both the Gate and
 * Console preview render this component so they are visually identical.
 */
export default function AuthPageLayout({
  variant = undefined,
  background = undefined,
  children,
  isLoading = false,
}: AuthPageLayoutProps): JSX.Element {
  return (
    <Stack
      direction="column"
      component="main"
      className={variant ? cn(`${variant}--root`) : undefined}
      sx={[
        {
          justifyContent: 'center',
          height: 'calc((1 - var(--template-frame-height, 0)) * 100%)',
          minHeight: '100%',
          ...(background ? {backgroundColor: background} : {}),
        },
      ]}
    >
      <Stack
        direction={{xs: 'row-reverse', md: 'row'}}
        sx={{
          justifyContent: 'center',
          gap: {xs: 6, sm: 12},
          p: 2,
          mx: 'auto',
        }}
      >
        <Stack
          direction={{xs: 'column', md: 'row'}}
          sx={{
            justifyContent: 'center',
            gap: {xs: 4, sm: 16},
            p: {xs: 2, sm: 4},
            m: 'auto',
          }}
        >
          {isLoading ? <CircularProgress /> : children}
        </Stack>
      </Stack>
    </Stack>
  );
}
