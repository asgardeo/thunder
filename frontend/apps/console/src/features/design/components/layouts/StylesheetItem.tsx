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

import type {Stylesheet} from '@thunderid/design';
import {Accordion, AccordionDetails, AccordionSummary, Box, Chip, Stack, Tooltip} from '@wso2/oxygen-ui';
import {Trash, ChevronUp, ChevronDown, Eye, EyeOff} from '@wso2/oxygen-ui-icons-react';
import type {JSX} from 'react';
import {useTranslation} from 'react-i18next';
import EditableTitle from './EditableTitle';
import InlineCSSField from './InlineCSSField';
import UrlField from './UrlField';

export interface StylesheetItemProps {
  sheet: Stylesheet;
  idx: number;
  total: number;
  expanded: boolean;
  colorMode: 'light' | 'dark';
  onToggle: () => void;
  onRemove: () => void;
  onMove: (direction: -1 | 1) => void;
  onUpdate: (patch: Partial<Stylesheet>) => void;
  registerFlush: (flush: (() => void) | null) => void;
}

export default function StylesheetItem({
  sheet,
  idx,
  total,
  expanded,
  colorMode,
  onToggle,
  onRemove,
  onMove,
  onUpdate,
  registerFlush,
}: StylesheetItemProps): JSX.Element {
  const {t} = useTranslation('design');
  const isInline = sheet.type === 'inline';
  const isDisabled = !!sheet.disabled;

  return (
    <Accordion
      expanded={expanded}
      onChange={onToggle}
      disableGutters
      square
      sx={{
        backgroundColor: 'transparent',
        '&:before': {display: 'none'},
        overflow: 'visible',
        opacity: isDisabled ? 0.5 : 1,
        transition: 'opacity 0.15s ease',
      }}
    >
      <AccordionSummary
        expandIcon={<ChevronDown size={16} />}
        sx={{
          '& .MuiAccordionSummary-content': {alignItems: 'center', gap: 0.5, overflow: 'hidden', minWidth: 0},
          minHeight: 40,
          '&.Mui-expanded': {minHeight: 40},
        }}
      >
        {/* Reorder arrows */}
        <Stack component="span" onClick={(e) => e.stopPropagation()} sx={{flexShrink: 0, mr: 0.25}}>
          <Box
            component="span"
            role="button"
            tabIndex={0}
            aria-label={t('layouts.config.custom_css.actions.move_up.label', 'Move up')}
            onClick={() => {
              if (idx > 0) onMove(-1);
            }}
            onKeyDown={(e) => {
              if (e.key === 'Enter' || e.key === ' ') {
                e.preventDefault();
                e.stopPropagation();
                if (idx > 0) onMove(-1);
              }
            }}
            sx={{
              display: 'flex',
              cursor: idx === 0 ? 'default' : 'pointer',
              opacity: idx === 0 ? 0.25 : 0.5,
              '&:hover': {opacity: idx === 0 ? 0.25 : 1},
              lineHeight: 0,
            }}
          >
            <ChevronUp size={12} />
          </Box>
          <Box
            component="span"
            role="button"
            tabIndex={0}
            aria-label={t('layouts.config.custom_css.actions.move_down.label', 'Move down')}
            onClick={() => {
              if (idx < total - 1) onMove(1);
            }}
            onKeyDown={(e) => {
              if (e.key === 'Enter' || e.key === ' ') {
                e.preventDefault();
                e.stopPropagation();
                if (idx < total - 1) onMove(1);
              }
            }}
            sx={{
              display: 'flex',
              cursor: idx === total - 1 ? 'default' : 'pointer',
              opacity: idx === total - 1 ? 0.25 : 0.5,
              '&:hover': {opacity: idx === total - 1 ? 0.25 : 1},
              lineHeight: 0,
            }}
          >
            <ChevronDown size={12} />
          </Box>
        </Stack>

        <EditableTitle value={sheet.id} onChange={(id) => onUpdate({id})} />

        <Chip
          label={isInline ? 'Inline' : 'URL'}
          size="small"
          variant="outlined"
          sx={{fontSize: '0.6rem', height: 18, flexShrink: 0, '& .MuiChip-label': {px: 0.75}}}
        />

        <Box sx={{flex: 1}} />

        {/* Show/hide toggle — uses <span> to avoid nested <button> inside AccordionSummary */}
        <Tooltip
          title={
            isDisabled
              ? t('layouts.config.custom_css.actions.show_in_preview.tooltip', 'Show in preview')
              : t('layouts.config.custom_css.actions.hide_from_preview.tooltip', 'Hide from preview')
          }
          placement="top"
        >
          <Box
            component="span"
            role="button"
            tabIndex={0}
            aria-label={
              isDisabled
                ? t('layouts.config.custom_css.actions.show_in_preview.tooltip', 'Show in preview')
                : t('layouts.config.custom_css.actions.hide_from_preview.tooltip', 'Hide from preview')
            }
            onClick={(e) => {
              e.stopPropagation();
              onUpdate({disabled: !isDisabled});
            }}
            onKeyDown={(e) => {
              if (e.key === 'Enter' || e.key === ' ') {
                e.preventDefault();
                e.stopPropagation();
                onUpdate({disabled: !isDisabled});
              }
            }}
            sx={{
              display: 'inline-flex',
              alignItems: 'center',
              justifyContent: 'center',
              cursor: 'pointer',
              borderRadius: '50%',
              p: 0.25,
              color: isDisabled ? 'text.disabled' : 'text.secondary',
              '&:hover': {bgcolor: 'action.hover'},
            }}
          >
            {isDisabled ? <EyeOff size={15} /> : <Eye size={15} />}
          </Box>
        </Tooltip>

        {/* Delete — uses <span> to avoid nested <button> inside AccordionSummary */}
        <Tooltip title={t('layouts.config.custom_css.actions.remove.tooltip', 'Remove stylesheet')} placement="top">
          <Box
            component="span"
            role="button"
            tabIndex={0}
            aria-label={t('layouts.config.custom_css.actions.remove.tooltip', 'Remove stylesheet')}
            onClick={(e) => {
              e.stopPropagation();
              onRemove();
            }}
            onKeyDown={(e) => {
              if (e.key === 'Enter' || e.key === ' ') {
                e.preventDefault();
                e.stopPropagation();
                onRemove();
              }
            }}
            sx={{
              display: 'inline-flex',
              alignItems: 'center',
              justifyContent: 'center',
              cursor: 'pointer',
              borderRadius: '50%',
              p: 0.25,
              color: 'text.disabled',
              '&:hover': {color: 'error.main', bgcolor: 'action.hover'},
            }}
          >
            <Trash size={15} />
          </Box>
        </Tooltip>
      </AccordionSummary>

      <AccordionDetails sx={{display: 'flex', flexDirection: 'column', gap: 1, overflow: 'visible'}}>
        {isInline ? (
          <InlineCSSField
            id={sheet.id}
            content={sheet.content}
            colorMode={colorMode}
            onChange={(content) => onUpdate({content})}
            registerFlush={registerFlush}
          />
        ) : (
          <UrlField sheet={sheet} onUpdate={onUpdate} />
        )}
      </AccordionDetails>
    </Accordion>
  );
}
