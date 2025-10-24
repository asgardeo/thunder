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
import TextField from '@mui/material/TextField';
import Button from '@mui/material/Button';
import InputAdornment from '@mui/material/InputAdornment';
import SearchIcon from '@mui/icons-material/Search';
import AddIcon from '@mui/icons-material/Add';
import RefreshIcon from '@mui/icons-material/Refresh';
import Select from '@mui/material/Select';
import MenuItem from '@mui/material/MenuItem';
import {useMemo, useState} from 'react';
import UsersList from '../components/UsersList';
import useGetUserSchemas from '../api/useGetUserSchemas';
import type {SchemaInterface} from '../types/users';

export default function UsersListPage() {
  const navigate = useNavigate();

  const [selectedSchema, setSelectedSchema] = useState<string>();

  const {data: originalUserSchemas} = useGetUserSchemas();

  const userSchemas: SchemaInterface[] = useMemo(() => {
    if (!originalUserSchemas?.schemas) {
      return [];
    }

    setSelectedSchema(originalUserSchemas.schemas[0]?.id);

    return originalUserSchemas?.schemas;
  }, [originalUserSchemas]);

  return (
    <Box>
      <Stack direction="row" justifyContent="space-between" alignItems="center" mb={4} flexWrap="wrap" gap={2}>
        <Box>
          <Typography variant="h4" component="h1" gutterBottom>
            User Management
          </Typography>
          <Typography variant="body2" color="text.secondary">
            Manage users, roles, and permissions across your organization
          </Typography>
        </Box>
        <Stack direction="row" spacing={2}>
          <Button variant="outlined" startIcon={<RefreshIcon />} onClick={() => window.location.reload()}>
            Refresh
          </Button>
          <Button variant="contained" startIcon={<AddIcon />} onClick={() => navigate('/users/new')}>
            Add User
          </Button>
        </Stack>
      </Stack>

      {/* Search and Filters */}
      <Stack direction="row" spacing={2} mb={4} flexWrap="wrap" useFlexGap>
        <TextField
          placeholder="Search users..."
          size="small"
          sx={{flexGrow: 1, minWidth: 300}}
          InputProps={{
            startAdornment: (
              <InputAdornment position="start">
                <SearchIcon />
              </InputAdornment>
            ),
          }}
        />
        <Select
          id="user-schema-select"
          label="User Schema"
          value={selectedSchema ?? ''}
          size="small"
          sx={{minWidth: 200}}
          onChange={(e) => setSelectedSchema(e.target.value)}
        >
          {userSchemas.map((schema) => (
            <MenuItem key={schema.id} value={schema.id}>
              {schema.name}
            </MenuItem>
          ))}
        </Select>
      </Stack>
      <UsersList selectedSchema={selectedSchema ?? ''} />
    </Box>
  );
}
