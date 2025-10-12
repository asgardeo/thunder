/**
 * Copyright (c) 2025, WSO2 LLC. (https://www.wso2.com).
 *
 * WSO2 LLC. licenses this file to you under the Apache License,
 * Version 2.0 (the "License"); you may not use this file except
 * in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */

import Box from '@oxygen-ui/react/src/components/Box/Box';
import ThemeToggle from '@oxygen-ui/react/src/components/ThemeToggle/ThemeToggle';
import Grid from '@oxygen-ui/react/src/components/Grid/Grid';
import Paper from '@oxygen-ui/react/src/components/Paper/Paper';
import Typography from '@oxygen-ui/react/src/components/Typography/Typography';
import AppConfig from '@/configs/app.json';
import BaseLayout from '@/layouts/base';
import SideImage from '@/images/layout-image';
import type { Metadata } from 'next';
import { ReactElement } from 'react';

export const metadata: Metadata = {
  title: 'Gate',
  description: 'This the gate of your app',
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>): ReactElement {
  return (
    <BaseLayout>
      <Box sx={{ height: '100vh', display: 'flex' }}>
        <Grid container sx={{ flex: 1 }}>
          <Grid
            size={{ xs: 12, md: 6 }}
            sx={{
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              padding: 4,
            }}
          >
            <Box>
              <SideImage />
              <Typography variant="h4" sx={{ mb: 2 }}>
                Welcome!
              </Typography>
              <Typography variant="body1">
                This login is powered by {AppConfig.productName}.<br />
                Which empowers developers to implement login experiences in no time.
              </Typography>
            </Box>
          </Grid>

          <Grid size={{ xs: 12, md: 6 }}>
            <Paper
              sx={{
                display: 'flex',
                padding: 4,
                width: '100%',
                height: '100%',
                flexDirection: 'column',
                position: 'relative',
              }}
            >
              <Box sx={{ position: 'absolute', right: '4rem' }}>
                <ThemeToggle />
              </Box>
              <Box
                sx={{
                  alignItems: 'center',
                  justifyContent: 'center',
                  padding: 4,
                  width: '100%',
                  maxWidth: 500,
                  margin: 'auto',
                }}
              >
                <Box>
                  {children}
                  <Box component="footer" sx={{ mt: 10 }}>
                    <Typography sx={{ textAlign: 'center' }}>© Copyright {new Date().getFullYear()}</Typography>
                  </Box>
                </Box>
              </Box>
            </Paper>
          </Grid>
        </Grid>
      </Box>
    </BaseLayout>
  );
}
