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

import Box from '@mui/material/Box';
import Typography from '@mui/material/Typography';
import Paper from '@mui/material/Paper';

export default function IntegrationsPage() {
  return (
    <Box>
      <Box mb={4}>
        <Typography variant="h4" component="h1" gutterBottom>
          Integrations
        </Typography>
        <Typography variant="body2" color="text.secondary">
          Manage your integrations and connections
        </Typography>
      </Box>

      <Paper
        sx={{
          p: 8,
          textAlign: 'center',
          minHeight: 400,
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
        }}
      >
        <Box>
          <Typography variant="h3" component="h2" gutterBottom color="text.secondary" fontWeight={500}>
            Coming Soon
          </Typography>
          <Typography variant="body1" color="text.secondary">
            Integrations management functionality will be available soon.
          </Typography>
        </Box>
      </Paper>
    </Box>
  );
}
