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

import {useState, useCallback, useEffect, useRef, useMemo, type JSX, type SyntheticEvent} from 'react';
import {Box, Avatar, Typography, CircularProgress, TreeView, useTheme} from '@wso2/oxygen-ui';
import {Building} from '@wso2/oxygen-ui-icons-react';
import {useTranslation} from 'react-i18next';
import {useConfig} from '@thunder/shared-contexts';
import {useAsgardeo} from '@asgardeo/react';
import {useQueryClient} from '@tanstack/react-query';
import {useLogger} from '@thunder/logger/react';
import useGetOrganizationUnits from '../api/useGetOrganizationUnits';
import type {OrganizationUnit} from '../models/organization-unit';
import type {OrganizationUnitTreeItem} from '../models/organization-unit-tree';
import type {OrganizationUnitListResponse} from '../models/responses';
import OrganizationUnitQueryKeys from '../constants/organization-unit-query-keys';

const PLACEHOLDER_SUFFIX = '__placeholder';
const EMPTY_SUFFIX = '__empty';

function buildTreeItems(ous: OrganizationUnit[]): OrganizationUnitTreeItem[] {
  return ous.map((ou) => ({
    id: ou.id,
    label: ou.name,
    handle: ou.handle,
    description: ou.description,
    logo_url: ou.logo_url,
    children: [
      {
        id: `${ou.id}${PLACEHOLDER_SUFFIX}`,
        label: '',
        handle: '',
        isPlaceholder: true,
      },
    ],
  }));
}

function buildItemMap(items: OrganizationUnitTreeItem[]): Map<string, OrganizationUnitTreeItem> {
  const map = new Map<string, OrganizationUnitTreeItem>();
  const visit = (list: OrganizationUnitTreeItem[]): void => {
    list.forEach((item) => {
      map.set(item.id, item);
      if (item.children) visit(item.children);
    });
  };
  visit(items);

  return map;
}

function updateTreeItemChildren(
  items: OrganizationUnitTreeItem[],
  parentId: string,
  children: OrganizationUnitTreeItem[],
): OrganizationUnitTreeItem[] {
  return items.map((item) => {
    if (item.id === parentId) {
      return {...item, children};
    }

    if (item.children && item.children.length > 0) {
      return {...item, children: updateTreeItemChildren(item.children, parentId, children)};
    }

    return item;
  });
}

interface PickerTreeItemProps extends TreeView.TreeItemProps {
  itemId: string;
  itemMap?: Map<string, OrganizationUnitTreeItem>;
  loadingItems?: Set<string>;
}

function PickerTreeItem(allProps: PickerTreeItemProps): JSX.Element {
  const {itemMap: itemMapProp, loadingItems: loadingItemsProp, itemId, label, ...restProps} = allProps;
  const treeItemProps = {itemId, label, ...restProps};
  const theme = useTheme();
  const {t} = useTranslation();
  const labelStr = typeof label === 'string' ? label : '';
  const itemData = itemMapProp?.get(itemId);
  const isEmptyPlaceholder = itemId.endsWith(EMPTY_SUFFIX);
  const isLoadingPlaceholder = !isEmptyPlaceholder && (itemData?.isPlaceholder ?? itemId.endsWith(PLACEHOLDER_SUFFIX));
  const isItemLoading = loadingItemsProp?.has(itemId);

  if (isEmptyPlaceholder) {
    return (
      <TreeView.TreeItem
        {...treeItemProps}
        sx={{
          '& > .MuiTreeItem-content': {
            border: 'none !important',
            backgroundColor: 'transparent !important',
          },
        }}
        label={
          <Typography variant="caption" color="text.secondary" sx={{fontStyle: 'italic', pl: 1}}>
            {labelStr}
          </Typography>
        }
      />
    );
  }

  if (isLoadingPlaceholder) {
    return (
      <TreeView.TreeItem
        {...treeItemProps}
        sx={{
          '& > .MuiTreeItem-content': {
            border: 'none !important',
            backgroundColor: 'transparent !important',
          },
        }}
        label={
          <Box sx={{display: 'flex', alignItems: 'center', gap: 1}}>
            <CircularProgress size={16} />
            <Typography variant="caption" color="text.secondary" sx={{fontStyle: 'italic'}}>
              {t('common:status.loading')}
            </Typography>
          </Box>
        }
      />
    );
  }

  return (
    <TreeView.TreeItem
      {...treeItemProps}
      label={
        <Box sx={{display: 'flex', alignItems: 'center', gap: 1.5}}>
          <Avatar
            sx={{
              p: 0.5,
              backgroundColor: theme?.vars?.palette.primary.main,
              width: 28,
              height: 28,
              fontSize: '0.75rem',
            }}
            src={itemData?.logo_url}
          >
            <Building size={12} />
          </Avatar>
          <Box sx={{flexGrow: 1, minWidth: 0}}>
            <Typography variant="body2" sx={{fontWeight: 500, lineHeight: 1.3}}>
              {labelStr}
            </Typography>
            {itemData?.handle && (
              <Typography variant="caption" color="text.secondary" sx={{lineHeight: 1.2, display: 'block'}}>
                {itemData.handle}
              </Typography>
            )}
          </Box>
          {isItemLoading && <CircularProgress size={16} />}
        </Box>
      }
    />
  );
}

interface OrganizationUnitTreePickerProps {
  id?: string;
  value: string;
  onChange: (ouId: string) => void;
  error?: boolean;
  helperText?: string;
}

export default function OrganizationUnitTreePicker({
  id = undefined,
  value,
  onChange,
  error = false,
  helperText = '',
}: OrganizationUnitTreePickerProps): JSX.Element {
  const theme = useTheme();
  const {t} = useTranslation();
  const logger = useLogger('OrganizationUnitTreePicker');
  const {http} = useAsgardeo();
  const {getServerUrl} = useConfig();
  const queryClient = useQueryClient();
  const {data, isLoading} = useGetOrganizationUnits();

  const [treeItems, setTreeItems] = useState<OrganizationUnitTreeItem[]>([]);
  const [expandedItems, setExpandedItems] = useState<string[]>([]);
  const [loadedItems, setLoadedItems] = useState<Set<string>>(new Set());
  const [loadingItems, setLoadingItems] = useState<Set<string>>(new Set());
  const loadingItemsRef = useRef<Set<string>>(loadingItems);
  loadingItemsRef.current = loadingItems;

  const itemMap = useMemo(() => buildItemMap(treeItems), [treeItems]);

  // Build root tree when data arrives
  useEffect(() => {
    if (data?.organizationUnits && data.organizationUnits.length > 0 && treeItems.length === 0) {
      setTreeItems(buildTreeItems(data.organizationUnits));
    }
  }, [data, treeItems.length]);

  const fetchChildOUs = useCallback(
    async (parentId: string): Promise<void> => {
      if (loadingItemsRef.current.has(parentId)) return;

      setLoadingItems((prev) => new Set(prev).add(parentId));

      try {
        const serverUrl = getServerUrl();
        const result = await queryClient.fetchQuery<OrganizationUnitListResponse>({
          queryKey: [OrganizationUnitQueryKeys.CHILD_ORGANIZATION_UNITS, parentId, {limit: 30, offset: 0}],
          queryFn: async (): Promise<OrganizationUnitListResponse> => {
            const queryParams = new URLSearchParams({limit: '30', offset: '0'});
            const response: {data: OrganizationUnitListResponse} = await http.request({
              url: `${serverUrl}/organization-units/${encodeURIComponent(parentId)}/ous?${queryParams.toString()}`,
              method: 'GET',
              headers: {'Content-Type': 'application/json'},
            } as unknown as Parameters<typeof http.request>[0]);

            return response.data;
          },
          staleTime: 0,
        });

        const childOUs = result.organizationUnits;
        const childItems =
          childOUs.length > 0
            ? buildTreeItems(childOUs)
            : [
                {
                  id: `${parentId}${EMPTY_SUFFIX}`,
                  label: t('organizationUnits:listing.treeView.noChildren'),
                  handle: '',
                  isPlaceholder: true,
                },
              ];

        setTreeItems((prev) => updateTreeItemChildren(prev, parentId, childItems));
        setLoadedItems((prev) => new Set(prev).add(parentId));
        setExpandedItems((prev) => (prev.includes(parentId) ? prev : [...prev, parentId]));
      } catch (_error: unknown) {
        logger.error('Failed to load child organization units', {error: _error, parentId});
      } finally {
        setLoadingItems((prev) => {
          const next = new Set(prev);
          next.delete(parentId);

          return next;
        });
      }
    },
    [getServerUrl, queryClient, http, logger, t],
  );

  const handleItemExpansionToggle = useCallback(
    (_event: SyntheticEvent | null, itemId: string, isExpanded: boolean) => {
      if (!isExpanded || loadedItems.has(itemId) || loadingItems.has(itemId)) {
        return;
      }

      fetchChildOUs(itemId).catch((_error: unknown) => {
        logger.error('Failed to load child organization units', {error: _error, parentId: itemId});
      });
    },
    [loadedItems, loadingItems, fetchChildOUs, logger],
  );

  const handleSelectedItemsChange = useCallback(
    (_event: SyntheticEvent | null, itemId: string | null) => {
      if (itemId && !itemId.endsWith(PLACEHOLDER_SUFFIX) && !itemId.endsWith(EMPTY_SUFFIX)) {
        onChange(itemId);
      }
    },
    [onChange],
  );

  if (isLoading) {
    return (
      <Box sx={{display: 'flex', justifyContent: 'center', py: 4}}>
        <CircularProgress size={24} />
      </Box>
    );
  }

  if (data && data.organizationUnits.length === 0) {
    return (
      <Typography variant="body2" color="text.secondary">
        {t('organizationUnits:treePicker.empty')}
      </Typography>
    );
  }

  return (
    <Box>
      <Box
        sx={{
          maxHeight: 300,
          overflow: 'auto',
        }}
      >
        <TreeView.RichTreeView
          id={id}
          items={treeItems}
          expandedItems={expandedItems}
          onExpandedItemsChange={(_event: SyntheticEvent | null, itemIds: string[]) => {
            const prevSet = new Set(expandedItems);
            const filtered = itemIds.filter((itemId) => prevSet.has(itemId) || loadedItems.has(itemId));
            setExpandedItems(filtered);
          }}
          onItemExpansionToggle={handleItemExpansionToggle}
          selectedItems={value || null}
          onSelectedItemsChange={handleSelectedItemsChange}
          slots={{item: PickerTreeItem}}
          slotProps={{
            item: {
              itemMap,
              loadingItems,
            } as Record<string, unknown>,
          }}
          getItemLabel={(item: OrganizationUnitTreeItem) => item.label}
          sx={{
            '& .MuiTreeItem-content': {
              cursor: 'pointer',
              border: '1px solid',
              borderColor: theme.vars?.palette.divider,
              borderRadius: 0.5,
              py: 0.75,
              px: 1,
              mb: 0.5,
              transition: 'all 0.15s ease-in-out',
              '&:hover': {
                backgroundColor: theme.vars?.palette.action.hover,
                borderColor: theme.vars?.palette.primary.main,
              },
            },
            '& .Mui-selected > .MuiTreeItem-content': {
              backgroundColor: `${theme.vars?.palette.primary.main}14`,
              borderColor: theme.vars?.palette.primary.main,
            },
            '& .MuiTreeItem-iconContainer': {
              color: theme.vars?.palette.text.secondary,
              mr: 0.5,
            },
            '& .MuiTreeItem-groupTransition': {
              ml: 2,
              pl: 2,
              borderLeft: '1px dashed',
              borderColor: theme.vars?.palette.divider,
            },
          }}
        />
      </Box>
      {helperText && (
        <Typography variant="caption" color={error ? 'error' : 'text.secondary'} sx={{mt: 0.5, ml: 1.75}}>
          {helperText}
        </Typography>
      )}
    </Box>
  );
}
