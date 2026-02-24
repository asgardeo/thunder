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
import type {JSX} from 'react';

export interface ScreenListItemProps {
  name: string;
  extendsBase?: string;
  isSelected: boolean;
  onClick: () => void;
}

export default function ScreenListItem({name, extendsBase, isSelected, onClick}: ScreenListItemProps): JSX.Element {
  return (
    <Card
      onClick={onClick}
      sx={{
        cursor: 'pointer',
        borderColor: isSelected ? 'primary.main' : undefined,
        '&:hover': {borderColor: isSelected ? 'primary.main' : 'divider'},
      }}
    >
      <CardContent sx={{display: 'flex', alignItems: 'center', gap: 1, '&:last-child': {pb: 1.25}}}>
        {/* Screen icon bar */}
        <Box
          sx={{
            width: 28,
            height: 36,
            borderRadius: '4px',
            border: '1.5px solid',
            borderColor: isSelected ? 'primary.main' : 'divider',
            bgcolor: isSelected ? 'primary.main' : 'background.paper',
            display: 'flex',
            flexDirection: 'column',
            overflow: 'hidden',
            flexShrink: 0,
            gap: '2px',
            p: '3px',
          }}
        >
          <Box
            sx={{height: 4, bgcolor: isSelected ? 'rgba(255,255,255,0.8)' : 'action.selected', borderRadius: '1px'}}
          />
          <Box sx={{flex: 1, bgcolor: isSelected ? 'rgba(255,255,255,0.25)' : 'grey.100', borderRadius: '1px'}} />
          <Box
            sx={{height: 3, bgcolor: isSelected ? 'rgba(255,255,255,0.5)' : 'action.selected', borderRadius: '1px'}}
          />
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
            {name}
          </Typography>
          <Typography
            variant="caption"
            sx={{
              fontSize: '0.68rem',
              color: isSelected ? 'primary.main' : 'text.secondary',
              opacity: isSelected ? 0.75 : 1,
              display: 'block',
            }}
          >
            {extendsBase ? `extends ${extendsBase}` : 'base screen'}
          </Typography>
        </Box>
      </CardContent>
    </Card>
  );
}
