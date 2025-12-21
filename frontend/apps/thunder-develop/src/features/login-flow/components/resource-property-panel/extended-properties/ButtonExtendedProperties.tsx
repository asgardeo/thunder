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

import {useMemo, type ReactNode} from 'react';
import {useTranslation} from 'react-i18next';
import classNames from 'classnames';
import isEqual from 'lodash-es/isEqual';
import omit from 'lodash-es/omit';
import {Avatar, Box, Card, Divider, FormHelperText, Grid, Stack, Typography, useColorScheme} from '@wso2/oxygen-ui';
import type {CommonResourcePropertiesPropsInterface} from '@/features/flows/components/resource-property-panel/ResourceProperties';
import useFlowBuilderCore from '@/features/flows/hooks/useFlowBuilderCore';
import useValidationStatus from '@/features/flows/hooks/useValidationStatus';
import {type Element} from '@/features/flows/models/elements';
import resolveStaticResourcePath from '@/features/flows/utils/resolveStaticResourcePath';
import useGetLoginFlowBuilderActions from '@/features/login-flow/api/useGetLoginFlowBuilderActions';
import './ButtonExtendedProperties.scss';

/**
 * Props interface of {@link ButtonExtendedProperties}
 */
export type ButtonExtendedPropertiesPropsInterface = CommonResourcePropertiesPropsInterface;
/**
 * Extended properties for the field elements.
 *
 * @param props - Props injected to the component.
 * @returns The ButtonExtendedProperties component.
 */
function ButtonExtendedProperties({resource, onChange}: ButtonExtendedPropertiesPropsInterface): ReactNode {
  const {t} = useTranslation();
  const {data: actions} = useGetLoginFlowBuilderActions();
  const {lastInteractedResource, setLastInteractedResource} = useFlowBuilderCore();
  const {selectedNotification} = useValidationStatus();
  const {mode, systemMode} = useColorScheme();

  // Determine the effective mode - if mode is 'system', use systemMode
  const effectiveMode = mode === 'system' ? systemMode : mode;

  /**
   * Get the error message for the identifier field.
   */
  const errorMessage: string = useMemo(() => {
    const key = `${resource?.id}_action`;

    if (selectedNotification?.hasResourceFieldNotification(key)) {
      return selectedNotification?.getResourceFieldNotification(key);
    }

    return '';
  }, [resource, selectedNotification]);

  return (
    <Stack className="button-extended-properties" gap={2}>
      <Divider sx={{marginY: 2}} />
      <div>
        <Typography className="button-extended-properties-heading">{t('flows:core.buttonExtendedProperties.type')}</Typography>
        {actions?.map((action: Partial<Element> & {types?: Element[]}) => (
          <Box key={`${action.type ?? 'action'}-${action.id ?? action.display?.label ?? ''}`}>
            <Typography className="button-extended-properties-sub-heading" variant="body1">
              {action?.display?.label}
            </Typography>
            <Grid container spacing={1}>
              {action.types?.map((actionType: Element) => (
                <Grid
                  key={`${actionType.type}-${actionType.id}`}
                  size={{xs: 12}}
                  onClick={() => {
                    onChange(
                      'action',
                      {
                        ...actionType.action,
                        ...((resource as Element)?.action?.next ? {next: (resource as Element)?.action?.next} : {}),
                      },
                      resource,
                    );

                    setLastInteractedResource({
                      ...lastInteractedResource,
                      action: actionType.action,
                    });
                  }}
                >
                  <Card
                    className={classNames('extended-property action-type', {
                      error: !!errorMessage,
                      selected: isEqual(omit((lastInteractedResource as Element)?.action, 'next'), actionType.action),
                    })}
                  >
                    <Box display="flex" flexDirection="row" gap={1} padding={1} alignItems="center">
                      <Avatar
                        className="action-type-icon"
                        src={resolveStaticResourcePath(actionType?.display?.image)}
                        variant="rounded"
                        sx={{
                          '& .MuiAvatar-img': {
                            filter: effectiveMode === 'dark' ? 'brightness(0.9) invert(1)' : 'none',
                          },
                        }}
                      />
                      <Typography variant="body2" className="action-type-name">
                        {actionType?.display?.label}
                      </Typography>
                    </Box>
                  </Card>
                </Grid>
              ))}
            </Grid>
          </Box>
        ))}
        {errorMessage && <FormHelperText error>{errorMessage}</FormHelperText>}
      </div>
      <Divider sx={{marginY: 2}} />
    </Stack>
  );
}

export default ButtonExtendedProperties;
