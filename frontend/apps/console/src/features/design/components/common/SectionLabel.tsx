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

import {Typography} from '@wso2/oxygen-ui';
import type {JSX} from 'react';

export interface SectionLabelProps {
  children: string;
}

/**
 * SectionLabel - A styled section heading label.
 * Used to group related configuration options within a config section.
 */
export default function SectionLabel({children}: SectionLabelProps): JSX.Element {
  return (
    <Typography
      variant="caption"
      sx={{
        display: 'block',
        fontWeight: 600,
        fontSize: '0.7rem',
        textTransform: 'uppercase',
        letterSpacing: '0.05em',
        color: 'text.secondary',
        mt: 1.25,
        mb: 0.5,
      }}
    >
      {children}
    </Typography>
  );
}
