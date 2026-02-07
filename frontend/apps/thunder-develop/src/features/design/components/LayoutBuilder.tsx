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

import {useState, type JSX} from 'react';
import {
  Box,
  Paper,
  Toolbar,
  Stack,
  Typography,
  Button,
  IconButton,
  Slider,
  ToggleButton,
  ToggleButtonGroup,
  TextField,
  FormControl,
  InputLabel,
  Select,
  MenuItem,
  Accordion,
  AccordionSummary,
  AccordionDetails,
  Divider,
} from '@wso2/oxygen-ui';
import {ArrowLeft, Save, ZoomIn, ZoomOut, Grid3x3, ChevronDown} from '@wso2/oxygen-ui-icons-react';
import {useTranslation} from 'react-i18next';
import {useLogger} from '@thunder/logger';
import Ruler from './Ruler';
import LayoutPreview from '@/features/layouts/components/LayoutPreview';
import useNotification from '@/features/design-system/hooks/useNotification';
import useCreateLayout from '@/features/layouts/api/useCreateLayout';
import {defaultLayoutFormData} from '@/features/layouts/schemas/layoutSchema';
import type {LayoutConfig} from '@/features/layouts/models/layout';

const RULER_SIZE = 24;

/**
 * Props for LayoutBuilder component
 */
interface LayoutBuilderProps {
  /** Callback when builder is closed */
  onClose: () => void;
  /** Callback when layout is saved */
  onSaved: (layoutId: string) => void;
}

/**
 * Layout builder component with rulers and layout controls
 */
export default function LayoutBuilder({onClose, onSaved}: LayoutBuilderProps): JSX.Element {
  const {t} = useTranslation();
  const logger = useLogger('LayoutBuilder');
  const {showSuccess, showError} = useNotification();
  const createLayout = useCreateLayout();

  const [layoutName, setLayoutName] = useState('');
  const [layoutConfig, setLayoutConfig] = useState<LayoutConfig>(defaultLayoutFormData.layout);
  const [zoom, setZoom] = useState(100);
  const [showRulers, setShowRulers] = useState(true);
  const [showGrid, setShowGrid] = useState(false);

  const canvasWidth = 1440;
  const canvasHeight = 900;
  const scaledWidth = (canvasWidth * zoom) / 100;
  const scaledHeight = (canvasHeight * zoom) / 100;

  const handleZoomIn = () => setZoom((prev) => Math.min(prev + 10, 200));
  const handleZoomOut = () => setZoom((prev) => Math.max(prev - 10, 25));

  const handleLayoutConfigChange = (path: string, value: any) => {
    const keys = path.split('.');
    const newConfig = {...layoutConfig};
    let current: any = newConfig;

    for (let i = 0; i < keys.length - 1; i++) {
      if (!current[keys[i]]) {
        current[keys[i]] = {};
      }
      current = current[keys[i]];
    }
    current[keys[keys.length - 1]] = value;

    setLayoutConfig(newConfig);
  };

  const handleSave = async () => {
    if (!layoutName.trim()) {
      showError(t('design:layout.nameRequired', {defaultValue: 'Please enter a layout name'}));
      return;
    }

    try {
      const result = await createLayout.mutateAsync({
        displayName: layoutName,
        layout: layoutConfig,
      });

      logger.info('Layout created successfully', {id: result.id, name: layoutName});
      showSuccess(t('design:layout.saved', {defaultValue: `Layout "${layoutName}" saved successfully`}));
      onSaved(result.id);
    } catch (error) {
      logger.error('Failed to save layout', error);
      showError(t('design:layout.saveFailed', {defaultValue: 'Failed to save layout. Please try again.'}));
    }
  };

  return (
    <Box sx={{display: 'flex', height: '100vh', overflow: 'hidden', flexDirection: 'column'}}>
      {/* Toolbar */}
      <Toolbar
        sx={{
          borderBottom: '1px solid',
          borderColor: 'divider',
          bgcolor: 'background.paper',
          gap: 2,
          justifyContent: 'space-between',
        }}
      >
        <Stack direction="row" spacing={2} alignItems="center">
          <IconButton size="small" onClick={onClose}>
            <ArrowLeft size={18} />
          </IconButton>
          <Typography variant="h6">{t('design:layout.layoutBuilder', {defaultValue: 'Layout Builder'})}</Typography>
          <TextField
            size="small"
            placeholder={t('design:layout.enterLayoutName', {defaultValue: 'Enter layout name'})}
            value={layoutName}
            onChange={(e) => setLayoutName(e.target.value)}
            sx={{minWidth: 250}}
          />
        </Stack>

        <Stack direction="row" spacing={1} alignItems="center">
          {/* Rulers Toggle */}
          <ToggleButton
            value="rulers"
            selected={showRulers}
            onChange={() => setShowRulers(!showRulers)}
            size="small"
          >
            Rulers
          </ToggleButton>

          {/* Grid Toggle */}
          <ToggleButton value="grid" selected={showGrid} onChange={() => setShowGrid(!showGrid)} size="small">
            <Grid3x3 size={18} />
          </ToggleButton>

          <Divider orientation="vertical" flexItem sx={{mx: 1}} />

          {/* Zoom Controls */}
          <IconButton size="small" onClick={handleZoomOut} disabled={zoom <= 25}>
            <ZoomOut size={18} />
          </IconButton>
          <Typography variant="caption" sx={{minWidth: 50, textAlign: 'center'}}>
            {zoom}%
          </Typography>
          <IconButton size="small" onClick={handleZoomIn} disabled={zoom >= 200}>
            <ZoomIn size={18} />
          </IconButton>

          <Divider orientation="vertical" flexItem sx={{mx: 1}} />

          {/* Save Button */}
          <Button
            variant="contained"
            startIcon={<Save size={18} />}
            onClick={handleSave}
            disabled={!layoutName.trim() || createLayout.isPending}
          >
            {createLayout.isPending ? t('common:actions.saving', {defaultValue: 'Saving...'}) : t('common:actions.save', {defaultValue: 'Save'})}
          </Button>
        </Stack>
      </Toolbar>

      <Box sx={{display: 'flex', flex: 1, overflow: 'hidden'}}>
        {/* Canvas Area with Rulers */}
        <Box sx={{flex: 1, display: 'flex', flexDirection: 'column', overflow: 'auto', bgcolor: 'grey.100'}}>
          {showRulers && (
            <Box sx={{display: 'flex'}}>
              <Box sx={{width: RULER_SIZE, height: RULER_SIZE, bgcolor: 'background.paper'}} />
              <Ruler direction="horizontal" size={scaledWidth} zoom={zoom} />
            </Box>
          )}

          <Box sx={{display: 'flex', flex: 1}}>
            {showRulers && <Ruler direction="vertical" size={scaledHeight} zoom={zoom} />}

            <Box
              sx={{
                flex: 1,
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                p: 4,
                position: 'relative',
              }}
            >
              {showGrid && (
                <Box
                  sx={{
                    position: 'absolute',
                    top: 0,
                    left: 0,
                    right: 0,
                    bottom: 0,
                    backgroundImage: 'linear-gradient(rgba(0,0,0,0.1) 1px, transparent 1px), linear-gradient(90deg, rgba(0,0,0,0.1) 1px, transparent 1px)',
                    backgroundSize: '20px 20px',
                    opacity: 0.3,
                    pointerEvents: 'none',
                  }}
                />
              )}

              <Paper
                elevation={4}
                sx={{
                  width: scaledWidth,
                  height: scaledHeight,
                  transform: `scale(${zoom / 100})`,
                  transformOrigin: 'center center',
                  overflow: 'hidden',
                }}
              >
                <LayoutPreview layout={layoutConfig} width={canvasWidth} height={canvasHeight} />
              </Paper>
            </Box>
          </Box>
        </Box>

        {/* Layout Controls Sidebar */}
        <Box
          sx={{
            width: 320,
            borderLeft: '1px solid',
            borderColor: 'divider',
            overflow: 'auto',
            bgcolor: 'background.paper',
          }}
        >
          <Box sx={{p: 2}}>
            <Typography variant="h6" sx={{mb: 2}}>
              {t('design:layout.properties', {defaultValue: 'Layout Properties'})}
            </Typography>

            <Stack spacing={1}>
              {/* Template Selection */}
              <Accordion defaultExpanded>
                <AccordionSummary expandIcon={<ChevronDown size={20} />}>
                  <Typography variant="subtitle2">{t('design:layout.template', {defaultValue: 'Template'})}</Typography>
                </AccordionSummary>
                <AccordionDetails>
                  <FormControl fullWidth size="small">
                    <InputLabel>{t('design:layout.selectTemplate', {defaultValue: 'Select Template'})}</InputLabel>
                    <Select
                      value={layoutConfig.screens?.signin?.template || 'centered'}
                      onChange={(e) => handleLayoutConfigChange('screens.signin.template', e.target.value)}
                      label={t('design:layout.selectTemplate', {defaultValue: 'Select Template'})}
                    >
                      <MenuItem value="centered">{t('design:layout.centered', {defaultValue: 'Centered'})}</MenuItem>
                      <MenuItem value="split">{t('design:layout.split', {defaultValue: 'Split Screen'})}</MenuItem>
                      <MenuItem value="fullscreen">{t('design:layout.fullscreen', {defaultValue: 'Fullscreen'})}</MenuItem>
                    </Select>
                  </FormControl>
                </AccordionDetails>
              </Accordion>

              {/* Container Settings */}
              <Accordion>
                <AccordionSummary expandIcon={<ChevronDown size={20} />}>
                  <Typography variant="subtitle2">{t('design:layout.container', {defaultValue: 'Container'})}</Typography>
                </AccordionSummary>
                <AccordionDetails>
                  <Stack spacing={2}>
                    <TextField
                      fullWidth
                      size="small"
                      type="number"
                      label={t('design:layout.maxWidth', {defaultValue: 'Max Width (px)'})}
                      value={layoutConfig.screens?.signin?.container?.maxWidth || 400}
                      onChange={(e) => handleLayoutConfigChange('screens.signin.container.maxWidth', parseInt(e.target.value, 10))}
                    />
                    <TextField
                      fullWidth
                      size="small"
                      type="number"
                      label={t('design:layout.padding', {defaultValue: 'Padding (px)'})}
                      value={layoutConfig.screens?.signin?.container?.padding || 24}
                      onChange={(e) => handleLayoutConfigChange('screens.signin.container.padding', parseInt(e.target.value, 10))}
                    />
                    <TextField
                      fullWidth
                      size="small"
                      label={t('design:layout.borderRadius', {defaultValue: 'Border Radius'})}
                      value={layoutConfig.screens?.signin?.container?.borderRadius || '8px'}
                      onChange={(e) => handleLayoutConfigChange('screens.signin.container.borderRadius', e.target.value)}
                      placeholder="8px"
                    />
                  </Stack>
                </AccordionDetails>
              </Accordion>
            </Stack>
          </Box>
        </Box>
      </Box>
    </Box>
  );
}
