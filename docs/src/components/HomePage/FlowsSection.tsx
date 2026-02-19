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

import React, {JSX, useCallback, useRef, useState} from 'react';
import {Box, Card, Container, Typography} from '@wso2/oxygen-ui';
import useIsDarkMode from '../../hooks/useIsDarkMode';
import useScrollAnimation from '../../hooks/useScrollAnimation';

/* ─── Types ─── */

type CategoryId = 'social' | 'passwordless' | 'mfa';
type LogoType = 'star' | 'plus' | 'bars';

const THEME_COLORS = ['#FF6B00', '#3b82f6', '#10b981', '#8b5cf6'];

interface ThemeBranding {
  logoType: LogoType;
  primaryColor: string;
  borderRadiusRounded: boolean;
}

interface FlowConfig {
  signinTitle: string;
  showPassword?: boolean;
  socialProvider?: 'google' | 'github' | 'facebook';
  verifyType: 'otp' | 'passkey' | 'biometric' | 'magic-link' | 'security-key' | 'email-otp' | 'totp';
  verifyTitle: string;
  otpLength?: number;
  caption: string;
}

interface IconDef {
  id: string;
  label: string;
  icon: JSX.Element;
}

interface CategoryDef {
  id: CategoryId;
  label: string;
  icons: IconDef[];
  flows: FlowConfig[];
}

/* ─── SVG Icon Helpers ─── */

function AsgardeoStarIcon({size = 24, color = '#FF6B00'}: {size?: number; color?: string}) {
  return (
    <svg width={size} height={size} viewBox="0 0 24 24" fill={color}>
      <path d="M12 2l2.4 7.2H22l-6 4.8 2.4 7.2L12 16.4l-6.4 4.8 2.4-7.2-6-4.8h7.6z" />
    </svg>
  );
}

const FacebookIcon = (
  <svg width="20" height="20" viewBox="0 0 24 24" fill="currentColor">
    <path d="M18 2h-3a5 5 0 0 0-5 5v3H7v4h3v8h4v-8h3l1-4h-4V7a1 1 0 0 1 1-1h3z" />
  </svg>
);

const GithubIcon = (
  <svg width="20" height="20" viewBox="0 0 24 24" fill="currentColor">
    <path d="M12 0C5.37 0 0 5.37 0 12c0 5.31 3.435 9.795 8.205 11.385.6.105.825-.255.825-.57 0-.285-.015-1.23-.015-2.235-3.015.555-3.795-.735-4.035-1.41-.135-.345-.72-1.41-1.23-1.695-.42-.225-1.02-.78-.015-.795.945-.015 1.62.87 1.845 1.23 1.08 1.815 2.805 1.305 3.495.99.105-.78.42-1.305.765-1.605-2.67-.3-5.46-1.335-5.46-5.925 0-1.305.465-2.385 1.23-3.225-.12-.3-.54-1.53.12-3.18 0 0 1.005-.315 3.3 1.23.96-.27 1.98-.405 3-.405s2.04.135 3 .405c2.295-1.56 3.3-1.23 3.3-1.23.66 1.65.24 2.88.12 3.18.765.84 1.23 1.905 1.23 3.225 0 4.605-2.805 5.625-5.475 5.925.435.375.81 1.095.81 2.22 0 1.605-.015 2.895-.015 3.3 0 .315.225.69.825.57A12.02 12.02 0 0 0 24 12c0-6.63-5.37-12-12-12z" />
  </svg>
);

const GoogleLetterIcon = <Box component="span" sx={{fontWeight: 700, fontSize: '1.1rem', color: 'inherit'}}>G</Box>;

const PersonIcon = (
  <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
    <path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2" />
    <circle cx="12" cy="7" r="4" />
  </svg>
);

const FingerprintIcon = (
  <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
    <path d="M2 12C2 6.5 6.5 2 12 2a10 10 0 0 1 8 4" />
    <path d="M5 19.5C5.5 18 6 15 6 12c0-3.5 2.5-6 6-6" />
    <path d="M12 10c-2 0-3 1.5-3 3 0 3-1 5.5-2 7.5" />
    <path d="M15 9.4a4 4 0 0 1 1 2.6c0 3-.5 5.5-1.5 8" />
    <path d="M18 6.5a9 9 0 0 1 2 5.5c0 4.5-1 7-2 9" />
  </svg>
);

const MagicLinkIcon = (
  <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
    <path d="M12 20h.01" />
    <path d="M2 8.82a15 15 0 0 1 20 0" />
    <path d="M5 12.859a10 10 0 0 1 14 0" />
    <path d="M8.5 16.429a5 5 0 0 1 7 0" />
  </svg>
);

const PhoneIcon = (
  <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
    <rect width="14" height="20" x="5" y="2" rx="2" ry="2" />
    <path d="M12 18h.01" />
  </svg>
);

const LockIcon = (
  <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
    <rect width="18" height="11" x="3" y="11" rx="2" ry="2" />
    <path d="M7 11V7a5 5 0 0 1 10 0v4" />
  </svg>
);

const EmailIcon = (
  <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
    <rect width="20" height="16" x="2" y="4" rx="2" />
    <path d="m22 7-8.97 5.7a1.94 1.94 0 0 1-2.06 0L2 7" />
  </svg>
);

const KeypadIcon = (
  <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
    <rect width="20" height="16" x="2" y="4" rx="2" />
    <path d="M6 8h.01M10 8h.01M14 8h.01M18 8h.01M8 12h.01M12 12h.01M16 12h.01M6 16h.01M10 16h.01M14 16h.01M18 16h.01" />
  </svg>
);

/* ─── Flow Configurations ─── */

const CATEGORIES: CategoryDef[] = [
  {
    id: 'social',
    label: 'Social Logins',
    icons: [
      {id: 'facebook', label: 'Facebook', icon: FacebookIcon},
      {id: 'github', label: 'GitHub', icon: GithubIcon},
      {id: 'google', label: 'Google', icon: GoogleLetterIcon},
    ],
    flows: [
      {
        signinTitle: 'Sign in to Thunder',
        socialProvider: 'facebook',
        verifyType: 'otp',
        verifyTitle: 'Verify your mobile',
        otpLength: 5,
        caption: 'Social Login with Facebook',
      },
      {
        signinTitle: 'Sign in to Thunder',
        socialProvider: 'github',
        verifyType: 'otp',
        verifyTitle: 'Verify your mobile',
        otpLength: 5,
        caption: 'Social Login with GitHub',
      },
      {
        signinTitle: 'Sign in to Thunder',
        socialProvider: 'google',
        verifyType: 'otp',
        verifyTitle: 'Verify your mobile',
        otpLength: 5,
        caption: 'Social Login',
      },
    ],
  },
  {
    id: 'passwordless',
    label: 'Passwordless Login',
    icons: [
      {id: 'passkey', label: 'Passkey', icon: PersonIcon},
      {id: 'biometric', label: 'Biometric', icon: FingerprintIcon},
      {id: 'magic-link', label: 'Magic Link', icon: MagicLinkIcon},
    ],
    flows: [
      {
        signinTitle: 'Sign in to Thunder',
        verifyType: 'passkey',
        verifyTitle: 'Authenticate with Passkey',
        caption: 'Passkey Authentication',
      },
      {
        signinTitle: 'Sign in to Thunder',
        verifyType: 'biometric',
        verifyTitle: 'Biometric Verification',
        caption: 'Biometric Login',
      },
      {
        signinTitle: 'Sign in to Thunder',
        verifyType: 'magic-link',
        verifyTitle: 'Check your email',
        caption: 'Magic Link Login',
      },
    ],
  },
  {
    id: 'mfa',
    label: 'Multi-Factor Authentication',
    icons: [
      {id: 'sms', label: 'SMS OTP', icon: PhoneIcon},
      {id: 'security-key', label: 'Security Key', icon: LockIcon},
      {id: 'email-otp', label: 'Email OTP', icon: EmailIcon},
      {id: 'totp', label: 'TOTP', icon: KeypadIcon},
    ],
    flows: [
      {
        signinTitle: 'Sign in to Thunder',
        showPassword: true,
        verifyType: 'otp',
        verifyTitle: 'Verify your mobile',
        otpLength: 5,
        caption: 'MFA with SMS OTP',
      },
      {
        signinTitle: 'Sign in to Thunder',
        showPassword: true,
        verifyType: 'security-key',
        verifyTitle: 'Security Key Verification',
        caption: 'MFA with Security Key',
      },
      {
        signinTitle: 'Sign in to Thunder',
        showPassword: true,
        verifyType: 'email-otp',
        verifyTitle: 'Verify your email',
        otpLength: 6,
        caption: 'MFA with Email OTP',
      },
      {
        signinTitle: 'Sign in to Thunder',
        showPassword: true,
        verifyType: 'totp',
        verifyTitle: 'Enter authenticator code',
        otpLength: 6,
        caption: 'MFA with Authenticator App',
      },
    ],
  },
];

/* ─── Sub-components ─── */

function LogoIcon({type, color, size = 28}: {type: LogoType; color: string; size?: number}): JSX.Element {
  if (type === 'star') return <AsgardeoStarIcon size={size} color={color} />;
  if (type === 'plus') {
    return (
      <svg width={size} height={size} viewBox="0 0 24 24" fill="none" stroke={color} strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
        <circle cx="12" cy="12" r="10" />
        <path d="M8 12h8M12 8v8" />
      </svg>
    );
  }
  return (
    <svg width={size} height={size} viewBox="0 0 24 24" fill="none" stroke={color} strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
      <path d="M12 4v4M12 16v4M8 8h8M6 12h12M8 16h8" />
    </svg>
  );
}

function BrandingConfigPanel({
  branding,
  onBrandingChange,
}: {
  branding: ThemeBranding;
  onBrandingChange: (update: Partial<ThemeBranding>) => void;
}) {
  const isDark = useIsDarkMode();
  const allLogoTypes: LogoType[] = ['star', 'plus', 'bars'];

  return (
    <Card
      sx={{
        p: 3,
        display: 'flex',
        gap: 4,
        flexWrap: 'wrap',
        border: '1px dashed',
        borderColor: isDark ? 'rgba(255, 255, 255, 0.12)' : 'rgba(0, 0, 0, 0.1)',
        bgcolor: isDark ? 'rgba(255, 255, 255, 0.02)' : 'rgba(0, 0, 0, 0.02)',
      }}
    >
      <Box>
        <Typography variant="body2" sx={{fontWeight: 600, mb: 1.5, fontSize: '0.8rem'}}>
          Logo
        </Typography>
        <Box sx={{display: 'flex', gap: 1}}>
          {allLogoTypes.map((type) => {
            const isActive = type === branding.logoType;
            return (
              <Box
                key={type}
                onClick={() => onBrandingChange({logoType: type})}
                sx={{
                  width: 52, height: 52, borderRadius: 1.5,
                  bgcolor: isActive ? `${branding.primaryColor}22` : (isDark ? 'rgba(255, 255, 255, 0.06)' : 'rgba(0, 0, 0, 0.04)'),
                  border: '1px dashed',
                  borderColor: isActive ? `${branding.primaryColor}88` : (isDark ? 'rgba(255, 255, 255, 0.1)' : 'rgba(0, 0, 0, 0.1)'),
                  display: 'flex', alignItems: 'center', justifyContent: 'center',
                  cursor: 'pointer',
                  transition: 'all 0.25s ease',
                  '&:hover': {
                    borderColor: isActive ? branding.primaryColor : (isDark ? 'rgba(255, 255, 255, 0.3)' : 'rgba(0, 0, 0, 0.2)'),
                    bgcolor: isActive ? `${branding.primaryColor}30` : (isDark ? 'rgba(255, 255, 255, 0.1)' : 'rgba(0, 0, 0, 0.06)'),
                  },
                }}
              >
                <LogoIcon type={type} color={isActive ? branding.primaryColor : (isDark ? 'rgba(255, 255, 255, 0.4)' : 'rgba(0, 0, 0, 0.3)')} />
              </Box>
            );
          })}
        </Box>
      </Box>
      <Box>
        <Typography variant="body2" sx={{fontWeight: 600, mb: 1.5, fontSize: '0.8rem'}}>
          Colors
        </Typography>
        <Box sx={{display: 'flex', gap: 0.8}}>
          {THEME_COLORS.map((color) => {
            const isActive = color === branding.primaryColor;
            return (
              <Box
                key={color}
                onClick={() => onBrandingChange({primaryColor: color})}
                sx={{
                  width: 32, height: 32, borderRadius: '50%', bgcolor: color,
                  cursor: 'pointer',
                  transition: 'all 0.25s ease',
                  border: isActive ? '2.5px solid rgba(255,255,255,0.5)' : '2.5px solid transparent',
                  transform: isActive ? 'scale(1.15)' : 'scale(1)',
                  '&:hover': {
                    transform: 'scale(1.15)',
                    border: '2.5px solid rgba(255,255,255,0.3)',
                  },
                }}
              />
            );
          })}
        </Box>
      </Box>
      <Box>
        <Typography variant="body2" sx={{fontWeight: 600, mb: 1.5, fontSize: '0.8rem'}}>
          Border Radius
        </Typography>
        <Box sx={{display: 'flex', gap: 1, alignItems: 'center'}}>
          <Box
            onClick={() => onBrandingChange({borderRadiusRounded: true})}
            sx={{
              width: 60, height: 28,
              borderRadius: 14,
              bgcolor: branding.borderRadiusRounded ? (isDark ? 'rgba(255, 255, 255, 0.15)' : 'rgba(0, 0, 0, 0.08)') : (isDark ? 'rgba(255, 255, 255, 0.08)' : 'rgba(0, 0, 0, 0.04)'),
              border: branding.borderRadiusRounded ? `1.5px solid ${branding.primaryColor}` : '1.5px solid transparent',
              cursor: 'pointer',
              transition: 'all 0.25s ease',
              '&:hover': {borderColor: branding.borderRadiusRounded ? branding.primaryColor : (isDark ? 'rgba(255, 255, 255, 0.3)' : 'rgba(0, 0, 0, 0.2)')},
            }}
          />
          <Box
            onClick={() => onBrandingChange({borderRadiusRounded: false})}
            sx={{
              width: 60, height: 28,
              borderRadius: 1,
              bgcolor: !branding.borderRadiusRounded ? (isDark ? 'rgba(255, 255, 255, 0.15)' : 'rgba(0, 0, 0, 0.08)') : (isDark ? 'rgba(255, 255, 255, 0.08)' : 'rgba(0, 0, 0, 0.04)'),
              border: !branding.borderRadiusRounded ? `1.5px solid ${branding.primaryColor}` : '1.5px solid transparent',
              cursor: 'pointer',
              transition: 'all 0.25s ease',
              '&:hover': {borderColor: !branding.borderRadiusRounded ? branding.primaryColor : (isDark ? 'rgba(255, 255, 255, 0.3)' : 'rgba(0, 0, 0, 0.2)')},
            }}
          />
        </Box>
      </Box>
    </Card>
  );
}

const iconBoxBase = {
  width: 44, height: 44, borderRadius: 1.5,
  display: 'flex', alignItems: 'center', justifyContent: 'center',
  cursor: 'pointer',
  transition: 'all 0.2s ease',
};

function AuthOptionsCard({
  selectedCategory,
  selectedIcon,
  onSelect,
}: {
  selectedCategory: CategoryId;
  selectedIcon: number;
  onSelect: (cat: CategoryId, icon: number) => void;
}) {
  const isDark = useIsDarkMode();
  return (
    <Card sx={{p: 3, border: '1px dashed', borderColor: isDark ? 'rgba(255, 255, 255, 0.12)' : 'rgba(0, 0, 0, 0.1)', bgcolor: isDark ? 'rgba(255, 255, 255, 0.02)' : 'rgba(0, 0, 0, 0.02)'}}>
      {CATEGORIES.map((cat, catIdx) => (
        <Box key={cat.id} sx={{mb: catIdx < CATEGORIES.length - 1 ? 2.5 : 0}}>
          <Typography
            variant="body2"
            onClick={() => onSelect(cat.id, 0)}
            sx={{
              fontWeight: 600,
              mb: 1.5,
              fontSize: '0.8rem',
              fontFamily: 'var(--ifm-font-family-monospace)',
              cursor: 'pointer',
              color: selectedCategory === cat.id ? '#FF6B00' : 'inherit',
              transition: 'color 0.2s ease',
              '&:hover': {color: '#FF6B00'},
            }}
          >
            {cat.label}
          </Typography>
          <Box sx={{display: 'flex', gap: 1}}>
            {cat.icons.map((iconDef, iconIdx) => {
              const isSelected = selectedCategory === cat.id && selectedIcon === iconIdx;
              return (
                <Box
                  key={iconDef.id}
                  onClick={() => onSelect(cat.id, iconIdx)}
                  sx={{
                    ...iconBoxBase,
                    color: isSelected ? '#FF6B00' : (isDark ? 'rgba(255, 255, 255, 0.6)' : 'rgba(0, 0, 0, 0.45)'),
                    bgcolor: isSelected ? 'rgba(255, 107, 0, 0.15)' : (isDark ? 'rgba(255, 255, 255, 0.06)' : 'rgba(0, 0, 0, 0.04)'),
                    border: '1px solid',
                    borderColor: isSelected ? 'rgba(255, 107, 0, 0.3)' : (isDark ? 'rgba(255, 255, 255, 0.1)' : 'rgba(0, 0, 0, 0.1)'),
                    '&:hover': {
                      borderColor: isSelected ? 'rgba(255, 107, 0, 0.5)' : (isDark ? 'rgba(255, 255, 255, 0.3)' : 'rgba(0, 0, 0, 0.2)'),
                      bgcolor: isSelected ? 'rgba(255, 107, 0, 0.2)' : (isDark ? 'rgba(255, 255, 255, 0.1)' : 'rgba(0, 0, 0, 0.06)'),
                    },
                  }}
                >
                  {iconDef.icon}
                </Box>
              );
            })}
          </Box>
        </Box>
      ))}
    </Card>
  );
}

const GoogleColoredIcon = (
  <svg width="14" height="14" viewBox="0 0 24 24">
    <path fill="#4285F4" d="M22.56 12.25c0-.78-.07-1.53-.2-2.25H12v4.26h5.92a5.06 5.06 0 0 1-2.2 3.32v2.77h3.57c2.08-1.92 3.28-4.74 3.28-8.1z" />
    <path fill="#34A853" d="M12 23c2.97 0 5.46-.98 7.28-2.66l-3.57-2.77c-.98.66-2.23 1.06-3.71 1.06-2.86 0-5.29-1.93-6.16-4.53H2.18v2.84C3.99 20.53 7.7 23 12 23z" />
    <path fill="#FBBC05" d="M5.84 14.09c-.22-.66-.35-1.36-.35-2.09s.13-1.43.35-2.09V7.07H2.18C1.43 8.55 1 10.22 1 12s.43 3.45 1.18 4.93l2.85-2.22.81-.62z" />
    <path fill="#EA4335" d="M12 5.38c1.62 0 3.06.56 4.21 1.64l3.15-3.15C17.45 2.09 14.97 1 12 1 7.7 1 3.99 3.47 2.18 7.07l3.66 2.84c.87-2.6 3.3-4.53 6.16-4.53z" />
  </svg>
);

const FacebookColoredIcon = (
  <svg width="14" height="14" viewBox="0 0 24 24" fill="#1877F2">
    <path d="M18 2h-3a5 5 0 0 0-5 5v3H7v4h3v8h4v-8h3l1-4h-4V7a1 1 0 0 1 1-1h3z" />
  </svg>
);

const GithubSmallIcon = (
  <svg width="14" height="14" viewBox="0 0 24 24" fill="currentColor">
    <path d="M12 0C5.37 0 0 5.37 0 12c0 5.31 3.435 9.795 8.205 11.385.6.105.825-.255.825-.57 0-.285-.015-1.23-.015-2.235-3.015.555-3.795-.735-4.035-1.41-.135-.345-.72-1.41-1.23-1.695-.42-.225-1.02-.78-.015-.795.945-.015 1.62.87 1.845 1.23 1.08 1.815 2.805 1.305 3.495.99.105-.78.42-1.305.765-1.605-2.67-.3-5.46-1.335-5.46-5.925 0-1.305.465-2.385 1.23-3.225-.12-.3-.54-1.53.12-3.18 0 0 1.005-.315 3.3 1.23.96-.27 1.98-.405 3-.405s2.04.135 3 .405c2.295-1.56 3.3-1.23 3.3-1.23.66 1.65.24 2.88.12 3.18.765.84 1.23 1.905 1.23 3.225 0 4.605-2.805 5.625-5.475 5.925.435.375.81 1.095.81 2.22 0 1.605-.015 2.895-.015 3.3 0 .315.225.69.825.57A12.02 12.02 0 0 0 24 12c0-6.63-5.37-12-12-12z" />
  </svg>
);

const socialButtonConfig: Record<string, {icon: JSX.Element; label: string}> = {
  google: {icon: GoogleColoredIcon, label: 'Continue with Google'},
  facebook: {icon: FacebookColoredIcon, label: 'Continue with Facebook'},
  github: {icon: GithubSmallIcon, label: 'Continue with GitHub'},
};

function SigninCard({flow, branding}: {flow: FlowConfig; branding: ThemeBranding}) {
  const isDark = useIsDarkMode();
  const btnRadius = branding.borderRadiusRounded ? '14px' : '2px';
  const inputRadius = branding.borderRadiusRounded ? '8px' : '2px';

  return (
    <Card sx={{p: 2.5, width: {xs: '100%', sm: 220}, flexShrink: 0, border: '1px dashed', borderColor: isDark ? 'rgba(255, 255, 255, 0.12)' : 'rgba(0, 0, 0, 0.1)', bgcolor: isDark ? 'rgba(255, 255, 255, 0.02)' : 'rgba(0, 0, 0, 0.02)'}}>
      <Box sx={{textAlign: 'center', mb: 1.5}}>
        <Box sx={{width: 28, height: 28, mx: 'auto', mb: 1}}>
          <LogoIcon type={branding.logoType} color={branding.primaryColor} />
        </Box>
        <Typography variant="body2" sx={{fontWeight: 600, fontSize: '0.85rem'}}>
          {flow.signinTitle}
        </Typography>
      </Box>
      <Typography variant="caption" sx={{fontSize: '0.7rem', display: 'block', mb: 0.5}}>
        Email
      </Typography>
      <Box
        sx={{
          height: 30, borderRadius: inputRadius,
          border: '1px solid', borderColor: isDark ? 'rgba(255, 255, 255, 0.15)' : 'rgba(0, 0, 0, 0.12)',
          bgcolor: isDark ? 'rgba(255, 255, 255, 0.05)' : 'rgba(0, 0, 0, 0.03)', mb: flow.showPassword ? 1 : 1.5,
          transition: 'border-radius 0.25s ease',
        }}
      />
      {flow.showPassword && (
        <>
          <Typography variant="caption" sx={{fontSize: '0.7rem', display: 'block', mb: 0.5}}>
            Password
          </Typography>
          <Box
            sx={{
              height: 30, borderRadius: inputRadius,
              border: '1px solid', borderColor: isDark ? 'rgba(255, 255, 255, 0.15)' : 'rgba(0, 0, 0, 0.12)',
              bgcolor: isDark ? 'rgba(255, 255, 255, 0.05)' : 'rgba(0, 0, 0, 0.03)', mb: 1.5,
              transition: 'border-radius 0.25s ease',
            }}
          />
        </>
      )}
      <Box
        sx={{
          height: 32, borderRadius: btnRadius,
          background: `linear-gradient(135deg, ${branding.primaryColor} 0%, ${branding.primaryColor}cc 100%)`,
          display: 'flex', alignItems: 'center', justifyContent: 'center',
          mb: flow.socialProvider ? 1.5 : 0,
          transition: 'background 0.25s ease, border-radius 0.25s ease',
        }}
      >
        <Typography variant="caption" sx={{color: '#fff', fontSize: '0.7rem', fontWeight: 600}}>
          {flow.showPassword ? 'Sign in' : 'Continue'}
        </Typography>
      </Box>
      {flow.socialProvider && (
        <>
          <Box sx={{display: 'flex', alignItems: 'center', gap: 1, my: 1}}>
            <Box sx={{flex: 1, height: '1px', bgcolor: isDark ? 'rgba(255, 255, 255, 0.1)' : 'rgba(0, 0, 0, 0.08)'}} />
            <Typography variant="caption" sx={{fontSize: '0.65rem', color: isDark ? 'rgba(255, 255, 255, 0.4)' : 'rgba(0, 0, 0, 0.35)', textTransform: 'uppercase'}}>
              or
            </Typography>
            <Box sx={{flex: 1, height: '1px', bgcolor: isDark ? 'rgba(255, 255, 255, 0.1)' : 'rgba(0, 0, 0, 0.08)'}} />
          </Box>
          <Box
            sx={{
              height: 32, borderRadius: btnRadius,
              bgcolor: isDark ? 'rgba(255, 255, 255, 0.08)' : 'rgba(0, 0, 0, 0.04)',
              border: '1px solid', borderColor: isDark ? 'rgba(255, 255, 255, 0.1)' : 'rgba(0, 0, 0, 0.1)',
              display: 'flex', alignItems: 'center', justifyContent: 'center', gap: 0.5,
              transition: 'border-radius 0.25s ease',
            }}
          >
            {socialButtonConfig[flow.socialProvider].icon}
            <Typography variant="caption" sx={{fontSize: '0.65rem', fontWeight: 500, color: isDark ? 'rgba(255, 255, 255, 0.7)' : 'rgba(0, 0, 0, 0.6)'}}>
              {socialButtonConfig[flow.socialProvider].label}
            </Typography>
          </Box>
        </>
      )}
    </Card>
  );
}

function VerifyCard({flow, branding}: {flow: FlowConfig; branding: ThemeBranding}) {
  const isDark = useIsDarkMode();
  const btnRadius = branding.borderRadiusRounded ? '14px' : '2px';
  const inputRadius = branding.borderRadiusRounded ? '8px' : '2px';

  const primaryButton = (label: string) => (
    <Box
      sx={{
        height: 32, borderRadius: btnRadius,
        background: `linear-gradient(135deg, ${branding.primaryColor} 0%, ${branding.primaryColor}cc 100%)`,
        display: 'flex', alignItems: 'center', justifyContent: 'center',
        transition: 'background 0.25s ease, border-radius 0.25s ease',
      }}
    >
      <Typography variant="caption" sx={{color: '#fff', fontSize: '0.7rem', fontWeight: 600}}>
        {label}
      </Typography>
    </Box>
  );

  const secondaryButton = (label: string) => (
    <Box
      sx={{
        height: 32, borderRadius: btnRadius,
        bgcolor: isDark ? 'rgba(255, 255, 255, 0.08)' : 'rgba(0, 0, 0, 0.04)',
        border: '1px solid', borderColor: isDark ? 'rgba(255, 255, 255, 0.1)' : 'rgba(0, 0, 0, 0.1)',
        display: 'flex', alignItems: 'center', justifyContent: 'center',
        transition: 'border-radius 0.25s ease',
      }}
    >
      <Typography variant="caption" sx={{fontSize: '0.7rem', fontWeight: 600, color: isDark ? 'rgba(255, 255, 255, 0.5)' : 'rgba(0, 0, 0, 0.4)'}}>
        {label}
      </Typography>
    </Box>
  );

  const otpBoxes = (length: number) => (
    <Box sx={{display: 'flex', gap: 0.5, mb: 1.5, justifyContent: 'center'}}>
      {Array.from({length}).map((_, i) => (
        <Box
          key={i}
          sx={{
            width: 30, height: 34, borderRadius: inputRadius,
            bgcolor: isDark ? 'rgba(255, 255, 255, 0.06)' : 'rgba(0, 0, 0, 0.03)',
            border: '1px solid', borderColor: isDark ? 'rgba(255, 255, 255, 0.12)' : 'rgba(0, 0, 0, 0.1)',
            transition: 'border-radius 0.25s ease',
          }}
        />
      ))}
    </Box>
  );

  return (
    <Card sx={{p: 2.5, width: {xs: '100%', sm: 220}, flexShrink: 0, border: '1px dashed', borderColor: isDark ? 'rgba(255, 255, 255, 0.12)' : 'rgba(0, 0, 0, 0.1)', bgcolor: isDark ? 'rgba(255, 255, 255, 0.02)' : 'rgba(0, 0, 0, 0.02)'}}>
      <Box sx={{textAlign: 'center', mb: 1.5}}>
        <Box sx={{width: 28, height: 28, mx: 'auto', mb: 1}}>
          <LogoIcon type={branding.logoType} color={branding.primaryColor} />
        </Box>
        <Typography variant="body2" sx={{fontWeight: 600, fontSize: '0.85rem'}}>
          {flow.verifyTitle}
        </Typography>
      </Box>

      {(flow.verifyType === 'otp' || flow.verifyType === 'email-otp') && (
        <>
          <Typography variant="caption" sx={{fontSize: '0.7rem', display: 'block', mb: 0.5}}>
            Enter OTP
          </Typography>
          {otpBoxes(flow.otpLength ?? 5)}
          {primaryButton('Submit')}
        </>
      )}

      {flow.verifyType === 'totp' && (
        <>
          <Typography variant="caption" sx={{fontSize: '0.65rem', display: 'block', mb: 0.5, color: isDark ? 'rgba(255, 255, 255, 0.5)' : 'rgba(0, 0, 0, 0.45)'}}>
            Enter code from your authenticator app
          </Typography>
          {otpBoxes(6)}
          {primaryButton('Verify')}
        </>
      )}

      {flow.verifyType === 'passkey' && (
        <Box sx={{textAlign: 'center'}}>
          <Box sx={{mx: 'auto', mb: 1.5, display: 'flex', justifyContent: 'center'}}>
            <svg width="32" height="32" viewBox="0 0 24 24" fill="none" stroke={branding.primaryColor} strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round">
              <path d="m21 2-2 2m-7.61 7.61a5.5 5.5 0 1 1-7.778 7.778 5.5 5.5 0 0 1 7.777-7.777zm0 0L15.5 7.5m0 0 3 3L22 7l-3-3m-3.5 3.5L19 4" />
            </svg>
          </Box>
          <Typography variant="caption" sx={{fontSize: '0.65rem', color: isDark ? 'rgba(255, 255, 255, 0.5)' : 'rgba(0, 0, 0, 0.45)', display: 'block', mb: 1.5}}>
            Use your device to verify your identity
          </Typography>
          {primaryButton('Verify')}
        </Box>
      )}

      {flow.verifyType === 'biometric' && (
        <Box sx={{textAlign: 'center'}}>
          <Box sx={{mx: 'auto', mb: 1.5, display: 'flex', justifyContent: 'center'}}>
            <svg width="32" height="32" viewBox="0 0 24 24" fill="none" stroke={branding.primaryColor} strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round">
              <path d="M2 12C2 6.5 6.5 2 12 2a10 10 0 0 1 8 4" />
              <path d="M5 19.5C5.5 18 6 15 6 12c0-3.5 2.5-6 6-6" />
              <path d="M12 10c-2 0-3 1.5-3 3 0 3-1 5.5-2 7.5" />
              <path d="M15 9.4a4 4 0 0 1 1 2.6c0 3-.5 5.5-1.5 8" />
              <path d="M18 6.5a9 9 0 0 1 2 5.5c0 4.5-1 7-2 9" />
            </svg>
          </Box>
          <Typography variant="caption" sx={{fontSize: '0.65rem', color: isDark ? 'rgba(255, 255, 255, 0.5)' : 'rgba(0, 0, 0, 0.45)', display: 'block', mb: 1.5}}>
            Place your finger on the sensor
          </Typography>
          {secondaryButton('Verifying...')}
        </Box>
      )}

      {flow.verifyType === 'magic-link' && (
        <Box sx={{textAlign: 'center'}}>
          <Box sx={{mx: 'auto', mb: 1.5, display: 'flex', justifyContent: 'center'}}>
            <svg width="32" height="32" viewBox="0 0 24 24" fill="none" stroke={branding.primaryColor} strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round">
              <rect width="20" height="16" x="2" y="4" rx="2" />
              <path d="m22 7-8.97 5.7a1.94 1.94 0 0 1-2.06 0L2 7" />
            </svg>
          </Box>
          <Typography variant="caption" sx={{fontSize: '0.65rem', color: isDark ? 'rgba(255, 255, 255, 0.5)' : 'rgba(0, 0, 0, 0.45)', display: 'block', mb: 1.5}}>
            We sent a magic link to your email
          </Typography>
          {primaryButton('Open Email App')}
        </Box>
      )}

      {flow.verifyType === 'security-key' && (
        <Box sx={{textAlign: 'center'}}>
          <Box sx={{mx: 'auto', mb: 1.5, display: 'flex', justifyContent: 'center'}}>
            <svg width="32" height="32" viewBox="0 0 24 24" fill="none" stroke={branding.primaryColor} strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round">
              <path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z" />
              <path d="m9 12 2 2 4-4" />
            </svg>
          </Box>
          <Typography variant="caption" sx={{fontSize: '0.65rem', color: isDark ? 'rgba(255, 255, 255, 0.5)' : 'rgba(0, 0, 0, 0.45)', display: 'block', mb: 1.5}}>
            Insert your security key and tap
          </Typography>
          {secondaryButton('Waiting for key...')}
        </Box>
      )}
    </Card>
  );
}

/* ─── Main Component ─── */

export default function FlowsSection(): JSX.Element {
  const isDark = useIsDarkMode();
  const [selectedCategory, setSelectedCategory] = useState<CategoryId>('social');
  const [selectedIcon, setSelectedIcon] = useState(2); // Google by default
  const [isSigninTransitioning, setIsSigninTransitioning] = useState(false);
  const [isVerifyTransitioning, setIsVerifyTransitioning] = useState(false);
  const [branding, setBranding] = useState<ThemeBranding>({
    logoType: 'star',
    primaryColor: '#FF6B00',
    borderRadiusRounded: true,
  });
  const signinTimeoutRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  const verifyTimeoutRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  const {ref: sectionRef, isVisible: sectionVisible} = useScrollAnimation({threshold: 0.1});

  const handleSelect = useCallback((cat: CategoryId, iconIdx: number) => {
    if (cat === selectedCategory && iconIdx === selectedIcon) return;

    const categoryChanging = cat !== selectedCategory;
    const signinChanges = categoryChanging || cat === 'social';

    if (signinTimeoutRef.current) clearTimeout(signinTimeoutRef.current);
    if (verifyTimeoutRef.current) clearTimeout(verifyTimeoutRef.current);

    setIsVerifyTransitioning(true);

    if (signinChanges) {
      setIsSigninTransitioning(true);
    }

    verifyTimeoutRef.current = setTimeout(() => {
      setSelectedCategory(cat);
      setSelectedIcon(iconIdx);
      setIsVerifyTransitioning(false);
      if (signinChanges) {
        setIsSigninTransitioning(false);
      }
    }, 200);
  }, [selectedCategory, selectedIcon]);

  const handleBrandingChange = useCallback((update: Partial<ThemeBranding>) => {
    setBranding((prev) => ({...prev, ...update}));
  }, []);

  const category = CATEGORIES.find((c) => c.id === selectedCategory)!;
  const flow = category.flows[selectedIcon];

  return (
    <Box sx={{py: {xs: 8, lg: 12}, background: isDark ? '#0a0a0a' : 'transparent'}}>
      <Container ref={sectionRef} maxWidth="lg" sx={{px: {xs: 2, sm: 4}}}>
        <Box
          sx={{
            textAlign: 'center',
            mb: 8,
            opacity: sectionVisible ? 1 : 0,
            transform: sectionVisible ? 'translateY(0)' : 'translateY(32px)',
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
            Beautiful, Customizable Login and Registration Flows
          </Typography>
          <Typography
            variant="body1"
            sx={{color: isDark ? 'rgba(255, 255, 255, 0.6)' : 'rgba(0, 0, 0, 0.55)', maxWidth: '600px', mx: 'auto', fontSize: {xs: '0.95rem', sm: '1.05rem'}}}
          >
            Streamlined, secure, and customizable visual user flows.
          </Typography>
        </Box>

        {/* Branding Config */}
        <Box sx={{mb: 6}}>
          <BrandingConfigPanel branding={branding} onBrandingChange={handleBrandingChange} />
        </Box>

        {/* Flow Diagram: Auth options + Login cards */}
        <Box
          sx={{
            display: 'grid',
            gridTemplateColumns: {xs: '1fr', md: '280px 1fr'},
            gap: 4,
            alignItems: 'start',
          }}
        >
          <AuthOptionsCard
            selectedCategory={selectedCategory}
            selectedIcon={selectedIcon}
            onSelect={handleSelect}
          />

          <Box sx={{position: 'relative'}}>
            <Box
              sx={{
                display: 'flex',
                gap: 3,
                alignItems: 'center',
                flexWrap: {xs: 'wrap', md: 'nowrap'},
                justifyContent: 'center',
              }}
            >
              <Box
                sx={{
                  opacity: isSigninTransitioning ? 0 : 1,
                  transform: isSigninTransitioning ? 'translateY(8px)' : 'translateY(0)',
                  transition: 'opacity 0.2s ease, transform 0.2s ease',
                }}
              >
                <SigninCard flow={flow} branding={branding} />
              </Box>

              {/* Arrow connector */}
              <Box
                sx={{
                  display: {xs: 'none', md: 'flex'},
                  alignItems: 'center',
                  color: isDark ? 'rgba(255, 255, 255, 0.4)' : 'rgba(0, 0, 0, 0.25)',
                }}
              >
                <svg width="50" height="2">
                  <line x1="0" y1="1" x2="44" y2="1" stroke="currentColor" strokeWidth="1.5" strokeDasharray="4 3" />
                  <polygon points="44,0 50,1 44,2" fill="currentColor" />
                </svg>
              </Box>

              <Box
                sx={{
                  opacity: isVerifyTransitioning ? 0 : 1,
                  transform: isVerifyTransitioning ? 'translateY(8px)' : 'translateY(0)',
                  transition: 'opacity 0.2s ease, transform 0.2s ease',
                }}
              >
                <VerifyCard flow={flow} branding={branding} />
              </Box>

              {/* Check mark */}
              <Box
                sx={{
                  display: {xs: 'none', md: 'flex'},
                  width: 48, height: 48,
                  borderRadius: '50%',
                  bgcolor: isDark ? 'rgba(255, 255, 255, 0.04)' : 'rgba(0, 0, 0, 0.03)',
                  border: '2px solid', borderColor: isDark ? 'rgba(255, 255, 255, 0.15)' : 'rgba(0, 0, 0, 0.1)',
                  alignItems: 'center', justifyContent: 'center',
                  flexShrink: 0,
                }}
              >
                <svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke={isDark ? 'rgba(255,255,255,0.7)' : 'rgba(0,0,0,0.5)'} strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round">
                  <polyline points="20 6 9 17 4 12" />
                </svg>
              </Box>
            </Box>

          </Box>
        </Box>
      </Container>
    </Box>
  );
}
