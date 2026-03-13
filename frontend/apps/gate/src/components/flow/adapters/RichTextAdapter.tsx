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

import type {JSX} from 'react';
import {Box} from '@wso2/oxygen-ui';
import {useDesign} from '@thunder/shared-design';
import {containsMetaTemplate, replaceMetaTemplate} from '@thunder/utils';
import DOMPurify from 'dompurify';
import type {FlowComponent} from '../../../models/flow';

/** The meta key used by the server to embed the application's sign-up URL. */
const SIGN_UP_URL_META_KEY = 'application.signUpUrl';

/** The meta key that controls whether self registration is enabled. */
const REGISTRATION_ENABLED_META_KEY = 'is_registration_flow_enabled';

interface RichTextAdapterProps {
  component: FlowComponent;
  resolve: (template: string | undefined) => string | undefined;
  /**
   * Fallback sign-up URL used when the flow meta does not supply
   * `application.signUpUrl` but self registration is enabled.
   */
  signUpFallbackUrl?: string;
}

export default function RichTextAdapter({
  component,
  resolve,
  signUpFallbackUrl = undefined,
}: RichTextAdapterProps): JSX.Element | null {
  const {isDesignEnabled} = useDesign();
  const rawLabel = typeof component.label === 'string' ? component.label : undefined;

  // When any component label embeds a sign-up URL meta template we treat the
  // whole element as the "sign up link" block.  Show it only when self
  // registration is enabled; hide it entirely otherwise.
  if (rawLabel && containsMetaTemplate(rawLabel, SIGN_UP_URL_META_KEY)) {
    const isRegistrationEnabled = resolve(`{{meta(${REGISTRATION_ENABLED_META_KEY})}}`) === 'true';

    if (!isRegistrationEnabled) {
      return null;
    }

    // Resolve the label so all other meta/i18n tokens are processed first.
    let resolvedLabel = resolve(rawLabel) ?? rawLabel;

    // If the sign-up URL token is still present after resolution (i.e. the
    // server did not provide application.signUpUrl in meta), substitute the
    // fallback URL so the link still works.
    if (containsMetaTemplate(resolvedLabel, SIGN_UP_URL_META_KEY) && signUpFallbackUrl) {
      resolvedLabel = replaceMetaTemplate(resolvedLabel, SIGN_UP_URL_META_KEY, signUpFallbackUrl);
    }

    return (
      <Box
        sx={{mb: 1, textAlign: isDesignEnabled ? 'center' : 'left'}}
        dangerouslySetInnerHTML={{__html: DOMPurify.sanitize(resolvedLabel)}}
      />
    );
  }

  const resolvedLabel = resolve(rawLabel);

  return (
    <Box
      sx={{mb: 1, textAlign: isDesignEnabled ? 'center' : 'left'}}
      dangerouslySetInnerHTML={{
        __html: DOMPurify.sanitize(resolvedLabel ?? rawLabel ?? ''),
      }}
    />
  );
}
