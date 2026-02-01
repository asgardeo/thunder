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

import React, {useEffect, type ReactNode} from 'react';
import ColorModeToggle from '@theme-original/ColorModeToggle';
import type ColorModeToggleType from '@theme/ColorModeToggle';
import type {WrapperProps} from '@docusaurus/types';
import {useColorScheme} from '@wso2/oxygen-ui';

type Props = WrapperProps<typeof ColorModeToggleType>;

export default function ColorModeToggleWrapper(props: Props): ReactNode {
  // MUI color mode setting
  const {setMode} = useColorScheme();

  // "value" holds the color theme. Either "light" or "dark"
  const {value} = props;

  // change mode based on "value" prop
  // "dark" or "light" are also used for MUI
  useEffect(() => {
    setMode(value);
  }, [setMode, value]);

  return <ColorModeToggle {...props} />;
}
