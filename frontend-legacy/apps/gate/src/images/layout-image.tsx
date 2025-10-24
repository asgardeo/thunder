/**
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
 * KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */

'use client';

import Box from '@oxygen-ui/react/src/components/Box/Box';
import { useColorScheme } from '@oxygen-ui/react/src/hooks/useColorScheme';
import { useTheme } from '@oxygen-ui/react/src/hooks/useTheme';
import Image from 'next/image';
import { ReactNode } from 'react';

const SideImage = (): ReactNode => {
  const theme = useTheme();
  const { mode, systemMode } = useColorScheme();

  const resolvedMode = mode === 'system' ? systemMode : mode;
  const imageSrc = resolvedMode === 'dark' ? '/images/login/image-light.svg' : '/images/login/image-light.svg';

  return (
    <Box
      sx={{
        position: 'relative',
        width: '100%',
        maxWidth: '500px',
        height: '520px',
        overflow: 'hidden',
        [theme.breakpoints.down('md')]: {
          maxWidth: '100px',
          height: '120px',
        },
      }}
    >
      <Image src={imageSrc} fill alt="Login Page Side Image" priority />
    </Box>
  );
};

export default SideImage;
