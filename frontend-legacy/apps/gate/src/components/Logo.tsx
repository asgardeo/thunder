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

import { useColorScheme } from '@oxygen-ui/react/src/hooks/useColorScheme';
import Image from 'next/image';
import { ReactElement } from 'react';

export default function Logo(): ReactElement {
  const { mode } = useColorScheme();

  const logoSrc: string = mode === 'dark' ? '/images/logo-dark.svg' : '/images/logo-light.svg';

  return <Image src={logoSrc} alt="Logo" width={200} height={50} priority />;
}
