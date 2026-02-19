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

import {Box, Typography} from '@wso2/oxygen-ui';
import {Plus} from '@wso2/oxygen-ui-icons-react';
import type {JSX} from 'react';

export interface AddCardProps {
  label: string;
  onClick: () => void;
}

export default function AddCard({label, onClick}: AddCardProps): JSX.Element {
  return (
    <Box
      onClick={onClick}
      sx={{
        cursor: 'pointer',
        borderRadius: 1,
        border: '1.5px dashed',
        borderColor: 'divider',
        overflow: 'hidden',
        aspectRatio: '4/3',
        display: 'flex',
        flexDirection: 'column',
        alignItems: 'center',
        justifyContent: 'center',
        gap: 0.75,
        color: 'text.disabled',
        transition: 'all 0.18s ease',
        '&:hover': {
          borderColor: 'primary.main',
          color: 'primary.main',
          bgcolor: 'primary.50',
        },
      }}
    >
      <Plus size={20} />
      <Typography variant="caption" sx={{fontSize: '0.75rem', fontWeight: 500}}>
        {label}
      </Typography>
    </Box>
  );
}
