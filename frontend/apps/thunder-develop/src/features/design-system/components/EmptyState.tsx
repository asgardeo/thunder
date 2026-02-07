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

import type {JSX, ReactNode} from 'react';
import {Box, Button, Stack, Typography} from '@mui/material';

/**
 * Action configuration for the empty state.
 */
interface EmptyStateAction {
  /** Button label */
  label: string;
  /** Click handler */
  onClick: () => void;
  /** Optional icon to show before label */
  icon?: ReactNode;
}

/**
 * Props for the EmptyState component.
 */
export interface EmptyStateProps {
  /** Icon to display above the title */
  icon?: ReactNode;
  /** Main title text */
  title: string;
  /** Optional description text */
  description?: string;
  /** Optional primary action button */
  action?: EmptyStateAction;
}

/**
 * Empty state component to show when no content is available.
 * Displays an icon, title, description, and optional action button.
 *
 * @param props - Component props
 * @returns EmptyState component
 *
 * @example
 * ```tsx
 * // No items state
 * <EmptyState
 *   icon={<Palette size={48} />}
 *   title="No themes yet"
 *   description="Create your first theme to get started"
 *   action={{
 *     label: "Create Theme",
 *     icon: <Plus size={20} />,
 *     onClick: () => navigate('/themes/builder')
 *   }}
 * />
 *
 * // No search results
 * <EmptyState
 *   icon={<Search size={48} />}
 *   title="No results found"
 *   description="Try a different search term"
 * />
 * ```
 */
export default function EmptyState({icon, title, description, action}: EmptyStateProps): JSX.Element {
  return (
    <Box
      sx={{
        display: 'flex',
        flexDirection: 'column',
        alignItems: 'center',
        justifyContent: 'center',
        minHeight: 400,
        textAlign: 'center',
        p: 4,
      }}
    >
      <Stack spacing={3} alignItems="center" maxWidth={480}>
        {/* Icon */}
        {icon && (
          <Box
            sx={{
              color: 'text.secondary',
              opacity: 0.5,
            }}
          >
            {icon}
          </Box>
        )}

        {/* Title */}
        <Typography variant="h5" color="text.primary">
          {title}
        </Typography>

        {/* Description */}
        {description && (
          <Typography variant="body1" color="text.secondary">
            {description}
          </Typography>
        )}

        {/* Action button */}
        {action && (
          <Button variant="contained" startIcon={action.icon} onClick={action.onClick} size="large">
            {action.label}
          </Button>
        )}
      </Stack>
    </Box>
  );
}
