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

import React, {JSX, useEffect, useState} from 'react';
import {Box, Typography, AvatarGroup, Avatar, Tooltip, Skeleton, Card, Container} from '@wso2/oxygen-ui';
import useIsDarkMode from '../../hooks/useIsDarkMode';
import useScrollAnimation from '../../hooks/useScrollAnimation';
import {MessagesSquare, CircleDot} from '@wso2/oxygen-ui-icons-react';
import {useLogger} from '@thunder/logger';

interface Contributor {
  login: string;
}

function CommunityCard({
  icon,
  iconBg,
  title,
  description,
  linkLabel,
  href,
}: {
  icon: JSX.Element;
  iconBg: string;
  title: string;
  description: string;
  linkLabel: string;
  href: string;
}) {
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
      <Typography
        variant="h6"
        sx={{fontWeight: 600, mb: 1, color: isDark ? '#ffffff' : '#1a1a2e', fontSize: '1.1rem'}}
      >
        {title}
      </Typography>
      <Typography
        variant="body2"
        sx={{
          mb: 3,
          color: isDark ? 'rgba(255, 255, 255, 0.5)' : 'rgba(0, 0, 0, 0.45)',
          lineHeight: 1.7,
          fontSize: '0.9rem',
        }}
      >
        {description}
      </Typography>
      <Typography
        variant="body2"
        sx={{
          mt: 'auto',
          color: '#FF8C00',
          fontWeight: 500,
          fontSize: '0.9rem',
          transition: 'color 0.2s ease',
          '&:hover': {color: '#FF6B00'},
        }}
      >
        {linkLabel} &rarr;
      </Typography>
    </Card>
  );
}

export default function CommunitySection(): JSX.Element {
  const logger = useLogger('CommunitySection');
  const isDark = useIsDarkMode();
  const {ref: sectionRef, isVisible: sectionVisible} = useScrollAnimation({threshold: 0.15});

  const [contributors, setContributors] = useState<string[]>([]);
  const [loading, setLoading] = useState(true);
  const [failedImages, setFailedImages] = useState<Set<string>>(new Set());

  const handleImageError = (username: string) => {
    setFailedImages((prev) => new Set(prev).add(username));
  };

  useEffect(() => {
    fetch('https://api.github.com/repos/asgardeo/thunder/contributors?per_page=12')
      .then((response) => response.json())
      .then((data: Contributor[]) => {
        setContributors(data.map((contributor) => contributor.login));
        setLoading(false);
      })
      .catch((error) => {
        logger.error('Error fetching contributors:', {error});
        setLoading(false);
      });
  }, [logger]);

  const hasContributors = !loading && contributors.length > 0;

  return (
    <Box component="section" sx={{py: {xs: 8, lg: 12}, background: isDark ? '#0a0a0a' : 'transparent'}}>
      <Container maxWidth="lg" sx={{px: {xs: 2, sm: 4}}}>
        <Box
          ref={sectionRef}
          sx={{
            display: 'flex',
            flexDirection: 'column',
            alignItems: 'center',
            textAlign: 'center',
            opacity: sectionVisible ? 1 : 0,
            transform: sectionVisible ? 'translateY(0)' : 'translateY(32px)',
            transition: 'opacity 0.7s cubic-bezier(0.16, 1, 0.3, 1), transform 0.7s cubic-bezier(0.16, 1, 0.3, 1)',
          }}
        >
          {/* Heading */}
          <Typography
            variant="h3"
            sx={{
              mb: 2,
              fontSize: {xs: '1.75rem', sm: '2.25rem', md: '2.5rem'},
              fontWeight: 700,
              color: isDark ? '#ffffff' : '#1a1a2e',
            }}
          >
            Join the{' '}
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
              mb: hasContributors || loading ? 6 : 5,
              fontSize: {xs: '0.95rem', sm: '1.05rem'},
              color: isDark ? 'rgba(255, 255, 255, 0.6)' : 'rgba(0, 0, 0, 0.55)',
              lineHeight: 1.7,
              maxWidth: '600px',
            }}
          >
            Engage with our ever-growing community to get the latest updates, product support, and more.
          </Typography>

          {/* Contributor avatars — only show when we have data */}
          {loading && (
            <Box
              sx={{
                mx: 'auto',
                mb: 6,
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                flexWrap: 'wrap',
                gap: 0.5,
              }}
            >
              {Array.from({length: 10}).map((_, index) => (
                <Skeleton
                  key={`${index + 1}-skeleton`}
                  variant="circular"
                  sx={{
                    height: {xs: 44, lg: 52},
                    width: {xs: 44, lg: 52},
                    bgcolor: isDark ? 'rgba(255, 255, 255, 0.06)' : 'rgba(0, 0, 0, 0.06)',
                  }}
                />
              ))}
            </Box>
          )}

          {hasContributors && (
            <AvatarGroup
              max={12}
              sx={{
                mx: 'auto',
                mb: 6,
                '& .MuiAvatar-root': {
                  width: {xs: 44, lg: 52},
                  height: {xs: 44, lg: 52},
                  border: isDark ? '2px solid #141414' : '2px solid #ffffff',
                  transition: 'all 0.2s ease',
                  cursor: 'pointer',
                  '&:hover': {
                    transform: 'translateY(-6px) scale(1.15)',
                    zIndex: 1000,
                  },
                },
              }}
            >
              {contributors
                .filter((username) => !failedImages.has(username))
                .map((username) => (
                  <Tooltip key={username} title={username} arrow>
                    <Avatar
                      alt={username}
                      src={`https://github.com/${username}.png?size=96`}
                      imgProps={{loading: 'lazy'}}
                      onError={() => handleImageError(username)}
                    />
                  </Tooltip>
                ))}
            </AvatarGroup>
          )}

          {/* Cards */}
          <Box
            sx={{
              display: 'grid',
              gridTemplateColumns: {xs: '1fr', md: '1fr 1fr'},
              width: '100%',
              maxWidth: 800,
              gap: 3,
            }}
          >
            <CommunityCard
              icon={<MessagesSquare size={26} />}
              iconBg="linear-gradient(135deg, #FF6B00 0%, #FF8C00 100%)"
              title="Join the Discussions"
              description="Connect with the community, ask questions, and share your ideas"
              linkLabel="Join Discussions"
              href="https://github.com/asgardeo/thunder/discussions"
            />
            <CommunityCard
              icon={<CircleDot size={26} />}
              iconBg="linear-gradient(135deg, #22c55e 0%, #16a34a 100%)"
              title="Good First Issues"
              description="Start contributing with beginner-friendly issues to get involved"
              linkLabel="View Issues"
              href="https://github.com/asgardeo/thunder/issues?q=is%3Aissue%20state%3Aopen%20label%3A%22good%20first%20issue%22"
            />
          </Box>
        </Box>
      </Container>
    </Box>
  );
}
