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

import {Box, Button, IconButton, Stack, TextField, Tooltip, Typography} from '@wso2/oxygen-ui';
import {ArrowLeft, Check, ChevronLeftIcon, Edit, X} from '@wso2/oxygen-ui-icons-react';
import {
  memo,
  useCallback,
  useEffect,
  useRef,
  useState,
  type HTMLAttributes,
  type KeyboardEvent,
  type ReactElement,
} from 'react';

/**
 * Props interface of {@link BuilderPanelHeader}
 */
export interface BuilderPanelHeaderProps extends HTMLAttributes<HTMLDivElement> {
  /**
   * Current title of the builder resource (e.g. flow name).
   */
  title?: string;
  /**
   * URL-friendly identifier displayed beneath the title.
   */
  handle?: string;
  /**
   * Callback to navigate back. When omitted the back button is not rendered.
   */
  onBack?: () => void;
  /**
   * Callback to collapse the side panel. When omitted the collapse button is not rendered.
   */
  onPanelToggle?: () => void /**
   * Callback invoked with the new title when the user saves an edit.
   * When omitted, the edit icon is not shown.
   */;
  onTitleChange?: (newTitle: string) => void;
  /**
   * Label for the back button.
   * @defaultValue "Back"
   */
  backLabel?: string;
  /**
   * Tooltip for the collapse (hide panel) button.
   * @defaultValue "Hide panel"
   */
  hidePanelTooltip?: string;
  /**
   * Tooltip for the edit title icon button.
   * @defaultValue "Edit title"
   */
  editTitleTooltip?: string;
  /**
   * Tooltip for the save title icon button.
   * @defaultValue "Save"
   */
  saveTitleTooltip?: string;
  /**
   * Tooltip for the cancel edit icon button.
   * @defaultValue "Cancel"
   */
  cancelEditTooltip?: string;
}

/**
 * Reusable header for builder side panels.
 *
 * Renders a back-navigation button, an optional panel-collapse toggle, and an
 * inline-editable title with an optional URL-friendly handle displayed below it.
 *
 * @param props - Props injected to the component.
 * @returns The BuilderPanelHeader component.
 */
function BuilderPanelHeader({
  title = '',
  handle = '',
  onBack = undefined,
  onPanelToggle = undefined,
  onTitleChange = undefined,
  backLabel = 'Back',
  hidePanelTooltip = 'Hide panel',
  editTitleTooltip = 'Edit title',
  saveTitleTooltip = 'Save',
  cancelEditTooltip = 'Cancel',
  ...rest
}: BuilderPanelHeaderProps): ReactElement {
  const [isEditing, setIsEditing] = useState(false);
  const [editedTitle, setEditedTitle] = useState(title);
  const inputRef = useRef<HTMLInputElement>(null);

  // Keep local state in sync when the title prop changes externally.
  useEffect(() => {
    setEditedTitle(title);
  }, [title]);

  // Auto-focus + select on entering edit mode.
  useEffect(() => {
    if (isEditing && inputRef.current) {
      inputRef.current.focus();
      inputRef.current.select();
    }
  }, [isEditing]);

  const handleEditClick = useCallback(() => {
    setIsEditing(true);
  }, []);

  const handleSave = useCallback(() => {
    const trimmed = editedTitle.trim();

    if (!trimmed) return; // keep editing if empty

    if (trimmed !== title) {
      onTitleChange?.(trimmed);
    }
    setIsEditing(false);
  }, [editedTitle, title, onTitleChange]);

  const handleCancel = useCallback(() => {
    setEditedTitle(title);
    setIsEditing(false);
  }, [title]);

  const handleKeyDown = useCallback(
    (event: KeyboardEvent) => {
      if (event.key === 'Enter') handleSave();
      else if (event.key === 'Escape') handleCancel();
    },
    [handleSave, handleCancel],
  );

  return (
    <Box
      sx={{
        pb: 1.5,
        borderBottom: '1px solid',
        borderColor: 'divider',
        mb: 1,
        flexShrink: 0,
      }}
      {...rest}
    >
      {/* Back button + collapse toggle row */}
      {(onBack !== undefined || onPanelToggle !== undefined) && (
        <Box display="flex" alignItems="center" justifyContent="space-between" sx={{mb: 1}}>
          {onBack !== undefined && (
            <Tooltip title={backLabel} placement="right">
              <Button onClick={onBack} size="small" startIcon={<ArrowLeft size={16} />}>
                {backLabel}
              </Button>
            </Tooltip>
          )}
          {onPanelToggle !== undefined && (
            <Tooltip title={hidePanelTooltip} placement="right">
              <IconButton onClick={onPanelToggle} size="small">
                <ChevronLeftIcon size={16} />
              </IconButton>
            </Tooltip>
          )}
        </Box>
      )}

      {/* Title section */}
      {title &&
        (isEditing ? (
          <Stack direction="row" alignItems="center" spacing={0.5}>
            <TextField
              inputRef={inputRef}
              value={editedTitle}
              onChange={(e) => setEditedTitle(e.target.value)}
              onKeyDown={handleKeyDown}
              size="small"
              variant="outlined"
              fullWidth
              sx={{
                '& .MuiInputBase-input': {
                  py: 0.5,
                  fontSize: '1rem',
                  fontWeight: 600,
                },
              }}
            />
            <Tooltip title={saveTitleTooltip}>
              <IconButton size="small" onClick={handleSave} color="primary">
                <Check size={16} />
              </IconButton>
            </Tooltip>
            <Tooltip title={cancelEditTooltip}>
              <IconButton size="small" onClick={handleCancel}>
                <X size={16} />
              </IconButton>
            </Tooltip>
          </Stack>
        ) : (
          <Stack direction="column" spacing={0}>
            <Stack direction="row" alignItems="center" spacing={0.5}>
              <Typography variant="h6" sx={{fontWeight: 600}}>
                {title}
              </Typography>
              {onTitleChange && (
                <Tooltip title={editTitleTooltip}>
                  <IconButton size="small" onClick={handleEditClick} sx={{p: 0.25}}>
                    <Edit size={14} />
                  </IconButton>
                </Tooltip>
              )}
            </Stack>
            {handle && (
              <Typography variant="caption" color="text.secondary">
                {handle}
              </Typography>
            )}
          </Stack>
        ))}
    </Box>
  );
}

export default memo(BuilderPanelHeader);
