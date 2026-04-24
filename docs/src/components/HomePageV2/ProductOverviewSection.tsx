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
import useBaseUrl from '@docusaurus/useBaseUrl';
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

function HighlightCard({icon, title, description, index, isVisible}: FeatureCardProps) {
  const isDark = useIsDarkMode();

  return (
    <Card
      sx={{
        p: 4,
        height: '100%',
        position: 'relative',
        overflow: 'hidden',
        background: isDark
          ? 'linear-gradient(160deg, rgba(255, 107, 0, 0.1) 0%, rgba(10, 10, 10, 0.95) 60%)'
          : 'linear-gradient(160deg, rgba(255, 107, 0, 0.07) 0%, rgba(255, 255, 255, 0.97) 60%)',
        border: '1px solid',
        borderColor: isDark ? 'rgba(255, 140, 0, 0.35)' : 'rgba(255, 107, 0, 0.3)',
        opacity: isVisible ? 1 : 0,
        transform: isVisible ? 'translateY(0)' : 'translateY(32px)',
        transitionProperty: 'opacity, transform, box-shadow',
        transitionDuration: '0.6s, 0.6s, 0.3s',
        transitionTimingFunction: 'cubic-bezier(0.16, 1, 0.3, 1)',
        transitionDelay: isVisible ? `${index * 0.07}s` : '0s',
        '&:hover': {
          transform: 'translateY(-6px)',
          boxShadow: isDark
            ? '0 20px 48px rgba(255, 107, 0, 0.22)'
            : '0 20px 48px rgba(255, 107, 0, 0.14)',
        },
        '&::before': {
          content: '""',
          position: 'absolute',
          top: 0,
          left: 0,
          right: 0,
          height: '3px',
          background: 'linear-gradient(90deg, #FF6B00 0%, #FF8C00 50%, #FFB347 100%)',
        },
      }}
    >
      <Box
        sx={{
          width: 56,
          height: 56,
          mb: 3,
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          borderRadius: '14px',
          background: 'linear-gradient(135deg, rgba(255, 107, 0, 0.22) 0%, rgba(255, 140, 0, 0.12) 100%)',
          color: '#FF6B00',
          boxShadow: '0 4px 16px rgba(255, 107, 0, 0.2)',
          transition: 'transform 0.3s ease, box-shadow 0.3s ease',
          '.MuiCard-root:hover &': {
            transform: 'scale(1.12)',
            boxShadow: '0 8px 24px rgba(255, 107, 0, 0.32)',
          },
        }}
      >
        {icon}
      </Box>
      <Typography
        variant="h6"
        sx={{
          fontWeight: 700,
          mb: 1.5,
          fontSize: '1.05rem',
          letterSpacing: '-0.01em',
          color: isDark ? '#ffffff' : '#1a1a2e',
        }}
      >
        {title}
      </Typography>
      <Typography
        variant="body2"
        sx={{
          fontSize: '0.875rem',
          lineHeight: 1.75,
          color: isDark ? 'rgba(255, 255, 255, 0.65)' : 'rgba(0, 0, 0, 0.6)',
        }}
      >
        {description}
      </Typography>
    </Card>
  );
}

function FeatureCard({icon, title, description, index, isVisible}: FeatureCardProps) {
  const isDark = useIsDarkMode();

  return (
    <Card
      sx={{
        p: 3,
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
        transitionDelay: isVisible ? `${index * 0.07}s` : '0s',
        '&:hover': {
          transform: 'translateY(-4px)',
          borderColor: 'rgba(255, 107, 0, 0.5)',
          boxShadow: isDark ? '0 8px 24px rgba(255, 107, 0, 0.1)' : '0 8px 24px rgba(255, 107, 0, 0.08)',
        },
      }}
    >
      <Box
        sx={{
          width: 44,
          height: 44,
          mb: 2,
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          borderRadius: '10px',
          bgcolor: isDark ? 'rgba(255, 140, 0, 0.1)' : 'rgba(255, 140, 0, 0.08)',
          color: '#FF8C00',
          transition: 'transform 0.3s ease',
          '.MuiCard-root:hover &': {transform: 'scale(1.1)'},
        }}
      >
        {icon}
      </Box>
      <Typography variant="h6" sx={{fontWeight: 600, mb: 1, fontSize: '0.95rem', color: isDark ? '#ffffff' : '#1a1a2e'}}>
        {title}
      </Typography>
      <Typography variant="body2" sx={{fontSize: '0.85rem', lineHeight: 1.65, color: isDark ? 'rgba(255, 255, 255, 0.55)' : 'rgba(0, 0, 0, 0.55)'}}>
        {description}
      </Typography>
    </Card>
  );
}

const features = [
  {
    icon: (
      <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round">
        <rect x="2" y="6" width="20" height="14" rx="3" />
        <path d="M8 6V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2" />
        <circle cx="9" cy="13" r="1.5" />
        <circle cx="15" cy="13" r="1.5" />
        <path d="M9 17h6" />
      </svg>
    ),
    title: 'Native agent identity',
    description:
      'Engineered with native Agent ID and inherent agentic AI capabilities to secure end-to-end workflows among humans, agents, and resources, including full MCP and A2A Authorization',
  },
  {
    icon: (
      <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round">
        <path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z" />
        <path d="M9 10l1.5 1.5L15 8" />
        <circle cx="12" cy="13" r="1" />
      </svg>
    ),
    title: 'Post-quantum ready',
    description:
      'Built upon a Post-Quantum Cryptographic (PQC) foundation to be inherently resistant to "Harvest Now, Decrypt Later" and "Trust Now, Forge Later" attacks and crypto agility by design.',
  },
  {
    icon: (
      <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round">
        <rect x="3" y="3" width="6" height="6" rx="1" />
        <rect x="15" y="3" width="6" height="6" rx="1" />
        <rect x="9" y="15" width="6" height="6" rx="1" />
        <path d="M6 9v3a3 3 0 0 0 3 3" />
        <path d="M18 9v3a3 3 0 0 1-3 3" />
      </svg>
    ),
    title: 'Built for how you work',
    description: 'Work your way, whether you are an app developer, IAM architect, or system admin. Your workflows, your default toolbox.',
  },
  {
    icon: (
      <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round">
        <path d="M13 2L4.5 13.5H11L10 22L20.5 10H14L13 2z" />
      </svg>
    ),
    title: 'Lightweight and high performing',
    description: 'A sub-25MB Go runtime with minimal resource overhead, optimized for speed and efficiency for edge use cases.',
  },
  {
    icon: (
      <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round">
        <circle cx="5" cy="6" r="2" />
        <circle cx="19" cy="6" r="2" />
        <circle cx="12" cy="18" r="2" />
        <path d="M7 6h10" />
        <path d="M6.5 7.5L12 16" />
        <path d="M17.5 7.5L12 16" />
      </svg>
    ),
    title: 'Every journey is a flow',
    description: 'Model and orchestrate identity journeys as composable flows using a drag-&-drop visual flow builder. ',
  },
  {
    icon: (
      <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round">
        <polyline points="16 18 22 12 16 6" />
        <polyline points="8 6 2 12 8 18" />
      </svg>
    ),
    title: 'API-first identity as code',
    description: 'Every capability is accessible programmatically over a secure, modern RESTful API, enabling you to build, deploy, and manage identity as code.',
  },
  {
    icon: (
      <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round">
        <path d="M21 16V8a2 2 0 0 0-1-1.73l-7-4a2 2 0 0 0-2 0l-7 4A2 2 0 0 0 3 8v8a2 2 0 0 0 1 1.73l7 4a2 2 0 0 0 2 0l7-4A2 2 0 0 0 21 16z" />
        <polyline points="3.29 7 12 12 20.71 7" />
        <line x1="12" y1="22" x2="12" y2="12" />
      </svg>
    ),
    title: 'Developer-first SDKs',
    description: 'Use drop-in UI components from pixel-perfect SDKs for React, Next.js, and more, and style with your own CSS.',
  },
  {
    icon: (
      <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round">
        <circle cx="12" cy="8" r="6" />
        <path d="M8.56 2.75c4.37 6.03 6.02 9.42 8.03 17.72m2.54-15.38c-3.72 4.35-8.94 5.66-16.88 5.85m19.5 1.9c-3.5-.93-6.63-.82-8.94 0-2.58.92-5.01 2.86-7.44 6.32" />
      </svg>
    ),
    title: 'Standards-first identity engine',
    description: 'Built on proven open standards including OpenID Connect, OAuth2, SCIM, and SAML and designed to evolve with next-generation standards.',
  },
  {
    icon: (
      <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round">
        <rect x="2" y="14" width="20" height="8" rx="2" />
        <path d="M6 14v-2a6 6 0 0 1 12 0v2" />
        <line x1="6" y1="18" x2="6.01" y2="18" />
        <line x1="10" y1="18" x2="10.01" y2="18" />
      </svg>
    ),
    title: 'Agnostic infrastructure and deployment',
    description: 'Deploy where your workloads live without infrastructure lock-in using a GitOps-driven approach and deployment artifacts for Kubernetes, Docker, and Helm.',
  },
];

export default function ProductOverviewSection(): JSX.Element {
  const isDark = useIsDarkMode();
  const diagramUrl = useBaseUrl('/assets/images/diagram.png');
  const {ref: titleRef, isVisible: titleVisible} = useScrollAnimation({threshold: 0.2});
  const {ref, isVisible} = useScrollAnimation({threshold: 0.05});

  return (
    <Box sx={{py: {xs: 8, lg: 12}, background: isDark ? '#0a0a0a' : 'transparent'}}>
      <Container maxWidth="lg" sx={{px: {xs: 2, sm: 4}}}>
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
            What is{' '}
            <Box
              component="span"
              sx={{
                background: 'linear-gradient(90deg, #FF6B00 0%, #FF8C00 100%)',
                WebkitBackgroundClip: 'text',
                WebkitTextFillColor: 'transparent',
                backgroundClip: 'text',
              }}
            >
              ThunderID?
            </Box>
          </Typography>
          <Typography
            variant="body1"
            sx={{
              maxWidth: '680px',
              mx: 'auto',
              fontSize: {xs: '0.95rem', sm: '1.05rem'},
              lineHeight: 1.7,
              color: isDark ? 'rgba(255, 255, 255, 0.6)' : 'rgba(0, 0, 0, 0.55)',
            }}
          >
            Thunder is an open source IAM stack built in Go, focused on open standards and designed to handle identity for humans, AI agents, and workloads with fully orchestratable identity flows.
          </Typography>
        </Box>

        <Box
          ref={ref}
          sx={{
            display: 'grid',
            gridTemplateColumns: {xs: '1fr', sm: 'repeat(2, 1fr)', md: 'repeat(3, 1fr)'},
            gap: 3,
          }}
        >
          {features.slice(0, 3).map((feature, index) => (
            <HighlightCard key={feature.title} {...feature} index={index} isVisible={isVisible} />
          ))}
          <Box
            sx={{
              gridColumn: '1 / -1',
              my: 2,
              borderRadius: '12px',
              overflow: 'hidden',
              border: '1px solid',
              borderColor: isDark ? 'rgba(255, 140, 0, 0.15)' : 'rgba(255, 140, 0, 0.2)',
              bgcolor: isDark ? 'rgba(255, 255, 255, 0.03)' : 'rgba(0, 0, 0, 0.02)',
              transition: 'transform 0.3s ease, border-color 0.3s ease, box-shadow 0.3s ease',
              '&:hover': {
                transform: 'translateY(-4px)',
                borderColor: 'rgba(255, 107, 0, 0.5)',
                boxShadow: isDark ? '0 8px 24px rgba(255, 107, 0, 0.1)' : '0 8px 24px rgba(255, 107, 0, 0.08)',
              },
            }}
          >
            <img
              src={diagramUrl}
              alt="ThunderID Architecture Diagram"
              style={{width: '100%', display: 'block'}}
            />
          </Box>
          {features.slice(3).map((feature, index) => (
            <FeatureCard key={feature.title} {...feature} index={index + 3} isVisible={isVisible} />
          ))}
        </Box>
      </Container>
    </Box>
  );
}
