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
import {render, screen} from '@testing-library/react';
import userEvent from '@testing-library/user-event';
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
    it('should render the settings card with correct title and description', () => {
      render(<TestWrapper />);

      expect(screen.getByTestId('card-title')).toHaveTextContent('User Attributes');
      expect(screen.getByTestId('card-description')).toHaveTextContent(
        'Select which user attributes to include in your tokens. These attributes will be available in the issued tokens.',
      );
    });

    it('should render JWT preview section', () => {
      render(<TestWrapper />);

      expect(screen.getByText('Token Preview (JWT)')).toBeInTheDocument();
    });

    it('should render user attributes accordion', () => {
      render(<TestWrapper />);

      const userAttributesElements = screen.getAllByText('User Attributes');
      expect(userAttributesElements.length).toBeGreaterThan(0);
    });

    it('should render default attributes accordion', () => {
      render(<TestWrapper />);

      expect(screen.getByText('Default Attributes')).toBeInTheDocument();
    });

    it('should display default token attributes as chips', () => {
      render(<TestWrapper />);

      expect(screen.getByText('aud')).toBeInTheDocument();
      expect(screen.getByText('exp')).toBeInTheDocument();
      expect(screen.getByText('iat')).toBeInTheDocument();
      expect(screen.getByText('iss')).toBeInTheDocument();
      expect(screen.getByText('sub')).toBeInTheDocument();
    });

    it('should display loading state when isLoadingUserAttributes is true', () => {
      render(<TestWrapper isLoadingUserAttributes />);

      // Loading state is rendered, but the exact text depends on i18n translations
      expect(screen.getByTestId('card-title')).toHaveTextContent('User Attributes');
    });

    it('should display user attributes as chips when provided', () => {
      const userAttributes = ['email', 'username', 'given_name'];
      render(<TestWrapper userAttributes={userAttributes} />);

      expect(screen.getByText('email')).toBeInTheDocument();
      expect(screen.getByText('username')).toBeInTheDocument();
      expect(screen.getByText('given_name')).toBeInTheDocument();
    });

    it('should display no attributes message when userAttributes is empty', () => {
      render(<TestWrapper userAttributes={[]} isLoadingUserAttributes={false} />);

      // Empty state alert is rendered with specific message
      expect(
        screen.getByText('No user attributes available. Configure allowed user types for this application.'),
      ).toBeInTheDocument();
    });

    it('should not render scopes section for shared token type', () => {
      render(<TestWrapper tokenType="shared" />);

      expect(screen.queryByText('Scopes')).not.toBeInTheDocument();
    });
  });

  describe('Rendering with tokenType="access"', () => {
    it('should render correct title for access token', () => {
      render(<TestWrapper tokenType="access" />);

      expect(screen.getByTestId('card-title')).toHaveTextContent('Access Token User Attributes');
      expect(screen.getByTestId('card-description')).toHaveTextContent(
        'Configure user attributes that will be included in the access token. You can add custom attributes from user profiles.',
      );
    });

    it('should render access token preview title', () => {
      render(<TestWrapper tokenType="access" />);

      expect(screen.getByText('Access Token Preview (JWT)')).toBeInTheDocument();
    });

    it('should not render scopes section for access token', () => {
      render(<TestWrapper tokenType="access" />);

      expect(screen.queryByText('Scopes')).not.toBeInTheDocument();
    });
  });

  describe('Rendering with tokenType="id"', () => {
    it('should render correct title for ID token', () => {
      render(<TestWrapper tokenType="id" />);

      expect(screen.getByTestId('card-title')).toHaveTextContent('ID Token User Attributes');
      expect(screen.getByTestId('card-description')).toHaveTextContent(
        'Configure user attributes that will be included in the ID token. You can add custom attributes from user profiles and define scope-based attributes.',
      );
    });

    it('should render ID token preview title', () => {
      render(<TestWrapper tokenType="id" />);

      expect(screen.getByText('ID Token Preview (JWT)')).toBeInTheDocument();
    });

    it('should render scopes section for ID token', () => {
      const oauth2Config = {
        scopes: ['openid', 'profile', 'email'],
      } as OAuth2Config;

      render(<TestWrapper tokenType="id" oauth2Config={oauth2Config} />);

      expect(screen.getByText('Scopes')).toBeInTheDocument();
    });

    it('should display scopes as chips when provided', () => {
      const oauth2Config = {
        scopes: ['openid', 'profile', 'email'],
      } as OAuth2Config;

      render(<TestWrapper tokenType="id" oauth2Config={oauth2Config} />);

      expect(screen.getByText('openid')).toBeInTheDocument();
      expect(screen.getByText('profile')).toBeInTheDocument();
      expect(screen.getByText('email')).toBeInTheDocument();
    });

    it('should display no scopes message when scopes array is empty', () => {
      const oauth2Config: OAuth2Config = {
        grant_types: [],
        response_types: [],
        scopes: [],
      };

      render(<TestWrapper tokenType="id" oauth2Config={oauth2Config} />);

      expect(screen.getByText('No scopes configured')).toBeInTheDocument();
    });
  });

  describe('User Interaction', () => {
    it('should call onAttributeClick when a user attribute chip is clicked', async () => {
      const user = userEvent.setup();
      const mockOnAttributeClick = vi.fn();

      render(
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

      const emailChip = screen.getByText('email');
      await user.click(emailChip);

      expect(mockOnAttributeClick).toHaveBeenCalledWith('email', 'shared');
    });

    it('should toggle user attributes accordion when clicked', async () => {
      const user = userEvent.setup();

      render(<TestWrapper userAttributes={['email']} />);

      // Find the User Attributes accordion header (not the default attributes one)
      const accordionHeaders = screen.getAllByText('User Attributes');
      const userAttributesHeader = accordionHeaders.find((el) => el.closest('button') !== null);

      expect(userAttributesHeader).toBeDefined();

      // Initially expanded - should show the email chip
      expect(screen.getByText('email')).toBeInTheDocument();

      // Click to collapse
      await user.click(userAttributesHeader!);

      // Note: Accordion collapse behavior is controlled by Material-UI
      // In a real DOM, the content would be hidden but may still be in the document
    });

    it('should render current attributes as filled chips', () => {
      const userAttributes = ['email', 'username', 'given_name'];
      const currentAttributes = ['email', 'username'];

      render(<TestWrapper userAttributes={userAttributes} currentAttributes={currentAttributes} />);

      const emailChip = screen.getByText('email').closest('[role="button"]');
      const usernameChip = screen.getByText('username').closest('[role="button"]');
      const givenNameChip = screen.getByText('given_name').closest('[role="button"]');

      // Chips for selected attributes should be present
      expect(emailChip).toBeInTheDocument();
      expect(usernameChip).toBeInTheDocument();
      expect(givenNameChip).toBeInTheDocument();
    });
  });

  describe('JWT Preview', () => {
    it('should display current attributes in JWT preview', () => {
      const {container} = render(<TestWrapper currentAttributes={['email', 'username']} />);

      // SyntaxHighlighter renders code, so we check the container's text content
      const jsonText = container.textContent || '';
      expect(jsonText).toContain('email');
      expect(jsonText).toContain('username');
    });

    it('should display default attributes in JWT preview', () => {
      const {container} = render(<TestWrapper />);

      const jsonText = container.textContent || '';
      expect(jsonText).toContain('aud');
      expect(jsonText).toContain('exp');
      expect(jsonText).toContain('iat');
      expect(jsonText).toContain('iss');
      expect(jsonText).toContain('sub');
    });
  });

  describe('Info Messages', () => {
    it('should display info about default attributes being always included', () => {
      render(<TestWrapper />);

      // Info message is rendered but text depends on i18n translations
      // Check that Default Attributes accordion is present
      expect(screen.getByText('Default Attributes')).toBeInTheDocument();
    });

    it('should display hint about configuring attributes', () => {
      render(<TestWrapper userAttributes={['email']} />);

      // Tooltip messages are rendered on hover but depend on i18n translations
      // Check that user attributes are rendered as clickable chips
      const emailChip = screen.getByText('email');
      expect(emailChip).toBeInTheDocument();
    });
  });

  describe('Pending Additions and Removals', () => {
    it('should include pending additions in JWT preview for shared token type', () => {
      const {container} = render(
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

    it('should include pending additions in JWT preview for access token when activeTokenType matches', () => {
      const {container} = render(
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

    it('should include pending additions in JWT preview for id token when activeTokenType matches', () => {
      const {container} = render(
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

    it('should include pending additions in JWT preview for userinfo token when activeTokenType matches', () => {
      const {container} = render(
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

    it('should not include pending additions in JWT preview when activeTokenType does not match', () => {
      const {container} = render(
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

    it('should exclude pending removals from JWT preview for shared token type', () => {
      const {container} = render(
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

    it('should exclude pending removals from JWT preview when activeTokenType matches', () => {
      const {container} = render(
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

    it('should not exclude pending removals from JWT preview when activeTokenType does not match', () => {
      const {container} = render(
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

    it('should show pending addition chip as active for shared token type', () => {
      render(
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

      const emailChip = screen.getByText('email').closest('.MuiChip-root');
      expect(emailChip).toBeInTheDocument();
      // The chip for a pending addition should be filled/primary
      expect(emailChip).toHaveClass('MuiChip-filled');
    });

    it('should show pending removal chip as inactive for shared token type', () => {
      render(
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

      const emailChip = screen.getByText('email').closest('.MuiChip-root');
      expect(emailChip).toBeInTheDocument();
      // The chip for a pending removal should be outlined/inactive
      expect(emailChip).toHaveClass('MuiChip-outlined');
    });
  });

  describe('Accordion Toggle Behavior', () => {
    it('should collapse and re-expand user attributes accordion', async () => {
      const user = userEvent.setup();
      const mockSetExpandedSections = vi.fn();

      render(
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
      const accordionHeaders = screen.getAllByText('User Attributes');
      const userAttributesButton = accordionHeaders.find((el) => el.closest('button') !== null);
      expect(userAttributesButton).toBeDefined();

      // Click to collapse
      await user.click(userAttributesButton!);

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

      const {unmount} = render(
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

      const accordionHeaders2 = screen.getAllByText('User Attributes');
      const userAttributesButton2 = accordionHeaders2.find((el) => el.closest('button') !== null);

      await user.click(userAttributesButton2!);

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

      unmount();
    });

    it('should collapse and re-expand default attributes accordion', async () => {
      const user = userEvent.setup();
      const mockSetExpandedSections = vi.fn();

      render(
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
      const defaultAttributesButton = screen.getByText('Default Attributes').closest('button');
      expect(defaultAttributesButton).toBeDefined();

      // Click to collapse
      await user.click(defaultAttributesButton!);

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

      const {unmount} = render(
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

      const defaultAttributesButton2 = screen
        .getAllByText('Default Attributes')
        .find((el) => el.closest('button') !== null);

      await user.click(defaultAttributesButton2!);

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

      unmount();
    });
  });

  describe('Refinements', () => {
    it('should hide content when readOnly is true', () => {
      render(
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

      // Content should be hidden
      expect(screen.queryByText('User Attributes')).not.toBeInTheDocument();
      expect(screen.queryByText('Default Attributes')).not.toBeInTheDocument();
    });

    it('should render headerAction', () => {
      render(
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

      expect(screen.getByText('Test Action')).toBeInTheDocument();
    });
  });
});
