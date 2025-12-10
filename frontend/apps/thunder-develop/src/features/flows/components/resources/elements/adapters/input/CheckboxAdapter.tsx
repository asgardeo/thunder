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

import {memo, useMemo, type CSSProperties, type ReactElement} from 'react';
import type {Element as FlowElement} from '@/features/flows/models/elements';
import {Trans} from 'react-i18next';
import type {RequiredFieldInterface} from '@/features/flows/hooks/useRequiredFields';
import useRequiredFields from '@/features/flows/hooks/useRequiredFields';
import {Checkbox, FormControlLabel, FormHelperText} from '@wso2/oxygen-ui';
import PlaceholderComponent from '../PlaceholderComponent';
import {Hint} from '../../hint';

// PERFORMANCE: Define fields outside component to prevent recreation on every render
const CHECKBOX_VALIDATION_FIELDS: RequiredFieldInterface[] = [
  {
    errorMessage: 'Label is required',
    name: 'label',
  },
  {
    errorMessage: 'Identifier is required',
    name: 'identifier',
  },
];

/**
 * Configuration interface for Checkbox element.
 */
interface CheckboxConfig {
  className?: string;
  defaultValue?: string;
  label?: string;
  required?: boolean;
  styles?: CSSProperties;
  hint?: string;
}

/**
 * Checkbox element type.
 */
export type CheckboxElement = FlowElement<CheckboxConfig>;

/**
 * Props interface of {@link CheckboxAdapter}
 */
export interface CheckboxAdapterPropsInterface {
  /**
   * The checkbox element properties.
   */
  resource: FlowElement;
}

/**
 * Adapter for the Checkbox component.
 *
 * PERFORMANCE: This component has been optimized to:
 * 1. Use static validation fields defined outside the component
 * 2. Remove useTranslation hook to avoid re-renders
 * 3. Memoize the general message
 *
 * @param props - Props injected to the component.
 * @returns The CheckboxAdapter component.
 */
function CheckboxAdapter({resource}: CheckboxAdapterPropsInterface): ReactElement {
  // PERFORMANCE: Memoize general message - only depends on resource.id
  const generalMessage: ReactElement = useMemo(
    () => (
      <Trans i18nKey="flows:core.validation.fields.checkbox.general" values={{id: resource.id}}>
        Required fields are not properly configured for the checkbox field with ID <code>{resource.id}</code>.
      </Trans>
    ),
    [resource?.id],
  );

  // PERFORMANCE: Use static fields array defined outside component
  useRequiredFields(resource, generalMessage, CHECKBOX_VALIDATION_FIELDS);

  // eslint-disable-next-line @typescript-eslint/no-unsafe-assignment -- Config type is validated at runtime
  const checkboxConfig = resource.config as CheckboxConfig | undefined;

  return (
    <div>
      <FormControlLabel
        control={<Checkbox defaultChecked />}
        className={checkboxConfig?.className}
        label={<PlaceholderComponent value={checkboxConfig?.label ?? ''} />}
        required={checkboxConfig?.required}
        style={checkboxConfig?.styles}
      />
      {checkboxConfig?.hint && (
        <FormHelperText>
          <Hint hint={checkboxConfig?.hint} />
        </FormHelperText>
      )}
    </div>
  );
}

// PERFORMANCE: Memoize to prevent re-renders during drag operations
export default memo(CheckboxAdapter, (prevProps, nextProps) =>
  prevProps.resource === nextProps.resource
);
