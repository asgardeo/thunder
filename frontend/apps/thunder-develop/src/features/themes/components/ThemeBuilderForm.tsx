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

import {useEffect, useState, type JSX} from 'react';
import {Controller, useForm} from 'react-hook-form';
import {zodResolver} from '@hookform/resolvers/zod';
import {
  Accordion,
  AccordionDetails,
  AccordionSummary,
  FormControl,
  FormLabel,
  MenuItem,
  Paper,
  Select,
  Stack,
  Tab,
  Tabs,
  TextField,
  Typography,
} from '@wso2/oxygen-ui';
import {ChevronDown} from '@wso2/oxygen-ui-icons-react';
import {useTranslation} from 'react-i18next';
import ColorPickerField from '@/features/design-system/components/ColorPickerField';
import {themeFormSchema, type ThemeFormData} from '../schemas/themeSchema';

/**
 * Props for ThemeBuilderForm component
 */
interface ThemeBuilderFormProps {
  /** Initial form values */
  initialValues: ThemeFormData;
  /** Callback when form is submitted */
  onSubmit: (data: ThemeFormData) => void;
  /** Callback when form values change (for preview) */
  onValuesChange?: (data: ThemeFormData) => void;
  /** Callback when isDirty state changes */
  onDirtyChange?: (isDirty: boolean) => void;
  /** Whether form is currently submitting */
  isSubmitting?: boolean;
}

/**
 * Form component for theme builder with react-hook-form and color pickers.
 * Supports editing both light and dark color schemes.
 */
export default function ThemeBuilderForm({
  initialValues,
  onSubmit,
  onValuesChange,
  onDirtyChange,
  isSubmitting = false,
}: ThemeBuilderFormProps): JSX.Element {
  const {t} = useTranslation();
  const [activeColorScheme, setActiveColorScheme] = useState<'light' | 'dark'>('light');

  const {
    control,
    handleSubmit,
    watch,
    formState: {errors, isDirty},
  } = useForm<ThemeFormData>({
    resolver: zodResolver(themeFormSchema),
    defaultValues: initialValues,
    mode: 'onChange',
  });

  // Watch all values for preview
  const formValues = watch();

  // Notify parent of form value changes
  useEffect(() => {
    onValuesChange?.(formValues);
  }, [formValues, onValuesChange]);

  // Notify parent of isDirty changes
  useEffect(() => {
    onDirtyChange?.(isDirty);
  }, [isDirty, onDirtyChange]);

  // Get background color for contrast checking
  const backgroundColor =
    formValues.theme.colorSchemes[activeColorScheme].colors.background.default;

  return (
    <Paper elevation={2} sx={{p: 3}}>
      <form onSubmit={handleSubmit(onSubmit)}>
        <Stack spacing={3}>
          {/* Display Name */}
          <Controller
            name="displayName"
            control={control}
            render={({field}) => (
              <TextField
                {...field}
                fullWidth
                label={t('themes:builder.displayName', {defaultValue: 'Display Name'})}
                error={!!errors.displayName}
                helperText={errors.displayName?.message}
                required
              />
            )}
          />

          {/* General Settings */}
          <Accordion defaultExpanded>
            <AccordionSummary expandIcon={<ChevronDown size={18} />}>
              <Typography variant="subtitle1">
                {t('themes:builder.general', {defaultValue: 'General Settings'})}
              </Typography>
            </AccordionSummary>
            <AccordionDetails>
              <Stack spacing={2}>
                <Controller
                  name="theme.direction"
                  control={control}
                  render={({field}) => (
                    <FormControl fullWidth>
                      <FormLabel>{t('themes:builder.direction', {defaultValue: 'Direction'})}</FormLabel>
                      <Select {...field} size="small">
                        <MenuItem value="ltr">Left to Right (LTR)</MenuItem>
                        <MenuItem value="rtl">Right to Left (RTL)</MenuItem>
                      </Select>
                    </FormControl>
                  )}
                />
                <Controller
                  name="theme.defaultColorScheme"
                  control={control}
                  render={({field}) => (
                    <FormControl fullWidth>
                      <FormLabel>
                        {t('themes:builder.defaultColorScheme', {defaultValue: 'Default Color Scheme'})}
                      </FormLabel>
                      <Select {...field} size="small">
                        <MenuItem value="light">Light</MenuItem>
                        <MenuItem value="dark">Dark</MenuItem>
                      </Select>
                    </FormControl>
                  )}
                />
              </Stack>
            </AccordionDetails>
          </Accordion>

          {/* Color Scheme Tabs */}
          <Paper variant="outlined" sx={{p: 2}}>
            <Tabs
              value={activeColorScheme}
              onChange={(_, newValue) => setActiveColorScheme(newValue as 'light' | 'dark')}
              sx={{borderBottom: 1, borderColor: 'divider', mb: 2}}
            >
              <Tab label={t('themes:builder.colorSchemes.light', {defaultValue: 'Light Mode'})} value="light" />
              <Tab label={t('themes:builder.colorSchemes.dark', {defaultValue: 'Dark Mode'})} value="dark" />
            </Tabs>

            <Stack spacing={3}>
              {/* Primary Color */}
              <Accordion>
                <AccordionSummary expandIcon={<ChevronDown size={18} />}>
                  <Typography variant="subtitle1">
                    {t('themes:builder.primaryColor', {defaultValue: 'Primary Color'})}
                  </Typography>
                </AccordionSummary>
                <AccordionDetails>
                  <Stack spacing={2}>
                    <ColorPickerField
                      name={`theme.colorSchemes.${activeColorScheme}.colors.primary.main`}
                      label="Main"
                      control={control}
                      required
                      contrastCheckAgainst={backgroundColor}
                    />
                    <ColorPickerField
                      name={`theme.colorSchemes.${activeColorScheme}.colors.primary.dark`}
                      label="Dark"
                      control={control}
                      required
                    />
                    <ColorPickerField
                      name={`theme.colorSchemes.${activeColorScheme}.colors.primary.contrastText`}
                      label="Contrast Text"
                      control={control}
                      required
                      contrastCheckAgainst={
                        formValues.theme.colorSchemes[activeColorScheme].colors.primary.main
                      }
                    />
                  </Stack>
                </AccordionDetails>
              </Accordion>

              {/* Secondary Color */}
              <Accordion>
                <AccordionSummary expandIcon={<ChevronDown size={18} />}>
                  <Typography variant="subtitle1">
                    {t('themes:builder.secondaryColor', {defaultValue: 'Secondary Color'})}
                  </Typography>
                </AccordionSummary>
                <AccordionDetails>
                  <Stack spacing={2}>
                    <ColorPickerField
                      name={`theme.colorSchemes.${activeColorScheme}.colors.secondary.main`}
                      label="Main"
                      control={control}
                      required
                      contrastCheckAgainst={backgroundColor}
                    />
                    <ColorPickerField
                      name={`theme.colorSchemes.${activeColorScheme}.colors.secondary.dark`}
                      label="Dark"
                      control={control}
                      required
                    />
                    <ColorPickerField
                      name={`theme.colorSchemes.${activeColorScheme}.colors.secondary.contrastText`}
                      label="Contrast Text"
                      control={control}
                      required
                      contrastCheckAgainst={
                        formValues.theme.colorSchemes[activeColorScheme].colors.secondary.main
                      }
                    />
                  </Stack>
                </AccordionDetails>
              </Accordion>

              {/* Background */}
              <Accordion>
                <AccordionSummary expandIcon={<ChevronDown size={18} />}>
                  <Typography variant="subtitle1">
                    {t('themes:builder.background', {defaultValue: 'Background'})}
                  </Typography>
                </AccordionSummary>
                <AccordionDetails>
                  <Stack spacing={2}>
                    <ColorPickerField
                      name={`theme.colorSchemes.${activeColorScheme}.colors.background.default`}
                      label="Default"
                      control={control}
                      required
                    />
                    <ColorPickerField
                      name={`theme.colorSchemes.${activeColorScheme}.colors.background.paper`}
                      label="Paper"
                      control={control}
                      required
                    />
                  </Stack>
                </AccordionDetails>
              </Accordion>

              {/* Text Colors */}
              <Accordion>
                <AccordionSummary expandIcon={<ChevronDown size={18} />}>
                  <Typography variant="subtitle1">
                    {t('themes:builder.textColors', {defaultValue: 'Text Colors'})}
                  </Typography>
                </AccordionSummary>
                <AccordionDetails>
                  <Stack spacing={2}>
                    <ColorPickerField
                      name={`theme.colorSchemes.${activeColorScheme}.colors.text.primary`}
                      label="Primary"
                      control={control}
                      required
                      contrastCheckAgainst={backgroundColor}
                    />
                    <ColorPickerField
                      name={`theme.colorSchemes.${activeColorScheme}.colors.text.secondary`}
                      label="Secondary"
                      control={control}
                      required
                      contrastCheckAgainst={backgroundColor}
                    />
                  </Stack>
                </AccordionDetails>
              </Accordion>
            </Stack>
          </Paper>

          {/* Typography */}
          <Accordion>
            <AccordionSummary expandIcon={<ChevronDown size={18} />}>
              <Typography variant="subtitle1">
                {t('themes:builder.typography', {defaultValue: 'Typography'})}
              </Typography>
            </AccordionSummary>
            <AccordionDetails>
              <Controller
                name="theme.typography.fontFamily"
                control={control}
                render={({field}) => (
                  <TextField
                    {...field}
                    fullWidth
                    label="Font Family"
                    error={!!errors.theme?.typography?.fontFamily}
                    helperText={errors.theme?.typography?.fontFamily?.message}
                    placeholder="Inter, sans-serif"
                  />
                )}
              />
            </AccordionDetails>
          </Accordion>

          {/* Shape */}
          <Accordion>
            <AccordionSummary expandIcon={<ChevronDown size={18} />}>
              <Typography variant="subtitle1">{t('themes:builder.shape', {defaultValue: 'Shape'})}</Typography>
            </AccordionSummary>
            <AccordionDetails>
              <Controller
                name="theme.shape.borderRadius"
                control={control}
                render={({field}) => (
                  <TextField
                    {...field}
                    fullWidth
                    label="Border Radius"
                    error={!!errors.theme?.shape?.borderRadius}
                    helperText={errors.theme?.shape?.borderRadius?.message || 'Format: 8px, 12px, etc.'}
                    placeholder="8px"
                  />
                )}
              />
            </AccordionDetails>
          </Accordion>
        </Stack>
      </form>
    </Paper>
  );
}
