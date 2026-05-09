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

import {Typography, Box} from '@wso2/oxygen-ui';
import {type ReactElement} from 'react';
import type {Resource} from '../../../../models/resources';

/**
 * Props interface for TimerAdapter
 */
export interface TimerAdapterPropsInterface {
  resource?: Resource;
}

/**
 * A canvas placeholder for the Timer element. Features a simulated timer value
 * based on the configured text value replacing the `{time}` dynamic placeholder.
 *
 * @param props - Custom props containing the resource.
 * @returns The TimerAdapter placeholder component.
 */
function TimerAdapter({resource = undefined}: TimerAdapterPropsInterface): ReactElement {
  // Extract text from resource label, default to generic string if missing
  const templateText = (resource as {label?: string})?.label ?? 'Time remaining: {time}';

  // Replace backend dynamic variable format `{time}` with a fake canvas placeholder `05:00`
  const displayText = templateText.replace('{time}', '05:00');

  return (
    <Box sx={{width: '100%', py: 1}}>
      <Typography variant="body2" color="textSecondary" sx={{fontFamily: 'monospace'}}>
        {displayText}
      </Typography>
    </Box>
  );
}

export default TimerAdapter;
