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

import React, {JSX, useState} from 'react';
import {Box, Card, Typography, Container, Stack, IconButton} from '@wso2/oxygen-ui';
import {ArrowUpRightIcon, ChevronDownIcon, ChevronRightIcon} from '@wso2/oxygen-ui-icons-react';
import Head from '@docusaurus/Head';
import Link from '@docusaurus/Link';

export default function APIReferenceSection(): JSX.Element {
  const [authExpanded, setAuthExpanded] = useState(true);
  const [bodyExpanded, setBodyExpanded] = useState(true);
  const [responseExpanded, setResponseExpanded] = useState(true);

  return (
    <Box sx={{px: {xs: 2, sm: 3}}}>
      <Head>
        <link rel="prefetch" href="/static/landing-page/api-ref-light.png" />
        <link rel="prefetch" href="/static/landing-page/api-ref-dark.png" />
      </Head>
      <Container
        maxWidth="lg"
        sx={{
          position: 'relative',
          borderRadius: 3,
          background: (theme) =>
            theme.palette.mode === 'dark'
              ? 'linear-gradient(to right, #18181b, #27272a)'
              : 'linear-gradient(to right, #09090b, #18181b)',
          px: {xs: 3, lg: 10},
          py: {xs: 10, lg: 10},
          color: 'white',
        }}
      >
        <IconButton
          component={Link}
          href="/api"
          aria-label="API Reference"
          target="_blank"
          sx={{
            position: 'absolute',
            top: 32,
            right: 32,
            width: 64,
            height: 64,
            bgcolor: 'rgba(255, 255, 255, 0.1)',
            '&:hover': {
              bgcolor: 'rgba(255, 255, 255, 0.2)',
            },
          }}
        >
          <ArrowUpRightIcon style={{fontSize: 24, color: 'rgba(255, 255, 255, 0.7)'}} />
        </IconButton>

        <Box
          sx={{
            display: 'flex',
            flexDirection: {xs: 'column', lg: 'row'},
            alignItems: 'center',
            gap: 5,
            textAlign: {xs: 'center', lg: 'left'},
          }}
        >
          <Box sx={{flex: 1}}>
            <Typography variant="h3" fontWeight={600} sx={{mb: 2, color: 'white'}}>
              REST API Reference
            </Typography>
            <Typography variant="body1" sx={{color: 'rgba(255, 255, 255, 0.7)', mb: 2}}>
              Integrate Thunder&apos;s authentication and identity management capabilities into your applications with
              our comprehensive REST APIs. Manage users, applications, flows, and more programmatically.
            </Typography>
            <Link
              href="/api"
              style={{
                color: '#a5b4fc',
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
                    color: 'white',
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
                <Typography variant="body2" sx={{color: 'rgba(255, 255, 255, 0.6)', mt: 0.5}}>
                  Register OAuth applications with custom flows
                </Typography>
              </Box>

              <Box>
                <Link
                  href="/api/flow-management.yaml"
                  style={{
                    color: 'white',
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
                <Typography variant="body2" sx={{color: 'rgba(255, 255, 255, 0.6)', mt: 0.5}}>
                  Build custom authentication flows with executors
                </Typography>
              </Box>

              <Box>
                <Link
                  href="/api/user.yaml"
                  style={{
                    color: 'white',
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
                <Typography variant="body2" sx={{color: 'rgba(255, 255, 255, 0.6)', mt: 0.5}}>
                  Create, update, and manage user accounts
                </Typography>
              </Box>
            </Stack>
          </Box>

          <Box
            sx={{
              flex: 1,
              display: 'flex',
              justifyContent: 'flex-end',
              maxWidth: {lg: '550px'},
            }}
          >
            <Card>
              <Box
                sx={{
                  bgcolor: '#f8fafc',
                  borderBottom: '1px solid #e2e8f0',
                  px: 3,
                  py: 2,
                }}
              >
                <Typography
                  sx={{
                    fontSize: '0.8rem',
                    fontWeight: 700,
                    color: '#334155',
                    letterSpacing: '0.5px',
                  }}
                >
                  POST /applications
                </Typography>
                <Typography
                  sx={{
                    fontSize: '0.7rem',
                    color: '#64748b',
                    mt: 0.5,
                  }}
                >
                  Create a new application
                </Typography>
              </Box>

              <Box sx={{px: 3, py: 2.5, bgcolor: '#ffffff'}}>
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
                      color: '#1e293b',
                      fontWeight: 500,
                    }}
                  >
                    /api/v1/applications
                  </Typography>
                </Box>

                <Box sx={{mb: 2.5}}>
                  <Box
                    onClick={() => setAuthExpanded(!authExpanded)}
                    sx={{
                      display: 'flex',
                      alignItems: 'center',
                      gap: 0.5,
                      fontSize: '0.7rem',
                      fontWeight: 600,
                      color: '#475569',
                      mb: 1,
                      textTransform: 'uppercase',
                      letterSpacing: '0.5px',
                      cursor: 'pointer',
                      userSelect: 'none',
                      '&:hover': {
                        color: '#1e293b',
                      },
                    }}
                  >
                    {authExpanded ? (
                      <ChevronDownIcon style={{fontSize: 14}} />
                    ) : (
                      <ChevronRightIcon style={{fontSize: 14}} />
                    )}
                    <Typography
                      component="span"
                      sx={{
                        fontSize: '0.7rem',
                        fontWeight: 600,
                        textTransform: 'uppercase',
                        letterSpacing: '0.5px',
                      }}
                    >
                      Authorization
                    </Typography>
                  </Box>
                  {authExpanded && (
                    <Box
                      sx={{
                        display: 'flex',
                        alignItems: 'center',
                        gap: 1,
                        ml: 2,
                      }}
                    >
                      <Typography
                        sx={{
                          fontSize: '0.7rem',
                          color: '#64748b',
                          minWidth: '80px',
                        }}
                      >
                        Bearer Token
                      </Typography>
                      <Box
                        sx={{
                          flex: 1,
                          height: 28,
                          bgcolor: '#f1f5f9',
                          borderRadius: 1,
                          border: '1px solid #cbd5e1',
                          display: 'flex',
                          alignItems: 'center',
                          px: 1.5,
                        }}
                      >
                        <Typography
                          sx={{
                            fontSize: '0.65rem',
                            color: '#94a3b8',
                            fontFamily: 'monospace',
                          }}
                        >
                          eyJhbGciOiJIUzI1NiIsInR5...
                        </Typography>
                      </Box>
                    </Box>
                  )}
                </Box>

                <Box>
                  <Box
                    onClick={() => setBodyExpanded(!bodyExpanded)}
                    sx={{
                      display: 'flex',
                      alignItems: 'center',
                      gap: 0.5,
                      fontSize: '0.7rem',
                      fontWeight: 600,
                      color: '#475569',
                      mb: 1,
                      textTransform: 'uppercase',
                      letterSpacing: '0.5px',
                      cursor: 'pointer',
                      userSelect: 'none',
                      '&:hover': {
                        color: '#1e293b',
                      },
                    }}
                  >
                    {bodyExpanded ? (
                      <ChevronDownIcon style={{fontSize: 14}} />
                    ) : (
                      <ChevronRightIcon style={{fontSize: 14}} />
                    )}
                    <Typography
                      component="span"
                      sx={{
                        fontSize: '0.7rem',
                        fontWeight: 600,
                        textTransform: 'uppercase',
                        letterSpacing: '0.5px',
                      }}
                    >
                      Request Body
                    </Typography>
                  </Box>
                  {bodyExpanded && (
                    <Box
                      sx={{
                        bgcolor: '#0f172a',
                        borderRadius: 1.5,
                        p: 2,
                        ml: 2,
                        fontSize: '0.7rem',
                        fontFamily: 'Consolas, Monaco, "Courier New", monospace',
                        lineHeight: 1.7,
                        border: '1px solid #1e293b',
                        position: 'relative',
                      }}
                    >
                      <Box
                        sx={{
                          position: 'absolute',
                          top: 8,
                          right: 8,
                          bgcolor: '#1e293b',
                          px: 1,
                          py: 0.5,
                          borderRadius: 0.5,
                          fontSize: '0.6rem',
                          color: '#64748b',
                        }}
                      >
                        JSON
                      </Box>
                      <Box sx={{color: '#e2e8f0'}}>
                        <Box sx={{display: 'flex'}}>
                          <span
                            style={{
                              color: '#475569',
                              width: '20px',
                              textAlign: 'right',
                              marginRight: '12px',
                              userSelect: 'none',
                            }}
                          >
                            1
                          </span>
                          <span style={{color: '#cbd5e1'}}>{'{'}</span>
                        </Box>
                        <Box sx={{display: 'flex'}}>
                          <span
                            style={{
                              color: '#475569',
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
                            <span style={{color: '#f472b6'}}>&quot;name&quot;</span>
                            <span style={{color: '#cbd5e1'}}>: </span>
                            <span style={{color: '#fbbf24'}}>&quot;My Web Application&quot;</span>
                            <span style={{color: '#cbd5e1'}}>,</span>
                          </span>
                        </Box>
                        <Box sx={{display: 'flex'}}>
                          <span
                            style={{
                              color: '#475569',
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
                            <span style={{color: '#f472b6'}}>&quot;description&quot;</span>
                            <span style={{color: '#cbd5e1'}}>: </span>
                            <span style={{color: '#fbbf24'}}>&quot;Customer portal&quot;</span>
                            <span style={{color: '#cbd5e1'}}>,</span>
                          </span>
                        </Box>
                        <Box sx={{display: 'flex'}}>
                          <span
                            style={{
                              color: '#475569',
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
                            <span style={{color: '#f472b6'}}>&quot;auth_flow_id&quot;</span>
                            <span style={{color: '#cbd5e1'}}>: </span>
                            <span style={{color: '#fbbf24'}}>&quot;edc013d0-e893-4dc0...&quot;</span>
                            <span style={{color: '#cbd5e1'}}>,</span>
                          </span>
                        </Box>
                        <Box sx={{display: 'flex'}}>
                          <span
                            style={{
                              color: '#475569',
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
                            <span style={{color: '#f472b6'}}>&quot;template&quot;</span>
                            <span style={{color: '#cbd5e1'}}>: </span>
                            <span style={{color: '#fbbf24'}}>&quot;spa&quot;</span>
                          </span>
                        </Box>
                        <Box sx={{display: 'flex'}}>
                          <span
                            style={{
                              color: '#475569',
                              width: '20px',
                              textAlign: 'right',
                              marginRight: '12px',
                              userSelect: 'none',
                            }}
                          >
                            6
                          </span>
                          <span style={{color: '#cbd5e1'}}>{'}'}</span>
                        </Box>
                      </Box>
                    </Box>
                  )}
                </Box>

                <Box sx={{mt: 2.5}}>
                  <Box
                    onClick={() => setResponseExpanded(!responseExpanded)}
                    sx={{
                      display: 'flex',
                      alignItems: 'center',
                      gap: 0.5,
                      mb: 1,
                      cursor: 'pointer',
                      userSelect: 'none',
                      '&:hover .response-title': {
                        color: '#1e293b',
                      },
                    }}
                  >
                    {responseExpanded ? (
                      <ChevronDownIcon style={{fontSize: 14, color: '#475569'}} />
                    ) : (
                      <ChevronRightIcon style={{fontSize: 14, color: '#475569'}} />
                    )}
                    <Typography
                      className="response-title"
                      sx={{
                        fontSize: '0.7rem',
                        fontWeight: 600,
                        color: '#475569',
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
                  {responseExpanded && (
                    <Box
                      sx={{
                        bgcolor: '#f8fafc',
                        borderRadius: 1,
                        p: 1.5,
                        ml: 2,
                        fontSize: '0.65rem',
                        fontFamily: 'monospace',
                        color: '#64748b',
                        border: '1px solid #e2e8f0',
                      }}
                    >
                      Application created successfully
                    </Box>
                  )}
                </Box>
              </Box>
            </Card>
          </Box>
        </Box>
      </Container>
    </Box>
  );
}
