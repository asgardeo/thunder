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
import {Box, Stack, Typography, TextField, Button, InputAdornment, Select, MenuItem} from '@wso2/oxygen-ui';
import {useMemo, useState} from 'react';
import {Plus, Search, Mail} from '@wso2/oxygen-ui-icons-react';
import {useTranslation} from 'react-i18next';
import {useLogger} from '@thunder/logger/react';
import UsersList from '../components/UsersList';
import useGetUserTypes from '../api/useGetUserTypes';
import type {UserTypeListItem} from '../types/users';
import InviteUserDialog from '../components/InviteUserDialog';

export default function UsersListPage() {
  const navigate = useNavigate();
  const {t} = useTranslation();
  const logger = useLogger('UsersListPage');

  const [selectedUserTypeId, setSelectedUserTypeId] = useState<string>();
  const [isInviteDialogOpen, setIsInviteDialogOpen] = useState(false);

  const {data: originalUserTypes} = useGetUserTypes();

  const userTypes: UserTypeListItem[] = useMemo(() => {
    if (!originalUserTypes?.schemas) {
      return [];
    }

    setSelectedUserTypeId(originalUserTypes.schemas[0]?.id);

    return originalUserTypes?.schemas;
  }, [originalUserTypes]);

  const handleOpenInviteDialog = () => {
    setIsInviteDialogOpen(true);
  };

  const handleCloseInviteDialog = () => {
    setIsInviteDialogOpen(false);
  };

  const handleInviteSuccess = (inviteLink: string) => {
    logger.info('Invite link generated successfully', {inviteLink});
  };

  return (
    <Box>
      <Stack direction="row" justifyContent="space-between" alignItems="center" mb={4} flexWrap="wrap" gap={2}>
        <Box>
          <Typography variant="h1" gutterBottom>
            {t('users:title')}
          </Typography>
          <Typography variant="subtitle1" color="text.secondary">
            {t('users:subtitle')}
          </Typography>
        </Box>
        <Stack direction="row" spacing={2}>
          <Button variant="outlined" startIcon={<Mail size={18} />} onClick={handleOpenInviteDialog}>
            {t('users:inviteUser', 'Invite User')}
          </Button>
          <Button
            variant="contained"
            startIcon={<Plus size={18} />}
            onClick={() => {
              (async () => {
                await navigate('/users/create');
              })().catch((error: unknown) => {
                logger.error('Failed to navigate to create user page', {error});
              });
            }}
          >
            {t('users:addUser')}
          </Button>
        </Stack>
      </Stack>

      {/* Search and Filters */}
      <Stack direction="row" spacing={2} mb={4} flexWrap="wrap" useFlexGap>
        <TextField
          placeholder={t('users:searchUsers')}
          size="small"
          sx={{flexGrow: 1, minWidth: 300}}
          InputProps={{
            startAdornment: (
              <InputAdornment position="start">
                <Search size={16} />
              </InputAdornment>
            ),
          }}
        />
        <Select
          id="user-type-select"
          value={selectedUserTypeId ?? ''}
          size="small"
          sx={{minWidth: 200}}
          onChange={(e) => setSelectedUserTypeId(e.target.value)}
        >
          {userTypes.map((userType) => (
            <MenuItem key={userType.id} value={userType.id}>
              {userType.name}
            </MenuItem>
          ))}
        </Select>
      </Stack>
      <UsersList selectedUserTypeId={selectedUserTypeId ?? ''} />

      {/* User Onboarding Dialog */}
      <InviteUserDialog open={isInviteDialogOpen} onClose={handleCloseInviteDialog} onSuccess={handleInviteSuccess} />
    </Box>
  );
}
