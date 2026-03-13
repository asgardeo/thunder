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

import {describe, it, expect} from 'vitest';
import {render, screen} from '@testing-library/react';
import getIntegrationIcon from '../getIntegrationIcon';
import {IdentityProviderTypes} from '../../models/identity-provider';

describe('getIntegrationIcon', () => {
  describe('Supported Provider Types', () => {
    it('should return Google icon for GOOGLE type', () => {
      const icon = getIntegrationIcon(IdentityProviderTypes.GOOGLE);

      expect(icon).not.toBeNull();
      expect(icon?.type).toBeDefined();
    });

    it('should render Google icon correctly', () => {
      const icon = getIntegrationIcon(IdentityProviderTypes.GOOGLE);

      const {container} = render(<div>{icon}</div>);
      const svgElement = container.querySelector('svg');

      expect(svgElement).toBeInTheDocument();
    });

    it('should return GitHub icon for GITHUB type', () => {
      const icon = getIntegrationIcon(IdentityProviderTypes.GITHUB);

      expect(icon).not.toBeNull();
      expect(icon?.type).toBeDefined();
    });

    it('should render GitHub icon correctly', () => {
      const icon = getIntegrationIcon(IdentityProviderTypes.GITHUB);

      const {container} = render(<div>{icon}</div>);
      const svgElement = container.querySelector('svg');

      expect(svgElement).toBeInTheDocument();
    });

    it('should return different icons for different provider types', () => {
      const googleIcon = getIntegrationIcon(IdentityProviderTypes.GOOGLE);
      const githubIcon = getIntegrationIcon(IdentityProviderTypes.GITHUB);

      expect(googleIcon).not.toBeNull();
      expect(githubIcon).not.toBeNull();
      expect(googleIcon?.type).not.toBe(githubIcon?.type);
    });
  });

  describe('Unsupported Provider Types', () => {
    it('should return null for OIDC type', () => {
      const icon = getIntegrationIcon(IdentityProviderTypes.OIDC);

      expect(icon).toBeNull();
    });

    it('should return null for OAUTH type', () => {
      const icon = getIntegrationIcon(IdentityProviderTypes.OAUTH);

      expect(icon).toBeNull();
    });

    it('should return null for unknown provider type', () => {
      const icon = getIntegrationIcon('UNKNOWN_PROVIDER');

      expect(icon).toBeNull();
    });

    it('should return null for empty string', () => {
      const icon = getIntegrationIcon('');

      expect(icon).toBeNull();
    });

    it('should return null for undefined type', () => {
      const icon = getIntegrationIcon(undefined as unknown as string);

      expect(icon).toBeNull();
    });

    it('should return null for null type', () => {
      const icon = getIntegrationIcon(null as unknown as string);

      expect(icon).toBeNull();
    });
  });

  describe('Case Sensitivity', () => {
    it('should be case-sensitive for GOOGLE type', () => {
      const icon = getIntegrationIcon('google');

      expect(icon).toBeNull();
    });

    it('should be case-sensitive for GITHUB type', () => {
      const icon = getIntegrationIcon('github');

      expect(icon).toBeNull();
    });

    it('should match exact case for GOOGLE', () => {
      const upperCaseIcon = getIntegrationIcon('GOOGLE');
      const mixedCaseIcon = getIntegrationIcon('Google');

      expect(upperCaseIcon).not.toBeNull();
      expect(mixedCaseIcon).toBeNull();
    });

    it('should match exact case for GITHUB', () => {
      const upperCaseIcon = getIntegrationIcon('GITHUB');
      const mixedCaseIcon = getIntegrationIcon('GitHub');

      expect(upperCaseIcon).not.toBeNull();
      expect(mixedCaseIcon).toBeNull();
    });
  });

  describe('Icon Rendering', () => {
    it('should render Google icon as a valid React element', () => {
      const icon = getIntegrationIcon(IdentityProviderTypes.GOOGLE);

      expect(icon).toBeTruthy();
      expect(typeof icon).toBe('object');
    });

    it('should render GitHub icon as a valid React element', () => {
      const icon = getIntegrationIcon(IdentityProviderTypes.GITHUB);

      expect(icon).toBeTruthy();
      expect(typeof icon).toBe('object');
    });

    it('should render Google icon in a container without errors', () => {
      const icon = getIntegrationIcon(IdentityProviderTypes.GOOGLE);

      expect(() => render(<div data-testid="icon-container">{icon}</div>)).not.toThrow();

      expect(screen.getByTestId('icon-container')).toBeInTheDocument();
    });

    it('should render GitHub icon in a container without errors', () => {
      const icon = getIntegrationIcon(IdentityProviderTypes.GITHUB);

      expect(() => render(<div data-testid="icon-container">{icon}</div>)).not.toThrow();

      expect(screen.getByTestId('icon-container')).toBeInTheDocument();
    });

    it('should not render anything when icon is null', () => {
      const icon = getIntegrationIcon('UNSUPPORTED');

      render(<div data-testid="icon-container">{icon}</div>);
      const iconContainer = screen.getByTestId('icon-container');

      expect(icon).toBeNull();
      expect(iconContainer).toBeEmptyDOMElement();
    });
  });

  describe('All Provider Types Coverage', () => {
    it('should handle all defined IdentityProviderTypes', () => {
      const results = Object.values(IdentityProviderTypes).map((type) => ({
        type,
        icon: getIntegrationIcon(type),
      }));

      // GOOGLE and GITHUB should have icons
      const googleResult = results.find((r) => r.type === IdentityProviderTypes.GOOGLE);
      const githubResult = results.find((r) => r.type === IdentityProviderTypes.GITHUB);

      expect(googleResult?.icon).not.toBeNull();
      expect(githubResult?.icon).not.toBeNull();

      // OIDC and OAUTH should return null
      const oidcResult = results.find((r) => r.type === IdentityProviderTypes.OIDC);
      const oauthResult = results.find((r) => r.type === IdentityProviderTypes.OAUTH);

      expect(oidcResult?.icon).toBeNull();
      expect(oauthResult?.icon).toBeNull();
    });

    it('should return consistent results for the same provider type', () => {
      const firstCall = getIntegrationIcon(IdentityProviderTypes.GOOGLE);
      const secondCall = getIntegrationIcon(IdentityProviderTypes.GOOGLE);

      // Both should be non-null
      expect(firstCall).not.toBeNull();
      expect(secondCall).not.toBeNull();

      // Should have the same type
      expect(firstCall?.type).toBe(secondCall?.type);
    });
  });

  describe('Edge Cases', () => {
    it('should handle whitespace in provider type', () => {
      const icon = getIntegrationIcon(' GOOGLE ');

      expect(icon).toBeNull();
    });

    it('should handle numeric input', () => {
      const icon = getIntegrationIcon('123' as string);

      expect(icon).toBeNull();
    });

    it('should handle special characters', () => {
      const icon = getIntegrationIcon('GOOGLE@#$%');

      expect(icon).toBeNull();
    });

    it('should handle very long strings', () => {
      const longString = 'A'.repeat(1000);
      const icon = getIntegrationIcon(longString);

      expect(icon).toBeNull();
    });
  });

  describe('Type Safety', () => {
    it('should accept string type parameter', () => {
      const stringType = 'GOOGLE';
      const icon = getIntegrationIcon(stringType);

      expect(icon).not.toBeNull();
    });

    it('should work with IdentityProviderType enum values', () => {
      const enumType = IdentityProviderTypes.GOOGLE;
      const icon = getIntegrationIcon(enumType);

      expect(icon).not.toBeNull();
    });
  });

  describe('Return Value Validation', () => {
    it('should return JSX.Element for supported types', () => {
      const googleIcon = getIntegrationIcon(IdentityProviderTypes.GOOGLE);

      expect(googleIcon).not.toBeNull();
      expect(googleIcon).toHaveProperty('type');
      expect(googleIcon).toHaveProperty('props');
    });

    it('should return null (not undefined) for unsupported types', () => {
      const icon = getIntegrationIcon('UNSUPPORTED');

      expect(icon).toBeNull();
      expect(icon).not.toBeUndefined();
    });

    it('should return exact null value for multiple unsupported types', () => {
      const icon1 = getIntegrationIcon('UNKNOWN1');
      const icon2 = getIntegrationIcon('UNKNOWN2');

      expect(icon1).toBeNull();
      expect(icon2).toBeNull();
      expect(icon1).toBe(icon2); // Both are null
    });
  });

  describe('Integration with React', () => {
    it('should be renderable in a React component tree', () => {
      function GoogleIconComponent() {
        const icon = getIntegrationIcon(IdentityProviderTypes.GOOGLE);
        return <div data-testid="google-wrapper">{icon}</div>;
      }

      render(<GoogleIconComponent />);

      expect(screen.getByTestId('google-wrapper')).toBeInTheDocument();
    });

    it('should be conditionally renderable based on provider type', () => {
      function ConditionalIcon({type}: {type: string}) {
        const icon = getIntegrationIcon(type);
        return <div data-testid="conditional-wrapper">{icon ?? <span>No icon</span>}</div>;
      }

      const {rerender} = render(<ConditionalIcon type={IdentityProviderTypes.GOOGLE} />);
      expect(screen.getByTestId('conditional-wrapper').querySelector('svg')).toBeInTheDocument();

      rerender(<ConditionalIcon type="UNSUPPORTED" />);
      expect(screen.getByText('No icon')).toBeInTheDocument();
    });

    it('should handle multiple icons in the same component', () => {
      function MultiIconComponent() {
        const googleIcon = getIntegrationIcon(IdentityProviderTypes.GOOGLE);
        const githubIcon = getIntegrationIcon(IdentityProviderTypes.GITHUB);

        return (
          <div data-testid="multi-icon-wrapper">
            <div data-testid="google-icon">{googleIcon}</div>
            <div data-testid="github-icon">{githubIcon}</div>
          </div>
        );
      }

      render(<MultiIconComponent />);

      expect(screen.getByTestId('google-icon').querySelector('svg')).toBeInTheDocument();
      expect(screen.getByTestId('github-icon').querySelector('svg')).toBeInTheDocument();
    });
  });
});
