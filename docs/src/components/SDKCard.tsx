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

import React, {ReactNode, useState, useEffect} from 'react';
import {Box, Card, CardContent, Typography, Chip} from '@wso2/oxygen-ui';
import Link from '@docusaurus/Link';

interface SDKCardProps {
  icon: ReactNode;
  title: string;
  packageName?: string;
  packageManager?: 'npm' | 'yarn' | 'pnpm' | 'go' | 'pip' | 'pod' | 'gradle';
  version?: string;
  description: string;
  comingSoon?: boolean;
  href?: string;
}

export default function SDKCard({
  icon,
  title,
  packageName = '',
  packageManager = 'npm',
  version = '',
  description,
  comingSoon = false,
  href = '',
}: SDKCardProps) {
  const [fetchedVersion, setFetchedVersion] = useState<string>('');

  useEffect(() => {
    // Only fetch version if:
    // 1. No version is provided
    // 2. Package name exists
    // 3. Package manager is npm
    // 4. Not coming soon
    if (!version && packageName && packageManager === 'npm' && !comingSoon) {
      fetch(`https://registry.npmjs.org/${packageName}/latest`)
        .then((res) => res.json())
        .then((data) => {
          if (data.version) {
            setFetchedVersion(`v${data.version}`);
          }
        })
        .catch(() => {
          // Silently fail if version fetch fails
        });
    }
  }, [packageName, packageManager, version, comingSoon]);

  const displayVersion = version || fetchedVersion;
  const cardContent = (
    <Card
      sx={{
        height: '100%',
        border: '1px solid',
        borderColor: 'divider',
        transition: 'all 0.3s ease',
        position: 'relative',
        overflow: 'visible',
        filter: comingSoon ? 'grayscale(1)' : 'none',
        opacity: comingSoon ? 0.7 : 1,
        '&:hover': {
          borderColor: comingSoon ? 'divider' : 'primary.main',
          transform: comingSoon ? 'none' : 'translateY(-4px)',
          boxShadow: comingSoon ? 'none' : 3,
        },
      }}
    >
      {comingSoon && (
        <Box
          sx={{
            position: 'absolute',
            top: 16,
            right: 0,
            bgcolor: 'warning.main',
            color: 'warning.contrastText',
            px: 2,
            py: 0.5,
            fontSize: '0.65rem',
            fontWeight: 600,
            textTransform: 'uppercase',
            transform: 'rotate(0deg)',
            boxShadow: 2,
            zIndex: 1,
            '&::before': {
              content: '""',
              position: 'absolute',
              left: 0,
              bottom: -8,
              width: 0,
              height: 0,
              borderLeft: '8px solid transparent',
              borderRight: '0px solid transparent',
              borderTop: '8px solid',
              borderTopColor: 'warning.dark',
            },
          }}
        >
          Coming Soon
        </Box>
      )}
      <CardContent sx={{p: 3}}>
        <Box sx={{display: 'flex', alignItems: 'flex-start', gap: 2, mb: 2}}>
          <Box
            sx={{
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              width: 42,
              height: 42,
              flexShrink: 0,
            }}
          >
            {icon}
          </Box>
          <Box sx={{flex: 1, minWidth: 0}}>
            <Box sx={{display: 'flex', alignItems: 'center', gap: 1, mb: 0.5, justifyContent: 'space-between'}}>
              <Typography variant="h6" sx={{fontWeight: 600}}>
                {title}
              </Typography>
              {displayVersion && (
                <Chip
                  label={displayVersion}
                  size="small"
                  sx={{
                    bgcolor: 'success.lighter',
                    color: 'success.dark',
                    fontWeight: 600,
                    height: 24,
                  }}
                />
              )}
            </Box>
            {packageName && (
              <Typography
                variant="body2"
                sx={{
                  color: 'text.secondary',
                  fontFamily: 'monospace',
                  fontSize: '0.75rem',
                  bgcolor: 'action.hover',
                  px: 1,
                  py: 0.5,
                  borderRadius: 0.5,
                  display: 'inline-block',
                }}
              >
                {packageName}
              </Typography>
            )}
          </Box>
        </Box>
        <Typography variant="body2" sx={{color: 'text.secondary', lineHeight: 1.6}}>
          {description}
        </Typography>
      </CardContent>
    </Card>
  );

  if (comingSoon || !href) {
    return cardContent;
  }

  return (
    <Link to={href} style={{textDecoration: 'none', display: 'block', height: '100%'}}>
      {cardContent}
    </Link>
  );
}
