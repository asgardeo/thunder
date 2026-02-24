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

import {Box, Stack, Typography} from '@wso2/oxygen-ui';
import type {JSX, ReactNode} from 'react';

export interface SectionHeaderProps {
  title: string;
  count: number;
  icon: ReactNode;
  action?: ReactNode;
  comingSoon?: boolean;
}

export default function SectionHeader({title, count, icon, action, comingSoon}: SectionHeaderProps): JSX.Element {
  return (
    <Stack direction="row" alignItems="center" mb={2}>
      <Stack direction="row" alignItems="center" spacing={1} sx={{flex: 1}}>
        {icon}
        <Typography variant="h6" sx={{fontWeight: 600, fontSize: '1rem'}}>
          {title}
        </Typography>
        <Box
          sx={{
            px: 0.75,
            py: 0.1,
            bgcolor: 'action.selected',
            borderRadius: 10,
            minWidth: 22,
            textAlign: 'center',
          }}
        >
          <Typography variant="caption" sx={{fontSize: '0.7rem', fontWeight: 600, color: 'text.secondary'}}>
            {count}
          </Typography>
        </Box>
        {comingSoon && (
          <Box
            sx={{
              px: 1,
              py: 0.15,
              bgcolor: 'warning.light',
              borderRadius: 10,
              border: '1px solid',
              borderColor: 'warning.main',
            }}
          >
            <Typography variant="caption" sx={{fontSize: '0.65rem', fontWeight: 700, color: 'warning.dark', letterSpacing: '0.04em'}}>
              COMING SOON
            </Typography>
          </Box>
        )}
      </Stack>
      {action}
    </Stack>
  );
}
