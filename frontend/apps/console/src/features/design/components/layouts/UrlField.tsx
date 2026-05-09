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

import {isValidStylesheetUrl, isInsecureStylesheetUrl, type UrlStylesheet} from '@thunderid/design';
import {FormControl, FormLabel, TextField} from '@wso2/oxygen-ui';
import type {JSX} from 'react';
import {useTranslation} from 'react-i18next';

export interface UrlFieldProps {
  sheet: UrlStylesheet;
  onUpdate: (patch: Partial<UrlStylesheet>) => void;
}

function UrlField({sheet, onUpdate}: UrlFieldProps): JSX.Element {
  const {t} = useTranslation('design');
  const hasError = !!sheet.href && !isValidStylesheetUrl(sheet.href);
  const isInsecure = Boolean(sheet.href) && !hasError && isInsecureStylesheetUrl(sheet.href);

  let helperText: string | undefined;
  if (hasError)
    helperText = t(
      'layouts.config.custom_css.fields.url.errors.invalid_url',
      'URL must be a valid http:// or https:// address',
    );
  else if (isInsecure)
    helperText = t(
      'layouts.config.custom_css.fields.url.warnings.insecure_protocol',
      'Using HTTP is insecure. Consider using HTTPS instead.',
    );

  return (
    <FormControl fullWidth>
      <FormLabel>{t('layouts.config.custom_css.fields.url.label', 'URL')}</FormLabel>
      <TextField
        size="small"
        value={sheet.href}
        onChange={(e) => onUpdate({href: e.target.value})}
        fullWidth
        error={hasError}
        color={isInsecure ? 'warning' : undefined}
        focused={isInsecure || undefined}
        helperText={helperText}
        slotProps={{
          input: {sx: {fontSize: '0.8rem', fontFamily: 'monospace'}},
          formHelperText: isInsecure ? {sx: {color: 'warning.main'}} : undefined,
        }}
      />
    </FormControl>
  );
}

export default UrlField;
