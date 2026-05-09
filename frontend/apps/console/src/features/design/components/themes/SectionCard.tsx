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

import {Box, Card, CardContent, Typography} from '@wso2/oxygen-ui';
import type {JSX, ReactNode} from 'react';

export interface SectionCardProps {
  label: string;
  description: string;
  icon: ReactNode;
  isSelected: boolean;
  onClick: () => void;
}

export default function SectionCard({label, description, icon, isSelected, onClick}: SectionCardProps): JSX.Element {
  return (
    <Card
      onClick={onClick}
      sx={{
        cursor: 'pointer',
        borderColor: isSelected ? 'primary.main' : 'divider',
        '&:hover': {borderColor: isSelected ? 'primary.main' : 'divider'},
      }}
    >
      <CardContent sx={{display: 'flex', alignItems: 'center', gap: 1.5, '&:last-child': {pb: 1.5}}}>
        <Box
          sx={{
            width: 32,
            height: 32,
            borderRadius: 1.5,
            bgcolor: isSelected ? 'primary.main' : 'action.selected',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            flexShrink: 0,
            color: isSelected ? 'primary.contrastText' : 'text.secondary',
          }}
        >
          {icon}
        </Box>
        <Box sx={{minWidth: 0}}>
          <Typography
            variant="body2"
            sx={{
              fontWeight: isSelected ? 600 : 500,
              fontSize: '0.8125rem',
              color: isSelected ? 'primary.main' : 'text.primary',
              lineHeight: 1.3,
            }}
          >
            {label}
          </Typography>
          <Typography
            variant="caption"
            sx={{
              fontSize: '0.7rem',
              color: isSelected ? 'primary.main' : 'text.secondary',
              opacity: isSelected ? 0.75 : 1,
              display: 'block',
              lineHeight: 1.3,
            }}
          >
            {description}
          </Typography>
        </Box>
      </CardContent>
    </Card>
  );
}
