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

import {Stack} from '@wso2/oxygen-ui';
import {useTranslation} from 'react-i18next';
import {useState, useEffect, useMemo, useRef} from 'react';
import {useForm} from 'react-hook-form';
import {zodResolver} from '@hookform/resolvers/zod';
import {z} from 'zod';
import {useQuery} from '@tanstack/react-query';
import {useAsgardeo} from '@asgardeo/react';
import {useConfig} from '@thunder/shared-contexts';
import {useLogger} from '@thunder/logger';
import type {OAuth2Config} from '../../../models/oauth';
import type {Application} from '../../../models/application';
import type {PropertyDefinition, ApiUserSchema} from '../../../../user-types/types/user-types';
import TokenIssuerSection from './TokenIssuerSection';
import TokenUserAttributesSection from './TokenUserAttributesSection';
import TokenValidationSection from './TokenValidationSection';

interface UserSchemaListResponse {
  totalResults: number;
  startIndex: number;
  count: number;
  schemas: {
    id: string;
    name: string;
  }[];
}

/**
 * Temporary local hook to fetch user types list.
 * TODO: Remove this once the parent hooks are fixed.
 * Tracker: https://github.com/asgardeo/thunder/issues/1159
 */
function useGetUserTypes() {
  const {http} = useAsgardeo();
  const {getServerUrl} = useConfig();

  return useQuery<UserSchemaListResponse>({
    queryKey: ['user-types-list'],
    queryFn: async (): Promise<UserSchemaListResponse> => {
      const serverUrl = getServerUrl();
      const response = await http.request({
        url: `${serverUrl}/user-schemas?limit=100`,
        method: 'GET',
      } as unknown as Parameters<typeof http.request>[0]);

      return response.data as UserSchemaListResponse;
    },
  });
}

/**
 * Props for the {@link EditTokenSettings} component.
 */
interface EditTokenSettingsProps {
  /**
   * The application being edited
   */
  application: Application;
  /**
   * OAuth2 configuration containing token settings (optional)
   */
  oauth2Config?: OAuth2Config;
  /**
   * Callback function to handle field value changes
   * @param field - The application field being updated
   * @param value - The new value for the field
   */
  onFieldChange: (field: keyof Application, value: unknown) => void;
}

const createTokenConfigSchema = (t: (key: string) => string) =>
  z.object({
    validityPeriod: z.number().min(1, t('applications:edit.token.validity.error')),
    accessTokenValidity: z.number().min(1, t('applications:edit.token.validity.error')),
    idTokenValidity: z.number().min(1, t('applications:edit.token.validity.error')),
    issuer: z.string().url(t('applications:edit.token.issuer.error')).or(z.literal('')).optional(),
  });

type TokenConfigFormData = z.infer<ReturnType<typeof createTokenConfigSchema>>;

/**
 * Container component for token configuration settings.
 *
 * Manages token settings for both OAuth2/OIDC mode and Native mode:
 * - OAuth2/OIDC mode: Separate access token and ID token configurations
 * - Native mode: Shared token configuration
 *
 * Provides sections for:
 * - Token validity periods (with real-time validation)
 * - Token issuer URL configuration
 * - User attributes to include in tokens
 * - JWT preview with syntax highlighting
 *
 * Features:
 * - Fetches user schemas from available user types
 * - Debounced updates (500ms) when changes are made
 * - Visual feedback for pending additions/removals
 * - Tab-based interface for access vs ID tokens in OAuth mode
 *
 * @param props - Component props
 * @returns Token settings UI sections wrapped in a Stack
 */
export default function EditTokenSettings({
  application,
  oauth2Config = undefined,
  onFieldChange,
}: EditTokenSettingsProps) {
  const logger = useLogger('EditTokenSettings');
  const {t} = useTranslation();
  const {http} = useAsgardeo();
  const {getServerUrl} = useConfig();

  const applyTimeoutRef = useRef<NodeJS.Timeout | null>(null);
  const [expandedSections, setExpandedSections] = useState<Set<string>>(new Set(['user', 'default']));
  const [userSchemas, setUserSchemas] = useState<ApiUserSchema[]>([]);

  const {data: userTypesData, isLoading: userTypesLoading} = useGetUserTypes();
  const [activeTokenType, setActiveTokenType] = useState<'access' | 'id'>('access');
  const [pendingAdditions, setPendingAdditions] = useState<Set<string>>(new Set());
  const [pendingRemovals, setPendingRemovals] = useState<Set<string>>(new Set());
  const [highlightedAttributes, setHighlightedAttributes] = useState<Set<string>>(new Set());

  // Stabilize allowed_user_types array reference
  const allowedUserTypes = useMemo(() => application.allowed_user_types ?? [], [application.allowed_user_types]);

  // Get schema IDs for allowed user types
  const schemaIds = useMemo(() => {
    if (!userTypesData?.schemas || allowedUserTypes.length === 0) {
      return [];
    }

    return userTypesData.schemas.filter((schema) => allowedUserTypes.includes(schema.name)).map((schema) => schema.id);
  }, [userTypesData, allowedUserTypes]);

  // Determine if this is OAuth/OIDC mode (has separate token configs) or Native mode
  const isOAuthMode = useMemo(
    () => oauth2Config?.token?.access_token !== undefined || oauth2Config?.token?.id_token !== undefined,
    [oauth2Config],
  );

  const tokenConfigSchema = useMemo(() => createTokenConfigSchema(t), [t]);

  const {
    control,
    formState: {errors},
    setValue,
    watch,
  } = useForm<TokenConfigFormData>({
    resolver: zodResolver(tokenConfigSchema),
    mode: 'onChange',
    defaultValues: {
      validityPeriod: oauth2Config?.token?.validity_period ?? application.token?.validity_period ?? 3600,
      accessTokenValidity: oauth2Config?.token?.access_token?.validity_period ?? 3600,
      idTokenValidity: oauth2Config?.token?.id_token?.validity_period ?? 3600,
      issuer: oauth2Config?.token?.issuer ?? application.token?.issuer ?? '',
    },
  });

  const validityPeriod = watch('validityPeriod');
  const accessTokenValidity = watch('accessTokenValidity');
  const idTokenValidity = watch('idTokenValidity');
  const issuer = watch('issuer');

  /**
   * Sync form values when the OAuth2 configuration or application token configuration changes.
   */
  useEffect(() => {
    if (isOAuthMode) {
      setValue('accessTokenValidity', oauth2Config?.token?.access_token?.validity_period ?? 3600);
      setValue('idTokenValidity', oauth2Config?.token?.id_token?.validity_period ?? 3600);
    } else {
      setValue('validityPeriod', oauth2Config?.token?.validity_period ?? application.token?.validity_period ?? 3600);
    }
    setValue('issuer', oauth2Config?.token?.issuer ?? application.token?.issuer ?? '');
  }, [isOAuthMode, oauth2Config, application.token?.validity_period, application.token?.issuer, setValue]);

  /**
   * Effect to sync form changes back to the parent component.
   */
  useEffect(() => {
    if (isOAuthMode && oauth2Config) {
      // OAuth mode: update separate access and ID token configs
      const updatedConfig = {
        ...oauth2Config,
        token: {
          ...oauth2Config.token,
          issuer,
          access_token: {
            ...oauth2Config.token?.access_token,
            validity_period: accessTokenValidity,
          },
          id_token: {
            ...oauth2Config.token?.id_token,
            validity_period: idTokenValidity,
          },
        },
      };

      const updatedInboundAuth = application.inbound_auth_config?.map((config) => {
        if (config.type === 'oauth2') {
          return {...config, config: updatedConfig};
        }

        return config;
      });

      onFieldChange('inbound_auth_config', updatedInboundAuth);
    } else if (!isOAuthMode) {
      // Native mode: update root-level token config
      const updatedToken = {
        ...application.token,
        validity_period: validityPeriod,
        issuer,
      };

      onFieldChange('token', updatedToken);
    }
  }, [
    validityPeriod,
    accessTokenValidity,
    idTokenValidity,
    issuer,
    isOAuthMode,
    oauth2Config,
    application.inbound_auth_config,
    application.token,
    onFieldChange,
  ]);

  /**
   * Fetch user schemas for all allowed user types
   */
  useEffect(() => {
    if (schemaIds.length === 0) {
      setUserSchemas([]);
      return;
    }

    const fetchSchemas = async () => {
      const serverUrl = getServerUrl();

      try {
        const schemaPromises = schemaIds.map(async (id) => {
          try {
            const response = await http.request({
              url: `${serverUrl}/user-schemas/${id}`,
              method: 'GET',
            } as unknown as Parameters<typeof http.request>[0]);
            return response.data as ApiUserSchema;
          } catch (err) {
            logger.error('Failed to fetch user schema', {error: err, userSchemaId: id});
            return null;
          }
        });

        const responses = await Promise.all(schemaPromises);
        const schemas = responses.filter((schema): schema is ApiUserSchema => schema !== null);
        setUserSchemas(schemas);
      } catch (err) {
        logger.error('Failed to fetch user schemas', {error: err});
        setUserSchemas([]);
      }
    };

    fetchSchemas().catch((err) => {
      logger.error('Unexpected error in fetchUserSchemas', {error: err});
    });
  }, [schemaIds, http, getServerUrl, logger]);

  const userAttributes = useMemo(() => {
    if (userSchemas.length === 0) return [];

    const flattenAttributes = (schema: Record<string, PropertyDefinition>, prefix = ''): string[] => {
      const attributes: string[] = [];

      Object.entries(schema).forEach(([key, value]) => {
        const fullKey = `${prefix}${key}`;

        if (value.type === 'object' && 'properties' in value) {
          // Recursively flatten nested objects
          attributes.push(...flattenAttributes(value.properties, `${fullKey}.`));
        } else if (value.type !== 'array') {
          // Add primitive types (string, number, boolean)
          attributes.push(fullKey);
        }
      });

      return attributes;
    };

    // Combine attributes from all allowed user types and remove duplicates
    const allAttributes = new Set<string>();
    userSchemas.forEach((userSchema) => {
      const attributes = flattenAttributes(userSchema.schema);
      attributes.forEach((attr) => allAttributes.add(attr));
    });

    return Array.from(allAttributes).sort();
  }, [userSchemas]);

  const isLoadingUserAttributes = userTypesLoading;

  const sharedUserAttributes = useMemo(() => {
    if (isOAuthMode) {
      // For OAuth mode, this is not used but kept for compatibility
      return [];
    }

    return oauth2Config?.token?.user_attributes ?? application.token?.user_attributes ?? [];
  }, [isOAuthMode, oauth2Config, application]);

  const currentAccessTokenAttributes = useMemo(
    () => oauth2Config?.token?.access_token?.user_attributes ?? [],
    [oauth2Config],
  );

  const currentIdTokenAttributes = useMemo(() => oauth2Config?.token?.id_token?.user_attributes ?? [], [oauth2Config]);

  /**
   * Effect to apply pending additions and removals after a debounce period.
   */
  useEffect(() => {
    if (pendingAdditions.size === 0 && pendingRemovals.size === 0) {
      return undefined;
    }

    // Clear existing timeout
    if (applyTimeoutRef.current) {
      clearTimeout(applyTimeoutRef.current);
    }

    // Set new timeout to apply changes
    applyTimeoutRef.current = setTimeout(() => {
      // Apply additions
      if (pendingAdditions.size > 0) {
        const additionsArray = Array.from(pendingAdditions);

        if (isOAuthMode && oauth2Config) {
          // OAuth mode: update the active token type
          if (activeTokenType === 'access') {
            const newAttributes = [
              ...currentAccessTokenAttributes,
              ...additionsArray.filter((attr) => !currentAccessTokenAttributes.includes(attr)),
            ];
            const updatedConfig = {
              ...oauth2Config,
              token: {
                ...oauth2Config.token,
                access_token: {
                  ...oauth2Config.token?.access_token,
                  user_attributes: newAttributes,
                },
              },
            };
            const updatedInboundAuth = application.inbound_auth_config?.map((config) => {
              if (config.type === 'oauth2') {
                return {...config, config: updatedConfig};
              }
              return config;
            });
            onFieldChange('inbound_auth_config', updatedInboundAuth);
          } else {
            const newAttributes = [
              ...currentIdTokenAttributes,
              ...additionsArray.filter((attr) => !currentIdTokenAttributes.includes(attr)),
            ];
            const updatedConfig = {
              ...oauth2Config,
              token: {
                ...oauth2Config.token,
                id_token: {
                  ...oauth2Config.token?.id_token,
                  user_attributes: newAttributes,
                },
              },
            };
            const updatedInboundAuth = application.inbound_auth_config?.map((config) => {
              if (config.type === 'oauth2') {
                return {...config, config: updatedConfig};
              }
              return config;
            });
            onFieldChange('inbound_auth_config', updatedInboundAuth);
          }
        } else {
          // Native mode: update root-level token attributes
          const newAttributes = [
            ...sharedUserAttributes,
            ...additionsArray.filter((attr) => !sharedUserAttributes.includes(attr)),
          ];
          const updatedToken = {
            ...application.token,
            user_attributes: newAttributes,
          };
          onFieldChange('token', updatedToken);
        }
      }

      // Apply removals
      if (pendingRemovals.size > 0) {
        const removalsArray = Array.from(pendingRemovals);

        if (isOAuthMode && oauth2Config) {
          // OAuth mode: update the active token type
          if (activeTokenType === 'access') {
            const newAttributes = currentAccessTokenAttributes.filter((attr) => !removalsArray.includes(attr));
            const updatedConfig = {
              ...oauth2Config,
              token: {
                ...oauth2Config.token,
                access_token: {
                  ...oauth2Config.token?.access_token,
                  user_attributes: newAttributes,
                },
              },
            };
            const updatedInboundAuth = application.inbound_auth_config?.map((config) => {
              if (config.type === 'oauth2') {
                return {...config, config: updatedConfig};
              }
              return config;
            });
            onFieldChange('inbound_auth_config', updatedInboundAuth);
          } else {
            const newAttributes = currentIdTokenAttributes.filter((attr) => !removalsArray.includes(attr));
            const updatedConfig = {
              ...oauth2Config,
              token: {
                ...oauth2Config.token,
                id_token: {
                  ...oauth2Config.token?.id_token,
                  user_attributes: newAttributes,
                },
              },
            };
            const updatedInboundAuth = application.inbound_auth_config?.map((config) => {
              if (config.type === 'oauth2') {
                return {...config, config: updatedConfig};
              }
              return config;
            });
            onFieldChange('inbound_auth_config', updatedInboundAuth);
          }
        } else {
          // Native mode: update root-level token attributes
          const newAttributes = sharedUserAttributes.filter((attr) => !removalsArray.includes(attr));
          const updatedToken = {
            ...application.token,
            user_attributes: newAttributes,
          };
          onFieldChange('token', updatedToken);
        }
      }

      // Don't clear pending changes immediately - let the next effect clean them up
      // when the config actually updates
    }, 800);

    // Cleanup timeout on unmount
    return () => {
      if (applyTimeoutRef.current) {
        clearTimeout(applyTimeoutRef.current);
      }
    };
  }, [
    pendingAdditions,
    pendingRemovals,
    isOAuthMode,
    oauth2Config,
    activeTokenType,
    currentAccessTokenAttributes,
    currentIdTokenAttributes,
    sharedUserAttributes,
    application.inbound_auth_config,
    application.token,
    onFieldChange,
  ]);

  // Clean up pending additions/removals once they're reflected in the actual config
  useEffect(() => {
    if (pendingAdditions.size > 0) {
      let currentAttrs: string[];
      if (isOAuthMode) {
        currentAttrs = activeTokenType === 'access' ? currentAccessTokenAttributes : currentIdTokenAttributes;
      } else {
        currentAttrs = sharedUserAttributes;
      }

      const stillPending = Array.from(pendingAdditions).filter((attr) => !currentAttrs.includes(attr));

      if (stillPending.length !== pendingAdditions.size) {
        setPendingAdditions(new Set(stillPending));
      }
    }

    if (pendingRemovals.size > 0) {
      let currentAttrs: string[];
      if (isOAuthMode) {
        currentAttrs = activeTokenType === 'access' ? currentAccessTokenAttributes : currentIdTokenAttributes;
      } else {
        currentAttrs = sharedUserAttributes;
      }

      const stillPending = Array.from(pendingRemovals).filter((attr) => currentAttrs.includes(attr));

      if (stillPending.length !== pendingRemovals.size) {
        setPendingRemovals(new Set(stillPending));
        // Clear highlights when removals are fully applied
        if (stillPending.length === 0 && pendingAdditions.size === 0) {
          setTimeout(() => setHighlightedAttributes(new Set()), 500);
        }
      }
    }
  }, [
    currentAccessTokenAttributes,
    currentIdTokenAttributes,
    sharedUserAttributes,
    isOAuthMode,
    activeTokenType,
    pendingAdditions,
    pendingRemovals,
  ]);

  // Handle attribute click
  const handleAttributeClick = (attr: string, tokenType: 'shared' | 'access' | 'id') => {
    if (tokenType !== 'shared') {
      setActiveTokenType(tokenType);
    }

    let currentAttributes: string[];
    if (tokenType === 'shared') {
      currentAttributes = sharedUserAttributes;
    } else if (tokenType === 'access') {
      currentAttributes = currentAccessTokenAttributes;
    } else {
      currentAttributes = currentIdTokenAttributes;
    }

    const isAdded = currentAttributes.includes(attr);
    const isPendingAddition = pendingAdditions.has(attr) && (tokenType === 'shared' || activeTokenType === tokenType);
    const isPendingRemoval = pendingRemovals.has(attr) && (tokenType === 'shared' || activeTokenType === tokenType);

    setHighlightedAttributes((prev) => new Set([...prev, attr]));
    const currentlyActive = (isAdded && !isPendingRemoval) || isPendingAddition;

    if (currentlyActive) {
      if (isPendingAddition) {
        setPendingAdditions((prev) => {
          const newSet = new Set(prev);
          newSet.delete(attr);
          return newSet;
        });
      } else if (isAdded) {
        setPendingRemovals((prev) => new Set([...prev, attr]));
      }
    } else if (isPendingRemoval) {
      setPendingRemovals((prev) => {
        const newSet = new Set(prev);
        newSet.delete(attr);
        return newSet;
      });
    } else {
      setPendingAdditions((prev) => new Set([...prev, attr]));
    }
  };

  return (
    <Stack spacing={3}>
      {/* OAuth/OIDC Mode */}
      {isOAuthMode ? (
        <>
          {/* Token Issuer - Common for both tokens */}
          <TokenIssuerSection control={control} errors={errors} />

          {/* Access Token User Attributes */}
          <TokenUserAttributesSection
            tokenType="access"
            currentAttributes={currentAccessTokenAttributes}
            userAttributes={userAttributes}
            isLoadingUserAttributes={isLoadingUserAttributes}
            expandedSections={expandedSections}
            setExpandedSections={setExpandedSections}
            pendingAdditions={pendingAdditions}
            pendingRemovals={pendingRemovals}
            highlightedAttributes={highlightedAttributes}
            onAttributeClick={handleAttributeClick}
            activeTokenType={activeTokenType}
            oauth2Config={oauth2Config}
          />

          {/* Access Token Validation */}
          <TokenValidationSection control={control} errors={errors} tokenType="access" />

          {/* ID Token User Attributes */}
          <TokenUserAttributesSection
            tokenType="id"
            currentAttributes={currentIdTokenAttributes}
            userAttributes={userAttributes}
            isLoadingUserAttributes={isLoadingUserAttributes}
            expandedSections={expandedSections}
            setExpandedSections={setExpandedSections}
            pendingAdditions={pendingAdditions}
            pendingRemovals={pendingRemovals}
            highlightedAttributes={highlightedAttributes}
            onAttributeClick={handleAttributeClick}
            activeTokenType={activeTokenType}
            oauth2Config={oauth2Config}
          />

          {/* ID Token Validation */}
          <TokenValidationSection control={control} errors={errors} tokenType="id" />
        </>
      ) : (
        <>
          {/* Native Flow Mode */}
          <TokenUserAttributesSection
            tokenType="shared"
            currentAttributes={sharedUserAttributes}
            userAttributes={userAttributes}
            isLoadingUserAttributes={isLoadingUserAttributes}
            expandedSections={expandedSections}
            setExpandedSections={setExpandedSections}
            pendingAdditions={pendingAdditions}
            pendingRemovals={pendingRemovals}
            highlightedAttributes={highlightedAttributes}
            onAttributeClick={handleAttributeClick}
            activeTokenType={activeTokenType}
          />

          {/* Token Validation with Issuer */}
          <TokenValidationSection control={control} errors={errors} tokenType="shared" />
          <TokenIssuerSection control={control} errors={errors} />
        </>
      )}
    </Stack>
  );
}
