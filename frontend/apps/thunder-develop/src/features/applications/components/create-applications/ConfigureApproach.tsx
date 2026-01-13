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

import {
  Box,
  Typography,
  Stack,
  Card,
  CardContent,
  Radio,
  RadioGroup,
  FormControlLabel,
  CardActionArea,
} from '@wso2/oxygen-ui';
import {ExternalLink, Code} from '@wso2/oxygen-ui-icons-react';
import type {JSX, ChangeEvent} from 'react';
import {useEffect} from 'react';
import {useTranslation} from 'react-i18next';
import {ApplicationCreateFlowSignInApproach} from '../../models/application-create-flow';

/**
 * Props for the {@link ConfigureApproach} component.
 *
 * @public
 */
export interface ConfigureApproachProps {
  /**
   * Currently selected sign-in approach
   */
  selectedApproach: ApplicationCreateFlowSignInApproach;

  /**
   * Callback function when the sign-in approach changes
   */
  onApproachChange: (approach: ApplicationCreateFlowSignInApproach) => void;

  /**
   * Callback function to broadcast whether this step is ready to proceed
   */
  onReadyChange?: (isReady: boolean) => void;
}

/**
 * React component that renders the sign-in approach selection step in the
 * application creation onboarding flow.
 *
 * This component allows users to choose between two authentication approaches:
 * 1. Inbuilt - Uses Thunder's hosted login pages for authentication
 * 2. Custom - Uses native/custom UI with Thunder as the authentication API
 *
 * The component displays two selectable cards with radio buttons, providing
 * clear descriptions of each approach. It automatically marks the step as ready
 * since a default selection is always available.
 *
 * @param props - The component props
 * @param props.selectedApproach - The currently selected sign-in approach (inbuilt or custom)
 * @param props.onApproachChange - Callback invoked when user selects a different approach
 * @param props.onReadyChange - Optional callback to notify parent that step is ready to proceed
 *
 * @returns JSX element displaying the sign-in approach selection interface
 *
 * @example
 * ```tsx
 * import ConfigureApproach from './ConfigureApproach';
 *
 * function OnboardingFlow() {
 *   const [approach, setApproach] = useState(
 *     ApplicationCreateFlowSignInApproach.INBUILT
 *   );
 *
 *   return (
 *     <ConfigureApproach
 *       selectedApproach={approach}
 *       onApproachChange={setApproach}
 *       onReadyChange={(isReady) => console.log('Step ready:', isReady)}
 *     />
 *   );
 * }
 * ```
 *
 * @public
 */
export default function ConfigureApproach({
  selectedApproach,
  onApproachChange,
  onReadyChange = undefined,
}: ConfigureApproachProps): JSX.Element {
  const {t} = useTranslation();

  /**
   * Broadcast readiness on mount since this step is always ready.
   */
  useEffect((): void => {
    onReadyChange?.(true);
  }, [onReadyChange]);

  const handleApproachChange = (event: ChangeEvent<HTMLInputElement>): void => {
    onApproachChange(event.target.value as ApplicationCreateFlowSignInApproach);
  };

  return (
    <Stack direction="column" spacing={3}>
      <Stack direction="column" spacing={1}>
        <Typography variant="h1" gutterBottom>
          {t('applications:onboarding.configure.approach.title')}
        </Typography>
        <Typography variant="subtitle1" gutterBottom>
          {t('applications:onboarding.configure.approach.subtitle')}
        </Typography>
      </Stack>

      <RadioGroup value={selectedApproach} onChange={handleApproachChange}>
        <Stack direction="column" spacing={2}>
          {/* Hosted Pages Option */}
          <Card variant="outlined" onClick={() => onApproachChange(ApplicationCreateFlowSignInApproach.INBUILT)}>
            <CardActionArea
              sx={{
                height: '100%',
                cursor: 'pointer',
                border: 1,
                borderColor:
                  selectedApproach === ApplicationCreateFlowSignInApproach.INBUILT ? 'primary.main' : 'divider',
                transition: 'all 0.2s ease-in-out',
                '&:hover': {
                  borderColor: 'primary.main',
                  bgcolor:
                    selectedApproach === ApplicationCreateFlowSignInApproach.INBUILT
                      ? 'action.selected'
                      : 'action.hover',
                },
              }}
            >
              <CardContent>
                <Stack direction="row" spacing={2} alignItems="flex-start">
                  <FormControlLabel
                    value={ApplicationCreateFlowSignInApproach.INBUILT}
                    control={<Radio />}
                    label=""
                    sx={{m: 0}}
                    onClick={(e) => e.stopPropagation()}
                  />
                  <Box sx={{flex: 1}}>
                    <Stack direction="row" spacing={1} alignItems="center" sx={{mb: 1}}>
                      <ExternalLink size={20} />
                      <Typography variant="h6">
                        {t('applications:onboarding.configure.approach.inbuilt.title', {
                          product: t('common:product.displayName'),
                        })}
                      </Typography>
                    </Stack>
                    <Typography variant="body2" color="text.secondary">
                      {t('applications:onboarding.configure.approach.inbuilt.description', {
                        product: t('common:product.displayName'),
                      })}
                    </Typography>
                  </Box>
                </Stack>
              </CardContent>
            </CardActionArea>
          </Card>

          {/* Native/Custom UI Option */}
          <Card variant="outlined" onClick={() => onApproachChange(ApplicationCreateFlowSignInApproach.CUSTOM)}>
            <CardActionArea
              sx={{
                height: '100%',
                cursor: 'pointer',
                border: 1,
                borderColor:
                  selectedApproach === ApplicationCreateFlowSignInApproach.CUSTOM ? 'primary.main' : 'divider',
                transition: 'all 0.2s ease-in-out',
                '&:hover': {
                  borderColor: 'primary.main',
                  bgcolor:
                    selectedApproach === ApplicationCreateFlowSignInApproach.CUSTOM
                      ? 'action.selected'
                      : 'action.hover',
                },
              }}
            >
              <CardContent>
                <Stack direction="row" spacing={2} alignItems="flex-start">
                  <FormControlLabel
                    value={ApplicationCreateFlowSignInApproach.CUSTOM}
                    control={<Radio />}
                    label=""
                    sx={{m: 0}}
                    onClick={(e) => e.stopPropagation()}
                  />
                  <Box sx={{flex: 1}}>
                    <Stack direction="row" spacing={1} alignItems="center" sx={{mb: 1}}>
                      <Code size={20} />
                      <Typography variant="h6">
                        {t('applications:onboarding.configure.approach.native.title', {
                          product: t('common:product.displayName'),
                        })}
                      </Typography>
                    </Stack>
                    <Typography variant="body2" color="text.secondary">
                      {t('applications:onboarding.configure.approach.native.description', {
                        product: t('common:product.displayName'),
                      })}
                    </Typography>
                  </Box>
                </Stack>
              </CardContent>
            </CardActionArea>
          </Card>
        </Stack>
      </RadioGroup>
    </Stack>
  );
}
