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

import {Box, TextField, Stack, Chip, InputAdornment} from '@wso2/oxygen-ui';
import {Mic, Sparkle} from '@wso2/oxygen-ui-icons-react';
import {useState, useEffect, type JSX} from 'react';

const SUGGESTIONS = [
  'Add login to a React app',
  'Secure a MCP Server',
  'Secure a application with role-based access control',
];

const PLACEHOLDERS = [
  'A web application that needs secure user authentication and authorization',
  'A SaaS product requiring single sign-on (SSO) across multiple applications',
  'An application that enforces role-based and attribute-based access control',
  'A multi-tenant platform with organization and tenant-level user management',
  'A consumer-facing app that supports social login and passwordless authentication',
  'An enterprise system integrating with external identity providers via OIDC or SAML',
  'A mobile or SPA application requiring secure token-based authentication',
  'An application that needs multi-factor authentication (MFA) and adaptive login flows',
  'A platform managing user lifecycles, provisioning, and deprovisioning',
  'An API-first product that needs centralized identity and access management',
];

export default function AppGenerationPrompt(): JSX.Element {
  const [prompt, setPrompt] = useState('');
  const [placeholderIndex, setPlaceholderIndex] = useState(0);
  const [isAnimating, setIsAnimating] = useState(false);

  useEffect(() => {
    const interval = setInterval(() => {
      setIsAnimating(true);
      setTimeout(() => {
        setPlaceholderIndex((prev) => (prev + 1) % PLACEHOLDERS.length);
        setIsAnimating(false);
      }, 300);
    }, 4000);

    return () => clearInterval(interval);
  }, []);

  const handleSuggestionClick = (suggestion: string) => {
    setPrompt(suggestion);
  };

  const handleSubmit = () => {
    if (prompt.trim()) {
      // TODO: Implement app generation logic
      // Generate app with prompt: {prompt}
    }
  };

  const handleKeyPress = (event: React.KeyboardEvent) => {
    if (event.key === 'Enter' && !event.shiftKey) {
      event.preventDefault();
      handleSubmit();
    }
  };

  return (
    <Box>
      <Stack spacing={2}>
        <Box
          sx={{
            fontSize: '0.875rem',
            fontWeight: 500,
            color: 'text.secondary',
          }}
        >
          Describe your application ✨
        </Box>

        <TextField
          fullWidth
          multiline
          minRows={3}
          maxRows={6}
          placeholder={PLACEHOLDERS[placeholderIndex]}
          value={prompt}
          onChange={(e) => setPrompt(e.target.value)}
          onKeyPress={handleKeyPress}
          slotProps={{
            input: {
              startAdornment: (
                <InputAdornment position="start">
                  <Box sx={{display: 'flex', color: 'text.secondary', fontSize: '1.25rem'}}>
                    <Sparkle size={20} />
                  </Box>
                </InputAdornment>
              ),
              endAdornment: (
                <InputAdornment position="end">
                  <Box
                    sx={{
                      display: 'flex',
                      gap: 1,
                      alignItems: 'center',
                    }}
                  >
                    <Box
                      sx={{
                        display: 'flex',
                        color: 'text.secondary',
                        fontSize: '1.25rem',
                        cursor: 'pointer',
                        '&:hover': {
                          color: 'text.primary',
                        },
                      }}
                    >
                      <Mic size={20} />
                    </Box>
                    <Chip
                      label="TAB"
                      size="small"
                      sx={{
                        height: '20px',
                        fontSize: '0.75rem',
                        fontWeight: 600,
                        bgcolor: 'action.hover',
                        color: 'text.secondary',
                      }}
                    />
                  </Box>
                </InputAdornment>
              ),
            },
          }}
          sx={{
            '& .MuiInputBase-root': {
              bgcolor: 'background.paper',
              borderRadius: 1,
              fontSize: '1rem',
              alignItems: 'flex-start',
              p: 2,
              '& textarea': {
                pt: 1.5,
              },
            },
            '& .MuiInputBase-input::placeholder': {
              opacity: isAnimating ? 0 : 1,
              transform: isAnimating ? 'translateY(-10px)' : 'translateY(0)',
              transition: 'opacity 0.3s ease-in-out, transform 0.3s ease-in-out',
            },
          }}
        />

        {/* Suggestion Chips */}
        <Stack direction="row" spacing={1} flexWrap="wrap" useFlexGap>
          {SUGGESTIONS.map((suggestion) => (
            <Chip
              key={suggestion}
              label={suggestion}
              onClick={() => handleSuggestionClick(suggestion)}
              sx={{
                bgcolor: 'action.hover',
                color: 'text.primary',
                fontWeight: 400,
                fontSize: '0.875rem',
                cursor: 'pointer',
                '&:hover': {
                  bgcolor: 'action.selected',
                },
                ...(suggestion === "I'm Feeling Lucky" && {
                  bgcolor: 'transparent',
                  border: '1px solid',
                  borderColor: 'divider',
                }),
              }}
            />
          ))}
        </Stack>
      </Stack>
    </Box>
  );
}
