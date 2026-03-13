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

/* eslint-disable @typescript-eslint/no-unsafe-assignment */
/* eslint-disable @typescript-eslint/no-explicit-any */
/* eslint-disable @typescript-eslint/no-unsafe-member-access */
/* eslint-disable @typescript-eslint/no-unsafe-call */

import type {JSX} from 'react';
import {
  Box,
  Button,
  Alert,
  Typography,
  styled,
  AlertTitle,
  Paper,
  Stack,
  ColorSchemeImage,
  CircularProgress,
} from '@wso2/oxygen-ui';
import {SignUp, type EmbeddedFlowComponent} from '@asgardeo/react';
import {useNavigate, useSearchParams} from 'react-router';
import {Trans, useTranslation} from 'react-i18next';
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

export default function SignUpBox(): JSX.Element {
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const {resolve} = useTemplateLiteralResolver();
  const {t} = useTranslation();

  const currentParams = searchParams.toString();
  const signInUrl = currentParams ? `${ROUTES.AUTH.SIGN_IN}?${currentParams}` : ROUTES.AUTH.SIGN_IN;

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
        sx={{
          display: {xs: 'flex', md: 'none'},
        }}
      />
      <StyledPaper variant="outlined">
        <SignUp afterSignUpUrl={signInUrl}>
          {({values, fieldErrors, error, touched, handleInputChange, handleSubmit, isLoading, components}: any) => (
            <>
              {!components ? (
                <Box sx={{display: 'flex', justifyContent: 'center', p: 3}}>
                  <CircularProgress />
                </Box>
              ) : (
                <>
                  {error && (
                    <Alert severity="error" sx={{mb: 2}}>
                      <AlertTitle>{t('signup:errors.signup.failed.message')}</AlertTitle>
                      {error.message ?? t('signup:errors.signup.failed.description')}
                    </Alert>
                  )}

                  {components && components.length > 0 ? (
                    <Box sx={{display: 'flex', flexDirection: 'column', gap: 2}}>
                      {isLoading && (
                        <Typography sx={{textAlign: 'center'}}>
                          {t('signup:creating', 'Creating account...')}
                        </Typography>
                      )}
                      {(components as EmbeddedFlowComponent[]).map((component, index) => (
                        <FlowComponentRenderer
                          key={component.id ?? index}
                          component={component}
                          index={index}
                          values={values ?? {}}
                          touched={touched}
                          fieldErrors={fieldErrors}
                          isLoading={isLoading}
                          resolve={resolve}
                          onInputChange={handleInputChange}
                          onSubmit={(action, inputs) => {
                            // Tracker: https://github.com/asgardeo/javascript/issues/222
                            handleSubmit(action, inputs).catch(() => {});
                          }}
                        />
                      ))}
                    </Box>
                  ) : (
                    <Alert severity="error" sx={{mb: 2}}>
                      <AlertTitle>{t("Oops, that didn't work")}</AlertTitle>
                      {t("We're sorry, we ran into a problem. Please try again!")}
                    </Alert>
                  )}
                </>
              )}

              <Typography sx={{textAlign: 'center', mt: 3}}>
                <Trans i18nKey="signup:redirect.to.signin">
                  Already have an account?
                  <Button
                    variant="text"
                    onClick={() => {
                      // eslint-disable-next-line @typescript-eslint/no-floating-promises
                      navigate(signInUrl);
                    }}
                    sx={{
                      p: 0,
                      minWidth: 'auto',
                      textTransform: 'none',
                      color: 'primary.main',
                      textDecoration: 'underline',
                      '&:hover': {
                        textDecoration: 'underline',
                        backgroundColor: 'transparent',
                      },
                    }}
                  >
                    Sign in
                  </Button>
                </Trans>
              </Typography>
            </>
          )}
        </SignUp>
      </StyledPaper>
    </Stack>
  );
}
