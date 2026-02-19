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
import ReactLogo from '../icons/ReactLogo';
import NextLogo from '../icons/NextLogo';
import ExpressLogo from '../icons/ExpressLogo';
import FlutterLogo from '../icons/FlutterLogo';
import AndroidLogo from '../icons/AndroidLogo';
import IOSLogo from '../icons/IOSLogo';

interface StepProps {
  number: number;
  title: string;
  children: React.ReactNode;
}

function Step({number, title, children}: StepProps) {
  return (
    <Box
      sx={{
        display: 'flex',
        gap: {xs: 2, md: 4},
        alignItems: 'flex-start',
        position: 'relative',
      }}
    >
      {/* Left side: title + number */}
      <Box
        sx={{
          display: 'flex',
          alignItems: 'center',
          gap: 2,
          minWidth: {xs: 'auto', md: '320px'},
          flexDirection: {xs: 'column', md: 'row'},
          textAlign: {xs: 'center', md: 'right'},
          flexShrink: 0,
        }}
      >
        <Typography
          variant="body1"
          sx={{
            fontSize: {xs: '0.9rem', md: '1rem'},
            color: 'text.secondary',
            order: {xs: 2, md: 1},
            flex: 1,
            textAlign: {xs: 'center', md: 'right'},
          }}
        >
          {title}
        </Typography>
        <Box
          sx={{
            width: 36,
            height: 36,
            borderRadius: '50%',
            bgcolor: (theme) =>
              theme.palette.mode === 'dark' ? 'rgba(255, 255, 255, 0.1)' : 'rgba(0, 0, 0, 0.08)',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            fontWeight: 600,
            fontSize: '0.9rem',
            flexShrink: 0,
            order: {xs: 1, md: 2},
            border: '1px solid',
            borderColor: (theme) =>
              theme.palette.mode === 'dark' ? 'rgba(255, 255, 255, 0.15)' : 'rgba(0, 0, 0, 0.12)',
          }}
        >
          {number}
        </Box>
      </Box>

      {/* Right side: content */}
      <Box sx={{flex: 1}}>{children}</Box>
    </Box>
  );
}

function TechIconBox({children}: {children: React.ReactNode}) {
  return (
    <Box
      sx={{
        width: 52,
        height: 52,
        borderRadius: 2,
        border: '1px solid',
        borderColor: (theme) =>
          theme.palette.mode === 'dark' ? 'rgba(255, 255, 255, 0.12)' : 'rgba(0, 0, 0, 0.1)',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        bgcolor: (theme) =>
          theme.palette.mode === 'dark' ? 'rgba(255, 255, 255, 0.03)' : 'rgba(0, 0, 0, 0.02)',
        transition: 'border-color 0.2s ease',
        '&:hover': {
          borderColor: 'primary.main',
        },
      }}
    >
      {children}
    </Box>
  );
}

export default function GetStartedSection(): JSX.Element {
  return (
    <Box sx={{py: {xs: 8, lg: 12}}}>
      <Container maxWidth="lg" sx={{px: {xs: 2, sm: 4}}}>
        <Box sx={{textAlign: 'center', mb: 8}}>
          <Typography
            variant="h3"
            sx={{
              mb: 2,
              fontSize: {xs: '1.75rem', sm: '2.25rem', md: '2.5rem'},
              fontWeight: 700,
            }}
          >
            Get up and running in minutes
          </Typography>
          <Typography
            variant="body1"
            color="text.secondary"
            sx={{maxWidth: '600px', mx: 'auto', fontSize: {xs: '0.95rem', sm: '1.05rem'}}}
          >
            Seamless authentication made simple. Add login to your app in just 3 steps.
          </Typography>
        </Box>

        <Stack spacing={6} sx={{maxWidth: '800px', mx: 'auto'}}>
          {/* Step 1: Pick your technology */}
          <Step number={1} title="Pick your technology and register your app in Thunder">
            <Box sx={{display: 'flex', flexWrap: 'wrap', gap: 1.5}}>
              <TechIconBox><ReactLogo size={28} /></TechIconBox>
              <TechIconBox><NextLogo size={28} /></TechIconBox>
              <TechIconBox><ExpressLogo size={28} /></TechIconBox>
              <TechIconBox>
                <Typography sx={{fontSize: '1.2rem', opacity: 0.5}}>...</Typography>
              </TechIconBox>
            </Box>
            <Box sx={{display: 'flex', flexWrap: 'wrap', gap: 1.5, mt: 1.5}}>
              <TechIconBox><FlutterLogo size={28} /></TechIconBox>
              <TechIconBox><AndroidLogo size={28} /></TechIconBox>
              <TechIconBox>
                <Typography sx={{fontWeight: 500, fontSize: '0.75rem', opacity: 0.7}}>iOS</Typography>
              </TechIconBox>
            </Box>
          </Step>

          {/* Step 2: Install SDK */}
          <Step number={2} title="Install the SDK package">
            <Box
              sx={{
                bgcolor: (theme) =>
                  theme.palette.mode === 'dark' ? '#1a1a2e' : '#1e293b',
                borderRadius: 2,
                px: 3,
                py: 2,
                fontFamily: 'var(--ifm-font-family-monospace)',
                fontSize: '0.9rem',
                color: '#e2e8f0',
                border: '1px solid',
                borderColor: 'rgba(255, 255, 255, 0.08)',
              }}
            >
              <Box component="span" sx={{color: '#6b7280', mr: 1}}>{'>'}</Box>
              npm i{' '}
              <Box component="span" sx={{fontWeight: 600, color: '#fff'}}>@asgardeo/react</Box>
            </Box>
          </Step>

          {/* Step 3: Add login component */}
          <Step number={3} title="Add login component to your application">
            <Box
              sx={{
                bgcolor: (theme) =>
                  theme.palette.mode === 'dark' ? '#1a1a2e' : '#1e293b',
                borderRadius: 2,
                p: 3,
                fontFamily: 'var(--ifm-font-family-monospace)',
                fontSize: '0.85rem',
                lineHeight: 1.8,
                color: '#e2e8f0',
                border: '1px solid',
                borderColor: 'rgba(255, 255, 255, 0.08)',
                overflow: 'auto',
              }}
            >
              <Box component="span" sx={{color: '#c084fc'}}>import</Box>
              {' { '}
              <Box component="span" sx={{color: '#e2e8f0'}}>AsgardeoProvider</Box>
              {', '}
              <Box component="span" sx={{color: '#e2e8f0'}}>SignInButton</Box>
              {' } '}
              <Box component="span" sx={{color: '#c084fc'}}>from</Box>
              {' '}
              <Box component="span" sx={{color: '#fbbf24'}}>&apos;@asgardeo/react&apos;</Box>
              <br /><br />
              <Box component="span" sx={{color: '#c084fc'}}>function</Box>
              {' '}
              <Box component="span" sx={{fontWeight: 600, color: '#fff'}}>LoginPage</Box>
              {'() {'}
              <br />
              {'  '}
              <Box component="span" sx={{color: '#c084fc'}}>return</Box>
              {' ('}
              <br />
              {'    '}
              <Box component="span" sx={{color: '#7dd3fc'}}>{'<SignInButton />'}</Box>
              <br />
              {'  )'}
              <br />
              {'}'}
            </Box>
          </Step>
        </Stack>

        {/* CTAs */}
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
              borderColor: (theme) =>
                theme.palette.mode === 'dark' ? 'rgba(255, 255, 255, 0.3)' : 'rgba(0, 0, 0, 0.2)',
            }}
            startIcon={
              <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
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
              borderColor: (theme) =>
                theme.palette.mode === 'dark' ? 'rgba(255, 255, 255, 0.3)' : 'rgba(0, 0, 0, 0.2)',
            }}
            startIcon={
              <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
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
