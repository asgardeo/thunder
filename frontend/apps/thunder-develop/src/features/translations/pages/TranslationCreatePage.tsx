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

import {Alert, Box, Breadcrumbs, Button, IconButton, LinearProgress, Typography} from '@wso2/oxygen-ui';
import {ChevronRight, X} from '@wso2/oxygen-ui-icons-react';
import {useAsgardeo} from '@asgardeo/react';
import {useQueryClient} from '@tanstack/react-query';
import {I18nQueryKeys, enUS} from '@thunder/i18n';
import type {CountryOption, LocaleOption} from '@thunder/i18n';
import {useConfig} from '@thunder/shared-contexts';
import {useLogger} from '@thunder/logger/react';
import {useCallback, useEffect, useState, type JSX} from 'react';
import {useTranslation} from 'react-i18next';
import {useNavigate} from 'react-router';
import SelectCountry from '../components/create-language/SelectCountry';
import SelectLanguage from '../components/create-language/SelectLanguage';
import ReviewLocaleCode from '../components/create-language/ReviewLocaleCode';
import InitializeLanguage from '../components/create-language/InitializeLanguage';

type CreateStep = 'COUNTRY' | 'LANGUAGE' | 'LOCALE_CODE' | 'INITIALIZE';

const STEPS: CreateStep[] = ['COUNTRY', 'LANGUAGE', 'LOCALE_CODE', 'INITIALIZE'];

export default function TranslationCreatePage(): JSX.Element {
  const {t} = useTranslation('translations');
  const {getServerUrl} = useConfig();
  const serverUrl = getServerUrl();
  const navigate = useNavigate();
  const logger = useLogger('TranslationCreatePage');
  const queryClient = useQueryClient();

  const {http} = useAsgardeo() as unknown as {
    http: {
      request: (config: {
        url: string;
        method: string;
        headers?: Record<string, string>;
        data?: string;
      }) => Promise<{data: unknown}>;
    };
  };

  const [currentStep, setCurrentStep] = useState<CreateStep>('COUNTRY');
  const [stepReady, setStepReady] = useState<Record<CreateStep, boolean>>({
    COUNTRY: false,
    LANGUAGE: false,
    LOCALE_CODE: true,
    INITIALIZE: true,
  });

  const [selectedCountry, setSelectedCountry] = useState<CountryOption | null>(null);
  const [selectedLocale, setSelectedLocale] = useState<LocaleOption | null>(null);
  const [localeCodeOverride, setLocaleCodeOverride] = useState('');
  const [populateFromEnglish, setPopulateFromEnglish] = useState(true);
  const [isCreating, setIsCreating] = useState(false);
  const [progress, setProgress] = useState(0);
  const [error, setError] = useState<string | null>(null);

  // Reset locale when country changes
  useEffect(() => {
    setSelectedLocale(null);
    setStepReady((prev) => ({...prev, LANGUAGE: false}));
  }, [selectedCountry]);

  const localeCode = (localeCodeOverride.trim() || (selectedLocale?.code ?? '')).trim();

  const stepLabels: Record<CreateStep, string> = {
    COUNTRY: t('language.create.steps.country'),
    LANGUAGE: t('language.create.steps.language'),
    LOCALE_CODE: t('language.create.steps.localeCode'),
    INITIALIZE: t('language.create.steps.initialize'),
  };

  const stepProgress = ((STEPS.indexOf(currentStep) + 1) / STEPS.length) * 100;

  const getBreadcrumbSteps = (): CreateStep[] => STEPS.slice(0, STEPS.indexOf(currentStep) + 1);

  const handleReadyChange = useCallback((step: CreateStep, isReady: boolean): void => {
    setStepReady((prev) => ({...prev, [step]: isReady}));
  }, []);

  const handleClose = (): void => {
    (async () => {
      await navigate('/translations');
    })().catch((_error: unknown) => {
      logger.error('Failed to navigate to translations page', {error: _error});
    });
  };

  const handleNext = (): void => {
    const idx = STEPS.indexOf(currentStep);
    if (idx < STEPS.length - 1) {
      if (currentStep === 'LANGUAGE') {
        setLocaleCodeOverride(selectedLocale?.code ?? '');
      }
      setCurrentStep(STEPS[idx + 1]);
    } else {
      void handleCreate();
    }
  };

  const handleBack = (): void => {
    const idx = STEPS.indexOf(currentStep);
    if (idx > 0) setCurrentStep(STEPS[idx - 1]);
  };

  const handleCreate = async (): Promise<void> => {
    if (!localeCode) return;
    setError(null);
    setIsCreating(true);
    setProgress(0);

    const entries: Array<{namespace: string; key: string; value: string}> = [];
    for (const [ns, nsValues] of Object.entries(enUS)) {
      for (const [key, val] of Object.entries(nsValues as Record<string, unknown>)) {
        if (typeof val === 'string') {
          entries.push({namespace: ns, key, value: populateFromEnglish ? val : ''});
        }
      }
    }

    let completed = 0;
    const total = entries.length;

    await Promise.allSettled(
      entries.map(async ({namespace, key, value}) => {
        try {
          await (http as {request: (config: unknown) => Promise<unknown>}).request({
            url: `${serverUrl}/i18n/languages/${localeCode}/translations/ns/${namespace}/keys/${key}`,
            method: 'POST',
            headers: {'Content-Type': 'application/json'},
            data: JSON.stringify({value}),
          });
        } finally {
          completed++;
          setProgress(Math.round((completed / total) * 100));
        }
      }),
    );

    try {
      await queryClient.invalidateQueries({queryKey: [I18nQueryKeys.TRANSLATIONS]});
      await queryClient.invalidateQueries({queryKey: [I18nQueryKeys.LANGUAGES]});
      await navigate(`/translations/${localeCode}`);
    } catch {
      setError(t('language.add.error'));
      setIsCreating(false);
    }
  };

  const renderStepContent = (): JSX.Element | null => {
    switch (currentStep) {
      case 'COUNTRY':
        return (
          <SelectCountry
            selectedCountry={selectedCountry}
            onCountryChange={setSelectedCountry}
            onReadyChange={(ready) => handleReadyChange('COUNTRY', ready)}
          />
        );
      case 'LANGUAGE':
        if (!selectedCountry) return null;
        return (
          <SelectLanguage
            selectedCountry={selectedCountry}
            selectedLocale={selectedLocale}
            onLocaleChange={setSelectedLocale}
            onReadyChange={(ready) => handleReadyChange('LANGUAGE', ready)}
          />
        );
      case 'LOCALE_CODE':
        if (!selectedLocale) return null;
        return (
          <ReviewLocaleCode
            derivedLocale={selectedLocale}
            localeCode={localeCodeOverride}
            onLocaleCodeChange={setLocaleCodeOverride}
            onReadyChange={(ready) => handleReadyChange('LOCALE_CODE', ready)}
          />
        );
      case 'INITIALIZE':
        return (
          <InitializeLanguage
            populateFromEnglish={populateFromEnglish}
            onPopulateChange={setPopulateFromEnglish}
            isCreating={isCreating}
            progress={progress}
          />
        );
      default:
        return null;
    }
  };

  const isFirstStep = currentStep === 'COUNTRY';

  return (
    <Box sx={{height: '100vh', display: 'flex', flexDirection: 'column', overflow: 'hidden'}}>
      <LinearProgress variant="determinate" value={stepProgress} sx={{height: 6, flexShrink: 0}} />

      <Box sx={{flex: 1, display: 'flex', flexDirection: 'column', minHeight: 0}}>
        {/* Header */}
        <Box sx={{p: 4, display: 'flex', alignItems: 'center', gap: 2, flexShrink: 0}}>
          <IconButton
            onClick={handleClose}
            disabled={isCreating}
            sx={{bgcolor: 'background.paper', '&:hover': {bgcolor: 'action.hover'}, boxShadow: 1}}
          >
            <X size={24} />
          </IconButton>
          <Breadcrumbs separator={<ChevronRight size={16} />} aria-label="breadcrumb">
            {getBreadcrumbSteps().map((step, index, array) => {
              const isLast = index === array.length - 1;
              return isLast ? (
                <Typography key={step} variant="h5" color="text.primary">
                  {stepLabels[step]}
                </Typography>
              ) : (
                <Typography
                  key={step}
                  variant="h5"
                  onClick={() => !isCreating && setCurrentStep(step)}
                  sx={{cursor: isCreating ? 'default' : 'pointer'}}
                >
                  {stepLabels[step]}
                </Typography>
              );
            })}
          </Breadcrumbs>
        </Box>

        {/* Left-aligned form content — mirrors ApplicationCreatePage inner layout */}
        <Box
          sx={{
            flex: 1,
            display: 'flex',
            flexDirection: 'column',
            overflowY: 'auto',
            py: 8,
            px: 20,
            alignItems: 'flex-start',
          }}
        >
          <Box sx={{width: '100%', maxWidth: 800, display: 'flex', flexDirection: 'column'}}>
            {error && (
              <Alert severity="error" sx={{mb: 3}} onClose={() => setError(null)}>
                {error}
              </Alert>
            )}

            {renderStepContent()}

            <Box
              sx={{
                mt: 4,
                display: 'flex',
                justifyContent: isFirstStep ? 'flex-start' : 'space-between',
                gap: 2,
              }}
            >
              {!isFirstStep && (
                <Button variant="outlined" onClick={handleBack} sx={{minWidth: 100}} disabled={isCreating}>
                  {t('common:actions.back', {ns: 'common'})}
                </Button>
              )}
              <Button
                variant="contained"
                onClick={handleNext}
                sx={{minWidth: 100}}
                disabled={!stepReady[currentStep] || isCreating}
              >
                {currentStep === 'INITIALIZE'
                  ? t('language.create.createButton')
                  : t('common:actions.continue', {ns: 'common'})}
              </Button>
            </Box>
          </Box>
        </Box>
      </Box>
    </Box>
  );
}
