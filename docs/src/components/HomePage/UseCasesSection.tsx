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
import {Box, Card, Container, Typography, Button} from '@wso2/oxygen-ui';

function B2CIllustration() {
  return (
    <Box
      sx={{
        width: '100%',
        height: 200,
        borderRadius: 2,
        background: 'linear-gradient(135deg, rgba(255, 107, 0, 0.08) 0%, rgba(99, 102, 241, 0.06) 100%)',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        position: 'relative',
        overflow: 'hidden',
        mb: 2,
      }}
    >
      {/* Decorative elements */}
      <Box
        sx={{
          position: 'absolute',
          width: 120,
          height: 80,
          borderRadius: 2,
          bgcolor: 'rgba(255, 255, 255, 0.06)',
          border: '1px solid',
          borderColor: 'rgba(255, 255, 255, 0.08)',
          top: '30%',
          left: '25%',
          transform: 'rotate(-5deg)',
        }}
      >
        <Box sx={{p: 1.5}}>
          <Box sx={{width: '60%', height: 6, borderRadius: 3, bgcolor: 'rgba(255, 107, 0, 0.3)', mb: 1}} />
          <Box sx={{width: '80%', height: 4, borderRadius: 2, bgcolor: 'rgba(255, 255, 255, 0.08)', mb: 0.5}} />
          <Box sx={{width: '50%', height: 4, borderRadius: 2, bgcolor: 'rgba(255, 255, 255, 0.06)'}} />
        </Box>
      </Box>
      <Box
        sx={{
          position: 'absolute',
          width: 100,
          height: 70,
          borderRadius: 2,
          bgcolor: 'rgba(255, 255, 255, 0.04)',
          border: '1px solid',
          borderColor: 'rgba(255, 255, 255, 0.06)',
          top: '20%',
          right: '20%',
          transform: 'rotate(3deg)',
        }}
      />
      {/* Lock icons */}
      <Box sx={{position: 'absolute', top: '15%', left: '15%', opacity: 0.3, color: '#FF6B00'}}>
        <svg width="24" height="24" viewBox="0 0 24 24" fill="currentColor">
          <path d="M12 2C9.24 2 7 4.24 7 7v3H6c-1.1 0-2 .9-2 2v8c0 1.1.9 2 2 2h12c1.1 0 2-.9 2-2v-8c0-1.1-.9-2-2-2h-1V7c0-2.76-2.24-5-5-5zm0 2c1.66 0 3 1.34 3 3v3H9V7c0-1.66 1.34-3 3-3zm0 10c1.1 0 2 .9 2 2s-.9 2-2 2-2-.9-2-2 .9-2 2-2z" />
        </svg>
      </Box>
      <Box sx={{position: 'absolute', bottom: '20%', right: '15%', opacity: 0.2, color: '#FF6B00'}}>
        <svg width="20" height="20" viewBox="0 0 24 24" fill="currentColor">
          <path d="M12 1L3 5v6c0 5.55 3.84 10.74 9 12 5.16-1.26 9-6.45 9-12V5l-9-4z" />
        </svg>
      </Box>
    </Box>
  );
}

function B2BIllustration() {
  return (
    <Box
      sx={{
        width: '100%',
        height: 200,
        borderRadius: 2,
        background: 'linear-gradient(135deg, rgba(99, 102, 241, 0.08) 0%, rgba(255, 107, 0, 0.06) 100%)',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        position: 'relative',
        overflow: 'hidden',
        mb: 2,
      }}
    >
      {/* Building blocks */}
      <Box
        sx={{
          position: 'absolute',
          width: 60,
          height: 100,
          borderRadius: 1,
          bgcolor: 'rgba(255, 107, 0, 0.15)',
          border: '1px solid',
          borderColor: 'rgba(255, 107, 0, 0.2)',
          left: '30%',
          bottom: '15%',
        }}
      />
      <Box
        sx={{
          position: 'absolute',
          width: 70,
          height: 120,
          borderRadius: 1,
          bgcolor: 'rgba(255, 107, 0, 0.1)',
          border: '1px solid',
          borderColor: 'rgba(255, 107, 0, 0.15)',
          right: '30%',
          bottom: '15%',
        }}
      />
      {/* Connection arrows */}
      <Box sx={{position: 'absolute', top: '25%', right: '25%', opacity: 0.3, color: '#FF6B00'}}>
        <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
          <path d="M7 17l9.2-9.2M17 17V7H7" />
        </svg>
      </Box>
      <Box sx={{position: 'absolute', bottom: '30%', left: '20%', opacity: 0.2, color: '#FF6B00'}}>
        <svg width="20" height="20" viewBox="0 0 24 24" fill="currentColor">
          <path d="M12 1L3 5v6c0 5.55 3.84 10.74 9 12 5.16-1.26 9-6.45 9-12V5l-9-4z" />
        </svg>
      </Box>
    </Box>
  );
}

export default function UseCasesSection(): JSX.Element {
  return (
    <Box sx={{py: {xs: 8, lg: 12}}}>
      <Container maxWidth="lg" sx={{px: {xs: 2, sm: 4}}}>
        <Typography
          variant="h3"
          sx={{
            textAlign: 'center',
            mb: 6,
            fontSize: {xs: '1.75rem', sm: '2.25rem', md: '2.5rem'},
            fontWeight: 700,
            color: '#ffffff',
          }}
        >
          Designed for your use case
        </Typography>

        <Box
          sx={{
            display: 'grid',
            gridTemplateColumns: {xs: '1fr', md: 'repeat(2, 1fr)'},
            gap: 3,
          }}
        >
          {/* B2C Card */}
          <Card
            sx={{
              p: 0,
              overflow: 'hidden',
              transition: 'transform 0.3s ease',
              '&:hover': {
                transform: 'translateY(-4px)',
              },
            }}
          >
            <B2CIllustration />
            <Box sx={{p: 3}}>
              <Typography variant="h6" sx={{fontWeight: 600, mb: 1, fontSize: '1rem', color: '#ffffff'}}>
                Building for Consumers (B2C) ?
              </Typography>
              <Typography variant="body2" sx={{mb: 3, fontSize: '0.85rem', lineHeight: 1.6, color: 'rgba(255, 255, 255, 0.6)'}}>
                Create seamless, secure login experiences for e-commerce, mobile apps, and communities.
              </Typography>
              <Button
                component={Link}
                href="/docs/guides/introduction"
                variant="outlined"
                size="small"
                sx={{
                  textTransform: 'none',
                  borderRadius: 2,
                  borderColor: 'rgba(255, 255, 255, 0.2)',
                }}
                endIcon={
                  <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                    <path d="M5 12h14M12 5l7 7-7 7" />
                  </svg>
                }
              >
                Explore Consumer App Use Cases
              </Button>
            </Box>
          </Card>

          {/* B2B Card */}
          <Card
            sx={{
              p: 0,
              overflow: 'hidden',
              transition: 'transform 0.3s ease',
              '&:hover': {
                transform: 'translateY(-4px)',
              },
            }}
          >
            <B2BIllustration />
            <Box sx={{p: 3}}>
              <Typography variant="h6" sx={{fontWeight: 600, mb: 1, fontSize: '1rem', color: '#ffffff'}}>
                Building Multi-Tenanted SaaS App (B2B) ?
              </Typography>
              <Typography variant="body2" sx={{mb: 3, fontSize: '0.85rem', lineHeight: 1.6, color: 'rgba(255, 255, 255, 0.6)'}}>
                Implement organizations, team invites, and enterprise SSO with just a few API calls.
              </Typography>
              <Button
                component={Link}
                href="/docs/guides/introduction"
                variant="outlined"
                size="small"
                sx={{
                  textTransform: 'none',
                  borderRadius: 2,
                  borderColor: 'rgba(255, 255, 255, 0.2)',
                }}
                endIcon={
                  <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                    <path d="M5 12h14M12 5l7 7-7 7" />
                  </svg>
                }
              >
                Discover Multi-Tenanted SaaS Use Cases
              </Button>
            </Box>
          </Card>
        </Box>
      </Container>
    </Box>
  );
}
