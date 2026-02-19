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

function BrandingConfigPanel() {
  return (
    <Card
      sx={{
        p: 3,
        display: 'flex',
        gap: 4,
        flexWrap: 'wrap',
      }}
    >
      {/* Logo */}
      <Box>
        <Typography variant="body2" sx={{fontWeight: 600, mb: 1.5, fontSize: '0.8rem'}}>
          Logo
        </Typography>
        <Box sx={{display: 'flex', gap: 1}}>
          {['#FF6B00', '#6366f1', '#06b6d4'].map((color) => (
            <Box
              key={color}
              sx={{
                width: 44,
                height: 44,
                borderRadius: 1.5,
                bgcolor: (theme) =>
                  theme.palette.mode === 'dark' ? 'rgba(255, 255, 255, 0.06)' : 'rgba(0, 0, 0, 0.04)',
                border: '1px solid',
                borderColor: (theme) =>
                  theme.palette.mode === 'dark' ? 'rgba(255, 255, 255, 0.1)' : 'rgba(0, 0, 0, 0.08)',
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
              }}
            >
              <Box
                sx={{
                  width: 20,
                  height: 20,
                  borderRadius: '50%',
                  bgcolor: color,
                  opacity: 0.8,
                }}
              />
            </Box>
          ))}
        </Box>
      </Box>

      {/* Colors */}
      <Box>
        <Typography variant="body2" sx={{fontWeight: 600, mb: 1.5, fontSize: '0.8rem'}}>
          Colors
        </Typography>
        <Box sx={{display: 'flex', gap: 0.8}}>
          {['#ef4444', '#3b82f6', '#10b981', '#8b5cf6'].map((color) => (
            <Box
              key={color}
              sx={{
                width: 28,
                height: 28,
                borderRadius: 1,
                bgcolor: color,
              }}
            />
          ))}
        </Box>
      </Box>

      {/* Border Radius */}
      <Box>
        <Typography variant="body2" sx={{fontWeight: 600, mb: 1.5, fontSize: '0.8rem'}}>
          Border Radius
        </Typography>
        <Box sx={{display: 'flex', gap: 1, alignItems: 'center'}}>
          <Box
            sx={{
              width: 60,
              height: 28,
              borderRadius: 14,
              bgcolor: (theme) =>
                theme.palette.mode === 'dark' ? 'rgba(255, 255, 255, 0.15)' : 'rgba(0, 0, 0, 0.1)',
            }}
          />
          <Box
            sx={{
              width: 60,
              height: 28,
              borderRadius: 1,
              bgcolor: (theme) =>
                theme.palette.mode === 'dark' ? 'rgba(255, 255, 255, 0.08)' : 'rgba(0, 0, 0, 0.05)',
            }}
          />
        </Box>
      </Box>
    </Card>
  );
}

function AuthOptionsCard() {
  return (
    <Card sx={{p: 3}}>
      <Box sx={{mb: 2.5}}>
        <Typography variant="body2" sx={{fontWeight: 600, mb: 1.5, fontSize: '0.8rem', fontFamily: 'var(--ifm-font-family-monospace)'}}>
          Social Logins
        </Typography>
        <Box sx={{display: 'flex', gap: 1}}>
          {['F', 'G', 'G'].map((letter, i) => (
            <Box
              key={i}
              sx={{
                width: 40,
                height: 40,
                borderRadius: '50%',
                bgcolor: (theme) =>
                  theme.palette.mode === 'dark' ? 'rgba(255, 255, 255, 0.08)' : 'rgba(0, 0, 0, 0.06)',
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                fontSize: '0.85rem',
                fontWeight: 600,
                opacity: 0.7,
              }}
            >
              {letter}
            </Box>
          ))}
        </Box>
      </Box>

      <Box sx={{mb: 2.5}}>
        <Typography variant="body2" sx={{fontWeight: 600, mb: 1.5, fontSize: '0.8rem', fontFamily: 'var(--ifm-font-family-monospace)'}}>
          Passwordless Login
        </Typography>
        <Box sx={{display: 'flex', gap: 1}}>
          {[1, 2, 3].map((i) => (
            <Box
              key={i}
              sx={{
                width: 40,
                height: 40,
                borderRadius: '50%',
                bgcolor: (theme) =>
                  theme.palette.mode === 'dark' ? 'rgba(255, 255, 255, 0.08)' : 'rgba(0, 0, 0, 0.06)',
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
              }}
            >
              <Box
                component="span"
                sx={{
                  width: 16,
                  height: 16,
                  borderRadius: '50%',
                  bgcolor: (theme) =>
                    theme.palette.mode === 'dark' ? 'rgba(255, 255, 255, 0.15)' : 'rgba(0, 0, 0, 0.1)',
                }}
              />
            </Box>
          ))}
        </Box>
      </Box>

      <Box>
        <Typography variant="body2" sx={{fontWeight: 600, mb: 1.5, fontSize: '0.8rem', fontFamily: 'var(--ifm-font-family-monospace)'}}>
          Multi-Factor Authentication
        </Typography>
        <Box sx={{display: 'flex', gap: 1}}>
          {[1, 2, 3].map((i) => (
            <Box
              key={i}
              sx={{
                width: 40,
                height: 40,
                borderRadius: '50%',
                bgcolor: (theme) =>
                  i === 2
                    ? theme.palette.mode === 'dark'
                      ? 'rgba(255, 107, 0, 0.2)'
                      : 'rgba(255, 107, 0, 0.1)'
                    : theme.palette.mode === 'dark'
                      ? 'rgba(255, 255, 255, 0.08)'
                      : 'rgba(0, 0, 0, 0.06)',
                border: i === 2 ? '1px solid' : 'none',
                borderColor: 'rgba(255, 107, 0, 0.3)',
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
              }}
            >
              <Box
                component="span"
                sx={{
                  width: 16,
                  height: 16,
                  borderRadius: '50%',
                  bgcolor: (theme) =>
                    theme.palette.mode === 'dark' ? 'rgba(255, 255, 255, 0.15)' : 'rgba(0, 0, 0, 0.1)',
                }}
              />
            </Box>
          ))}
        </Box>
      </Box>
    </Card>
  );
}

function MiniLoginCard({title}: {title: string}) {
  return (
    <Card sx={{p: 2.5, width: {xs: '100%', sm: 220}, flexShrink: 0}}>
      <Box sx={{textAlign: 'center', mb: 1.5}}>
        <Box
          sx={{
            width: 28,
            height: 28,
            mx: 'auto',
            mb: 1,
            color: '#FF6B00',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            fontSize: '1.2rem',
          }}
        >
          &#x2726;
        </Box>
        <Typography variant="body2" sx={{fontWeight: 600, fontSize: '0.8rem'}}>
          {title}
        </Typography>
      </Box>
      {title.includes('Sign in') && (
        <>
          <Typography variant="caption" sx={{fontSize: '0.7rem', display: 'block', mb: 0.5}}>
            Email
          </Typography>
          <Box
            sx={{
              height: 28,
              borderRadius: 1,
              border: '1px solid',
              borderColor: (theme) =>
                theme.palette.mode === 'dark' ? 'rgba(255, 255, 255, 0.1)' : 'rgba(0, 0, 0, 0.1)',
              mb: 1.5,
            }}
          />
          <Box
            sx={{
              height: 30,
              borderRadius: 14,
              background: 'linear-gradient(135deg, #FF6B00 0%, #FF8C00 100%)',
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
            }}
          >
            <Typography variant="caption" sx={{color: '#fff', fontSize: '0.7rem', fontWeight: 600}}>
              Continue
            </Typography>
          </Box>
        </>
      )}
      {title.includes('Verify') && (
        <>
          <Typography variant="caption" sx={{fontSize: '0.7rem', display: 'block', mb: 0.5}}>
            Enter OTP
          </Typography>
          <Box sx={{display: 'flex', gap: 0.5, mb: 1.5, justifyContent: 'center'}}>
            {[1, 2, 3, 4, 5].map((i) => (
              <Box
                key={i}
                sx={{
                  width: 28,
                  height: 32,
                  borderRadius: 1,
                  bgcolor: (theme) =>
                    theme.palette.mode === 'dark' ? 'rgba(255, 255, 255, 0.08)' : 'rgba(0, 0, 0, 0.06)',
                  border: '1px solid',
                  borderColor: (theme) =>
                    theme.palette.mode === 'dark' ? 'rgba(255, 255, 255, 0.1)' : 'rgba(0, 0, 0, 0.08)',
                }}
              />
            ))}
          </Box>
          <Box
            sx={{
              height: 30,
              borderRadius: 14,
              background: 'linear-gradient(135deg, #FF6B00 0%, #FF8C00 100%)',
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
            }}
          >
            <Typography variant="caption" sx={{color: '#fff', fontSize: '0.7rem', fontWeight: 600}}>
              Submit
            </Typography>
          </Box>
        </>
      )}
    </Card>
  );
}

export default function AIFlowsSection(): JSX.Element {
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
            Beautiful, AI-assisted Login and Registration Flows
          </Typography>
          <Typography
            variant="body1"
            color="text.secondary"
            sx={{maxWidth: '600px', mx: 'auto', fontSize: {xs: '0.95rem', sm: '1.05rem'}}}
          >
            Streamlined, secure, and customizable visual user flows powered by AI.
          </Typography>
        </Box>

        {/* AI Prompt + Branding Config */}
        <Box
          sx={{
            display: 'grid',
            gridTemplateColumns: {xs: '1fr', md: '1fr 1.5fr'},
            gap: 4,
            mb: 6,
          }}
        >
          {/* AI Prompt Input */}
          <Box>
            <Typography
              variant="body1"
              sx={{
                mb: 2,
                fontSize: '1rem',
              }}
            >
              Generate the{' '}
              <Box component="span" sx={{color: '#FF6B00', fontWeight: 600}}>
                Login flow with AI
              </Box>{' '}
              &#x2728;
            </Typography>
            <Card
              sx={{
                p: 2,
                position: 'relative',
              }}
            >
              <Typography
                variant="body2"
                sx={{
                  fontSize: '0.85rem',
                  color: 'text.secondary',
                  lineHeight: 1.6,
                }}
              >
                Create a flow with basic login with google social login and SMS OTP as the second factor
              </Typography>
              <Box
                sx={{
                  position: 'absolute',
                  bottom: 8,
                  right: 8,
                  width: 28,
                  height: 28,
                  borderRadius: '50%',
                  bgcolor: '#FF6B00',
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'center',
                  cursor: 'pointer',
                  opacity: 0.8,
                }}
              >
                <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="#fff" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round">
                  <path d="M21 12a9 9 0 1 1-9-9c2.52 0 4.93 1 6.74 2.74L21 8" />
                  <path d="M21 3v5h-5" />
                </svg>
              </Box>
            </Card>
          </Box>

          {/* Branding Config Panel */}
          <BrandingConfigPanel />
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
          <AuthOptionsCard />

          <Box
            sx={{
              display: 'flex',
              gap: 3,
              alignItems: 'center',
              flexWrap: {xs: 'wrap', md: 'nowrap'},
              justifyContent: 'center',
            }}
          >
            <MiniLoginCard title="Sign in to Thunder" />

            {/* Arrow connector */}
            <Box
              sx={{
                display: {xs: 'none', md: 'flex'},
                alignItems: 'center',
                color: 'text.secondary',
                opacity: 0.4,
              }}
            >
              <svg width="40" height="2" fill="currentColor">
                <line x1="0" y1="1" x2="36" y2="1" stroke="currentColor" strokeWidth="1.5" strokeDasharray="4 3" />
                <polygon points="36,0 40,1 36,2" />
              </svg>
            </Box>

            <MiniLoginCard title="Verify your mobile" />

            {/* Check mark */}
            <Box
              sx={{
                display: {xs: 'none', md: 'flex'},
                width: 40,
                height: 40,
                borderRadius: '50%',
                bgcolor: (theme) =>
                  theme.palette.mode === 'dark' ? 'rgba(255, 255, 255, 0.06)' : 'rgba(0, 0, 0, 0.04)',
                border: '1px solid',
                borderColor: (theme) =>
                  theme.palette.mode === 'dark' ? 'rgba(255, 255, 255, 0.1)' : 'rgba(0, 0, 0, 0.08)',
                alignItems: 'center',
                justifyContent: 'center',
                flexShrink: 0,
              }}
            >
              <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round">
                <polyline points="20 6 9 17 4 12" />
              </svg>
            </Box>
          </Box>
        </Box>
      </Container>
    </Box>
  );
}
