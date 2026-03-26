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
import {useState} from 'react';
import {Box, Button, Alert, Typography, AlertTitle, CircularProgress} from '@wso2/oxygen-ui';
import {SignUp, useAsgardeo, type EmbeddedFlowComponent} from '@asgardeo/react';
import {useNavigate, useSearchParams} from 'react-router';
import {Trans, useTranslation} from 'react-i18next';
import {FlowComponentRenderer, AuthCardLayout} from '@thunder/shared-design';
import ROUTES from '../../constants/routes';

export default function SignUpBox(): JSX.Element {
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const {resolveFlowTemplateLiterals: resolve} = useAsgardeo();
  const {t} = useTranslation();
  const [emailSent, setEmailSent] = useState<boolean | null>(null);
  const [flowError, setFlowError] = useState<string | null>(null);

  const currentParams = searchParams.toString();
  const signInUrl = currentParams ? `${ROUTES.AUTH.SIGN_IN}?${currentParams}` : ROUTES.AUTH.SIGN_IN;

  const renderFlowContent = (
    components: EmbeddedFlowComponent[],
    error: any,
    isLoading: boolean,
    values: any,
    touched: any,
    fieldErrors: any,
    handleInputChange: any,
    handleSubmit: any,
  ): JSX.Element | null => {
    if (components.length > 0) {
      return (
        <Box sx={{display: 'flex', flexDirection: 'column', gap: 2}}>
          {isLoading && (
            <Typography sx={{textAlign: 'center'}}>
              {t('signup:create_account.loading', 'Creating account...')}
            </Typography>
          )}
          {components.map((component, index) => (
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
                setFlowError(null);
                void handleSubmit(action, inputs);
              }}
            />
          ))}
        </Box>
      );
    }
    if (emailSent === true) {
      return (
        <Alert severity="info" sx={{mb: 2}}>
          <AlertTitle>{t('signup:email.sent.title', 'Check your email')}</AlertTitle>
          {t(
            'signup:email.sent.description',
            'We sent you an invitation link. Please check your inbox and click the link to continue.',
          )}
        </Alert>
      );
    }
    if (emailSent === false) {
      return (
        <Alert severity="warning" sx={{mb: 2}}>
          <AlertTitle>{t('signup:email.not.sent.title', 'Email service unavailable')}</AlertTitle>
          {t(
            'signup:email.not.sent.description',
            'Your registration was received but the verification email could not be sent. Please contact your administrator.',
          )}
        </Alert>
      );
    }
    if (!error) {
      return (
        <Alert severity="error" sx={{mb: 2}}>
          <AlertTitle>{t("Oops, that didn't work")}</AlertTitle>
          {t("We're sorry, we ran into a problem. Please try again!")}
        </Alert>
      );
    }
    return null;
  };

  return (
    <AuthCardLayout
      variant="SignUpBox"
      logo={{
        src: {
          light: `${import.meta.env.BASE_URL}/assets/images/logo.svg`,
          dark: `${import.meta.env.BASE_URL}/assets/images/logo-inverted.svg`,
        },
        alt: {light: '', dark: ''},
      }}
    >
      <SignUp
        afterSignUpUrl={signInUrl}
        onFlowChange={(response: any) => {
          const emailSentValue = response?.data?.additionalData?.emailSent;
          if (emailSentValue === 'true') {
            setEmailSent(true);
          } else if (emailSentValue === 'false') {
            setEmailSent(false);
          }
          if (response?.failureReason) {
            setFlowError(response.failureReason as string);
          } else {
            setFlowError(null);
          }
        }}
      >
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
                {flowError && (
                  <Alert severity="error" sx={{mb: 2}}>
                    {flowError}
                  </Alert>
                )}

                {renderFlowContent(
                  components as EmbeddedFlowComponent[],
                  error,
                  isLoading as boolean,
                  values,
                  touched,
                  fieldErrors,
                  handleInputChange,
                  handleSubmit,
                )}
              </>
            )}

            {emailSent !== true && (
              <Typography sx={{textAlign: 'center', mt: 3}}>
                <Trans i18nKey="signup:redirect.to.signin">
                  Already have an account?
                  <Button
                    variant="text"
                    onClick={() => {
                      void navigate(signInUrl);
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
            )}
          </>
        )}
      </SignUp>
    </AuthCardLayout>
  );
}
