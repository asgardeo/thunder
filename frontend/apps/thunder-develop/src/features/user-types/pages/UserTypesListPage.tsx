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
import Box from '@mui/material/Box';
import Stack from '@mui/material/Stack';
import Typography from '@mui/material/Typography';
import Button from '@mui/material/Button';
import {Plus, RefreshCw} from 'lucide-react';
import UserTypesList from '../components/UserTypesList';

export default function UserTypesListPage() {
  const navigate = useNavigate();

  return (
    <Box>
      <Stack direction="row" justifyContent="space-between" alignItems="center" mb={4} flexWrap="wrap" gap={2}>
        <Box>
          <Typography variant="h4" component="h1" gutterBottom>
            User Type Management
          </Typography>
          <Typography variant="body2" color="text.secondary">
            Define and manage user type schemas for your organization
          </Typography>
        </Box>
        <Stack direction="row" spacing={2}>
          <Button variant="outlined" startIcon={<RefreshCw size={16} />} onClick={() => window.location.reload()}>
            Refresh
          </Button>
          <Button
            variant="contained"
            startIcon={<Plus size={18} />}
            onClick={() => {
              const handler = async () => {
                await navigate('/user-types/create');
              };

              handler().catch(() => {
                // TODO: Log the errors
                // Tracker: https://github.com/asgardeo/thunder/issues/618
              });
            }}
          >
            Add User Type
          </Button>
        </Stack>
      </Stack>

      <UserTypesList />
    </Box>
  );
}
