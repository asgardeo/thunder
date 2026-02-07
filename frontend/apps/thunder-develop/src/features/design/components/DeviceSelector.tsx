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
import {ButtonGroup, Button, Tooltip, Box} from '@wso2/oxygen-ui';
import {Monitor, Tablet, Smartphone} from '@wso2/oxygen-ui-icons-react';
import type {DeviceType} from '../models/design';
import {DEVICE_VIEWPORTS} from '../models/design';

/**
 * Props for DeviceSelector component
 */
interface DeviceSelectorProps {
  /** Currently selected device */
  selectedDevice: DeviceType;
  /** Callback when device selection changes */
  onDeviceChange: (device: DeviceType) => void;
}

/**
 * Device selector component for responsive preview
 */
export default function DeviceSelector({selectedDevice, onDeviceChange}: DeviceSelectorProps): JSX.Element {
  const devices: Array<{type: DeviceType; icon: JSX.Element; label: string}> = [
    {type: 'desktop', icon: <Monitor size={18} />, label: DEVICE_VIEWPORTS.desktop.label},
    {type: 'tablet', icon: <Tablet size={18} />, label: DEVICE_VIEWPORTS.tablet.label},
    {type: 'mobile', icon: <Smartphone size={18} />, label: DEVICE_VIEWPORTS.mobile.label},
  ];

  return (
    <Box>
      <ButtonGroup variant="outlined" size="small">
        {devices.map(({type, icon, label}) => (
          <Tooltip key={type} title={`${label} (${DEVICE_VIEWPORTS[type].width}x${DEVICE_VIEWPORTS[type].height})`}>
            <Button
              variant={selectedDevice === type ? 'contained' : 'outlined'}
              onClick={() => onDeviceChange(type)}
              sx={{minWidth: 48}}
            >
              {icon}
            </Button>
          </Tooltip>
        ))}
      </ButtonGroup>
    </Box>
  );
}
