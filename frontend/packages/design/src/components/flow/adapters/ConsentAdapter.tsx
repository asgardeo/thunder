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

import {Consent, ConsentCheckboxList, type ConsentPurpose, type ConsentRenderProps} from '@asgardeo/react';
import {cn} from '@thunderid/utils';
import {Box, Divider, FormControlLabel, Switch, Typography} from '@wso2/oxygen-ui';
import type {JSX} from 'react';

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
        <Box className={cn('Flow--consent')} sx={{display: 'flex', flexDirection: 'column', gap: 2, mt: 1}}>
          {purposes.map((purpose, idx) => (
            <Box key={purpose.purposeId ?? idx}>
              {purpose.essential && purpose.essential.length > 0 && (
                <Box sx={{mt: 1}}>
                  <Typography className={cn('Text--subtitle2')} variant="subtitle2" fontWeight="bold" sx={{mb: 0.5}}>
                    Essential Attributes
                  </Typography>
                  <ConsentCheckboxList
                    variant="ESSENTIAL"
                    purpose={purpose}
                    formValues={formValues}
                    onInputChange={onInputChange}
                  >
                    {({attributes, isChecked}) => (
                      <Box sx={{display: 'flex', flexDirection: 'column'}}>
                        {attributes.map((attr) => (
                          <Box key={attr} sx={{px: 1}}>
                            <FormControlLabel
                              className={cn('FormControlLabel--root')}
                              control={
                                <Switch
                                  className={cn('Switch--root')}
                                  checked={isChecked(attr)}
                                  disabled
                                  size="small"
                                />
                              }
                              label={
                                <Box sx={{display: 'flex', alignItems: 'center', gap: 1.5}}>
                                  <Box
                                    sx={{
                                      width: 6,
                                      height: 6,
                                      borderRadius: '50%',
                                      backgroundColor: 'text.disabled',
                                      flexShrink: 0,
                                    }}
                                  />
                                  <Typography className={cn('Text--body2')} variant="body2" sx={{fontWeight: 500}}>
                                    {attr}
                                  </Typography>
                                </Box>
                              }
                              labelPlacement="start"
                              sx={{
                                m: 0,
                                width: '100%',
                                justifyContent: 'space-between',
                                py: 0.5,
                              }}
                            />
                            <Divider className={cn('Divider--root')} sx={{opacity: 0.5}} />
                          </Box>
                        ))}
                      </Box>
                    )}
                  </ConsentCheckboxList>
                </Box>
              )}
              {purpose.optional && purpose.optional.length > 0 && (
                <Box sx={{mt: 1}}>
                  <Typography className={cn('Text--subtitle2')} variant="subtitle2" fontWeight="bold" sx={{mb: 0.5}}>
                    Optional Attributes
                  </Typography>
                  <ConsentCheckboxList
                    variant="OPTIONAL"
                    purpose={purpose}
                    formValues={formValues}
                    onInputChange={onInputChange}
                  >
                    {({attributes, isChecked, handleChange}) => (
                      <Box sx={{display: 'flex', flexDirection: 'column'}}>
                        {attributes.map((attr) => (
                          <Box key={attr} sx={{px: 1}}>
                            <FormControlLabel
                              className={cn('FormControlLabel--root')}
                              control={
                                <Switch
                                  className={cn('Switch--root')}
                                  checked={isChecked(attr)}
                                  onChange={(e) => handleChange(attr, (e.target as HTMLInputElement).checked)}
                                  size="small"
                                />
                              }
                              label={
                                <Box sx={{display: 'flex', alignItems: 'center', gap: 1.5}}>
                                  <Box
                                    sx={{
                                      width: 6,
                                      height: 6,
                                      borderRadius: '50%',
                                      backgroundColor: 'text.disabled',
                                      flexShrink: 0,
                                    }}
                                  />
                                  <Typography className={cn('Text--body2')} variant="body2" sx={{fontWeight: 500}}>
                                    {attr}
                                  </Typography>
                                </Box>
                              }
                              labelPlacement="start"
                              sx={{
                                m: 0,
                                width: '100%',
                                justifyContent: 'space-between',
                                py: 0.5,
                              }}
                            />
                            <Divider className={cn('Divider--root')} sx={{opacity: 0.5}} />
                          </Box>
                        ))}
                      </Box>
                    )}
                  </ConsentCheckboxList>
                </Box>
              )}
              {idx < purposes.length - 1 && <Divider className={cn('Divider--root')} sx={{mt: 2}} />}
            </Box>
          ))}
        </Box>
      )}
    </Consent>
  );
}
