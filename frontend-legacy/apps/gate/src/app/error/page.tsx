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
import AlertTitle from '@oxygen-ui/react/src/components/AlertTitle/AlertTitle';
import Typography from '@oxygen-ui/react/src/components/Typography/Typography';
import React, { useState, useEffect, ReactElement } from 'react';

const FallbackErrorMessage: string = 'Sorry, but we encountered an error while processing your request.';

export default function ErrorPage(): ReactElement {
  const [errorCode, setErrorCode] = useState('');
  const [errorMsg, setErrorMsg] = useState(FallbackErrorMessage);

  useEffect(() => {
    const params: URLSearchParams = new URLSearchParams(window.location.search);

    setErrorCode(params.get('errorCode') || '');
    setErrorMsg(params.get('errorMessage') || FallbackErrorMessage);
  }, []);

  return (
    <Alert severity="error">
      <AlertTitle>
        <Typography variant="h6">Something didn&apos;t go as expected!</Typography>
      </AlertTitle>
      <Typography variant="body1" sx={{ mt: 3 }}>
        {errorMsg}
      </Typography>
      {errorCode !== '' && (
        <Typography variant="body1" sx={{ mt: 2 }}>
          Error Code: {errorCode}
        </Typography>
      )}
    </Alert>
  );
}
