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

import {useEffect, type JSX} from 'react';
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
  Switch,
  TextField,
  Typography,
} from '@wso2/oxygen-ui';
import {ChevronDown} from '@wso2/oxygen-ui-icons-react';
import {useTranslation} from 'react-i18next';
import ColorPickerField from '@/features/design-system/components/ColorPickerField';
import {layoutFormSchema, type LayoutFormData} from '../schemas/layoutSchema';

/**
 * Props for LayoutBuilderForm component
 */
interface LayoutBuilderFormProps {
  /** Initial form values */
  initialValues: LayoutFormData;
  /** Callback when form is submitted */
  onSubmit: (data: LayoutFormData) => void;
  /** Callback when form values change (for preview) */
  onValuesChange?: (data: LayoutFormData) => void;
  /** Callback when isDirty state changes */
  onDirtyChange?: (isDirty: boolean) => void;
  /** Whether form is currently submitting */
  isSubmitting?: boolean;
}

/**
 * Form component for layout builder with react-hook-form and validation.
 */
export default function LayoutBuilderForm({
  initialValues,
  onSubmit,
  onValuesChange,
  onDirtyChange,
  isSubmitting = false,
}: LayoutBuilderFormProps): JSX.Element {
  const {t} = useTranslation();

  const {
    control,
    handleSubmit,
    watch,
    formState: {errors, isDirty},
  } = useForm<LayoutFormData>({
    resolver: zodResolver(layoutFormSchema),
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
                label={t('layouts:builder.displayName', {defaultValue: 'Display Name'})}
                error={!!errors.displayName}
                helperText={errors.displayName?.message}
                required
              />
            )}
          />

          {/* Template */}
          <Accordion defaultExpanded>
            <AccordionSummary expandIcon={<ChevronDown size={18} />}>
              <Typography variant="subtitle1">{t('layouts:builder.template', {defaultValue: 'Template'})}</Typography>
            </AccordionSummary>
            <AccordionDetails>
              <Controller
                name="layout.screens.signin.template"
                control={control}
                render={({field}) => (
                  <FormControl fullWidth>
                    <Select {...field} size="small">
                      <MenuItem value="centered">Centered</MenuItem>
                      <MenuItem value="split">Split</MenuItem>
                      <MenuItem value="fullscreen">Fullscreen</MenuItem>
                    </Select>
                  </FormControl>
                )}
              />
            </AccordionDetails>
          </Accordion>

          {/* Background */}
          <Accordion>
            <AccordionSummary expandIcon={<ChevronDown size={18} />}>
              <Typography variant="subtitle1">
                {t('layouts:builder.background', {defaultValue: 'Background'})}
              </Typography>
            </AccordionSummary>
            <AccordionDetails>
              <Stack spacing={2}>
                <Controller
                  name="layout.screens.signin.background.type"
                  control={control}
                  render={({field}) => (
                    <FormControl fullWidth>
                      <FormLabel>Type</FormLabel>
                      <Select {...field} size="small">
                        <MenuItem value="gradient">Gradient</MenuItem>
                        <MenuItem value="solid">Solid Color</MenuItem>
                        <MenuItem value="image">Image URL</MenuItem>
                      </Select>
                    </FormControl>
                  )}
                />
                <Controller
                  name="layout.screens.signin.background.value"
                  control={control}
                  render={({field}) => (
                    <TextField
                      {...field}
                      fullWidth
                      label="Value"
                      error={!!errors.layout?.screens?.signin?.background?.value}
                      helperText={errors.layout?.screens?.signin?.background?.value?.message}
                      placeholder={
                        formValues.layout.screens.signin?.background?.type === 'gradient'
                          ? 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)'
                          : formValues.layout.screens.signin?.background?.type === 'solid'
                            ? '#FFFFFF'
                            : 'https://example.com/image.jpg'
                      }
                    />
                  )}
                />
              </Stack>
            </AccordionDetails>
          </Accordion>

          {/* Container */}
          <Accordion>
            <AccordionSummary expandIcon={<ChevronDown size={18} />}>
              <Typography variant="subtitle1">{t('layouts:builder.container', {defaultValue: 'Container'})}</Typography>
            </AccordionSummary>
            <AccordionDetails>
              <Stack spacing={2}>
                <Controller
                  name="layout.screens.signin.container.maxWidth"
                  control={control}
                  render={({field}) => (
                    <TextField
                      {...field}
                      fullWidth
                      type="number"
                      label="Max Width (px)"
                      error={!!errors.layout?.screens?.signin?.container?.maxWidth}
                      helperText={
                        errors.layout?.screens?.signin?.container?.maxWidth?.message || 'Range: 300-1200px'
                      }
                      onChange={(e) => field.onChange(Number(e.target.value))}
                    />
                  )}
                />
                <Controller
                  name="layout.screens.signin.container.padding"
                  control={control}
                  render={({field}) => (
                    <TextField
                      {...field}
                      fullWidth
                      type="number"
                      label="Padding (px)"
                      error={!!errors.layout?.screens?.signin?.container?.padding}
                      helperText={errors.layout?.screens?.signin?.container?.padding?.message || 'Range: 0-100px'}
                      onChange={(e) => field.onChange(Number(e.target.value))}
                    />
                  )}
                />
                <Controller
                  name="layout.screens.signin.container.borderRadius"
                  control={control}
                  render={({field}) => (
                    <TextField
                      {...field}
                      fullWidth
                      type="number"
                      label="Border Radius (px)"
                      error={!!errors.layout?.screens?.signin?.container?.borderRadius}
                      helperText={
                        errors.layout?.screens?.signin?.container?.borderRadius?.message || 'Range: 0-50px'
                      }
                      onChange={(e) => field.onChange(Number(e.target.value))}
                    />
                  )}
                />
                <Controller
                  name="layout.screens.signin.container.elevation"
                  control={control}
                  render={({field}) => (
                    <TextField
                      {...field}
                      fullWidth
                      type="number"
                      label="Elevation"
                      error={!!errors.layout?.screens?.signin?.container?.elevation}
                      helperText={errors.layout?.screens?.signin?.container?.elevation?.message || 'Range: 0-24'}
                      onChange={(e) => field.onChange(Number(e.target.value))}
                    />
                  )}
                />
                <ColorPickerField
                  name="layout.screens.signin.container.background"
                  label="Background Color"
                  control={control}
                  required
                />
              </Stack>
            </AccordionDetails>
          </Accordion>

          {/* Spacing */}
          <Accordion>
            <AccordionSummary expandIcon={<ChevronDown size={18} />}>
              <Typography variant="subtitle1">{t('layouts:builder.spacing', {defaultValue: 'Spacing'})}</Typography>
            </AccordionSummary>
            <AccordionDetails>
              <Stack spacing={2}>
                <Controller
                  name="layout.screens.signin.spacing.componentGap"
                  control={control}
                  render={({field}) => (
                    <TextField
                      {...field}
                      fullWidth
                      type="number"
                      label="Component Gap (px)"
                      error={!!errors.layout?.screens?.signin?.spacing?.componentGap}
                      helperText={
                        errors.layout?.screens?.signin?.spacing?.componentGap?.message || 'Range: 0-100px'
                      }
                      onChange={(e) => field.onChange(Number(e.target.value))}
                    />
                  )}
                />
                <Controller
                  name="layout.screens.signin.spacing.sectionGap"
                  control={control}
                  render={({field}) => (
                    <TextField
                      {...field}
                      fullWidth
                      type="number"
                      label="Section Gap (px)"
                      error={!!errors.layout?.screens?.signin?.spacing?.sectionGap}
                      helperText={errors.layout?.screens?.signin?.spacing?.sectionGap?.message || 'Range: 0-100px'}
                      onChange={(e) => field.onChange(Number(e.target.value))}
                    />
                  )}
                />
              </Stack>
            </AccordionDetails>
          </Accordion>

          {/* Header */}
          <Accordion>
            <AccordionSummary expandIcon={<ChevronDown size={18} />}>
              <Typography variant="subtitle1">{t('layouts:builder.header', {defaultValue: 'Header'})}</Typography>
            </AccordionSummary>
            <AccordionDetails>
              <Stack spacing={2}>
                <Controller
                  name="layout.screens.signin.header.enabled"
                  control={control}
                  render={({field}) => (
                    <Stack direction="row" justifyContent="space-between" alignItems="center">
                      <Typography>Enabled</Typography>
                      <Switch checked={field.value} onChange={field.onChange} />
                    </Stack>
                  )}
                />
                <Controller
                  name="layout.screens.signin.header.height"
                  control={control}
                  render={({field}) => (
                    <TextField
                      {...field}
                      fullWidth
                      type="number"
                      label="Height (px)"
                      error={!!errors.layout?.screens?.signin?.header?.height}
                      helperText={errors.layout?.screens?.signin?.header?.height?.message || 'Range: 0-200px'}
                      onChange={(e) => field.onChange(Number(e.target.value))}
                    />
                  )}
                />
                <Controller
                  name="layout.screens.signin.header.padding"
                  control={control}
                  render={({field}) => (
                    <TextField
                      {...field}
                      fullWidth
                      type="number"
                      label="Padding (px)"
                      error={!!errors.layout?.screens?.signin?.header?.padding}
                      helperText={errors.layout?.screens?.signin?.header?.padding?.message || 'Range: 0-50px'}
                      onChange={(e) => field.onChange(Number(e.target.value))}
                    />
                  )}
                />
                <Controller
                  name="layout.screens.signin.header.showLogo"
                  control={control}
                  render={({field}) => (
                    <Stack direction="row" justifyContent="space-between" alignItems="center">
                      <Typography>Show Logo</Typography>
                      <Switch checked={field.value} onChange={field.onChange} />
                    </Stack>
                  )}
                />
                <Controller
                  name="layout.screens.signin.header.showLanguageSelector"
                  control={control}
                  render={({field}) => (
                    <Stack direction="row" justifyContent="space-between" alignItems="center">
                      <Typography>Show Language Selector</Typography>
                      <Switch checked={field.value} onChange={field.onChange} />
                    </Stack>
                  )}
                />
                <Controller
                  name="layout.screens.signin.header.showBackButton"
                  control={control}
                  render={({field}) => (
                    <Stack direction="row" justifyContent="space-between" alignItems="center">
                      <Typography>Show Back Button</Typography>
                      <Switch checked={field.value} onChange={field.onChange} />
                    </Stack>
                  )}
                />
              </Stack>
            </AccordionDetails>
          </Accordion>

          {/* Footer */}
          <Accordion>
            <AccordionSummary expandIcon={<ChevronDown size={18} />}>
              <Typography variant="subtitle1">{t('layouts:builder.footer', {defaultValue: 'Footer'})}</Typography>
            </AccordionSummary>
            <AccordionDetails>
              <Stack spacing={2}>
                <Controller
                  name="layout.screens.signin.footer.enabled"
                  control={control}
                  render={({field}) => (
                    <Stack direction="row" justifyContent="space-between" alignItems="center">
                      <Typography>Enabled</Typography>
                      <Switch checked={field.value} onChange={field.onChange} />
                    </Stack>
                  )}
                />
                <Controller
                  name="layout.screens.signin.footer.height"
                  control={control}
                  render={({field}) => (
                    <TextField
                      {...field}
                      fullWidth
                      type="number"
                      label="Height (px)"
                      error={!!errors.layout?.screens?.signin?.footer?.height}
                      helperText={errors.layout?.screens?.signin?.footer?.height?.message || 'Range: 0-200px'}
                      onChange={(e) => field.onChange(Number(e.target.value))}
                    />
                  )}
                />
                <Controller
                  name="layout.screens.signin.footer.padding"
                  control={control}
                  render={({field}) => (
                    <TextField
                      {...field}
                      fullWidth
                      type="number"
                      label="Padding (px)"
                      error={!!errors.layout?.screens?.signin?.footer?.padding}
                      helperText={errors.layout?.screens?.signin?.footer?.padding?.message || 'Range: 0-50px'}
                      onChange={(e) => field.onChange(Number(e.target.value))}
                    />
                  )}
                />
                <Controller
                  name="layout.screens.signin.footer.showLinks"
                  control={control}
                  render={({field}) => (
                    <Stack direction="row" justifyContent="space-between" alignItems="center">
                      <Typography>Show Links</Typography>
                      <Switch checked={field.value} onChange={field.onChange} />
                    </Stack>
                  )}
                />
              </Stack>
            </AccordionDetails>
          </Accordion>
        </Stack>
      </form>
    </Paper>
  );
}
