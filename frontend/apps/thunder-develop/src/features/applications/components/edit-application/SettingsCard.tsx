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

import {Box, Paper, Stack, Switch, Typography} from '@wso2/oxygen-ui';
import type {ReactNode} from 'react';

interface SettingsCardProps {
  /**
   * Card title
   */
  title: string;
  /**
   * Optional description text shown below the title
   */
  description?: string;
  /**
   * Content of the card
   */
  children: ReactNode;
  /**
   * Optional toggle switch state
   */
  enabled?: boolean;
  /**
   * Optional toggle change handler
   */
  onToggle?: (enabled: boolean) => void;
}

/**
 * Reusable settings card component for application edit pages.
 * Provides consistent styling with optional enable/disable toggle.
 *
 * @example
 * ```tsx
 * <SettingsCard
 *   title="Quick Copy"
 *   description="Copy application credentials"
 * >
 *   <TextField label="Application ID" />
 * </SettingsCard>
 * ```
 *
 * @example With toggle
 * ```tsx
 * <SettingsCard
 *   title="Registration Flow"
 *   description="Allow users to register"
 *   enabled={isEnabled}
 *   onToggle={(enabled) => handleToggle(enabled)}
 * >
 *   <TextField label="Flow ID" />
 * </SettingsCard>
 * ```
 */
export default function SettingsCard({
  title,
  description = undefined,
  children,
  enabled = undefined,
  onToggle = undefined,
}: SettingsCardProps) {
  const hasToggle = enabled !== undefined && onToggle !== undefined;

  return (
    <Paper>
      <Box sx={{p: 3}}>
        <Stack direction="row" alignItems="center" justifyContent="space-between" spacing={2}>
          <Typography variant="h5">{title}</Typography>
          {hasToggle && (
            <Switch
              checked={enabled}
              onChange={(e) => onToggle(e.target.checked)}
              inputProps={{'aria-label': `Toggle ${title}`}}
            />
          )}
        </Stack>
        {description && (
          <Typography variant="body2" sx={{mt: 0.5, color: 'text.disabled'}}>
            {description}
          </Typography>
        )}
      </Box>
      <Paper sx={{p: 3}}>
        {/* Only show content if toggle is enabled or if there's no toggle */}
        {(!hasToggle || enabled) && children}
      </Paper>
    </Paper>
  );
}
