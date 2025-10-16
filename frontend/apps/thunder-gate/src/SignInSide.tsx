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

import Stack from '@mui/material/Stack';
import {alpha} from '@mui/material/styles';
import SignInCard from './components/SignInCard';
import Content from './components/Content';

export type SignInSideProps = Record<string, unknown>;

export default function SignInSide() {
  return (
    <Stack
      direction="column"
      component="main"
      sx={[
        {
          justifyContent: 'center',
          height: 'calc((1 - var(--template-frame-height, 0)) * 100%)',
          marginTop: 'max(40px - var(--template-frame-height, 0px), 0px)',
          minHeight: '100%',
        },
        (theme) => ({
          '&::before': {
            content: '""',
            display: 'block',
            position: 'absolute',
            zIndex: -1,
            inset: 0,
            backgroundImage: `radial-gradient(ellipse at 50% 50%, ${alpha(theme.palette.primary.light, 0.1)}, hsl(0, 0%, 100%))`,
            backgroundRepeat: 'no-repeat',
            ...theme.applyStyles('dark', {
              backgroundImage: `radial-gradient(at 50% 50%, ${alpha(theme.palette.primary.dark, 0.15)}, hsl(220, 30%, 5%))`,
            }),
          },
        }),
      ]}
    >
      <Stack
        direction={{xs: 'column-reverse', md: 'row'}}
        sx={{
          justifyContent: 'center',
          gap: {xs: 6, sm: 12},
          p: 2,
          mx: 'auto',
        }}
      >
        <Stack
          direction={{xs: 'column-reverse', md: 'row'}}
          sx={{
            justifyContent: 'center',
            gap: {xs: 6, sm: 12},
            p: {xs: 2, sm: 4},
            m: 'auto',
          }}
        >
          <Content />
          <SignInCard />
        </Stack>
      </Stack>
    </Stack>
  );
}
