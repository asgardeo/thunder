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
import {Box, Card, Typography, Container, Stack} from '@wso2/oxygen-ui';
import useIsDarkMode from '../../hooks/useIsDarkMode';
import useScrollAnimation from '../../hooks/useScrollAnimation';
import Link from '@docusaurus/Link';

export default function APIReferenceSection(): JSX.Element {
  const isDark = useIsDarkMode();
  const {ref: textRef, isVisible: textVisible} = useScrollAnimation({threshold: 0.2});
  const {ref: cardRef, isVisible: cardVisible} = useScrollAnimation({threshold: 0.1});

  return (
    <Box sx={{py: {xs: 8, lg: 12}, background: isDark ? '#0a0a0a' : 'transparent'}}>
      <Container maxWidth="lg" sx={{px: {xs: 2, sm: 4}}}>
        <Box
          sx={{
            display: 'flex',
            flexDirection: {xs: 'column', lg: 'row'},
            alignItems: 'center',
            gap: 5,
            textAlign: {xs: 'center', lg: 'left'},
          }}
        >
          <Box
            ref={textRef}
            sx={{
              flex: 1,
              opacity: textVisible ? 1 : 0,
              transform: textVisible ? 'translateX(0)' : 'translateX(-32px)',
              transition: 'opacity 0.7s cubic-bezier(0.16, 1, 0.3, 1), transform 0.7s cubic-bezier(0.16, 1, 0.3, 1)',
            }}
          >
            <Typography variant="h3" fontWeight={600} sx={{mb: 2, color: isDark ? '#ffffff' : '#1a1a2e'}}>
              REST API Reference
            </Typography>
            <Typography variant="body1" sx={{color: isDark ? 'rgba(255, 255, 255, 0.6)' : 'rgba(0, 0, 0, 0.55)', mb: 2}}>
              Integrate Thunder&apos;s authentication and identity management capabilities into your applications with
              our comprehensive REST APIs. Manage users, applications, flows, and more programmatically.
            </Typography>
            <Link
              href="/apis"
              style={{
                color: '#FF8C00',
                fontWeight: 500,
                textDecoration: 'none',
              }}
            >
              Get started with Thunder REST APIs →
            </Link>

            <Stack spacing={2} sx={{mt: 5, textAlign: 'left'}}>
              <Box>
                <Link
                  href="/api/application.yaml"
                  style={{
                    color: isDark ? '#ffffff' : '#1a1a2e',
                    fontWeight: 600,
                    textDecoration: 'none',
                    display: 'inline-block',
                  }}
                  className="api-link"
                >
                  Create an application
                  <Box
                    component="span"
                    className="arrow"
                    sx={{
                      ml: 1,
                      opacity: 0,
                      transition: 'all 0.3s',
                      display: 'inline-block',
                      '.api-link:hover &': {
                        opacity: 1,
                        transform: 'translateX(8px)',
                      },
                    }}
                  >
                    →
                  </Box>
                </Link>
                <Typography variant="body2" sx={{color: isDark ? 'rgba(255, 255, 255, 0.5)' : 'rgba(0, 0, 0, 0.45)', mt: 0.5}}>
                  Register OAuth applications with custom flows
                </Typography>
              </Box>

              <Box>
                <Link
                  href="/api/flow-management.yaml"
                  style={{
                    color: isDark ? '#ffffff' : '#1a1a2e',
                    fontWeight: 600,
                    textDecoration: 'none',
                    display: 'inline-block',
                  }}
                  className="api-link"
                >
                  Create an auth flow
                  <Box
                    component="span"
                    className="arrow"
                    sx={{
                      ml: 1,
                      opacity: 0,
                      transition: 'all 0.3s',
                      display: 'inline-block',
                      '.api-link:hover &': {
                        opacity: 1,
                        transform: 'translateX(8px)',
                      },
                    }}
                  >
                    →
                  </Box>
                </Link>
                <Typography variant="body2" sx={{color: isDark ? 'rgba(255, 255, 255, 0.5)' : 'rgba(0, 0, 0, 0.45)', mt: 0.5}}>
                  Build custom authentication flows with executors
                </Typography>
              </Box>

              <Box>
                <Link
                  href="/api/user.yaml"
                  style={{
                    color: isDark ? '#ffffff' : '#1a1a2e',
                    fontWeight: 600,
                    textDecoration: 'none',
                    display: 'inline-block',
                  }}
                  className="api-link"
                >
                  Manage users
                  <Box
                    component="span"
                    className="arrow"
                    sx={{
                      ml: 1,
                      opacity: 0,
                      transition: 'all 0.3s',
                      display: 'inline-block',
                      '.api-link:hover &': {
                        opacity: 1,
                        transform: 'translateX(8px)',
                      },
                    }}
                  >
                    →
                  </Box>
                </Link>
                <Typography variant="body2" sx={{color: isDark ? 'rgba(255, 255, 255, 0.5)' : 'rgba(0, 0, 0, 0.45)', mt: 0.5}}>
                  Create, update, and manage user accounts
                </Typography>
              </Box>
            </Stack>
          </Box>

          <Box
            ref={cardRef}
            sx={{
              flex: 1,
              display: 'flex',
              justifyContent: 'flex-end',
              maxWidth: {lg: '550px'},
              opacity: cardVisible ? 1 : 0,
              transform: cardVisible ? 'translateX(0)' : 'translateX(32px)',
              transition: 'opacity 0.7s cubic-bezier(0.16, 1, 0.3, 1) 0.15s, transform 0.7s cubic-bezier(0.16, 1, 0.3, 1) 0.15s',
            }}
          >
            <Card
              sx={{
                bgcolor: isDark ? '#0c0c0e' : '#f8f9fa',
                border: '1px solid',
                borderColor: isDark ? 'rgba(255, 255, 255, 0.08)' : 'rgba(0, 0, 0, 0.1)',
                boxShadow: isDark ? '0 8px 32px rgba(0, 0, 0, 0.4)' : '0 8px 32px rgba(0, 0, 0, 0.08)',
                borderRadius: 2,
              }}
            >
              <Box
                sx={{
                  bgcolor: isDark ? 'rgba(255, 255, 255, 0.04)' : 'rgba(0, 0, 0, 0.03)',
                  borderBottom: '1px solid',
                  borderColor: isDark ? 'rgba(255, 255, 255, 0.08)' : 'rgba(0, 0, 0, 0.08)',
                  px: 3,
                  py: 2,
                }}
              >
                <Typography
                  sx={{
                    fontSize: '0.8rem',
                    fontWeight: 700,
                    color: isDark ? 'rgba(255, 255, 255, 0.9)' : 'rgba(0, 0, 0, 0.85)',
                    letterSpacing: '0.5px',
                  }}
                >
                  POST /applications
                </Typography>
                <Typography
                  sx={{
                    fontSize: '0.7rem',
                    color: isDark ? 'rgba(255, 255, 255, 0.5)' : 'rgba(0, 0, 0, 0.45)',
                    mt: 0.5,
                  }}
                >
                  Create a new application
                </Typography>
              </Box>

              <Box
                sx={{
                  px: 3,
                  py: 2.5,
                  bgcolor: isDark ? '#0c0c0e' : '#f8f9fa',
                }}
              >
                <Box
                  sx={{
                    display: 'flex',
                    alignItems: 'center',
                    gap: 1.5,
                    mb: 2,
                  }}
                >
                  <Box
                    sx={{
                      bgcolor: '#10b981',
                      color: 'white',
                      px: 1.5,
                      py: 0.5,
                      borderRadius: 1,
                      fontSize: '0.7rem',
                      fontWeight: 700,
                      fontFamily: 'monospace',
                    }}
                  >
                    POST
                  </Box>
                  <Typography
                    sx={{
                      fontFamily: 'monospace',
                      fontSize: '0.75rem',
                      color: isDark ? 'rgba(255, 255, 255, 0.85)' : 'rgba(0, 0, 0, 0.8)',
                      fontWeight: 500,
                    }}
                  >
                    /api/v1/applications
                  </Typography>
                </Box>

                <Box sx={{mb: 2.5}}>
                  <Box
                    sx={{
                      display: 'flex',
                      alignItems: 'center',
                      gap: 0.5,
                      fontSize: '0.7rem',
                      fontWeight: 600,
                      color: isDark ? 'rgba(255, 255, 255, 0.6)' : 'rgba(0, 0, 0, 0.5)',
                      mb: 1,
                      textTransform: 'uppercase',
                      letterSpacing: '0.5px',
                    }}
                  >
                    <Typography
                      component="span"
                      sx={{
                        fontSize: '0.7rem',
                        fontWeight: 600,
                        color: isDark ? 'rgba(255, 255, 255, 0.6)' : 'rgba(0, 0, 0, 0.5)',
                        textTransform: 'uppercase',
                        letterSpacing: '0.5px',
                      }}
                    >
                      Authorization
                    </Typography>
                  </Box>
                  <Box
                    sx={{
                      display: 'flex',
                      alignItems: 'center',
                      gap: 1,
                    }}
                  >
                    <Typography
                      sx={{
                        fontSize: '0.7rem',
                        color: isDark ? 'rgba(255, 255, 255, 0.45)' : 'rgba(0, 0, 0, 0.4)',
                        minWidth: '80px',
                      }}
                    >
                      Bearer Token
                    </Typography>
                    <Box
                      sx={{
                        flex: 1,
                        height: 28,
                        bgcolor: isDark ? 'rgba(255, 255, 255, 0.06)' : 'rgba(0, 0, 0, 0.04)',
                        borderRadius: 1,
                        border: '1px solid',
                        borderColor: isDark ? 'rgba(255, 255, 255, 0.1)' : 'rgba(0, 0, 0, 0.08)',
                        display: 'flex',
                        alignItems: 'center',
                        px: 1.5,
                      }}
                    >
                      <Typography
                        sx={{
                          fontSize: '0.65rem',
                          color: isDark ? 'rgba(255, 255, 255, 0.35)' : 'rgba(0, 0, 0, 0.35)',
                          fontFamily: 'monospace',
                        }}
                      >
                        eyJhbGciOiJIUzI1NiIsInR5...
                      </Typography>
                    </Box>
                  </Box>
                </Box>

                <Box>
                  <Box
                    sx={{
                      display: 'flex',
                      alignItems: 'center',
                      gap: 0.5,
                      fontSize: '0.7rem',
                      fontWeight: 600,
                      color: isDark ? 'rgba(255, 255, 255, 0.6)' : 'rgba(0, 0, 0, 0.5)',
                      mb: 1,
                      textTransform: 'uppercase',
                      letterSpacing: '0.5px',
                    }}
                  >
                    <Typography
                      component="span"
                      sx={{
                        fontSize: '0.7rem',
                        fontWeight: 600,
                        color: isDark ? 'rgba(255, 255, 255, 0.6)' : 'rgba(0, 0, 0, 0.5)',
                        textTransform: 'uppercase',
                        letterSpacing: '0.5px',
                      }}
                    >
                      Request Body
                    </Typography>
                  </Box>
                  <Box
                    sx={{
                      bgcolor: isDark ? 'rgba(255, 255, 255, 0.03)' : 'rgba(0, 0, 0, 0.02)',
                      borderRadius: 1.5,
                      p: 2,
                      fontSize: '0.7rem',
                      fontFamily: 'Consolas, Monaco, "Courier New", monospace',
                      lineHeight: 1.7,
                      border: '1px solid',
                      borderColor: isDark ? 'rgba(255, 255, 255, 0.06)' : 'rgba(0, 0, 0, 0.06)',
                      position: 'relative',
                    }}
                  >
                    <Box
                      sx={{
                        position: 'absolute',
                        top: 8,
                        right: 8,
                        bgcolor: isDark ? 'rgba(255, 255, 255, 0.06)' : 'rgba(0, 0, 0, 0.05)',
                        px: 1,
                        py: 0.5,
                        borderRadius: 0.5,
                        fontSize: '0.6rem',
                        color: isDark ? 'rgba(255, 255, 255, 0.35)' : 'rgba(0, 0, 0, 0.4)',
                      }}
                    >
                      JSON
                    </Box>
                    <Box sx={{color: isDark ? '#e2e8f0' : '#1e293b'}}>
                      <Box sx={{display: 'flex'}}>
                        <span
                          style={{
                            color: isDark ? '#475569' : '#94a3b8',
                            width: '20px',
                            textAlign: 'right',
                            marginRight: '12px',
                            userSelect: 'none',
                          }}
                        >
                          1
                        </span>
                        <span style={{color: isDark ? '#cbd5e1' : '#334155'}}>{'{'}</span>
                      </Box>
                      <Box sx={{display: 'flex'}}>
                        <span
                          style={{
                            color: isDark ? '#475569' : '#94a3b8',
                            width: '20px',
                            textAlign: 'right',
                            marginRight: '12px',
                            userSelect: 'none',
                          }}
                        >
                          2
                        </span>
                        <span>
                          {'  '}
                          <span style={{color: isDark ? '#f472b6' : '#be185d'}}>&quot;name&quot;</span>
                          <span style={{color: isDark ? '#cbd5e1' : '#334155'}}>: </span>
                          <span style={{color: isDark ? '#fbbf24' : '#b45309'}}>&quot;My Web Application&quot;</span>
                          <span style={{color: isDark ? '#cbd5e1' : '#334155'}}>,</span>
                        </span>
                      </Box>
                      <Box sx={{display: 'flex'}}>
                        <span
                          style={{
                            color: isDark ? '#475569' : '#94a3b8',
                            width: '20px',
                            textAlign: 'right',
                            marginRight: '12px',
                            userSelect: 'none',
                          }}
                        >
                          3
                        </span>
                        <span>
                          {'  '}
                          <span style={{color: isDark ? '#f472b6' : '#be185d'}}>&quot;description&quot;</span>
                          <span style={{color: isDark ? '#cbd5e1' : '#334155'}}>: </span>
                          <span style={{color: isDark ? '#fbbf24' : '#b45309'}}>&quot;Customer portal&quot;</span>
                          <span style={{color: isDark ? '#cbd5e1' : '#334155'}}>,</span>
                        </span>
                      </Box>
                      <Box sx={{display: 'flex'}}>
                        <span
                          style={{
                            color: isDark ? '#475569' : '#94a3b8',
                            width: '20px',
                            textAlign: 'right',
                            marginRight: '12px',
                            userSelect: 'none',
                          }}
                        >
                          4
                        </span>
                        <span>
                          {'  '}
                          <span style={{color: isDark ? '#f472b6' : '#be185d'}}>&quot;auth_flow_id&quot;</span>
                          <span style={{color: isDark ? '#cbd5e1' : '#334155'}}>: </span>
                          <span style={{color: isDark ? '#fbbf24' : '#b45309'}}>&quot;edc013d0-e893-4dc0...&quot;</span>
                          <span style={{color: isDark ? '#cbd5e1' : '#334155'}}>,</span>
                        </span>
                      </Box>
                      <Box sx={{display: 'flex'}}>
                        <span
                          style={{
                            color: isDark ? '#475569' : '#94a3b8',
                            width: '20px',
                            textAlign: 'right',
                            marginRight: '12px',
                            userSelect: 'none',
                          }}
                        >
                          5
                        </span>
                        <span>
                          {'  '}
                          <span style={{color: isDark ? '#f472b6' : '#be185d'}}>&quot;template&quot;</span>
                          <span style={{color: isDark ? '#cbd5e1' : '#334155'}}>: </span>
                          <span style={{color: isDark ? '#fbbf24' : '#b45309'}}>&quot;spa&quot;</span>
                        </span>
                      </Box>
                      <Box sx={{display: 'flex'}}>
                        <span
                          style={{
                            color: isDark ? '#475569' : '#94a3b8',
                            width: '20px',
                            textAlign: 'right',
                            marginRight: '12px',
                            userSelect: 'none',
                          }}
                        >
                          6
                        </span>
                        <span style={{color: isDark ? '#cbd5e1' : '#334155'}}>{'}'}</span>
                      </Box>
                    </Box>
                  </Box>
                </Box>

                <Box sx={{mt: 2.5}}>
                  <Box
                    sx={{
                      display: 'flex',
                      alignItems: 'center',
                      gap: 0.5,
                      mb: 1,
                    }}
                  >
                    <Typography
                      sx={{
                        fontSize: '0.7rem',
                        fontWeight: 600,
                        color: isDark ? 'rgba(255, 255, 255, 0.6)' : 'rgba(0, 0, 0, 0.5)',
                        textTransform: 'uppercase',
                        letterSpacing: '0.5px',
                      }}
                    >
                      Response
                    </Typography>
                    <Box
                      sx={{
                        bgcolor: '#10b981',
                        color: 'white',
                        px: 1,
                        py: 0.25,
                        borderRadius: 0.5,
                        fontSize: '0.65rem',
                        fontWeight: 600,
                      }}
                    >
                      201
                    </Box>
                  </Box>
                  <Box
                    sx={{
                      bgcolor: isDark ? 'rgba(255, 255, 255, 0.04)' : 'rgba(0, 0, 0, 0.03)',
                      borderRadius: 1,
                      p: 1.5,
                      fontSize: '0.65rem',
                      fontFamily: 'monospace',
                      color: isDark ? 'rgba(255, 255, 255, 0.5)' : 'rgba(0, 0, 0, 0.45)',
                      border: '1px solid',
                      borderColor: isDark ? 'rgba(255, 255, 255, 0.08)' : 'rgba(0, 0, 0, 0.06)',
                    }}
                  >
                    Application created successfully
                  </Box>
                </Box>
              </Box>
            </Card>
          </Box>
        </Box>
      </Container>
    </Box>
  );
}
