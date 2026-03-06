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

import type {JSX, ComponentType} from 'react';
import {useEffect} from 'react';
import {useAsgardeo} from '@asgardeo/react';
import i18next from 'i18next';
import {initReactI18next, useTranslation} from 'react-i18next';
import enUS from '@thunder/i18n/locales/en-US';

interface I18nMeta {
  i18n?: {
    language?: string;
    translations?: Record<string, Record<string, string>>;
  };
}

await i18next.use(initReactI18next).init({
  resources: {
    'en-US': enUS,
  },
  lng: 'en-US',
  fallbackLng: 'en-US',
  defaultNS: 'common',
  interpolation: {
    escapeValue: false,
  },
  debug: import.meta.env.DEV,
});

export default function withI18n<P extends object>(WrappedComponent: ComponentType<P>) {
  return function WithI18n(props: P): JSX.Element {
    const {meta} = useAsgardeo();
    const {i18n} = useTranslation();

    useEffect(() => {
      const i18nMeta = (meta as I18nMeta)?.i18n;
      const translations = i18nMeta?.translations;
      if (!translations) {
        return;
      }

      const language: string = i18nMeta?.language ?? i18n.language;
      let added = false;

      Object.entries(translations).forEach(([namespace, keys]) => {
        if (!keys || Object.keys(keys).length === 0) {
          return;
        }

        const existing = (i18n.getResourceBundle(language, namespace) as Record<string, string> | undefined) ?? {};

        i18n.addResourceBundle(language, namespace, {...existing, ...keys}, true, true);
        added = true;
      });

      if (added) {
        i18n.emit('added', language, Object.keys(translations));
      }

      if (language !== i18n.language) {
        i18n.changeLanguage(language).catch(() => undefined);
      }
    }, [meta, i18n]);

    return <WrappedComponent {...props} />;
  };
}
