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
import type {IdentityProvider} from '@/features/integrations/models/identity-provider';
import {AuthenticatorTypes} from '@/features/integrations/models/authenticators';
import {IdentityProviderTypes} from '@/features/integrations/models/identity-provider';

const IDP_ICONS: Record<string, string> = {
  GOOGLE: 'assets/images/icons/google.svg',
  GITHUB: 'assets/images/icons/github.svg',
};

/**
 * Builds a preview mock component list based on the enabled integrations and available identity providers.
 *
 * The mock is used to render a realistic sign-in preview via {@link GatePreview} without making real API calls.
 * The generated components reflect which authenticators (basic auth, passkey, social) are currently enabled.
 */
const DEFAULT_INTEGRATIONS: Record<string, boolean> = {
  [AuthenticatorTypes.BASIC_AUTH]: true,
  [AuthenticatorTypes.PASSKEY]: true,
  google: true,
  github: true,
};

const DEFAULT_IDENTITY_PROVIDERS: IdentityProvider[] = [
  {id: 'google', name: 'Google', type: IdentityProviderTypes.GOOGLE},
  {id: 'github', name: 'GitHub', type: IdentityProviderTypes.GITHUB},
];

export default function buildPreviewMock(
  integrations: Record<string, boolean> = DEFAULT_INTEGRATIONS,
  identityProviders: IdentityProvider[] = DEFAULT_IDENTITY_PROVIDERS,
): EmbeddedFlowComponent[] {
  const hasBasicAuth: boolean = integrations[AuthenticatorTypes.BASIC_AUTH] ?? false;
  const hasPasskey: boolean = integrations[AuthenticatorTypes.PASSKEY] ?? false;
  const selectedProviders: IdentityProvider[] = identityProviders.filter(
    (idp: IdentityProvider): boolean => integrations[idp.id] ?? false,
  );
  const hasSocial: boolean = selectedProviders.length > 0;

  const components: Record<string, unknown>[] = [];

  // Heading
  components.push({
    category: 'DISPLAY',
    id: 'text_heading',
    label: '{{t(signin:heading)}}',
    resourceType: 'ELEMENT',
    type: 'TEXT',
    variant: 'HEADING_3',
  });

  // Basic auth block
  if (hasBasicAuth) {
    components.push({
      category: 'BLOCK',
      components: [
        {
          category: 'FIELD',
          hint: '',
          id: 'text_input_username',
          inputType: 'text',
          label: '{{t(elements:fields.username.label)}}',
          placeholder: '{{t(elements:fields.username.placeholder)}}',
          ref: 'username',
          required: true,
          resourceType: 'ELEMENT',
          type: 'TEXT_INPUT',
        },
        {
          category: 'FIELD',
          hint: '',
          id: 'password_input',
          inputType: 'password',
          label: '{{t(elements:fields.password.label)}}',
          placeholder: '{{t(elements:fields.password.placeholder)}}',
          ref: 'password',
          required: true,
          resourceType: 'ELEMENT',
          type: 'PASSWORD_INPUT',
        },
        {
          actionRef: 'ID_basic',
          category: 'ACTION',
          eventType: 'SUBMIT',
          id: 'action_submit',
          label: '{{t(elements:buttons.submit.text)}}',
          resourceType: 'ELEMENT',
          type: 'ACTION',
          variant: 'PRIMARY',
        },
      ],
      id: 'block_basic_auth',
      resourceType: 'ELEMENT',
      type: 'BLOCK',
    });
  }

  // Divider — shown when basic/passkey coexist with social or each other
  const showDivider: boolean = (hasBasicAuth || hasPasskey) && (hasSocial || (hasBasicAuth && hasPasskey));
  if (showDivider) {
    components.push({
      category: 'DISPLAY',
      id: 'divider_or',
      label: '{{t(elements:display.divider.or_separator)}}',
      resourceType: 'ELEMENT',
      type: 'DIVIDER',
      variant: 'HORIZONTAL',
    });
  }

  // Social provider buttons
  selectedProviders.forEach((provider: IdentityProvider) => {
    components.push({
      category: 'ACTION',
      components: [
        {
          actionRef: `ID_${provider.id}`,
          category: 'ACTION',
          eventType: 'TRIGGER',
          id: `action_${provider.id}`,
          image: IDP_ICONS[provider.type] ?? '',
          label: `Continue with ${provider.name}`,
          resourceType: 'ELEMENT',
          type: 'ACTION',
          variant: 'SOCIAL',
        },
      ],
      eventType: 'TRIGGER',
      id: `block_${provider.id}`,
      resourceType: 'ELEMENT',
      type: 'BLOCK',
    });
  });

  // Passkey button
  if (hasPasskey) {
    components.push({
      category: 'ACTION',
      eventType: 'TRIGGER',
      id: 'action_passkey',
      label: 'Use a Passkey',
      resourceType: 'ELEMENT',
      startIcon: `${import.meta.env.BASE_URL}/assets/images/icons/fingerprint.svg`,
      type: 'ACTION',
      variant: 'SOCIAL',
    });
  }

  return components as unknown as EmbeddedFlowComponent[];
}
