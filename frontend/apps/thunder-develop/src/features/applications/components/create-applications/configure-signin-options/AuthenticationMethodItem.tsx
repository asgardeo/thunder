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

import {ListItem, ListItemButton, ListItemIcon, ListItemText, Switch} from '@wso2/oxygen-ui';
import type {JSX, ReactNode} from 'react';
import {useTranslation} from 'react-i18next';

/**
 * Props for the AuthenticationMethodItem component
 */
export interface AuthenticationMethodItemProps {
  /**
   * Unique identifier for the authentication method
   */
  id: string;

  /**
   * Display name for the authentication method
   */
  name: string;

  /**
   * Secondary text (optional)
   */
  secondary?: string;

  /**
   * Icon for the authentication method
   */
  icon: ReactNode;

  /**
   * Whether this method is currently enabled
   */
  isEnabled: boolean;

  /**
   * Whether this method is available (affects disabled state)
   */
  isAvailable: boolean;

  /**
   * Whether this method should be disabled
   */
  isDisabled?: boolean;

  /**
   * Callback when the method is toggled
   */
  onToggle: (id: string) => void;
}

/**
 * Individual authentication method list item component
 */
export default function AuthenticationMethodItem({
  id,
  name,
  secondary = undefined,
  icon,
  isEnabled,
  isAvailable,
  isDisabled = false,
  onToggle,
}: AuthenticationMethodItemProps): JSX.Element {
  const {t} = useTranslation();
  const handleToggle = () => onToggle(id);

  if (!isAvailable) {
    return (
      <ListItem disablePadding>
        <ListItemButton disabled>
          <ListItemIcon>{icon}</ListItemIcon>
          <ListItemText primary={name} secondary={t('applications:onboarding.configure.SignInOptions.notConfigured')} />
        </ListItemButton>
      </ListItem>
    );
  }

  return (
    <ListItem
      disablePadding
      secondaryAction={
        <Switch edge="end" checked={isEnabled} onChange={handleToggle} disabled={isDisabled} color="primary" />
      }
    >
      <ListItemButton onClick={handleToggle} disabled={isDisabled}>
        <ListItemIcon>{icon}</ListItemIcon>
        <ListItemText primary={name} secondary={secondary} />
      </ListItemButton>
    </ListItem>
  );
}
