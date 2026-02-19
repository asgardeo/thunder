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
import useIsDarkMode from '../../hooks/useIsDarkMode';
import useScrollAnimation from '../../hooks/useScrollAnimation';
import ReactLogo from '../icons/ReactLogo';
import NextLogo from '../icons/NextLogo';
import VueLogo from '../icons/VueLogo';
import ExpressLogo from '../icons/ExpressLogo';
import GoLogo from '../icons/GoLogo';
import FlutterLogo from '../icons/FlutterLogo';

interface StepProps {
  number: number;
  title: string;
  children: React.ReactNode;
  isLast?: boolean;
}

function Step({number, title, children, isLast = false}: StepProps) {
  const isDark = useIsDarkMode();

  return (
    <Box
      sx={{
        display: 'grid',
        gridTemplateColumns: {xs: '1fr', md: 'minmax(0, 240px) 36px minmax(0, 1fr)'},
        gap: {xs: 2, md: 0},
        alignItems: 'flex-start',
        position: 'relative',
      }}
    >
      {/* Title - right aligned */}
      <Typography
        variant="body1"
        sx={{
          fontSize: {xs: '0.9rem', md: '0.95rem'},
          color: isDark ? 'rgba(255, 255, 255, 0.5)' : 'rgba(0, 0, 0, 0.45)',
          textAlign: {xs: 'center', md: 'right'},
          pr: {md: 2.5},
          pt: '7px',
          lineHeight: 1.5,
          display: {xs: 'none', md: 'block'},
        }}
      >
        {title}
      </Typography>

      {/* Number circle column */}
      <Box
        sx={{
          position: 'relative',
          display: 'flex',
          justifyContent: 'center',
          alignSelf: 'stretch',
        }}
      >
        <Box
          sx={{
            width: 36,
            height: 36,
            borderRadius: '50%',
            bgcolor: isDark ? 'rgba(255, 255, 255, 0.08)' : 'rgba(0, 0, 0, 0.04)',
            color: isDark ? '#ffffff' : '#1a1a2e',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            fontWeight: 600,
            fontSize: '0.9rem',
            flexShrink: 0,
            border: '1px solid',
            borderColor: isDark ? 'rgba(255, 255, 255, 0.15)' : 'rgba(0, 0, 0, 0.1)',
            position: 'relative',
            zIndex: 1,
          }}
        >
          {number}
        </Box>
        {!isLast && (
          <Box
            sx={{
              position: 'absolute',
              top: 42,
              bottom: -46,
              left: '50%',
              transform: 'translateX(-50%)',
              width: 0,
              borderLeft: isDark ? '1px dashed rgba(255, 255, 255, 0.12)' : '1px dashed rgba(0, 0, 0, 0.1)',
              display: {xs: 'none', md: 'block'},
            }}
          />
        )}
      </Box>

      {/* Content */}
      <Box sx={{pl: {md: 3}}}>
        {/* Title shown on mobile only */}
        <Typography
          variant="body1"
          sx={{
            fontSize: '0.9rem',
            color: isDark ? 'rgba(255, 255, 255, 0.5)' : 'rgba(0, 0, 0, 0.45)',
            textAlign: 'center',
            mb: 1,
            display: {xs: 'block', md: 'none'},
          }}
        >
          {title}
        </Typography>
        {children}
      </Box>
    </Box>
  );
}

function TechIconBox({
  children,
  comingSoon = false,
  selected = false,
  onClick,
}: {
  children: React.ReactNode;
  comingSoon?: boolean;
  selected?: boolean;
  onClick?: () => void;
}) {
  const isDark = useIsDarkMode();

  return (
    <Box
      sx={{
        display: 'flex',
        flexDirection: 'column',
        alignItems: 'center',
        gap: 0.5,
        cursor: comingSoon ? 'default' : 'pointer',
      }}
      onClick={comingSoon ? undefined : onClick}
    >
      <Box
        sx={{
          width: 48,
          height: 48,
          borderRadius: 1.5,
          border: '1px solid',
          borderColor: selected
            ? 'rgba(255, 107, 0, 0.7)'
            : isDark
              ? 'rgba(255, 255, 255, 0.12)'
              : 'rgba(0, 0, 0, 0.1)',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          bgcolor: selected
            ? isDark
              ? 'rgba(255, 107, 0, 0.1)'
              : 'rgba(255, 107, 0, 0.06)'
            : isDark
              ? 'rgba(255, 255, 255, 0.03)'
              : 'rgba(0, 0, 0, 0.02)',
          transition: 'border-color 0.2s ease, background-color 0.2s ease, transform 0.2s ease',
          opacity: comingSoon ? 0.4 : 1,
          '&:hover': {
            borderColor: comingSoon ? undefined : 'rgba(255, 107, 0, 0.5)',
            transform: comingSoon ? undefined : 'translateY(-2px)',
          },
          '&:active': {
            transform: comingSoon ? undefined : 'translateY(0) scale(0.97)',
          },
        }}
      >
        {children}
      </Box>
      {comingSoon && (
        <Typography
          sx={{
            fontSize: '0.55rem',
            color: isDark ? 'rgba(255, 255, 255, 0.35)' : 'rgba(0, 0, 0, 0.35)',
            fontWeight: 500,
            whiteSpace: 'nowrap',
          }}
        >
          Coming Soon
        </Typography>
      )}
    </Box>
  );
}

export default function GetStartedSection(): JSX.Element {
  const isDark = useIsDarkMode();
  const {ref: titleRef, isVisible: titleVisible} = useScrollAnimation({threshold: 0.2});
  const {ref: stepsRef, isVisible: stepsVisible} = useScrollAnimation({threshold: 0.1});

  return (
    <Box
      sx={{
        py: {xs: 8, lg: 12},
        position: 'relative',
        background: isDark ? '#0a0a0a' : 'transparent',
        '&::before': {
          content: '""',
          position: 'absolute',
          top: 0,
          left: 0,
          right: 0,
          bottom: 0,
          background: isDark
            ? 'radial-gradient(ellipse at 50% 40%, rgba(255, 107, 0, 0.08) 0%, transparent 55%)'
            : 'radial-gradient(ellipse at 50% 40%, rgba(255, 107, 0, 0.05) 0%, transparent 55%)',
          pointerEvents: 'none',
        },
      }}
    >
      <Container maxWidth="lg" sx={{px: {xs: 2, sm: 4}, position: 'relative', zIndex: 1}}>
        <Box
          ref={titleRef}
          sx={{
            textAlign: 'center',
            mb: 8,
            opacity: titleVisible ? 1 : 0,
            transform: titleVisible ? 'translateY(0)' : 'translateY(32px)',
            transition: 'opacity 0.7s cubic-bezier(0.16, 1, 0.3, 1), transform 0.7s cubic-bezier(0.16, 1, 0.3, 1)',
          }}
        >
          <Typography
            variant="h3"
            sx={{
              mb: 2,
              fontSize: {xs: '1.75rem', sm: '2.25rem', md: '2.5rem'},
              fontWeight: 700,
              color: isDark ? '#ffffff' : '#1a1a2e',
            }}
          >
            Get up and running in minutes
          </Typography>
          <Typography
            variant="body1"
            sx={{
              maxWidth: '600px',
              mx: 'auto',
              fontSize: {xs: '0.95rem', sm: '1.05rem'},
              color: isDark ? 'rgba(255, 255, 255, 0.6)' : 'rgba(0, 0, 0, 0.55)',
            }}
          >
            Seamless authentication made simple. Add login to your app in just 3 steps.
          </Typography>
        </Box>

        <Stack
          ref={stepsRef}
          spacing={5}
          sx={{
            maxWidth: '860px',
            mx: 'auto',
            opacity: stepsVisible ? 1 : 0,
            transform: stepsVisible ? 'translateY(0)' : 'translateY(32px)',
            transition:
              'opacity 0.7s cubic-bezier(0.16, 1, 0.3, 1) 0.15s, transform 0.7s cubic-bezier(0.16, 1, 0.3, 1) 0.15s',
          }}
        >
          <Step number={1} title="Pick your technology and register your app in Thunder">
            <Box sx={{display: 'flex', flexWrap: 'wrap', gap: 1.5, alignItems: 'flex-start'}}>
              <TechIconBox selected>
                <ReactLogo size={26} />
              </TechIconBox>
              <TechIconBox comingSoon>
                <NextLogo size={26} />
              </TechIconBox>
              <TechIconBox comingSoon>
                <VueLogo size={26} />
              </TechIconBox>
              <TechIconBox comingSoon>
                <ExpressLogo size={26} />
              </TechIconBox>
              <TechIconBox comingSoon>
                <GoLogo size={26} />
              </TechIconBox>
              <TechIconBox comingSoon>
                <FlutterLogo size={26} />
              </TechIconBox>
              <Box
                sx={{
                  display: 'flex',
                  alignItems: 'center',
                  height: 48,
                  px: 1,
                }}
              >
                <Typography
                  sx={{
                    fontSize: '0.75rem',
                    color: isDark ? 'rgba(255, 255, 255, 0.35)' : 'rgba(0, 0, 0, 0.35)',
                    fontWeight: 500,
                    fontStyle: 'italic',
                  }}
                >
                  and more
                </Typography>
              </Box>
            </Box>
          </Step>

          <Step number={2} title="Install the SDK package">
            <Box
              sx={{
                bgcolor: isDark ? '#1a1a1f' : '#f8f9fa',
                borderRadius: 2,
                px: 3,
                py: 2,
                fontFamily: 'var(--ifm-font-family-monospace)',
                fontSize: '0.9rem',
                color: isDark ? '#e2e8f0' : '#1a1a2e',
                border: '1px solid',
                borderColor: isDark ? 'rgba(255, 255, 255, 0.08)' : 'rgba(0, 0, 0, 0.08)',
              }}
            >
              <Box component="span" sx={{color: '#6b7280', mr: 1}}>
                {'>'}
              </Box>
              npm i{' '}
              <Box component="span" sx={{fontWeight: 600, color: isDark ? '#fff' : '#1a1a2e'}}>
                @asgardeo/react
              </Box>
            </Box>
          </Step>

          <Step number={3} title="Integrate the SDK" isLast>
            <Box
              component="pre"
              sx={{
                bgcolor: isDark ? '#1a1a1f' : '#f8f9fa',
                borderRadius: 2,
                p: 3,
                fontFamily: 'var(--ifm-font-family-monospace)',
                fontSize: '0.82rem',
                lineHeight: 1.9,
                color: isDark ? '#e2e8f0' : '#1a1a2e',
                border: '1px solid',
                borderColor: isDark ? 'rgba(255, 255, 255, 0.08)' : 'rgba(0, 0, 0, 0.08)',
                overflow: 'auto',
                m: 0,
                whiteSpace: 'pre',
              }}
            >
              <Box component="span" sx={{color: isDark ? '#c084fc' : '#7c3aed'}}>
                import
              </Box>
              {' { StrictMode } '}
              <Box component="span" sx={{color: isDark ? '#c084fc' : '#7c3aed'}}>
                from
              </Box>{' '}
              <Box component="span" sx={{color: isDark ? '#fbbf24' : '#b45309'}}>{`'react'`}</Box>
              {'\n'}
              <Box component="span" sx={{color: isDark ? '#c084fc' : '#7c3aed'}}>
                import
              </Box>
              {' { createRoot } '}
              <Box component="span" sx={{color: isDark ? '#c084fc' : '#7c3aed'}}>
                from
              </Box>{' '}
              <Box component="span" sx={{color: isDark ? '#fbbf24' : '#b45309'}}>{`'react-dom/client'`}</Box>
              {'\n'}
              <Box component="span" sx={{color: isDark ? '#c084fc' : '#7c3aed'}}>
                import
              </Box>
              {' { AsgardeoProvider } '}
              <Box component="span" sx={{color: isDark ? '#c084fc' : '#7c3aed'}}>
                from
              </Box>{' '}
              <Box component="span" sx={{color: isDark ? '#fbbf24' : '#b45309'}}>{`'@asgardeo/react'`}</Box>
              {'\n'}
              <Box component="span" sx={{color: isDark ? '#c084fc' : '#7c3aed'}}>
                import
              </Box>
              {' App '}
              <Box component="span" sx={{color: isDark ? '#c084fc' : '#7c3aed'}}>
                from
              </Box>{' '}
              <Box component="span" sx={{color: isDark ? '#fbbf24' : '#b45309'}}>{`'./App.jsx'`}</Box>
              {'\n'}
              <Box component="span" sx={{color: isDark ? '#c084fc' : '#7c3aed'}}>
                import
              </Box>{' '}
              <Box component="span" sx={{color: isDark ? '#fbbf24' : '#b45309'}}>{`'./index.css'`}</Box>
              {'\n\n'}
              <Box component="span" sx={{color: isDark ? '#7dd3fc' : '#0284c7'}}>
                createRoot
              </Box>
              (document.
              <Box component="span" sx={{color: isDark ? '#7dd3fc' : '#0284c7'}}>
                getElementById
              </Box>
              (<Box component="span" sx={{color: isDark ? '#fbbf24' : '#b45309'}}>{`'root'`}</Box>
              )).
              <Box component="span" sx={{color: isDark ? '#7dd3fc' : '#0284c7'}}>
                render
              </Box>
              {'(\n  '}
              <Box component="span" sx={{color: isDark ? '#7dd3fc' : '#0284c7'}}>
                {'<StrictMode>'}
              </Box>
              {'\n    '}
              <Box component="span" sx={{color: isDark ? '#7dd3fc' : '#0284c7'}}>
                {'<AsgardeoProvider'}
              </Box>
              {'\n      '}
              <Box component="span" sx={{color: isDark ? '#93c5fd' : '#2563eb'}}>
                clientId
              </Box>
              =<Box component="span" sx={{color: isDark ? '#fbbf24' : '#b45309'}}>{`"<your-client-id>"`}</Box>
              {'\n      '}
              <Box component="span" sx={{color: isDark ? '#93c5fd' : '#2563eb'}}>
                baseUrl
              </Box>
              =<Box component="span" sx={{color: isDark ? '#fbbf24' : '#b45309'}}>{`"https://localhost:8090"`}</Box>
              {'\n      '}
              <Box component="span" sx={{color: isDark ? '#93c5fd' : '#2563eb'}}>
                platform
              </Box>
              =<Box component="span" sx={{color: isDark ? '#fbbf24' : '#b45309'}}>{`"AsgardeoV2"`}</Box>
              {'\n    '}
              <Box component="span" sx={{color: isDark ? '#7dd3fc' : '#0284c7'}}>
                {'>'}
              </Box>
              {'\n      '}
              <Box component="span" sx={{color: isDark ? '#7dd3fc' : '#0284c7'}}>
                {'<App />'}
              </Box>
              {'\n    '}
              <Box component="span" sx={{color: isDark ? '#7dd3fc' : '#0284c7'}}>
                {'</AsgardeoProvider>'}
              </Box>
              {'\n  '}
              <Box component="span" sx={{color: isDark ? '#7dd3fc' : '#0284c7'}}>
                {'</StrictMode>'}
              </Box>
              {'\n)'}
            </Box>
          </Step>
        </Stack>

        <Stack direction={{xs: 'column', sm: 'row'}} spacing={2} justifyContent="center" sx={{mt: 6}}>
          <Button
            component={Link}
            href="/docs/guides/introduction"
            variant="outlined"
            size="large"
            sx={{
              textTransform: 'none',
              borderRadius: 2,
              px: 3,
              borderColor: isDark ? 'rgba(255, 255, 255, 0.3)' : 'rgba(0, 0, 0, 0.2)',
              color: isDark ? '#ffffff' : '#1a1a2e',
              transition: 'transform 0.2s ease, border-color 0.2s ease, background-color 0.2s ease',
              '&:hover': {
                borderColor: isDark ? 'rgba(255, 255, 255, 0.5)' : 'rgba(0, 0, 0, 0.4)',
                bgcolor: isDark ? 'rgba(255, 255, 255, 0.05)' : 'rgba(0, 0, 0, 0.04)',
                transform: 'translateY(-2px)',
              },
              '&:active': {
                transform: 'translateY(0)',
              },
            }}
            startIcon={
              <svg
                width="16"
                height="16"
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                strokeWidth="2"
                strokeLinecap="round"
                strokeLinejoin="round"
              >
                <path d="M4.5 16.5c-1.5 1.26-2 5-2 5s3.74-.5 5-2c.71-.84.7-2.13-.09-2.91a2.18 2.18 0 0 0-2.91-.09z" />
                <path d="m12 15-3-3a22 22 0 0 1 2-3.95A12.88 12.88 0 0 1 22 2c0 2.72-.78 7.5-6 11a22.35 22.35 0 0 1-4 2z" />
              </svg>
            }
          >
            Try Quickstart
          </Button>
          <Button
            component={Link}
            href="/docs/sdks/overview"
            variant="outlined"
            size="large"
            sx={{
              textTransform: 'none',
              borderRadius: 2,
              px: 3,
              borderColor: isDark ? 'rgba(255, 255, 255, 0.3)' : 'rgba(0, 0, 0, 0.2)',
              color: isDark ? '#ffffff' : '#1a1a2e',
              transition: 'transform 0.2s ease, border-color 0.2s ease, background-color 0.2s ease',
              '&:hover': {
                borderColor: isDark ? 'rgba(255, 255, 255, 0.5)' : 'rgba(0, 0, 0, 0.4)',
                bgcolor: isDark ? 'rgba(255, 255, 255, 0.05)' : 'rgba(0, 0, 0, 0.04)',
                transform: 'translateY(-2px)',
              },
              '&:active': {
                transform: 'translateY(0)',
              },
            }}
            startIcon={
              <svg
                width="16"
                height="16"
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                strokeWidth="2"
                strokeLinecap="round"
                strokeLinejoin="round"
              >
                <path d="M4 19.5v-15A2.5 2.5 0 0 1 6.5 2H20v20H6.5a2.5 2.5 0 0 1 0-5H20" />
              </svg>
            }
          >
            Read SDK Docs
          </Button>
        </Stack>
      </Container>
    </Box>
  );
}
