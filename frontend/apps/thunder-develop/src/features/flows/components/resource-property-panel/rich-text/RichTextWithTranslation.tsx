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

import type {Resource} from '@/features/flows/models/resources';
import {useMemo, useRef, useState, type ReactElement} from 'react';
import {useTranslation} from 'react-i18next';
import useValidationStatus from '@/features/flows/hooks/useValidationStatus';
import {Box, FormHelperText, IconButton, Tooltip} from '@wso2/oxygen-ui';
import {Languages} from '@wso2/oxygen-ui-icons-react';
import type {ToolbarPluginProps} from './helper-plugins/ToolbarPlugin';
import RichText from './RichText';
import I18nConfigurationCard from '../I18nConfigurationCard';

/**
 * Props interface for the RichTextWithTranslation component.
 */
export interface RichTextWithTranslationProps {
  /**
   * Options to customize the rich text editor toolbar.
   */
  ToolbarProps?: ToolbarPluginProps;
  /**
   * Listener for changes in the rich text editor content.
   *
   * @param value - The HTML string representation of the rich text editor content.
   */
  onChange: (value: string) => void;
  /**
   * Additional CSS class names to apply to the rich text editor container.
   */
  className?: string;
  /**
   * The resource associated with the rich text editor.
   */
  resource: Resource;
}

/**
 * Rich text editor component with translation support.
 */
function RichTextWithTranslation({
  ToolbarProps = {},
  className = '',
  onChange,
  resource,
}: RichTextWithTranslationProps): ReactElement {
  const {t} = useTranslation();
  const [isI18nCardOpen, setIsI18nCardOpen] = useState<boolean>(false);
  const buttonRef = useRef(null);
  const {selectedNotification} = useValidationStatus();

  /**
   * Get the error message for the rich text field.
   */
  const errorMessage: string = useMemo(() => {
    const key = `${resource?.id}_text`;

    if (selectedNotification?.hasResourceFieldNotification(key)) {
      return selectedNotification?.getResourceFieldNotification(key);
    }

    return '';
  }, [resource, selectedNotification]);

  return (
    <Box sx={{position: 'relative'}}>
      <RichText ToolbarProps={ToolbarProps} className={className} onChange={onChange} resource={resource} />
      {errorMessage && <FormHelperText error>{errorMessage}</FormHelperText>}
      <Tooltip title={t('flows:core.elements.textPropertyField.tooltip.configureTranslation')}>
        <IconButton
          ref={buttonRef}
          onClick={() => setIsI18nCardOpen(!isI18nCardOpen)}
          size="small"
          sx={{position: 'absolute', top: 8, right: 8}}
        >
          <Languages size={13} />
        </IconButton>
      </Tooltip>
      <I18nConfigurationCard
        open={isI18nCardOpen}
        anchorEl={buttonRef.current}
        propertyKey="richText"
        onClose={() => setIsI18nCardOpen(false)}
        i18nKey={(() => {
          const text = String((resource as Resource & {label?: string})?.label ?? '');
          const match = /^\{\{t\(([^)]+)\)\}\}$/.exec(text);
          return match?.[1] ?? '';
        })()}
        onChange={(i18nKey: string) => onChange(i18nKey ? `{{t(${i18nKey})}}` : '')}
      />
    </Box>
  );
}

export default RichTextWithTranslation;
