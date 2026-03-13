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

import type {JSX} from 'react';
import {Stack} from '@wso2/oxygen-ui';
import {useDesign} from '@thunder/shared-design';
import SignInBox from './SignInBox';
import SignInSlogan from './SignInSlogan';

export default function SignIn(): JSX.Element {
  const {isDesignEnabled} = useDesign();

  const showSlogan = !isDesignEnabled;

  return (
    <Stack
      direction="column"
      component="main"
      sx={[
        {
          justifyContent: 'center',
          height: 'calc((1 - var(--template-frame-height, 0)) * 100%)',
          minHeight: '100%',
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
          {showSlogan && <SignInSlogan />}
          <SignInBox />
        </Stack>
      </Stack>
    </Stack>
  );
}
