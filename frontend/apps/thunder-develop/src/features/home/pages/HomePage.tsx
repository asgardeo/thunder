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

import {Box, Divider, Stack, Typography} from '@wso2/oxygen-ui';
import {User} from '@asgardeo/react';
import type {JSX} from 'react';
import AppGenerationPrompt from '../components/AppGenerationPrompt';
import StartCoding from '../components/StartCoding';

export default function HomePage(): JSX.Element {
  return (
    <User>
      {(user: {name?: string; email?: string} | null) => (
        <Box
          sx={{
            minHeight: '80vh',
            display: 'flex',
            flexDirection: 'column',
            justifyContent: 'center',
            alignItems: 'flex-start',
            maxWidth: '960px',
            margin: '0 auto',
            py: 4,
          }}
        >
          <Stack spacing={6} width="100%">
            {/* Welcome Header */}
            <Box>
              <Typography
                variant="h1"
                sx={{
                  fontSize: {xs: '2rem', sm: '2.5rem', md: '3rem'},
                  fontWeight: 500,
                  mb: 1,
                }}
              >
                👋 Hello,{' '}
                <Box
                  component="span"
                  sx={{
                    background: 'linear-gradient(90deg, #FF6B00 0%, #FF8C00 100%)',
                    WebkitBackgroundClip: 'text',
                    WebkitTextFillColor: 'transparent',
                    backgroundClip: 'text',
                  }}
                >
                  {user?.name?.split(' ')[0] ?? 'there'}
                </Box>
              </Typography>
              <Typography
                variant="h4"
                sx={{
                  fontSize: {xs: '1rem', sm: '1.25rem', md: '1.5rem'},
                  fontWeight: 400,
                  color: 'text.secondary',
                }}
              >
                What do you want to secure today?
              </Typography>
            </Box>
            <Box>
              {/* App Generation Prompt */}
              <AppGenerationPrompt />

              <Divider sx={{my: 6}}>Or</Divider>

              {/* Start Coding Section */}
              <StartCoding />
            </Box>
          </Stack>
        </Box>
      )}
    </User>
  );
}
