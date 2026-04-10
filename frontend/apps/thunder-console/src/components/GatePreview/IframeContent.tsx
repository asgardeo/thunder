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

import type {EmbeddedFlowComponent} from '@asgardeo/react';
import createCache from '@emotion/cache';
import {CacheProvider} from '@emotion/react';
import {
  FlowComponentRenderer,
  DesignProvider,
  AuthPageLayout,
  AuthCardLayout,
  GoogleFontLoader,
  type Theme,
  type DesignResolveResponse,
  type Stylesheet,
} from '@thunder/design';
import {useTemplateLiteralResolver} from '@thunder/hooks';
import {TemplateLiteralType} from '@thunder/utils';
import {Box} from '@wso2/oxygen-ui';
import {useEffect, useMemo, type JSX} from 'react';
import PreviewThemeProvider from './PreviewThemeProvider';
import ElementInspector from '../../features/design/components/layouts/ElementInspector';

/**
 * MUI's extendTheme only accepts 'light' or 'dark' for defaultColorScheme.
 * Thunder stores 'system' as a runtime concept — strip it before passing to MUI.
 */
function sanitizeThemeForMui(theme: Theme): Theme {
  const raw = theme as unknown as Record<string, unknown>;
  if (raw.defaultColorScheme === 'system') {
    const copy = {...raw};
    delete copy.defaultColorScheme;
    return copy as unknown as Theme;
  }
  return theme;
}

/** No-op handlers for preview mode — the form is purely visual. */
const noopSubmit = (): void => {
  /* no-op */
};
const noopInputChange = (): void => {
  /* no-op */
};

export interface IframeContentProps {
  iframeDoc: Document;
  colorScheme: 'light' | 'dark';
  theme: Theme | undefined;
  stylesheets: Stylesheet[];
  pageBackground: string | undefined;
  mock: EmbeddedFlowComponent[];
  inspectorEnabled: boolean;
  onSelectSelector?: (selector: string) => void;
}

/**
 * Wraps preview content with an emotion CacheProvider that injects MUI styles
 * into the iframe's <head>, and injects custom stylesheets into the iframe document.
 */
export default function IframeContent({
  iframeDoc,
  colorScheme,
  theme,
  stylesheets,
  pageBackground,
  mock,
  inspectorEnabled,
  onSelectSelector = undefined,
}: IframeContentProps): JSX.Element {
  // Strip {{t(key)}} wrappers, returning the bare i18n key so the shared
  // adapters' own t() call performs the actual translation (avoids double-translation).
  const {resolveAll} = useTemplateLiteralResolver();
  const previewResolve = useMemo(
    () =>
      (template: string | undefined): string | undefined =>
        resolveAll(template, {[TemplateLiteralType.TRANSLATION]: (key: string) => key}),
    [resolveAll],
  );

  // Create an emotion cache that injects styles into the iframe's <head>.
  // An insertion point element is prepended so that emotion styles are always
  // positioned before any custom branding stylesheets, preventing source-order
  // conflicts when MUI regenerates CSS variables on color-scheme toggle.
  const cache = useMemo(() => {
    const insertionPoint = iframeDoc.createElement('meta');
    insertionPoint.setAttribute('name', 'emotion-insertion-point');
    iframeDoc.head.prepend(insertionPoint);
    return createCache({key: 'preview', container: iframeDoc.head, insertionPoint});
  }, [iframeDoc]);

  const themeTypography = theme?.typography as {fontFamily?: string} | undefined;
  const fontFamily = themeTypography?.fontFamily;

  // Inject custom stylesheets into the iframe document (not the parent).
  // The effect also depends on `colorScheme` so that stylesheets are
  // re-appended after MUI/emotion re-injects its Global CSS-variable styles
  // on color-scheme toggle. Emotion's Global component uses useInsertionEffect
  // (which fires before useEffect), so by the time this runs the emotion styles
  // are already in place and our custom sheets land after them in source order.
  const serializedSheets = JSON.stringify(stylesheets);
  useEffect(() => {
    const parsed: Stylesheet[] = JSON.parse(serializedSheets) as Stylesheet[];
    const injectedIds: string[] = [];

    parsed.forEach((sheet) => {
      const elementId = `thunder-preview-${sheet.id}`;
      iframeDoc.getElementById(elementId)?.remove();

      if (sheet.disabled) return;

      if (sheet.type === 'inline') {
        const style = iframeDoc.createElement('style');
        style.id = elementId;
        style.textContent = sheet.content;
        iframeDoc.head.appendChild(style);
        injectedIds.push(elementId);
      } else if (sheet.type === 'url') {
        try {
          const url = new URL(sheet.href);
          if (url.protocol !== 'https:') return;
        } catch {
          return;
        }
        const link = iframeDoc.createElement('link');
        link.id = elementId;
        link.rel = 'stylesheet';
        link.href = sheet.href;
        iframeDoc.head.appendChild(link);
        injectedIds.push(elementId);
      }
    });

    return () => {
      injectedIds.forEach((id) => iframeDoc.getElementById(id)?.remove());
    };
  }, [iframeDoc, serializedSheets, colorScheme]);

  return (
    <CacheProvider value={cache}>
      <GoogleFontLoader fontFamily={fontFamily} targetDocument={iframeDoc} />
      <DesignProvider
        shouldResolveDesignInternally={false}
        design={theme ? ({theme: sanitizeThemeForMui(theme)} as DesignResolveResponse) : undefined}
      >
        <PreviewThemeProvider colorScheme={colorScheme} colorSchemeNode={iframeDoc.documentElement}>
          <ElementInspector enabled={inspectorEnabled} onSelectSelector={onSelectSelector}>
            <AuthPageLayout variant="SignIn" background={pageBackground}>
              <AuthCardLayout variant="SignInBox" showLogo={false}>
                <Box sx={{display: 'flex', flexDirection: 'column', alignItems: 'center', gap: 2}}>
                  {mock.map((component, index) => (
                    <FlowComponentRenderer
                      key={component.id ?? index}
                      component={component}
                      index={index}
                      values={{}}
                      isLoading={false}
                      resolve={previewResolve}
                      onInputChange={noopInputChange}
                      onSubmit={noopSubmit}
                    />
                  ))}
                </Box>
              </AuthCardLayout>
            </AuthPageLayout>
          </ElementInspector>
        </PreviewThemeProvider>
      </DesignProvider>
    </CacheProvider>
  );
}
