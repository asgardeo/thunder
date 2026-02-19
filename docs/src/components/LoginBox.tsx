/**
 * Copyright (c) 2026, WSO2 LLC. (https://www.wso2.com).
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

import React, {useEffect, useState} from 'react';
import {Box, Typography, TextField, Button, Divider, Card, FormControl, FormLabel} from '@wso2/oxygen-ui';
import useIsDarkMode from '../hooks/useIsDarkMode';

interface LoginBoxProps {
  variant: 'email' | 'social' | 'mfa';
  delay?: number;
  sideCard?: boolean;
  sx?: object;
}

export default function LoginBox({variant, delay = 0, sideCard = false, sx = {}}: LoginBoxProps) {
  const isDark = useIsDarkMode();

  // Clear entry animation after it plays so transitions work smoothly on hover.
  const [entryDone, setEntryDone] = useState(false);

  useEffect(() => {
    const timer = setTimeout(() => setEntryDone(true), (delay + 0.6) * 1000 + 100);

    return () => clearTimeout(timer);
  }, [delay]);

  const renderEmailLogin = () => (
    <>
      <Box sx={{textAlign: 'center', mb: 3}}>
        <Box
          sx={{
            width: 48,
            height: 48,
            margin: '0 auto 12px',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            transition: 'transform 0.8s cubic-bezier(0.4, 0, 0.2, 1), filter 0.8s cubic-bezier(0.4, 0, 0.2, 1)',
            '.MuiCard-root:hover &': {
              transform: 'scale(1.05)',
              filter: isDark
                ? 'drop-shadow(0 0 6px rgba(255, 140, 0, 0.15))'
                : 'drop-shadow(0 0 6px rgba(255, 107, 0, 0.1))',
            },
          }}
        >
          <svg width="40" height="40" viewBox="0 0 64 64" fill="currentColor" style={{opacity: 0.8}}>
            <circle opacity="0.3" cx="32" cy="32" r="11" />
            <circle opacity="0.6" cx="32" cy="32" r="5" />
            <path
              opacity="0.3"
              d="M39,14.65c0,3.51-3.13,6.35-7,6.35s-7-2.84-7-6.35,7-11.65,7-11.65c0,0,7,8.14,7,11.65Z"
            />
            <path opacity="0.5" d="M36,17c0,2.21-1.79,4-4,4s-4-1.79-4-4,4-6,4-6c0,0,4,3.79,4,6Z" />
            <path
              opacity="0.3"
              d="M20.47,17.26c3.04,1.75,3.93,5.89,2,9.24s-5.96,4.64-9,2.89-6.59-11.89-6.59-11.89c0,0,10.55-1.99,13.59-.24Z"
            />
            <path
              opacity="0.5"
              d="M21.01,21.04c1.91,1.1,2.57,3.55,1.46,5.46s-3.55,2.57-5.46,1.46-3.2-6.46-3.2-6.46c0,0,5.28-1.57,7.2-.46Z"
            />
            <path
              opacity="0.3"
              d="M13.47,34.61c3.04-1.75,7.07-.46,9,2.89s1.04,7.48-2,9.24c-3.04,1.75-13.59-.24-13.59-.24,0,0,3.55-10.13,6.59-11.89Z"
            />
            <path
              opacity="0.5"
              d="M17.01,36.04c1.91-1.1,4.36-.45,5.46,1.46s.45,4.36-1.46,5.46-7.2-.46-7.2-.46c0,0,1.28-5.36,3.2-6.46Z"
            />
            <path
              opacity="0.3"
              d="M25,49.35c0-3.51,3.13-6.35,7-6.35,3.87,0,7,2.84,7,6.35,0,3.51-7,11.65-7,11.65,0,0-7-8.14-7-11.65Z"
            />
            <path opacity="0.5" d="M28,47c0-2.21,1.79-4,4-4s4,1.79,4,4-4,6-4,6c0,0-4-3.79-4-6Z" />
            <path
              opacity="0.3"
              d="M43.53,46.74c-3.04-1.75-3.93-5.89-2-9.24s5.96-4.64,9-2.89c3.04,1.75,6.59,11.89,6.59,11.89,0,0-10.55,1.99-13.59,.24Z"
            />
            <path
              opacity="0.5"
              d="M42.99,42.96c-1.91-1.1-2.57-3.55-1.46-5.46s3.55-2.57,5.46-1.46,3.2,6.46,3.2,6.46c0,0-5.28,1.57-7.2,.46Z"
            />
            <path
              opacity="0.3"
              d="M50.53,29.39c-3.04,1.75-7.07,.46-9-2.89s-1.04-7.48,2-9.24c3.04-1.75,13.59,.24,13.59,.24,0,0-3.55,10.13-6.59,11.89Z"
            />
            <path
              opacity="0.5"
              d="M46.99,27.96c-1.91,1.1-4.36,.45-5.46-1.46s-.45-4.36,1.46-5.46,7.2,.46,7.2,.46c0,0-1.28,5.36-3.2,6.46Z"
            />
          </svg>
        </Box>
        <Typography
          variant="h6"
          sx={{
            fontWeight: 500,
            mb: 0.5,
            fontSize: '1.1rem',
          }}
        >
          Sign in to ACME
        </Typography>
        <Typography
          variant="body2"
          sx={{
            fontSize: '0.8rem',
            px: 2,
            opacity: 0.7,
          }}
        >
          Enter your basic credentials or go through social login to continue
        </Typography>
      </Box>
      <FormControl fullWidth sx={{mb: 2.5}}>
        <FormLabel htmlFor="email-input">Email</FormLabel>
        <TextField
          fullWidth
          id="email-input"
          placeholder="Your email address"
          size="small"
          slotProps={{
            input: {
              readOnly: true,
            },
          }}
          sx={{
            '& .MuiOutlinedInput-root': {
              borderRadius: 2,
            },
          }}
        />
      </FormControl>
      <Button
        fullWidth
        variant="outlined"
        sx={{
          textTransform: 'none',
          mb: 2.5,
          py: 1.3,
          borderRadius: 10,
          borderColor: !isDark ? 'rgba(0, 0, 0, 0.15)' : 'rgba(255, 255, 255, 0.15)',
          color: !isDark ? 'rgba(0, 0, 0, 0.7)' : 'rgba(255, 255, 255, 0.7)',
          bgcolor: !isDark ? 'rgba(0, 0, 0, 0.03)' : 'rgba(255, 255, 255, 0.03)',
          '&:hover': {
            borderColor: !isDark ? 'rgba(0, 0, 0, 0.25)' : 'rgba(255, 255, 255, 0.25)',
            bgcolor: !isDark ? 'rgba(0, 0, 0, 0.05)' : 'rgba(255, 255, 255, 0.05)',
          },
        }}
      >
        Continue
      </Button>
      <Divider sx={{my: 2.5, borderColor: !isDark ? 'rgba(0, 0, 0, 0.1)' : 'rgba(255, 255, 255, 0.1)'}}>
        <Typography variant="body2" sx={{fontSize: '0.75rem', opacity: 0.6}}>
          OR
        </Typography>
      </Divider>
      <Button
        fullWidth
        variant="outlined"
        startIcon={
          <Box
            component="span"
            sx={{
              width: 18,
              height: 18,
              display: 'flex',
              alignItems: 'center',
              color: !isDark ? 'rgba(0, 0, 0, 0.7)' : 'rgba(255, 255, 255, 0.7)',
            }}
          >
            <svg width="18" height="18" viewBox="0 0 18 18" fill="currentColor" style={{opacity: 0.8}}>
              <path
                opacity="0.5"
                d="M17.64 9.2c0-.637-.057-1.251-.164-1.84H9v3.481h4.844c-.209 1.125-.843 2.078-1.796 2.717v2.258h2.908c1.702-1.567 2.684-3.875 2.684-6.615z"
              />
              <path
                opacity="0.4"
                d="M9 18c2.43 0 4.467-.806 5.956-2.18l-2.908-2.259c-.806.54-1.837.86-3.048.86-2.344 0-4.328-1.584-5.036-3.711H.957v2.332C2.438 15.983 5.482 18 9 18z"
              />
              <path
                opacity="0.3"
                d="M3.964 10.71c-.18-.54-.282-1.117-.282-1.71s.102-1.17.282-1.71V4.958H.957C.347 6.173 0 7.548 0 9s.348 2.827.957 4.042l3.007-2.332z"
              />
              <path
                opacity="0.6"
                d="M9 3.58c1.321 0 2.508.454 3.44 1.345l2.582-2.58C13.463.891 11.426 0 9 0 5.482 0 2.438 2.017.957 4.958L3.964 7.29C4.672 5.163 6.656 3.58 9 3.58z"
              />
            </svg>
          </Box>
        }
        sx={{
          textTransform: 'none',
          mb: 1.5,
          py: 1.3,
          borderRadius: 10,
          borderColor: !isDark ? 'rgba(0, 0, 0, 0.15)' : 'rgba(255, 255, 255, 0.15)',
          color: !isDark ? 'rgba(0, 0, 0, 0.7)' : 'rgba(255, 255, 255, 0.7)',
          bgcolor: !isDark ? 'rgba(0, 0, 0, 0.03)' : 'rgba(255, 255, 255, 0.03)',
          '&:hover': {
            borderColor: !isDark ? 'rgba(0, 0, 0, 0.25)' : 'rgba(255, 255, 255, 0.25)',
            bgcolor: !isDark ? 'rgba(0, 0, 0, 0.05)' : 'rgba(255, 255, 255, 0.05)',
          },
        }}
      >
        Continue with Google
      </Button>
      <Button
        fullWidth
        variant="outlined"
        startIcon={
          <Box
            component="span"
            sx={{
              width: 18,
              height: 18,
              display: 'flex',
              alignItems: 'center',
              color: !isDark ? 'rgba(0, 0, 0, 0.7)' : 'rgba(255, 255, 255, 0.7)',
            }}
          >
            <svg width="18" height="18" viewBox="0 0 24 24" fill="currentColor" style={{opacity: 0.8}}>
              <rect x="1" y="1" width="10" height="10" />
              <rect x="13" y="1" width="10" height="10" />
              <rect x="1" y="13" width="10" height="10" />
              <rect x="13" y="13" width="10" height="10" />
            </svg>
          </Box>
        }
        sx={{
          textTransform: 'none',
          py: 1.3,
          borderRadius: 10,
          borderColor: !isDark ? 'rgba(0, 0, 0, 0.15)' : 'rgba(255, 255, 255, 0.15)',
          color: !isDark ? 'rgba(0, 0, 0, 0.7)' : 'rgba(255, 255, 255, 0.7)',
          bgcolor: !isDark ? 'rgba(0, 0, 0, 0.03)' : 'rgba(255, 255, 255, 0.03)',
          '&:hover': {
            borderColor: !isDark ? 'rgba(0, 0, 0, 0.25)' : 'rgba(255, 255, 255, 0.25)',
            bgcolor: !isDark ? 'rgba(0, 0, 0, 0.05)' : 'rgba(255, 255, 255, 0.05)',
          },
        }}
      >
        Continue with Microsoft
      </Button>
      <Typography
        variant="caption"
        sx={{
          display: 'block',
          textAlign: 'center',
          mt: 2.5,
          fontSize: '0.8rem',
          opacity: 0.7,
        }}
      >
        Don&apos;t have an account?{' '}
        <Box
          component="span"
          sx={{
            cursor: 'pointer',
            fontWeight: 600,
            opacity: 1,
            '&:hover': {textDecoration: 'underline'},
          }}
        >
          Sign up
        </Box>
      </Typography>
    </>
  );

  const renderSocialLogin = () => (
    <>
      <Box sx={{textAlign: 'center', mb: 3}}>
        <Box
          sx={{
            width: 48,
            height: 48,
            margin: '0 auto 12px',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            transition: 'transform 0.8s cubic-bezier(0.4, 0, 0.2, 1), filter 0.8s cubic-bezier(0.4, 0, 0.2, 1)',
            '.MuiCard-root:hover &': {
              transform: 'scale(1.05)',
              filter: isDark
                ? 'drop-shadow(0 0 6px rgba(255, 140, 0, 0.15))'
                : 'drop-shadow(0 0 6px rgba(255, 107, 0, 0.1))',
            },
          }}
        >
          <svg width="40" height="40" viewBox="0 0 64 64" fill="currentColor" style={{opacity: 0.8}}>
            <path opacity="0.3" d="M50,31c0,9.94-8.06,18-18,18S14,40.94,14,31,32,3,32,3c0,0,18,18.06,18,28Z" />
            <path
              opacity="0.5"
              d="M31.54,48.99c-9.73-.24-17.54-8.2-17.54-17.99,0-2.57,1.21-5.69,3-8.9h0c.02,3.36,.73,6.77,3.16,8.53,2.21,1.6,3.43,3.51,4.09,5.25-.76,1.52-1.25,2.96-1.25,4.12,0,4.81,3.78,8.75,8.54,8.99Z"
            />
            <path
              opacity="0.5"
              d="M32.46,48.99c9.73-.24,17.54-8.2,17.54-17.99,0-2.57-1.21-5.69-3-8.9h0c-.02,3.36-.73,6.77-3.16,8.53-2.21,1.6-3.43,3.51-4.09,5.25,.76,1.52,1.25,2.96,1.25,4.12,0,4.81-3.78,8.75-8.54,8.99Z"
            />
            <path
              opacity="0.6"
              d="M30.53,26.72c.88-1.06,1.47-1.72,1.47-1.72,0,0,.59,.66,1.47,1.72,.57-1.95,1.79-4.23,4.37-6.09,2.15-1.56,2.96-4.42,3.13-7.4-4.49-5.73-8.97-10.23-8.97-10.23,0,0-4.48,4.5-8.97,10.23,.17,2.98,.98,5.84,3.13,7.4,2.58,1.86,3.8,4.14,4.37,6.09Z"
            />
            <path
              opacity="0.4"
              d="M35.75,48.6c-1.21,.26-2.46,.4-3.75,.4s-2.54-.14-3.75-.4c.41,2.33,.75,6.15,.75,12.4h6c0-6.25,.34-10.07,.75-12.4Z"
            />
            <path opacity="0.5" d="M41,40c0,4.97-4.03,9-9,9s-9-4.03-9-9,9-15,9-15c0,0,9,10.03,9,15Z" />
            <path
              opacity="0.3"
              d="M35.75,48.6c4.08-.86,7.65-3.1,10.19-6.22,2.81,1.36,6.3,.88,8.63-1.45,2.86-2.86,2.96-7.43,.27-10.38-.08-.1-.17-.19-.27-.28-.09-.1-.18-.19-.28-.27l.55,.55c5.64,5.9,5.54,15.24-.27,21.03-5.26,5.26-13.45,5.82-19.35,1.69,.14-2,.32-3.52,.53-4.67Z"
            />
            <path
              opacity="0.3"
              d="M28.25,48.6c-4.08-.86-7.65-3.1-10.19-6.22-2.81,1.36-6.3,.88-8.63-1.45-2.86-2.86-2.96-7.43-.27-10.38l.55-.55c-.10,.08-.19,.17-.28,.27-.10,.09-.19,.18-.27,.28-5.64,5.9-5.54,15.24,.27,21.03,5.26,5.26,13.45,5.82,19.35,1.69-.14-2-.32-3.52-.53-4.67Z"
            />
          </svg>
        </Box>
        <Typography
          variant="h6"
          sx={{
            fontWeight: 500,
            mb: 0.5,
            fontSize: '1.1rem',
          }}
        >
          Welcome to Teamspace
        </Typography>
        <Typography
          variant="body2"
          sx={{
            fontSize: '0.8rem',
            px: 2,
            opacity: 0.7,
          }}
        >
          Enter your username to continue sign-in to the application
        </Typography>
      </Box>
      <FormControl fullWidth sx={{mb: 2.5}}>
        <FormLabel htmlFor="username-input">Username</FormLabel>
        <TextField
          fullWidth
          id="username-input"
          placeholder="Your username/email"
          size="small"
          slotProps={{
            input: {
              readOnly: true,
            },
          }}
          sx={{
            '& .MuiOutlinedInput-root': {
              borderRadius: 2,
            },
          }}
        />
      </FormControl>
      <Button
        fullWidth
        variant="outlined"
        sx={{
          textTransform: 'none',
          py: 1.3,
          borderRadius: 10,
          borderColor: !isDark ? 'rgba(0, 0, 0, 0.15)' : 'rgba(255, 255, 255, 0.15)',
          color: !isDark ? 'rgba(0, 0, 0, 0.7)' : 'rgba(255, 255, 255, 0.7)',
          bgcolor: !isDark ? 'rgba(0, 0, 0, 0.03)' : 'rgba(255, 255, 255, 0.03)',
          '&:hover': {
            borderColor: !isDark ? 'rgba(0, 0, 0, 0.25)' : 'rgba(255, 255, 255, 0.25)',
            bgcolor: !isDark ? 'rgba(0, 0, 0, 0.05)' : 'rgba(255, 255, 255, 0.05)',
          },
        }}
      >
        Continue
      </Button>
      <Typography
        variant="caption"
        sx={{
          display: 'block',
          textAlign: 'center',
          mt: 2.5,
          fontSize: '0.8rem',
          opacity: 0.7,
        }}
      >
        Don&apos;t have an account?{' '}
        <Box
          component="span"
          sx={{
            cursor: 'pointer',
            fontWeight: 600,
            opacity: 1,
            '&:hover': {textDecoration: 'underline'},
          }}
        >
          Sign up
        </Box>
      </Typography>
    </>
  );

  const renderMfaLogin = () => (
    <>
      <Box sx={{textAlign: 'center', mb: 3}}>
        <Box
          sx={{
            width: 48,
            height: 48,
            margin: '0 auto 12px',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            transition: 'transform 0.8s cubic-bezier(0.4, 0, 0.2, 1), filter 0.8s cubic-bezier(0.4, 0, 0.2, 1)',
            '.MuiCard-root:hover &': {
              transform: 'scale(1.05)',
              filter: isDark
                ? 'drop-shadow(0 0 6px rgba(255, 140, 0, 0.15))'
                : 'drop-shadow(0 0 6px rgba(255, 107, 0, 0.1))',
            },
          }}
        >
          <svg width="40" height="40" viewBox="0 0 64 64" fill="currentColor" style={{opacity: 0.8}}>
            <path
              opacity="0.3"
              d="M44,44c0,2.64-1.9,6.03-4.19,9.15-.38-.81-1.28-1.32-2.28-1.1-.71,.17-1.3,.75-1.47,1.46-.05,.21-.07,.42-.06,.62,.14,2.07-.36,4.13-1.78,5.63-1.31,1.38-2.22,2.24-2.22,2.24,0,0-1.02-.96-2.46-2.49-2.07-2.21-2.82-5.23-2.55-8.25,.02-.25,.01-.52-.04-.79-.21-1.24-1.24-2.24-2.48-2.43-1.16-.18-2.2,.31-2.83,1.13-.99-1.84-1.64-3.63-1.64-5.17-2.64,0-6.03-1.9-9.15-4.19,.81-.38,1.32-1.28,1.1-2.28-.17-.71-.75-1.3-1.46-1.47-.21-.05-.42-.07-.62-.06-2.07,.14-4.13-.36-5.63-1.78-1.38-1.31-2.24-2.22-2.24-2.22,0,0,.96-1.02,2.49-2.46,2.21-2.07,5.23-2.82,8.25-2.55,.25,.02,.52,.01,.79-.04,1.24-.21,2.24-1.24,2.43-2.48,.18-1.16-.31-2.2-1.13-2.83,1.84-.99,3.63-1.64,5.17-1.64,0-2.64,1.9-6.03,4.19-9.15,.38,.81,1.28,1.32,2.28,1.1,.71-.17,1.3-.75,1.47-1.46,.05-.21,.07-.42,.06-.62-.14-2.07,.36-4.13,1.78-5.63,1.31-1.38,2.22-2.24,2.22-2.24,0,0,1.02,.96,2.46,2.49,2.07,2.21,2.82,5.23,2.55,8.25-.02,.25-.01,.52,.04,.79,.21,1.24,1.24,2.24,2.48,2.43,1.16,.18,2.2-.31,2.83-1.13,.99,1.84,1.64,3.63,1.64,5.17,2.64,0,6.03,1.9,9.15,4.19-.81,.38-1.32,1.28-1.1,2.28,.17,.71,.75,1.3,1.46,1.47,.21,.05,.42,.07,.62,.06,2.07-.14,4.13,.36,5.63,1.78,1.38,1.31,2.24,2.22,2.24,2.22,0,0-.96,1.02-2.49,2.46-2.21,2.07-5.23,2.82-8.25,2.55-.25-.02-.52-.01-.79,.04-1.24,.21-2.24,1.24-2.43,2.48-.18,1.16,.31,2.2,1.13,2.83-1.84,.99-3.63,1.64-5.17,1.64Z"
            />
            <path
              opacity="0.5"
              d="M38.94,38.94c3.75-.65,11.06-6.94,11.06-6.94,0,0-7.32-6.29-11.06-6.94-.65-3.75-6.94-11.06-6.94-11.06,0,0-6.29,7.32-6.94,11.06-3.75,.65-11.06,6.94-11.06,6.94,0,0,7.32,6.29,11.06,6.94,.65,3.75,6.94,11.06,6.94,11.06,0,0,6.29-7.32,6.94-11.06Z"
            />
            <path
              opacity="0.6"
              d="M36,28c0-2.21-4-6-4-6,0,0-4,3.79-4,6-2.21,0-6,4-6,4,0,0,3.79,4,6,4,0,2.21,4,6,4,6,0,0,4-3.79,4-6,2.21,0,6-4,6-4,0,0-3.79-4-6-4Z"
            />
            <circle opacity="0.7" cx="32" cy="32" r="3" />
          </svg>
        </Box>
        <Typography
          variant="h6"
          sx={{
            fontWeight: 500,
            mb: 0.5,
            fontSize: '1.1rem',
          }}
        >
          Verify OTP
        </Typography>
        <Typography
          variant="body2"
          sx={{
            fontSize: '0.8rem',
            px: 2,
            opacity: 0.7,
          }}
        >
          Enter the temporary passcode sent to your email to continue
        </Typography>
      </Box>
      <Box sx={{display: 'flex', gap: 1, mb: 3, justifyContent: 'center'}}>
        {[1, 2, 3, 4, 5, 6].map((i) => (
          <TextField
            key={i}
            variant="outlined"
            slotProps={{
              input: {
                readOnly: true,
                style: {
                  textAlign: 'center',
                  padding: 0,
                  fontSize: '1.25rem',
                  fontWeight: 500,
                },
              },
            }}
            sx={{
              width: 40,
              '& .MuiOutlinedInput-root': {
                height: 52,
                borderRadius: 2,
              },
            }}
          />
        ))}
      </Box>
      <Button
        fullWidth
        variant="outlined"
        sx={{
          textTransform: 'none',
          py: 1.3,
          borderRadius: 10,
          borderColor: !isDark ? 'rgba(0, 0, 0, 0.15)' : 'rgba(255, 255, 255, 0.15)',
          color: !isDark ? 'rgba(0, 0, 0, 0.7)' : 'rgba(255, 255, 255, 0.7)',
          bgcolor: !isDark ? 'rgba(0, 0, 0, 0.03)' : 'rgba(255, 255, 255, 0.03)',
          '&:hover': {
            borderColor: !isDark ? 'rgba(0, 0, 0, 0.25)' : 'rgba(255, 255, 255, 0.25)',
            bgcolor: !isDark ? 'rgba(0, 0, 0, 0.05)' : 'rgba(255, 255, 255, 0.05)',
          },
        }}
      >
        Continue
      </Button>
      <Typography
        variant="caption"
        sx={{
          display: 'block',
          textAlign: 'center',
          mt: 2.5,
          fontSize: '0.8rem',
          opacity: 0.7,
        }}
      >
        Don&apos;t have an account?{' '}
        <Box
          component="span"
          sx={{
            cursor: 'pointer',
            fontWeight: 600,
            opacity: 1,
            '&:hover': {textDecoration: 'underline'},
          }}
        >
          Sign up
        </Box>
      </Typography>
    </>
  );

  return (
    <Card
      sx={{
        width: sideCard ? 300 : 340,
        p: 3.5,
        textAlign: 'left',
        '@keyframes fadeInUp': {
          '0%': {
            opacity: 0,
            transform: 'translateY(30px)',
          },
          '100%': {
            opacity: 1,
            transform: 'translateY(0)',
          },
        },
        ...(entryDone
          ? {}
          : {
              animation: 'fadeInUp 0.6s ease-out',
              animationDelay: `${delay}s`,
              animationFillMode: 'both',
            }),
        transition: 'transform 0.8s cubic-bezier(0.4, 0, 0.2, 1), opacity 0.8s cubic-bezier(0.4, 0, 0.2, 1), box-shadow 0.8s cubic-bezier(0.4, 0, 0.2, 1), border-color 0.8s cubic-bezier(0.4, 0, 0.2, 1)',
        bgcolor: isDark ? '#14141e' : '#ffffff',
        // Glowing border
        border: '1px solid',
        borderColor: isDark ? 'rgba(255, 140, 0, 0.15)' : 'rgba(255, 107, 0, 0.12)',
        // Shimmer sweep overlay — very subtle
        overflow: 'hidden',
        position: 'relative',
        '&::before': {
          content: '""',
          position: 'absolute',
          top: 0,
          left: 0,
          width: '35%',
          height: '100%',
          background: isDark
            ? 'linear-gradient(90deg, transparent, rgba(255, 255, 255, 0.015), transparent)'
            : 'linear-gradient(90deg, transparent, rgba(255, 255, 255, 0.08), transparent)',
          animation: 'shimmerSweep 8s ease-in-out infinite',
          animationDelay: `${delay + 3}s`,
          pointerEvents: 'none',
          zIndex: 1,
        },
        // Border glow animation
        '&::after': {
          content: '""',
          position: 'absolute',
          inset: -1,
          borderRadius: 'inherit',
          padding: '1px',
          background: isDark
            ? 'linear-gradient(135deg, rgba(255, 140, 0, 0.25), transparent 40%, transparent 60%, rgba(255, 140, 0, 0.15))'
            : 'linear-gradient(135deg, rgba(255, 107, 0, 0.15), transparent 40%, transparent 60%, rgba(255, 107, 0, 0.1))',
          mask: 'linear-gradient(#fff 0 0) content-box, linear-gradient(#fff 0 0)',
          maskComposite: 'exclude',
          WebkitMask: 'linear-gradient(#fff 0 0) content-box, linear-gradient(#fff 0 0)',
          WebkitMaskComposite: 'xor',
          animation: 'borderGlow 4s ease-in-out infinite',
          animationDelay: `${delay * 0.5}s`,
          pointerEvents: 'none',
          zIndex: 0,
        },
        // Inner top-edge light reflection
        backgroundImage: isDark
          ? 'linear-gradient(180deg, rgba(255, 255, 255, 0.04) 0%, transparent 40%)'
          : 'linear-gradient(180deg, rgba(255, 255, 255, 0.6) 0%, transparent 40%)',
        // Hover: border brightens subtly
        '&:hover': {
          borderColor: isDark ? 'rgba(255, 140, 0, 0.22)' : 'rgba(255, 107, 0, 0.18)',
        },
        ...sx,
      }}
    >
      {variant === 'social' && renderSocialLogin()}
      {variant === 'email' && renderEmailLogin()}
      {variant === 'mfa' && renderMfaLogin()}
    </Card>
  );
}
