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

import {Box, Typography, Stack, Card, CardContent, Divider, Paper, IconButton, Button, Tooltip} from '@wso2/oxygen-ui';
import {Sparkles, Copy} from '@wso2/oxygen-ui-icons-react';
import type {JSX} from 'react';
import {useState} from 'react';
import {useTranslation} from 'react-i18next';
import {Prism as SyntaxHighlighter} from 'react-syntax-highlighter';
import {vscDarkPlus} from 'react-syntax-highlighter/dist/esm/styles/prism';
import type {IntegrationGuides, IntegrationStep} from '../../models/application-templates';
import {ApplicationCreateFlowSignInApproach} from '../../models/application-create-flow';

/**
 * Props for the {@link TechnologyGuide} component.
 *
 * @public
 */
export interface TechnologyGuideProps {
  /**
   * Integration guides structure containing LLM prompt and manual steps
   */
  guides: IntegrationGuides | null;
  /**
   * The selected sign-in approach (INBUILT or CUSTOM)
   */
  signInApproach: ApplicationCreateFlowSignInApproach;
  /**
   * The OAuth2 client ID to replace {{clientId}} placeholders
   */
  clientId?: string;
  /**
   * The application ID to replace {{applicationId}} placeholders
   */
  applicationId?: string;
}

/**
 * React component that displays integration guide options for technology templates.
 *
 * This component renders:
 * 1. LLM prompt option as a clickable card
 * 2. Divider with "or" text
 * 3. Step-by-step integration guide using custom timeline layout
 *
 * The displayed steps vary based on the sign-in approach:
 * - INBUILT: Shows SDK integration steps for Thunder-hosted login
 * - CUSTOM: Shows API integration steps for custom login implementation
 *
 * @param props - The component props
 * @param props.guides - Integration guides structure
 * @param props.onLLMGuideSelect - Callback invoked when user selects LLM guide
 * @param props.signInApproach - The selected sign-in approach
 *
 * @returns JSX element displaying the integration guide options
 *
 * @public
 */
export default function TechnologyGuide({
  guides,
  signInApproach,
  clientId = '',
  applicationId = '',
}: TechnologyGuideProps): JSX.Element | null {
  const {t} = useTranslation();
  const [copiedStep, setCopiedStep] = useState<number | null>(null);
  const [copiedPrompt, setCopiedPrompt] = useState<boolean>(false);

  // Don't render anything if there are no guides
  if (!guides) {
    return null;
  }

  // Get the guide key based on sign-in approach
  const guideKey = signInApproach === ApplicationCreateFlowSignInApproach.INBUILT ? 'inbuilt' : 'custom';
  const selectedGuide = guides[guideKey];

  if (!selectedGuide) {
    return null;
  }

  const {llm_prompt: llmPrompt, manual_steps: manualSteps} = selectedGuide;

  /**
   * Replace placeholders in text with actual values
   * If clientId or applicationId is not available, keep the placeholder
   */
  const replacePlaceholders = (text: string): string => {
    let result = text;

    // Replace clientId if available
    if (clientId && clientId.trim() !== '') {
      result = result.replace(/\{\{clientId\}\}/g, clientId);
    }

    // Replace applicationId if available
    if (applicationId && applicationId.trim() !== '') {
      result = result.replace(/\{\{applicationId\}\}/g, applicationId);
    }

    return result;
  };

  const handleCopyPrompt = async (e: React.MouseEvent): Promise<void> => {
    e.stopPropagation();
    if (!llmPrompt.content) return;

    const contentWithReplacements = replacePlaceholders(llmPrompt.content);

    try {
      await navigator.clipboard.writeText(contentWithReplacements);
      setCopiedPrompt(true);
      setTimeout(() => setCopiedPrompt(false), 2000);
    } catch {
      // Fallback for older browsers
      const textArea = document.createElement('textarea');
      textArea.value = contentWithReplacements;
      textArea.style.position = 'fixed';
      textArea.style.opacity = '0';
      document.body.appendChild(textArea);
      textArea.select();
      try {
        document.execCommand('copy');
        setCopiedPrompt(true);
        setTimeout(() => setCopiedPrompt(false), 2000);
      } catch {
        // Ignore copy errors
      }
      document.body.removeChild(textArea);
    }
  };

  const handleCopyCode = async (code: string, stepNumber: number): Promise<void> => {
    const codeWithReplacements = replacePlaceholders(code);

    try {
      await navigator.clipboard.writeText(codeWithReplacements);
      setCopiedStep(stepNumber);
      setTimeout(() => setCopiedStep(null), 2000);
    } catch {
      // Fallback for older browsers
      const textArea = document.createElement('textarea');
      textArea.value = codeWithReplacements;
      textArea.style.position = 'fixed';
      textArea.style.opacity = '0';
      document.body.appendChild(textArea);
      textArea.select();
      try {
        document.execCommand('copy');
        setCopiedStep(stepNumber);
        setTimeout(() => setCopiedStep(null), 2000);
      } catch {
        // Ignore copy errors
      }
      document.body.removeChild(textArea);
    }
  };

  const renderCodeBlock = (step: IntegrationStep): JSX.Element | null => {
    if (!step.code) {
      return null;
    }

    // Map language aliases to Prism language identifiers
    const getLanguage = (lang: string): string => {
      const languageMap: Record<string, string> = {
        terminal: 'bash',
        '.env': 'properties',
        typescript: 'tsx',
      };
      return languageMap[lang] || lang;
    };

    const codeWithReplacements = replacePlaceholders(step.code.content);

    return (
      <Paper
        variant="outlined"
        sx={{
          mt: 1.5,
          bgcolor: 'transparent',
          p: 0,
          overflow: 'hidden',
          position: 'relative',
        }}
      >
        {step.code.filename && (
          <Box
            sx={{
              px: 2,
              py: 1,
              bgcolor: 'grey.900',
              borderBottom: 1,
              borderColor: 'grey.800',
            }}
          >
            <Typography variant="caption" sx={{fontFamily: 'monospace', color: 'grey.300'}}>
              {step.code.filename}
            </Typography>
          </Box>
        )}
        <Box sx={{position: 'relative'}}>
          <SyntaxHighlighter
            language={getLanguage(step.code.language)}
            style={vscDarkPlus}
            customStyle={{
              margin: 0,
              padding: '16px',
              fontSize: '0.875rem',
              lineHeight: 1.6,
              backgroundColor: '#1E1E1E',
              borderRadius: step.code.filename ? '0 0 4px 4px' : '4px',
            }}
            showLineNumbers={false}
            wrapLines
          >
            {codeWithReplacements}
          </SyntaxHighlighter>
          <IconButton
            size="small"
            onClick={() => {
              handleCopyCode(step.code!.content, step.step).catch(() => {
                // Error already handled
              });
            }}
            sx={{
              position: 'absolute',
              top: 8,
              right: 8,
              color: 'grey.400',
              bgcolor: 'grey.800',
              '&:hover': {
                bgcolor: 'grey.700',
              },
            }}
          >
            <Copy size={16} />
          </IconButton>
          {copiedStep === step.step && (
            <Typography
              variant="caption"
              sx={{
                position: 'absolute',
                top: 8,
                right: 48,
                color: 'success.main',
                bgcolor: 'grey.800',
                px: 1,
                py: 0.5,
                borderRadius: 1,
              }}
            >
              {t('applications:clientSecret.copied')}
            </Typography>
          )}
        </Box>
      </Paper>
    );
  };

  return (
    <Stack direction="column" spacing={3} sx={{width: '100%'}}>
      {/* LLM Prompt Option */}
      {llmPrompt && (
        <Card
          variant="outlined"
          sx={{
            position: 'relative',
            background:
              'linear-gradient(white, white) padding-box, linear-gradient(135deg, #667eea 0%, #764ba2 25%, #f093fb 50%, #4facfe 75%, #00f2fe 100%) border-box',
            border: '2px solid transparent',
            borderRadius: 2,
          }}
        >
          <CardContent sx={{p: 3}}>
            <Stack direction="row" spacing={2} alignItems="flex-start">
              <Box
                sx={{
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'center',
                  width: 48,
                  height: 48,
                  borderRadius: 1,
                  color: 'primary.main',
                }}
              >
                <Sparkles size={24} />
              </Box>
              <Box sx={{flex: 1}}>
                <Typography variant="subtitle1" sx={{mb: 0.5, fontWeight: 600}}>
                  {llmPrompt.title}
                </Typography>
                <Typography variant="body2" color="text.secondary">
                  {llmPrompt.description}
                </Typography>
              </Box>
              {llmPrompt.content && (
                <Tooltip title={copiedPrompt ? t('applications:clientSecret.copied') : ''} open={copiedPrompt} arrow>
                  <Button
                    onClick={(e) => {
                      handleCopyPrompt(e).catch(() => {
                        /* Error already handled */
                      });
                    }}
                    variant="contained"
                    color="primary"
                    startIcon={<Copy size={16} />}
                  >
                    Copy Prompt
                  </Button>
                </Tooltip>
              )}
            </Stack>
          </CardContent>
        </Card>
      )}

      {/* Divider with "or" */}
      {manualSteps && manualSteps.length > 0 && (
        <Divider sx={{my: 2}}>
          <Typography variant="body2" color="text.secondary" sx={{px: 2}}>
            {t('applications:onboarding.summary.guides.divider')}
          </Typography>
        </Divider>
      )}

      {/* Manual Steps Timeline */}
      {manualSteps && manualSteps.length > 0 && (
        <Box>
          {manualSteps.map((step, index) => (
            <Box key={step.step} sx={{display: 'flex', gap: 2, mb: 4}}>
              {/* Timeline dot and connector */}
              <Box sx={{display: 'flex', flexDirection: 'column', alignItems: 'center'}}>
                <Box
                  sx={{
                    width: 32,
                    height: 32,
                    borderRadius: '50%',
                    bgcolor: 'primary.main',
                    color: 'white',
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'center',
                    flexShrink: 0,
                  }}
                >
                  <Typography variant="body2" sx={{fontWeight: 600}}>
                    {step.step}
                  </Typography>
                </Box>
                {index < manualSteps.length - 1 && (
                  <Box
                    sx={{
                      width: 2,
                      flex: 1,
                      bgcolor: 'divider',
                      mt: 1,
                      minHeight: 40,
                    }}
                  />
                )}
              </Box>

              {/* Content */}
              <Box sx={{flex: 1, pb: 2}}>
                <Typography variant="subtitle1" sx={{fontWeight: 600, mb: 1}}>
                  {step.title}
                </Typography>
                <Typography variant="body2" color="text.secondary" sx={{mb: 0.5}}>
                  {step.description}
                </Typography>
                {step.subDescription && (
                  <Typography variant="body2" color="text.secondary" sx={{mb: 0.5}}>
                    {step.subDescription}
                  </Typography>
                )}
                {step.bullets && step.bullets.length > 0 && (
                  <Box component="ul" sx={{mt: 1, pl: 2, mb: 1}}>
                    {step.bullets.map((bullet) => (
                      <Box component="li" key={bullet} sx={{mb: 0.5}}>
                        <Typography variant="body2" color="text.secondary">
                          {bullet}
                        </Typography>
                      </Box>
                    ))}
                  </Box>
                )}
                {renderCodeBlock(step)}
              </Box>
            </Box>
          ))}
        </Box>
      )}
    </Stack>
  );
}
