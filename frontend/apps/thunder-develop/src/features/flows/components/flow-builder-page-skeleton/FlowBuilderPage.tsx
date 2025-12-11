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

import {useEffect, type HTMLAttributes} from 'react';
import './FlowBuilderPage.scss';
import useUserPreferences from '@/features/common/hooks/useUserPreferences';
import classNames from 'classnames';
import type {FloatingPublishButtonProps} from './FloatingPublishButton';
import FlowBuilderPageHeader from './FlowBuilderPageHeader';
import FloatingPublishButton from './FloatingPublishButton';

/**
 * Props interface of {@link FlowBuilderPage}
 */
export type FlowBuilderPageProps = HTMLAttributes<HTMLDivElement> & FloatingPublishButtonProps;

/**
 * Skeleton for the flow builder page.
 *
 * @param props - Props injected to the component.
 * @returns FlowBuilderPage component.
 */
function FlowBuilderPage({
  children,
  className,
  flowType,
  flowTypeDisplayName,
  isPublishing,
  onPublish,
}: FlowBuilderPageProps) {
  const {setPreferences, leftNavbarCollapsed} = useUserPreferences();

  /**
   * If the user doesn't have a `leftNavbarCollapsed` preference saved, collapse the navbar for better UX.
   * @remarks Since the builder needs more real estate, it's better to have the navbar collapsed.
   */
  useEffect(() => {
    if (leftNavbarCollapsed === undefined) {
      setPreferences({leftNavbarCollapsed: true});
    }
  }, [leftNavbarCollapsed, setPreferences]);

  return (
    <div className={classNames('flow-builder-page', className)}>
      <div className="page-layout">
        <FlowBuilderPageHeader
          flowType={flowType}
          isPublishing={isPublishing}
          onPublish={onPublish}
          flowTypeDisplayName={flowTypeDisplayName}
        />
      </div>
      {children}
      <FloatingPublishButton
        flowType={flowType}
        isPublishing={isPublishing}
        onPublish={onPublish}
        flowTypeDisplayName={flowTypeDisplayName}
      />
    </div>
  );
}

export default FlowBuilderPage;
