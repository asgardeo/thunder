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

    // Short action words (used as button labels, etc.)
    show: 'Show',
    publish: 'Publish',
    saveDraft: 'Save Draft',
    common: 'Common',
    new: 'New',
    edit: 'Edit',
    delete: 'Delete',
    close: 'Close',
    back: 'Back',
    create: 'Create',
    update: 'Update',
    save: 'Save',

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
    'onboarding.steps.name': 'Create an Application',
    'onboarding.steps.design': 'Design',
    'onboarding.steps.options': 'Sign In Options',
    'onboarding.steps.configure': 'Configure OAuth',
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
    'onboarding.configure.SignInOptions.title': 'Sign In Options',
    'onboarding.configure.SignInOptions.subtitle': 'Choose how users will sign-in to your application',
    'onboarding.configure.SignInOptions.usernamePassword': 'Username & Password',
    'onboarding.configure.SignInOptions.google': 'Google',
    'onboarding.configure.SignInOptions.github': 'GitHub',
    'onboarding.configure.SignInOptions.notConfigured': 'Not configured',
    'onboarding.configure.SignInOptions.noSelectionWarning':
      'At least one login option is required. Please select at least one authentication method.',
    'onboarding.configure.SignInOptions.noIntegrations':
      'No social login integrations available. Please configure an integration first.',
    'onboarding.configure.SignInOptions.hint':
      'You can always change these settings later in the application settings.',
    'onboarding.configure.SignInOptions.loading': 'Loading...',
    'onboarding.configure.SignInOptions.error': 'Failed to load authentication methods: {{error}}',
    'onboarding.configure.oauth.title': 'Configure OAuth',
    'onboarding.configure.oauth.subtitle': 'Configure OAuth2/OIDC settings for your application (optional)',
    'onboarding.configure.oauth.optional': 'This step is optional.',
    'onboarding.configure.oauth.publicClient.label': 'Public Client',
    'onboarding.configure.oauth.pkce.label': 'Enable PKCE',
    'onboarding.configure.oauth.redirectURIs.fieldLabel': 'Redirect URIs',
    'onboarding.configure.oauth.redirectURIs.placeholder': 'https://localhost:3000/callback',
    'onboarding.configure.oauth.redirectURIs.addButton': 'Add',
    'onboarding.configure.oauth.redirectURIs.errors.empty': 'Please enter a redirect URI',
    'onboarding.configure.oauth.redirectURIs.errors.invalid': 'Please enter a valid URL (must start with http:// or https://)',
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
    'onboarding.configure.oauth.errors.atLeastOneGrantTypeRequired':
      'At least one grant type must be selected.',
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
    'clientSecret.warning':
      'Please copy your client credentials now. The client secret will not be shown again.',
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

  // ============================================================================
  // Flows namespace - Flow builder feature translations
  // ============================================================================
  flows: {
    // Flow labels and navigation
    label: 'Flows',
    'core.breadcrumb': '{{flowType}}',
    'core.autoSave.savingInProgress': 'Saving...',
    'core.labels.enableFlow': 'Enable Flow',
    'core.labels.disableFlow': 'Disable Flow',
    'core.tooltips.enableFlow': 'Enable this {{flowType}}',
    'core.tooltips.disableFlow': 'Disable this {{flowType}}',

    // Notification panel
    'core.notificationPanel.header': 'Notifications',
    'core.notificationPanel.trigger.label': 'View notifications',
    'core.notificationPanel.tabs.errors': 'Errors',
    'core.notificationPanel.tabs.warnings': 'Warnings',
    'core.notificationPanel.tabs.info': 'Info',
    'core.notificationPanel.emptyMessages.errors': 'No errors found',
    'core.notificationPanel.emptyMessages.warnings': 'No warnings found',
    'core.notificationPanel.emptyMessages.info': 'No information messages',

    // Execution steps - names
    'core.executions.names.google': 'Google',
    'core.executions.names.apple': 'Apple',
    'core.executions.names.github': 'GitHub',
    'core.executions.names.facebook': 'Facebook',
    'core.executions.names.microsoft': 'Microsoft',
    'core.executions.names.passkeyEnrollment': 'Passkey Enrollment',
    'core.executions.names.confirmationCode': 'Confirmation Code',
    'core.executions.names.magicLink': 'Magic Link',
    'core.executions.names.sendEmailOTP': 'Send Email OTP',
    'core.executions.names.verifyEmailOTP': 'Verify Email OTP',
    'core.executions.names.sendSMS': 'Send SMS',
    'core.executions.names.verifySMSOTP': 'Verify SMS OTP',
    'core.executions.names.default': 'Execution',

    // Execution steps - tooltips and messages
    'core.executions.tooltip.configurationHint': 'Click to configure this step',
    'core.executions.landing.message': 'This {{executor}} step will redirect users to a landing page.',

    // Steps - end
    'core.steps.end.flowCompletionProperties': 'Flow Completion Properties',

    // Validation messages - input fields
    'core.validation.fields.input.general':
      'Required fields are not properly configured for the input field with ID <code>{{id}}</code>.',
    'core.validation.fields.input.idpName': 'Identity provider name is required',
    'core.validation.fields.input.label': 'Label is required',
    'core.validation.fields.input.identifier': 'Identifier is required',

    // Validation messages - button
    'core.validation.fields.button.text': 'Button text is required',
    'core.validation.fields.button.action': 'Button action is required',

    // Validation messages - checkbox
    'core.validation.fields.checkbox.label': 'Label is required',
    'core.validation.fields.checkbox.identifier': 'Identifier is required',

    // Validation messages - divider
    'core.validation.fields.divider.variant': 'Divider variant is required',

    // Validation messages - typography
    'core.validation.fields.typography.text': 'Text content is required',
    'core.validation.fields.typography.variant': 'Typography variant is required',

    // Validation messages - image
    'core.validation.fields.image.src': 'Image source is required',
    'core.validation.fields.image.variant': 'Image variant is required',

    // Validation messages - rich text
    'core.validation.fields.richText.text': 'Rich text content is required',

    // Validation messages - OTP input
    'core.validation.fields.otpInput.label': 'OTP input label is required',

    // Validation messages - phone number input
    'core.validation.fields.phoneNumberInput.label': 'Phone number label is required',
    'core.validation.fields.phoneNumberInput.identifier': 'Phone number identifier is required',

    // Elements - rich text
    'core.elements.richText.placeholder': 'Enter text here...',
    'core.elements.richText.linkEditor.urlTypeLabel': 'URL Type',
    'core.elements.richText.linkEditor.placeholder': 'Enter URL',

    // Elements - text property field
    'core.elements.textPropertyField.placeholder': 'Enter {{propertyName}}',
    'core.elements.textPropertyField.tooltip.configureTranslation': 'Configure translation',

    // Elements - i18n card
    'core.elements.textPropertyField.i18nCard.title': 'Translation for {{field}}',
    'core.elements.textPropertyField.i18nCard.createTitle': 'Create Translation',
    'core.elements.textPropertyField.i18nCard.updateTitle': 'Update Translation',
    'core.elements.textPropertyField.i18nCard.i18nKey': 'Translation Key',
    'core.elements.textPropertyField.i18nCard.i18nKeyInputPlaceholder': 'Enter a unique translation key',
    'core.elements.textPropertyField.i18nCard.i18nKeyInputHint': 'Use format: screen.{{key}}',
    'core.elements.textPropertyField.i18nCard.selectI18nKey': 'Select an existing key',
    'core.elements.textPropertyField.i18nCard.language': 'Language',
    'core.elements.textPropertyField.i18nCard.languageText': 'Translation Text',
    'core.elements.textPropertyField.i18nCard.languageTextPlaceholder': 'Enter translation text',
    'core.elements.textPropertyField.i18nCard.commonKeyWarning':
      'This is a common key shared across screens. Changes will affect all usages.',
    'core.elements.textPropertyField.i18nCard.chip.commonScreen.label': 'Common',
    'core.elements.textPropertyField.i18nCard.tooltip.commonKeyTooltip': 'This key is shared across multiple screens',
    'core.elements.textPropertyField.i18nCard.tooltip.editExistingTranslation': 'Edit existing translation',
    'core.elements.textPropertyField.i18nCard.tooltip.addNewTranslation': 'Add new translation',
  },
} as const;

export default translations;
