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

import {useNavigate, useParams} from 'react-router';
import {
  Box,
  Stack,
  Typography,
  Button,
  Paper,
  Divider,
  CircularProgress,
  Alert,
  Chip,
  Avatar,
} from '@wso2/oxygen-ui';
import {ArrowLeft, AppWindow} from '@wso2/oxygen-ui-icons-react';
import {useTranslation} from 'react-i18next';
import useGetApplication from '../api/useGetApplication';
import type {OAuth2Config} from '../models/oauth';

export default function ViewApplicationPage() {
  const {t} = useTranslation();
  const navigate = useNavigate();
  const {applicationId} = useParams<{applicationId: string}>();

  const {data: application, isLoading, error, isError} = useGetApplication(applicationId ?? '');

  const handleBack = async () => {
    await navigate('/applications');
  };

  // Loading state
  if (isLoading) {
    return (
      <Box sx={{display: 'flex', justifyContent: 'center', alignItems: 'center', minHeight: '400px'}}>
        <CircularProgress />
      </Box>
    );
  }

  // Error state
  if (isError || error) {
    return (
      <Box sx={{maxWidth: 1000, mx: 'auto', px: 2, pt: 6}}>
        <Alert severity="error" sx={{mb: 2}}>
          {error?.message ?? t('applications:view.error')}
        </Alert>
        <Button
          onClick={() => {
            handleBack().catch(() => {
              // Handle navigation error
            });
          }}
          startIcon={<ArrowLeft size={16} />}
        >
          {t('applications:view.back')}
        </Button>
      </Box>
    );
  }

  // No application found
  if (!application) {
    return (
      <Box sx={{maxWidth: 1000, mx: 'auto', px: 2, pt: 6}}>
        <Alert severity="warning" sx={{mb: 2}}>
          {t('applications:view.notFound')}
        </Alert>
        <Button
          onClick={() => {
            handleBack().catch(() => {
              // Handle navigation error
            });
          }}
          startIcon={<ArrowLeft size={16} />}
        >
          {t('applications:view.back')}
        </Button>
      </Box>
    );
  }

  const oauth2Config: OAuth2Config | undefined = application.inbound_auth_config?.find((config) => config.type === 'oauth2')
    ?.config;

  return (
    <Box sx={{maxWidth: 1000, mx: 'auto', px: 2, position: 'relative'}}>
      <Button
        onClick={() => {
          handleBack().catch(() => {
            // Handle navigation error
          });
        }}
        variant="text"
        sx={{mb: 3}}
        aria-label="Go back"
        startIcon={<ArrowLeft size={16} />}
      >
        {t('applications:view.back')}
      </Button>

      <Stack direction="row" alignItems="flex-start" justifyContent="space-between" mb={4} gap={2}>
        <Stack direction="row" spacing={2} alignItems="center">
          <Avatar
            src={application.logo_url}
            slotProps={{
              img: {
                onError: (e: React.SyntheticEvent<HTMLImageElement>) => {
                  e.currentTarget.style.display = 'none';
                },
              },
            }}
            sx={{
              width: 56,
              height: 56,
            }}
          >
            <AppWindow size={24} />
          </Avatar>
          <Box>
            <Typography variant="h4" component="h1" gutterBottom>
              {application.name}
            </Typography>
            <Typography variant="body2" color="text.secondary">
              {t('applications:view.subtitle')}
            </Typography>
          </Box>
        </Stack>
      </Stack>

      <Paper sx={{p: 4}}>
        {/* Basic Information */}
        <Box sx={{mb: 3}}>
          <Typography variant="h6" gutterBottom>
            {t('applications:view.sections.basicInformation')}
          </Typography>
          <Divider sx={{mb: 2}} />
          <Stack spacing={2}>
            <Box>
              <Typography variant="caption" color="text.secondary">
                {t('applications:view.fields.applicationId')}
              </Typography>
              <Typography variant="body1" sx={{fontFamily: 'monospace', fontSize: '0.875rem'}}>
                {application.id}
              </Typography>
            </Box>
            {application.description && (
              <Box>
                <Typography variant="caption" color="text.secondary">
                  {t('applications:view.fields.description')}
                </Typography>
                <Typography variant="body1">{application.description}</Typography>
              </Box>
            )}
            {application.url && (
              <Box>
                <Typography variant="caption" color="text.secondary">
                  {t('applications:view.fields.url')}
                </Typography>
                <Typography variant="body1">
                  <a href={application.url} target="_blank" rel="noopener noreferrer">
                    {application.url}
                  </a>
                </Typography>
              </Box>
            )}
            {application.tos_uri && (
              <Box>
                <Typography variant="caption" color="text.secondary">
                  {t('applications:view.fields.tosUri')}
                </Typography>
                <Typography variant="body1">
                  <a href={application.tos_uri} target="_blank" rel="noopener noreferrer">
                    {application.tos_uri}
                  </a>
                </Typography>
              </Box>
            )}
            {application.policy_uri && (
              <Box>
                <Typography variant="caption" color="text.secondary">
                  {t('applications:view.fields.policyUri')}
                </Typography>
                <Typography variant="body1">
                  <a href={application.policy_uri} target="_blank" rel="noopener noreferrer">
                    {application.policy_uri}
                  </a>
                </Typography>
              </Box>
            )}
            {application.contacts && application.contacts.length > 0 && (
              <Box>
                <Typography variant="caption" color="text.secondary">
                  {t('applications:view.fields.contacts')}
                </Typography>
                <Stack direction="row" spacing={1} sx={{mt: 0.5}}>
                  {application.contacts.map((contact) => (
                    <Chip key={contact} label={contact} size="small" variant="outlined" />
                  ))}
                </Stack>
              </Box>
            )}
          </Stack>
        </Box>

        <Divider sx={{my: 3}} />

        {/* Flow Configuration */}
        <Box sx={{mb: 3}}>
          <Typography variant="h6" gutterBottom>
            {t('applications:view.sections.flowConfiguration')}
          </Typography>
          <Divider sx={{mb: 2}} />
          <Stack spacing={2}>
            {application.auth_flow_id && (
              <Box>
                <Typography variant="caption" color="text.secondary">
                  {t('applications:view.fields.authFlowId')}
                </Typography>
                <Typography variant="body1">{application.auth_flow_id}</Typography>
              </Box>
            )}
            {application.registration_flow_id && (
              <Box>
                <Typography variant="caption" color="text.secondary">
                  {t('applications:view.fields.registrationFlowId')}
                </Typography>
                <Typography variant="body1">{application.registration_flow_id}</Typography>
              </Box>
            )}
            <Box>
              <Typography variant="caption" color="text.secondary">
                {t('applications:view.fields.registrationFlowEnabled')}
              </Typography>
              <Typography variant="body1">
                {application.is_registration_flow_enabled ? t('applications:view.values.yes') : t('applications:view.values.no')}
              </Typography>
            </Box>
          </Stack>
        </Box>

        <Divider sx={{my: 3}} />

        {/* User Attributes */}
        {application.user_attributes && application.user_attributes.length > 0 && (
          <>
            <Box sx={{mb: 3}}>
              <Typography variant="h6" gutterBottom>
                {t('applications:view.sections.userAttributes')}
              </Typography>
              <Divider sx={{mb: 2}} />
              <Stack direction="row" spacing={1} flexWrap="wrap" useFlexGap>
                {application.user_attributes.map((attr) => (
                  <Chip key={attr} label={attr} size="small" variant="outlined" />
                ))}
              </Stack>
            </Box>
            <Divider sx={{my: 3}} />
          </>
        )}

        {/* OAuth2 Configuration */}
        {oauth2Config && (
          <Box sx={{mb: 3}}>
            <Typography variant="h6" gutterBottom>
              {t('applications:view.sections.oauth2Configuration')}
            </Typography>
            <Divider sx={{mb: 2}} />
            <Stack spacing={2}>
              {oauth2Config.client_id && (
                <Box>
                  <Typography variant="caption" color="text.secondary">
                    {t('applications:view.fields.clientId')}
                  </Typography>
                  <Typography variant="body1" sx={{fontFamily: 'monospace', fontSize: '0.875rem'}}>
                    {oauth2Config.client_id}
                  </Typography>
                </Box>
              )}
              {oauth2Config.redirect_uris && oauth2Config.redirect_uris.length > 0 && (
                <Box>
                  <Typography variant="caption" color="text.secondary">
                    {t('applications:view.fields.redirectUris')}
                  </Typography>
                  <Stack spacing={0.5} sx={{mt: 0.5}}>
                    {oauth2Config.redirect_uris.map((uri) => (
                      <Typography
                        key={uri}
                        variant="body2"
                        sx={{fontFamily: 'monospace', fontSize: '0.875rem'}}
                      >
                        {uri}
                      </Typography>
                    ))}
                  </Stack>
                </Box>
              )}
              {oauth2Config.grant_types && oauth2Config.grant_types.length > 0 && (
                <Box>
                  <Typography variant="caption" color="text.secondary">
                    {t('applications:view.fields.grantTypes')}
                  </Typography>
                  <Stack direction="row" spacing={1} sx={{mt: 0.5}} flexWrap="wrap" useFlexGap>
                    {oauth2Config.grant_types.map((grant) => (
                      <Chip key={grant} label={grant} size="small" variant="outlined" />
                    ))}
                  </Stack>
                </Box>
              )}
              {oauth2Config.response_types && oauth2Config.response_types.length > 0 && (
                <Box>
                  <Typography variant="caption" color="text.secondary">
                    {t('applications:view.fields.responseTypes')}
                  </Typography>
                  <Stack direction="row" spacing={1} sx={{mt: 0.5}} flexWrap="wrap" useFlexGap>
                    {oauth2Config.response_types.map((response) => (
                      <Chip key={response} label={response} size="small" variant="outlined" />
                    ))}
                  </Stack>
                </Box>
              )}
              {oauth2Config.scopes && oauth2Config.scopes.length > 0 && (
                <Box>
                  <Typography variant="caption" color="text.secondary">
                    {t('applications:view.fields.scopes')}
                  </Typography>
                  <Stack direction="row" spacing={1} sx={{mt: 0.5}} flexWrap="wrap" useFlexGap>
                    {oauth2Config.scopes.map((scope) => (
                      <Chip key={scope} label={scope} size="small" variant="outlined" />
                    ))}
                  </Stack>
                </Box>
              )}
              {oauth2Config.public_client !== undefined && (
                <Box>
                  <Typography variant="caption" color="text.secondary">
                    {t('applications:view.fields.publicClient')}
                  </Typography>
                  <Typography variant="body1">{oauth2Config.public_client ? t('applications:view.values.yes') : t('applications:view.values.no')}</Typography>
                </Box>
              )}
              {oauth2Config.pkce_required !== undefined && (
                <Box>
                  <Typography variant="caption" color="text.secondary">
                    {t('applications:view.fields.pkceRequired')}
                  </Typography>
                  <Typography variant="body1">{oauth2Config.pkce_required ? t('applications:view.values.yes') : t('applications:view.values.no')}</Typography>
                </Box>
              )}
            </Stack>
          </Box>
        )}

        {/* Timestamps */}
        {(application.created_at ?? application.updated_at) && (
          <>
            <Divider sx={{my: 3}} />
            <Box>
              <Typography variant="h6" gutterBottom>
                {t('applications:view.sections.timestamps')}
              </Typography>
              <Divider sx={{mb: 2}} />
              <Stack spacing={2}>
                {application.created_at && (
                  <Box>
                    <Typography variant="caption" color="text.secondary">
                      {t('applications:view.fields.createdAt')}
                    </Typography>
                    <Typography variant="body1">
                      {new Date(application.created_at).toLocaleString()}
                    </Typography>
                  </Box>
                )}
                {application.updated_at && (
                  <Box>
                    <Typography variant="caption" color="text.secondary">
                      {t('applications:view.fields.updatedAt')}
                    </Typography>
                    <Typography variant="body1">
                      {new Date(application.updated_at).toLocaleString()}
                    </Typography>
                  </Box>
                )}
              </Stack>
            </Box>
          </>
        )}
      </Paper>
    </Box>
  );
}

