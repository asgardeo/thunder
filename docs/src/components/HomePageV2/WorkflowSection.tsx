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
import useIsDarkMode from '../../hooks/useIsDarkMode';
import useScrollAnimation from '../../hooks/useScrollAnimation';

export default function WorkflowSection(): JSX.Element {
  const isDark = useIsDarkMode();
  const {ref, isVisible} = useScrollAnimation({threshold: 0.2});

  return (
    <Box
      sx={{
        py: {xs: 8, lg: 12},
        position: 'relative',
        background: isDark ? '#0a0a0a' : 'transparent',
        '&::before': {
          content: '""',
          position: 'absolute',
          top: 0,
          left: 0,
          right: 0,
          bottom: 0,
          background: isDark
            ? 'radial-gradient(ellipse at 50% 50%, rgba(255, 107, 0, 0.07) 0%, transparent 60%)'
            : 'radial-gradient(ellipse at 50% 50%, rgba(255, 107, 0, 0.04) 0%, transparent 60%)',
          pointerEvents: 'none',
        },
      }}
    >
      <Container maxWidth="lg" sx={{px: {xs: 2, sm: 4}, position: 'relative', zIndex: 1}}>
        <Box
          ref={ref}
          sx={{
            textAlign: 'center',
            maxWidth: '720px',
            mx: 'auto',
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
            Built around your workflow,{' '}
            <Box
              component="span"
              sx={{
                background: 'linear-gradient(90deg, #FF6B00 0%, #FF8C00 100%)',
                WebkitBackgroundClip: 'text',
                WebkitTextFillColor: 'transparent',
                backgroundClip: 'text',
              }}
            >
              optimized for productivity
            </Box>
          </Typography>
          <Typography
            variant="body1"
            sx={{
              fontSize: {xs: '0.95rem', sm: '1.1rem'},
              lineHeight: 1.7,
              color: isDark ? 'rgba(255, 255, 255, 0.6)' : 'rgba(0, 0, 0, 0.55)',
            }}
          >
            Thunder is engineered from the ground up to fit your workflows and toolbox, not dictate them.
          </Typography>
        </Box>
      </Container>
    </Box>
  );
}
