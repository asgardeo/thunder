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

import lowerCase from 'lodash-es/lowerCase';
import startCase from 'lodash-es/startCase';
import {type ReactElement, type SyntheticEvent, useCallback, useEffect, useState} from 'react';
import {useTranslation} from 'react-i18next';
import {Box, Card, CardContent, CardHeader, IconButton, Popover, Tab, Tabs} from '@wso2/oxygen-ui';
import {XIcon} from '@wso2/oxygen-ui-icons-react';
import {extractI18nKey, extractMetaKey, isMetaPattern} from '../../utils/i18nPatternUtils';
import {I18nConfigurationCardContent} from './I18nConfigurationCard';
import {MetaConfigurationCardContent} from './MetaConfigurationCard';

/**
 * Tab identifiers for the dynamic value popover.
 */
const TAB_TRANSLATION = 0;
const TAB_VARIABLES = 1;

/**
 * Props interface of {@link DynamicValuePopover}
 */
export interface DynamicValuePopoverPropsInterface {
  open: boolean;
  anchorEl: HTMLElement | null;
  propertyKey: string;
  onClose: () => void;
  /**
   * The current full value, e.g. `{{t(flowI18n:key)}}` or `{{meta(application.name)}}`.
   */
  value: string;
  /**
   * Called with the new formatted value when the user selects/changes a dynamic value.
   */
  onChange: (newValue: string) => void;
}

/**
 * Tabbed popover for configuring dynamic values in property fields.
 * Provides two tabs: Translation ({{t(...)}}) and Variables ({{meta(...)}}).
 */
function DynamicValuePopover({
  open,
  anchorEl,
  propertyKey,
  onClose,
  value,
  onChange,
}: DynamicValuePopoverPropsInterface): ReactElement {
  const {t} = useTranslation();
  const [activeTab, setActiveTab] = useState<number>(() => (isMetaPattern(value) ? TAB_VARIABLES : TAB_TRANSLATION));

  /**
   * Reset active tab based on current value when popover opens.
   */
  useEffect(() => {
    if (open) {
      setActiveTab(isMetaPattern(value) ? TAB_VARIABLES : TAB_TRANSLATION);
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [open]);

  const handleClose = useCallback(() => {
    onClose();
  }, [onClose]);

  const handleTabChange = (_event: SyntheticEvent, newTab: number) => {
    setActiveTab(newTab);
  };

  return (
    <Popover
      open={open}
      anchorEl={anchorEl}
      onClose={handleClose}
      anchorOrigin={{
        vertical: 'top',
        horizontal: 'right',
      }}
      transformOrigin={{
        vertical: 'top',
        horizontal: 'left',
      }}
    >
      <Card sx={{width: 400}}>
        <CardHeader
          title={t('flows:core.elements.textPropertyField.dynamicValuePopover.title', {
            field: startCase(lowerCase(propertyKey)),
          })}
          action={
            <IconButton aria-label={t('common:close')} onClick={handleClose} size="small">
              <XIcon />
            </IconButton>
          }
          sx={{pb: 0}}
        />
        <Tabs
          value={activeTab}
          onChange={handleTabChange}
          sx={{px: 2, borderBottom: 1, borderColor: 'divider'}}
        >
          <Tab label={t('flows:core.elements.textPropertyField.dynamicValuePopover.tabs.translation')} />
          <Tab label={t('flows:core.elements.textPropertyField.dynamicValuePopover.tabs.variables')} />
        </Tabs>
        <CardContent>
          <Box role="tabpanel" hidden={activeTab !== TAB_TRANSLATION}>
            {activeTab === TAB_TRANSLATION && (
              <I18nConfigurationCardContent
                propertyKey={propertyKey}
                i18nKey={extractI18nKey(value) ?? ''}
                isActive={open && activeTab === TAB_TRANSLATION}
                onChange={(i18nKey: string) => onChange(i18nKey ? `{{t(${i18nKey})}}` : '')}
              />
            )}
          </Box>
          <Box role="tabpanel" hidden={activeTab !== TAB_VARIABLES}>
            {activeTab === TAB_VARIABLES && (
              <MetaConfigurationCardContent
                metaKey={extractMetaKey(value) ?? ''}
                onChange={(metaKey: string) => onChange(metaKey ? `{{meta(${metaKey})}}` : '')}
              />
            )}
          </Box>
        </CardContent>
      </Card>
    </Popover>
  );
}

export default DynamicValuePopover;
