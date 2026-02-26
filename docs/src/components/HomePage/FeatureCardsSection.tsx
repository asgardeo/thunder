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

import React, {JSX, ReactNode} from 'react';
import {Box, Card, Container, Typography} from '@wso2/oxygen-ui';
import useIsDarkMode from '../../hooks/useIsDarkMode';
import useScrollAnimation from '../../hooks/useScrollAnimation';

interface FeatureCardProps {
  icon: ReactNode;
  title: string;
  description: string;
  index: number;
  isVisible: boolean;
}

function FeatureCard({icon, title, description, index, isVisible}: FeatureCardProps) {
  const isDark = useIsDarkMode();

  return (
    <Card
      sx={{
        p: 3,
        textAlign: 'center',
        height: '100%',
        transition: 'transform 0.3s ease, border-color 0.3s ease, box-shadow 0.3s ease',
        bgcolor: isDark ? 'rgba(255, 255, 255, 0.03)' : 'rgba(0, 0, 0, 0.02)',
        border: '1px solid',
        borderColor: isDark ? 'rgba(255, 140, 0, 0.15)' : 'rgba(255, 140, 0, 0.2)',
        opacity: isVisible ? 1 : 0,
        transform: isVisible ? 'translateY(0)' : 'translateY(32px)',
        transitionProperty: 'opacity, transform, border-color, box-shadow',
        transitionDuration: '0.6s, 0.6s, 0.3s, 0.3s',
        transitionTimingFunction: 'cubic-bezier(0.16, 1, 0.3, 1)',
        transitionDelay: isVisible ? `${index * 0.1}s` : '0s',
        '&:hover': {
          transform: 'translateY(-4px)',
          borderColor: 'rgba(255, 107, 0, 0.5)',
          boxShadow: isDark ? '0 8px 24px rgba(255, 107, 0, 0.1)' : '0 8px 24px rgba(255, 107, 0, 0.08)',
        },
      }}
    >
      <Box
        sx={{
          width: 48,
          height: 48,
          mx: 'auto',
          mb: 2,
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          color: isDark ? '#ffffff' : '#1a1a2e',
          opacity: 0.8,
          transition: 'transform 0.3s ease',
          '.MuiCard-root:hover &': {
            transform: 'scale(1.1)',
          },
        }}
      >
        {icon}
      </Box>
      <Typography variant="h6" sx={{fontWeight: 600, mb: 1, fontSize: '1rem', color: isDark ? '#ffffff' : '#1a1a2e'}}>
        {title}
      </Typography>
      <Typography variant="body2" sx={{fontSize: '0.85rem', lineHeight: 1.6, color: isDark ? 'rgba(255, 255, 255, 0.6)' : 'rgba(0, 0, 0, 0.55)'}}>
        {description}
      </Typography>
    </Card>
  );
}

function RocketIcon() {
  return (
    <svg width="28" height="28" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round">
      <path d="M4.5 16.5c-1.5 1.26-2 5-2 5s3.74-.5 5-2c.71-.84.7-2.13-.09-2.91a2.18 2.18 0 0 0-2.91-.09z" />
      <path d="m12 15-3-3a22 22 0 0 1 2-3.95A12.88 12.88 0 0 1 22 2c0 2.72-.78 7.5-6 11a22.35 22.35 0 0 1-4 2z" />
      <path d="M9 12H4s.55-3.03 2-4c1.62-1.08 5 0 5 0" />
      <path d="M12 15v5s3.03-.55 4-2c1.08-1.62 0-5 0-5" />
    </svg>
  );
}

function CheckSquareIcon() {
  return (
    <svg width="28" height="28" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round">
      <path d="m9 11 3 3L22 4" />
      <path d="M21 12v7a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h11" />
    </svg>
  );
}

function GlobeShieldIcon() {
  return (
    <svg width="28" height="28" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round">
      <circle cx="12" cy="12" r="10" />
      <path d="M2 12h20" />
      <path d="M12 2a15.3 15.3 0 0 1 4 10 15.3 15.3 0 0 1-4 10 15.3 15.3 0 0 1-4-10 15.3 15.3 0 0 1 4-10z" />
    </svg>
  );
}

function ShieldCheckIcon() {
  return (
    <svg width="28" height="28" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round">
      <path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z" />
      <path d="m9 12 2 2 4-4" />
    </svg>
  );
}

const features = [
  {
    icon: <RocketIcon />,
    title: 'Proven Open Source',
    description: 'Built on open source tech trusted to secure over one billion identities worldwide',
  },
  {
    icon: <CheckSquareIcon />,
    title: 'Start Free and Scale',
    description: 'Start free and scale as you grow. No credit card required',
  },
  {
    icon: <GlobeShieldIcon />,
    title: 'Any App, Any Stack, Anywhere',
    description: 'From mobile to SPAs to server-side apps. From self deploy to SaaS platforms',
  },
  {
    icon: <ShieldCheckIcon />,
    title: 'Standards You Trust',
    description: 'OAuth 2.1, OpenID Connect, SAML 2.0 and SCIM 2.0',
  },
];

export default function FeatureCardsSection(): JSX.Element {
  const isDark = useIsDarkMode();
  const {ref, isVisible} = useScrollAnimation({threshold: 0.1});

  return (
    <Box sx={{py: {xs: 4, lg: 6}, background: isDark ? '#0a0a0a' : 'transparent'}}>
      <Container maxWidth="lg" sx={{px: {xs: 2, sm: 4}}}>
        <Box
          ref={ref}
          sx={{
            display: 'grid',
            gridTemplateColumns: {xs: '1fr', sm: 'repeat(2, 1fr)', md: 'repeat(4, 1fr)'},
            gap: 3,
          }}
        >
          {features.map((feature, index) => (
            <FeatureCard key={feature.title} {...feature} index={index} isVisible={isVisible} />
          ))}
        </Box>
      </Container>
    </Box>
  );
}
