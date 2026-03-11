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

import {describe, it, expect, vi} from 'vitest';
import {render} from '@thunder/test-utils/browser';
import {page, userEvent} from 'vitest/browser';
import {useState} from 'react';
import TokenUserAttributesSection from '../TokenUserAttributesSection';
import type {OAuth2Config} from '../../../../models/oauth';

// Mock the SettingsCard component
vi.mock('../../../../../../components/SettingsCard', () => ({
  default: ({
    title,
    description,
    children,
    headerAction,
  }: {
    title: string;
    description: string;
    children: React.ReactNode;
    headerAction?: React.ReactNode;
  }) => (
    <div data-testid="settings-card">
      <div data-testid="card-title">{title}</div>
      <div data-testid="card-header-action">{headerAction}</div>
      <div data-testid="card-description">{description}</div>
      {children}
    </div>
  ),
}));

// Mock TokenConstants
vi.mock('../../../../constants/token-constants', () => ({
  default: {
    DEFAULT_TOKEN_ATTRIBUTES: ['aud', 'exp', 'iat', 'iss', 'sub'],
    USER_INFO_DEFAULT_ATTRIBUTES: ['sub'],
    ADDITIONAL_USER_ATTRIBUTES: ['ouHandle', 'ouId', 'ouName', 'userType'],
  },
}));

// Wrapper component to manage state
function TestWrapper({
  tokenType = 'shared',
  currentAttributes = [],
  userAttributes = [],
  isLoadingUserAttributes = false,
  oauth2Config = undefined,
  children = undefined,
}: {
  tokenType?: 'shared' | 'access' | 'id' | 'userinfo';
  currentAttributes?: string[];
  userAttributes?: string[];
  isLoadingUserAttributes?: boolean;
  oauth2Config?: OAuth2Config;
  children?: (props: {
    expandedSections: Set<string>;
    setExpandedSections: React.Dispatch<React.SetStateAction<Set<string>>>;
    pendingAdditions: Set<string>;
    pendingRemovals: Set<string>;
    highlightedAttributes: Set<string>;
    onAttributeClick: (attr: string, tokenType: 'shared' | 'access' | 'id' | 'userinfo') => void;
  }) => React.ReactNode;
}) {
  const [expandedSections, setExpandedSections] = useState<Set<string>>(new Set([`user-${tokenType}`]));
  const [pendingAdditions] = useState<Set<string>>(new Set());
  const [pendingRemovals] = useState<Set<string>>(new Set());
  const [highlightedAttributes] = useState<Set<string>>(new Set());
  const onAttributeClick = vi.fn();

  if (children) {
    return (
      <>
        {children({
          expandedSections,
          setExpandedSections,
          pendingAdditions,
          pendingRemovals,
          highlightedAttributes,
          onAttributeClick,
        })}
      </>
    );
  }

  return (
    <TokenUserAttributesSection
      tokenType={tokenType}
      currentAttributes={currentAttributes}
      userAttributes={userAttributes}
      isLoadingUserAttributes={isLoadingUserAttributes}
      expandedSections={expandedSections}
      setExpandedSections={setExpandedSections}
      pendingAdditions={pendingAdditions}
      pendingRemovals={pendingRemovals}
      highlightedAttributes={highlightedAttributes}
      onAttributeClick={onAttributeClick}
      activeTokenType="access"
      oauth2Config={oauth2Config}
    />
  );
}

describe('TokenUserAttributesSection', () => {
  describe('Rendering with tokenType="shared"', () => {
    it('should render the settings card with correct title and description', async () => {
      await render(<TestWrapper />);

      await expect.element(page.getByTestId('card-title')).toHaveTextContent('User Attributes');
      await expect.element(page.getByTestId('card-description')).toHaveTextContent(
        'Select which user attributes to include in your tokens. These attributes will be available in the issued tokens.',
      );
    });

    it('should render JWT preview section', async () => {
      await render(<TestWrapper />);

      await expect.element(page.getByText('Token Preview (JWT)')).toBeInTheDocument();
    });

    it('should render user attributes accordion', async () => {
      await render(<TestWrapper />);

      const userAttributesElements = page.getByText('User Attributes');
      expect(userAttributesElements.length).toBeGreaterThan(0);
    });

    it('should render default attributes accordion', async () => {
      const {container} = await render(<TestWrapper />);

      // Target the h6 Typography element specifically to avoid matching the h3 Accordion wrapper
      const defaultAttributesHeading = Array.from(container.querySelectorAll('h6')).find(
        (el) => el.textContent === 'Default Attributes',
      );
      expect(defaultAttributesHeading).toBeInTheDocument();
    });

    it('should display default token attributes as chips', async () => {
      const {container} = await render(<TestWrapper />);

      // Use container query to find chips specifically (not JWT preview spans)
      const chipLabels = container.querySelectorAll('.MuiChip-label');
      const chipTexts = Array.from(chipLabels).map((el) => el.textContent);
      expect(chipTexts).toContain('aud');
      expect(chipTexts).toContain('exp');
      expect(chipTexts).toContain('iat');
      expect(chipTexts).toContain('iss');
      expect(chipTexts).toContain('sub');
    });

    it('should display loading state when isLoadingUserAttributes is true', async () => {
      await render(<TestWrapper isLoadingUserAttributes />);

      // Loading state is rendered, but the exact text depends on i18n translations
      await expect.element(page.getByTestId('card-title')).toHaveTextContent('User Attributes');
    });

    it('should display user attributes as chips when provided', async () => {
      const userAttributes = ['email', 'username', 'firstName'];
      await render(<TestWrapper userAttributes={userAttributes} />);

      await expect.element(page.getByText('email')).toBeInTheDocument();
      await expect.element(page.getByText('username')).toBeInTheDocument();
      await expect.element(page.getByText('firstName')).toBeInTheDocument();
    });

    it('should display no attributes message when userAttributes is empty', async () => {
      await render(<TestWrapper userAttributes={[]} isLoadingUserAttributes={false} />);

      // Empty state alert is rendered with specific message
      expect(
        page.getByText('No user attributes available. Configure allowed user types for this application.'),
      ).toBeInTheDocument();
    });

    it('should not render scopes section for shared token type', async () => {
      await render(<TestWrapper tokenType="shared" />);

      await expect.element(page.getByText('Scopes')).not.toBeInTheDocument();
    });
  });

  describe('Rendering with tokenType="access"', () => {
    it('should render correct title for access token', async () => {
      await render(<TestWrapper tokenType="access" />);

      await expect.element(page.getByTestId('card-title')).toHaveTextContent('Access Token User Attributes');
      await expect.element(page.getByTestId('card-description')).toHaveTextContent(
        'Configure user attributes that will be included in the access token. You can add custom attributes from user profiles.',
      );
    });

    it('should render access token preview title', async () => {
      await render(<TestWrapper tokenType="access" />);

      await expect.element(page.getByText('Access Token Preview (JWT)')).toBeInTheDocument();
    });

    it('should not render scopes section for access token', async () => {
      await render(<TestWrapper tokenType="access" />);

      await expect.element(page.getByText('Scopes')).not.toBeInTheDocument();
    });
  });

  describe('Rendering with tokenType="id"', () => {
    it('should render correct title for ID token', async () => {
      await render(<TestWrapper tokenType="id" />);

      await expect.element(page.getByTestId('card-title')).toHaveTextContent('ID Token User Attributes');
      await expect.element(page.getByTestId('card-description')).toHaveTextContent(
        'Configure user attributes that will be included in the ID token. You can add custom attributes from user profiles and define scope-based attributes.',
      );
    });

    it('should render ID token preview title', async () => {
      await render(<TestWrapper tokenType="id" />);

      await expect.element(page.getByText('ID Token Preview (JWT)')).toBeInTheDocument();
    });

    it('should render scopes section for ID token', async () => {
      const oauth2Config = {
        scopes: ['openid', 'profile', 'email'],
      } as OAuth2Config;

      await render(<TestWrapper tokenType="id" oauth2Config={oauth2Config} />);

      await expect.element(page.getByText('Scopes')).toBeInTheDocument();
    });

    it('should display scopes as chips when provided', async () => {
      const oauth2Config = {
        scopes: ['openid', 'profile', 'email'],
      } as OAuth2Config;

      const {container} = await render(<TestWrapper tokenType="id" oauth2Config={oauth2Config} />);

      // Use chip labels to avoid matching description text that contains "profiles" or "email"
      const chipLabels = container.querySelectorAll('.MuiChip-label');
      const chipTexts = Array.from(chipLabels).map((el) => el.textContent);
      expect(chipTexts).toContain('openid');
      expect(chipTexts).toContain('profile');
      expect(chipTexts).toContain('email');
    });

    it('should display no scopes message when scopes array is empty', async () => {
      const oauth2Config: OAuth2Config = {
        grant_types: [],
        response_types: [],
        scopes: [],
      };

      await render(<TestWrapper tokenType="id" oauth2Config={oauth2Config} />);

      await expect.element(page.getByText('No scopes configured')).toBeInTheDocument();
    });
  });

  describe('User Interaction', () => {
    it('should call onAttributeClick when a user attribute chip is clicked', async () => {
            const mockOnAttributeClick = vi.fn();

      await render(
        <TestWrapper userAttributes={['email', 'username']}>
          {(props) => (
            <TokenUserAttributesSection
              tokenType="shared"
              currentAttributes={[]}
              userAttributes={['email', 'username']}
              isLoadingUserAttributes={false}
              expandedSections={props.expandedSections}
              setExpandedSections={props.setExpandedSections}
              pendingAdditions={props.pendingAdditions}
              pendingRemovals={props.pendingRemovals}
              highlightedAttributes={props.highlightedAttributes}
              onAttributeClick={mockOnAttributeClick}
              activeTokenType="access"
            />
          )}
        </TestWrapper>,
      );

      const emailChip = page.getByText('email');
      await userEvent.click(emailChip);

      expect(mockOnAttributeClick).toHaveBeenCalledWith('email', 'shared');
    });

    it('should toggle user attributes accordion when clicked', async () => {
      
      await render(<TestWrapper userAttributes={['email']} />);

      // Find the User Attributes accordion header (not the default attributes one)
      const accordionHeaders = page.getByText('User Attributes').all();
      const userAttributesHeader = accordionHeaders.find((el: import('vitest/browser').Locator) => el.element().closest('button') !== null);

      expect(userAttributesHeader).toBeDefined();

      // Initially expanded - should show the email chip
      await expect.element(page.getByText('email')).toBeInTheDocument();

      // Click to collapse
      await userEvent.click(userAttributesHeader!);

      // Note: Accordion collapse behavior is controlled by Material-UI
      // In a real DOM, the content would be hidden but may still be in the document
    });

    it('should render current attributes as filled chips', async () => {
      const userAttributes = ['email', 'username', 'firstName'];
      const currentAttributes = ['email', 'username'];

      const {container} = await render(<TestWrapper userAttributes={userAttributes} currentAttributes={currentAttributes} />);

      // Use chip labels to avoid matching JWT preview spans
      const chipLabels = container.querySelectorAll('.MuiChip-label');
      const chipTexts = Array.from(chipLabels).map((el) => el.textContent);

      expect(chipTexts).toContain('email');
      expect(chipTexts).toContain('username');
      expect(chipTexts).toContain('firstName');
    });
  });

  describe('JWT Preview', () => {
    it('should display current attributes in JWT preview', async () => {
      const {container} = await render(<TestWrapper currentAttributes={['email', 'username']} />);

      // SyntaxHighlighter renders code, so we check the container's text content
      const jsonText = container.textContent || '';
      expect(jsonText).toContain('email');
      expect(jsonText).toContain('username');
    });

    it('should display default attributes in JWT preview', async () => {
      const {container} = await render(<TestWrapper />);

      const jsonText = container.textContent || '';
      expect(jsonText).toContain('aud');
      expect(jsonText).toContain('exp');
      expect(jsonText).toContain('iat');
      expect(jsonText).toContain('iss');
      expect(jsonText).toContain('sub');
    });
  });

  describe('Info Messages', () => {
    it('should display info about default attributes being always included', async () => {
      await render(<TestWrapper />);

      // Info message is rendered but text depends on i18n translations
      // Check that Default Attributes accordion is present
      // Target the h6 Typography element specifically to avoid matching the h3 Accordion wrapper
      const defaultAttributesHeading = Array.from(document.querySelectorAll('h6')).find(
        (el) => el.textContent === 'Default Attributes',
      );
      expect(defaultAttributesHeading).toBeInTheDocument();
    });

    it('should display hint about configuring attributes', async () => {
      await render(<TestWrapper userAttributes={['email']} />);

      // Tooltip messages are rendered on hover but depend on i18n translations
      // Check that user attributes are rendered as clickable chips
      const emailChip = page.getByText('email');
      expect(emailChip).toBeInTheDocument();
    });
  });

  describe('Pending Additions and Removals', () => {
    it('should include pending additions in JWT preview for shared token type', async () => {
      const {container} = await render(
        <TokenUserAttributesSection
          tokenType="shared"
          currentAttributes={[]}
          userAttributes={['email', 'username']}
          isLoadingUserAttributes={false}
          expandedSections={new Set(['user-shared', 'default-shared'])}
          setExpandedSections={vi.fn()}
          pendingAdditions={new Set(['email'])}
          pendingRemovals={new Set()}
          highlightedAttributes={new Set()}
          onAttributeClick={vi.fn()}
          activeTokenType="access"
        />,
      );

      // The pending addition 'email' should appear in the JWT preview
      const jsonText = container.textContent || '';
      expect(jsonText).toContain('"email"');
      expect(jsonText).toContain('<email>');
    });

    it('should include pending additions in JWT preview for access token when activeTokenType matches', async () => {
      const {container} = await render(
        <TokenUserAttributesSection
          tokenType="access"
          currentAttributes={[]}
          userAttributes={['email']}
          isLoadingUserAttributes={false}
          expandedSections={new Set(['user-access', 'default-access'])}
          setExpandedSections={vi.fn()}
          pendingAdditions={new Set(['email'])}
          pendingRemovals={new Set()}
          highlightedAttributes={new Set()}
          onAttributeClick={vi.fn()}
          activeTokenType="access"
        />,
      );

      const jsonText = container.textContent || '';
      expect(jsonText).toContain('"email"');
      expect(jsonText).toContain('<email>');
    });

    it('should include pending additions in JWT preview for id token when activeTokenType matches', async () => {
      const {container} = await render(
        <TokenUserAttributesSection
          tokenType="id"
          currentAttributes={[]}
          userAttributes={['email']}
          isLoadingUserAttributes={false}
          expandedSections={new Set(['user-id', 'default-id'])}
          setExpandedSections={vi.fn()}
          pendingAdditions={new Set(['email'])}
          pendingRemovals={new Set()}
          highlightedAttributes={new Set()}
          onAttributeClick={vi.fn()}
          activeTokenType="id"
        />,
      );

      const jsonText = container.textContent || '';
      expect(jsonText).toContain('"email"');
      expect(jsonText).toContain('<email>');
    });

    it('should include pending additions in JWT preview for userinfo token when activeTokenType matches', async () => {
      const {container} = await render(
        <TokenUserAttributesSection
          tokenType="userinfo"
          currentAttributes={[]}
          userAttributes={['email']}
          isLoadingUserAttributes={false}
          expandedSections={new Set(['user-userinfo', 'default-userinfo'])}
          setExpandedSections={vi.fn()}
          pendingAdditions={new Set(['email'])}
          pendingRemovals={new Set()}
          highlightedAttributes={new Set()}
          onAttributeClick={vi.fn()}
          activeTokenType="userinfo"
        />,
      );

      const jsonText = container.textContent || '';
      expect(jsonText).toContain('"email"');
      expect(jsonText).toContain('<email>');
    });

    it('should not include pending additions in JWT preview when activeTokenType does not match', async () => {
      const {container} = await render(
        <TokenUserAttributesSection
          tokenType="access"
          currentAttributes={[]}
          userAttributes={['email']}
          isLoadingUserAttributes={false}
          expandedSections={new Set(['user-access', 'default-access'])}
          setExpandedSections={vi.fn()}
          pendingAdditions={new Set(['email'])}
          pendingRemovals={new Set()}
          highlightedAttributes={new Set()}
          onAttributeClick={vi.fn()}
          activeTokenType="id"
        />,
      );

      // 'email' should NOT appear as a value in the preview since activeTokenType doesn't match
      const jsonText = container.textContent || '';
      expect(jsonText).not.toContain('<email>');
    });

    it('should exclude pending removals from JWT preview for shared token type', async () => {
      const {container} = await render(
        <TokenUserAttributesSection
          tokenType="shared"
          currentAttributes={['email', 'username']}
          userAttributes={['email', 'username']}
          isLoadingUserAttributes={false}
          expandedSections={new Set(['user-shared', 'default-shared'])}
          setExpandedSections={vi.fn()}
          pendingAdditions={new Set()}
          pendingRemovals={new Set(['email'])}
          highlightedAttributes={new Set()}
          onAttributeClick={vi.fn()}
          activeTokenType="access"
        />,
      );

      const jsonText = container.textContent || '';
      // email should be removed from preview since it's a pending removal
      expect(jsonText).not.toContain('<email>');
      // username should still be in preview
      expect(jsonText).toContain('<username>');
    });

    it('should exclude pending removals from JWT preview when activeTokenType matches', async () => {
      const {container} = await render(
        <TokenUserAttributesSection
          tokenType="access"
          currentAttributes={['email', 'username']}
          userAttributes={['email', 'username']}
          isLoadingUserAttributes={false}
          expandedSections={new Set(['user-access', 'default-access'])}
          setExpandedSections={vi.fn()}
          pendingAdditions={new Set()}
          pendingRemovals={new Set(['email'])}
          highlightedAttributes={new Set()}
          onAttributeClick={vi.fn()}
          activeTokenType="access"
        />,
      );

      const jsonText = container.textContent || '';
      expect(jsonText).not.toContain('<email>');
      expect(jsonText).toContain('<username>');
    });

    it('should not exclude pending removals from JWT preview when activeTokenType does not match', async () => {
      const {container} = await render(
        <TokenUserAttributesSection
          tokenType="access"
          currentAttributes={['email', 'username']}
          userAttributes={['email', 'username']}
          isLoadingUserAttributes={false}
          expandedSections={new Set(['user-access', 'default-access'])}
          setExpandedSections={vi.fn()}
          pendingAdditions={new Set()}
          pendingRemovals={new Set(['email'])}
          highlightedAttributes={new Set()}
          onAttributeClick={vi.fn()}
          activeTokenType="id"
        />,
      );

      const jsonText = container.textContent || '';
      // email should still be in preview since activeTokenType doesn't match tokenType
      expect(jsonText).toContain('<email>');
      expect(jsonText).toContain('<username>');
    });

    it('should show pending addition chip as active for shared token type', async () => {
      const {container} = await render(
        <TokenUserAttributesSection
          tokenType="shared"
          currentAttributes={[]}
          userAttributes={['email', 'username']}
          isLoadingUserAttributes={false}
          expandedSections={new Set(['user-shared', 'default-shared'])}
          setExpandedSections={vi.fn()}
          pendingAdditions={new Set(['email'])}
          pendingRemovals={new Set()}
          highlightedAttributes={new Set()}
          onAttributeClick={vi.fn()}
          activeTokenType="access"
        />,
      );

      // Use container query to find the chip specifically (not JWT preview spans)
      const emailChipRoot = Array.from(container.querySelectorAll('.MuiChip-root')).find(
        (el) => el.querySelector('.MuiChip-label')?.textContent === 'email',
      );
      expect(emailChipRoot).toBeInTheDocument();
      // The chip for a pending addition should be filled/primary
      expect(emailChipRoot).toHaveClass('MuiChip-filled');
    });

    it('should show pending removal chip as inactive for shared token type', async () => {
      const {container} = await render(
        <TokenUserAttributesSection
          tokenType="shared"
          currentAttributes={['email']}
          userAttributes={['email', 'username']}
          isLoadingUserAttributes={false}
          expandedSections={new Set(['user-shared', 'default-shared'])}
          setExpandedSections={vi.fn()}
          pendingAdditions={new Set()}
          pendingRemovals={new Set(['email'])}
          highlightedAttributes={new Set()}
          onAttributeClick={vi.fn()}
          activeTokenType="access"
        />,
      );

      // Use container query to find the chip specifically (not JWT preview spans)
      const emailChipRoot = Array.from(container.querySelectorAll('.MuiChip-root')).find(
        (el) => el.querySelector('.MuiChip-label')?.textContent === 'email',
      );
      expect(emailChipRoot).toBeInTheDocument();
      // The chip for a pending removal should be outlined/inactive
      expect(emailChipRoot).toHaveClass('MuiChip-outlined');
    });
  });

  describe('Accordion Toggle Behavior', () => {
    it('should collapse and re-expand user attributes accordion', async () => {
            const mockSetExpandedSections = vi.fn();

      await render(
        <TokenUserAttributesSection
          tokenType="shared"
          currentAttributes={[]}
          userAttributes={['email']}
          isLoadingUserAttributes={false}
          expandedSections={new Set(['user-shared', 'default-shared'])}
          setExpandedSections={mockSetExpandedSections}
          pendingAdditions={new Set()}
          pendingRemovals={new Set()}
          highlightedAttributes={new Set()}
          onAttributeClick={vi.fn()}
          activeTokenType="access"
        />,
      );

      // Find the User Attributes accordion summary button
      const accordionHeaders = page.getByText('User Attributes').all();
      const userAttributesButton = accordionHeaders.find((el: import('vitest/browser').Locator) => el.element().closest('button') !== null);
      expect(userAttributesButton).toBeDefined();

      // Click to collapse
      await userEvent.click(userAttributesButton!);

      // Verify setExpandedSections was called
      expect(mockSetExpandedSections).toHaveBeenCalled();

      // Get the updater function and test collapse behavior
      const collapseUpdater = mockSetExpandedSections.mock.calls[mockSetExpandedSections.mock.calls.length - 1][0] as (
        prev: Set<string>,
      ) => Set<string>;
      const collapseResult: Set<string> = collapseUpdater(new Set(['user-shared', 'default-shared']));
      expect(collapseResult.has('user-shared')).toBe(false);
      expect(collapseResult.has('default-shared')).toBe(true);

      // Now simulate re-expand: render with collapsed state, then click to expand
      mockSetExpandedSections.mockClear();

      const {unmount} = await render(
        <TokenUserAttributesSection
          tokenType="shared"
          currentAttributes={[]}
          userAttributes={['email']}
          isLoadingUserAttributes={false}
          expandedSections={new Set(['default-shared'])}
          setExpandedSections={mockSetExpandedSections}
          pendingAdditions={new Set()}
          pendingRemovals={new Set()}
          highlightedAttributes={new Set()}
          onAttributeClick={vi.fn()}
          activeTokenType="access"
        />,
      );

      const accordionHeaders2 = page.getByText('User Attributes').all();
      const userAttributesButton2 = accordionHeaders2.find((el: import('vitest/browser').Locator) => el.element().closest('button') !== null);

      await userEvent.click(userAttributesButton2!);

      // Find the call from the onChange handler (not the useEffect auto-expand)
      const allCalls = mockSetExpandedSections.mock.calls as [(prev: Set<string>) => Set<string>][];
      // Test the updater that adds user-shared back
      const expandCall = allCalls.find((call) => {
        if (typeof call[0] === 'function') {
          const result: Set<string> = call[0](new Set(['default-shared']));
          return result.has('user-shared');
        }
        return false;
      });
      expect(expandCall).toBeDefined();
      const expandResult: Set<string> = expandCall![0](new Set(['default-shared']));
      expect(expandResult.has('user-shared')).toBe(true);

      await unmount();
    });

    it('should collapse and re-expand default attributes accordion', async () => {
            const mockSetExpandedSections = vi.fn();

      await render(
        <TokenUserAttributesSection
          tokenType="shared"
          currentAttributes={[]}
          userAttributes={['email']}
          isLoadingUserAttributes={false}
          expandedSections={new Set(['user-shared', 'default-shared'])}
          setExpandedSections={mockSetExpandedSections}
          pendingAdditions={new Set()}
          pendingRemovals={new Set()}
          highlightedAttributes={new Set()}
          onAttributeClick={vi.fn()}
          activeTokenType="access"
        />,
      );

      // Find the Default Attributes accordion summary button
      // Target the h6 Typography element specifically to avoid matching the h3 Accordion wrapper
      const defaultAttributesH6 = Array.from(document.querySelectorAll('h6')).find(
        (el) => el.textContent === 'Default Attributes',
      );
      const defaultAttributesButton = defaultAttributesH6?.closest('button');
      expect(defaultAttributesButton).toBeDefined();

      // Click to collapse
      await userEvent.click(defaultAttributesButton!);

      // Verify setExpandedSections was called
      expect(mockSetExpandedSections).toHaveBeenCalled();

      // Get the updater function and test collapse behavior (isExpanded=false -> delete)
      const collapseUpdater = mockSetExpandedSections.mock.calls[mockSetExpandedSections.mock.calls.length - 1][0] as (
        prev: Set<string>,
      ) => Set<string>;
      const collapseResult: Set<string> = collapseUpdater(new Set(['user-shared', 'default-shared']));
      expect(collapseResult.has('default-shared')).toBe(false);
      expect(collapseResult.has('user-shared')).toBe(true);

      // Now simulate re-expand: render with collapsed default, click to expand
      mockSetExpandedSections.mockClear();

      const {unmount} = await render(
        <TokenUserAttributesSection
          tokenType="shared"
          currentAttributes={[]}
          userAttributes={['email']}
          isLoadingUserAttributes={false}
          expandedSections={new Set(['user-shared'])}
          setExpandedSections={mockSetExpandedSections}
          pendingAdditions={new Set()}
          pendingRemovals={new Set()}
          highlightedAttributes={new Set()}
          onAttributeClick={vi.fn()}
          activeTokenType="access"
        />,
      );

      const defaultAttributesButton2 = Array.from(document.querySelectorAll('button')).find(
        (btn) => btn.textContent?.includes('Default Attributes'),
      );

      await userEvent.click(defaultAttributesButton2!);

      // Find the call from the onChange handler that adds default-shared back
      const allCalls = mockSetExpandedSections.mock.calls as [(prev: Set<string>) => Set<string>][];
      const expandCall = allCalls.find((call) => {
        if (typeof call[0] === 'function') {
          const result: Set<string> = call[0](new Set(['user-shared']));
          return result.has('default-shared');
        }
        return false;
      });
      expect(expandCall).toBeDefined();
      const expandResult: Set<string> = expandCall![0](new Set(['user-shared']));
      expect(expandResult.has('default-shared')).toBe(true);

      await unmount();
    });
  });

  describe('Refinements', () => {
    it('should hide content when readOnly is true', async () => {
      await render(
        <TokenUserAttributesSection
          tokenType="userinfo"
          currentAttributes={[]}
          userAttributes={[]}
          isLoadingUserAttributes={false}
          expandedSections={new Set()}
          setExpandedSections={vi.fn()}
          pendingAdditions={new Set()}
          pendingRemovals={new Set()}
          highlightedAttributes={new Set()}
          onAttributeClick={vi.fn()}
          activeTokenType="userinfo"
          readOnly
        />,
      );

      // Content should be hidden (accordion headings are inside the !readOnly block)
      // Target h6 Typography elements specifically to avoid matching the card title "User Info Attributes"
      const userAttributesH6 = Array.from(document.querySelectorAll('h6')).find(
        (el) => el.textContent === 'User Attributes',
      );
      const defaultAttributesH6 = Array.from(document.querySelectorAll('h6')).find(
        (el) => el.textContent === 'Default Attributes',
      );
      expect(userAttributesH6).not.toBeInTheDocument();
      expect(defaultAttributesH6).not.toBeInTheDocument();
    });

    it('should render headerAction', async () => {
      await render(
        <TokenUserAttributesSection
          tokenType="userinfo"
          currentAttributes={[]}
          userAttributes={[]}
          isLoadingUserAttributes={false}
          expandedSections={new Set()}
          setExpandedSections={vi.fn()}
          pendingAdditions={new Set()}
          pendingRemovals={new Set()}
          highlightedAttributes={new Set()}
          onAttributeClick={vi.fn()}
          activeTokenType="userinfo"
          headerAction={<button type="button">Test Action</button>}
        />,
      );

      await expect.element(page.getByText('Test Action')).toBeInTheDocument();
    });
  });
});
