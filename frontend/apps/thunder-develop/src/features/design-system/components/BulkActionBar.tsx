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
import {Button, Paper, Stack, Typography} from '@mui/material';

/**
 * Configuration for a bulk action button.
 */
export interface BulkAction {
  /** Button label */
  label: string;
  /** Icon to display before label */
  icon: ReactNode;
  /** Click handler */
  onClick: () => void;
  /** Whether the button is disabled */
  disabled?: boolean;
  /** Button color */
  color?: 'primary' | 'error' | 'secondary';
}

/**
 * Props for the BulkActionBar component.
 */
export interface BulkActionBarProps {
  /** Number of selected items */
  selectedCount: number;
  /** Callback to clear all selections */
  onClearSelection: () => void;
  /** Array of action buttons to display */
  actions: BulkAction[];
}

/**
 * Sticky action bar that appears when items are selected.
 * Shows selected count and action buttons for bulk operations.
 *
 * @param props - Component props
 * @returns BulkActionBar component
 *
 * @example
 * ```tsx
 * const { selectedIds, clearSelection } = useMultiSelect();
 *
 * {selectedIds.size > 0 && (
 *   <BulkActionBar
 *     selectedCount={selectedIds.size}
 *     onClearSelection={clearSelection}
 *     actions={[
 *       {
 *         label: 'Delete',
 *         icon: <Trash2 size={20} />,
 *         onClick: handleBulkDelete,
 *         color: 'error'
 *       },
 *       {
 *         label: 'Duplicate',
 *         icon: <Copy size={20} />,
 *         onClick: handleBulkDuplicate,
 *       }
 *     ]}
 *   />
 * )}
 * ```
 */
export default function BulkActionBar({selectedCount, onClearSelection, actions}: BulkActionBarProps): JSX.Element {
  return (
    <Paper
      elevation={8}
      sx={{
        position: 'fixed',
        bottom: 24,
        left: '50%',
        transform: 'translateX(-50%)',
        zIndex: 1000,
        px: 3,
        py: 2,
        minWidth: 400,
        animation: 'slideUp 0.3s ease-out',
        '@keyframes slideUp': {
          from: {
            opacity: 0,
            transform: 'translate(-50%, 20px)',
          },
          to: {
            opacity: 1,
            transform: 'translate(-50%, 0)',
          },
        },
      }}
    >
      <Stack direction="row" alignItems="center" spacing={3}>
        {/* Selected count */}
        <Typography variant="body1" fontWeight={600} color="text.primary">
          {selectedCount} selected
        </Typography>

        {/* Action buttons */}
        <Stack direction="row" spacing={1} flex={1} justifyContent="flex-end">
          {actions.map((action, index) => (
            <Button
              key={index}
              variant={action.color === 'error' ? 'outlined' : 'text'}
              color={action.color || 'primary'}
              startIcon={action.icon}
              onClick={action.onClick}
              disabled={action.disabled}
              size="small"
            >
              {action.label}
            </Button>
          ))}
        </Stack>

        {/* Clear selection button */}
        <Button variant="outlined" onClick={onClearSelection} size="small">
          Clear
        </Button>
      </Stack>
    </Paper>
  );
}
