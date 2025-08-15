/**
 * Copyright (c) 2025, WSO2 LLC. (https://www.wso2.com).
 *
 * WSO2 LLC. licenses this file to you under the Apache License,
 * Version 2.0 (the "License"); you may not use this file except
 * in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */

'use client';

import Alert from '@oxygen-ui/react/src/components/Alert/Alert';
import Box from '@oxygen-ui/react/src/components/Box/Box';
import Button from '@oxygen-ui/react/src/components/Button/Button';
import OutlinedInput from '@oxygen-ui/react/src/components/OutlinedInput/OutlinedInput';
import InputLabel from '@oxygen-ui/react/src/components/InputLabel/InputLabel';
import Typography from '@oxygen-ui/react/src/components/Typography/Typography';
import React, { useState, useEffect, ReactElement } from 'react';
import axios from 'axios';
import AppConfig from '@/configs/app.json';

interface LoginInput {
  name: string;
  type: string;
  required?: boolean;
}

const LoginPageContent = function (): ReactElement {
  const [sessionDataKey, setSessionDataKey] = useState<string>('');
  const [_appId, setAppId] = useState<string>('');
  const [insecureWarning, setInsecureWarning] = useState<boolean>(false);
  const [flowId, setFlowId] = useState<string>('');
  const [inputs, setInputs] = useState<LoginInput[]>([]);
  const [formValues, setFormValues] = useState<Record<string, string>>({});
  const [error, setError] = useState<string>('');
  const [loading, setLoading] = useState<boolean>(true);
  const [showPassword, setShowPassword] = useState(false);

  useEffect(() => {
    const params: URLSearchParams = new URLSearchParams(window.location.search);
    const key = params.get('sessionDataKey') || '';
    const appId = params.get('applicationId') || '';

    setSessionDataKey(key);
    setAppId(appId);
    setInsecureWarning(params.get('showInsecureWarning') === 'true');

    if (key) {
      axios.post(
          AppConfig.flowExecutionEndpoint,
          { applicationId: appId, flowType: 'AUTHENTICATION' },
          {
            headers: {
              Accept: 'application/json',
              'Content-Type': 'application/json',
            },
            withCredentials: true,
          },
        )
        .then(res => {
          setFlowId(res.data.flowId);
          setInputs(res.data.data?.inputs || []);
          setLoading(false);
        })
        .catch(() => {
          setError('Failed to initiate authentication flow.');
          setLoading(false);
        });
    } else {
      setLoading(false);
    }
  }, []);

  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement>): void => {
    setFormValues({ ...formValues, [e.target.name]: e.target.value });
  };

  const handleTogglePasswordVisibility = (): void => {
    setShowPassword(prev => !prev);
  };

  const handleSubmit = async (e: React.FormEvent): Promise<any> => {
    e.preventDefault();
    setError('');
    const currentFlowId = flowId;
    const currentInputs = formValues;
    let continueFlow = true;
    while (continueFlow) {
      try {
        const res = await axios.post(
          AppConfig.flowExecutionEndpoint,
          {
            flowId: currentFlowId,
            inputs: currentInputs,
          },
          {
            headers: { 'Content-Type': 'application/json' },
            withCredentials: true,
            validateStatus: () => true,
          },
        );
        if (res.status !== 200) {
          setError(res.data?.message || `Error response status: ${res.status}`);
          continueFlow = false;
          break;
        }
        if (res.data.flowStatus === 'ERROR') {
          setError(res.data.failureReason || 'Authentication failed.');
          continueFlow = false;
        } else if (res.data.flowStatus === 'COMPLETE') {
          const assertion = res?.data?.assertion;

          const authZRes = await axios.post(
            AppConfig.authorizationEndpoint,
            {
              sessionDataKey,
              assertion,
            },
            {
              headers: { 'Content-Type': 'application/json' },
              withCredentials: true,
              validateStatus: () => true,
            },
          );

          if (authZRes.status !== 200) {
            setError(authZRes.data?.message || `Authorization failed with status: ${authZRes.status}`);
            continueFlow = false;
            break;
          }

          const redirectUrl = authZRes?.data?.redirect_uri;
          if (redirectUrl) {
            window.location.href = redirectUrl;
          } else {
            setError('Authorization completed but no redirect URL provided.');
          }
          continueFlow = false;
        } else if (res.data.flowStatus === 'INCOMPLETE') {
          // Update inputs and flowId for next step
          setInputs(res.data.data?.inputs || []);
          setFlowId(res.data.flowId);
          setFormValues({}); // Clear previous values for new input
          continueFlow = false; // Stop loop, wait for user to submit again
          setError(''); // Clear any previous error
        }
      } catch {
        setError('Failed to authenticate.');
        continueFlow = false;
      }
    }
  };

  return (
    <form onSubmit={handleSubmit}>
      <Box sx={{ mb: 4 }}>
        <Typography variant="h5" gutterBottom>
          Login to Account
        </Typography>
      </Box>
      {insecureWarning && (
        <Alert severity="warning" sx={{ my: 2 }}>
          You are about to access a non-secure site. Proceed with caution!
        </Alert>
      )}
      {error && (
        <Alert severity="error" sx={{ my: 2 }}>
          {error}
        </Alert>
      )}
      {loading ? (
        <Typography>Loading...</Typography>
      ) : (
        <Box display="flex" flexDirection="column" gap={2}>
          {inputs.map(input => {
            const words = input.name.replace(/([A-Z])/g, ' $1').split(' ');
            let formattedLabel = words.map(w => w.charAt(0).toUpperCase() + w.slice(1)).join(' ');
            let formattedPlaceholder = words.join(' ').toLowerCase();

            if (input.name === 'otp') {
              formattedLabel = 'OTP';
              formattedPlaceholder = 'one time password';
            }

            return (
              <Box key={input.name} display="flex" flexDirection="column" gap={0.5}>
                <InputLabel htmlFor={input.name}>{formattedLabel}</InputLabel>
                <OutlinedInput
                  type={input.name === 'password' ? (showPassword ? 'text' : 'password') : 'text'}
                  id={input.name}
                  name={input.name}
                  placeholder={`Enter your ${formattedPlaceholder}`}
                  size="small"
                  required={input.required}
                  value={formValues[input.name] || ''}
                  onChange={handleInputChange}
                  endAdornment={
                    input.name === 'password' && (
                      <Button
                        variant="text"
                        onClick={handleTogglePasswordVisibility}
                        style={{
                          background: 'none',
                          border: 'none',
                          cursor: 'pointer',
                          fontSize: '0.875rem',
                          color: '#1976d2',
                        }}
                      >
                        {showPassword ? 'Hide' : 'Show'}
                      </Button>
                    )
                  }
                />
              </Box>
            );
          })}
          <Button variant="contained" color="primary" type="submit" fullWidth sx={{ mt: 2 }}>
            Sign In
          </Button>
        </Box>
      )}
    </form>
  );
};

export default LoginPageContent;
