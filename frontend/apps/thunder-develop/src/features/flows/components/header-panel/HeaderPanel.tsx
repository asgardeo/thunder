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

import {type HTMLAttributes, type ReactElement, useCallback, useState, useRef, useEffect} from 'react';
import {useTranslation} from 'react-i18next';
import {useNavigate} from 'react-router';
import {
  Box,
  Button,
  Card,
  type CardProps,
  ColorSchemeToggle,
  IconButton,
  Stack,
  Typography,
  Tooltip,
  TextField,
} from '@wso2/oxygen-ui';
import {ArrowLeft, Save, LayoutGrid, Edit, X, Check} from '@wso2/oxygen-ui-icons-react';
import ValidationStatusLabels from '../validation-panel/ValidationStatusLabels';
import EdgeStyleSelector from './EdgeStyleSelector';

/**
 * Props interface of {@link HeaderPanel}
 */
export interface HeaderPanelPropsInterface extends Omit<CardProps, 'open' | 'title'>, HTMLAttributes<HTMLDivElement> {
  /**
   * Title to display in the header.
   */
  title: string;
  /**
   * URL-friendly handle for the flow.
   */
  handle: string;
  /**
   * Callback to be triggered when title changes.
   */
  onTitleChange?: (newTitle: string) => void;
  /**
   * Callback to be triggered when back button is clicked.
   */
  onBack?: () => void;
  /**
   * Callback to be triggered when save button is clicked.
   */
  onSave?: () => void;
  /**
   * Callback to be triggered when auto-layout button is clicked.
   */
  onAutoLayout?: () => void;
}

/**
 * Flow builder header panel that appears at the top.
 *
 * @param props - Props injected to the component.
 * @returns The HeaderPanel component.
 */
function HeaderPanel({
  title,
  handle,
  onTitleChange = undefined,
  onBack = undefined,
  onSave = undefined,
  onAutoLayout = undefined,
  ...rest
}: HeaderPanelPropsInterface): ReactElement {
  const {t} = useTranslation();
  const navigate = useNavigate();

  const [isEditingTitle, setIsEditingTitle] = useState(false);
  const [editedTitle, setEditedTitle] = useState(title);
  const inputRef = useRef<HTMLInputElement>(null);

  // Sync editedTitle when title prop changes
  useEffect(() => {
    setEditedTitle(title);
  }, [title]);

  // Focus input when entering edit mode
  useEffect(() => {
    if (isEditingTitle && inputRef.current) {
      inputRef.current.focus();
      inputRef.current.select();
    }
  }, [isEditingTitle]);

  const handleBackClick = useCallback(() => {
    if (onBack) {
      onBack();
      return;
    }

    // Go back to the flows list
    // eslint-disable-next-line @typescript-eslint/no-floating-promises
    navigate('/flows');
  }, [onBack, navigate]);

  const handleSaveClick = useCallback(() => {
    // Note: getNodes/getEdges returns visually displayed nodes (may be filtered)
    // The actual save handler uses the full unfiltered data
    onSave?.();
  }, [onSave]);

  const handleEditClick = useCallback(() => {
    setIsEditingTitle(true);
  }, []);

  const handleTitleSave = useCallback(() => {
    const trimmedTitle = editedTitle.trim();
    if (trimmedTitle && trimmedTitle !== title) {
      onTitleChange?.(trimmedTitle);
    } else {
      setEditedTitle(title);
    }
    setIsEditingTitle(false);
  }, [editedTitle, title, onTitleChange]);

  const handleTitleCancel = useCallback(() => {
    setEditedTitle(title);
    setIsEditingTitle(false);
  }, [title]);

  const handleKeyDown = useCallback(
    (event: React.KeyboardEvent) => {
      if (event.key === 'Enter') {
        handleTitleSave();
      } else if (event.key === 'Escape') {
        handleTitleCancel();
      }
    },
    [handleTitleSave, handleTitleCancel],
  );

  return (
    <Box sx={{flexShrink: 0, pb: 1}} {...rest}>
      <Card
        elevation={0}
        sx={{
          borderRadius: 1,
          height: 50,
          display: 'flex',
          alignItems: 'center',
          px: 2,
        }}
      >
        <Stack direction="row" alignItems="center" spacing={2} width="100%">
          {/* Left section - Back button and title */}
          <Stack direction="row" alignItems="center">
            <Button onClick={handleBackClick} startIcon={<ArrowLeft size={20} />}>
              {t('flows:core.headerPanel.goBack')}
            </Button>
          </Stack>

          {/* Title section - editable or display mode */}
          {isEditingTitle ? (
            <Stack direction="row" alignItems="center" spacing={0.5}>
              <TextField
                inputRef={inputRef}
                value={editedTitle}
                onChange={(e) => setEditedTitle(e.target.value)}
                onKeyDown={handleKeyDown}
                size="small"
                variant="outlined"
                sx={{
                  '& .MuiInputBase-input': {
                    py: 0.5,
                    fontSize: '1.25rem',
                    fontWeight: 500,
                  },
                }}
              />
              <Tooltip title={t('flows:core.headerPanel.saveTitle')}>
                <IconButton size="small" onClick={handleTitleSave} color="primary">
                  <Check size={18} />
                </IconButton>
              </Tooltip>
              <Tooltip title={t('flows:core.headerPanel.cancelEdit')}>
                <IconButton size="small" onClick={handleTitleCancel}>
                  <X size={18} />
                </IconButton>
              </Tooltip>
            </Stack>
          ) : (
            <Stack direction="row" alignItems="center" spacing={1}>
              <Stack direction="column" spacing={0}>
                <Stack direction="row" alignItems="center" spacing={0.5}>
                  <Typography variant="h4">{title}</Typography>
                  {onTitleChange && (
                    <Tooltip title={t('flows:core.headerPanel.editTitle')}>
                      <IconButton size="small" onClick={handleEditClick}>
                        <Edit size={16} />
                      </IconButton>
                    </Tooltip>
                  )}
                </Stack>
                {handle && (
                  <Typography variant="caption" color="text.secondary" sx={{mt: -0.5}}>
                    {handle}
                  </Typography>
                )}
              </Stack>
            </Stack>
          )}

          {/* Right section - Action buttons */}
          <Stack direction="row" alignItems="center" spacing={1} ml="auto">
            <ValidationStatusLabels />
            {onAutoLayout && (
              <Tooltip title={t('flows:core.headerPanel.autoLayout')}>
                <IconButton size="small" onClick={onAutoLayout}>
                  <LayoutGrid size={20} />
                </IconButton>
              </Tooltip>
            )}
            <EdgeStyleSelector />
            <ColorSchemeToggle data-testid="theme-toggle" />
            <Button variant="contained" startIcon={<Save size={20} />} onClick={handleSaveClick}>
              {t('flows:core.headerPanel.save')}
            </Button>
          </Stack>
        </Stack>
      </Card>
    </Box>
  );
}

export default HeaderPanel;
