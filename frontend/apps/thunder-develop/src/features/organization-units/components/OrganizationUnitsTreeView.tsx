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

import {useState, useCallback, useEffect, useRef, useMemo} from 'react';
import type {ReactNode, MouseEvent, KeyboardEvent, SyntheticEvent, JSX} from 'react';
import {useNavigate} from 'react-router';
import {useLogger} from '@thunder/logger/react';
import {
  Box,
  Avatar,
  IconButton,
  Typography,
  CircularProgress,
  TreeView,
  Snackbar,
  Alert,
  Menu,
  MenuItem,
  ListItemIcon,
  ListItemText,
  useTheme,
} from '@wso2/oxygen-ui';
import {Building, EllipsisVertical, Pencil, Plus, Trash2} from '@wso2/oxygen-ui-icons-react';
import {useTranslation} from 'react-i18next';
import {useConfig} from '@thunder/shared-contexts';
import {useAsgardeo} from '@asgardeo/react';
import {useQueryClient} from '@tanstack/react-query';
import useGetOrganizationUnits from '../api/useGetOrganizationUnits';
import type {OrganizationUnit} from '../models/organization-unit';
import type {OrganizationUnitTreeItem} from '../models/organization-unit-tree';
import type {OrganizationUnitListResponse} from '../models/responses';
import OrganizationUnitQueryKeys from '../constants/organization-unit-query-keys';
import OrganizationUnitDeleteDialog from './OrganizationUnitDeleteDialog';
import useOrganizationUnit from '../contexts/useOrganizationUnit';

const PLACEHOLDER_SUFFIX = '__placeholder';
const ERROR_SUFFIX = '__error';
const ADD_CHILD_SUFFIX = '__addChild';
const LOAD_MORE_SUFFIX = '__loadMore';
const ROOT_PARENT_ID = '__root';
const ROOT_LOAD_MORE_ID = `${ROOT_PARENT_ID}${LOAD_MORE_SUFFIX}`;
const PAGE_SIZE = 30;

function TreeViewLoadingIcon(): JSX.Element {
  return <CircularProgress size={18} />;
}

function buildAddChildItem(parentId: string, parentName: string, parentHandle: string): OrganizationUnitTreeItem {
  return {
    id: `${parentId}${ADD_CHILD_SUFFIX}`,
    label: parentName,
    handle: parentHandle,
    isPlaceholder: true,
  };
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

interface CustomTreeItemProps extends TreeView.TreeItemProps {
  itemId: string;
  label?: ReactNode;
  onEdit?: (event: MouseEvent<HTMLElement>, ou: {id: string; name: string}) => void;
  onDelete?: (event: MouseEvent<HTMLElement>, ou: {id: string; name: string}) => void;
  onAddChild?: (event: MouseEvent<HTMLElement>, ou: {id: string; name: string; handle: string}) => void;
  onLoadMore?: (parentId: string) => void;
  addChildTooltip?: string;
  addChildButtonText?: string;
  editTooltip?: string;
  deleteTooltip?: string;
  loadingItems?: Set<string>;
  loadMoreLoadingItems?: Set<string>;
  itemMap?: Map<string, OrganizationUnitTreeItem>;
}

function findItem(items: OrganizationUnitTreeItem[], id: string): OrganizationUnitTreeItem | undefined {
  return items.reduce<OrganizationUnitTreeItem | undefined>((found, item) => {
    if (found) return found;
    if (item.id === id) return item;

    return item.children ? findItem(item.children, id) : undefined;
  }, undefined);
}

function CustomTreeItem(allProps: CustomTreeItemProps): JSX.Element {
  const {
    onEdit,
    onDelete,
    onAddChild,
    onLoadMore: onLoadMoreProp,
    addChildTooltip = '',
    addChildButtonText = '',
    editTooltip = '',
    deleteTooltip = '',
    loadingItems: loadingItemsProp,
    loadMoreLoadingItems: loadMoreLoadingItemsProp,
    itemMap: itemMapProp,
    itemId,
    label,
    ...restProps
  } = allProps;
  const treeItemProps = {itemId, label, ...restProps};
  const theme = useTheme();
  const {t} = useTranslation();
  const [menuAnchor, setMenuAnchor] = useState<null | HTMLElement>(null);
  const labelStr = typeof label === 'string' ? label : '';
  const itemData = itemMapProp?.get(itemId);
  const isLoadMoreItem = itemId.endsWith(LOAD_MORE_SUFFIX);
  const isAddChildButton = itemId.endsWith(ADD_CHILD_SUFFIX);
  const isPlaceholder =
    !isAddChildButton &&
    !isLoadMoreItem &&
    (itemData?.isPlaceholder ??
      (itemId.endsWith(PLACEHOLDER_SUFFIX) || itemId.endsWith(ERROR_SUFFIX) || itemId.endsWith('__empty')));
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
            borderRadius: 1,
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
            onClick={(e: MouseEvent<HTMLElement>) => {
              e.stopPropagation();
              if (!isLoadingMore) {
                onLoadMoreProp?.(parentId);
              }
            }}
            onKeyDown={(e: KeyboardEvent) => {
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

  if (isAddChildButton) {
    const parentId = itemId.replace(ADD_CHILD_SUFFIX, '');
    const parentItem = itemMapProp?.get(parentId);

    return (
      <TreeView.TreeItem
        {...treeItemProps}
        sx={{
          '& > .MuiTreeItem-content': {
            border: '1px dashed',
            borderColor: theme.vars?.palette.primary.main,
            borderRadius: 1,
            backgroundColor: 'transparent !important',
            cursor: 'pointer',
            transition: 'all 0.15s ease-in-out',
            '&:hover': {
              backgroundColor: `${theme.vars?.palette.primary.main} !important`,
              '& .add-child-avatar': {
                backgroundColor: theme.vars?.palette.primary.contrastText,
                color: theme.vars?.palette.primary.main,
              },
              '& .add-child-text': {
                color: theme.vars?.palette.primary.contrastText,
              },
            },
          },
        }}
        label={
          <Box
            role="button"
            tabIndex={0}
            onClick={(e: MouseEvent<HTMLElement>) => {
              e.stopPropagation();
              onAddChild?.(e, {id: parentId, name: parentItem?.label ?? '', handle: parentItem?.handle ?? ''});
            }}
            onKeyDown={(e: KeyboardEvent) => {
              if (e.key === 'Enter' || e.key === ' ') {
                e.preventDefault();
                e.stopPropagation();
                onAddChild?.(e as unknown as MouseEvent<HTMLElement>, {
                  id: parentId,
                  name: parentItem?.label ?? '',
                  handle: parentItem?.handle ?? '',
                });
              }
            }}
            sx={{
              display: 'flex',
              alignItems: 'center',
              gap: 1.5,
            }}
          >
            <Avatar
              className="add-child-avatar"
              sx={{
                p: 0.5,
                backgroundColor: theme.vars?.palette.primary.main,
                width: 32,
                height: 32,
                fontSize: '0.875rem',
                transition: 'all 0.15s ease-in-out',
              }}
            >
              <Plus size={14} />
            </Avatar>
            <Typography
              className="add-child-text"
              variant="body2"
              sx={{fontWeight: 500, transition: 'color 0.15s ease-in-out'}}
            >
              {addChildButtonText}
            </Typography>
          </Box>
        }
      />
    );
  }

  if (isPlaceholder) {
    const isLoadingPlaceholder = itemId.endsWith(PLACEHOLDER_SUFFIX);
    const isErrorPlaceholder = itemId.endsWith(ERROR_SUFFIX);

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
            {isLoadingPlaceholder ? (
              <>
                <CircularProgress size={16} />
                <Typography variant="caption" color="text.secondary" sx={{fontStyle: 'italic'}}>
                  Loading...
                </Typography>
              </>
            ) : (
              <Typography
                variant="caption"
                color={isErrorPlaceholder ? 'error' : 'text.secondary'}
                sx={{fontStyle: 'italic', pl: 1}}
              >
                {labelStr}
              </Typography>
            )}
          </Box>
        }
      />
    );
  }

  return (
    <>
      <TreeView.TreeItem
        {...treeItemProps}
        {...(isItemLoading
          ? {slots: {collapseIcon: TreeViewLoadingIcon, expandIcon: TreeViewLoadingIcon}}
          : {})}
        label={
          <Box
            sx={{
              display: 'flex',
              alignItems: 'center',
              gap: 1.5,
            }}
          >
            <Avatar
              sx={{
                p: 0.5,
                backgroundColor: theme?.vars?.palette.primary.main,
                width: 32,
                height: 32,
                fontSize: '0.875rem',
              }}
              src={itemData?.logo_url}
            >
              <Building size={14} />
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
            <IconButton
              size="small"
              aria-label="Actions"
              onClick={(e: MouseEvent<HTMLButtonElement>) => {
                e.stopPropagation();
                setMenuAnchor(e.currentTarget);
              }}
              sx={{
                color: theme.vars?.palette.text.secondary,
                '&:hover': {color: theme.vars?.palette.text.primary},
              }}
            >
              <EllipsisVertical size={16} />
            </IconButton>
          </Box>
        }
      />
      <Menu anchorEl={menuAnchor} open={Boolean(menuAnchor)} onClose={() => setMenuAnchor(null)}>
        <MenuItem
          onClick={(e: MouseEvent<HTMLLIElement>) => {
            e.stopPropagation();
            setMenuAnchor(null);
            onAddChild?.(e as unknown as MouseEvent<HTMLElement>, {
              id: itemId,
              name: labelStr,
              handle: itemData?.handle ?? '',
            });
          }}
        >
          <ListItemIcon>
            <Plus size={16} />
          </ListItemIcon>
          <ListItemText>{addChildTooltip}</ListItemText>
        </MenuItem>
        <MenuItem
          onClick={(e: MouseEvent<HTMLLIElement>) => {
            e.stopPropagation();
            setMenuAnchor(null);
            onEdit?.(e as unknown as MouseEvent<HTMLElement>, {id: itemId, name: labelStr});
          }}
        >
          <ListItemIcon>
            <Pencil size={16} />
          </ListItemIcon>
          <ListItemText>{editTooltip}</ListItemText>
        </MenuItem>
        <MenuItem
          onClick={(e: MouseEvent<HTMLLIElement>) => {
            e.stopPropagation();
            setMenuAnchor(null);
            onDelete?.(e as unknown as MouseEvent<HTMLElement>, {id: itemId, name: labelStr});
          }}
        >
          <ListItemIcon>
            <Trash2 size={16} color="red" />
          </ListItemIcon>
          <ListItemText sx={{color: 'error.main'}}>{deleteTooltip}</ListItemText>
        </MenuItem>
      </Menu>
    </>
  );
}

export default function OrganizationUnitsTreeView(): JSX.Element {
  const theme = useTheme();
  const navigate = useNavigate();
  const {t} = useTranslation();
  const logger = useLogger('OrganizationUnitsTreeView');
  const {http} = useAsgardeo();
  const {getServerUrl} = useConfig();
  const queryClient = useQueryClient();
  const {data, isLoading, error} = useGetOrganizationUnits();
  const {treeItems, setTreeItems, expandedItems, setExpandedItems, loadedItems, setLoadedItems, resetTreeState} =
    useOrganizationUnit();

  const itemMap = useMemo(() => buildItemMap(treeItems), [treeItems]);

  const [loadingItems, setLoadingItems] = useState<Set<string>>(new Set());
  const [loadMoreLoadingItems, setLoadMoreLoadingItems] = useState<Set<string>>(new Set());
  const [childOffsets, setChildOffsets] = useState<Map<string, number>>(new Map());
  const [rootOffset, setRootOffset] = useState<number>(0);
  const [rootLoadMoreLoading, setRootLoadMoreLoading] = useState<boolean>(false);
  const rootLoadMoreLoadingRef = useRef<boolean>(false);
  rootLoadMoreLoadingRef.current = rootLoadMoreLoading;
  const loadingItemsRef = useRef<Set<string>>(loadingItems);
  loadingItemsRef.current = loadingItems;
  const expandedItemsRef = useRef<string[]>(expandedItems);
  expandedItemsRef.current = expandedItems;
  const treeItemsRef = useRef<OrganizationUnitTreeItem[]>(treeItems);
  treeItemsRef.current = treeItems;
  const rebuildIdRef = useRef(0);
  const builtFromDataRef = useRef<unknown>(null);
  const [selectedOU, setSelectedOU] = useState<{id: string; name: string} | null>(null);
  const [deleteDialogOpen, setDeleteDialogOpen] = useState<boolean>(false);
  const [snackbar, setSnackbar] = useState<{open: boolean; message: string; severity: 'success' | 'error'}>({
    open: false,
    message: '',
    severity: 'success',
  });

  const updateTreeItemChildren = useCallback(
    (
      items: OrganizationUnitTreeItem[],
      parentId: string,
      children: OrganizationUnitTreeItem[],
    ): OrganizationUnitTreeItem[] =>
      items.map((item) => {
        if (item.id === parentId) {
          return {...item, children};
        }

        if (item.children && item.children.length > 0) {
          return {...item, children: updateTreeItemChildren(item.children, parentId, children)};
        }

        return item;
      }),
    [],
  );

  const appendTreeItemChildren = useCallback(
    (
      items: OrganizationUnitTreeItem[],
      parentId: string,
      newChildren: OrganizationUnitTreeItem[],
    ): OrganizationUnitTreeItem[] =>
      items.map((item) => {
        if (item.id === parentId) {
          const existing = (item.children ?? []).filter((c) => !c.id.endsWith(LOAD_MORE_SUFFIX));

          return {...item, children: [...existing, ...newChildren]};
        }

        if (item.children && item.children.length > 0) {
          return {...item, children: appendTreeItemChildren(item.children, parentId, newChildren)};
        }

        return item;
      }),
    [],
  );

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

  // Fetch children for a single parent and return the built tree items.
  // Does NOT update React state — caller is responsible for that.
  const fetchChildItems = useCallback(
    async (parentId: string): Promise<OrganizationUnitTreeItem[]> => {
      const result = await fetchChildPage(parentId, 0);
      const childOUs = result.organizationUnits;
      const parentItem = findItem(treeItemsRef.current, parentId);
      const addChildItem = buildAddChildItem(parentId, parentItem?.label ?? '', parentItem?.handle ?? '');
      const items = childOUs.length > 0 ? [addChildItem, ...buildTreeItems(childOUs)] : [addChildItem];

      if (childOUs.length < result.totalResults) {
        items.push({
          id: `${parentId}${LOAD_MORE_SUFFIX}`,
          label: '',
          handle: '',
          isPlaceholder: true,
        });
      }

      return items;
    },
    [fetchChildPage],
  );

  // Fetch and update state for user-triggered node expansion
  const fetchChildOUs = useCallback(
    async (parentId: string): Promise<void> => {
      if (loadingItemsRef.current.has(parentId)) return;

      setLoadingItems((prev) => new Set(prev).add(parentId));

      try {
        const childItems = await fetchChildItems(parentId);

        setChildOffsets((prev) => new Map(prev).set(parentId, childItems.filter(
          (c) => !c.isPlaceholder,
        ).length));
        // Update tree items, mark as loaded, then expand in one synchronous block.
        // The node stays collapsed until this point, so it opens directly with real children.
        setTreeItems((prev) => updateTreeItemChildren(prev, parentId, childItems));
        setLoadedItems((prev) => new Set(prev).add(parentId));
        setExpandedItems((prev) => (prev.includes(parentId) ? prev : [...prev, parentId]));
      } catch (_error: unknown) {
        logger.error('Failed to load child organization units', {error: _error, parentId});
        // Replace the loading placeholder with an error placeholder so the user sees feedback.
        // The node is NOT marked as loaded, so collapsing and re-expanding will retry the fetch.
        const errorItem: OrganizationUnitTreeItem = {
          id: `${parentId}${ERROR_SUFFIX}`,
          label: t('organizationUnits:listing.treeView.loadError'),
          handle: '',
          isPlaceholder: true,
        };
        setTreeItems((prev) => updateTreeItemChildren(prev, parentId, [errorItem]));
        setExpandedItems((prev) => (prev.includes(parentId) ? prev : [...prev, parentId]));
      } finally {
        setLoadingItems((prev) => {
          const next = new Set(prev);
          next.delete(parentId);

          return next;
        });
      }
    },
    [fetchChildItems, updateTreeItemChildren, setTreeItems, setLoadedItems, setExpandedItems, logger, t],
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
    [rootOffset, getServerUrl, queryClient, http, setTreeItems, logger],
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
        const childOUs = result.organizationUnits;
        const newItems = buildTreeItems(childOUs);
        const loadedSoFar = offset + childOUs.length;

        if (loadedSoFar < result.totalResults) {
          newItems.push({
            id: `${parentId}${LOAD_MORE_SUFFIX}`,
            label: '',
            handle: '',
            isPlaceholder: true,
          });
        }

        setChildOffsets((prev) => new Map(prev).set(parentId, loadedSoFar));
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
    [childOffsets, fetchChildPage, appendTreeItemChildren, setTreeItems, logger, handleRootLoadMore],
  );

  // Process one level of the tree: fetch children for the given IDs,
  // insert them into the tree, then recurse for the next deeper level.
  const expandLevel = useCallback(
    (
      tree: OrganizationUnitTreeItem[],
      levelIds: string[],
      expandedSet: Set<string>,
      loaded: Set<string>,
    ): Promise<{tree: OrganizationUnitTreeItem[]; loaded: Set<string>}> => {
      if (levelIds.length === 0) {
        return Promise.resolve({tree, loaded});
      }

      return Promise.all(
        levelIds.map((parentId) =>
          fetchChildItems(parentId)
            .then((children) => ({parentId, children, success: true as const}))
            .catch(() => ({parentId, children: [] as OrganizationUnitTreeItem[], success: false as const})),
        ),
      ).then((results) => {
        // Insert fetched children into the tree and collect next-level IDs
        let updatedTree = tree;
        const nextLoaded = new Set(loaded);
        const nextLevelIds: string[] = [];

        results
          .filter((r) => r.success)
          .forEach((r) => {
            updatedTree = updateTreeItemChildren(updatedTree, r.parentId, r.children);
            nextLoaded.add(r.parentId);

            r.children
              .filter((child) => !child.isPlaceholder && expandedSet.has(child.id))
              .forEach((child) => {
                nextLevelIds.push(child.id);
              });
          });

        // Recurse to the next level
        return expandLevel(updatedTree, nextLevelIds, expandedSet, nextLoaded);
      });
    },
    [fetchChildItems, updateTreeItemChildren],
  );

  // Build the full tree with all previously expanded nodes restored.
  // Returns the computed tree and loaded set without setting state — the caller
  // is responsible for applying the result so it can guard against stale rebuilds.
  const rebuildTree = useCallback(
    (
      rootOUs: OrganizationUnit[],
      expandedIds: string[],
    ): Promise<{tree: OrganizationUnitTreeItem[]; loaded: Set<string>}> => {
      const rootTree = buildTreeItems(rootOUs);
      const expandedSet = new Set(expandedIds);

      // Start with root-level IDs that are expanded
      const rootLevelIds = rootTree.map((item) => item.id).filter((id) => expandedSet.has(id));

      return expandLevel(rootTree, rootLevelIds, expandedSet, new Set<string>());
    },
    [expandLevel],
  );

  const buildRootTreeItems = useCallback(
    (response: OrganizationUnitListResponse): OrganizationUnitTreeItem[] => {
      const items = buildTreeItems(response.organizationUnits);

      if (response.organizationUnits.length < response.totalResults) {
        items.push({
          id: ROOT_LOAD_MORE_ID,
          label: '',
          handle: '',
          isPlaceholder: true,
        });
      }

      setRootOffset(response.organizationUnits.length);

      return items;
    },
    [],
  );

  // Rebuild tree when query data is available and either the tree is empty (after
  // reset) or the data has changed since the last build (fresh fetch after mutation).
  // rebuildIdRef guards against stale rebuilds: if a newer rebuild starts while an
  // older one is in-flight, the older result is silently ignored.
  useEffect(() => {
    if (!data?.organizationUnits || data.organizationUnits.length === 0) return;

    // Skip if tree is already built from this exact data reference
    if (treeItems.length > 0 && builtFromDataRef.current === data) return;

    const currentExpanded = expandedItemsRef.current;
    rebuildIdRef.current += 1;
    const id = rebuildIdRef.current;

    if (currentExpanded.length > 0) {
      // Rebuild with expanded nodes restored
      rebuildTree(data.organizationUnits, currentExpanded)
        .then(({tree, loaded}) => {
          // Only apply if no newer rebuild was triggered
          if (rebuildIdRef.current === id) {
            const items = tree;

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
            setLoadedItems(loaded);
            builtFromDataRef.current = data;
          }
        })
        .catch((_err: unknown) => {
          logger.error('Failed to rebuild tree with expanded items', {error: _err});
          if (rebuildIdRef.current === id) {
            // Fallback: just set root items
            setTreeItems(buildRootTreeItems(data));
            builtFromDataRef.current = data;
          }
        });
    } else {
      setTreeItems(buildRootTreeItems(data));
      builtFromDataRef.current = data;
    }
  }, [data, treeItems.length, rebuildTree, buildRootTreeItems, setTreeItems, setLoadedItems, logger]);

  // Clear builtFromDataRef when tree is reset so the effect rebuilds from current data
  useEffect(() => {
    if (treeItems.length === 0) {
      builtFromDataRef.current = null;
    }
  }, [treeItems.length]);

  const handleItemExpansionToggle = useCallback(
    (_event: SyntheticEvent | null, itemId: string, isExpanded: boolean) => {
      if (!isExpanded || loadedItems.has(itemId) || loadingItems.has(itemId)) {
        return;
      }

      // Don't expand yet — fetchChildOUs will expand after children are loaded.
      fetchChildOUs(itemId).catch((_error: unknown) => {
        logger.error('Failed to load child organization units', {error: _error, parentId: itemId});
      });
    },
    [loadedItems, loadingItems, fetchChildOUs, logger],
  );

  const handleEditClick = useCallback(
    (_event: MouseEvent<HTMLElement>, ou: {id: string; name: string}): void => {
      (async (): Promise<void> => {
        await navigate(`/organization-units/${ou.id}`);
      })().catch((_error: unknown) => {
        logger.error('Failed to navigate to organization unit', {error: _error, ouId: ou.id});
      });
    },
    [navigate, logger],
  );

  const handleDeleteClick = useCallback((_event: MouseEvent<HTMLElement>, ou: {id: string; name: string}): void => {
    setSelectedOU(ou);
    setDeleteDialogOpen(true);
  }, []);

  const handleDeleteDialogClose = (): void => {
    setDeleteDialogOpen(false);
    setSelectedOU(null);
  };

  const handleDeleteSuccess = useCallback((): void => {
    resetTreeState();
    setSnackbar({
      open: true,
      message: t('organizationUnits:edit.general.dangerZone.delete.success'),
      severity: 'success',
    });
  }, [resetTreeState, t]);

  const handleDeleteError = useCallback((message: string): void => {
    setSnackbar({open: true, message, severity: 'error'});
  }, []);

  const handleAddChildClick = useCallback(
    (_event: MouseEvent<HTMLElement>, ou: {id: string; name: string; handle: string}): void => {
      (async (): Promise<void> => {
        await navigate('/organization-units/create', {
          state: {parentId: ou.id, parentName: ou.name, parentHandle: ou.handle},
        });
      })().catch((_error: unknown) => {
        logger.error('Failed to navigate to create child organization unit', {error: _error, parentId: ou.id});
      });
    },
    [navigate, logger],
  );

  const handleAddRootClick = useCallback((): void => {
    (async (): Promise<void> => {
      await navigate('/organization-units/create');
    })().catch((_error: unknown) => {
      logger.error('Failed to navigate to create organization unit page', {error: _error});
    });
  }, [navigate, logger]);

  const combinedLoadMoreLoadingItems = useMemo(() => {
    if (!rootLoadMoreLoading) return loadMoreLoadingItems;
    const combined = new Set(loadMoreLoadingItems);
    combined.add(ROOT_PARENT_ID);

    return combined;
  }, [loadMoreLoadingItems, rootLoadMoreLoading]);

  if (error) {
    return (
      <Box sx={{textAlign: 'center', py: 8}}>
        <Typography variant="h6" color="error" gutterBottom>
          {t('organizationUnits:listing.error.title')}
        </Typography>
        <Typography variant="body2" color="text.secondary">
          {error.message ?? t('organizationUnits:listing.error.unknown')}
        </Typography>
      </Box>
    );
  }

  if (isLoading) {
    return (
      <Box sx={{display: 'flex', justifyContent: 'center', py: 8}}>
        <CircularProgress />
      </Box>
    );
  }

  if (!treeItems.length) {
    // Data loaded but no organization units exist — show empty state
    if (data && data.organizationUnits.length === 0) {
      return (
        <Box sx={{textAlign: 'center', py: 8}}>
          <Typography variant="body2" color="text.secondary" sx={{mb: 2}}>
            {t('organizationUnits:listing.treeView.empty')}
          </Typography>
          <Box
            role="button"
            tabIndex={0}
            onClick={handleAddRootClick}
            onKeyDown={(e: KeyboardEvent) => {
              if (e.key === 'Enter' || e.key === ' ') {
                e.preventDefault();
                handleAddRootClick();
              }
            }}
            sx={{
              display: 'inline-flex',
              alignItems: 'center',
              gap: 1.5,
              border: '1px dashed',
              borderColor: theme.vars?.palette.primary.main,
              borderRadius: 1,
              py: 1,
              px: 2,
              cursor: 'pointer',
              transition: 'all 0.15s ease-in-out',
              '&:hover': {
                backgroundColor: theme.vars?.palette.primary.main,
                '& .add-root-avatar': {
                  backgroundColor: theme.vars?.palette.primary.contrastText,
                  color: theme.vars?.palette.primary.main,
                },
                '& .add-root-text': {
                  color: theme.vars?.palette.primary.contrastText,
                },
              },
            }}
          >
            <Avatar
              className="add-root-avatar"
              sx={{
                p: 0.5,
                backgroundColor: theme.vars?.palette.primary.main,
                width: 32,
                height: 32,
                fontSize: '0.875rem',
                transition: 'all 0.15s ease-in-out',
              }}
            >
              <Plus size={14} />
            </Avatar>
            <Typography
              className="add-root-text"
              variant="body2"
              sx={{fontWeight: 500, transition: 'color 0.15s ease-in-out'}}
            >
              {t('organizationUnits:listing.addRootOrganizationUnit')}
            </Typography>
          </Box>
        </Box>
      );
    }

    // Still loading tree items
    return (
      <Box sx={{display: 'flex', justifyContent: 'center', py: 8}}>
        <CircularProgress />
      </Box>
    );
  }

  return (
    <>
      <Box sx={{width: '100%', minHeight: 400}}>
        <Box
          role="button"
          tabIndex={0}
          onClick={handleAddRootClick}
          onKeyDown={(e: KeyboardEvent) => {
            if (e.key === 'Enter' || e.key === ' ') {
              e.preventDefault();
              handleAddRootClick();
            }
          }}
          sx={{
            display: 'flex',
            alignItems: 'center',
            gap: 1.5,
            border: '1px dashed',
            borderColor: theme.vars?.palette.primary.main,
            borderRadius: 1,
            py: 1,
            pl: 5,
            pr: 1.5,
            mb: 0.75,
            cursor: 'pointer',
            transition: 'all 0.15s ease-in-out',
            '&:hover': {
              backgroundColor: theme.vars?.palette.primary.main,
              '& .add-root-avatar': {
                backgroundColor: theme.vars?.palette.primary.contrastText,
                color: theme.vars?.palette.primary.main,
              },
              '& .add-root-text': {
                color: theme.vars?.palette.primary.contrastText,
              },
            },
          }}
        >
          <Avatar
            className="add-root-avatar"
            sx={{
              p: 0.5,
              backgroundColor: theme.vars?.palette.primary.main,
              width: 32,
              height: 32,
              fontSize: '0.875rem',
              transition: 'all 0.15s ease-in-out',
            }}
          >
            <Plus size={14} />
          </Avatar>
          <Typography
            className="add-root-text"
            variant="body2"
            sx={{fontWeight: 500, transition: 'color 0.15s ease-in-out'}}
          >
            {t('organizationUnits:listing.addRootOrganizationUnit')}
          </Typography>
        </Box>
        <TreeView.RichTreeView
          items={treeItems}
          expandedItems={expandedItems}
          onExpandedItemsChange={(_event: SyntheticEvent | null, itemIds: string[]) => {
            // Block expansion of items whose children haven't been loaded yet.
            // fetchChildOUs will add them to expandedItems after children are fetched.
            const prevSet = new Set(expandedItems);
            const filtered = itemIds.filter((id) => prevSet.has(id) || loadedItems.has(id));
            setExpandedItems(filtered);
          }}
          onItemExpansionToggle={handleItemExpansionToggle}
          disableSelection
          slots={{item: CustomTreeItem}}
          slotProps={{
            item: {
              onEdit: handleEditClick,
              onDelete: handleDeleteClick,
              onAddChild: handleAddChildClick,
              onLoadMore: (parentId: string) => {
                handleLoadMore(parentId).catch((_error: unknown) => {
                  logger.error('Failed to load more child organization units', {error: _error, parentId});
                });
              },
              addChildTooltip: t('organizationUnits:listing.treeView.addChild'),
              addChildButtonText: t('organizationUnits:listing.treeView.addChildOrganizationUnit'),
              editTooltip: t('common:actions.edit'),
              deleteTooltip: t('common:actions.delete'),
              loadingItems,
              loadMoreLoadingItems: combinedLoadMoreLoadingItems,
              itemMap,
            } as Record<string, unknown>,
          }}
          getItemLabel={(item: OrganizationUnitTreeItem) => item.label}
          sx={{
            '& .MuiTreeItem-root': {
              position: 'relative',
            },
            '& .MuiTreeItem-content': {
              cursor: 'pointer',
              border: '1px solid',
              borderColor: theme.vars?.palette.divider,
              py: 1,
              px: 1.5,
              mb: 0.75,
              transition: 'all 0.15s ease-in-out',
              '&:hover': {
                backgroundColor: theme.vars?.palette.action.hover,
                borderColor: theme.vars?.palette.primary.main,
              },
            },
            '& .MuiTreeItem-iconContainer': {
              color: theme.vars?.palette.text.secondary,
              mr: 0.5,
            },
            // Hierarchy connector lines
            '& .MuiTreeItem-groupTransition': {
              ml: 3,
              pl: 3,
              borderLeft: '1px dashed',
              borderColor: theme.vars?.palette.divider,
            },
          }}
        />
      </Box>

      <OrganizationUnitDeleteDialog
        open={deleteDialogOpen}
        organizationUnitId={selectedOU?.id ?? null}
        onClose={handleDeleteDialogClose}
        onSuccess={handleDeleteSuccess}
        onError={handleDeleteError}
      />

      <Snackbar
        open={snackbar.open}
        autoHideDuration={6000}
        onClose={() => setSnackbar((prev) => ({...prev, open: false}))}
        anchorOrigin={{vertical: 'bottom', horizontal: 'right'}}
      >
        <Alert
          onClose={() => setSnackbar((prev) => ({...prev, open: false}))}
          severity={snackbar.severity}
          sx={{width: '100%'}}
        >
          {snackbar.message}
        </Alert>
      </Snackbar>
    </>
  );
}
