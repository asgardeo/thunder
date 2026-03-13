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

import {useNavigate, useSearchParams} from 'react-router';
import {Stack, TextField, Button, InputAdornment, PageContent, PageTitle} from '@wso2/oxygen-ui';
import {useState, useEffect} from 'react';
import {Plus, Search, Mail} from '@wso2/oxygen-ui-icons-react';
import {useTranslation} from 'react-i18next';
import {useLogger} from '@thunder/logger/react';
import UsersList from '../components/UsersList';
import InviteUserDialog from '../components/InviteUserDialog';

export default function UsersListPage() {
  const navigate = useNavigate();
  const {t} = useTranslation();
  const logger = useLogger('UsersListPage');
  const [searchParams, setSearchParams] = useSearchParams();

  const [isInviteDialogOpen, setIsInviteDialogOpen] = useState(false);

  useEffect(() => {
    if (searchParams.get('invite') === 'true') {
      setIsInviteDialogOpen(true);
      setSearchParams({}, {replace: true});
    }
  }, [searchParams, setSearchParams]);

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
    <PageContent>
      {/* Header */}
      <PageTitle>
        <PageTitle.Header>{t('users:title')}</PageTitle.Header>
        <PageTitle.SubHeader>{t('users:subtitle')}</PageTitle.SubHeader>
        <PageTitle.Actions>
          <Stack direction="row" spacing={2}>
            <Button variant="outlined" startIcon={<Mail size={18} />} onClick={handleOpenInviteDialog}>
              {t('users:inviteUser', 'Invite User')}
            </Button>
            <Button
              variant="contained"
              startIcon={<Plus size={20} />}
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
        </PageTitle.Actions>
      </PageTitle>

      {/* Search */}
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
      </Stack>
      <UsersList />

      {/* User Onboarding Dialog */}
      <InviteUserDialog open={isInviteDialogOpen} onClose={handleCloseInviteDialog} onSuccess={handleInviteSuccess} />
    </PageContent>
  );
}
