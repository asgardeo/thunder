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
import {Box, Container, Typography} from '@wso2/oxygen-ui';
import GoogleLogo from '../icons/GoogleLogo';
import AWSLogo from '../icons/AWSLogo';
import HerokuLogo from '../icons/HerokuLogo';
import CiscoLogo from '../icons/CiscoLogo';
import LinkedInLogo from '../icons/LinkedInLogo';

const logos = [
  {Component: GoogleLogo, key: 'google'},
  {Component: AWSLogo, key: 'aws'},
  {Component: HerokuLogo, key: 'heroku'},
  {Component: CiscoLogo, key: 'cisco'},
  {Component: LinkedInLogo, key: 'linkedin'},
];

export default function SocialProofSection(): JSX.Element {
  return (
    <Box
      sx={{
        py: {xs: 8, lg: 10},
        textAlign: 'center',
        background: '#0a0a0a',
        color: '#ffffff',
      }}
    >
      <Container maxWidth="lg" sx={{px: {xs: 2, sm: 4}}}>
        <Typography
          variant="h3"
          sx={{
            mb: 2,
            fontSize: {xs: '1.75rem', sm: '2.25rem', md: '2.5rem'},
            fontWeight: 700,
            color: '#ffffff',
          }}
        >
          Production-ready identity. Built for your stack
        </Typography>
        <Typography
          variant="body1"
          sx={{
            mb: 6,
            maxWidth: '700px',
            mx: 'auto',
            fontSize: {xs: '0.95rem', sm: '1.05rem'},
            color: 'rgba(255, 255, 255, 0.6)',
          }}
        >
          Trusted by developers at the world&apos;s fastest-growing companies to build secure login
          experiences that customers love.
        </Typography>

        {/* Scrolling marquee */}
        <Box
          sx={{
            overflow: 'hidden',
            position: 'relative',
            '&::before, &::after': {
              content: '""',
              position: 'absolute',
              top: 0,
              bottom: 0,
              width: '100px',
              zIndex: 1,
              pointerEvents: 'none',
            },
            '&::before': {
              left: 0,
              background: 'linear-gradient(to right, #0a0a0a, transparent)',
            },
            '&::after': {
              right: 0,
              background: 'linear-gradient(to left, #0a0a0a, transparent)',
            },
          }}
        >
          <Box
            sx={{
              display: 'flex',
              alignItems: 'center',
              gap: {xs: 6, md: 8},
              animation: 'marquee-scroll 25s linear infinite',
              width: 'max-content',
              '&:hover': {
                animationPlayState: 'paused',
              },
              '@keyframes marquee-scroll': {
                '0%': {transform: 'translateX(0)'},
                '100%': {transform: 'translateX(-50%)'},
              },
            }}
          >
            {/* Duplicate logos for seamless loop */}
            {[...logos, ...logos].map((logo, i) => (
              <Box
                key={`${logo.key}-${i}`}
                sx={{
                  opacity: 0.5,
                  transition: 'opacity 0.3s ease',
                  display: 'flex',
                  alignItems: 'center',
                  flexShrink: 0,
                  '&:hover': {opacity: 0.8},
                }}
              >
                <logo.Component size={32} />
              </Box>
            ))}
          </Box>
        </Box>
      </Container>
    </Box>
  );
}
