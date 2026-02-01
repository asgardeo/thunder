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

import {describe, it, expect, vi} from 'vitest';
import {renderHook} from '@thunder/test-utils';
import useDataGridLocaleText from '../useDataGridLocaleText';

// Unmock the hook for testing the actual implementation
vi.unmock('@/hooks/useDataGridLocaleText');

describe('useDataGridLocaleText', () => {
  it('should return localized text object for DataGrid', () => {
    const {result} = renderHook(() => useDataGridLocaleText());

    expect(result.current).toBeDefined();
    expect(typeof result.current).toBe('object');
  });

  it('should return root labels', () => {
    const {result} = renderHook(() => useDataGridLocaleText());

    expect(result.current.noRowsLabel).toBe('No rows');
    expect(result.current.noResultsOverlayLabel).toBe('No results found.');
    expect(result.current.noColumnsOverlayLabel).toBe('No columns');
    expect(result.current.noColumnsOverlayManageColumns).toBe('Manage columns');
  });

  it('should return toolbar density labels', () => {
    const {result} = renderHook(() => useDataGridLocaleText());

    expect(result.current.toolbarDensity).toBe('Density');
    expect(result.current.toolbarDensityLabel).toBe('Density');
    expect(result.current.toolbarDensityCompact).toBe('Compact');
    expect(result.current.toolbarDensityStandard).toBe('Standard');
    expect(result.current.toolbarDensityComfortable).toBe('Comfortable');
  });

  it('should return toolbar columns labels', () => {
    const {result} = renderHook(() => useDataGridLocaleText());

    expect(result.current.toolbarColumns).toBe('Columns');
    expect(result.current.toolbarColumnsLabel).toBe('Select columns');
  });

  it('should return toolbar filters labels', () => {
    const {result} = renderHook(() => useDataGridLocaleText());

    expect(result.current.toolbarFilters).toBe('Filters');
    expect(result.current.toolbarFiltersLabel).toBe('Show filters');
    expect(result.current.toolbarFiltersTooltipHide).toBe('Hide filters');
    expect(result.current.toolbarFiltersTooltipShow).toBe('Show filters');
  });

  it('should return toolbar filters tooltip active as a function', () => {
    const {result} = renderHook(() => useDataGridLocaleText());

    expect(typeof result.current.toolbarFiltersTooltipActive).toBe('function');
    if (result.current.toolbarFiltersTooltipActive) {
      expect(result.current.toolbarFiltersTooltipActive(5)).toBe('5 active filters');
      expect(result.current.toolbarFiltersTooltipActive(1)).toBe('1 active filter');
    }
  });

  it('should return quick filter labels', () => {
    const {result} = renderHook(() => useDataGridLocaleText());

    expect(result.current.toolbarQuickFilterPlaceholder).toBe('Search…');
    expect(result.current.toolbarQuickFilterLabel).toBe('Search');
    expect(result.current.toolbarQuickFilterDeleteIconLabel).toBe('Clear');
  });

  it('should return export labels', () => {
    const {result} = renderHook(() => useDataGridLocaleText());

    expect(result.current.toolbarExport).toBe('Export');
    expect(result.current.toolbarExportLabel).toBe('Export');
    expect(result.current.toolbarExportCSV).toBe('Download as CSV');
    expect(result.current.toolbarExportPrint).toBe('Print');
  });

  it('should return columns management labels', () => {
    const {result} = renderHook(() => useDataGridLocaleText());

    expect(result.current.columnsManagementSearchTitle).toBe('Search');
    expect(result.current.columnsManagementNoColumns).toBe('No columns');
    expect(result.current.columnsManagementShowHideAllText).toBe('Show/Hide All');
    expect(result.current.columnsManagementReset).toBe('Reset');
  });

  it('should return filter panel labels', () => {
    const {result} = renderHook(() => useDataGridLocaleText());

    expect(result.current.filterPanelAddFilter).toBe('Add filter');
    expect(result.current.filterPanelRemoveAll).toBe('Remove all');
    expect(result.current.filterPanelDeleteIconLabel).toBe('Delete');
    expect(result.current.filterPanelLogicOperator).toBe('Logic operator');
    expect(result.current.filterPanelOperator).toBe('Operator');
    expect(result.current.filterPanelOperatorAnd).toBe('And');
    expect(result.current.filterPanelOperatorOr).toBe('Or');
    expect(result.current.filterPanelColumns).toBe('Columns');
    expect(result.current.filterPanelInputLabel).toBe('Value');
    expect(result.current.filterPanelInputPlaceholder).toBe('Filter value');
  });

  it('should return filter operator labels', () => {
    const {result} = renderHook(() => useDataGridLocaleText());

    expect(result.current.filterOperatorContains).toBe('contains');
    expect(result.current.filterOperatorDoesNotContain).toBe('does not contain');
    expect(result.current.filterOperatorEquals).toBe('equals');
    expect(result.current.filterOperatorDoesNotEqual).toBe('does not equal');
    expect(result.current.filterOperatorStartsWith).toBe('starts with');
    expect(result.current.filterOperatorEndsWith).toBe('ends with');
    expect(result.current.filterOperatorIs).toBe('is');
    expect(result.current.filterOperatorNot).toBe('is not');
    expect(result.current.filterOperatorAfter).toBe('is after');
    expect(result.current.filterOperatorOnOrAfter).toBe('is on or after');
    expect(result.current.filterOperatorBefore).toBe('is before');
    expect(result.current.filterOperatorOnOrBefore).toBe('is on or before');
    expect(result.current.filterOperatorIsEmpty).toBe('is empty');
    expect(result.current.filterOperatorIsNotEmpty).toBe('is not empty');
    expect(result.current.filterOperatorIsAnyOf).toBe('is any of');
  });

  it('should return filter value labels', () => {
    const {result} = renderHook(() => useDataGridLocaleText());

    expect(result.current.filterValueAny).toBe('any');
    expect(result.current.filterValueTrue).toBe('true');
    expect(result.current.filterValueFalse).toBe('false');
  });

  it('should return column menu labels', () => {
    const {result} = renderHook(() => useDataGridLocaleText());

    expect(result.current.columnMenuLabel).toBe('Menu');
    expect(result.current.columnMenuShowColumns).toBe('Show columns');
    expect(result.current.columnMenuManageColumns).toBe('Manage columns');
    expect(result.current.columnMenuFilter).toBe('Filter');
    expect(result.current.columnMenuHideColumn).toBe('Hide column');
    expect(result.current.columnMenuUnsort).toBe('Unsort');
    expect(result.current.columnMenuSortAsc).toBe('Sort by ASC');
    expect(result.current.columnMenuSortDesc).toBe('Sort by DESC');
  });

  it('should return column header labels as functions', () => {
    const {result} = renderHook(() => useDataGridLocaleText());

    expect(typeof result.current.columnHeaderFiltersTooltipActive).toBe('function');
    expect(result.current.columnHeaderFiltersLabel).toBe('Show filters');
    expect(result.current.columnHeaderSortIconLabel).toBe('Sort');
    if (result.current.columnHeaderFiltersTooltipActive) {
      expect(result.current.columnHeaderFiltersTooltipActive(3)).toBe('3 active filters');
      expect(result.current.columnHeaderFiltersTooltipActive(1)).toBe('1 active filter');
    }
  });

  it('should return footer labels as functions', () => {
    const {result} = renderHook(() => useDataGridLocaleText());

    expect(typeof result.current.footerRowSelected).toBe('function');
    expect(result.current.footerTotalRows).toBe('Total Rows:');
    expect(typeof result.current.footerTotalVisibleRows).toBe('function');
    if (result.current.footerRowSelected) {
      expect(result.current.footerRowSelected(10)).toBe('10 rows selected');
      expect(result.current.footerRowSelected(1)).toBe('1 row selected');
    }
    if (result.current.footerTotalVisibleRows) {
      expect(result.current.footerTotalVisibleRows(50, 100)).toBe('50 of 100');
    }
  });

  it('should return checkbox selection labels', () => {
    const {result} = renderHook(() => useDataGridLocaleText());

    expect(result.current.checkboxSelectionHeaderName).toBe('Checkbox selection');
    expect(result.current.checkboxSelectionSelectAllRows).toBe('Select all rows');
    expect(result.current.checkboxSelectionUnselectAllRows).toBe('Unselect all rows');
    expect(result.current.checkboxSelectionSelectRow).toBe('Select row');
    expect(result.current.checkboxSelectionUnselectRow).toBe('Unselect row');
  });

  it('should return boolean cell labels', () => {
    const {result} = renderHook(() => useDataGridLocaleText());

    expect(result.current.booleanCellTrueLabel).toBe('yes');
    expect(result.current.booleanCellFalseLabel).toBe('no');
  });

  it('should return actions cell label', () => {
    const {result} = renderHook(() => useDataGridLocaleText());

    expect(result.current.actionsCellMore).toBe('more');
  });

  it('should return column pinning labels', () => {
    const {result} = renderHook(() => useDataGridLocaleText());

    expect(result.current.pinToLeft).toBe('Pin to left');
    expect(result.current.pinToRight).toBe('Pin to right');
    expect(result.current.unpin).toBe('Unpin');
  });

  it('should return tree data labels', () => {
    const {result} = renderHook(() => useDataGridLocaleText());

    expect(result.current.treeDataGroupingHeaderName).toBe('Group');
    expect(result.current.treeDataExpand).toBe('see children');
    expect(result.current.treeDataCollapse).toBe('hide children');
  });

  it('should return grouping labels as functions', () => {
    const {result} = renderHook(() => useDataGridLocaleText());

    expect(result.current.groupingColumnHeaderName).toBe('Group');
    expect(typeof result.current.groupColumn).toBe('function');
    expect(typeof result.current.unGroupColumn).toBe('function');
    if (result.current.groupColumn) {
      expect(result.current.groupColumn('status')).toBe('Group by status');
    }
    if (result.current.unGroupColumn) {
      expect(result.current.unGroupColumn('status')).toBe('Stop grouping by status');
    }
  });

  it('should return master/detail labels', () => {
    const {result} = renderHook(() => useDataGridLocaleText());

    expect(result.current.detailPanelToggle).toBe('Detail panel toggle');
    expect(result.current.expandDetailPanel).toBe('Expand');
    expect(result.current.collapseDetailPanel).toBe('Collapse');
  });

  it('should return pagination labels', () => {
    const {result} = renderHook(() => useDataGridLocaleText());

    expect(result.current.paginationRowsPerPage).toBe('Rows per page:');
    expect(typeof result.current.paginationDisplayedRows).toBe('function');
    if (result.current.paginationDisplayedRows) {
      expect(result.current.paginationDisplayedRows({from: 1, to: 10, count: 100, estimated: undefined})).toBe(
        '1–10 of 100',
      );
      expect(result.current.paginationDisplayedRows({from: 1, to: 10, count: -1, estimated: undefined})).toBe(
        '1–10 of more than 10',
      );
    }
  });

  it('should return row reordering label', () => {
    const {result} = renderHook(() => useDataGridLocaleText());

    expect(result.current.rowReorderingHeaderName).toBe('Row reordering');
  });

  it('should return aggregation labels', () => {
    const {result} = renderHook(() => useDataGridLocaleText());

    expect(result.current.aggregationMenuItemHeader).toBe('Aggregation');
    expect(result.current.aggregationFunctionLabelSum).toBe('sum');
    expect(result.current.aggregationFunctionLabelAvg).toBe('avg');
    expect(result.current.aggregationFunctionLabelMin).toBe('min');
    expect(result.current.aggregationFunctionLabelMax).toBe('max');
    expect(result.current.aggregationFunctionLabelSize).toBe('size');
  });

  it('should memoize the result', () => {
    const {result, rerender} = renderHook(() => useDataGridLocaleText());
    const firstResult = result.current;

    rerender();

    // The result should be the same object reference due to memoization
    expect(result.current).toBe(firstResult);
  });
});
