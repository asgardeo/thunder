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

import {useNavigate} from 'react-router';
import {Box, Stack, Typography, Button} from '@wso2/oxygen-ui';
import {Plus} from '@wso2/oxygen-ui-icons-react';

export default function FlowsListPage() {
  const navigate = useNavigate();

  return (
    <Box>
      <Stack direction="row" justifyContent="space-between" alignItems="center" mb={4} flexWrap="wrap" gap={2}>
        <Box>
          <Typography variant="h4" component="h1" gutterBottom>
            Login Flows
          </Typography>
          <Typography variant="body2" color="text.secondary">
            Create and manage authentication flows for your applications
          </Typography>
        </Box>
        <Stack direction="row" spacing={2}>
          <Button
            variant="contained"
            startIcon={<Plus size={18} />}
            onClick={() => {
              const handler = async () => {
                await navigate('/flows/login-builder');
              };

              handler().catch(() => {
                // TODO: Log the errors
              });
            }}
          >
            Create New Flow
          </Button>
        </Stack>
      </Stack>

      {/* <FlowsList /> */}
    </Box>
  );
}
