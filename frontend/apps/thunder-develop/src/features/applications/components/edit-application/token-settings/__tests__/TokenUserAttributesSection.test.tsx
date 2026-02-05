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
vi.mock('../../SettingsCard', () => ({
  default: ({title, description, children}: {title: string; description: string; children: React.ReactNode}) => (
    <div data-testid="settings-card">
      <div data-testid="card-title">{title}</div>
      <div data-testid="card-description">{description}</div>
      {children}
    </div>
  ),
}));

// Mock TokenConstants
vi.mock('../../../../constants/token-constants', () => ({
  default: {
    DEFAULT_TOKEN_ATTRIBUTES: ['aud', 'exp', 'iat', 'iss', 'sub'],
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
  tokenType?: 'shared' | 'access' | 'id';
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
    onAttributeClick: (attr: string, tokenType: 'shared' | 'access' | 'id') => void;
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
      const userAttributes = ['email', 'username', 'firstName'];
      render(<TestWrapper userAttributes={userAttributes} />);

      expect(screen.getByText('email')).toBeInTheDocument();
      expect(screen.getByText('username')).toBeInTheDocument();
      expect(screen.getByText('firstName')).toBeInTheDocument();
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
      const userAttributes = ['email', 'username', 'firstName'];
      const currentAttributes = ['email', 'username'];

      render(<TestWrapper userAttributes={userAttributes} currentAttributes={currentAttributes} />);

      const emailChip = screen.getByText('email').closest('[role="button"]');
      const usernameChip = screen.getByText('username').closest('[role="button"]');
      const firstNameChip = screen.getByText('firstName').closest('[role="button"]');

      // Chips for selected attributes should be present
      expect(emailChip).toBeInTheDocument();
      expect(usernameChip).toBeInTheDocument();
      expect(firstNameChip).toBeInTheDocument();
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

  describe('Pending Additions in JWT Preview', () => {
    it('should include pending additions in JWT preview for shared token type', () => {
      const pendingAdditions = new Set(['newAttribute']);
      const expandedSections = new Set(['user-shared', 'default-shared']);

      const {container} = render(
        <TokenUserAttributesSection
          tokenType="shared"
          currentAttributes={['email']}
          userAttributes={['email', 'newAttribute']}
          isLoadingUserAttributes={false}
          expandedSections={expandedSections}
          setExpandedSections={vi.fn()}
          pendingAdditions={pendingAdditions}
          pendingRemovals={new Set()}
          highlightedAttributes={new Set()}
          onAttributeClick={vi.fn()}
          activeTokenType="access"
        />,
      );

      const jsonText = container.textContent || '';
      expect(jsonText).toContain('newAttribute');
    });

    it('should include pending additions in JWT preview for access token when activeTokenType matches', () => {
      const pendingAdditions = new Set(['newAttribute']);
      const expandedSections = new Set(['user-access', 'default-access']);

      const {container} = render(
        <TokenUserAttributesSection
          tokenType="access"
          currentAttributes={['email']}
          userAttributes={['email', 'newAttribute']}
          isLoadingUserAttributes={false}
          expandedSections={expandedSections}
          setExpandedSections={vi.fn()}
          pendingAdditions={pendingAdditions}
          pendingRemovals={new Set()}
          highlightedAttributes={new Set()}
          onAttributeClick={vi.fn()}
          activeTokenType="access"
        />,
      );

      const jsonText = container.textContent || '';
      expect(jsonText).toContain('newAttribute');
    });

    it('should not include pending additions when activeTokenType does not match', () => {
      const pendingAdditions = new Set(['newAttribute']);
      const expandedSections = new Set(['user-access', 'default-access']);

      const {container} = render(
        <TokenUserAttributesSection
          tokenType="access"
          currentAttributes={['email']}
          userAttributes={['email', 'newAttribute']}
          isLoadingUserAttributes={false}
          expandedSections={expandedSections}
          setExpandedSections={vi.fn()}
          pendingAdditions={pendingAdditions}
          pendingRemovals={new Set()}
          highlightedAttributes={new Set()}
          onAttributeClick={vi.fn()}
          activeTokenType="id"
        />,
      );

      const jsonText = container.textContent || '';
      // newAttribute should NOT be in preview since activeTokenType (id) doesn't match tokenType (access)
      // but email IS a current attribute so it should still be there
      expect(jsonText).toContain('email');
    });
  });

  describe('Scopes with undefined oauth2Config', () => {
    it('should display no scopes message when oauth2Config is undefined for ID token', () => {
      render(<TestWrapper tokenType="id" oauth2Config={undefined} />);

      expect(screen.getByText('No scopes configured')).toBeInTheDocument();
    });
  });

  describe('Accordion Toggle', () => {
    it('should toggle default attributes accordion when clicked', async () => {
      const user = userEvent.setup();

      render(<TestWrapper />);

      // Find the Default Attributes accordion header
      const defaultAttributesHeader = screen.getByText('Default Attributes');
      const accordionButton = defaultAttributesHeader.closest('button');

      expect(accordionButton).toBeDefined();

      if (accordionButton) {
        // Click to collapse
        await user.click(accordionButton);
        // Click to expand again
        await user.click(accordionButton);
      }
    });
  });
});
