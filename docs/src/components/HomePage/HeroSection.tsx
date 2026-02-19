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

import React, {JSX, useEffect, useState} from 'react';
import Link from '@docusaurus/Link';
import {Box, Container, Typography, Stack, Button} from '@wso2/oxygen-ui';
import useIsDarkMode from '../../hooks/useIsDarkMode';
import LoginBox from '../LoginBox';
import ConstellationBackground from './ConstellationBackground';

export default function HeroSection(): JSX.Element {
  const isDark = useIsDarkMode();

  // After entry animations finish, clear them so CSS transitions can take over.
  // animation-fill-mode: both locks the transform, preventing smooth hover transitions.
  const [animDone, setAnimDone] = useState(false);

  useEffect(() => {
    const timer = setTimeout(() => setAnimDone(true), 1800);

    return () => clearTimeout(timer);
  }, []);

  return (
    <Box
      sx={{
        py: {xs: 7, lg: 10},
        position: 'relative',
        overflow: 'hidden',
        background: isDark ? '#0a0a0a' : 'transparent',
        '&::before': {
          content: '""',
          position: 'absolute',
          top: 0,
          left: 0,
          right: 0,
          bottom: 0,
          background: isDark
            ? 'radial-gradient(ellipse at 60% 35%, rgba(255, 107, 0, 0.10) 0%, transparent 50%)'
            : 'radial-gradient(ellipse at 60% 35%, rgba(255, 107, 0, 0.06) 0%, transparent 50%)',
          pointerEvents: 'none',
        },
      }}
    >
      <ConstellationBackground />
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
          {/* Lightning bolt icon with glow */}
          <Box
            sx={{
              mb: 3,
              position: 'relative',
              width: 80,
              height: 120,
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              animation: 'fadeInUp 0.8s cubic-bezier(0.16, 1, 0.3, 1) both',
            }}
          >
            {/* Glow effect */}
            <Box
              sx={{
                position: 'absolute',
                width: 120,
                height: 120,
                borderRadius: '50%',
                background: 'radial-gradient(circle, rgba(255, 170, 50, 0.25) 0%, transparent 70%)',
                filter: 'blur(20px)',
                animation: 'pulseGlow 3s ease-in-out infinite',
              }}
            />
            <svg width="56" height="80" viewBox="0 0 24 32" fill="none" style={{position: 'relative', zIndex: 1, animation: 'subtleFloat 4s ease-in-out infinite'}}>
              <path
                d="M13.5 1L4 18h7l-1.5 13L20 14h-7L13.5 1z"
                stroke="#FF8C00"
                strokeWidth="1.5"
                strokeLinecap="round"
                strokeLinejoin="round"
                fill="none"
              />
            </svg>
          </Box>

          {/* INTRODUCING label */}
          <Typography
            variant="overline"
            sx={{
              mb: 1.5,
              fontSize: '0.8rem',
              letterSpacing: '0.25em',
              color: isDark ? 'rgba(255, 170, 80, 0.8)' : 'rgba(200, 100, 0, 0.8)',
              fontWeight: 500,
              animation: 'fadeInUp 0.8s cubic-bezier(0.16, 1, 0.3, 1) 0.1s both',
            }}
          >
            INTRODUCING
          </Typography>

          {/* [ THUNDER ] title */}
          <Typography
            variant="h2"
            sx={{
              mb: 3,
              fontSize: {xs: '2rem', sm: '2.5rem', md: '3rem'},
              fontWeight: 300,
              letterSpacing: '0.15em',
              color: isDark ? '#ffffff' : '#1a1a2e',
              animation: 'fadeInUp 0.8s cubic-bezier(0.16, 1, 0.3, 1) 0.2s both',
            }}
          >
            {'[  THUNDER  ]'}
          </Typography>

          {/* Main heading */}
          <Typography
            variant="h1"
            sx={{
              mb: 3,
              fontSize: {xs: '2.75rem', sm: '3.5rem', md: '4.5rem'},
              fontWeight: 700,
              lineHeight: 1.1,
              color: isDark ? '#ffffff' : '#1a1a2e',
              animation: 'fadeInUp 0.8s cubic-bezier(0.16, 1, 0.3, 1) 0.3s both',
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

          {/* Description */}
          <Typography
            variant="body1"
            sx={{
              maxWidth: '680px',
              textAlign: 'center',
              mb: 5,
              fontSize: {xs: '1rem', sm: '1.15rem'},
              lineHeight: 1.7,
              color: isDark ? 'rgba(255, 255, 255, 0.6)' : 'rgba(0, 0, 0, 0.55)',
              animation: 'fadeInUp 0.8s cubic-bezier(0.16, 1, 0.3, 1) 0.4s both',
            }}
          >
            The world&apos;s most flexible, truly open source identity platform,
            powered by open source innovation.
          </Typography>

          {/* Buttons */}
          <Stack
            direction={{xs: 'column', sm: 'row'}}
            spacing={2}
            sx={{mb: 8, animation: 'fadeInUp 0.8s cubic-bezier(0.16, 1, 0.3, 1) 0.5s both'}}
            alignItems="center"
          >
            <Button
              component={Link}
              href="/docs/guides/introduction"
              variant="contained"
              color="primary"
              size="large"
              sx={{
                px: 5,
                py: 1.5,
                fontWeight: 600,
                textTransform: 'none',
                fontSize: '1.05rem',
                borderRadius: '28px',
                background: 'linear-gradient(135deg, #FF6B00 0%, #FF8C00 100%)',
                transition: 'transform 0.2s ease, box-shadow 0.2s ease, background 0.2s ease',
                '&:hover': {
                  background: 'linear-gradient(135deg, #e65e00 0%, #e67d00 100%)',
                  transform: 'translateY(-2px)',
                  boxShadow: '0 6px 20px rgba(255, 107, 0, 0.3)',
                },
                '&:active': {
                  transform: 'translateY(0)',
                },
              }}
            >
              Get Started
            </Button>
            <Button
              component={Link}
              href="https://github.com/asgardeo/thunder"
              variant="outlined"
              size="large"
              sx={{
                px: 4,
                py: 1.5,
                textTransform: 'none',
                fontSize: '1.05rem',
                borderRadius: '28px',
                borderColor: isDark ? 'rgba(255, 140, 0, 0.4)' : 'rgba(255, 107, 0, 0.5)',
                color: '#FF8C00',
                transition: 'transform 0.2s ease, border-color 0.2s ease, background-color 0.2s ease',
                '&:hover': {
                  borderColor: 'rgba(255, 140, 0, 0.7)',
                  bgcolor: 'rgba(255, 140, 0, 0.05)',
                  transform: 'translateY(-2px)',
                },
                '&:active': {
                  transform: 'translateY(0)',
                },
              }}
              startIcon={
                <svg width="20" height="20" viewBox="0 0 24 24" fill="none">
                  <path
                    d="M12 .297c-6.63 0-12 5.373-12 12 0 5.303 3.438 9.8 8.205 11.385.6.113.82-.258.82-.577 0-.285-.01-1.04-.015-2.04-3.338.724-4.042-1.61-4.042-1.61C4.422 18.07 3.633 17.7 3.633 17.7c-1.087-.744.084-.729.084-.729 1.205.084 1.838 1.236 1.838 1.236 1.07 1.835 2.809 1.305 3.495.998.108-.776.417-1.305.76-1.605-2.665-.3-5.466-1.332-5.466-5.93 0-1.31.465-2.38 1.235-3.22-.135-.303-.54-1.523.105-3.176 0 0 1.005-.322 3.3 1.23.96-.267 1.98-.399 3-.405 1.02.006 2.04.138 3 .405 2.28-1.552 3.285-1.23 3.285-1.23.645 1.653.24 2.873.12 3.176.765.84 1.23 1.91 1.23 3.22 0 4.61-2.805 5.625-5.475 5.92.42.36.81 1.096.81 2.22 0 1.606-.015 2.896-.015 3.286 0 .315.21.69.825.57C20.565 22.092 24 17.592 24 12.297c0-6.627-5.373-12-12-12"
                    fill="#FF8C00"
                  />
                </svg>
              }
            >
              Star on GitHub
            </Button>
          </Stack>

          {/* Login Box Showcase with dashed arc borders */}
          <Box
            sx={{
              mt: 2,
              position: 'relative',
              maxWidth: '1100px',
              width: '100%',
              mx: 'auto',
              animation: 'fadeInScale 1s cubic-bezier(0.16, 1, 0.3, 1) 0.6s both',
            }}
          >
            {/* Ambient glow behind card group */}
            <Box
              sx={{
                position: 'absolute',
                top: '50%',
                left: '50%',
                transform: 'translate(-50%, -45%)',
                width: '70%',
                height: '80%',
                background: isDark
                  ? 'radial-gradient(ellipse at center, rgba(255, 107, 0, 0.06) 0%, transparent 70%)'
                  : 'radial-gradient(ellipse at center, rgba(255, 107, 0, 0.04) 0%, transparent 70%)',
                pointerEvents: 'none',
                zIndex: 0,
                filter: 'blur(40px)',
                display: {xs: 'none', md: 'block'},
              }}
            />

            {/* Subtle connecting lines between cards */}
            <Box
              component="svg"
              sx={{
                position: 'absolute',
                top: 0,
                left: 0,
                width: '100%',
                height: '100%',
                pointerEvents: 'none',
                zIndex: 1,
                display: {xs: 'none', md: 'block'},
              }}
              viewBox="0 0 1100 600"
              preserveAspectRatio="xMidYMid slice"
            >
              {/* Left-to-center connecting line */}
              <line
                x1="320" y1="280" x2="430" y2="200"
                stroke={isDark ? 'rgba(255, 140, 0, 0.08)' : 'rgba(255, 107, 0, 0.06)'}
                strokeWidth="1"
                strokeDasharray="4 6"
                style={{animation: 'dashFlow 6s linear infinite'}}
              />
              {/* Center-to-right connecting line */}
              <line
                x1="670" y1="200" x2="780" y2="280"
                stroke={isDark ? 'rgba(255, 140, 0, 0.08)' : 'rgba(255, 107, 0, 0.06)'}
                strokeWidth="1"
                strokeDasharray="4 6"
                style={{animation: 'dashFlow 6s linear infinite'}}
              />
            </Box>

            {/* Floating particles */}
            <Box
              component="svg"
              sx={{
                position: 'absolute',
                top: 0,
                left: 0,
                width: '100%',
                height: '100%',
                pointerEvents: 'none',
                zIndex: 3,
                display: {xs: 'none', md: 'block'},
              }}
              viewBox="0 0 1100 600"
              preserveAspectRatio="xMidYMid slice"
            >
              {[
                {cx: 120, cy: 80, r: 2, dur: '6s', delay: '0s'},
                {cx: 980, cy: 100, r: 1.5, dur: '7s', delay: '1s'},
                {cx: 200, cy: 400, r: 1.5, dur: '5s', delay: '2s'},
                {cx: 900, cy: 420, r: 2, dur: '8s', delay: '0.5s'},
                {cx: 550, cy: 500, r: 1.5, dur: '6s', delay: '3s'},
                {cx: 50, cy: 250, r: 1.5, dur: '7s', delay: '1.5s'},
                {cx: 1050, cy: 280, r: 1.5, dur: '5.5s', delay: '2.5s'},
              ].map((dot, i) => (
                <circle key={i} cx={dot.cx} cy={dot.cy} r={dot.r} fill={isDark ? 'rgba(255, 140, 0, 0.3)' : 'rgba(255, 107, 0, 0.2)'}>
                  <animate attributeName="cy" values={`${dot.cy};${dot.cy - 15};${dot.cy}`} dur={dot.dur} begin={dot.delay} repeatCount="indefinite" />
                  <animate attributeName="opacity" values="0.15;0.5;0.15" dur={dot.dur} begin={dot.delay} repeatCount="indefinite" />
                </circle>
              ))}
            </Box>

            {/* Login cards container */}
            <Box
              sx={{
                display: 'flex',
                flexWrap: 'nowrap',
                alignItems: 'flex-start',
                justifyContent: 'center',
                position: 'relative',
                pt: {xs: 0, md: 4},
              }}
            >
              {/* Left card - social login */}
              <LoginBox
                variant="social"
                delay={0.3}
                sideCard
                sx={{
                  display: {xs: 'none', md: 'block'},
                  mr: '-60px',
                  mt: '40px',
                  transform: 'translateY(0px)',
                  opacity: 0.85,
                  zIndex: 0,
                  transition: 'transform 0.8s cubic-bezier(0.4, 0, 0.2, 1), opacity 0.8s cubic-bezier(0.4, 0, 0.2, 1), box-shadow 0.8s cubic-bezier(0.4, 0, 0.2, 1)',
                  ...(animDone
                    ? {}
                    : {animation: 'slideInLeft 0.9s cubic-bezier(0.16, 1, 0.3, 1) 0.8s both'}),
                  boxShadow: isDark
                    ? '0 12px 40px rgba(0, 0, 0, 0.4)'
                    : '0 12px 40px rgba(0, 0, 0, 0.06)',
                  '&:hover': {
                    transform: 'translateY(-4px)',
                    opacity: 0.95,
                    boxShadow: isDark
                      ? '0 16px 48px rgba(0, 0, 0, 0.45), 0 0 24px rgba(255, 140, 0, 0.04)'
                      : '0 16px 48px rgba(0, 0, 0, 0.08), 0 0 24px rgba(255, 107, 0, 0.03)',
                  },
                }}
              />

              {/* Center card - email login (most prominent) */}
              <LoginBox
                variant="email"
                delay={0}
                sx={{
                  zIndex: 2,
                  transform: 'translateY(0px)',
                  boxShadow: isDark
                    ? '0 20px 60px rgba(0, 0, 0, 0.5), 0 0 80px rgba(255, 107, 0, 0.1)'
                    : '0 20px 60px rgba(0, 0, 0, 0.1), 0 0 80px rgba(255, 107, 0, 0.07)',
                  transition: 'transform 0.8s cubic-bezier(0.4, 0, 0.2, 1), box-shadow 0.8s cubic-bezier(0.4, 0, 0.2, 1)',
                  ...(animDone
                    ? {}
                    : {animation: 'fadeInUp 0.9s cubic-bezier(0.16, 1, 0.3, 1) 0.7s both'}),
                  '&:hover': {
                    transform: 'translateY(-4px)',
                    boxShadow: isDark
                      ? '0 24px 68px rgba(0, 0, 0, 0.55), 0 0 90px rgba(255, 107, 0, 0.12)'
                      : '0 24px 68px rgba(0, 0, 0, 0.12), 0 0 90px rgba(255, 107, 0, 0.09)',
                  },
                }}
              />

              {/* Right card - MFA */}
              <LoginBox
                variant="mfa"
                delay={0.6}
                sideCard
                sx={{
                  display: {xs: 'none', md: 'block'},
                  ml: '-60px',
                  mt: '40px',
                  transform: 'translateY(0px)',
                  opacity: 0.85,
                  zIndex: 0,
                  transition: 'transform 0.8s cubic-bezier(0.4, 0, 0.2, 1), opacity 0.8s cubic-bezier(0.4, 0, 0.2, 1), box-shadow 0.8s cubic-bezier(0.4, 0, 0.2, 1)',
                  ...(animDone
                    ? {}
                    : {animation: 'slideInRight 0.9s cubic-bezier(0.16, 1, 0.3, 1) 0.8s both'}),
                  boxShadow: isDark
                    ? '0 12px 40px rgba(0, 0, 0, 0.4)'
                    : '0 12px 40px rgba(0, 0, 0, 0.06)',
                  '&:hover': {
                    transform: 'translateY(-4px)',
                    opacity: 0.95,
                    boxShadow: isDark
                      ? '0 16px 48px rgba(0, 0, 0, 0.45), 0 0 24px rgba(255, 140, 0, 0.04)'
                      : '0 16px 48px rgba(0, 0, 0, 0.08), 0 0 24px rgba(255, 107, 0, 0.03)',
                  },
                }}
              />
            </Box>
          </Box>
        </Box>
      </Container>
    </Box>
  );
}
