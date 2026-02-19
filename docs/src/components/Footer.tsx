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
import {useBaseUrlUtils} from '@docusaurus/useBaseUrl';
import {Box, Container, Typography} from '@wso2/oxygen-ui';
import useIsDarkMode from '../hooks/useIsDarkMode';
import ThemedImage from '@theme/ThemedImage';

interface FooterColumnProps {
  title: string;
  links: {label: string; href: string}[];
}

function FooterColumn({title, links}: FooterColumnProps) {
  const isDark = useIsDarkMode();

  return (
    <Box>
      <Typography
        variant="body2"
        sx={{
          fontWeight: 600,
          mb: 2,
          fontSize: '0.85rem',
          color: isDark ? 'rgba(255, 255, 255, 0.9)' : 'rgba(0, 0, 0, 0.85)',
        }}
      >
        {title}
      </Typography>
      {links.map((link) => (
        <Typography
          key={link.label}
          component={Link}
          href={link.href}
          variant="body2"
          sx={{
            display: 'block',
            mb: 1.5,
            fontSize: '0.8rem',
            color: isDark ? 'rgba(255, 255, 255, 0.5)' : 'rgba(0, 0, 0, 0.5)',
            textDecoration: 'none',
            '&:hover': {
              color: isDark ? 'rgba(255, 255, 255, 0.8)' : 'rgba(0, 0, 0, 0.7)',
              textDecoration: 'none',
            },
          }}
        >
          {link.label}
        </Typography>
      ))}
    </Box>
  );
}

export default function Footer(): JSX.Element {
  const {withBaseUrl} = useBaseUrlUtils();
  const isDark = useIsDarkMode();

  return (
    <Box
      sx={{
        bgcolor: isDark ? '#0a0a0a' : '#f8f9fa',
        color: isDark ? '#fff' : '#1a1a2e',
        borderTop: isDark ? '1px solid rgba(255, 255, 255, 0.06)' : '1px solid rgba(0, 0, 0, 0.06)',
        pt: {xs: 6, lg: 8},
        pb: 3,
      }}
    >
      <Container maxWidth="lg" sx={{px: {xs: 2, sm: 4}}}>
        <Box
          sx={{
            display: 'grid',
            gridTemplateColumns: {xs: '1fr', sm: 'repeat(2, 1fr)', md: '2fr 1fr 1fr'},
            gap: {xs: 4, md: 6},
            mb: 6,
          }}
        >
          {/* Brand column */}
          <Box>
            <Box sx={{mb: 3}}>
              <ThemedImage
                sources={{
                  light: withBaseUrl('/assets/images/logo.svg'),
                  dark: withBaseUrl('/assets/images/logo-inverted.svg'),
                }}
                alt="Thunder Logo"
                style={{height: 32}}
              />
            </Box>
            <Typography
              variant="body2"
              sx={{
                color: isDark ? 'rgba(255, 255, 255, 0.5)' : 'rgba(0, 0, 0, 0.5)',
                fontSize: '0.85rem',
                lineHeight: 1.7,
                maxWidth: '280px',
                mb: 3,
              }}
            >
              Work together seamlessly with secure your applications with ease.
            </Typography>
            <Typography
              variant="caption"
              sx={{
                color: isDark ? 'rgba(255, 255, 255, 0.35)' : 'rgba(0, 0, 0, 0.35)',
                fontSize: '0.75rem',
              }}
            >
              Terms & Policy
            </Typography>
          </Box>

          {/* Pages column */}
          <FooterColumn
            title="Pages"
            links={[
              {label: 'Home', href: '/'},
              {label: 'Docs', href: '/docs/guides/introduction'},
              {label: 'APIs', href: '/apis'},
              {label: 'SDKs', href: '/docs/sdks/overview'},
            ]}
          />

          {/* Resources column */}
          <FooterColumn
            title="Resources"
            links={[
              {label: 'Community', href: '/docs/community/overview'},
              {label: 'Releases', href: 'https://github.com/asgardeo/thunder/releases'},
              {label: 'Discussions', href: 'https://github.com/asgardeo/thunder/discussions'},
              {label: 'Report an Issue', href: 'https://github.com/asgardeo/thunder/issues'},
            ]}
          />

        </Box>

        {/* Copyright */}
        <Box
          sx={{
            borderTop: isDark ? '1px solid rgba(255, 255, 255, 0.08)' : '1px solid rgba(0, 0, 0, 0.08)',
            pt: 3,
            textAlign: 'center',
          }}
        >
          <Typography
            variant="caption"
            sx={{
              color: isDark ? 'rgba(255, 255, 255, 0.35)' : 'rgba(0, 0, 0, 0.35)',
              fontSize: '0.75rem',
            }}
          >
            &copy; WSO2 LLC. All rights reserved.
          </Typography>
        </Box>
      </Container>
    </Box>
  );
}
