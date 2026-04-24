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
import {Box, Card, Container, Typography} from '@wso2/oxygen-ui';
import useIsDarkMode from '../../hooks/useIsDarkMode';
import useScrollAnimation from '../../hooks/useScrollAnimation';

interface CommunityCardProps {
  icon: JSX.Element;
  iconBg: string;
  title: string;
  description: string;
  linkLabel: string;
  href: string;
}

function CommunityCard({icon, iconBg, title, description, linkLabel, href}: CommunityCardProps) {
  const isDark = useIsDarkMode();

  return (
    <Card
      sx={{
        flex: 1,
        p: {xs: 3, sm: 4},
        pt: {xs: 4, sm: 5},
        pb: {xs: 3, sm: 4},
        display: 'flex',
        flexDirection: 'column',
        alignItems: 'center',
        textAlign: 'center',
        cursor: 'pointer',
        transition: 'all 0.3s ease',
        bgcolor: isDark ? 'rgba(255, 255, 255, 0.025)' : 'rgba(0, 0, 0, 0.02)',
        border: '1px solid',
        borderColor: isDark ? 'rgba(255, 255, 255, 0.08)' : 'rgba(0, 0, 0, 0.06)',
        borderRadius: '16px',
        '&:hover': {
          transform: 'translateY(-4px)',
          boxShadow: isDark ? '0 12px 32px rgba(0, 0, 0, 0.4)' : '0 12px 32px rgba(0, 0, 0, 0.1)',
          borderColor: 'rgba(255, 140, 0, 0.25)',
          bgcolor: isDark ? 'rgba(255, 255, 255, 0.035)' : 'rgba(0, 0, 0, 0.03)',
        },
      }}
      onClick={() => window.open(href, '_blank', 'noopener noreferrer')}
    >
      <Box
        sx={{
          width: 56,
          height: 56,
          borderRadius: '14px',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          background: iconBg,
          color: '#ffffff',
          mb: 3,
        }}
      >
        {icon}
      </Box>
      <Typography variant="h6" sx={{fontWeight: 600, mb: 1, color: isDark ? '#ffffff' : '#1a1a2e', fontSize: '1.1rem'}}>
        {title}
      </Typography>
      <Typography
        variant="body2"
        sx={{mb: 3, color: isDark ? 'rgba(255, 255, 255, 0.5)' : 'rgba(0, 0, 0, 0.45)', lineHeight: 1.7, fontSize: '0.9rem'}}
      >
        {description}
      </Typography>
      <Typography
        variant="body2"
        sx={{mt: 'auto', color: '#FF8C00', fontWeight: 500, fontSize: '0.9rem', transition: 'color 0.2s ease', '&:hover': {color: '#FF6B00'}}}
      >
        {linkLabel} &rarr;
      </Typography>
    </Card>
  );
}

function GitForkIcon() {
  return (
    <svg width="26" height="26" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
      <circle cx="12" cy="18" r="3" />
      <circle cx="6" cy="6" r="3" />
      <circle cx="18" cy="6" r="3" />
      <path d="M18 9v1a2 2 0 0 1-2 2H8a2 2 0 0 1-2-2V9" />
      <path d="M12 12v3" />
    </svg>
  );
}

function IssueIcon() {
  return (
    <svg width="26" height="26" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
      <circle cx="12" cy="12" r="10" />
      <line x1="12" y1="8" x2="12" y2="12" />
      <line x1="12" y1="16" x2="12.01" y2="16" />
    </svg>
  );
}

function DiscordIcon() {
  return (
    <svg width="26" height="26" viewBox="0 0 24 24" fill="currentColor">
      <path d="M20.317 4.37a19.791 19.791 0 0 0-4.885-1.515.074.074 0 0 0-.079.037c-.21.375-.444.864-.608 1.25a18.27 18.27 0 0 0-5.487 0 12.64 12.64 0 0 0-.617-1.25.077.077 0 0 0-.079-.037A19.736 19.736 0 0 0 3.677 4.37a.07.07 0 0 0-.032.027C.533 9.046-.32 13.58.099 18.057c.002.022.015.043.032.055a19.9 19.9 0 0 0 5.993 3.03.078.078 0 0 0 .084-.028c.462-.63.874-1.295 1.226-1.994a.076.076 0 0 0-.041-.106 13.107 13.107 0 0 1-1.872-.892.077.077 0 0 1-.008-.128 10.2 10.2 0 0 0 .372-.292.074.074 0 0 1 .077-.01c3.928 1.793 8.18 1.793 12.062 0a.074.074 0 0 1 .078.01c.12.098.246.198.373.292a.077.077 0 0 1-.006.127 12.299 12.299 0 0 1-1.873.892.077.077 0 0 0-.041.107c.36.698.772 1.362 1.225 1.993a.076.076 0 0 0 .084.028 19.839 19.839 0 0 0 6.002-3.03.077.077 0 0 0 .032-.054c.5-5.177-.838-9.674-3.549-13.66a.061.061 0 0 0-.031-.03zM8.02 15.33c-1.183 0-2.157-1.085-2.157-2.419 0-1.333.956-2.419 2.157-2.419 1.21 0 2.176 1.096 2.157 2.42 0 1.333-.956 2.418-2.157 2.418zm7.975 0c-1.183 0-2.157-1.085-2.157-2.419 0-1.333.955-2.419 2.157-2.419 1.21 0 2.176 1.096 2.157 2.42 0 1.333-.946 2.418-2.157 2.418z" />
    </svg>
  );
}

export default function CommunitySection(): JSX.Element {
  const isDark = useIsDarkMode();
  const {ref, isVisible} = useScrollAnimation({threshold: 0.15});

  return (
    <Box component="section" sx={{py: {xs: 8, lg: 12}, background: isDark ? '#0a0a0a' : 'transparent'}}>
      <Container maxWidth="lg" sx={{px: {xs: 2, sm: 4}}}>
        <Box
          ref={ref}
          sx={{
            display: 'flex',
            flexDirection: 'column',
            alignItems: 'center',
            textAlign: 'center',
            opacity: isVisible ? 1 : 0,
            transform: isVisible ? 'translateY(0)' : 'translateY(32px)',
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
            Join the ThunderID{' '}
            <Box
              component="span"
              sx={{
                background: 'linear-gradient(90deg, #FF6B00 0%, #FF8C00 100%)',
                WebkitBackgroundClip: 'text',
                WebkitTextFillColor: 'transparent',
                backgroundClip: 'text',
              }}
            >
              community
            </Box>
          </Typography>
          <Typography
            variant="body1"
            sx={{
              mb: 6,
              fontSize: {xs: '0.95rem', sm: '1.05rem'},
              color: isDark ? 'rgba(255, 255, 255, 0.6)' : 'rgba(0, 0, 0, 0.55)',
              lineHeight: 1.7,
              maxWidth: '600px',
            }}
          >
            We're building Thunder with you. Engage with our ever-growing community to get the latest updates, product support, and more.
          </Typography>

          <Box
            sx={{
              display: 'grid',
              gridTemplateColumns: {xs: '1fr', sm: 'repeat(2, 1fr)', md: 'repeat(3, 1fr)'},
              width: '100%',
              maxWidth: 900,
              gap: 3,
            }}
          >
            <CommunityCard
              icon={<GitForkIcon />}
              iconBg="linear-gradient(135deg, #FF6B00 0%, #FF8C00 100%)"
              title="Contribute"
              description="Help shape Thunder by submitting features, fixes, or improvements."
              linkLabel="Start Contributing"
              href="https://github.com/asgardeo/thunder/blob/main/CONTRIBUTING.md"
            />
            <CommunityCard
              icon={<IssueIcon />}
              iconBg="linear-gradient(135deg, #22c55e 0%, #16a34a 100%)"
              title="Report issues"
              description="Identify bugs and suggest enhancements to make the Thunder better for everyone."
              linkLabel="Open an Issue"
              href="https://github.com/asgardeo/thunder/issues"
            />
            <CommunityCard
              icon={<DiscordIcon />}
              iconBg="linear-gradient(135deg, #5865F2 0%, #4752C4 100%)"
              title="Join Discord"
              description="Join Discord to get real-time support, ask questions, and engage with other users"
              linkLabel="Join Discord"
              href="https://discord.gg/wso2"
            />
          </Box>
        </Box>
      </Container>
    </Box>
  );
}
