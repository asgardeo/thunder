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

function SignInMockup() {
  return (
    <Card sx={{p: 2.5, width: '100%'}}>
      <Box sx={{textAlign: 'center', mb: 2}}>
        <Box sx={{width: 24, height: 24, mx: 'auto', mb: 1, color: '#FF6B00', fontSize: '1.1rem'}}>&#x2726;</Box>
        <Typography variant="body2" sx={{fontWeight: 600, fontSize: '0.85rem'}}>Sign in to Thunder</Typography>
      </Box>
      <Typography variant="caption" sx={{fontSize: '0.7rem', display: 'block', mb: 0.5}}>Email</Typography>
      <Box
        sx={{
          height: 30,
          borderRadius: 1,
          border: '1px solid',
          borderColor: (theme) =>
            theme.palette.mode === 'dark' ? 'rgba(255, 255, 255, 0.1)' : 'rgba(0, 0, 0, 0.1)',
          mb: 1.5,
          px: 1,
          display: 'flex',
          alignItems: 'center',
        }}
      >
        <Typography variant="caption" sx={{fontSize: '0.65rem', opacity: 0.4}}>Your email address</Typography>
      </Box>
      <Box
        sx={{
          height: 32,
          borderRadius: 14,
          background: 'linear-gradient(135deg, #FF6B00 0%, #FF8C00 100%)',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          mb: 1.5,
        }}
      >
        <Typography variant="caption" sx={{color: '#fff', fontSize: '0.7rem', fontWeight: 600}}>Continue</Typography>
      </Box>
      <Box
        sx={{
          display: 'flex',
          alignItems: 'center',
          gap: 1,
          my: 1.5,
        }}
      >
        <Box sx={{flex: 1, height: '1px', bgcolor: 'divider'}} />
        <Typography variant="caption" sx={{fontSize: '0.6rem', opacity: 0.5}}>OR</Typography>
        <Box sx={{flex: 1, height: '1px', bgcolor: 'divider'}} />
      </Box>
      {['Continue with Google', 'Continue with GitHub'].map((label) => (
        <Box
          key={label}
          sx={{
            height: 32,
            borderRadius: 14,
            background: 'linear-gradient(135deg, rgba(255, 107, 0, 0.15) 0%, rgba(255, 140, 0, 0.1) 100%)',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            mb: 1,
          }}
        >
          <Typography variant="caption" sx={{fontSize: '0.65rem', fontWeight: 500}}>{label}</Typography>
        </Box>
      ))}
    </Card>
  );
}

function UserProfileMockup() {
  return (
    <Card sx={{p: 2, width: '100%'}}>
      <Box sx={{display: 'flex', alignItems: 'center', gap: 1.5, mb: 1}}>
        <Box
          sx={{
            width: 32,
            height: 32,
            borderRadius: '50%',
            bgcolor: '#6366f1',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            color: '#fff',
            fontSize: '0.75rem',
            fontWeight: 600,
          }}
        >
          MA
        </Box>
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
                <><circle cx="12" cy="8" r="4" /><path d="M5 20c0-3.87 3.13-7 7-7s7 3.13 7 7" /></>
              ) : (
                <><path d="M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4" /><polyline points="16 17 21 12 16 7" /><line x1="21" y1="12" x2="9" y2="12" /></>
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
  return (
    <Card sx={{p: 2.5, width: '100%'}}>
      <Box sx={{textAlign: 'center', mb: 2}}>
        <Box sx={{width: 24, height: 24, mx: 'auto', mb: 1, color: '#FF6B00', fontSize: '1.1rem'}}>&#x2726;</Box>
        <Typography variant="body2" sx={{fontWeight: 600, fontSize: '0.85rem'}}>Sign Up to Thunder</Typography>
      </Box>
      {['Email', 'Password'].map((label) => (
        <Box key={label} sx={{mb: 1.5}}>
          <Typography variant="caption" sx={{fontSize: '0.7rem', display: 'block', mb: 0.5}}>{label}</Typography>
          <Box
            sx={{
              height: 30,
              borderRadius: 1,
              border: '1px solid',
              borderColor: (theme) =>
                theme.palette.mode === 'dark' ? 'rgba(255, 255, 255, 0.1)' : 'rgba(0, 0, 0, 0.1)',
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
          borderRadius: 14,
          background: 'linear-gradient(135deg, #FF6B00 0%, #FF8C00 100%)',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          mb: 1.5,
        }}
      >
        <Typography variant="caption" sx={{color: '#fff', fontSize: '0.7rem', fontWeight: 600}}>Continue</Typography>
      </Box>
      <Box sx={{display: 'flex', alignItems: 'center', gap: 1, my: 1.5}}>
        <Box sx={{flex: 1, height: '1px', bgcolor: 'divider'}} />
        <Typography variant="caption" sx={{fontSize: '0.6rem', opacity: 0.5}}>OR</Typography>
        <Box sx={{flex: 1, height: '1px', bgcolor: 'divider'}} />
      </Box>
      {['Continue with Google', 'Continue with GitHub', 'Continue with Microsoft'].map((label) => (
        <Box
          key={label}
          sx={{
            height: 32,
            borderRadius: 14,
            background: 'linear-gradient(135deg, rgba(255, 107, 0, 0.15) 0%, rgba(255, 140, 0, 0.1) 100%)',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            mb: 1,
          }}
        >
          <Typography variant="caption" sx={{fontSize: '0.65rem', fontWeight: 500}}>{label}</Typography>
        </Box>
      ))}
    </Card>
  );
}

function UserProfileTableMockup() {
  return (
    <Card sx={{p: 2, width: '100%'}}>
      <Box sx={{display: 'flex', alignItems: 'center', gap: 1, mb: 2}}>
        <Box
          sx={{
            width: 28,
            height: 28,
            borderRadius: '50%',
            bgcolor: '#06b6d4',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            color: '#fff',
            fontSize: '0.6rem',
            fontWeight: 600,
          }}
        >
          MA
        </Box>
      </Box>
      {[
        {label: 'Name', value: 'Mathew Asgardi'},
        {label: 'Email Address', value: 'mathew@thunder.dev'},
        {label: 'Country', value: 'United States'},
        {label: 'Phone Number', value: '+1 000 000 000'},
      ].map((row) => (
        <Box
          key={row.label}
          sx={{
            display: 'flex',
            justifyContent: 'space-between',
            alignItems: 'center',
            py: 1,
            borderBottom: '1px solid',
            borderColor: 'divider',
          }}
        >
          <Typography variant="caption" sx={{fontSize: '0.7rem', opacity: 0.6}}>{row.label}</Typography>
          <Box sx={{display: 'flex', alignItems: 'center', gap: 1}}>
            <Typography variant="caption" sx={{fontSize: '0.7rem'}}>{row.value}</Typography>
            <svg width="10" height="10" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5" opacity={0.4}>
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
      <Box
        sx={{
          width: 28,
          height: 28,
          borderRadius: '50%',
          bgcolor: '#8b5cf6',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          color: '#fff',
          fontSize: '0.6rem',
          fontWeight: 600,
        }}
      >
        MA
      </Box>
      <Typography variant="body2" sx={{fontSize: '0.8rem', fontWeight: 500}}>
        Mathew Asgardi
      </Typography>
      <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
        <polyline points="6 9 12 15 18 9" />
      </svg>
    </Card>
  );
}

export default function UIComponentsSection(): JSX.Element {
  return (
    <Box sx={{py: {xs: 8, lg: 12}}}>
      <Container maxWidth="lg" sx={{px: {xs: 2, sm: 4}}}>
        <Box
          sx={{
            display: 'grid',
            gridTemplateColumns: {xs: '1fr', md: '1fr 1.5fr'},
            gap: {xs: 4, md: 8},
            alignItems: 'center',
          }}
        >
          {/* Left: Description */}
          <Box>
            <Typography
              variant="h3"
              sx={{
                mb: 3,
                fontSize: {xs: '1.75rem', sm: '2rem', md: '2.25rem'},
                fontWeight: 700,
                lineHeight: 1.2,
              }}
            >
              Clean, Customizable UI Components Built for Devs
            </Typography>
            <Typography
              variant="body1"
              color="text.secondary"
              sx={{
                fontSize: {xs: '0.95rem', sm: '1rem'},
                lineHeight: 1.8,
              }}
            >
              Ready-to-use UI components for{' '}
              <Box component="code" sx={{color: '#FF6B00', fontSize: '0.9em'}}>{'<SignInButton />'}</Box>,{' '}
              <Box component="code" sx={{color: '#FF6B00', fontSize: '0.9em'}}>{'<SignUpButton />'}</Box>,{' '}
              <Box component="code" sx={{color: '#FF6B00', fontSize: '0.9em'}}>{'<UserProfile />'}</Box>,{' '}
              <Box component="code" sx={{color: '#FF6B00', fontSize: '0.9em'}}>{'<UserDropdown />'}</Box>{' '}
              and more for full user journeys, and style with your own CSS, and with the flexibility
              to choose between redirects or in-app experiences.
            </Typography>
          </Box>

          {/* Right: Component mockups */}
          <Box
            sx={{
              display: 'grid',
              gridTemplateColumns: {xs: '1fr', sm: 'repeat(2, 1fr)'},
              gap: 2,
            }}
          >
            <SignInMockup />
            <Box sx={{display: 'flex', flexDirection: 'column', gap: 2}}>
              <UserProfileMockup />
              <SignUpMockup />
            </Box>
          </Box>
        </Box>

        {/* Bottom row: Profile table + dropdown */}
        <Box
          sx={{
            display: 'grid',
            gridTemplateColumns: {xs: '1fr', md: '1fr 1fr 1fr'},
            gap: 2,
            mt: 3,
            alignItems: 'start',
          }}
        >
          <Box />
          <UserProfileTableMockup />
          <Box sx={{display: 'flex', flexDirection: 'column', gap: 2, alignItems: {xs: 'stretch', md: 'flex-end'}}}>
            <UserDropdownMockup />
          </Box>
        </Box>
      </Container>
    </Box>
  );
}
