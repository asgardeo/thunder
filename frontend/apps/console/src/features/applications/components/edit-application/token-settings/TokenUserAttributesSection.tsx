/**
 * Copyright (c) 2025, WSO2 LLC. (https://www.wso2.com).
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

import {
  Box,
  Stack,
  Typography,
  Chip,
  Divider,
  Accordion,
  AccordionSummary,
  AccordionDetails,
  Alert,
  Tooltip,
  Grid,
} from '@wso2/oxygen-ui';
import {ChevronDownIcon} from '@wso2/oxygen-ui-icons-react';
import {useTranslation} from 'react-i18next';
import type {ReactNode, Dispatch, SetStateAction} from 'react';
import {useEffect} from 'react';
import {Prism as SyntaxHighlighter} from 'react-syntax-highlighter';
import {vscDarkPlus} from 'react-syntax-highlighter/dist/esm/styles/prism';
import type {OAuth2Config} from '../../../models/oauth';
import SettingsCard from '../../../../../components/SettingsCard';
import TokenConstants from '../../../constants/token-constants';

/**
 * Props for the {@link TokenUserAttributesSection} component.
 */
interface TokenUserAttributesSectionProps {
  /**
   * Type of token being configured
   * - 'shared': Single token for native apps
   * - 'access': OAuth2 access token
   * - 'id': OIDC ID token
   * - 'userinfo': User Info response attributes
   */
  tokenType: 'shared' | 'access' | 'id' | 'userinfo';
  /**
   * Array of currently selected user attributes
   */
  currentAttributes: string[];
  /**
   * Array of all available user attributes from schemas
   */
  userAttributes: string[];
  /**
   * Loading state for user attributes fetch
   */
  isLoadingUserAttributes: boolean;
  /**
   * Set of expanded accordion section keys
   */
  expandedSections: Set<string>;
  /**
   * Setter function for expanded sections state
   */
  setExpandedSections: Dispatch<SetStateAction<Set<string>>>;
  /**
   * Set of attributes pending addition (visual feedback)
   */
  pendingAdditions: Set<string>;
  /**
   * Set of attributes pending removal (visual feedback)
   */
  pendingRemovals: Set<string>;
  /**
   * Set of attributes to highlight in the preview
   */
  highlightedAttributes: Set<string>;
  /**
   * Callback function when an attribute chip is clicked
   * @param attr - The attribute name
   * @param tokenType - The token type being modified
   */
  onAttributeClick: (attr: string, tokenType: 'shared' | 'access' | 'id' | 'userinfo') => void;
  /**
   * Currently active token tab ('access', 'id' or 'userinfo')
   */
  activeTokenType: 'access' | 'id' | 'userinfo';
  /**
   * OAuth2 configuration (optional)
   */
  oauth2Config?: OAuth2Config;
  /**
   * Optional custom action element to render in the header
   */
  headerAction?: ReactNode;
  /**
   * Optional flag to hide/disable content (e.g. when inherited)
   */
  readOnly?: boolean;
}

/**
 * Section component for managing user attributes in JWT tokens.
 *
 * Provides:
 * - Expandable sections for user attributes and default attributes
 * - Clickable chips to add/remove attributes from tokens
 * - Visual feedback for pending changes (green for additions, red for removals)
 * - JWT preview with syntax highlighting
 * - Separate attribute management for access vs ID tokens in OAuth mode
 *
 * Default attributes (aud, client_id, exp, etc.) are always included and shown in the preview.
 *
 * @param props - Component props
 * @returns User attributes selection UI within a SettingsCard
 */
export default function TokenUserAttributesSection({
  tokenType,
  currentAttributes,
  userAttributes,
  isLoadingUserAttributes,
  expandedSections,
  setExpandedSections,
  pendingAdditions,
  pendingRemovals,
  highlightedAttributes,
  onAttributeClick,
  activeTokenType,
  oauth2Config = undefined,
  headerAction = undefined,
  readOnly = false,
}: TokenUserAttributesSectionProps) {
  const {t} = useTranslation();

  /**
   * Automatically expand both the "User Attributes" and "Default Attributes" sections
   */
  useEffect(() => {
    const userKey = `user-${tokenType}`;
    const defaultKey = `default-${tokenType}`;

    if (!expandedSections.has(userKey) || !expandedSections.has(defaultKey)) {
      setExpandedSections((prev) => {
        const newSet = new Set(prev);
        newSet.add(userKey);
        newSet.add(defaultKey);
        return newSet;
      });
    }
  }, [tokenType, expandedSections, setExpandedSections]);

  const jwtPreviewForToken: Record<string, string> = {};

  const defaultAttributes =
    tokenType === 'userinfo' ? TokenConstants.USER_INFO_DEFAULT_ATTRIBUTES : TokenConstants.DEFAULT_TOKEN_ATTRIBUTES;

  defaultAttributes.forEach((attribute) => {
    jwtPreviewForToken[attribute] = `<${attribute}>`;
  });
  currentAttributes.forEach((attr: string) => {
    const isPendingRemoval = pendingRemovals.has(attr) && (tokenType === 'shared' || activeTokenType === tokenType);

    if (!isPendingRemoval) {
      jwtPreviewForToken[attr] = `<${attr}>`;
    }
  });
  pendingAdditions.forEach((attr) => {
    if (
      tokenType === 'shared' ||
      (tokenType === 'access' && activeTokenType === 'access') ||
      (tokenType === 'id' && activeTokenType === 'id') ||
      (tokenType === 'userinfo' && activeTokenType === 'userinfo')
    ) {
      jwtPreviewForToken[attr] = `<${attr}>`;
    }
  });

  const getTitle = () => {
    if (tokenType === 'access') return t('applications:edit.token.accessTokenUserAttributes.title');
    if (tokenType === 'id') return t('applications:edit.token.idTokenUserAttributes.title');
    if (tokenType === 'userinfo') return t('applications:edit.token.userInfoAttributes.title', 'User Info Attributes');
    return t('applications:edit.token.userAttributes.title');
  };

  const getDescription = () => {
    if (tokenType === 'access') return t('applications:edit.token.accessTokenUserAttributes.description');
    if (tokenType === 'id') return t('applications:edit.token.idTokenUserAttributes.description');
    if (tokenType === 'userinfo')
      return t(
        'applications:edit.token.userInfoAttributes.description',
        'Configure the user attributes to include in the User Info response.',
      );
    return t('applications:edit.token.userAttributes.description');
  };

  const getPreviewTitle = () => {
    if (tokenType === 'shared') return t('applications:edit.token.token.preview.title');
    if (tokenType === 'access') return t('applications:edit.token.accessToken.preview.title');
    if (tokenType === 'userinfo') return t('applications:edit.token.userInfo.preview.title', 'User Info JSON');
    return t('applications:edit.token.idToken.preview.title');
  };

  return (
    <SettingsCard title={getTitle()} description={getDescription()} headerAction={headerAction}>
      {!readOnly && (
        <Stack spacing={3}>
          <Box>
            <Grid container spacing={3}>
              {/* Left Column - JWT Preview */}
              <Grid size={{xs: 12, md: 6}}>
                <Box
                  sx={{
                    bgcolor: 'background.paper',
                    border: 1,
                    borderColor: 'divider',
                    borderRadius: 1,
                    p: 2,
                    height: '100%',
                  }}
                >
                  <Stack spacing={2}>
                    <Box>
                      <Stack direction="row" spacing={1} alignItems="center" sx={{mb: 1}}>
                        <Box
                          component="svg"
                          xmlns="http://www.w3.org/2000/svg"
                          xmlnsXlink="http://www.w3.org/1999/xlink"
                          viewBox="0 0 256 257"
                          width="24"
                          height="24"
                          preserveAspectRatio="xMidYMid"
                          sx={{height: 20}}
                        >
                          <path
                            fill="#FFF"
                            d="M147.386 69.071 147.129 0h-38.515l.257 69.071 19.257 26.448zM108.871 187.442v69.328h38.515v-69.328l-19.258-26.447z"
                          />
                          <path
                            fill="#00F2E6"
                            d="m147.386 187.442 40.57 55.976 31.069-22.596-40.57-55.975-31.069-10.015zM108.871 69.071 68.044 13.095 36.975 35.691l40.57 55.976 31.326 10.014z"
                          />
                          <path
                            fill="#00B9F1"
                            d="M77.545 91.667 11.811 70.355 0 106.816l65.733 21.569 31.069-10.271zM159.198 138.399l19.257 26.448 65.734 21.311L256 149.697l-65.733-21.312z"
                          />
                          <path
                            fill="#D63AFF"
                            d="M190.267 128.385 256 106.816l-11.811-36.461-65.734 21.312-19.257 26.447zM65.733 128.385 0 149.697l11.811 36.461 65.734-21.311 19.257-26.448z"
                          />
                          <path
                            fill="#FB015B"
                            d="m77.545 164.847-40.57 55.975 31.069 22.596 40.827-55.976v-32.61zM178.455 91.667l40.57-55.976-31.069-22.596-40.57 55.976v32.61z"
                          />
                        </Box>
                        <Typography variant="body1">{getPreviewTitle()}</Typography>
                      </Stack>
                    </Box>
                    <Box sx={{overflow: 'hidden', borderRadius: 1}}>
                      <SyntaxHighlighter
                        language="json"
                        style={vscDarkPlus}
                        customStyle={{
                          margin: 0,
                          padding: '16px',
                          fontSize: '0.875rem',
                          lineHeight: 1.6,
                          backgroundColor: '#1E1E1E',
                          borderRadius: '4px',
                          maxHeight: 600,
                          overflow: 'auto',
                        }}
                        showLineNumbers={false}
                        wrapLines
                      >
                        {JSON.stringify(jwtPreviewForToken, null, 2)}
                      </SyntaxHighlighter>
                    </Box>
                  </Stack>
                </Box>
              </Grid>

              {/* Right Column - Attribute Selection */}
              <Grid size={{xs: 12, md: 6}}>
                <Stack spacing={2}>
                  <Box>
                    <Typography variant="body1" sx={{mb: 1}}>
                      {t('applications:edit.token.configureAttributes')}
                    </Typography>
                    <Typography variant="body2" color="text.disabled" sx={{mb: 2}}>
                      {t('applications:edit.token.configureAttributes.hint')}
                    </Typography>

                    {/* Accordions */}
                    <Box sx={{border: 1, borderColor: 'divider', borderRadius: 1}}>
                      <Accordion
                        defaultExpanded
                        expanded={expandedSections.has(`user-${tokenType}`)}
                        onChange={(_, isExpanded) => {
                          setExpandedSections((prev) => {
                            const newSet = new Set(prev);
                            if (isExpanded) {
                              newSet.add(`user-${tokenType}`);
                            } else {
                              newSet.delete(`user-${tokenType}`);
                            }
                            return newSet;
                          });
                        }}
                        elevation={0}
                        sx={{
                          '&:before': {display: 'none'},
                          border: 'none',
                        }}
                      >
                        <AccordionSummary expandIcon={<ChevronDownIcon />}>
                          <Typography variant="subtitle2">{t('applications:edit.token.userAttributes')}</Typography>
                        </AccordionSummary>
                        <AccordionDetails>
                          {isLoadingUserAttributes && (
                            <Typography variant="body2" color="text.secondary">
                              {t('applications:edit.token.loadingAttributes')}
                            </Typography>
                          )}
                          {!isLoadingUserAttributes && userAttributes.length > 0 && (
                            <Stack direction="row" spacing={1} flexWrap="wrap" useFlexGap>
                              {(() => {
                                // Filter out system attributes for userinfo token type
                                // Create shallow copy to avoid mutating props with sort()
                                const availableAttributesRaw = Array.from(
                                  new Set([...userAttributes, ...TokenConstants.ADDITIONAL_USER_ATTRIBUTES]),
                                );

                                const availableAttributes = availableAttributesRaw.filter(
                                  (attr) => !(defaultAttributes as readonly string[]).includes(attr),
                                );

                                return availableAttributes.sort().map((attr) => {
                                  const isAdded = currentAttributes.includes(attr);
                                  const isPendingAddition =
                                    pendingAdditions.has(attr) &&
                                    (tokenType === 'shared' || activeTokenType === tokenType);
                                  const isPendingRemoval =
                                    pendingRemovals.has(attr) &&
                                    (tokenType === 'shared' || activeTokenType === tokenType);
                                  const isHighlighted = highlightedAttributes.has(attr);
                                  const isActive = (isAdded && !isPendingRemoval) || isPendingAddition;

                                  return (
                                    <Tooltip
                                      key={attr}
                                      title={
                                        isActive
                                          ? t('applications:edit.token.clickToRemove')
                                          : t('applications:edit.token.clickToAdd')
                                      }
                                    >
                                      <Chip
                                        label={attr}
                                        size="small"
                                        variant={isActive ? 'filled' : 'outlined'}
                                        color={isActive ? 'primary' : 'default'}
                                        onClick={() => onAttributeClick(attr, tokenType)}
                                        sx={{
                                          cursor: 'pointer',
                                          transition: 'all 0.3s ease',
                                          transform: isHighlighted ? 'scale(1.05)' : 'scale(1)',
                                          boxShadow: isHighlighted ? '0 0 0 2px rgba(25, 118, 210, 0.4)' : 'none',
                                          '&:hover': {
                                            backgroundColor: 'action.hover',
                                          },
                                        }}
                                      />
                                    </Tooltip>
                                  );
                                });
                              })()}
                            </Stack>
                          )}
                          {!isLoadingUserAttributes && userAttributes.length === 0 && (
                            <Alert severity="info">{t('applications:edit.token.noUserAttributes')}</Alert>
                          )}
                        </AccordionDetails>
                      </Accordion>

                      <Divider />

                      <Accordion
                        defaultExpanded
                        expanded={expandedSections.has(`default-${tokenType}`)}
                        onChange={(_, isExpanded) => {
                          setExpandedSections((prev) => {
                            const newSet = new Set(prev);
                            if (isExpanded) {
                              newSet.add(`default-${tokenType}`);
                            } else {
                              newSet.delete(`default-${tokenType}`);
                            }
                            return newSet;
                          });
                        }}
                        elevation={0}
                        sx={{
                          '&:before': {display: 'none'},
                          border: 'none',
                        }}
                      >
                        <AccordionSummary expandIcon={<ChevronDownIcon />}>
                          <Typography variant="subtitle2">{t('applications:edit.token.defaultAttributes')}</Typography>
                        </AccordionSummary>
                        <AccordionDetails>
                          <Alert severity="info" sx={{mb: 2}}>
                            {t('applications:edit.token.defaultAttributes.info')}
                          </Alert>
                          <Stack direction="row" spacing={1} flexWrap="wrap" useFlexGap>
                            {defaultAttributes.map((attribute) => (
                              <Chip key={attribute} label={attribute} size="small" variant="filled" color="default" />
                            ))}
                          </Stack>
                        </AccordionDetails>
                      </Accordion>
                    </Box>
                  </Box>
                </Stack>
              </Grid>
            </Grid>
          </Box>

          {/* Scopes Sub-section - Only for ID tokens */}
          {tokenType === 'id' && (
            <>
              <Divider />
              <Box>
                <Typography variant="h6" gutterBottom>
                  {t('applications:edit.token.labels.scopes')}
                </Typography>
                <Stack direction="row" spacing={1} flexWrap="wrap" useFlexGap>
                  {oauth2Config?.scopes && oauth2Config.scopes.length > 0 ? (
                    oauth2Config.scopes.map((scope) => (
                      <Chip key={scope} label={scope} variant="outlined" size="small" />
                    ))
                  ) : (
                    <Typography variant="body2" color="text.secondary">
                      {t('applications:edit.token.noScopes')}
                    </Typography>
                  )}
                </Stack>
              </Box>
            </>
          )}
        </Stack>
      )}
    </SettingsCard>
  );
}
