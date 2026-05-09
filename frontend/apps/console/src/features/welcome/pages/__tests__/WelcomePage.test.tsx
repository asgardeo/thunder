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

import {render, screen, userEvent} from '@thunderid/test-utils';
import {afterEach, beforeEach, describe, expect, it, vi} from 'vitest';

const mockNavigate = vi.fn();
const mockSessionStorageSetItem = vi.fn();

vi.mock('@thunderid/contexts', async (importOriginal) => {
  const actual = await importOriginal<typeof import('@thunderid/contexts')>();
  return {
    ...actual,
    useConfig: () => ({
      config: {
        brand: {
          product_name: 'ThunderID',
          favicon: {light: 'assets/images/favicon.ico', dark: 'assets/images/favicon-inverted.ico'},
        },
      },
    }),
  };
});

vi.mock('react-i18next', () => ({
  useTranslation: () => ({
    t: (key: string, opts?: Record<string, unknown>) => {
      if (opts?.productName) {
        const productName = opts.productName as string;
        return `${key}:${productName}`;
      }
      return key;
    },
  }),
}));

vi.mock('react-router', async () => {
  const actual = await vi.importActual<typeof import('react-router')>('react-router');
  return {...actual, useNavigate: () => mockNavigate};
});

vi.mock('@wso2/oxygen-ui', async (importOriginal) => {
  const actual = await importOriginal<typeof import('@wso2/oxygen-ui')>();
  return {
    ...actual,
    useTheme: () => ({palette: {mode: 'light'}}),
  };
});

vi.mock('framer-motion', () => ({
  motion: {
    create: (Component: React.ElementType) => Component,
  },
}));

vi.mock('@wso2/oxygen-ui-icons-react', async (importOriginal) => {
  const actual = await importOriginal<typeof import('@wso2/oxygen-ui-icons-react')>();
  return {
    ...actual,
    FolderOpen: () => <span data-testid="icon-folder-open" />,
    X: () => <span data-testid="icon-x" />,
    ChevronRight: () => <span data-testid="icon-chevron-right" />,
    BookOpen: () => <span data-testid="icon-book-open" />,
    Lightbulb: () => <span data-testid="icon-lightbulb" />,
    PackagePlus: () => <span data-testid="icon-package-plus" />,
  };
});

import WelcomePage from '../WelcomePage';

describe('WelcomePage', () => {
  beforeEach(() => {
    vi.stubGlobal('sessionStorage', {
      setItem: mockSessionStorageSetItem,
      getItem: vi.fn(),
      removeItem: vi.fn(),
      clear: vi.fn(),
    });
    vi.stubGlobal('open', vi.fn());
  });

  afterEach(() => {
    vi.clearAllMocks();
    vi.unstubAllGlobals();
  });

  it('renders without crashing', () => {
    const {container} = render(<WelcomePage />);
    expect(container).toBeInTheDocument();
  });

  it('renders close button', () => {
    render(<WelcomePage />);
    expect(screen.getByRole('button', {name: /common:actions\.close/i})).toBeInTheDocument();
  });

  it('navigates to /home and sets session storage on close', async () => {
    const user = userEvent.setup();
    render(<WelcomePage />);

    await user.click(screen.getByRole('button', {name: /common:actions\.close/i}));

    expect(mockSessionStorageSetItem).toHaveBeenCalledWith('thunderid:welcome:dismissed', 'true');
    expect(mockNavigate).toHaveBeenCalledWith('/home');
  });

  it('navigates to /welcome/create-project on new project click', async () => {
    const user = userEvent.setup();
    render(<WelcomePage />);

    const newProjectButton = screen.getByText('common:welcome.start.newProject');
    await user.click(newProjectButton.closest('[role="button"]') ?? newProjectButton);

    expect(mockSessionStorageSetItem).toHaveBeenCalledWith('thunderid:welcome:dismissed', 'true');
    expect(mockNavigate).toHaveBeenCalledWith('/welcome/create-project');
  });

  it('renders start action items', () => {
    render(<WelcomePage />);
    expect(screen.getByText('common:welcome.start.newProject')).toBeInTheDocument();
    expect(screen.getByText('common:welcome.start.openImport')).toBeInTheDocument();
  });

  it('renders walkthrough items', () => {
    render(<WelcomePage />);
    expect(screen.getByText('common:welcome.walkthrough.getStartedDesigner')).toBeInTheDocument();
    expect(screen.getByText('common:welcome.walkthrough.learnFundamentals')).toBeInTheDocument();
  });

  it('opens external link with noopener,noreferrer on walkthrough click', async () => {
    const mockOpen = vi.fn();
    vi.stubGlobal('open', mockOpen);
    const user = userEvent.setup();
    render(<WelcomePage />);

    const getStartedButton = screen.getByText('common:welcome.walkthrough.getStartedDesigner');
    await user.click(getStartedButton.closest('[role="button"]') ?? getStartedButton);

    expect(mockOpen).toHaveBeenCalledWith(expect.stringContaining('quickstart'), '_blank', 'noopener,noreferrer');
  });

  it('uses product name from config', () => {
    render(<WelcomePage />);
    // The openImportDesc key is interpolated with productName
    expect(screen.getByText(/openImportDesc.*ThunderID/i)).toBeInTheDocument();
  });
});
