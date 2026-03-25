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
import type {OAuth2Config, ScopeClaims} from '../../../models/oauth';
import type {Application} from '../../../models/application';
import type {PropertyDefinition, ApiUserSchema} from '../../../../user-types/types/user-types';
import TokenUserAttributesSection from './TokenUserAttributesSection';
import TokenValidationSection from './TokenValidationSection';
import ScopeSection from './ScopeSection';

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
  });

type TokenConfigFormData = z.infer<ReturnType<typeof createTokenConfigSchema>>;

type TokenAttributeScope = 'shared' | 'access' | 'id' | 'userinfo';
type OAuthTokenAttributeScope = Exclude<TokenAttributeScope, 'shared'>;

const OAUTH_TOKEN_SCOPES: OAuthTokenAttributeScope[] = ['access', 'id', 'userinfo'];

const createEmptyAttributeSetState = (): Record<TokenAttributeScope, Set<string>> => ({
  shared: new Set(),
  access: new Set(),
  id: new Set(),
  userinfo: new Set(),
});

const areAttributesEqual = (arr1: string[], arr2: string[]): boolean => {
  if (arr1.length !== arr2.length) return false;
  const sorted1 = [...arr1].sort();
  const sorted2 = [...arr2].sort();
  return sorted1.every((val, index) => val === sorted2[index]);
};

const areSetsEqual = (set1: Set<string>, set2: Set<string>): boolean => {
  if (set1.size !== set2.size) return false;

  return Array.from(set1).every((value) => set2.has(value));
};

/**
 * Container component for token configuration settings.
 *
 * Manages token settings for both OAuth2/OIDC mode and Native mode:
 * - OAuth2/OIDC mode: Separate access token and ID token configurations
 * - Native mode: Shared token configuration
 *
 * Provides sections for:
 * - Token validity periods (with real-time validation)
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

  const applyTimeoutRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  const [userSchemas, setUserSchemas] = useState<ApiUserSchema[]>([]);

  const {data: userTypesData, isLoading: userTypesLoading} = useGetUserTypes();
  const [activeTokenType, setActiveTokenType] = useState<'access' | 'id' | 'userinfo'>('access');
  const [pendingAdditionsByToken, setPendingAdditionsByToken] = useState<Record<TokenAttributeScope, Set<string>>>(() =>
    createEmptyAttributeSetState(),
  );
  const [pendingRemovalsByToken, setPendingRemovalsByToken] = useState<Record<TokenAttributeScope, Set<string>>>(() =>
    createEmptyAttributeSetState(),
  );
  const [highlightedAttributesByToken, setHighlightedAttributesByToken] = useState<
    Record<TokenAttributeScope, Set<string>>
  >(() => createEmptyAttributeSetState());

  // Stabilize allowedUserTypes array reference
  const allowedUserTypes = useMemo(() => application.allowedUserTypes ?? [], [application.allowedUserTypes]);

  // Get schema IDs for allowed user types
  const schemaIds = useMemo(() => {
    if (!userTypesData?.schemas || allowedUserTypes.length === 0) {
      return [];
    }

    return userTypesData.schemas.filter((schema) => allowedUserTypes.includes(schema.name)).map((schema) => schema.id);
  }, [userTypesData, allowedUserTypes]);

  // Determine if this is OAuth/OIDC mode (has separate token configs) or Native mode
  const isOAuthMode = useMemo(
    () => oauth2Config?.token?.accessToken !== undefined || oauth2Config?.token?.idToken !== undefined,
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
      validityPeriod: oauth2Config?.token?.validityPeriod ?? application.assertion?.validityPeriod ?? 3600,
      accessTokenValidity: oauth2Config?.token?.accessToken?.validityPeriod ?? 3600,
      idTokenValidity: oauth2Config?.token?.idToken?.validityPeriod ?? 3600,
    },
  });

  const validityPeriod = watch('validityPeriod');
  const accessTokenValidity = watch('accessTokenValidity');
  const idTokenValidity = watch('idTokenValidity');

  /**
   * Sync form values when the OAuth2 configuration or application token configuration changes.
   */
  useEffect(() => {
    if (isOAuthMode) {
      setValue('accessTokenValidity', oauth2Config?.token?.accessToken?.validityPeriod ?? 3600);
      setValue('idTokenValidity', oauth2Config?.token?.idToken?.validityPeriod ?? 3600);
    } else {
      setValue('validityPeriod', oauth2Config?.token?.validityPeriod ?? application.assertion?.validityPeriod ?? 3600);
    }
  }, [isOAuthMode, oauth2Config, application.assertion?.validityPeriod, setValue]);

  /**
   * Effect to sync form changes back to the parent component.
   */
  useEffect(() => {
    if (isOAuthMode && oauth2Config) {
      const currentAccessTokenValidity = oauth2Config.token?.accessToken?.validityPeriod ?? 3600;
      const currentIdTokenValidity = oauth2Config.token?.idToken?.validityPeriod ?? 3600;

      if (accessTokenValidity === currentAccessTokenValidity && idTokenValidity === currentIdTokenValidity) {
        return;
      }

      // OAuth mode: update separate access and ID token configs
      const updatedConfig = {
        ...oauth2Config,
        token: {
          ...oauth2Config.token,
          accessToken: {
            ...oauth2Config.token?.accessToken,
            validityPeriod: accessTokenValidity,
          },
          idToken: {
            ...oauth2Config.token?.idToken,
            validityPeriod: idTokenValidity,
          },
        },
      };

      const updatedInboundAuth = application.inboundAuthConfig?.map((config) => {
        if (config.type === 'oauth2') {
          return {...config, config: updatedConfig};
        }

        return config;
      });

      onFieldChange('inboundAuthConfig', updatedInboundAuth);
    } else if (!isOAuthMode) {
      const currentValidityPeriod =
        oauth2Config?.token?.validityPeriod ?? application.assertion?.validityPeriod ?? 3600;

      if (validityPeriod === currentValidityPeriod) {
        return;
      }

      // Native mode: update root-level assertion config
      const updatedAssertion = {
        ...application.assertion,
        validityPeriod,
      };

      onFieldChange('assertion', updatedAssertion);
    }
  }, [
    validityPeriod,
    accessTokenValidity,
    idTokenValidity,
    isOAuthMode,
    oauth2Config,
    application.inboundAuthConfig,
    application.assertion,
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

    return oauth2Config?.token?.userAttributes ?? application.assertion?.userAttributes ?? [];
  }, [isOAuthMode, oauth2Config, application]);

  const [isUserInfoCustomAttributes, setIsUserInfoCustomAttributes] = useState<boolean>(false);
  const [currentUserInfoAttributes, setCurrentUserInfoAttributes] = useState<string[]>([]);

  const currentAccessTokenAttributes = useMemo(
    () => oauth2Config?.token?.accessToken?.userAttributes ?? [],
    [oauth2Config],
  );

  const currentIdTokenAttributes = useMemo(() => oauth2Config?.token?.idToken?.userAttributes ?? [], [oauth2Config]);

  // Initialize userinfoEnabled based on config presence and difference from ID token
  useEffect(() => {
    if (!isOAuthMode || !oauth2Config) return;

    const idTokenAttrs = oauth2Config.token?.idToken?.userAttributes ?? [];

    const userInfoConfig = oauth2Config.userInfo;

    if (userInfoConfig) {
      const userInfoAttrs = userInfoConfig.userAttributes || [];
      const idTokenAttrsRef = idTokenAttrs || [];
      setCurrentUserInfoAttributes(userInfoAttrs);
      // Enable toggle only if attributes differ from ID token attributes
      const isDifferent = !areAttributesEqual(userInfoAttrs, idTokenAttrsRef);
      setIsUserInfoCustomAttributes(isDifferent);
    } else {
      // If userInfo is undefined, fallback logic applies, so toggle is OFF
      setIsUserInfoCustomAttributes(false);
      setCurrentUserInfoAttributes(idTokenAttrs);
    }
  }, [isOAuthMode, oauth2Config]); // Run when config structure changes

  const handleToggleUserInfo = (checked: boolean) => {
    setIsUserInfoCustomAttributes(checked);

    if (!checked && activeTokenType === 'userinfo') {
      // Cancel any pending changes when disabling explicit configuration
      if (applyTimeoutRef.current) {
        clearTimeout(applyTimeoutRef.current);
        applyTimeoutRef.current = null;
      }
      setPendingAdditionsByToken((prev) => ({...prev, userinfo: new Set()}));
      setPendingRemovalsByToken((prev) => ({...prev, userinfo: new Set()}));
      setHighlightedAttributesByToken((prev) => ({...prev, userinfo: new Set()}));
    }

    if (checked) {
      // When enabling, start with ID token attributes if current UserInfo attrs are empty/undefined
      if (!oauth2Config?.userInfo) {
        setCurrentUserInfoAttributes([...currentIdTokenAttributes]);

        // Update config immediately to initialize the structure
        const updatedConfig = {
          ...oauth2Config,
          userInfo: {
            userAttributes: [...currentIdTokenAttributes],
          },
        };

        const updatedInboundAuth = application.inboundAuthConfig?.map((config) => {
          if (config.type === 'oauth2') {
            return {...config, config: updatedConfig};
          }
          return config;
        });
        onFieldChange('inboundAuthConfig', updatedInboundAuth);
      }
    } else if (oauth2Config) {
      // When disabling, remove userInfo from config to use fallback
      const {userInfo, ...restConfig} = oauth2Config;
      const updatedConfig = restConfig;

      const updatedInboundAuth = application.inboundAuthConfig?.map((config) => {
        if (config.type === 'oauth2') {
          return {...config, config: updatedConfig};
        }
        return config;
      });
      onFieldChange('inboundAuthConfig', updatedInboundAuth);
    }
  };

  const handleScopesChange = (newScopes: string[]) => {
    const updatedConfig = {...oauth2Config, scopes: newScopes};
    const updatedInboundAuth = application.inboundAuthConfig?.map((config) => {
      if (config.type === 'oauth2') {
        return {...config, config: updatedConfig};
      }
      return config;
    });
    onFieldChange('inboundAuthConfig', updatedInboundAuth);
  };

  const handleScopeClaimsChange = (newScopeClaims: ScopeClaims) => {
    const updatedConfig = {
      ...oauth2Config,
      scopeClaims: newScopeClaims,
    };
    const updatedInboundAuth = application.inboundAuthConfig?.map((config) => {
      if (config.type === 'oauth2') {
        return {...config, config: updatedConfig};
      }

      return config;
    });
    onFieldChange('inboundAuthConfig', updatedInboundAuth);
  };

  /**
   * Effect to apply pending additions and removals after a debounce period.
   */
  useEffect(() => {
    const hasPendingChanges =
      Object.values(pendingAdditionsByToken).some((set) => set.size > 0) ||
      Object.values(pendingRemovalsByToken).some((set) => set.size > 0);

    if (!hasPendingChanges) {
      return undefined;
    }

    // Clear existing timeout
    if (applyTimeoutRef.current) {
      clearTimeout(applyTimeoutRef.current);
    }

    // Set new timeout to apply changes
    applyTimeoutRef.current = setTimeout(() => {
      if (isOAuthMode && oauth2Config) {
        let updatedConfig = oauth2Config;
        let hasConfigChanges = false;

        const currentAttributesByToken: Record<OAuthTokenAttributeScope, string[]> = {
          access: currentAccessTokenAttributes,
          id: currentIdTokenAttributes,
          userinfo: currentUserInfoAttributes,
        };

        OAUTH_TOKEN_SCOPES.forEach((tokenType) => {
          const additionsArray = Array.from(pendingAdditionsByToken[tokenType]);
          const removalsArray = Array.from(pendingRemovalsByToken[tokenType]);

          if (additionsArray.length === 0 && removalsArray.length === 0) {
            return;
          }

          const currentAttrs = currentAttributesByToken[tokenType];
          const nextAttrs = [...currentAttrs, ...additionsArray.filter((attr) => !currentAttrs.includes(attr))].filter(
            (attr) => !removalsArray.includes(attr),
          );

          if (areAttributesEqual(nextAttrs, currentAttrs)) {
            return;
          }

          hasConfigChanges = true;

          if (tokenType === 'access') {
            const currentAccessConfig = updatedConfig.token?.accessToken ?? {
              validityPeriod: accessTokenValidity,
              userAttributes: currentAccessTokenAttributes,
            };
            const currentIdConfig = updatedConfig.token?.idToken ?? {
              validityPeriod: idTokenValidity,
              userAttributes: currentIdTokenAttributes,
            };

            updatedConfig = {
              ...updatedConfig,
              token: {
                ...updatedConfig.token,
                accessToken: {
                  ...currentAccessConfig,
                  userAttributes: nextAttrs,
                },
                idToken: currentIdConfig,
              },
            };
          } else if (tokenType === 'id') {
            const currentAccessConfig = updatedConfig.token?.accessToken ?? {
              validityPeriod: accessTokenValidity,
              userAttributes: currentAccessTokenAttributes,
            };
            const currentIdConfig = updatedConfig.token?.idToken ?? {
              validityPeriod: idTokenValidity,
              userAttributes: currentIdTokenAttributes,
            };

            updatedConfig = {
              ...updatedConfig,
              token: {
                ...updatedConfig.token,
                accessToken: currentAccessConfig,
                idToken: {
                  ...currentIdConfig,
                  userAttributes: nextAttrs,
                },
              },
            };
          } else {
            updatedConfig = {
              ...updatedConfig,
              userInfo: {
                userAttributes: nextAttrs,
              },
            };
          }
        });

        if (hasConfigChanges) {
          const updatedInboundAuth = application.inboundAuthConfig?.map((config) => {
            if (config.type === 'oauth2') {
              return {...config, config: updatedConfig};
            }

            return config;
          });

          onFieldChange('inboundAuthConfig', updatedInboundAuth);
        }

        return;
      }

      const sharedAdditions = Array.from(pendingAdditionsByToken.shared);
      const sharedRemovals = Array.from(pendingRemovalsByToken.shared);

      if (sharedAdditions.length === 0 && sharedRemovals.length === 0) {
        return;
      }

      const nextSharedAttributes = [
        ...sharedUserAttributes,
        ...sharedAdditions.filter((attr) => !sharedUserAttributes.includes(attr)),
      ].filter((attr) => !sharedRemovals.includes(attr));

      if (areAttributesEqual(nextSharedAttributes, sharedUserAttributes)) {
        return;
      }

      const updatedAssertion = {
        ...application.assertion,
        userAttributes: nextSharedAttributes,
      };
      onFieldChange('assertion', updatedAssertion);

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
    pendingAdditionsByToken,
    pendingRemovalsByToken,
    isOAuthMode,
    oauth2Config,
    accessTokenValidity,
    idTokenValidity,
    currentAccessTokenAttributes,
    currentIdTokenAttributes,
    currentUserInfoAttributes,
    sharedUserAttributes,
    application.inboundAuthConfig,
    application.assertion,
    onFieldChange,
  ]);

  // Clean up pending additions/removals once they're reflected in the actual config
  useEffect(() => {
    const currentAttributesByToken: Record<TokenAttributeScope, string[]> = {
      shared: sharedUserAttributes,
      access: currentAccessTokenAttributes,
      id: currentIdTokenAttributes,
      userinfo: currentUserInfoAttributes,
    };

    const allScopes: TokenAttributeScope[] = ['shared', 'access', 'id', 'userinfo'];

    setPendingAdditionsByToken((prev) => {
      let hasUpdates = false;
      const next = {...prev};

      allScopes.forEach((scope) => {
        const remaining = new Set(
          Array.from(prev[scope]).filter((attr) => !currentAttributesByToken[scope].includes(attr)),
        );

        if (!areSetsEqual(prev[scope], remaining)) {
          next[scope] = remaining;
          hasUpdates = true;
        }
      });

      return hasUpdates ? next : prev;
    });

    setPendingRemovalsByToken((prev) => {
      let hasUpdates = false;
      const next = {...prev};
      const clearedScopes: TokenAttributeScope[] = [];

      allScopes.forEach((scope) => {
        const remaining = new Set(
          Array.from(prev[scope]).filter((attr) => currentAttributesByToken[scope].includes(attr)),
        );

        if (!areSetsEqual(prev[scope], remaining)) {
          next[scope] = remaining;
          hasUpdates = true;
        }

        if (prev[scope].size > 0 && remaining.size === 0 && pendingAdditionsByToken[scope].size === 0) {
          clearedScopes.push(scope);
        }
      });

      if (clearedScopes.length > 0) {
        setTimeout(() => {
          setHighlightedAttributesByToken((prevHighlights) => {
            let hasHighlightUpdates = false;
            const nextHighlights = {...prevHighlights};

            clearedScopes.forEach((scope) => {
              if (nextHighlights[scope].size > 0) {
                nextHighlights[scope] = new Set();
                hasHighlightUpdates = true;
              }
            });

            return hasHighlightUpdates ? nextHighlights : prevHighlights;
          });
        }, 500);
      }

      return hasUpdates ? next : prev;
    });
  }, [
    currentAccessTokenAttributes,
    currentIdTokenAttributes,
    currentUserInfoAttributes,
    sharedUserAttributes,
    pendingAdditionsByToken,
  ]);

  // Handle attribute click
  const handleAttributeClick = (attr: string, tokenType: 'shared' | 'access' | 'id' | 'userinfo') => {
    let currentAttributes: string[];
    if (tokenType === 'shared') {
      currentAttributes = sharedUserAttributes;
    } else if (tokenType === 'access') {
      currentAttributes = currentAccessTokenAttributes;
    } else if (tokenType === 'id') {
      currentAttributes = currentIdTokenAttributes;
    } else {
      currentAttributes = currentUserInfoAttributes;
    }

    const isAdded = currentAttributes.includes(attr);
    const tokenPendingAdditions = pendingAdditionsByToken[tokenType];
    const tokenPendingRemovals = pendingRemovalsByToken[tokenType];
    const isPendingAddition = tokenPendingAdditions.has(attr);
    const isPendingRemoval = tokenPendingRemovals.has(attr);

    setHighlightedAttributesByToken((prev) => ({
      ...prev,
      [tokenType]: new Set([...prev[tokenType], attr]),
    }));

    const currentlyActive = (isAdded && !isPendingRemoval) || isPendingAddition;

    if (currentlyActive) {
      if (isPendingAddition) {
        setPendingAdditionsByToken((prev) => {
          const newSet = new Set(prev[tokenType]);
          newSet.delete(attr);
          return {...prev, [tokenType]: newSet};
        });
      } else if (isAdded) {
        setPendingRemovalsByToken((prev) => ({
          ...prev,
          [tokenType]: new Set([...prev[tokenType], attr]),
        }));
      }
    } else if (isPendingRemoval) {
      setPendingRemovalsByToken((prev) => {
        const newSet = new Set(prev[tokenType]);
        newSet.delete(attr);
        return {...prev, [tokenType]: newSet};
      });
    } else {
      setPendingAdditionsByToken((prev) => ({
        ...prev,
        [tokenType]: new Set([...prev[tokenType], attr]),
      }));
    }
  };

  const visibleScope: TokenAttributeScope = isOAuthMode ? activeTokenType : 'shared';
  const visiblePendingAdditions = pendingAdditionsByToken[visibleScope];
  const visiblePendingRemovals = pendingRemovalsByToken[visibleScope];
  const visibleHighlightedAttributes = highlightedAttributesByToken[visibleScope];

  return (
    <Stack spacing={3}>
      {/* OAuth/OIDC Mode */}
      {isOAuthMode ? (
        <>
          {/* Merged User Attributes (Access Token / ID Token / User Info tabs) */}
          <TokenUserAttributesSection
            accessTokenAttributes={currentAccessTokenAttributes}
            idTokenAttributes={currentIdTokenAttributes}
            userInfoAttributes={currentUserInfoAttributes}
            activeTab={activeTokenType}
            onTabChange={setActiveTokenType}
            isUserInfoCustomAttributes={isUserInfoCustomAttributes}
            onToggleUserInfo={handleToggleUserInfo}
            userAttributes={userAttributes}
            isLoadingUserAttributes={isLoadingUserAttributes}
            pendingAdditions={visiblePendingAdditions}
            pendingRemovals={visiblePendingRemovals}
            highlightedAttributes={visibleHighlightedAttributes}
            onAttributeClick={handleAttributeClick}
          />

          {/* Scopes & Attribute Mapping */}
          <ScopeSection
            scopes={oauth2Config?.scopes ?? []}
            scopeClaims={oauth2Config?.scopeClaims ?? {}}
            userAttributes={userAttributes}
            isLoadingUserAttributes={isLoadingUserAttributes}
            onScopesChange={handleScopesChange}
            onScopeClaimsChange={handleScopeClaimsChange}
          />

          {/* Merged Token Validation (Access Token / ID Token tabs) */}
          <TokenValidationSection control={control} errors={errors} tokenType="oauth" />
        </>
      ) : (
        <>
          {/* Native Flow Mode */}
          <TokenUserAttributesSection
            sharedAttributes={sharedUserAttributes}
            userAttributes={userAttributes}
            isLoadingUserAttributes={isLoadingUserAttributes}
            pendingAdditions={visiblePendingAdditions}
            pendingRemovals={visiblePendingRemovals}
            highlightedAttributes={visibleHighlightedAttributes}
            onAttributeClick={handleAttributeClick}
          />

          {/* Token Validation */}
          <TokenValidationSection control={control} errors={errors} tokenType="shared" />
        </>
      )}
    </Stack>
  );
}
