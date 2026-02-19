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
import {Box, Card, Container, Grid, Typography} from '@wso2/oxygen-ui';
import useIsDarkMode from '../../hooks/useIsDarkMode';
import useScrollAnimation from '../../hooks/useScrollAnimation';

/* ─── Shared Icons ─── */

const GoogleColoredIcon = (
  <svg width="14" height="14" viewBox="0 0 24 24">
    <path
      fill="#4285F4"
      d="M22.56 12.25c0-.78-.07-1.53-.2-2.25H12v4.26h5.92a5.06 5.06 0 0 1-2.2 3.32v2.77h3.57c2.08-1.92 3.28-4.74 3.28-8.1z"
    />
    <path
      fill="#34A853"
      d="M12 23c2.97 0 5.46-.98 7.28-2.66l-3.57-2.77c-.98.66-2.23 1.06-3.71 1.06-2.86 0-5.29-1.93-6.16-4.53H2.18v2.84C3.99 20.53 7.7 23 12 23z"
    />
    <path
      fill="#FBBC05"
      d="M5.84 14.09c-.22-.66-.35-1.36-.35-2.09s.13-1.43.35-2.09V7.07H2.18C1.43 8.55 1 10.22 1 12s.43 3.45 1.18 4.93l2.85-2.22.81-.62z"
    />
    <path
      fill="#EA4335"
      d="M12 5.38c1.62 0 3.06.56 4.21 1.64l3.15-3.15C17.45 2.09 14.97 1 12 1 7.7 1 3.99 3.47 2.18 7.07l3.66 2.84c.87-2.6 3.3-4.53 6.16-4.53z"
    />
  </svg>
);

const GithubSmallIcon = (
  <svg width="14" height="14" viewBox="0 0 24 24" fill="currentColor">
    <path d="M12 0C5.37 0 0 5.37 0 12c0 5.31 3.435 9.795 8.205 11.385.6.105.825-.255.825-.57 0-.285-.015-1.23-.015-2.235-3.015.555-3.795-.735-4.035-1.41-.135-.345-.72-1.41-1.23-1.695-.42-.225-1.02-.78-.015-.795.945-.015 1.62.87 1.845 1.23 1.08 1.815 2.805 1.305 3.495.99.105-.78.42-1.305.765-1.605-2.67-.3-5.46-1.335-5.46-5.925 0-1.305.465-2.385 1.23-3.225-.12-.3-.54-1.53.12-3.18 0 0 1.005-.315 3.3 1.23.96-.27 1.98-.405 3-.405s2.04.135 3 .405c2.295-1.56 3.3-1.23 3.3-1.23.66 1.65.24 2.88.12 3.18.765.84 1.23 1.905 1.23 3.225 0 4.605-2.805 5.625-5.475 5.925.435.375.81 1.095.81 2.22 0 1.605-.015 2.895-.015 3.3 0 .315.225.69.825.57A12.02 12.02 0 0 0 24 12c0-6.63-5.37-12-12-12z" />
  </svg>
);

const MicrosoftIcon = (
  <svg width="14" height="14" viewBox="0 0 24 24">
    <rect x="1" y="1" width="10" height="10" fill="#F25022" />
    <rect x="13" y="1" width="10" height="10" fill="#7FBA00" />
    <rect x="1" y="13" width="10" height="10" fill="#00A4EF" />
    <rect x="13" y="13" width="10" height="10" fill="#FFB900" />
  </svg>
);

function AvatarIcon({size = 32, bgcolor = '#4a6cf7'}: {size?: number; bgcolor?: string}) {
  return (
    <Box
      sx={{
        width: size,
        height: size,
        borderRadius: '50%',
        bgcolor,
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        overflow: 'hidden',
        flexShrink: 0,
      }}
    >
      <svg width={size * 0.7} height={size * 0.7} viewBox="0 0 24 24" fill="none">
        <circle cx="12" cy="9" r="4" fill="rgba(255,255,255,0.85)" />
        <path d="M4 22c0-4.42 3.58-8 8-8s8 3.58 8 8" fill="rgba(255,255,255,0.85)" />
      </svg>
    </Box>
  );
}

const socialProviders: Record<string, {icon: JSX.Element; label: string}> = {
  google: {icon: GoogleColoredIcon, label: 'Continue with Google'},
  github: {icon: GithubSmallIcon, label: 'Continue with GitHub'},
  microsoft: {icon: MicrosoftIcon, label: 'Continue with Microsoft'},
};

function SocialButton({provider}: {provider: string}) {
  const isDark = useIsDarkMode();
  const config = socialProviders[provider];
  return (
    <Box
      sx={{
        height: 32,
        borderRadius: '14px',
        bgcolor: isDark ? 'rgba(255, 255, 255, 0.05)' : 'rgba(0, 0, 0, 0.03)',
        border: '1px solid',
        borderColor: isDark ? 'rgba(255, 255, 255, 0.1)' : 'rgba(0, 0, 0, 0.08)',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        gap: 0.75,
        mb: 1,
      }}
    >
      {config.icon}
      <Typography
        variant="caption"
        sx={{
          fontSize: '0.65rem',
          fontWeight: 500,
          color: isDark ? 'rgba(255, 255, 255, 0.7)' : 'rgba(0, 0, 0, 0.6)',
        }}
      >
        {config.label}
      </Typography>
    </Box>
  );
}

/* ─── Fade Card ─── */

function FadeCard({
  height = 200,
  width = 120,
  direction = 'right',
}: {
  height?: number;
  width?: number;
  direction?: 'left' | 'right';
}) {
  const isDark = useIsDarkMode();
  const gradientDir = direction === 'right' ? 'to right' : 'to left';
  return (
    <Card
      sx={{
        width,
        height,
        flexShrink: 0,
        bgcolor: isDark ? 'rgba(255, 255, 255, 0.02)' : 'rgba(0, 0, 0, 0.02)',
        border: '1px solid',
        borderColor: isDark ? 'rgba(255, 255, 255, 0.08)' : 'rgba(0, 0, 0, 0.06)',
        maskImage: `linear-gradient(${gradientDir}, rgba(0,0,0,0.6) 0%, rgba(0,0,0,0) 100%)`,
        WebkitMaskImage: `linear-gradient(${gradientDir}, rgba(0,0,0,0.6) 0%, rgba(0,0,0,0) 100%)`,
      }}
    />
  );
}

/* ─── Mockup Components ─── */

function SignInMockup() {
  const isDark = useIsDarkMode();

  return (
    <Card sx={{p: 2.5, width: '100%'}}>
      <Box sx={{textAlign: 'center', mb: 2}}>
        <Box sx={{width: 24, height: 24, mx: 'auto', mb: 1, color: '#FF6B00', fontSize: '1.1rem'}}>&#x2726;</Box>
        <Typography variant="body2" sx={{fontWeight: 600, fontSize: '0.85rem'}}>
          Sign in to ACME
        </Typography>
      </Box>
      <Typography variant="caption" sx={{fontSize: '0.7rem', display: 'block', mb: 0.5}}>
        Email
      </Typography>
      <Box
        sx={{
          height: 30,
          borderRadius: 1,
          border: '1px solid',
          borderColor: isDark ? 'rgba(255, 255, 255, 0.15)' : 'rgba(0, 0, 0, 0.12)',
          mb: 1.5,
          px: 1,
          display: 'flex',
          alignItems: 'center',
        }}
      >
        <Typography variant="caption" sx={{fontSize: '0.65rem', opacity: 0.4}}>
          Your email address
        </Typography>
      </Box>
      <Box
        sx={{
          height: 32,
          borderRadius: '14px',
          background: 'linear-gradient(135deg, #FF6B00 0%, #FF8C00 100%)',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          mb: 1.5,
        }}
      >
        <Typography variant="caption" sx={{color: '#fff', fontSize: '0.7rem', fontWeight: 600}}>
          Continue
        </Typography>
      </Box>
      <Box sx={{display: 'flex', alignItems: 'center', gap: 1, my: 1.5}}>
        <Box sx={{flex: 1, height: '1px', bgcolor: 'divider'}} />
        <Typography variant="caption" sx={{fontSize: '0.6rem', opacity: 0.5}}>
          OR
        </Typography>
        <Box sx={{flex: 1, height: '1px', bgcolor: 'divider'}} />
      </Box>
      <SocialButton provider="google" />
      <SocialButton provider="github" />
    </Card>
  );
}

function UserProfileMockup() {
  return (
    <Card sx={{p: 2, width: '100%'}}>
      <Box sx={{display: 'flex', alignItems: 'center', gap: 1.5, mb: 1}}>
        <AvatarIcon size={32} bgcolor="#4a6cf7" />
        <Box>
          <Typography variant="body2" sx={{fontWeight: 600, fontSize: '0.8rem', lineHeight: 1.2}}>
            Mathew Asgardi
          </Typography>
          <Typography variant="caption" sx={{fontSize: '0.65rem', opacity: 0.6}}>
            mathew@thunder.dev
          </Typography>
        </Box>
      </Box>
      <Box sx={{borderTop: '1px solid', borderColor: 'divider', pt: 1, mt: 1}}>
        {['Manage Profile', 'Sign out'].map((item) => (
          <Box
            key={item}
            sx={{
              py: 0.8,
              px: 0.5,
              display: 'flex',
              alignItems: 'center',
              gap: 1,
              fontSize: '0.75rem',
              opacity: 0.7,
            }}
          >
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5">
              {item === 'Manage Profile' ? (
                <>
                  <circle cx="12" cy="8" r="4" />
                  <path d="M5 20c0-3.87 3.13-7 7-7s7 3.13 7 7" />
                </>
              ) : (
                <>
                  <path d="M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4" />
                  <polyline points="16 17 21 12 16 7" />
                  <line x1="21" y1="12" x2="9" y2="12" />
                </>
              )}
            </svg>
            {item}
          </Box>
        ))}
      </Box>
    </Card>
  );
}

function SignUpMockup() {
  const isDark = useIsDarkMode();

  return (
    <Card sx={{p: 2.5, width: '100%'}}>
      <Box sx={{textAlign: 'center', mb: 2}}>
        <Box sx={{width: 24, height: 24, mx: 'auto', mb: 1, color: '#FF6B00', fontSize: '1.1rem'}}>&#x2726;</Box>
        <Typography variant="body2" sx={{fontWeight: 600, fontSize: '0.85rem'}}>
          Sign Up to ACME
        </Typography>
      </Box>
      {['Email', 'Password'].map((label) => (
        <Box key={label} sx={{mb: 1.5}}>
          <Typography variant="caption" sx={{fontSize: '0.7rem', display: 'block', mb: 0.5}}>
            {label}
          </Typography>
          <Box
            sx={{
              height: 30,
              borderRadius: 1,
              border: '1px solid',
              borderColor: isDark ? 'rgba(255, 255, 255, 0.15)' : 'rgba(0, 0, 0, 0.12)',
              px: 1,
              display: 'flex',
              alignItems: 'center',
            }}
          >
            <Typography variant="caption" sx={{fontSize: '0.65rem', opacity: 0.4}}>
              Your {label.toLowerCase()}
            </Typography>
          </Box>
        </Box>
      ))}
      <Box
        sx={{
          height: 32,
          borderRadius: '14px',
          background: 'linear-gradient(135deg, #FF6B00 0%, #FF8C00 100%)',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          mb: 1.5,
        }}
      >
        <Typography variant="caption" sx={{color: '#fff', fontSize: '0.7rem', fontWeight: 600}}>
          Continue
        </Typography>
      </Box>
      <Box sx={{display: 'flex', alignItems: 'center', gap: 1, my: 1.5}}>
        <Box sx={{flex: 1, height: '1px', bgcolor: 'divider'}} />
        <Typography variant="caption" sx={{fontSize: '0.6rem', opacity: 0.5}}>
          OR
        </Typography>
        <Box sx={{flex: 1, height: '1px', bgcolor: 'divider'}} />
      </Box>
      <SocialButton provider="google" />
      <SocialButton provider="github" />
      <SocialButton provider="microsoft" />
    </Card>
  );
}

function UserProfileTableMockup() {
  return (
    <Card sx={{p: 2, width: '100%'}}>
      <Box sx={{display: 'flex', alignItems: 'center', gap: 1, mb: 2}}>
        <AvatarIcon size={28} bgcolor="#06b6d4" />
      </Box>
      {[
        {label: 'Name', value: 'Mathew Asgardi'},
        {label: 'Email Address', value: 'mathew@thunder.dev'},
        {label: 'Country', value: 'United States'},
        {label: 'Phone Number', value: '+1 000 000 000'},
        {label: 'Phone Number', value: '+1 000 000 000'},
      ].map((row, idx) => (
        <Box
          key={`${row.label}-${idx}`}
          sx={{
            display: 'flex',
            justifyContent: 'space-between',
            alignItems: 'center',
            py: 1,
            borderBottom: '1px solid',
            borderColor: 'divider',
          }}
        >
          <Typography variant="caption" sx={{fontSize: '0.7rem', opacity: 0.6}}>
            {row.label}
          </Typography>
          <Box sx={{display: 'flex', alignItems: 'center', gap: 1}}>
            <Typography variant="caption" sx={{fontSize: '0.7rem'}}>
              {row.value}
            </Typography>
            <svg
              width="10"
              height="10"
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              strokeWidth="1.5"
              opacity={0.4}
            >
              <path d="M17 3a2.83 2.83 0 1 1 4 4L7.5 20.5 2 22l1.5-5.5L17 3z" />
            </svg>
          </Box>
        </Box>
      ))}
    </Card>
  );
}

function UserDropdownMockup() {
  return (
    <Card
      sx={{
        p: 1.5,
        display: 'flex',
        alignItems: 'center',
        gap: 1,
        width: 'fit-content',
      }}
    >
      <AvatarIcon size={28} bgcolor="#8b5cf6" />
      <Typography variant="body2" sx={{fontSize: '0.8rem', fontWeight: 500}}>
        Mathew Asgardi
      </Typography>
      <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
        <polyline points="6 9 12 15 18 9" />
      </svg>
    </Card>
  );
}

/* ─── Main Section ─── */

export default function UIComponentsSection(): JSX.Element {
  const isDark = useIsDarkMode();
  const {ref: textRef, isVisible: textVisible} = useScrollAnimation({threshold: 0.2});
  const {ref: cardsRef, isVisible: cardsVisible} = useScrollAnimation({threshold: 0.1});

  return (
    <Box sx={{py: {xs: 8, lg: 12}, overflow: 'hidden', background: isDark ? '#0a0a0a' : 'transparent'}}>
      <Container maxWidth="lg" sx={{px: {xs: 2, sm: 4}}}>
        {/* ── Main layout: Text (left) + 2-col card grid (right) ── */}
        <Box
          sx={{
            display: 'grid',
            gridTemplateColumns: {xs: '1fr', md: '1fr 1.5fr'},
            gap: {xs: 4, md: 6},
            alignItems: 'start',
          }}
        >
          {/* Left: Text — top-aligned */}
          <Box
            ref={textRef}
            sx={{
              pt: {md: 2},
              opacity: textVisible ? 1 : 0,
              transform: textVisible ? 'translateX(0)' : 'translateX(-32px)',
              transition: 'opacity 0.7s cubic-bezier(0.16, 1, 0.3, 1), transform 0.7s cubic-bezier(0.16, 1, 0.3, 1)',
            }}
          >
            <Typography
              variant="h3"
              sx={{
                mb: 3,
                fontSize: {xs: '1.75rem', sm: '2rem', md: '2.25rem'},
                fontWeight: 700,
                lineHeight: 1.2,
                color: isDark ? '#ffffff' : '#1a1a2e',
              }}
            >
              Clean, Customizable UI Components Built for Devs
            </Typography>
            <Typography
              variant="body1"
              sx={{
                fontSize: {xs: '0.95rem', sm: '1rem'},
                lineHeight: 1.8,
                color: isDark ? 'rgba(255, 255, 255, 0.6)' : 'rgba(0, 0, 0, 0.55)',
              }}
            >
              Ready-to-use UI components for{' '}
              <Box component="code" sx={{color: '#FF6B00', fontSize: '0.9em'}}>
                {'<SignIn />'}
              </Box>
              ,{' '}
              <Box component="code" sx={{color: '#FF6B00', fontSize: '0.9em'}}>
                {'<SignUp />'}
              </Box>
              ,{' '}
              <Box component="code" sx={{color: '#FF6B00', fontSize: '0.9em'}}>
                {'<UserProfile />'}
              </Box>
              ,{' '}
              <Box component="code" sx={{color: '#FF6B00', fontSize: '0.9em'}}>
                {'<UserDropdown />'}
              </Box>{' '}
              and more for full user journeys, and style with your own CSS, and with the flexibility to choose between
              redirects or in-app experiences.
            </Typography>
          </Box>

          {/* Right: Card area with fades as absolute positioned overlays */}
          <Box
            ref={cardsRef}
            sx={{
              position: 'relative',
              opacity: cardsVisible ? 1 : 0,
              transform: cardsVisible ? 'translateX(0)' : 'translateX(32px)',
              transition: 'opacity 0.7s cubic-bezier(0.16, 1, 0.3, 1) 0.15s, transform 0.7s cubic-bezier(0.16, 1, 0.3, 1) 0.15s',
            }}
          >
            {/* Card layout: 2 columns stacked */}
            <Box sx={{display: 'flex', gap: 2, alignItems: 'start'}}>
              {/* Left column: SignIn + ProfileTable */}
              <Box sx={{flex: 1.3, minWidth: 0, display: 'flex', flexDirection: 'column', gap: 2}}>
                <SignInMockup />
                <UserProfileTableMockup />
              </Box>
              {/* Right column: UserProfile + SignUp */}
              <Box sx={{flex: 1, minWidth: 0, display: 'flex', flexDirection: 'column', gap: 2}}>
                <UserProfileMockup />
                <SignUpMockup />
              </Box>
            </Box>

            {/* Right fade — absolute, right edge */}
            <Box
              sx={{
                display: {xs: 'none', md: 'block'},
                position: 'absolute',
                top: 100,
                right: -65,
              }}
            >
              <FadeCard width={50} height={400} direction="right" />
            </Box>

            {/* Left fade — absolute, left edge, starts below SignIn */}
            <Box
              sx={{
                display: {xs: 'none', md: 'block'},
                position: 'absolute',
                bottom: 60,
                left: -140,
              }}
            >
              <FadeCard height={280} direction="left" />
            </Box>
          </Box>
        </Box>

        {/* ── Bottom: dropdown row ── */}
        <Box sx={{display: 'flex', gap: 2, mt: 2, justifyContent: 'center', alignItems: 'start'}}>
          <Box sx={{display: {xs: 'none', md: 'block'}, flexShrink: 0}}>
            <FadeCard height={54} direction="left" />
          </Box>
          <UserDropdownMockup />
          <Box sx={{display: {xs: 'none', md: 'block'}, flexShrink: 0}}>
            <FadeCard height={54} direction="right" />
          </Box>
        </Box>
      </Container>
    </Box>
  );
}
