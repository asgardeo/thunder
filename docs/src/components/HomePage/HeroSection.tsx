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

import React, {JSX} from 'react';
import Link from '@docusaurus/Link';
import {Box, Container, Typography, Stack, Button} from '@wso2/oxygen-ui';
import LoginBox from '../LoginBox';

export default function HeroSection(): JSX.Element {
  return (
    <Box
      sx={{
        py: {xs: 7, lg: 10},
        position: 'relative',
        overflow: 'hidden',
        background: (theme) =>
          theme.palette.mode === 'dark'
            ? 'radial-gradient(ellipse at 50% 0%, rgba(255, 107, 0, 0.08) 0%, transparent 60%)'
            : 'radial-gradient(ellipse at 50% 0%, rgba(255, 107, 0, 0.04) 0%, transparent 60%)',
        '&::before': {
          content: '""',
          position: 'absolute',
          top: 0,
          left: 0,
          right: 0,
          bottom: 0,
          backgroundImage: (theme) =>
            theme.palette.mode === 'dark'
              ? `radial-gradient(circle, rgba(255, 255, 255, 0.08) 1px, transparent 1px)`
              : `radial-gradient(circle, rgba(0, 0, 0, 0.04) 1px, transparent 1px)`,
          backgroundSize: '32px 32px',
          opacity: 0.6,
          pointerEvents: 'none',
        },
      }}
    >
      <Container maxWidth="lg" sx={{px: {xs: 2, sm: 4}, position: 'relative', zIndex: 1}}>
        <Box
          sx={{
            display: 'flex',
            flexDirection: 'column',
            alignItems: 'center',
            justifyContent: 'center',
            py: {xs: 5, lg: 8},
            textAlign: 'center',
          }}
        >
          <Typography
            variant="h1"
            sx={{
              mb: 2,
              fontSize: {xs: '2.75rem', sm: '3.5rem', md: '4rem'},
              fontWeight: 700,
              lineHeight: 1.1,
            }}
          >
            <Box
              component="span"
              sx={{
                background: 'linear-gradient(90deg, #FF6B00 0%, #FF8C00 100%)',
                WebkitBackgroundClip: 'text',
                WebkitTextFillColor: 'transparent',
                backgroundClip: 'text',
              }}
            >
              Auth
            </Box>{' '}
            for Modern Apps
          </Typography>

          <Typography
            variant="body1"
            sx={{
              mb: 2,
              fontSize: {xs: '1rem', sm: '1.15rem'},
              color: 'text.secondary',
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              gap: 1,
              flexWrap: 'wrap',
            }}
          >
            Drop-in components for{' '}
            <Box
              component="span"
              sx={{
                fontFamily: 'var(--ifm-font-family-monospace)',
                fontSize: '0.9em',
                bgcolor: (theme) =>
                  theme.palette.mode === 'dark' ? 'rgba(255, 255, 255, 0.1)' : 'rgba(0, 0, 0, 0.08)',
                px: 1.2,
                py: 0.3,
                borderRadius: 1,
                border: '1px solid',
                borderColor: (theme) =>
                  theme.palette.mode === 'dark' ? 'rgba(255, 255, 255, 0.15)' : 'rgba(0, 0, 0, 0.12)',
              }}
            >
              React.js
            </Box>{' '}
            and more
          </Typography>

          <Typography
            variant="body1"
            color="text.secondary"
            sx={{
              maxWidth: '720px',
              textAlign: 'center',
              mb: 4,
              fontSize: {xs: '0.95rem', sm: '1.05rem'},
              lineHeight: 1.7,
            }}
          >
            Add secure login with SSO and MFA, user management, Role-Based Access Control (RBAC),
            and much more &mdash; get your app up and running in minutes.
          </Typography>

          <Stack direction={{xs: 'column', sm: 'row'}} spacing={2} sx={{mb: 2}} alignItems="center">
            <Button
              component={Link}
              href="/docs/guides/introduction"
              variant="contained"
              color="primary"
              size="large"
              sx={{
                px: 4,
                py: 1.5,
                fontWeight: 600,
                textTransform: 'none',
                fontSize: '1rem',
                borderRadius: 2,
                background: 'linear-gradient(135deg, #FF6B00 0%, #FF8C00 100%)',
                '&:hover': {
                  background: 'linear-gradient(135deg, #e65e00 0%, #e67d00 100%)',
                },
              }}
            >
              Start building for <Box component="span" sx={{fontWeight: 800, ml: 0.5}}>FREE</Box>
            </Button>
            <Button
              component={Link}
              href="/docs/guides/introduction"
              variant="outlined"
              size="large"
              sx={{
                px: 3,
                py: 1.5,
                textTransform: 'none',
                fontSize: '1rem',
                borderRadius: 2,
                borderColor: (theme) =>
                  theme.palette.mode === 'dark' ? 'rgba(255, 255, 255, 0.3)' : 'rgba(0, 0, 0, 0.2)',
                color: (theme) =>
                  theme.palette.mode === 'dark' ? '#fff' : 'text.primary',
              }}
              startIcon={
                <Box component="span" sx={{fontSize: '1.2rem'}}>
                  &#x2728;
                </Box>
              }
            >
              Build with your AI Helper
            </Button>
          </Stack>

          <Typography
            variant="body2"
            sx={{
              mb: 8,
              fontSize: '0.85rem',
              color: 'text.secondary',
              opacity: 0.7,
              cursor: 'pointer',
              '&:hover': {textDecoration: 'underline'},
            }}
          >
            Connect your AI supported IDE
          </Typography>

          {/* Login Box Showcase */}
          <Box
            sx={{
              mt: 2,
              display: 'flex',
              gap: 3,
              flexWrap: 'nowrap',
              alignItems: 'center',
              justifyContent: 'center',
              maxWidth: '1100px',
              perspective: '1000px',
              '& > *': {
                transition: 'transform 0.4s ease',
                '&:hover': {
                  transform: 'translateY(-2px) scale(1.02)',
                },
              },
            }}
          >
            <LoginBox variant="social" sx={{mr: {xs: 0, md: '-100px'}, display: {xs: 'none', md: 'block'}}} />
            <LoginBox variant="email" sx={{zIndex: 1}} />
            <LoginBox variant="mfa" sx={{ml: {xs: 0, md: '-100px'}, display: {xs: 'none', md: 'block'}}} />
          </Box>
        </Box>
      </Container>
    </Box>
  );
}
