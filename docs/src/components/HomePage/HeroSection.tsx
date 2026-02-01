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
import {useBaseUrlUtils} from '@docusaurus/useBaseUrl';
import {Box, Container, Typography, Stack, Button} from '@wso2/oxygen-ui';
import {GithubIcon} from '@wso2/oxygen-ui-icons-react';
import ThemedImage from '@theme/ThemedImage';
import LoginBox from '../LoginBox';

function BoltIcon() {
  return (
    <Box
      component="svg"
      viewBox="495 168 360 654"
      sx={{
        width: 100,
        filter: 'drop-shadow(0 0 5px rgba(234, 179, 8, 1)) drop-shadow(0 0 20px rgba(234, 179, 8, 1))',
        gridArea: 'logo',
        placeSelf: 'center end',
        animation: 'bolt-pulse 2s ease-in-out infinite',
        '@keyframes bolt-pulse': {
          '0%, 100%': {
            filter: 'drop-shadow(0 0 5px rgba(234, 179, 8, 0.8)) drop-shadow(0 0 20px rgba(234, 179, 8, 0.8))',
            transform: 'scale(1)',
          },
          '50%': {
            filter: 'drop-shadow(0 0 10px rgba(234, 179, 8, 1)) drop-shadow(0 0 30px rgba(234, 179, 8, 1))',
            transform: 'scale(1.05)',
          },
        },
        '& .outer': {
          fill: (theme) => (theme.palette.mode === 'dark' ? '#fbbf24' : '#eab308'),
        },
        '& .inner': {
          fill: (theme) => (theme.palette.mode === 'dark' ? '#fef08a' : '#facc15'),
        },
      }}
    >
      <path
        d="M594.41 805c-.71 0-1.43-.15-2.11-.47a5.015 5.015 0 0 1-2.72-5.83l67.98-253.71H517.11a5.03 5.03 0 0 1-4.44-2.69c-.86-1.65-.73-3.65.34-5.18l26.85-38.35q25.56-36.51 104.91-149.83l106.31-151.82a4.99 4.99 0 0 1 6.21-1.66 5.015 5.015 0 0 1 2.72 5.83L692.03 455h140.45c1.86 0 3.57 1.04 4.43 2.69s.73 3.65-.34 5.18l-238.07 340a5 5 0 0 1-4.1 2.13Zm-67.69-270h137.37c1.55 0 3.02.72 3.97 1.96.95 1.23 1.27 2.84.86 4.34l-62.33 232.61 216.29-308.9H685.52c-1.55 0-3.02-.72-3.97-1.96a4.97 4.97 0 0 1-.86-4.34l62.33-232.61-90.04 128.59q-79.35 113.32-104.91 149.83L526.73 535Z"
        className="outer"
      />
      <path
        d="M594.41 805c-.71 0-1.43-.15-2.11-.47a5.015 5.015 0 0 1-2.72-5.83l67.98-253.71H517.11a5.03 5.03 0 0 1-4.44-2.69c-.86-1.65-.73-3.65.34-5.18l26.85-38.35q25.56-36.51 104.91-149.83l106.31-151.82a4.99 4.99 0 0 1 6.21-1.66 5.015 5.015 0 0 1 2.72 5.83L692.03 455h140.45c1.86 0 3.57 1.04 4.43 2.69s.73 3.65-.34 5.18l-238.07 340a5 5 0 0 1-4.1 2.13Zm-67.69-270h137.37c1.55 0 3.02.72 3.97 1.96.95 1.23 1.27 2.84.86 4.34l-62.33 232.61 216.29-308.9H685.52c-1.55 0-3.02-.72-3.97-1.96a4.97 4.97 0 0 1-.86-4.34l62.33-232.61-90.04 128.59q-79.35 113.32-104.91 149.83L526.73 535Z"
        className="inner"
      />
    </Box>
  );
}

export default function HeroSection(): JSX.Element {
  const {withBaseUrl} = useBaseUrlUtils();
  return (
    <Box
      sx={{
        py: {xs: 7, lg: 10},
        position: 'relative',
        overflow: 'hidden',
        background: (theme) =>
          theme.palette.mode === 'dark'
            ? 'radial-gradient(circle at 50% 0%, rgba(99, 102, 241, 0.1) 0%, transparent 50%)'
            : 'radial-gradient(circle at 50% 0%, rgba(99, 102, 241, 0.05) 0%, transparent 50%)',
        '&::before': {
          content: '""',
          position: 'absolute',
          top: 0,
          left: 0,
          right: 0,
          bottom: 0,
          backgroundImage: (theme) =>
            theme.palette.mode === 'dark'
              ? 'linear-gradient(rgba(255, 255, 255, 0.05) 1px, transparent 1px), linear-gradient(90deg, rgba(255, 255, 255, 0.05) 1px, transparent 1px)'
              : 'linear-gradient(rgba(0, 0, 0, 0.03) 1px, transparent 1px), linear-gradient(90deg, rgba(0, 0, 0, 0.03) 1px, transparent 1px)',
          backgroundSize: '50px 50px',
          opacity: 0.4,
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
          <Stack direction="row" spacing={2} alignItems="center" sx={{mb: 2}}>
            <BoltIcon />
          </Stack>
          <Typography
            variant="body1"
            sx={{
              mb: 2,
              fontSize: {xs: '0.875rem', sm: '1rem'},
              color: 'text.secondary',
              letterSpacing: '0.1em',
              textTransform: 'uppercase',
            }}
          >
            Introducing
          </Typography>
          <Box sx={{mb: 3}}>
            <ThemedImage
              sources={{
                light: withBaseUrl('/assets/images/logo.svg'),
                dark: withBaseUrl('/assets/images/logo-inverted.svg'),
              }}
              alt="Thunder Logo"
              style={{height: 50}}
            />
          </Box>
          <Typography
            variant="h2"
            sx={{
              mb: 2,
              fontSize: {xs: '3.5rem', sm: '4rem'},
              fontWeight: 700,
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
            for the Modern Dev
          </Typography>
          <Typography
            variant="h5"
            color="text.secondary"
            sx={{
              maxWidth: '700px',
              textAlign: 'center',
              mb: 4,
              fontSize: {xs: '1.1rem', sm: '1.3rem'},
            }}
          >
            The world&apos;s most flexible, truly open source identity platform,
            <br />
            powered by open source innovation.
          </Typography>
          <Stack direction="row" spacing={2} sx={{mb: 8}}>
            <Button component={Link} href="/docs/guides/introduction" variant="contained" color="primary" size="large">
              Get Started
            </Button>
            <Button
              component={Link}
              href="https://github.com/asgardeo/thunder"
              target="_blank"
              variant="outlined"
              size="large"
              startIcon={<GithubIcon style={{fontSize: 20}} />}
            >
              Star on GitHub
            </Button>
          </Stack>

          {/* Login Box Showcase */}
          <Box
            sx={{
              mt: 4,
              display: 'flex',
              gap: 3,
              flexWrap: 'wrap',
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
            <LoginBox variant="social" sx={{mr: '-100px'}} />
            <LoginBox variant="email" sx={{zIndex: 1}} />
            <LoginBox variant="mfa" sx={{ml: '-100px'}} />
          </Box>
        </Box>
      </Container>
    </Box>
  );
}
