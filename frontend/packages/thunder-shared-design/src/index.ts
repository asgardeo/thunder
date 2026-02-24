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

// Theme API hooks
export {default as useGetThemes} from './api/useGetThemes';
export {default as useGetTheme} from './api/useGetTheme';
export {default as useCreateTheme} from './api/useCreateTheme';
export {default as useUpdateTheme} from './api/useUpdateTheme';
export {default as useDeleteTheme} from './api/useDeleteTheme';

// Layout API hooks
export {default as useGetLayouts} from './api/useGetLayouts';
export {default as useGetLayout} from './api/useGetLayout';
export {default as useCreateLayout} from './api/useCreateLayout';
export {default as useUpdateLayout} from './api/useUpdateLayout';
export {default as useDeleteLayout} from './api/useDeleteLayout';

// Design resolve API hook
export {default as useGetDesignResolve} from './api/useGetDesignResolve';

// Query keys
export {default as DesignQueryKeys} from './constants/design-query-keys';

// Context
export {default as DesignContext} from './contexts/Design/DesignContext';
export * from './contexts/Design/DesignContext';

export {default as DesignProvider} from './contexts/Design/DesignProvider';
export * from './contexts/Design/DesignProvider';

export {default as useDesign} from './contexts/Design/useDesign';

// Models
export * from './models/design';
export * from './models/layout';
export * from './models/requests';
export * from './models/responses';
export * from './models/theme';

// Utils
export {default as extractLayoutFromDesign} from './utils/extractLayoutFromDesign';
export {default as mapEmbeddedFlowTextVariant} from './utils/mapEmbeddedFlowTextVariant';
