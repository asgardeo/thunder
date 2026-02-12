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
import {Box, Typography, Button, AvatarGroup, Avatar, Tooltip, Skeleton, Card} from '@wso2/oxygen-ui';
import {MessagesSquare, CircleDot, ArrowUpRight} from '@wso2/oxygen-ui-icons-react';
import {useLogger} from '@thunder/logger';

interface Contributor {
  login: string;
}

export default function CommunitySection(): JSX.Element {
  const logger = useLogger('CommunitySection');

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

  return (
    <Box component="section" sx={{textDecoration: 'none'}}>
      <Box
        sx={{
          mx: 'auto',
          display: 'flex',
          width: '100%',
          flexDirection: 'column',
          alignItems: 'center',
          justifyContent: 'center',
          px: 4,
          py: 16,
        }}
      >
        <Typography variant="h2" fontWeight={600} sx={{mb: 2}}>
          Join the{' '}
          <Box component="span" sx={{color: 'primary.main'}}>
            community
          </Box>
        </Typography>
        <Typography
          variant="body1"
          sx={{
            mb: 10,
            color: (theme) => (theme.palette.mode === 'light' ? 'rgb(113, 113, 122)' : 'rgb(113, 113, 122)'),
          }}
        >
          Engage with our ever-growing community to get the latest updates, product support, and more.
        </Typography>
        {loading ? (
          <Box
            sx={{
              mx: 'auto',
              mb: 16,
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              flexWrap: 'wrap',
              gap: 1,
            }}
          >
            {Array.from({length: 12}).map((_, index) => (
              <Skeleton
                key={`${index + 1}-skeleton`}
                variant="circular"
                sx={{
                  height: {xs: 48, lg: 60},
                  width: {xs: 48, lg: 60},
                  border: 2,
                  borderColor: 'background.paper',
                }}
              />
            ))}
          </Box>
        ) : (
          <AvatarGroup
            max={12}
            sx={{
              mx: 'auto',
              mb: 16,
              '& .MuiAvatar-root': {
                width: {xs: 48, lg: 60},
                height: {xs: 48, lg: 60},
                border: 2,
                borderColor: 'background.paper',
                transition: 'all 0.2s ease',
                cursor: 'pointer',
                '&:hover': {
                  transform: 'translateY(-8px) scale(1.2)',
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
        <Box
          sx={{
            display: 'flex',
            width: '100%',
            maxWidth: 900,
            flexDirection: {xs: 'column', md: 'row'},
            alignItems: 'stretch',
            justifyContent: 'center',
            gap: 3,
          }}
        >
          <Card
            sx={{
              flex: 1,
              p: 4,
              display: 'flex',
              flexDirection: 'column',
              alignItems: 'center',
              textAlign: 'center',
              cursor: 'pointer',
              transition: 'all 0.3s ease',
              border: '1px solid',
              borderColor: 'divider',
              '&:hover': {
                transform: 'translateY(-4px)',
                boxShadow: (theme) =>
                  theme.palette.mode === 'light' ? '0 8px 24px rgba(0, 0, 0, 0.12)' : '0 8px 24px rgba(0, 0, 0, 0.4)',
                borderColor: 'primary.main',
                '& .arrow-icon': {
                  opacity: 1,
                },
              },
            }}
            onClick={() =>
              window.open('https://github.com/asgardeo/thunder/discussions', '_blank', 'noopener noreferrer')
            }
          >
            <Box
              sx={{
                width: 56,
                height: 56,
                borderRadius: 2,
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                bgcolor: 'primary.main',
                color: 'primary.contrastText',
                mb: 2,
              }}
            >
              <MessagesSquare size={28} />
            </Box>
            <Typography variant="h6" fontWeight={600} sx={{mb: 1}}>
              Join the Discussions
            </Typography>
            <Typography variant="body2" color="text.secondary" sx={{mb: 2}}>
              Connect with the community, ask questions, and share your ideas
            </Typography>
            <Button
              variant="text"
              endIcon={
                <Box
                  className="arrow-icon"
                  component="span"
                  sx={{
                    display: 'flex',
                    opacity: 0,
                    transition: 'opacity 0.2s ease',
                  }}
                >
                  <ArrowUpRight size={18} />
                </Box>
              }
              sx={{mt: 'auto'}}
            >
              Join Discussions
            </Button>
          </Card>

          <Card
            sx={{
              flex: 1,
              p: 4,
              display: 'flex',
              flexDirection: 'column',
              alignItems: 'center',
              textAlign: 'center',
              cursor: 'pointer',
              transition: 'all 0.3s ease',
              border: '1px solid',
              borderColor: 'divider',
              '&:hover': {
                transform: 'translateY(-4px)',
                boxShadow: (theme) =>
                  theme.palette.mode === 'light' ? '0 8px 24px rgba(0, 0, 0, 0.12)' : '0 8px 24px rgba(0, 0, 0, 0.4)',
                borderColor: 'primary.main',
                '& .arrow-icon': {
                  opacity: 1,
                },
              },
            }}
            onClick={() =>
              window.open(
                'https://github.com/asgardeo/thunder/issues?q=is%3Aissue%20state%3Aopen%20label%3A%22good%20first%20issue%22',
                '_blank',
                'noopener noreferrer',
              )
            }
          >
            <Box
              sx={{
                width: 56,
                height: 56,
                borderRadius: 2,
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                bgcolor: 'success.main',
                color: 'success.contrastText',
                mb: 2,
              }}
            >
              <CircleDot size={28} />
            </Box>
            <Typography variant="h6" fontWeight={600} sx={{mb: 1}}>
              Good First Issues
            </Typography>
            <Typography variant="body2" color="text.secondary" sx={{mb: 2}}>
              Start contributing with beginner-friendly issues to get involved
            </Typography>
            <Button
              variant="text"
              endIcon={
                <Box
                  className="arrow-icon"
                  component="span"
                  sx={{
                    display: 'flex',
                    opacity: 0,
                    transition: 'opacity 0.2s ease',
                  }}
                >
                  <ArrowUpRight size={18} />
                </Box>
              }
              sx={{mt: 'auto'}}
            >
              View Issues
            </Button>
          </Card>
        </Box>
      </Box>
    </Box>
  );
}
