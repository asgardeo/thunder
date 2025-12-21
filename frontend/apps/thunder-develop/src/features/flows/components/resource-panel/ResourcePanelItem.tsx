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

import {Avatar, Box, Card, CardContent, IconButton, Stack, Typography, useColorScheme} from '@wso2/oxygen-ui';
import {PlusIcon} from '@wso2/oxygen-ui-icons-react';
import React, {type HTMLAttributes, type ReactElement} from 'react';
import type {Resource} from '../../models/resources';

/**
 * Props interface of {@link ResourcePanelItem}
 */
export interface ResourcePanelItemProps extends Omit<HTMLAttributes<HTMLDivElement>, 'resource'> {
  /**
   * The resource item.
   */
  resource: Resource;
  /**
   * The type of the resource item.
   */
  type?: 'draggable' | 'static';
  /**
   * Callback to be triggered when a resource add button is clicked.
   * @param resource - Added resource.
   */
  onAdd?: (resource: Resource) => void;
  /**
   * Flag to disable the panel item.
   */
  disabled?: boolean;
}

/**
 * Resource panel item component.
 *
 * @param props - Props injected to the component.
 * @returns The ResourcePanelItem component.
 */
function ResourcePanelItem({
  children,
  resource,
  type = 'static',
  onAdd = undefined,
  disabled = false,
}: ResourcePanelItemProps): ReactElement | React.ReactNode {
  const {mode, systemMode} = useColorScheme();

  // Determine the effective mode - if mode is 'system', use systemMode
  const effectiveMode = mode === 'system' ? systemMode : mode;

  return (
    children ?? (
      <Card
        elevation={0}
        sx={{
          border: '1px solid',
          borderColor: 'divider',
          borderRadius: 1.5,
          transition: 'all 0.2s ease-in-out',
          cursor: type === 'draggable' ? 'grab' : 'default',
          backgroundColor: 'background.paper',
          '&:hover': {
            borderColor: type === 'draggable' ? 'primary.main' : 'divider',
            boxShadow: type === 'draggable' ? 2 : 'none',
            transform: type === 'draggable' ? 'translateY(-2px)' : 'none',
            backgroundColor: type === 'draggable' ? 'action.hover' : 'background.paper',
          },
          '&:active': {
            cursor: type === 'draggable' ? 'grabbing' : 'default',
          },
        }}
      >
        <CardContent
          sx={{
            p: 1.5,
            '&:last-child': {
              pb: 1.5,
            },
          }}
        >
          <Box display="flex" justifyContent="space-between" alignItems="center" gap={1}>
            <Stack direction="row" spacing={1.5} alignItems="center" flex={1}>
              <Avatar
                src={resource?.display?.image ? `${import.meta.env.BASE_URL}/${resource.display.image}` : undefined}
                variant="square"
                sx={{
                  height: 20,
                  width: 20,
                  backgroundColor: 'transparent !important',
                  color: 'text.primary',
                  '& .MuiAvatar-img': {
                    filter: effectiveMode === 'dark' ? 'brightness(0.9) invert(1)' : 'none',
                  },
                }}
              />
              <Stack direction="column" spacing={0.25} flex={1}>
                <Typography variant="body2" fontWeight={500} color="text.primary">
                  {resource?.display?.label}
                </Typography>
                {resource?.display?.description && (
                  <Typography variant="caption" color="text.secondary" sx={{lineHeight: 1.3}}>
                    {resource?.display?.description}
                  </Typography>
                )}
              </Stack>
            </Stack>
            {onAdd && (
              <IconButton
                onClick={() => onAdd(resource)}
                disabled={disabled}
                size="small"
                sx={{
                  height: 28,
                  width: 28,
                  border: '1px solid',
                  borderColor: 'divider',
                  borderRadius: 1,
                  backgroundColor: 'action.selected',
                  '&:hover': {
                    backgroundColor: 'primary.main',
                    borderColor: 'primary.main',
                    color: 'primary.contrastText',
                  },
                  '&.Mui-disabled': {
                    backgroundColor: 'action.disabledBackground',
                    borderColor: 'divider',
                  },
                }}
              >
                <PlusIcon size={14} />
              </IconButton>
            )}
          </Box>
        </CardContent>
      </Card>
    )
  );
}

export default ResourcePanelItem;
