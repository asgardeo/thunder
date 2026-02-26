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
const LOAD_MORE_SUFFIX = '__loadMore';
const ROOT_PARENT_ID = '__root';
const ROOT_LOAD_MORE_ID = `${ROOT_PARENT_ID}${LOAD_MORE_SUFFIX}`;
const PAGE_SIZE = 30;

function PickerLoadingIcon(): JSX.Element {
  return <CircularProgress size={16} />;
}

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

function appendTreeItemChildren(
  items: OrganizationUnitTreeItem[],
  parentId: string,
  newChildren: OrganizationUnitTreeItem[],
): OrganizationUnitTreeItem[] {
  return items.map((item) => {
    if (item.id === parentId) {
      // Remove any existing load-more placeholder before appending
      const existing = (item.children ?? []).filter((c) => !c.id.endsWith(LOAD_MORE_SUFFIX));

      return {...item, children: [...existing, ...newChildren]};
    }

    if (item.children && item.children.length > 0) {
      return {...item, children: appendTreeItemChildren(item.children, parentId, newChildren)};
    }

    return item;
  });
}

interface PickerTreeItemProps extends TreeView.TreeItemProps {
  itemId: string;
  itemMap?: Map<string, OrganizationUnitTreeItem>;
  loadingItems?: Set<string>;
  loadMoreLoadingItems?: Set<string>;
  onLoadMore?: (parentId: string) => void;
}

function PickerTreeItem(allProps: PickerTreeItemProps): JSX.Element {
  const {
    itemMap: itemMapProp,
    loadingItems: loadingItemsProp,
    loadMoreLoadingItems: loadMoreLoadingItemsProp,
    onLoadMore: onLoadMoreProp,
    itemId,
    label,
    ...restProps
  } = allProps;
  const treeItemProps = {itemId, label, ...restProps};
  const theme = useTheme();
  const {t} = useTranslation();
  const labelStr = typeof label === 'string' ? label : '';
  const itemData = itemMapProp?.get(itemId);
  const isLoadMoreItem = itemId.endsWith(LOAD_MORE_SUFFIX);
  const isEmptyPlaceholder = itemId.endsWith(EMPTY_SUFFIX);
  const isLoadingPlaceholder =
    !isEmptyPlaceholder && !isLoadMoreItem && (itemData?.isPlaceholder ?? itemId.endsWith(PLACEHOLDER_SUFFIX));
  const isItemLoading = loadingItemsProp?.has(itemId);

  if (isLoadMoreItem) {
    const parentId = itemId.replace(LOAD_MORE_SUFFIX, '');
    const isLoadingMore = loadMoreLoadingItemsProp?.has(parentId);

    return (
      <TreeView.TreeItem
        {...treeItemProps}
        sx={{
          '& > .MuiTreeItem-content': {
            border: '1px dashed',
            borderColor: theme.vars?.palette.divider,
            borderRadius: 0.5,
            backgroundColor: 'transparent !important',
            cursor: isLoadingMore ? 'default' : 'pointer',
            transition: 'all 0.15s ease-in-out',
            '&:hover': {
              borderColor: isLoadingMore ? undefined : theme.vars?.palette.primary.main,
            },
          },
        }}
        label={
          <Box
            role="button"
            tabIndex={0}
            onClick={(e) => {
              e.stopPropagation();
              if (!isLoadingMore) {
                onLoadMoreProp?.(parentId);
              }
            }}
            onKeyDown={(e) => {
              if ((e.key === 'Enter' || e.key === ' ') && !isLoadingMore) {
                e.preventDefault();
                e.stopPropagation();
                onLoadMoreProp?.(parentId);
              }
            }}
            sx={{display: 'flex', alignItems: 'center', justifyContent: 'center', gap: 1, py: 0.25}}
          >
            {isLoadingMore ? (
              <>
                <CircularProgress size={14} />
                <Typography variant="caption" color="text.secondary">
                  {t('common:status.loading')}
                </Typography>
              </>
            ) : (
              <Typography variant="caption" color="primary" sx={{fontWeight: 500}}>
                {t('organizationUnits:listing.treeView.loadMore')}
              </Typography>
            )}
          </Box>
        }
      />
    );
  }

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
      {...(isItemLoading
        ? {slots: {collapseIcon: PickerLoadingIcon, expandIcon: PickerLoadingIcon}}
        : {})}
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
  const [loadMoreLoadingItems, setLoadMoreLoadingItems] = useState<Set<string>>(new Set());
  const [childOffsets, setChildOffsets] = useState<Map<string, number>>(new Map());
  const [rootOffset, setRootOffset] = useState<number>(0);
  const [rootLoadMoreLoading, setRootLoadMoreLoading] = useState<boolean>(false);
  const rootLoadMoreLoadingRef = useRef<boolean>(false);
  rootLoadMoreLoadingRef.current = rootLoadMoreLoading;
  const loadingItemsRef = useRef<Set<string>>(loadingItems);
  loadingItemsRef.current = loadingItems;

  const itemMap = useMemo(() => buildItemMap(treeItems), [treeItems]);

  // Build root tree when data arrives
  useEffect(() => {
    if (data?.organizationUnits && data.organizationUnits.length > 0 && treeItems.length === 0) {
      const items = buildTreeItems(data.organizationUnits);

      if (data.organizationUnits.length < data.totalResults) {
        items.push({
          id: ROOT_LOAD_MORE_ID,
          label: '',
          handle: '',
          isPlaceholder: true,
        });
      }

      setRootOffset(data.organizationUnits.length);
      setTreeItems(items);
    }
  }, [data, treeItems.length]);

  const fetchChildPage = useCallback(
    async (parentId: string, offset: number): Promise<OrganizationUnitListResponse> => {
      const serverUrl = getServerUrl();

      return queryClient.fetchQuery<OrganizationUnitListResponse>({
        queryKey: [OrganizationUnitQueryKeys.CHILD_ORGANIZATION_UNITS, parentId, {limit: PAGE_SIZE, offset}],
        queryFn: async (): Promise<OrganizationUnitListResponse> => {
          const queryParams = new URLSearchParams({limit: String(PAGE_SIZE), offset: String(offset)});
          const response: {data: OrganizationUnitListResponse} = await http.request({
            url: `${serverUrl}/organization-units/${encodeURIComponent(parentId)}/ous?${queryParams.toString()}`,
            method: 'GET',
            headers: {'Content-Type': 'application/json'},
          } as unknown as Parameters<typeof http.request>[0]);

          return response.data;
        },
        staleTime: 0,
      });
    },
    [getServerUrl, queryClient, http],
  );

  const buildChildItems = useCallback(
    (parentId: string, result: OrganizationUnitListResponse, offset: number): OrganizationUnitTreeItem[] => {
      const childOUs = result.organizationUnits;

      if (childOUs.length === 0 && offset === 0) {
        return [
          {
            id: `${parentId}${EMPTY_SUFFIX}`,
            label: t('organizationUnits:listing.treeView.noChildren'),
            handle: '',
            isPlaceholder: true,
          },
        ];
      }

      const items = buildTreeItems(childOUs);
      const loadedSoFar = offset + childOUs.length;

      if (loadedSoFar < result.totalResults) {
        items.push({
          id: `${parentId}${LOAD_MORE_SUFFIX}`,
          label: '',
          handle: '',
          isPlaceholder: true,
        });
      }

      return items;
    },
    [t],
  );

  const fetchChildOUs = useCallback(
    async (parentId: string): Promise<void> => {
      if (loadingItemsRef.current.has(parentId)) return;

      setLoadingItems((prev) => new Set(prev).add(parentId));

      try {
        const result = await fetchChildPage(parentId, 0);
        const childItems = buildChildItems(parentId, result, 0);

        setChildOffsets((prev) => new Map(prev).set(parentId, result.organizationUnits.length));
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
    [fetchChildPage, buildChildItems, logger],
  );

  const handleRootLoadMore = useCallback(
    async (): Promise<void> => {
      if (rootLoadMoreLoadingRef.current) return;

      setRootLoadMoreLoading(true);

      try {
        const serverUrl = getServerUrl();
        const result = await queryClient.fetchQuery<OrganizationUnitListResponse>({
          queryKey: [OrganizationUnitQueryKeys.ORGANIZATION_UNITS, {limit: PAGE_SIZE, offset: rootOffset}],
          queryFn: async (): Promise<OrganizationUnitListResponse> => {
            const queryParams = new URLSearchParams({limit: String(PAGE_SIZE), offset: String(rootOffset)});
            const response: {data: OrganizationUnitListResponse} = await http.request({
              url: `${serverUrl}/organization-units?${queryParams.toString()}`,
              method: 'GET',
              headers: {'Content-Type': 'application/json'},
            } as unknown as Parameters<typeof http.request>[0]);

            return response.data;
          },
          staleTime: 0,
        });

        const newItems = buildTreeItems(result.organizationUnits);
        const loadedSoFar = rootOffset + result.organizationUnits.length;

        if (loadedSoFar < result.totalResults) {
          newItems.push({
            id: ROOT_LOAD_MORE_ID,
            label: '',
            handle: '',
            isPlaceholder: true,
          });
        }

        setRootOffset(loadedSoFar);
        setTreeItems((prev) => {
          const withoutLoadMore = prev.filter((item) => item.id !== ROOT_LOAD_MORE_ID);

          return [...withoutLoadMore, ...newItems];
        });
      } catch (_error: unknown) {
        logger.error('Failed to load more root organization units', {error: _error});
      } finally {
        setRootLoadMoreLoading(false);
      }
    },
    [rootOffset, getServerUrl, queryClient, http, logger],
  );

  const handleLoadMore = useCallback(
    async (parentId: string): Promise<void> => {
      if (parentId === ROOT_PARENT_ID) {
        await handleRootLoadMore();

        return;
      }

      setLoadMoreLoadingItems((prev) => new Set(prev).add(parentId));

      try {
        const offset = childOffsets.get(parentId) ?? PAGE_SIZE;
        const result = await fetchChildPage(parentId, offset);
        const newItems = buildChildItems(parentId, result, offset);

        setChildOffsets((prev) => new Map(prev).set(parentId, offset + result.organizationUnits.length));
        setTreeItems((prev) => appendTreeItemChildren(prev, parentId, newItems));
      } catch (_error: unknown) {
        logger.error('Failed to load more child organization units', {error: _error, parentId});
      } finally {
        setLoadMoreLoadingItems((prev) => {
          const next = new Set(prev);
          next.delete(parentId);

          return next;
        });
      }
    },
    [childOffsets, fetchChildPage, buildChildItems, logger, handleRootLoadMore],
  );

  const combinedLoadMoreLoadingItems = useMemo(() => {
    if (!rootLoadMoreLoading) return loadMoreLoadingItems;
    const combined = new Set(loadMoreLoadingItems);
    combined.add(ROOT_PARENT_ID);

    return combined;
  }, [loadMoreLoadingItems, rootLoadMoreLoading]);

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
      if (
        itemId &&
        !itemId.endsWith(PLACEHOLDER_SUFFIX) &&
        !itemId.endsWith(EMPTY_SUFFIX) &&
        !itemId.endsWith(LOAD_MORE_SUFFIX)
      ) {
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
              loadMoreLoadingItems: combinedLoadMoreLoadingItems,
              onLoadMore: (parentId: string) => {
                handleLoadMore(parentId).catch((_error: unknown) => {
                  logger.error('Failed to load more child organization units', {error: _error, parentId});
                });
              },
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
