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

/* eslint-disable @typescript-eslint/no-unsafe-member-access */
/* eslint-disable @typescript-eslint/no-explicit-any */
/* eslint-disable @typescript-eslint/no-unsafe-assignment */
/* eslint-disable @typescript-eslint/no-unsafe-call */
/* eslint-disable @typescript-eslint/prefer-nullish-coalescing */

import type {JSX} from 'react';
import {
  Box,
  Alert,
  Typography,
  styled,
  AlertTitle,
  Paper,
  Stack,
  ColorSchemeImage,
  CircularProgress,
} from '@wso2/oxygen-ui';
import {AcceptInvite, type EmbeddedFlowComponent} from '@asgardeo/react';
import {useNavigate} from 'react-router';
import {useTranslation} from 'react-i18next';
import {useConfig} from '@thunder/shared-contexts';
import {useTemplateLiteralResolver} from '@thunder/shared-hooks';
import ROUTES from '../../constants/routes';
import FlowComponentRenderer from '../flow/FlowComponentRenderer';

const StyledPaper = styled(Paper)(({theme}) => ({
  display: 'flex',
  flexDirection: 'column',
  alignSelf: 'center',
  width: '100%',
  padding: theme.spacing(4),
  gap: theme.spacing(2),
  [theme.breakpoints.up('sm')]: {
    width: '450px',
  },
}));

export default function AcceptInviteBox(): JSX.Element {
  const navigate = useNavigate();
  const {resolve} = useTemplateLiteralResolver();
  const {t} = useTranslation();
  const {getServerUrl} = useConfig();

  const baseUrl = getServerUrl() ?? (import.meta.env.VITE_ASGARDEO_BASE_URL as string);

  const handleGoToSignIn = () => {
    const result = navigate(ROUTES.AUTH.SIGN_IN);
    if (result instanceof Promise) result.catch(() => {});
  };

  return (
    <Stack gap={2}>
      <ColorSchemeImage
        src={{
          light: `${import.meta.env.BASE_URL}/assets/images/logo.svg`,
          dark: `${import.meta.env.BASE_URL}/assets/images/logo-inverted.svg`,
        }}
        alt={{
          light: 'Logo (Light)',
          dark: 'Logo (Dark)',
        }}
        height={30}
        width="auto"
        sx={{display: {xs: 'flex', md: 'none'}}}
      />
      <StyledPaper variant="outlined">
        <AcceptInvite
          baseUrl={baseUrl}
          onGoToSignIn={handleGoToSignIn}
          onError={(error: Error) => {
            // eslint-disable-next-line no-console
            console.error('Invite acceptance error:', error);
          }}
        >
          {({
            values,
            fieldErrors,
            error,
            touched,
            isLoading,
            components,
            handleInputChange,
            handleSubmit,
            isComplete,
            isValidatingToken,
            isTokenInvalid,
            isValid = true,
          }: any) => {
            // Validating token
            if (isValidatingToken) {
              return (
                <Box sx={{display: 'flex', flexDirection: 'column', alignItems: 'center', p: 3, gap: 2}}>
                  <CircularProgress />
                  <Typography>{t('invite:validating', 'Validating your invite link...')}</Typography>
                </Box>
              );
            }

            // Invalid token
            if (isTokenInvalid) {
              return (
                <Alert severity="error">
                  <AlertTitle>{t('invite:errors.invalid.title', 'Unable to verify invite')}</AlertTitle>
                  {t('invite:errors.invalid.description', 'This invite link is invalid or has expired.')}
                </Alert>
              );
            }

            // Completed
            if (isComplete) {
              return (
                <Box sx={{textAlign: 'center', py: 2}}>
                  <Alert severity="success">
                    {t('invite:complete.description', 'Your account has been successfully set up.')}
                  </Alert>
                </Box>
              );
            }

            // Loading
            if (isLoading && !components?.length) {
              return (
                <Box sx={{display: 'flex', justifyContent: 'center', p: 3}}>
                  <CircularProgress />
                </Box>
              );
            }

            return (
              <>
                {error && (
                  <Alert severity="error" sx={{mb: 2}}>
                    <AlertTitle>{t('invite:errors.failed.title', 'Error')}</AlertTitle>
                    {error.message ?? t('invite:errors.failed.description', 'An error occurred.')}
                  </Alert>
                )}
                {components?.length > 0 && (
                  <Box sx={{display: 'flex', flexDirection: 'column', gap: 2}}>
                    {(components as EmbeddedFlowComponent[]).map((component, index) => (
                      <FlowComponentRenderer
                        key={component.id ?? index}
                        component={component}
                        index={index}
                        values={values ?? {}}
                        touched={touched}
                        fieldErrors={fieldErrors}
                        isLoading={isLoading || !isValid}
                        resolve={resolve}
                        onInputChange={handleInputChange}
                        onSubmit={(action, inputs) => {
                          handleSubmit(action, inputs).catch(() => {});
                        }}
                      />
                    ))}
                  </Box>
                )}
              </>
            );
          }}
        </AcceptInvite>
      </StyledPaper>
    </Stack>
  );
}
