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

import {useMemo} from 'react';
import type {DataGrid} from '@wso2/oxygen-ui';
import {useTranslation} from 'react-i18next';

/**
 * Helper function to safely extract and cast function values from translation resources
 *
 * @template T - The function type to cast to
 * @param commonBundle - The common namespace resource bundle
 * @param key - The key of the function in the resource bundle (e.g., 'dataTable.toolbarFiltersTooltipActive')
 * @returns The function cast to type T, or undefined if not found or invalid
 *
 * @remarks
 * This helper provides a consistent way to handle type assertions for function values
 * that cannot be accessed via i18next's t() function. It performs runtime validation
 * to ensure the value is actually a function before casting, preventing runtime errors
 * if the translation resource structure changes.
 *
 * Expected common bundle structure (flat with dot notation keys):
 * - Keys should map to either string values (handled by i18next t() function)
 * - Or function values that accept parameters and return strings
 *
 * @example
 * ```typescript
 * // Valid function in translation resource
 * {
 *   'dataTable.toolbarFiltersTooltipActive': (count: number) => `${count} active filter${count !== 1 ? 's' : ''}`
 * }
 * ```
 */
function getTranslationFunction<T>(commonBundle: Record<string, unknown>, key: string): T | undefined {
  const value = commonBundle[key];

  // Runtime validation: ensure the value is actually a function
  if (typeof value === 'function') {
    return value as T;
  }

  // Log a warning in development if the key exists but is not a function
  if (import.meta.env.DEV && value !== undefined) {
    // eslint-disable-next-line no-console
    console.warn(`Translation key '${key}' exists but is not a function. Expected a function, got ${typeof value}.`);
  }

  return undefined;
}

/**
 * Custom hook to get localized text for MUI DataGrid
 * @returns Localized text object for DataGrid
 */
export default function useDataGridLocaleText(): Partial<DataGrid.GridLocaleText> {
  const {t, i18n} = useTranslation();

  return useMemo(() => {
    // Get the common namespace translations directly from the resource store
    // This is necessary because i18next's t() function doesn't execute function values
    // With flat structure, function keys are like 'dataTable.toolbarFiltersTooltipActive'
    const commonBundle = (i18n.getResourceBundle(i18n.language, 'common') ?? {}) as Record<string, unknown>;

    return {
      // Root
      noRowsLabel: t('common:dataTable.noRowsLabel'),
      noResultsOverlayLabel: t('common:dataTable.noResultsOverlayLabel'),
      noColumnsOverlayLabel: t('common:dataTable.noColumnsOverlayLabel'),
      noColumnsOverlayManageColumns: t('common:dataTable.noColumnsOverlayManageColumns'),

      // Density selector toolbar button text
      toolbarDensity: t('common:dataTable.toolbarDensity'),
      toolbarDensityLabel: t('common:dataTable.toolbarDensityLabel'),
      toolbarDensityCompact: t('common:dataTable.toolbarDensityCompact'),
      toolbarDensityStandard: t('common:dataTable.toolbarDensityStandard'),
      toolbarDensityComfortable: t('common:dataTable.toolbarDensityComfortable'),

      // Columns selector toolbar button text
      toolbarColumns: t('common:dataTable.toolbarColumns'),
      toolbarColumnsLabel: t('common:dataTable.toolbarColumnsLabel'),

      // Filters toolbar button text
      toolbarFilters: t('common:dataTable.toolbarFilters'),
      toolbarFiltersLabel: t('common:dataTable.toolbarFiltersLabel'),
      toolbarFiltersTooltipHide: t('common:dataTable.toolbarFiltersTooltipHide'),
      toolbarFiltersTooltipShow: t('common:dataTable.toolbarFiltersTooltipShow'),
      toolbarFiltersTooltipActive: getTranslationFunction<(count: number) => string>(
        commonBundle,
        'dataTable.toolbarFiltersTooltipActive',
      ),

      // Quick filter toolbar field
      toolbarQuickFilterPlaceholder: t('common:dataTable.toolbarQuickFilterPlaceholder'),
      toolbarQuickFilterLabel: t('common:dataTable.toolbarQuickFilterLabel'),
      toolbarQuickFilterDeleteIconLabel: t('common:dataTable.toolbarQuickFilterDeleteIconLabel'),

      // Export selector toolbar button text
      toolbarExport: t('common:dataTable.toolbarExport'),
      toolbarExportLabel: t('common:dataTable.toolbarExportLabel'),
      toolbarExportCSV: t('common:dataTable.toolbarExportCSV'),
      toolbarExportPrint: t('common:dataTable.toolbarExportPrint'),

      // Columns management text
      columnsManagementSearchTitle: t('common:dataTable.columnsManagementSearchTitle'),
      columnsManagementNoColumns: t('common:dataTable.columnsManagementNoColumns'),
      columnsManagementShowHideAllText: t('common:dataTable.columnsManagementShowHideAllText'),
      columnsManagementReset: t('common:dataTable.columnsManagementReset'),

      // Filter panel text
      filterPanelAddFilter: t('common:dataTable.filterPanelAddFilter'),
      filterPanelRemoveAll: t('common:dataTable.filterPanelRemoveAll'),
      filterPanelDeleteIconLabel: t('common:dataTable.filterPanelDeleteIconLabel'),
      filterPanelLogicOperator: t('common:dataTable.filterPanelLogicOperator'),
      filterPanelOperator: t('common:dataTable.filterPanelOperator'),
      filterPanelOperatorAnd: t('common:dataTable.filterPanelOperatorAnd'),
      filterPanelOperatorOr: t('common:dataTable.filterPanelOperatorOr'),
      filterPanelColumns: t('common:dataTable.filterPanelColumns'),
      filterPanelInputLabel: t('common:dataTable.filterPanelInputLabel'),
      filterPanelInputPlaceholder: t('common:dataTable.filterPanelInputPlaceholder'),

      // Filter operators text
      filterOperatorContains: t('common:dataTable.filterOperatorContains'),
      filterOperatorDoesNotContain: t('common:dataTable.filterOperatorDoesNotContain'),
      filterOperatorEquals: t('common:dataTable.filterOperatorEquals'),
      filterOperatorDoesNotEqual: t('common:dataTable.filterOperatorDoesNotEqual'),
      filterOperatorStartsWith: t('common:dataTable.filterOperatorStartsWith'),
      filterOperatorEndsWith: t('common:dataTable.filterOperatorEndsWith'),
      filterOperatorIs: t('common:dataTable.filterOperatorIs'),
      filterOperatorNot: t('common:dataTable.filterOperatorNot'),
      filterOperatorAfter: t('common:dataTable.filterOperatorAfter'),
      filterOperatorOnOrAfter: t('common:dataTable.filterOperatorOnOrAfter'),
      filterOperatorBefore: t('common:dataTable.filterOperatorBefore'),
      filterOperatorOnOrBefore: t('common:dataTable.filterOperatorOnOrBefore'),
      filterOperatorIsEmpty: t('common:dataTable.filterOperatorIsEmpty'),
      filterOperatorIsNotEmpty: t('common:dataTable.filterOperatorIsNotEmpty'),
      filterOperatorIsAnyOf: t('common:dataTable.filterOperatorIsAnyOf'),

      // Filter values text
      filterValueAny: t('common:dataTable.filterValueAny'),
      filterValueTrue: t('common:dataTable.filterValueTrue'),
      filterValueFalse: t('common:dataTable.filterValueFalse'),

      // Column menu text
      columnMenuLabel: t('common:dataTable.columnMenuLabel'),
      columnMenuShowColumns: t('common:dataTable.columnMenuShowColumns'),
      columnMenuManageColumns: t('common:dataTable.columnMenuManageColumns'),
      columnMenuFilter: t('common:dataTable.columnMenuFilter'),
      columnMenuHideColumn: t('common:dataTable.columnMenuHideColumn'),
      columnMenuUnsort: t('common:dataTable.columnMenuUnsort'),
      columnMenuSortAsc: t('common:dataTable.columnMenuSortAsc'),
      columnMenuSortDesc: t('common:dataTable.columnMenuSortDesc'),

      // Column header text
      columnHeaderFiltersTooltipActive: getTranslationFunction<(count: number) => string>(
        commonBundle,
        'dataTable.columnHeaderFiltersTooltipActive',
      ),
      columnHeaderFiltersLabel: t('common:dataTable.columnHeaderFiltersLabel'),
      columnHeaderSortIconLabel: t('common:dataTable.columnHeaderSortIconLabel'),

      // Rows selected footer text
      footerRowSelected: getTranslationFunction<(count: number) => string>(
        commonBundle,
        'dataTable.footerRowSelected',
      ),

      // Total row amount footer text
      footerTotalRows: t('common:dataTable.footerTotalRows'),

      // Total visible row amount footer text
      footerTotalVisibleRows: getTranslationFunction<(visibleCount: number, totalCount: number) => string>(
        commonBundle,
        'dataTable.footerTotalVisibleRows',
      ),

      // Checkbox selection text
      checkboxSelectionHeaderName: t('common:dataTable.checkboxSelectionHeaderName'),
      checkboxSelectionSelectAllRows: t('common:dataTable.checkboxSelectionSelectAllRows'),
      checkboxSelectionUnselectAllRows: t('common:dataTable.checkboxSelectionUnselectAllRows'),
      checkboxSelectionSelectRow: t('common:dataTable.checkboxSelectionSelectRow'),
      checkboxSelectionUnselectRow: t('common:dataTable.checkboxSelectionUnselectRow'),

      // Boolean cell text
      booleanCellTrueLabel: t('common:dataTable.booleanCellTrueLabel'),
      booleanCellFalseLabel: t('common:dataTable.booleanCellFalseLabel'),

      // Actions cell more text
      actionsCellMore: t('common:dataTable.actionsCellMore'),

      // Column pinning text
      pinToLeft: t('common:dataTable.pinToLeft'),
      pinToRight: t('common:dataTable.pinToRight'),
      unpin: t('common:dataTable.unpin'),

      // Tree Data
      treeDataGroupingHeaderName: t('common:dataTable.treeDataGroupingHeaderName'),
      treeDataExpand: t('common:dataTable.treeDataExpand'),
      treeDataCollapse: t('common:dataTable.treeDataCollapse'),

      // Grouping columns
      groupingColumnHeaderName: t('common:dataTable.groupingColumnHeaderName'),
      groupColumn: getTranslationFunction<(name: string) => string>(commonBundle, 'dataTable.groupColumn'),
      unGroupColumn: getTranslationFunction<(name: string) => string>(commonBundle, 'dataTable.unGroupColumn'),

      // Master/detail
      detailPanelToggle: t('common:dataTable.detailPanelToggle'),
      expandDetailPanel: t('common:dataTable.expandDetailPanel'),
      collapseDetailPanel: t('common:dataTable.collapseDetailPanel'),

      // Pagination
      paginationRowsPerPage: t('common:dataTable.paginationRowsPerPage'),
      paginationDisplayedRows: getTranslationFunction<
        ({from, to, count}: {from: number; to: number; count: number}) => string
      >(commonBundle, 'dataTable.paginationDisplayedRows'),

      // Row reordering text
      rowReorderingHeaderName: t('common:dataTable.rowReorderingHeaderName'),

      // Aggregation
      aggregationMenuItemHeader: t('common:dataTable.aggregationMenuItemHeader'),
      aggregationFunctionLabelSum: t('common:dataTable.aggregationFunctionLabelSum'),
      aggregationFunctionLabelAvg: t('common:dataTable.aggregationFunctionLabelAvg'),
      aggregationFunctionLabelMin: t('common:dataTable.aggregationFunctionLabelMin'),
      aggregationFunctionLabelMax: t('common:dataTable.aggregationFunctionLabelMax'),
      aggregationFunctionLabelSize: t('common:dataTable.aggregationFunctionLabelSize'),
    };
  }, [t, i18n]);
}
