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

import {User} from '@asgardeo/react';
import Box from '@mui/material/Box';
import Stack from '@mui/material/Stack';
import Typography from '@mui/material/Typography';
import type {JSX} from 'react';

export default function HomePage(): JSX.Element {
  return (
    <Stack
      sx={{
        flexDirection: 'column',
        alignSelf: 'center',
        gap: 4,
        px: 20,
        py: 10,
      }}
    >
      <Box sx={{display: {xs: 'none', md: 'flex'}}}>
        <User fallback={<p>Please sign in</p>}>
          {(user) => (
            <Typography variant="h2">
              👋 Welcome, {user?.givenName} {user?.familyName}!
            </Typography>
          )}
        </User>
      </Box>
    </Stack>
  );
}
