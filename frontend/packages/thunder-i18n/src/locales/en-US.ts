/**
 * Copyright (c) 2025-2026, WSO2 LLC. (https://www.wso2.com).
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
    // Product
    'product.displayName': 'Thunder',

    // Actions
    'actions.add': 'Add',
    'actions.edit': 'Edit',
    'actions.delete': 'Delete',
    'actions.cancel': 'Cancel',
    'actions.save': 'Save',
    'actions.create': 'Create',
    'actions.update': 'Update',
    'actions.remove': 'Remove',
    'actions.search': 'Search',
    'actions.filter': 'Filter',
    'actions.reset': 'Reset',
    'actions.submit': 'Submit',
    'actions.close': 'Close',
    'actions.back': 'Back',
    'actions.next': 'Next',
    'actions.previous': 'Previous',
    'actions.confirm': 'Confirm',
    'actions.ok': 'OK',
    'actions.yes': 'Yes',
    'actions.no': 'No',
    'actions.continue': 'Continue',
    'actions.skip': 'Skip',
    'actions.finish': 'Finish',
    'actions.done': 'Done',
    'actions.refresh': 'Refresh',
    'actions.copy': 'Copy',
    'actions.copied': 'Copied!',
    'actions.download': 'Download',
    'actions.upload': 'Upload',
    'actions.export': 'Export',
    'actions.import': 'Import',
    'actions.view': 'View',
    'actions.details': 'Details',
    'actions.settings': 'Settings',
    'actions.logout': 'Logout',
    'actions.login': 'Login',

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
    cancel: 'Cancel',
    or: 'or',

    // Page names (for breadcrumbs and navigation)
    home: 'Home',
    flows: 'Flows',

    // Status messages
    'status.loading': 'Loading...',
    'status.saving': 'Saving...',
    'status.deleting': 'Deleting...',
    'status.success': 'Success',
    'status.error': 'Error',
    'status.warning': 'Warning',
    'status.info': 'Info',
    'status.pending': 'Pending',
    'status.active': 'Active',
    'status.inactive': 'Inactive',
    'status.enabled': 'Enabled',
    'status.disabled': 'Disabled',
    'status.completed': 'Completed',
    'status.failed': 'Failed',

    // Form labels
    'edit.general.name.label': 'Name',
    'edit.general.description.label': 'Description',
    'form.email': 'Email',
    'form.password': 'Password',
    'form.username': 'Username',
    'form.required': 'Required',
    'form.optional': 'Optional',
    'form.requiredField': 'This field is required',
    'form.invalidEmail': 'Invalid email address',
    'form.invalidFormat': 'Invalid format',
    'form.searchPlaceholder': 'Search...',

    // Messages
    'messages.confirmDelete': 'Are you sure you want to delete this item?',
    'messages.deleteSuccess': 'Item deleted successfully',
    'messages.deleteError': 'Failed to delete item',
    'messages.saveSuccess': 'Saved successfully',
    'messages.saveError': 'Failed to save',
    'messages.updateSuccess': 'Updated successfully',
    'messages.updateError': 'Failed to update',
    'messages.createSuccess': 'Created successfully',
    'messages.createError': 'Failed to create',
    'messages.noData': 'No data available',
    'messages.noResults': 'No results found',
    'messages.somethingWentWrong': 'Something went wrong',
    'messages.tryAgain': 'Please try again',

    // Navigation
    'navigation.home': 'Home',
    'navigation.dashboard': 'Dashboard',
    'navigation.profile': 'Profile',
    'navigation.help': 'Help',
    'navigation.documentation': 'Documentation',

    // User menu
    'userMenu.profile': 'Profile',
    'userMenu.myAccount': 'My account',
    'userMenu.addAnotherAccount': 'Add another account',
    'userMenu.settings': 'Settings',
    'userMenu.signOut': 'Sign Out',

    // Header
    'header.notifications': 'Coming soon',
    'header.openNotifications': 'Open notifications',

    // Data table - MUI DataGrid locale text
    // Root
    'dataTable.noRowsLabel': 'No rows',
    'dataTable.noResultsOverlayLabel': 'No results found.',
    'dataTable.noColumnsOverlayLabel': 'No columns',
    'dataTable.noColumnsOverlayManageColumns': 'Manage columns',

    // Density selector toolbar button text
    'dataTable.toolbarDensity': 'Density',
    'dataTable.toolbarDensityLabel': 'Density',
    'dataTable.toolbarDensityCompact': 'Compact',
    'dataTable.toolbarDensityStandard': 'Standard',
    'dataTable.toolbarDensityComfortable': 'Comfortable',

    // Columns selector toolbar button text
    'dataTable.toolbarColumns': 'Columns',
    'dataTable.toolbarColumnsLabel': 'Select columns',

    // Filters toolbar button text
    'dataTable.toolbarFilters': 'Filters',
    'dataTable.toolbarFiltersLabel': 'Show filters',
    'dataTable.toolbarFiltersTooltipHide': 'Hide filters',
    'dataTable.toolbarFiltersTooltipShow': 'Show filters',
    'dataTable.toolbarFiltersTooltipActive': (count: number) =>
      count !== 1 ? `${count} active filters` : `${count} active filter`,

    // Quick filter toolbar field
    'dataTable.toolbarQuickFilterPlaceholder': 'Search…',
    'dataTable.toolbarQuickFilterLabel': 'Search',
    'dataTable.toolbarQuickFilterDeleteIconLabel': 'Clear',

    // Export selector toolbar button text
    'dataTable.toolbarExport': 'Export',
    'dataTable.toolbarExportLabel': 'Export',
    'dataTable.toolbarExportCSV': 'Download as CSV',
    'dataTable.toolbarExportPrint': 'Print',

    // Columns management text
    'dataTable.columnsManagementSearchTitle': 'Search',
    'dataTable.columnsManagementNoColumns': 'No columns',
    'dataTable.columnsManagementShowHideAllText': 'Show/Hide All',
    'dataTable.columnsManagementReset': 'Reset',

    // Filter panel text
    'dataTable.filterPanelAddFilter': 'Add filter',
    'dataTable.filterPanelRemoveAll': 'Remove all',
    'dataTable.filterPanelDeleteIconLabel': 'Delete',
    'dataTable.filterPanelLogicOperator': 'Logic operator',
    'dataTable.filterPanelOperator': 'Operator',
    'dataTable.filterPanelOperatorAnd': 'And',
    'dataTable.filterPanelOperatorOr': 'Or',
    'dataTable.filterPanelColumns': 'Columns',
    'dataTable.filterPanelInputLabel': 'Value',
    'dataTable.filterPanelInputPlaceholder': 'Filter value',

    // Filter operators text
    'dataTable.filterOperatorContains': 'contains',
    'dataTable.filterOperatorDoesNotContain': 'does not contain',
    'dataTable.filterOperatorEquals': 'equals',
    'dataTable.filterOperatorDoesNotEqual': 'does not equal',
    'dataTable.filterOperatorStartsWith': 'starts with',
    'dataTable.filterOperatorEndsWith': 'ends with',
    'dataTable.filterOperatorIs': 'is',
    'dataTable.filterOperatorNot': 'is not',
    'dataTable.filterOperatorAfter': 'is after',
    'dataTable.filterOperatorOnOrAfter': 'is on or after',
    'dataTable.filterOperatorBefore': 'is before',
    'dataTable.filterOperatorOnOrBefore': 'is on or before',
    'dataTable.filterOperatorIsEmpty': 'is empty',
    'dataTable.filterOperatorIsNotEmpty': 'is not empty',
    'dataTable.filterOperatorIsAnyOf': 'is any of',

    // Filter values text
    'dataTable.filterValueAny': 'any',
    'dataTable.filterValueTrue': 'true',
    'dataTable.filterValueFalse': 'false',

    // Column menu text
    'dataTable.columnMenuLabel': 'Menu',
    'dataTable.columnMenuShowColumns': 'Show columns',
    'dataTable.columnMenuManageColumns': 'Manage columns',
    'dataTable.columnMenuFilter': 'Filter',
    'dataTable.columnMenuHideColumn': 'Hide column',
    'dataTable.columnMenuUnsort': 'Unsort',
    'dataTable.columnMenuSortAsc': 'Sort by ASC',
    'dataTable.columnMenuSortDesc': 'Sort by DESC',

    // Column header text
    'dataTable.columnHeaderFiltersTooltipActive': (count: number) =>
      count !== 1 ? `${count} active filters` : `${count} active filter`,
    'dataTable.columnHeaderFiltersLabel': 'Show filters',
    'dataTable.columnHeaderSortIconLabel': 'Sort',

    // Rows selected footer text
    'dataTable.footerRowSelected': (count: number) =>
      count !== 1 ? `${count.toLocaleString()} rows selected` : `${count.toLocaleString()} row selected`,

    // Total row amount footer text
    'dataTable.footerTotalRows': 'Total Rows:',

    // Total visible row amount footer text
    'dataTable.footerTotalVisibleRows': (visibleCount: number, totalCount: number) =>
      `${visibleCount.toLocaleString()} of ${totalCount.toLocaleString()}`,

    // Checkbox selection text
    'dataTable.checkboxSelectionHeaderName': 'Checkbox selection',
    'dataTable.checkboxSelectionSelectAllRows': 'Select all rows',
    'dataTable.checkboxSelectionUnselectAllRows': 'Unselect all rows',
    'dataTable.checkboxSelectionSelectRow': 'Select row',
    'dataTable.checkboxSelectionUnselectRow': 'Unselect row',

    // Boolean cell text
    'dataTable.booleanCellTrueLabel': 'yes',
    'dataTable.booleanCellFalseLabel': 'no',

    // Actions cell more text
    'dataTable.actionsCellMore': 'more',

    // Column pinning text
    'dataTable.pinToLeft': 'Pin to left',
    'dataTable.pinToRight': 'Pin to right',
    'dataTable.unpin': 'Unpin',

    // Tree Data
    'dataTable.treeDataGroupingHeaderName': 'Group',
    'dataTable.treeDataExpand': 'see children',
    'dataTable.treeDataCollapse': 'hide children',

    // Grouping columns
    'dataTable.groupingColumnHeaderName': 'Group',
    'dataTable.groupColumn': (name: string) => `Group by ${name}`,
    'dataTable.unGroupColumn': (name: string) => `Stop grouping by ${name}`,

    // Master/detail
    'dataTable.detailPanelToggle': 'Detail panel toggle',
    'dataTable.expandDetailPanel': 'Expand',
    'dataTable.collapseDetailPanel': 'Collapse',

    // Pagination
    'dataTable.paginationRowsPerPage': 'Rows per page:',
    'dataTable.paginationDisplayedRows': ({from, to, count}: {from: number; to: number; count: number}) =>
      `${from}–${to} of ${count !== -1 ? count : `more than ${to}`}`,

    // Row reordering text
    'dataTable.rowReorderingHeaderName': 'Row reordering',

    // Aggregation
    'dataTable.aggregationMenuItemHeader': 'Aggregation',
    'dataTable.aggregationFunctionLabelSum': 'sum',
    'dataTable.aggregationFunctionLabelAvg': 'avg',
    'dataTable.aggregationFunctionLabelMin': 'min',
    'dataTable.aggregationFunctionLabelMax': 'max',
    'dataTable.aggregationFunctionLabelSize': 'size',
  },

  // ============================================================================
  // Navigation namespace - Navigation related translations
  // ============================================================================
  navigation: {
    'pages.home': 'Home',
    'pages.users': 'Users',
    'pages.userTypes': 'User Types',
    'pages.organizationUnits': 'Organization Units',
    'pages.integrations': 'Integrations',
    'pages.applications': 'Applications',
    'pages.dashboard': 'Dashboard',
    'pages.flows': 'Flows',
    'breadcrumb.develop': 'Develop',
  },

  // ============================================================================
  // Users namespace - User management feature translations
  // ============================================================================
  users: {
    title: 'User Management',
    subtitle: 'Manage users, roles, and permissions across your organization',
    addUser: 'Add User',
    inviteUser: 'Invite User',
    inviteUserDescription: 'Send an invite link to a new user to complete their registration',
    inviteLinkGenerated: 'Invite Link Generated!',
    inviteLinkDescription: 'Share this link with the user to complete their registration.',
    inviteLink: 'Invite Link',
    inviteAnother: 'Invite Another User',
    'invite.steps.userdetails': 'User Details',
    'invite.steps.invitelink': 'Invite Link',
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
    'errors.failed.title': 'Error',
    'errors.failed.description': 'An error occurred. Please try again.',
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
    'types.string': 'String',
    'types.number': 'Number',
    'types.boolean': 'Boolean',
    'types.enum': 'Enum',
    'types.array': 'Array',
    'types.object': 'Object',
    'validationErrors.nameRequired': 'Please enter a user type name',
    'validationErrors.ouIdRequired': 'Please provide an organization unit ID',
    'validationErrors.propertiesRequired': 'Please add at least one property',
    'validationErrors.duplicateProperties': 'Duplicate property names found: {{duplicates}}',
    'errors.organizationUnitsFailedTitle': 'Failed to load organization units',
    noUserTypes: 'No user types found',
    noOrganizationUnits: 'No organization units available',
    confirmDeleteUserType: 'Are you sure you want to delete this user type?',
  },

  // ============================================================================
  // Organization Units namespace - Organization unit management feature translations
  // ============================================================================
  organizationUnits: {
    // Listing page
    'listing.title': 'Organization Units (OU)',
    'listing.subtitle': 'Manage organization units and hierarchies',
    'listing.addRootOrganizationUnit': 'Add Root Organization Unit',
    'listing.error.title': 'Failed to load organization units',
    'listing.error.unknown': 'An unknown error occurred',
    'listing.treeView.empty': 'No organization units found',
    'listing.treeView.noChildren': 'No child organization units',
    'listing.treeView.loadError': 'Failed to load child organization units',
    'listing.treeView.addChild': 'Add Child OU',
    'listing.treeView.addChildOrganizationUnit': 'Add Child Organization Unit',
    'listing.columns.name': 'Name',
    'listing.columns.handle': 'Handle',
    'listing.columns.description': 'Description',

    // Create page
    'create.title': 'Create Organization Unit',
    'create.heading': "Let's set up your organization unit",
    'create.subtitle': 'Define a new organization unit',
    'create.error': 'Failed to create organization unit. Please try again.',
    'create.suggestions.label': 'In a hurry? Pick a random name:',

    'delete.dialog.title': 'Delete Organization Unit',
    'delete.dialog.message': 'Are you sure you want to delete this organization unit? This action cannot be undone.',
    'delete.dialog.disclaimer': 'Warning: All associated data, configurations, and user assignments will be permanently removed.',
    'delete.dialog.error': 'Failed to delete organization unit. Please try again.',

    /* -------------------- Edit page -------------------- */
    // Common
    'edit.page.error': 'Failed to load organization unit',
    'edit.page.notFound': 'Organization unit not found',
    'edit.page.back': 'Back to Organization Units',
    'edit.page.backToOU': 'Back to {{name}}',
    'edit.page.description.empty': 'No description',
    'edit.page.description.placeholder': 'Enter a description...',
    'edit.page.tabs.general': 'General',
    'edit.page.tabs.childOUs': 'Child OUs',
    'edit.page.tabs.users': 'Users',
    'edit.page.tabs.groups': 'Groups',
    'edit.page.tabs.customization': 'Customization',
    'edit.page.tabs.advanced': 'Advanced Settings',
    'edit.actions.unsavedChanges.label': 'You have unsaved changes',
    'edit.actions.reset.label': 'Reset',
    'edit.actions.save.label': 'Save Changes',
    'edit.actions.saving.label': 'Saving...',

    // General section
    'edit.general.sections.quickCopy.title': 'Quick Copy',
    'edit.general.sections.quickCopy.description': 'Copy organization unit identifiers for quick reference.',
    'edit.general.sections.parentOUSettings.title': 'Parent Organization Unit',
    'edit.general.sections.parentOUSettings.description': 'The parent organization unit in the hierarchy.',
    'edit.general.sections.dangerZone.title': 'Danger Zone',
    'edit.general.sections.dangerZone.description': 'Actions in this section are irreversible. Proceed with caution.',
    'edit.general.sections.dangerZone.deleteOU.title': 'Delete Organization Unit',
    'edit.general.sections.dangerZone.deleteOU.description':
      'Deleting this organization unit is permanent and cannot be undone.',
    'edit.general.ou.id.label': 'Organization Unit ID',
    'edit.general.ou.parent.label': 'Parent Organization Unit',
    'edit.general.ou.noParent.label': 'Root Organization Unit',
    'edit.general.dangerZone.delete.button.label': 'Delete Organization Unit',
    // Form fields
    'edit.general.handle.label': 'Handle',
    'edit.general.handle.placeholder': 'e.g., engineering, sales, hr',
    'edit.general.handle.hint': 'A unique identifier for this organization unit',
    'edit.general.handle.validations.required': 'Handle is required',
    'edit.general.handle.validations.format': 'Handle must be lowercase alphanumeric with hyphens only',
    'edit.general.name.label': 'Name',
    'edit.general.name.placeholder': 'e.g., Engineering Department',
    'edit.general.name.validations.required': 'Name is required',
    'edit.general.description.label': 'Description',
    'edit.general.description.placeholder': 'Enter a description for this organization unit',
    'edit.general.parent.label': 'Parent Organization Unit',
    'edit.general.parent.hint': 'The parent organization unit for this new unit',
    'edit.general.dangerZone.delete.title': 'Delete Organization Unit',
    'edit.general.dangerZone.delete.message':
      'Are you sure you want to delete this organization unit? This action cannot be undone.',
    'edit.general.dangerZone.delete.error': 'Failed to delete organization unit. Please try again.',
    'edit.general.dangerZone.delete.success': 'Organization unit deleted successfully.',

    // Child OUs Section
    'edit.childOUs.sections.manage.title': 'Child Organization Units',
    'edit.childOUs.sections.manage.description': 'View and manage child organization units under this OU',

    // Users Section
    'edit.users.sections.manage.title': 'Users',
    'edit.users.sections.manage.description': 'View users belonging to this organization unit',
    'edit.users.sections.manage.listing.columns.id': 'User ID',
    'edit.users.sections.manage.listing.columns.type': 'User Type',
    'edit.users.sections.manage.listing.columns.name': 'Group Name',

    // Groups Section
    'edit.groups.sections.manage.title': 'Groups',
    'edit.groups.sections.manage.description': 'View groups belonging to this organization unit',
    'edit.groups.sections.manage.listing.columns.id': 'Group ID',

    // Customization tab
    'edit.customization.sections.appearance': 'Appearance',
    'edit.customization.sections.appearance.description': 'Customize the look and feel of this organization unit.',
    'edit.customization.labels.theme': 'Theme',
    'edit.customization.theme.placeholder': 'Select a theme',
    'edit.customization.theme.hint': 'The theme applied to this organization unit.',
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
    'listing.columns.template': 'Type',
    'listing.search.placeholder': 'Search ..',
    'delete.title': 'Delete Application',
    'delete.message': 'Are you sure you want to delete this application? This action cannot be undone.',
    'delete.disclaimer': 'Warning: All associated data, configurations, and access tokens will be permanently removed.',
    'onboarding.preview.title': 'Preview',
    'onboarding.preview.signin': 'Sign In',
    'onboarding.preview.username': 'Username',
    'onboarding.preview.usernamePlaceholder': 'Enter your Username',
    'onboarding.preview.password': 'Password',
    'onboarding.preview.passwordPlaceholder': 'Enter your Password',
    'onboarding.preview.signInButton': 'Sign In',
    'onboarding.preview.passkeySignIn': 'Sign in with Passkey',
    'onboarding.preview.mobileNumber': 'Mobile Number',
    'onboarding.preview.mobileNumberPlaceholder': 'Enter your mobile number',
    'onboarding.preview.sendOtpButton': 'Send OTP',
    'onboarding.preview.dividerText': 'or',
    'onboarding.preview.continueWith': 'Continue with {{providerName}}',
    'onboarding.steps.name': 'Create an Application',
    'onboarding.steps.design': 'Design',
    'onboarding.steps.options': 'Sign In Options',
    'onboarding.steps.experience': 'Sign-In Experience',
    'onboarding.steps.stack': 'Technology Stack',
    'onboarding.steps.configure': 'Configuration',
    'onboarding.steps.complete': 'Setup Complete',
    'onboarding.steps.summary': 'Summary',
    'onboarding.configure.name.title': "Let's give a name to your application",
    'onboarding.configure.name.fieldLabel': 'Application Name',
    'onboarding.configure.name.placeholder': 'Enter your application name',
    'onboarding.configure.name.suggestions.label': 'In a hurry? Pick a random name:',
    'onboarding.configure.design.title': 'Design Your Application',
    'onboarding.configure.design.subtitle': 'Customize the appearance of your application',
    'onboarding.configure.design.logo.title': 'Application Logo',
    'onboarding.configure.design.logo.shuffle': 'Shuffle',
    'onboarding.configure.design.theme.title': 'Theme',
    'onboarding.configure.design.theme.emptyState': 'No themes configured',
    'onboarding.configure.design.theme.noDescription': 'No description',
    'onboarding.configure.design.theme.emptyStateHint': 'You can configure themes later from the Design settings.',
    'onboarding.configure.SignInOptions.title': 'Sign In Options',
    'onboarding.configure.SignInOptions.subtitle': 'Choose how users will sign-in to your application',
    'onboarding.configure.SignInOptions.usernamePassword': 'Username & Password',
    'onboarding.configure.SignInOptions.google': 'Google',
    'onboarding.configure.SignInOptions.github': 'GitHub',
    'onboarding.configure.SignInOptions.passkey': 'Passkey',
    'onboarding.configure.SignInOptions.notConfigured': 'Not configured',
    'onboarding.configure.SignInOptions.noFlowFound':
      'No flow found for the selected sign-in options. Please try a different combination.',
    'onboarding.configure.SignInOptions.noSelectionWarning':
      'At least one sign-in option is required. Please select at least one authentication method.',
    'onboarding.configure.SignInOptions.noIntegrations':
      'No social sign-in integrations available. Please configure an integration first.',
    'onboarding.configure.SignInOptions.hint':
      'You can always change these settings later in the application settings.',
    'onboarding.configure.SignInOptions.preConfiguredFlows.selectFlow': 'Select already configured flow',
    'onboarding.configure.SignInOptions.preConfiguredFlows.searchFlows': 'Search flows...',
    'onboarding.configure.SignInOptions.smsOtp': 'SMS OTP',
    'onboarding.configure.SignInOptions.loading': 'Loading...',
    'onboarding.configure.SignInOptions.error': 'Failed to load authentication methods: {{error}}',
    'onboarding.configure.experience.title': 'Sign-In Experience',
    'onboarding.configure.experience.subtitle': 'Select how and who can access your application',
    'onboarding.configure.experience.subtitleWithoutUserTypes': 'Select how users access your application',
    'onboarding.configure.experience.access.userTypes.title': 'User Access',
    'onboarding.configure.experience.access.userTypes.subtitle': 'Select which user types can access this application',
    'onboarding.configure.experience.approach.title': 'Sign-In Approach',
    'onboarding.configure.experience.approach.subtitle': 'Select how users will access this application',
    'onboarding.configure.approach.inbuilt.title': 'Redirect to {{product}} sign-in/sign-up handling pages',
    'onboarding.configure.approach.inbuilt.description':
      'Users will be redirected to system-hosted sign-in and sign-up pages, which can be customized and branded using the Flow Designer and easily integrated with SDKs in just a few steps.',
    'onboarding.configure.approach.native.title': 'Embedded sign-in/sign-up components in your app',
    'onboarding.configure.approach.native.description':
      'Users will sign in or sign up through your app using the UI components or APIs provided by {{product}}. You can customize and brand the flows using the designer or through code.',
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
    'onboarding.configure.details.passkey.title': 'Passkey Settings',

    'onboarding.configure.passkey.title': 'Passkey Configuration',
    'onboarding.configure.passkey.description': 'Configure the Relying Party details for Passkey authentication.',
    'onboarding.configure.details.relyingPartyId.label': 'Relying Party ID',
    'onboarding.configure.details.relyingPartyId.placeholder': 'e.g., example.com',
    'onboarding.configure.details.relyingPartyId.helperText': 'The domain where the WebAuthn credential is valid',
    'onboarding.configure.details.relyingPartyName.label': 'Relying Party Name',
    'onboarding.configure.details.relyingPartyName.placeholder': 'e.g., My Application',
    'onboarding.configure.details.relyingPartyName.helperText': 'A user-friendly name for the Relying Party',
    'onboarding.configure.details.noConfigRequired.title': 'No Additional Configuration Needed',
    'onboarding.configure.details.noConfigRequired.description':
      'Your application is ready to go! You can proceed to the next step.',
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
    'onboarding.complete.title': 'Application Created Successfully',
    'onboarding.complete.description':
      'Your application has been created. Please save the client secret below as it will only be shown once.',
    'onboarding.complete.warning.title': 'Important: Save Your Client Secret',
    'onboarding.complete.warning.message':
      'This is the only time you will see this client secret. Please copy it now and store it securely. You will not be able to retrieve it later.',
    'onboarding.complete.clientSecret.label': 'Client Secret',
    'onboarding.creating': 'Creating...',
    'onboarding.skipAndCreate': 'Skip & Create',
    'onboarding.createApplication': 'Create Application',
    'onboarding.summary.title': 'Application Created Successfully!',
    'onboarding.summary.subtitle': 'Your application is ready to use',
    'onboarding.summary.viewAppAriaLabel': 'Click to view application details',
    'onboarding.summary.appDetails': 'Application successfully created',
    'onboarding.summary.guides.subtitle': 'Choose how you want to integrate sign-in to your application',
    'onboarding.summary.guides.divider': 'or',
    'clientSecret.saveTitle': 'Save Your Client Secret',
    'clientSecret.saveSubtitle': "This is the only time you'll see this secret. Store it somewhere safe.",
    'clientSecret.appNameLabel': 'App Name',
    'clientSecret.warning':
      "Make sure to copy your client secret now. You won't be able to see it again for security reasons.",
    'clientSecret.clientIdLabel': 'Client ID',
    'clientSecret.clientSecretLabel': 'Client Secret',
    'clientSecret.copied': 'Copied to clipboard',
    'clientSecret.copySecret': 'Copy Secret',
    'clientSecret.securityReminder.title': 'Security Reminder',
    'clientSecret.securityReminder.description':
      'Your client secret is a confidential key used to authenticate your application. It should be treated with the same level of security as a password. Never expose it in browser console, version control, or logs.',
    'view.title': 'Application Details',
    'view.subtitle': 'View application details and configuration',
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
    'view.fields.authFlowId': 'Authentication Flow ID',
    'view.fields.registrationFlowId': 'Registration Flow ID',
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
    'edit.customization.tosUri.hint':
      "URL to your application's Terms of Service. May be displayed to users during consent or user sign-in, sign-up or recovery flows.",
    'edit.customization.policyUri.hint':
      "URL to your application's Privacy Policy. May be displayed to users during consent or user sign-in, sign-up or recovery flows.",
    'edit.advanced.oauth2Config.intro':
      'These OAuth 2.0 settings are read-only and were configured during application creation.',
    'edit.advanced.redirectUris.hint':
      'Allowed callback URLs where users will be redirected after authentication. Must be exact matches for security.',
    'edit.advanced.grantTypes.hint':
      'OAuth 2.0 flows this application can use (e.g., authorization_code, client_credentials, refresh_token).',
    'edit.advanced.publicClient.public':
      'This is a public client (SPA, mobile app) that cannot securely store credentials.',
    'edit.advanced.publicClient.confidential':
      'This is a confidential client (server-side app) that can securely store credentials.',
    'edit.advanced.pkce.enabled': 'PKCE is required for authorization code flow, providing additional security.',
    'edit.advanced.pkce.disabled': 'PKCE is not required. Consider enabling for public clients (SPAs, mobile apps).',
    'edit.advanced.certificate.intro': 'Configure certificates for client authentication and token encryption.',
    'edit.advanced.certificate.type.none': 'None',
    'edit.advanced.certificate.type.jwks': 'JWKS (Inline JSON Web Key Set)',
    'edit.advanced.certificate.type.jwksUri': 'JWKS URI (URL to JWKS endpoint)',
    // Logo update modal
    'logoModal.title': 'Update Application Logo',
    'logoModal.preview.title': 'Preview',
    'logoModal.customUrl.title': 'Custom Logo URL',
    'logoModal.customUrl.placeholder': 'https://example.com/logo.png',
    'logoModal.customUrl.hint': 'Enter a direct URL to your custom logo image',
    'logoModal.logos.title': 'Application Logo',
    'logoModal.logos.shuffle': 'Shuffle',
    'logoModal.cancel': 'Cancel',
    'logoModal.update': 'Update Logo',

    /* -------------------- Edit page -------------------- */
    // Common
    'edit.page.error': 'Failed to load application information',
    'edit.page.notFound': 'Application not found',
    'edit.page.back': 'Back to Applications',
    'edit.page.description.empty': 'No description',
    'edit.page.description.placeholder': 'Add a description',
    'edit.page.tabs.overview': 'Guide',
    'edit.page.tabs.general': 'General',
    'edit.page.tabs.flows': 'Flows',
    'edit.page.tabs.customization': 'Customization',
    'edit.page.tabs.token': 'Token',
    'edit.page.tabs.advanced': 'Advanced Settings',
    'edit.page.unsavedChanges': 'Unsaved changes',
    'edit.page.reset': 'Reset',
    'edit.page.save': 'Save',
    'edit.page.saving': 'Saving...',

    // Overview section
    'edit.overview.noGuides': 'No integration guides available for this application type.',

    // General section
    'edit.general.sections.quickCopy': 'Quick Copy',
    'edit.general.sections.quickCopy.description': 'Copy application identifiers for use in your code.',
    'edit.general.sections.access': 'Access',
    'edit.general.sections.access.description': "Configure who can access this application, where it's hosted, etc.",
    'edit.general.sections.contacts': 'Contacts',
    'edit.general.sections.contacts.description': 'Contact email addresses for application administrators.',
    'edit.general.labels.applicationId': 'Application ID',
    'edit.general.labels.clientId': 'Client ID',
    'edit.general.labels.allowedUserTypes': 'Allowed User Types',
    'edit.general.labels.applicationUrl': 'Application URL',
    'edit.general.labels.contacts': 'Contacts',
    'edit.general.applicationId.hint': 'Unique identifier for your application',
    'edit.general.clientId.hint': 'OAuth2 client identifier used for authentication',
    'edit.general.noUserTypes': 'No user types configured',
    'edit.general.contacts.placeholder': 'admin@example.com, support@example.com',
    'edit.general.contacts.hint': 'Enter email addresses separated by commas',
    'edit.general.redirectUris.title': 'Authorized redirect URIs',
    'edit.general.redirectUris.description': 'For use with requests from a web server',
    'edit.general.redirectUris.tooltip': 'OAuth2 redirect URIs where users will be redirected after authentication',
    'edit.general.redirectUris.addUri': 'Add URI',
    'edit.general.redirectUris.error.empty': 'Invalid Redirect: URI must not be empty.',
    'edit.general.redirectUris.error.invalid': 'Invalid Redirect: Please enter a valid URL.',
    'edit.general.allowedUserTypes.placeholder': 'Select user types',
    'edit.general.allowedUserTypes.hint': 'Users of these types can authenticate with this application',
    'edit.general.applicationUrl.hint': 'The homepage URL of your application',

    // Flows section
    'edit.flows.labels.authFlow': 'Authentication Flow',
    'edit.flows.labels.authFlow.description': 'Choose the flow that handles user login and authentication.',
    'edit.flows.labels.registrationFlow': 'Registration Flow',
    'edit.flows.labels.registrationFlow.description': 'Choose the flow that handles user sign-up and account creation.',
    'edit.flows.labels.enableRegistration': 'Enable Registration Flow',
    'edit.flows.authFlow.placeholder': 'Select an authentication flow',
    'edit.flows.authFlow.hint': 'Select the flow that handles user sign-in for this application.',
    'edit.flows.authFlow.alert':
      'To modify the selected flow, <0>open the flow builder</0>. To create a new flow, visit the <1>Flows page</1>.',
    'edit.flows.registrationFlow.placeholder': 'Select a registration flow',
    'edit.flows.registrationFlow.hint': 'Select the flow that handles user registration for this application.',
    'edit.flows.registrationFlow.alert':
      'To modify the selected flow, <0>open the flow builder</0>. To create a new flow, visit the <1>Flows page</1>.',
    'edit.flows.enableRegistration.hint': 'Allow users to register new accounts through this application',
    'edit.flows.editFlow': 'Edit flow',

    // Customization section
    'edit.customization.sections.appearance': 'Appearance',
    'edit.customization.sections.appearance.description': 'Customize the visual appearance of your application.',
    'edit.customization.sections.urls': 'URLs',
    'edit.customization.sections.urls.description': 'Configure legal and policy URLs for your application.',
    'edit.customization.labels.theme': 'Theme',
    'edit.customization.labels.tosUri': 'Terms of Service URI',
    'edit.customization.labels.policyUri': 'Privacy Policy URI',
    'edit.customization.theme.placeholder': 'Select a Theme',
    'edit.customization.theme.hint':
      'Choose a theme to customize authentication pages. Select the Default Theme (shared across all applications) or pick an app-specific theme.',
    'edit.customization.tosUri.placeholder': 'https://example.com/terms',
    'edit.customization.policyUri.placeholder': 'https://example.com/privacy',

    // Token section
    'edit.token.labels.title': 'Token Settings',
    'edit.token.intro': 'Customize your access and ID tokens by adding user attributes.',
    'edit.token.labels.accessToken': 'Access Token Configuration',
    'edit.token.labels.idToken': 'ID Token Configuration',
    'edit.token.labels.accessTokenValidity': 'Access Token Validity (seconds)',
    'edit.token.labels.accessTokenAttributes': 'Access Token User Attributes',
    'edit.token.labels.idTokenValidity': 'ID Token Validity (seconds)',
    'edit.token.labels.idTokenAttributes': 'ID Token User Attributes',
    'edit.token.labels.scopes': 'Scopes',
    'edit.token.labels.customAttributes': 'Custom Attributes',
    'edit.token.labels.tokenValidity': 'Token Validity',
    'edit.token.notConfigured': 'Not configured',
    'edit.token.noScopes': 'No scopes configured',
    'edit.token.noCustomAttributes': 'No custom attributes added',
    'edit.token.seconds': 'seconds',
    'edit.token.accessTokenValidity.hint': 'Duration in seconds before the access token expires',
    'edit.token.idTokenValidity.hint': 'Duration in seconds before the ID token expires',
    'edit.token.customAttributes.hint': 'Add custom attributes to your tokens by selecting user attributes below.',
    'edit.token.scopes.hint': 'OAuth2 scopes allowed for this application',
    'edit.token.userAttributes': 'User Attributes',
    'edit.token.defaultAttributes': 'Default Attributes',
    'edit.token.defaultAttributes.info':
      'These attributes are automatically included in all tokens for compatibility and standard compliance.',
    'edit.token.loadingAttributes': 'Loading user attributes...',
    'edit.token.noUserAttributes': 'No user attributes available. Configure allowed user types for this application.',
    'edit.token.alreadyAdded': 'Already added',
    'edit.token.clickToAdd': 'Click to add',
    'edit.token.clickToRemove': 'Click to remove',
    'edit.token.configureAttributes': 'Add or Remove Attributes',
    'edit.token.configureAttributes.hint':
      'Click on user attributes to add them to your token. You can also view the default attributes that are automatically included.',
    'edit.token.accessTokenPreview': 'Preview of access token attributes',
    'edit.token.token.preview.title': 'Token Preview (JWT)',
    'edit.token.accessToken.preview.title': 'Access Token Preview (JWT)',
    'edit.token.idToken.preview.title': 'ID Token Preview (JWT)',
    'edit.token.useSeparateConfigs': 'Use separate configurations for Access Token and ID Token',
    'edit.token.useSeparateTokenConfigs': 'Use separate configurations for Access Token and ID Token',
    'edit.token.separateConfigHint':
      'When enabled, you can configure different user attributes and validity periods for access tokens and ID tokens separately.',
    'edit.token.sharedConfigs.hint': 'User attributes will be shared between Access Token and ID Token',
    'edit.token.sharedValidity.hint': 'Token validity period in seconds',
    'edit.token.tokenValidation.title': 'Token Validation',
    'edit.token.tokenValidation.description': 'Configure token validation settings such as audience and validity',
    'edit.token.accessTokenValidation.title': 'Access Token Validation',
    'edit.token.accessTokenValidation.description': 'Configure how long access tokens remain valid before expiration',
    'edit.token.idTokenValidation.title': 'ID Token Validation',
    'edit.token.idTokenValidation.description': 'Configure how long ID tokens remain valid before expiration',
    'edit.token.userAttributes.title': 'User Attributes',
    'edit.token.userAttributes.description':
      'Select which user attributes to include in your tokens. These attributes will be available in the issued tokens.',
    'edit.token.accessTokenUserAttributes.title': 'Access Token User Attributes',
    'edit.token.accessTokenUserAttributes.description':
      'Configure user attributes that will be included in the access token. You can add custom attributes from user profiles.',
    'edit.token.idTokenUserAttributes.title': 'ID Token User Attributes',
    'edit.token.idTokenUserAttributes.description':
      'Configure user attributes that will be included in the ID token. You can add custom attributes from user profiles and define scope-based attributes.',
    'edit.token.tokenConfiguration.title': 'Token Configuration',
    'edit.token.tokenConfiguration.description':
      'Configure token validity settings that apply to all issued tokens.',
    'edit.token.validity.title': 'Token Validity',
    'edit.token.validity.description':
      'Configure how long tokens remain valid before expiration. Shorter validity periods enhance security.',
    'edit.token.validity.hint': 'Token validity period in seconds (e.g., 3600 for 1 hour)',
    'edit.token.validity.error': 'Validity period must be at least 1 second',
    'edit.token.scopes.title': 'OAuth Scopes',
    'edit.token.scopes.description':
      'Scopes define the level of access that the application is requesting. Users will see these during consent.',
    'edit.token.other': 'Other',
    'edit.token.noOtherAttributes': 'No other attributes available',

    // Advanced section
    'edit.advanced.labels.oauth2Config': 'OAuth2 Configuration',
    'edit.advanced.labels.redirectUris': 'Redirect URIs',
    'edit.advanced.labels.grantTypes': 'Grant Types',
    'edit.advanced.labels.responseTypes': 'Response Types',
    'edit.advanced.labels.publicClient': 'Public Client',
    'edit.advanced.labels.pkceRequired': 'PKCE Required',
    'edit.advanced.labels.certificate': 'Certificate',
    'edit.advanced.labels.certificateType': 'Certificate Type',
    'edit.advanced.labels.metadata': 'Metadata',
    'edit.advanced.labels.createdAt': 'Created At',
    'edit.advanced.labels.updatedAt': 'Updated At',
    'edit.advanced.publicClient.yes': 'Yes',
    'edit.advanced.publicClient.no': 'No',
    'edit.advanced.pkce.yes': 'Yes',
    'edit.advanced.pkce.no': 'No',
    'edit.advanced.certificate.placeholder.jwksUri': 'https://example.com/.well-known/jwks',
    'edit.advanced.certificate.placeholder.jwks': 'Enter JWKS JSON',
    'edit.advanced.certificate.hint.jwksUri': 'URL to the JWKS endpoint',
    'edit.advanced.certificate.hint.jwks': 'JSON Web Key Set',
  },

  // ============================================================================
  // Sign In - Sign In page translations
  // ============================================================================
  signin: {
    'errors.signin.failed.message': 'Error',
    'errors.signin.failed.description': 'We are sorry, something has gone wrong here. Please try again.',
    'errors.passkey.failed': 'Passkey authentication failed. Please try again.',
    'redirect.to.signup': "Don't have an account? <1>Sign up</1>",
    heading: 'Sign In',
    // Passkey authentication
    'passkey.button.use': 'Sign in with Passkey',
    'passkey.signin.heading': 'Sign in with Passkey',
    'passkey.signin.description': 'Use your passkey to securely sign in to your account without a password.',
    'passkey.register.heading': 'Register Passkey',
    'passkey.register.description': 'Create a passkey to securely sign in to your account without a password.',
  },

  // ============================================================================
  // Sign Up - Sign Up page translations
  // ============================================================================
  signup: {
    'errors.signup.failed.message': 'Error',
    'errors.signup.failed.description': 'We are sorry, but we were unable to create your account. Please try again.',
    'redirect.to.signin': 'Already have an account? <1>Sign in</1>',
    heading: 'Sign Up',
    // Passkey registration translations
    'passkey.setup.heading': 'Set Up Passkey',
    'passkey.setup.description': 'Create a passkey to securely sign in to your account without a password.',
    'passkey.button.create': 'Create Passkey',
    'passkey.registering': 'Creating passkey...',
    'errors.passkey.failed': 'Failed to create passkey. Please try again.',
  },

  // ============================================================================
  // Invite namespace - Invite acceptance feature translations (for Thunder Gate)
  // ============================================================================
  invite: {
    validating: 'Validating your invite link...',
    'complete.title': 'Welcome Aboard!',
    'complete.description': 'Your account has been successfully set up.',
    'errors.invalid.title': 'Unable to verify invite',
    'errors.invalid.description': 'This invite link is invalid or has expired.',
    'errors.failed.title': 'Error',
    'errors.failed.description': 'An error occurred. Please try again.',
    goToSignIn: 'Go to Sign In',
    signIn: 'Sign In',
  },

  // ============================================================================
  // Components namespace - SDK component error translations
  // ============================================================================
  components: {
    'signIn.errors.generic': 'An error occurred during sign in. Please try again.',
    'signUp.errors.generic': 'An error occurred during sign up. Please try again.',
    'inviteUser.errors.generic': 'An error occurred while inviting the user. Please try again.',
    'acceptInvite.errors.generic': 'An error occurred while accepting the invite. Please try again.',
  },

  // ============================================================================
  // Elements - Low level reusable element translations
  // ============================================================================
  elements: {
    'buttons.github.text': 'Continue with GitHub',
    'buttons.google.text': 'Continue with Google',
    'buttons.submit.text': 'Continue',
    'display.divider.or_separator': 'OR',
    'fields.first_name.label': 'First Name',
    'fields.first_name.placeholder': 'Enter your first name',
    'fields.last_name.label': 'Last Name',
    'fields.last_name.placeholder': 'Enter your last name',
    'fields.email.label': 'Email',
    'fields.email.placeholder': 'Enter your email address',
    'fields.mobile.label': 'Mobile Number',
    'fields.mobile.placeholder': 'Enter your mobile number',
    'fields.password.label': 'Password',
    'fields.password.placeholder': 'Enter your password',
    'fields.username.label': 'Username',
    'fields.username.placeholder': 'Enter your username',
    'fields.usertype.label': 'User Type',
    'fields.usertype.placeholder': 'Select User Type',
  },

  // ============================================================================
  // Validations - Form & other validation messages translations
  // ============================================================================
  validations: {
    'form.field.required': '{{field}} is required.',
  },

  // ============================================================================
  // Flows - Flow builder feature translations
  // ============================================================================
  flows: {
    // Flow listing page
    'listing.title': 'Flows',
    'listing.subtitle': 'Create and manage authentication and registration flows for your applications',
    'listing.addFlow': 'Create New Flow',
    'listing.columns.name': 'Name',
    'listing.columns.flowType': 'Type',
    'listing.columns.version': 'Version',
    'listing.columns.updatedAt': 'Last Updated',
    'listing.columns.actions': 'Actions',
    'listing.error.title': 'Failed to load flows',
    'listing.error.unknown': 'An unknown error occurred',
    'delete.title': 'Delete Flow',
    'delete.message': 'Are you sure you want to delete this flow? This action cannot be undone.',
    'delete.disclaimer': 'Warning: All associated configurations will be permanently removed.',
    'delete.error': 'Failed to delete flow. Please try again.',

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
    'core.executions.names.PasskeyAuthentication': 'Passkey Authentication',
    'core.executions.names.confirmationCode': 'Confirmation Code',
    'core.executions.names.magicLink': 'Magic Link',
    'core.executions.names.sendEmailOTP': 'Send Email OTP',
    'core.executions.names.verifyEmailOTP': 'Verify Email OTP',
    'core.executions.names.sendSMS': 'Send SMS',
    'core.executions.names.verifySMSOTP': 'Verify SMS OTP',
    'core.executions.names.default': 'Execution',

    // SMS OTP executor modes
    'core.executions.smsOtp.mode.send': 'Send OTP',
    'core.executions.smsOtp.mode.verify': 'Verify OTP',
    'core.executions.smsOtp.mode.label': 'Mode',
    'core.executions.smsOtp.mode.placeholder': 'Select a mode',
    'core.executions.smsOtp.description': 'Configure the SMS OTP executor settings.',

    // SMS OTP sender selection
    'core.executions.smsOtp.sender.label': 'Notification Sender',
    'core.executions.smsOtp.sender.placeholder': 'Select a notification sender',
    'core.executions.smsOtp.sender.required': 'Notification sender is required and must be selected.',
    'core.executions.smsOtp.sender.noSenders':
      'No notification senders available. Please create a notification sender first.',

    // Passkey executor modes
    'core.executions.passkey.mode.challenge': 'Challenge',
    'core.executions.passkey.mode.verify': 'Verify',
    'core.executions.passkey.mode.registerStart': 'Start Registration',
    'core.executions.passkey.mode.registerFinish': 'Finish Registration',
    'core.executions.passkey.mode.label': 'Mode',
    'core.executions.passkey.mode.placeholder': 'Select a mode',
    'core.executions.passkey.description': 'Configure the Passkey executor settings.',

    // Passkey relying party configuration
    'core.executions.passkey.relyingPartyId.label': 'Relying Party ID',
    'core.executions.passkey.relyingPartyId.placeholder': 'e.g., localhost or example.com',
    'core.executions.passkey.relyingPartyId.hint':
      'The domain identifier for passkey registration (typically your domain name)',
    'core.executions.passkey.relyingPartyName.label': 'Relying Party Name',
    'core.executions.passkey.relyingPartyName.placeholder': 'e.g., My Application',
    'core.executions.passkey.relyingPartyName.hint': 'A human-readable name shown to users during passkey registration',

    // Execution steps - tooltips and messages
    'core.executions.tooltip.configurationHint': 'Click to configure this step',
    'core.executions.landing.message': 'This {{executor}} step will redirect users to a landing page.',

    // Execution steps - branching handles
    'core.executions.handles.success': 'onSuccess',
    'core.executions.handles.failure': 'onFailure',

    // Canvas hints and tips
    'core.canvas.hints.autoLayout': 'Tip: Use auto-layout to organize your flow',
    'core.canvas.buttons.autoLayout': 'Auto Layout',

    // Steps - end
    'core.steps.end.flowCompletionProperties': 'Flow Completion Properties',

    // Validation messages - input fields
    'core.validation.fields.input.general':
      'Required fields are not properly configured for the input field with ID <code>{{id}}</code>.',
    'core.validation.fields.input.idpName': 'Identity provider name is required',
    'core.validation.fields.input.idpId': 'Connection is required',
    'core.validation.fields.input.senderId': 'Notification sender is required',
    'core.validation.fields.input.label': 'Label is required',
    'core.validation.fields.input.ref': 'Attribute is required',

    // Validation messages - executor
    'core.validation.fields.executor.general': 'The executor <0>{{id}}</0> is not properly configured.',

    // Validation messages - button
    'core.validation.fields.button.general':
      'Required fields are not properly configured for the button with ID <code>{{id}}</code>.',
    'core.validation.fields.button.text': 'Button text is required',
    'core.validation.fields.button.label': 'Label is required',
    'core.validation.fields.button.action': 'Action is required',
    'core.validation.fields.button.variant': 'Variant is required',

    // Validation messages - checkbox
    'core.validation.fields.checkbox.general':
      'Required fields are not properly configured for the checkbox with ID <code>{{id}}</code>.',
    'core.validation.fields.checkbox.label': 'Label is required',
    'core.validation.fields.checkbox.ref': 'Attribute is required',

    // Validation messages - divider
    'core.validation.fields.divider.general':
      'Required fields are not properly configured for the divider with ID <code>{{id}}</code>.',
    'core.validation.fields.divider.variant': 'Variant is required',

    // Validation messages - typography
    'core.validation.fields.typography.general':
      'Required fields are not properly configured for the typography with ID <code>{{id}}</code>.',
    'core.validation.fields.typography.text': 'Text content is required',
    'core.validation.fields.typography.label': 'Label is required',
    'core.validation.fields.typography.variant': 'Variant is required',

    // Validation messages - image
    'core.validation.fields.image.general':
      'Required fields are not properly configured for the image with ID <code>{{id}}</code>.',
    'core.validation.fields.image.src': 'Image source is required',
    'core.validation.fields.image.variant': 'Variant is required',

    // Placeholders
    'core.placeholders.image': 'No image source',

    // Validation messages - rich text
    'core.validation.fields.richText.general':
      'Required fields are not properly configured for the rich text with ID <code>{{id}}</code>.',
    'core.validation.fields.richText.text': 'Rich text content is required',
    'core.validation.fields.richText.label': 'Label is required',

    // Validation messages - OTP input
    'core.validation.fields.otpInput.label': 'OTP input label is required',

    // Validation messages - phone number input
    'core.validation.fields.phoneNumberInput.label': 'Phone number label is required',
    'core.validation.fields.phoneNumberInput.ref': 'Phone number attribute is required',

    // Elements - rich text
    'core.elements.richText.placeholder': 'Enter text here...',
    'core.elements.richText.resolvedI18nValue': 'Resolved i18n value',
    'core.elements.richText.linkEditor.urlTypeLabel': 'URL Type',
    'core.elements.richText.linkEditor.placeholder': 'Enter URL',
    'core.elements.richText.linkEditor.editLink': 'Edit Link',
    'core.elements.richText.linkEditor.viewLink': 'Link',

    // Elements - text property field
    'core.elements.textPropertyField.placeholder': 'Enter {{propertyName}}',
    'core.elements.textPropertyField.tooltip.configureTranslation': 'Configure translation',
    'core.elements.textPropertyField.i18nKey': 'Translation Key',
    'core.elements.textPropertyField.resolvedValue': 'Resolved Value',

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
    'core.elements.textPropertyField.i18nCard.invalidKeyFormat':
      'Invalid key format. Use only letters, numbers, dots, underscores, and hyphens.',

    // Form requires view dialog
    'core.dialogs.formRequiresView.formOnCanvas.title': 'Form Requires a View',
    'core.dialogs.formRequiresView.formOnCanvas.description':
      'Form components cannot be placed directly on the canvas. They must be inside a View component.',
    'core.dialogs.formRequiresView.formOnCanvas.alertMessage':
      'Would you like to create a View and add the Form inside it?',
    'core.dialogs.formRequiresView.formOnCanvas.confirmButton': 'Add View with Form',
    'core.dialogs.formRequiresView.inputOnCanvas.title': 'Input Requires a Form and View',
    'core.dialogs.formRequiresView.inputOnCanvas.description':
      'Input components cannot be placed directly on the canvas. They must be inside a Form, which must be inside a View.',
    'core.dialogs.formRequiresView.inputOnCanvas.alertMessage':
      'Would you like to create a View with a Form and add the Input inside it?',
    'core.dialogs.formRequiresView.inputOnCanvas.confirmButton': 'Add View, Form and Input',
    'core.dialogs.formRequiresView.inputOnView.title': 'Input Requires a Form',
    'core.dialogs.formRequiresView.inputOnView.description':
      'Input components cannot be placed directly inside a View. They must be inside a Form component.',
    'core.dialogs.formRequiresView.inputOnView.alertMessage':
      'Would you like to create a Form and add the Input inside it?',
    'core.dialogs.formRequiresView.inputOnView.confirmButton': 'Add Form with Input',
    'core.dialogs.formRequiresView.widgetOnCanvas.title': 'Widget Requires a View',
    'core.dialogs.formRequiresView.widgetOnCanvas.description':
      'Widgets cannot be placed directly on the canvas. They must be inside a View component.',
    'core.dialogs.formRequiresView.widgetOnCanvas.alertMessage':
      'Would you like to create a View and add the Widget inside it?',
    'core.dialogs.formRequiresView.widgetOnCanvas.confirmButton': 'Add View with Widget',
    'core.dialogs.formRequiresView.cancelButton': 'Cancel',

    // Form adapter
    'core.adapters.form.badgeLabel': 'Form',
    'core.adapters.form.placeholder': 'DROP FORM COMPONENTS HERE',

    // Header panel
    'core.headerPanel.goBack': 'Go back to Flows',
    'core.headerPanel.autoLayout': 'Auto Layout',
    'core.headerPanel.save': 'Save',
    'core.headerPanel.editTitle': 'Edit flow name',
    'core.headerPanel.saveTitle': 'Save flow name',
    'core.headerPanel.cancelEdit': 'Cancel',
    'core.headerPanel.edgeStyleTooltip': 'Change edge style',
    'core.headerPanel.edgeStyles.bezier': 'Bezier',
    'core.headerPanel.edgeStyles.smoothStep': 'Smooth Step',
    'core.headerPanel.edgeStyles.step': 'Step',

    // Resource panel
    'core.resourcePanel.title': 'Resources',
    'core.resourcePanel.showResources': 'Show Resources',
    'core.resourcePanel.hideResources': 'Hide Resources',
    'core.resourcePanel.starterTemplates.title': 'Starter Templates',
    'core.resourcePanel.starterTemplates.description':
      'Choose one of these templates to start building registration experience',
    'core.resourcePanel.widgets.title': 'Widgets',
    'core.resourcePanel.widgets.description': 'Use these widgets to build up the flow using pre-created flow blocks',
    'core.resourcePanel.steps.title': 'Steps',
    'core.resourcePanel.steps.description': 'Use these as steps in your flow',
    'core.resourcePanel.components.title': 'Components',
    'core.resourcePanel.components.description': 'Use these components to build up your views',
    'core.resourcePanel.executors.title': 'Executors',
    'core.resourcePanel.executors.description': 'Add authentication executors to your flow',

    // View step
    'core.steps.view.addComponent': 'Add Component',
    'core.steps.view.configure': 'Configure',
    'core.steps.view.remove': 'Remove',
    'core.steps.view.noComponentsAvailable': 'No components available',

    // Rule
    'core.rule.conditionalRule': 'Conditional Rule',
    'core.rule.remove': 'Remove',

    // Field extended properties
    'core.fieldExtendedProperties.attribute': 'Attribute',
    'core.fieldExtendedProperties.selectAttribute': 'Select an attribute',

    // Button extended properties
    'core.buttonExtendedProperties.type': 'Type',
    'core.buttonExtendedProperties.startIcon.label': 'Start Icon',
    'core.buttonExtendedProperties.startIcon.placeholder': 'Enter icon path (e.g., assets/images/icons/icon.svg)',
    'core.buttonExtendedProperties.startIcon.hint': 'Optional icon displayed before the button label',
    'core.buttonExtendedProperties.endIcon.label': 'End Icon',
    'core.buttonExtendedProperties.endIcon.placeholder': 'Enter icon path (e.g., assets/images/icons/icon.svg)',
    'core.buttonExtendedProperties.endIcon.hint': 'Optional icon displayed after the button label',

    // Rules properties
    'core.rulesProperties.description': 'Define a rule to how conditionally proceed to next steps in the flow',

    // Login flow builder
    'core.loginFlowBuilder.form': 'Form',
    'core.loginFlowBuilder.errors.validationRequired': 'Please fix all validation errors before saving.',
    'core.loginFlowBuilder.errors.structureValidationFailed': 'Flow structure validation failed: {{error}}',
    'core.loginFlowBuilder.errors.saveFailed': 'Failed to save flow. Please try again.',
    'core.loginFlowBuilder.success.flowCreated': 'Flow created successfully.',
    'core.loginFlowBuilder.success.flowUpdated': 'Flow updated successfully.',
  },

  /**
   * Appearance namespace - Theme and layout related translations
   */
  appearance: {
    'theme.defaultTheme': 'Default Theme',
    'theme.appTheme.displayName': '{{appName}} Theme',
  },
} as const;

export default translations;
