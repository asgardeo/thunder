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

import type {JSX} from 'react';
import {Box, Checkbox, Divider, FormControlLabel, Typography} from '@wso2/oxygen-ui';
import {Consent, ConsentCheckboxList, type ConsentPurpose, type ConsentRenderProps} from '@asgardeo/react';

/**
 * Props for the ConsentAdapter component.
 *  Includes the raw consent data from the backend, current form values for tracking optional checkbox state,
 *  and a handler for when the user toggles an optional attribute.
 */
interface ConsentAdapterProps {
  /** Raw consent data from additionalData.consentPrompt */
  consentData?: string | ConsentPurpose[] | {purposes: ConsentPurpose[]};
  /** Current form values for tracking optional checkbox state */
  formValues: Record<string, string>;
  /** Handler invoked when the user toggles an optional attribute */
  onInputChange: (name: string, value: string) => void;
}

/**
 * Oxygen-UI styled consent adapter.
 *
 * Uses the SDK's `Consent` render-prop component to parse the backend data,
 * then renders each purpose section with oxygen-ui `Checkbox` and `Typography`.
 */
export default function ConsentAdapter({
  consentData = undefined,
  formValues,
  onInputChange,
}: ConsentAdapterProps): JSX.Element | null {
  if (!consentData) return null;

  return (
    <Consent consentData={consentData} formValues={formValues} onInputChange={onInputChange}>
      {({purposes}: ConsentRenderProps) => (
        <Box sx={{display: 'flex', flexDirection: 'column', gap: 2, mt: 1}}>
          {purposes.map((purpose, idx) => (
            <Box key={purpose.purpose_id ?? idx}>
              {purpose.essential && purpose.essential.length > 0 && (
                <Box sx={{mt: 1}}>
                  <Typography variant="subtitle2" fontWeight="bold" sx={{mb: 0.5}}>
                    Essential Attributes
                  </Typography>
                  <ConsentCheckboxList
                    variant="ESSENTIAL"
                    purpose={purpose}
                    formValues={formValues}
                    onInputChange={onInputChange}
                  >
                    {({attributes, isChecked}) => (
                      <Box sx={{display: 'flex', flexDirection: 'column', gap: 0.5, pl: 1}}>
                        {attributes.map((attr) => (
                          <FormControlLabel
                            key={attr}
                            control={<Checkbox checked={isChecked(attr)} disabled size="small" />}
                            label={
                              <Typography variant="body2" sx={{opacity: 0.7}}>
                                {attr}
                              </Typography>
                            }
                          />
                        ))}
                      </Box>
                    )}
                  </ConsentCheckboxList>
                </Box>
              )}
              {purpose.optional && purpose.optional.length > 0 && (
                <Box sx={{mt: 1}}>
                  <Typography variant="subtitle2" fontWeight="bold" sx={{mb: 0.5}}>
                    Optional Attributes
                  </Typography>
                  <ConsentCheckboxList
                    variant="OPTIONAL"
                    purpose={purpose}
                    formValues={formValues}
                    onInputChange={onInputChange}
                  >
                    {({attributes, isChecked, handleChange}) => (
                      <Box sx={{display: 'flex', flexDirection: 'column', gap: 0.5, pl: 1}}>
                        {attributes.map((attr) => (
                          <FormControlLabel
                            key={attr}
                            control={
                              <Checkbox
                                checked={isChecked(attr)}
                                onChange={(e) => handleChange(attr, (e.target as HTMLInputElement).checked)}
                                size="small"
                              />
                            }
                            label={<Typography variant="body2">{attr}</Typography>}
                          />
                        ))}
                      </Box>
                    )}
                  </ConsentCheckboxList>
                </Box>
              )}
              {idx < purposes.length - 1 && <Divider sx={{mt: 2}} />}
            </Box>
          ))}
        </Box>
      )}
    </Consent>
  );
}
