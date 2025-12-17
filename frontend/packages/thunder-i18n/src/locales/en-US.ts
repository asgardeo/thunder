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

/**
 * English (US) translations for Thunder applications
 * All namespaces organized in a single file for better maintainability
 */
const translations = {
  // ============================================================================
  // Common namespace - Shared translations across all Thunder applications
  // ============================================================================
  common: {
    product: {
      displayName: 'Asgardeo',
    },

    // Actions
    actions: {
      add: 'Add',
      edit: 'Edit',
      delete: 'Delete',
      cancel: 'Cancel',
      save: 'Save',
      create: 'Create',
      update: 'Update',
      remove: 'Remove',
      search: 'Search',
      filter: 'Filter',
      reset: 'Reset',
      submit: 'Submit',
      close: 'Close',
      back: 'Back',
      next: 'Next',
      previous: 'Previous',
      confirm: 'Confirm',
      ok: 'OK',
      yes: 'Yes',
      no: 'No',
      continue: 'Continue',
      skip: 'Skip',
      finish: 'Finish',
      done: 'Done',
      refresh: 'Refresh',
      copy: 'Copy',
      download: 'Download',
      upload: 'Upload',
      export: 'Export',
      import: 'Import',
      view: 'View',
      details: 'Details',
      settings: 'Settings',
      logout: 'Logout',
      login: 'Login',
    },

    // Dictionary
    'dictionary.unknown': 'Unknown',

    // Status messages
    status: {
      loading: 'Loading...',
      saving: 'Saving...',
      deleting: 'Deleting...',
      success: 'Success',
      error: 'Error',
      warning: 'Warning',
      info: 'Info',
      pending: 'Pending',
      active: 'Active',
      inactive: 'Inactive',
      enabled: 'Enabled',
      disabled: 'Disabled',
      completed: 'Completed',
      failed: 'Failed',
    },

    // Form labels
    form: {
      name: 'Name',
      description: 'Description',
      email: 'Email',
      password: 'Password',
      username: 'Username',
      required: 'Required',
      optional: 'Optional',
      requiredField: 'This field is required',
      invalidEmail: 'Invalid email address',
      invalidFormat: 'Invalid format',
      searchPlaceholder: 'Search...',
    },

    // Messages
    messages: {
      confirmDelete: 'Are you sure you want to delete this item?',
      deleteSuccess: 'Item deleted successfully',
      deleteError: 'Failed to delete item',
      saveSuccess: 'Saved successfully',
      saveError: 'Failed to save',
      updateSuccess: 'Updated successfully',
      updateError: 'Failed to update',
      createSuccess: 'Created successfully',
      createError: 'Failed to create',
      noData: 'No data available',
      noResults: 'No results found',
      somethingWentWrong: 'Something went wrong',
      tryAgain: 'Please try again',
    },

    // Navigation
    navigation: {
      home: 'Home',
      dashboard: 'Dashboard',
      profile: 'Profile',
      help: 'Help',
      documentation: 'Documentation',
    },

    // User menu
    userMenu: {
      profile: 'Profile',
      myAccount: 'My account',
      addAnotherAccount: 'Add another account',
      settings: 'Settings',
      signOut: 'Sign Out',
    },

    // Header
    header: {
      notifications: 'Coming soon',
      openNotifications: 'Open notifications',
    },

    // Data table - MUI DataGrid locale text
    dataTable: {
      // Root
      noRowsLabel: 'No rows',
      noResultsOverlayLabel: 'No results found.',
      noColumnsOverlayLabel: 'No columns',
      noColumnsOverlayManageColumns: 'Manage columns',

      // Density selector toolbar button text
      toolbarDensity: 'Density',
      toolbarDensityLabel: 'Density',
      toolbarDensityCompact: 'Compact',
      toolbarDensityStandard: 'Standard',
      toolbarDensityComfortable: 'Comfortable',

      // Columns selector toolbar button text
      toolbarColumns: 'Columns',
      toolbarColumnsLabel: 'Select columns',

      // Filters toolbar button text
      toolbarFilters: 'Filters',
      toolbarFiltersLabel: 'Show filters',
      toolbarFiltersTooltipHide: 'Hide filters',
      toolbarFiltersTooltipShow: 'Show filters',
      toolbarFiltersTooltipActive: (count: number) =>
        count !== 1 ? `${count} active filters` : `${count} active filter`,

      // Quick filter toolbar field
      toolbarQuickFilterPlaceholder: 'Search…',
      toolbarQuickFilterLabel: 'Search',
      toolbarQuickFilterDeleteIconLabel: 'Clear',

      // Export selector toolbar button text
      toolbarExport: 'Export',
      toolbarExportLabel: 'Export',
      toolbarExportCSV: 'Download as CSV',
      toolbarExportPrint: 'Print',

      // Columns management text
      columnsManagementSearchTitle: 'Search',
      columnsManagementNoColumns: 'No columns',
      columnsManagementShowHideAllText: 'Show/Hide All',
      columnsManagementReset: 'Reset',

      // Filter panel text
      filterPanelAddFilter: 'Add filter',
      filterPanelRemoveAll: 'Remove all',
      filterPanelDeleteIconLabel: 'Delete',
      filterPanelLogicOperator: 'Logic operator',
      filterPanelOperator: 'Operator',
      filterPanelOperatorAnd: 'And',
      filterPanelOperatorOr: 'Or',
      filterPanelColumns: 'Columns',
      filterPanelInputLabel: 'Value',
      filterPanelInputPlaceholder: 'Filter value',

      // Filter operators text
      filterOperatorContains: 'contains',
      filterOperatorDoesNotContain: 'does not contain',
      filterOperatorEquals: 'equals',
      filterOperatorDoesNotEqual: 'does not equal',
      filterOperatorStartsWith: 'starts with',
      filterOperatorEndsWith: 'ends with',
      filterOperatorIs: 'is',
      filterOperatorNot: 'is not',
      filterOperatorAfter: 'is after',
      filterOperatorOnOrAfter: 'is on or after',
      filterOperatorBefore: 'is before',
      filterOperatorOnOrBefore: 'is on or before',
      filterOperatorIsEmpty: 'is empty',
      filterOperatorIsNotEmpty: 'is not empty',
      filterOperatorIsAnyOf: 'is any of',

      // Filter values text
      filterValueAny: 'any',
      filterValueTrue: 'true',
      filterValueFalse: 'false',

      // Column menu text
      columnMenuLabel: 'Menu',
      columnMenuShowColumns: 'Show columns',
      columnMenuManageColumns: 'Manage columns',
      columnMenuFilter: 'Filter',
      columnMenuHideColumn: 'Hide column',
      columnMenuUnsort: 'Unsort',
      columnMenuSortAsc: 'Sort by ASC',
      columnMenuSortDesc: 'Sort by DESC',

      // Column header text
      columnHeaderFiltersTooltipActive: (count: number) =>
        count !== 1 ? `${count} active filters` : `${count} active filter`,
      columnHeaderFiltersLabel: 'Show filters',
      columnHeaderSortIconLabel: 'Sort',

      // Rows selected footer text
      footerRowSelected: (count: number) =>
        count !== 1 ? `${count.toLocaleString()} rows selected` : `${count.toLocaleString()} row selected`,

      // Total row amount footer text
      footerTotalRows: 'Total Rows:',

      // Total visible row amount footer text
      footerTotalVisibleRows: (visibleCount: number, totalCount: number) =>
        `${visibleCount.toLocaleString()} of ${totalCount.toLocaleString()}`,

      // Checkbox selection text
      checkboxSelectionHeaderName: 'Checkbox selection',
      checkboxSelectionSelectAllRows: 'Select all rows',
      checkboxSelectionUnselectAllRows: 'Unselect all rows',
      checkboxSelectionSelectRow: 'Select row',
      checkboxSelectionUnselectRow: 'Unselect row',

      // Boolean cell text
      booleanCellTrueLabel: 'yes',
      booleanCellFalseLabel: 'no',

      // Actions cell more text
      actionsCellMore: 'more',

      // Column pinning text
      pinToLeft: 'Pin to left',
      pinToRight: 'Pin to right',
      unpin: 'Unpin',

      // Tree Data
      treeDataGroupingHeaderName: 'Group',
      treeDataExpand: 'see children',
      treeDataCollapse: 'hide children',

      // Grouping columns
      groupingColumnHeaderName: 'Group',
      groupColumn: (name: string) => `Group by ${name}`,
      unGroupColumn: (name: string) => `Stop grouping by ${name}`,

      // Master/detail
      detailPanelToggle: 'Detail panel toggle',
      expandDetailPanel: 'Expand',
      collapseDetailPanel: 'Collapse',

      // Pagination
      paginationRowsPerPage: 'Rows per page:',
      paginationDisplayedRows: ({from, to, count}: {from: number; to: number; count: number}) =>
        `${from}–${to} of ${count !== -1 ? count : `more than ${to}`}`,

      // Row reordering text
      rowReorderingHeaderName: 'Row reordering',

      // Aggregation
      aggregationMenuItemHeader: 'Aggregation',
      aggregationFunctionLabelSum: 'sum',
      aggregationFunctionLabelAvg: 'avg',
      aggregationFunctionLabelMin: 'min',
      aggregationFunctionLabelMax: 'max',
      aggregationFunctionLabelSize: 'size',
    },
  },

  // ============================================================================
  // Navigation namespace - Navigation related translations
  // ============================================================================
  navigation: {
    pages: {
      users: 'Users',
      userTypes: 'User Types',
      integrations: 'Integrations',
      applications: 'Applications',
      dashboard: 'Dashboard',
    },
    breadcrumb: {
      develop: 'Develop',
    },
  },

  // ============================================================================
  // Users namespace - User management feature translations
  // ============================================================================
  users: {
    title: 'User Management',
    subtitle: 'Manage users, roles, and permissions across your organization',
    addUser: 'Add User',
    editUser: 'Edit User',
    deleteUser: 'Delete User',
    userDetails: 'User Details',
    firstName: 'First Name',
    lastName: 'Last Name',
    email: 'Email Address',
    username: 'Username',
    role: 'Role',
    status: 'Status',
    createdAt: 'Created At',
    lastLogin: 'Last Login',
    actions: 'Actions',
    noUsers: 'No users found',
    searchUsers: 'Search users...',
    confirmDeleteUser: 'Are you sure you want to delete this user?',
    userCreatedSuccess: 'User created successfully',
    userUpdatedSuccess: 'User updated successfully',
    userDeletedSuccess: 'User deleted successfully',
  },

  // ============================================================================
  // User Types namespace - User types feature translations
  // ============================================================================
  userTypes: {
    title: 'User Types',
    addUserType: 'Add User Type',
    createUserType: 'Create User Type',
    editUserType: 'Edit User Type',
    deleteUserType: 'Delete User Type',
    userTypeDetails: 'User Type Details',
    typeName: 'Type Name',
    typeNamePlaceholder: 'e.g., Employee, Customer, Partner',
    organizationUnit: 'Organization Unit',
    ouSelectPlaceholder: 'Select an organization unit',
    allowSelfRegistration: 'Allow Self Registration',
    description: 'Description',
    createDescription: 'Define a new user type schema for your organization',
    permissions: 'Permissions',
    schemaProperties: 'Schema Properties',
    propertyName: 'Property Name',
    propertyNamePlaceholder: 'e.g., email, age, address',
    propertyType: 'Type',
    addProperty: 'Add Property',
    unique: 'Unique',
    regexPattern: 'Regular Expression Pattern (Optional)',
    regexPlaceholder: 'e.g., ^[a-zA-Z0-9]+$',
    enumValues: 'Allowed Values (Enum) - Optional',
    enumPlaceholder: 'Add value and press Enter',
    types: {
      string: 'String',
      number: 'Number',
      boolean: 'Boolean',
      enum: 'Enum',
      array: 'Array',
      object: 'Object',
    },
    validationErrors: {
      nameRequired: 'Please enter a user type name',
      ouIdRequired: 'Please provide an organization unit ID',
      propertiesRequired: 'Please add at least one property',
      duplicateProperties: 'Duplicate property names found: {{duplicates}}',
    },
    errors: {
      organizationUnitsFailedTitle: 'Failed to load organization units',
    },
    noUserTypes: 'No user types found',
    noOrganizationUnits: 'No organization units available',
    confirmDeleteUserType: 'Are you sure you want to delete this user type?',
  },

  // ============================================================================
  // Integrations namespace - Integrations feature translations
  // ============================================================================
  integrations: {
    title: 'Integrations',
    subtitle: 'Manage your integrations and connections',
    addIntegration: 'Add Integration',
    editIntegration: 'Edit Integration',
    deleteIntegration: 'Delete Integration',
    integrationDetails: 'Integration Details',
    provider: 'Provider',
    apiKey: 'API Key',
    endpoint: 'Endpoint',
    status: 'Status',
    connected: 'Connected',
    disconnected: 'Disconnected',
    testConnection: 'Test Connection',
    noIntegrations: 'No integrations found',
    comingSoon: 'Coming Soon',
    comingSoonDescription: 'Integrations management functionality will be available soon.',
  },

  // ============================================================================
  // Dashboard namespace - Dashboard feature translations
  // ============================================================================
  dashboard: {
    welcomeMessage: 'Welcome to Thunder Develop',
    totalUsers: 'Total Users',
    activeUsers: 'Active Users',
    totalApplications: 'Total Applications',
    recentActivity: 'Recent Activity',
    quickActions: 'Quick Actions',
  },

  // ============================================================================
  // Authentication namespace - Authentication feature translations
  // ============================================================================
  auth: {
    signIn: 'Sign In',
    signUp: 'Sign Up',
    signOut: 'Sign Out',
    forgotPassword: 'Forgot Password?',
    resetPassword: 'Reset Password',
    changePassword: 'Change Password',
    rememberMe: 'Remember Me',
    welcomeBack: 'Welcome Back',
    createAccount: 'Create Account',
    alreadyHaveAccount: 'Already have an account?',
    dontHaveAccount: "Don't have an account?",
    enterEmail: 'Enter your email',
    enterPassword: 'Enter your password',
    confirmPassword: 'Confirm Password',
    passwordMismatch: 'Passwords do not match',
    invalidCredentials: 'Invalid credentials',
    accountLocked: 'Account is locked',
    sessionExpired: 'Session expired. Please sign in again.',
    signInSuccess: 'Signed in successfully',
    signUpSuccess: 'Account created successfully',
    passwordResetSent: 'Password reset link sent to your email',
    passwordResetSuccess: 'Password reset successfully',
  },

  // ============================================================================
  // MFA namespace - Multi-factor authentication feature translations
  // ============================================================================
  mfa: {
    title: 'Multi-Factor Authentication',
    setupMfa: 'Set Up MFA',
    enableMfa: 'Enable MFA',
    disableMfa: 'Disable MFA',
    verificationCode: 'Verification Code',
    enterCode: 'Enter verification code',
    sendCode: 'Send Code',
    resendCode: 'Resend Code',
    invalidCode: 'Invalid verification code',
    codeExpired: 'Verification code expired',
    scanQrCode: 'Scan this QR code with your authenticator app',
    backupCodes: 'Backup Codes',
    saveBackupCodes: 'Save these backup codes in a secure place',
  },

  // ============================================================================
  // Social namespace - Social login feature translations
  // ============================================================================
  social: {
    continueWith: 'Continue with',
    signInWith: 'Sign in with',
    google: 'Google',
    facebook: 'Facebook',
    github: 'GitHub',
    microsoft: 'Microsoft',
  },

  // ============================================================================
  // Consent namespace - Consent feature translations
  // ============================================================================
  consent: {
    title: 'Consent Required',
    message: 'The application is requesting access to your information',
    requestedPermissions: 'Requested Permissions',
    allow: 'Allow',
    deny: 'Deny',
    learnMore: 'Learn More',
  },

  // ============================================================================
  // Errors namespace - Error messages translations
  // ============================================================================
  errors: {
    authenticationFailed: 'Authentication failed',
    unauthorizedAccess: 'Unauthorized access',
    accessDenied: 'Access denied',
    invalidRequest: 'Invalid request',
    serverError: 'Server error occurred',
    networkError: 'Network error. Please check your connection.',
    redirectFailed: 'Redirect failed',
  },

  // ============================================================================
  // Applications - Applications feature translations
  // ============================================================================
  applications: {
    'listing.title': 'Applications',
    'listing.subtitle': 'Manage your applications and services',
    'listing.addApplication': 'Add Application',
    'listing.columns.name': 'Name',
    'listing.columns.clientId': 'Client ID',
    'listing.columns.actions': 'Actions',
    'listing.search.placeholder': 'Search ..',
    'delete.title': 'Delete Application',
    'delete.message': 'Are you sure you want to delete this application? This action cannot be undone.',
    'delete.disclaimer': 'Warning: All associated data, configurations, and access tokens will be permanently removed.',
    'onboarding.preview.title': 'Preview',
    'onboarding.preview.signInTo': 'Sign in to {{appName}}',
    'onboarding.preview.welcomeMessage': 'Welcome back! Please sign in to continue.',
    'onboarding.preview.username': 'Username',
    'onboarding.preview.usernamePlaceholder': 'Enter your Username',
    'onboarding.preview.password': 'Password',
    'onboarding.preview.passwordPlaceholder': 'Enter your Password',
    'onboarding.preview.signInButton': 'Sign In',
    'onboarding.preview.dividerText': 'or',
    'onboarding.preview.continueWith': 'Continue with {{providerName}}',
    'onboarding.steps.name': 'Create an Application',
    'onboarding.steps.design': 'Design',
    'onboarding.steps.options': 'Sign In Options',
    'onboarding.steps.approach': 'Sign-In Experience',
    'onboarding.steps.stack': 'Technology Stack',
    'onboarding.steps.configure': 'Configuration',
    'onboarding.steps.summary': 'Summary',
    'onboarding.configure.name.title': "Let's give a name to your application",
    'onboarding.configure.name.fieldLabel': 'Application Name',
    'onboarding.configure.name.placeholder': 'Enter your application name',
    'onboarding.configure.name.suggestions.label': 'In a hurry? Pick a random name:',
    'onboarding.configure.design.title': 'Design Your Application',
    'onboarding.configure.design.subtitle': 'Customize the appearance of your application',
    'onboarding.configure.design.logo.title': 'Application Logo',
    'onboarding.configure.design.logo.shuffle': 'Shuffle',
    'onboarding.configure.design.color.title': 'Brand Color',
    'onboarding.configure.design.color.customLabel': 'Custom',
    'onboarding.configure.design.color.defaultBranding.withAppName': 'will use the default brand color',
    'onboarding.configure.design.color.defaultBranding.withoutAppName': 'Using the default brand color',
    'onboarding.configure.design.color.pickDifferent': 'Pick a different color',
    'onboarding.configure.SignInOptions.title': 'Sign In Options',
    'onboarding.configure.SignInOptions.subtitle': 'Choose how users will sign-in to your application',
    'onboarding.configure.SignInOptions.usernamePassword': 'Username & Password',
    'onboarding.configure.SignInOptions.google': 'Google',
    'onboarding.configure.SignInOptions.github': 'GitHub',
    'onboarding.configure.SignInOptions.notConfigured': 'Not configured',
    'onboarding.configure.SignInOptions.noSelectionWarning':
      'At least one sign-in option is required. Please select at least one authentication method.',
    'onboarding.configure.SignInOptions.noIntegrations':
      'No social sign-in integrations available. Please configure an integration first.',
    'onboarding.configure.SignInOptions.hint':
      'You can always change these settings later in the application settings.',
    'onboarding.configure.SignInOptions.loading': 'Loading...',
    'onboarding.configure.SignInOptions.error': 'Failed to load authentication methods: {{error}}',
    'onboarding.configure.approach.title': 'Sign-In Experience',
    'onboarding.configure.approach.subtitle': 'Select how users will authenticate in your application',
    'onboarding.configure.approach.inbuilt.title': 'Redirect to {{product}} Default Pages',
    'onboarding.configure.approach.inbuilt.description':
      "Users are redirected to {{product}}'s default sign-in pages. Easy to set up and you can customize the look and feel.",
    'onboarding.configure.approach.native.title': 'Bring Your Own Sign-In',
    'onboarding.configure.approach.native.description':
      'Build a fully custom authentication experience within your application. Gives you maximum flexibility over the user journey.',
    'onboarding.configure.stack.technology.title': 'Technology',
    'onboarding.configure.stack.technology.subtitle': 'What technology are you using to build your application?',
    'onboarding.configure.stack.technology.react.title': 'React',
    'onboarding.configure.stack.technology.react.description': 'Single-page application built with React',
    'onboarding.configure.stack.technology.nextjs.title': 'Next.js',
    'onboarding.configure.stack.technology.nextjs.description': 'Full-stack React framework with server-side rendering',
    'onboarding.configure.stack.technology.angular.title': 'Angular',
    'onboarding.configure.stack.technology.angular.description': 'Single-page application built with Angular',
    'onboarding.configure.stack.technology.vue.title': 'Vue',
    'onboarding.configure.stack.technology.vue.description': 'Single-page application built with Vue.js',
    'onboarding.configure.stack.technology.ios.title': 'iOS',
    'onboarding.configure.stack.technology.ios.description': 'Native iOS application (Swift or Objective-C)',
    'onboarding.configure.stack.technology.android.title': 'Android',
    'onboarding.configure.stack.technology.android.description': 'Native Android application (Kotlin or Java)',
    'onboarding.configure.stack.technology.springboot.title': 'Spring Boot',
    'onboarding.configure.stack.technology.springboot.description': 'Java backend application with Spring Boot',
    'onboarding.configure.stack.technology.nodejs.title': 'Node.js',
    'onboarding.configure.stack.technology.nodejs.description': 'Backend service built with Node.js',
    'onboarding.configure.stack.platform.title': 'Application Type',
    'onboarding.configure.stack.platform.subtitle': 'This helps us configure the right settings for your app',
    'onboarding.configure.stack.dividerLabel': 'OR',
    'onboarding.configure.stack.platform.browser.title': 'Browser App',
    'onboarding.configure.stack.platform.browser.description': 'Single-page apps running in browsers',
    'onboarding.configure.stack.platform.server.title': 'Full-Stack App',
    'onboarding.configure.stack.platform.server.description': 'Apps with both server and client code',
    'onboarding.configure.stack.platform.mobile.title': 'Mobile App',
    'onboarding.configure.stack.platform.mobile.description': 'Native or hybrid mobile applications',
    'onboarding.configure.stack.platform.backend.title': 'Backend Service',
    'onboarding.configure.stack.platform.backend.description': 'Server-to-server APIs and services',
    'onboarding.configure.details.title': 'Configuration',
    'onboarding.configure.details.description': 'Configure where your application is hosted and callback settings',
    'onboarding.configure.details.hostingUrl.label': 'Where is your application hosted?',
    'onboarding.configure.details.hostingUrl.placeholder': 'https://myapp.example.com',
    'onboarding.configure.details.hostingUrl.helperText': 'The URL where users will access your application',
    'onboarding.configure.details.hostingUrl.error.required': 'Application hosting URL is required',
    'onboarding.configure.details.hostingUrl.error.invalid':
      'Please enter a valid URL (must start with http:// or https://)',
    'onboarding.configure.details.callbackUrl.label': 'After Sign-in URL (Optional)',
    'onboarding.configure.details.callbackUrl.placeholder': 'https://myapp.example.com/callback',
    'onboarding.configure.details.callbackUrl.helperText': 'The URL where users will be redirected after signing in',
    'onboarding.configure.details.callbackUrl.error.required': 'After sign-in URL is required',
    'onboarding.configure.details.callbackUrl.error.invalid':
      'Please enter a valid URL (must start with http:// or https://)',
    'onboarding.configure.details.callbackUrl.info':
      'The URL is where users will be redirected after sign-in. For most applications, using the same URL as your application access URL is recommended.',
    'onboarding.configure.details.sameAsHosting': 'Same as the application URL',
    'onboarding.configure.details.callbackMode.same': 'Same as Application Access URL',
    'onboarding.configure.details.callbackMode.custom': 'Custom URL',
    'onboarding.configure.details.mobile.title': 'Mobile Application Configuration',
    'onboarding.configure.details.mobile.description':
      'Configure the deep link or universal link for your mobile application',
    'onboarding.configure.details.mobile.info':
      'Deep links (e.g., myapp://callback) or universal links (e.g., https://example.com/callback) are used to redirect users back to your mobile app after authentication.',
    'onboarding.configure.details.deeplink.label': 'Deep Link / Universal Link',
    'onboarding.configure.details.deeplink.placeholder': 'myapp://callback or https://example.com/callback',
    'onboarding.configure.details.deeplink.helperText': 'The custom URL scheme or universal link for your mobile app',
    'onboarding.configure.details.noConfigRequired.title': 'No Additional Configuration Needed',
    'onboarding.configure.details.noConfigRequired.description':
      'Your application is ready to go! You can proceed to the next step.',
    'onboarding.configure.details.userTypes.label': 'User Types',
    'onboarding.configure.details.userTypes.description': 'Select which user types can access this application',
    'onboarding.configure.details.userTypes.error': 'Please select at least one user type',
    'onboarding.configure.setup.title': 'Application Setup',
    'onboarding.configure.setup.subtitle': 'Select the technology stack for your application',
    'onboarding.configure.setup.platform.label': 'What technology are you using?',
    'onboarding.configure.setup.platform.browser.title': 'Browser',
    'onboarding.configure.setup.platform.browser.description': 'Single page apps (React, Vue, Angular)',
    'onboarding.configure.setup.platform.server.title': 'Server + Browser',
    'onboarding.configure.setup.platform.server.description': 'Full-stack apps (Next.js, Remix)',
    'onboarding.configure.setup.platform.mobile.title': 'Mobile Device',
    'onboarding.configure.setup.platform.mobile.description': 'iOS, Android, React Native',
    'onboarding.configure.setup.platform.desktop.title': 'Desktop',
    'onboarding.configure.setup.platform.desktop.description': 'Electron, Tauri apps',
    'onboarding.configure.setup.platform.backend.title': 'Backend Service',
    'onboarding.configure.setup.platform.backend.description': 'Server-to-server APIs',
    'onboarding.configure.setup.info':
      'Your {{platform}} configuration is set up automatically. You can customize these settings later.',
    'onboarding.configure.oauth.title': 'Configure OAuth',
    'onboarding.configure.oauth.subtitle': 'Configure OAuth2/OIDC settings for your application (optional)',
    'onboarding.configure.oauth.optional': 'This step is optional.',
    'onboarding.configure.oauth.hostedGuidance':
      'OAuth configuration is recommended for redirect-based authentication. Configure OAuth settings to enable secure authentication flows.',
    'onboarding.configure.oauth.nativeGuidance':
      'OAuth configuration is optional for custom sign-in UI. You can skip this step and use the Flow API for authentication instead.',
    'onboarding.configure.oauth.publicClient.label': 'Public Client',
    'onboarding.configure.oauth.pkce.label': 'Enable PKCE',
    'onboarding.configure.oauth.redirectURIs.fieldLabel': 'Redirect URIs',
    'onboarding.configure.oauth.redirectURIs.placeholder': 'https://localhost:3000/callback',
    'onboarding.configure.oauth.redirectURIs.addButton': 'Add',
    'onboarding.configure.oauth.redirectURIs.errors.empty': 'Please enter a redirect URI',
    'onboarding.configure.oauth.redirectURIs.errors.invalid':
      'Please enter a valid URL (must start with http:// or https://)',
    'onboarding.configure.oauth.redirectURIs.errors.duplicate': 'This redirect URI has already been added',
    'onboarding.configure.oauth.grantTypes.label': 'Grant Types',
    'onboarding.configure.oauth.grantTypes.authorizationCode': 'Authorization Code',
    'onboarding.configure.oauth.grantTypes.refreshToken': 'Refresh Token',
    'onboarding.configure.oauth.grantTypes.clientCredentials': 'Client Credentials',
    'onboarding.configure.oauth.tokenEndpointAuthMethod.label': 'Token Endpoint Authentication Method',
    'onboarding.configure.oauth.tokenEndpointAuthMethod.clientSecretBasic': 'Client Secret Basic',
    'onboarding.configure.oauth.tokenEndpointAuthMethod.clientSecretPost': 'Client Secret Post',
    'onboarding.configure.oauth.tokenEndpointAuthMethod.none': 'None',
    'onboarding.configure.oauth.errors.publicClientRequiresPKCE':
      'Public clients must have PKCE enabled. PKCE is automatically enabled for public clients.',
    'onboarding.configure.oauth.errors.publicClientRequiresNone':
      'Public clients must use "None" as the token endpoint authentication method.',
    'onboarding.configure.oauth.errors.publicClientNoClientCredentials':
      'Public clients cannot use the client_credentials grant type.',
    'onboarding.configure.oauth.errors.authorizationCodeRequiresRedirectURIs':
      'Authorization Code grant type requires at least one redirect URI.',
    'onboarding.configure.oauth.errors.clientCredentialsRequiresAuth':
      'Client Credentials grant type cannot use "None" authentication method.',
    'onboarding.configure.oauth.errors.atLeastOneGrantTypeRequired': 'At least one grant type must be selected.',
    'onboarding.configure.oauth.errors.refreshTokenRequiresAuthorizationCode':
      'Refresh Token grant type requires Authorization Code grant type to be selected.',
    'onboarding.creating': 'Creating...',
    'onboarding.skipAndCreate': 'Skip & Create',
    'onboarding.createApplication': 'Create Application',
    'onboarding.summary.title': 'Application Created!',
    'onboarding.summary.subtitle': 'Your application has been successfully created and is ready to use.',
    'onboarding.summary.appDetails': 'Application is ready to use',
    'onboarding.summary.viewApplication': 'View Application',
    'onboarding.summary.viewAppAriaLabel': 'View application details',
    'clientSecret.warning': 'Please copy your client credentials now. The client secret will not be shown again.',
    'clientSecret.clientIdLabel': 'Client ID',
    'clientSecret.clientSecretLabel': 'Client Secret',
    'clientSecret.copied': 'Copied!',
    'view.title': 'Application Details',
    'view.subtitle': 'View application details and configuration',
    'view.back': 'Back to Applications',
    'view.notFound': 'Application not found',
    'view.error': 'Failed to load application information',
    'view.sections.basicInformation': 'Basic Information',
    'view.sections.flowConfiguration': 'Flow Configuration',
    'view.sections.userAttributes': 'User Attributes',
    'view.sections.oauth2Configuration': 'OAuth2 Configuration',
    'view.sections.timestamps': 'Timestamps',
    'view.fields.applicationId': 'Application ID',
    'view.fields.description': 'Description',
    'view.fields.url': 'URL',
    'view.fields.tosUri': 'Terms of Service URI',
    'view.fields.policyUri': 'Privacy Policy URI',
    'view.fields.contacts': 'Contacts',
    'view.fields.authFlowGraphId': 'Authentication Flow Graph ID',
    'view.fields.registrationFlowGraphId': 'Registration Flow Graph ID',
    'view.fields.registrationFlowEnabled': 'Registration Flow Enabled',
    'view.fields.clientId': 'Client ID',
    'view.fields.redirectUris': 'Redirect URIs',
    'view.fields.grantTypes': 'Grant Types',
    'view.fields.responseTypes': 'Response Types',
    'view.fields.scopes': 'Scopes',
    'view.fields.publicClient': 'Public Client',
    'view.fields.pkceRequired': 'PKCE Required',
    'view.fields.createdAt': 'Created At',
    'view.fields.updatedAt': 'Updated At',
    'view.values.yes': 'Yes',
    'view.values.no': 'No',
  },
} as const;

export default translations;
