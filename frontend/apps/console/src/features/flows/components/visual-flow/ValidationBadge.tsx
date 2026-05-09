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

import {keyframes} from '@emotion/react';
import {Box, ButtonBase} from '@wso2/oxygen-ui';
import {CircleXIcon, TriangleAlertIcon} from '@wso2/oxygen-ui-icons-react';
import type {ReactElement} from 'react';
import useValidationStatus from '../../hooks/useValidationStatus';

const pulseError = keyframes`
  0% { box-shadow: 0 0 0 0 rgba(var(--oxygen-palette-error-mainChannel) / 1); }
  70% { box-shadow: 0 0 0 10px rgba(var(--oxygen-palette-error-mainChannel) / 0); }
  100% { box-shadow: 0 0 0 0 rgba(var(--oxygen-palette-error-mainChannel) / 0); }
`;

const pulseWarning = keyframes`
  0% { box-shadow: 0 0 0 0 rgba(var(--oxygen-palette-warning-mainChannel) / 1); }
  70% { box-shadow: 0 0 0 10px rgba(var(--oxygen-palette-warning-mainChannel) / 0); }
  100% { box-shadow: 0 0 0 0 rgba(var(--oxygen-palette-warning-mainChannel) / 0); }
`;

export interface ValidationBadgeProps {
  errorCount: number;
  warningCount: number;
}

export default function ValidationBadge({errorCount, warningCount}: ValidationBadgeProps): ReactElement | null {
  const {setCurrentActiveTab, setOpenValidationPanel} = useValidationStatus();

  if (errorCount === 0 && warningCount === 0) {
    return null;
  }

  const hasErrors = errorCount > 0;

  return (
    <ButtonBase
      onClick={() => {
        setCurrentActiveTab?.(hasErrors ? 0 : 1);
        setOpenValidationPanel?.(true);
      }}
      aria-label={`${errorCount} errors, ${warningCount} warnings. Open validation panel.`}
      sx={{
        display: 'flex',
        alignItems: 'center',
        gap: 1,
        px: 1.5,
        height: 36.5,
        boxSizing: 'border-box',
        borderRadius: 1,
        border: '1px solid',
        borderColor: hasErrors ? 'error.main' : 'warning.main',
        cursor: 'pointer',
        animation: hasErrors ? `${pulseError} 1s infinite` : `${pulseWarning} 1s infinite`,
      }}
    >
      {errorCount > 0 && (
        <Box sx={{display: 'flex', alignItems: 'center', gap: 0.5, color: 'error.main'}}>
          <CircleXIcon size={14} />
          <span style={{fontSize: '0.75rem', fontWeight: 600}}>{errorCount}</span>
        </Box>
      )}
      {warningCount > 0 && (
        <Box sx={{display: 'flex', alignItems: 'center', gap: 0.5, color: 'warning.main'}}>
          <TriangleAlertIcon size={14} />
          <span style={{fontSize: '0.75rem', fontWeight: 600}}>{warningCount}</span>
        </Box>
      )}
    </ButtonBase>
  );
}
