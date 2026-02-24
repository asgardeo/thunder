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

import {MenuItem, Select, Stack, Typography} from '@wso2/oxygen-ui';
import type {JSX} from 'react';
import type {Theme} from '@thunder/shared-design';
import ConfigCard from '../common/ConfigCard';

export interface GeneralBuilderContentProps {
  draft: Theme;
  onUpdate: (updater: (d: Theme) => void) => void;
}

/**
 * GeneralBuilderContent - Theme builder section for internationalization settings.
 * Configures text direction (LTR/RTL).
 */
export default function GeneralBuilderContent({draft, onUpdate}: GeneralBuilderContentProps): JSX.Element {
  return (
    <ConfigCard title="Internationalization">
      <Stack direction="row" alignItems="center" justifyContent="space-between" sx={{py: 0.75}}>
        <Typography variant="body2" sx={{fontWeight: 500, fontSize: '0.875rem'}}>
          Text direction
        </Typography>
        <Select
          value={draft.direction ?? 'ltr'}
          onChange={(e) =>
            onUpdate((d) => {
              d.direction = String(e.target.value);
            })
          }
          size="small"
          sx={{fontSize: '0.8125rem', height: 36, minWidth: 90}}
        >
          <MenuItem value="ltr" sx={{fontSize: '0.8125rem'}}>
            LTR
          </MenuItem>
          <MenuItem value="rtl" sx={{fontSize: '0.8125rem'}}>
            RTL
          </MenuItem>
        </Select>
      </Stack>
    </ConfigCard>
  );
}
