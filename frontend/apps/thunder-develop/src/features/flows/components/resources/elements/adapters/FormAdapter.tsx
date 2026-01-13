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

import {useMemo, type ReactElement} from 'react';
import {useTranslation} from 'react-i18next';
import classNames from 'classnames';
import {Badge, Box, Typography} from '@wso2/oxygen-ui';
import {CollisionPriority} from '@dnd-kit/abstract';
import VisualFlowConstants from '@/features/flows/constants/VisualFlowConstants';
import {ElementCategories, type Element as FlowElement} from '@/features/flows/models/elements';
import FlowEventTypes from '@/features/flows/models/extension';
import PluginRegistry from '@/features/flows/plugins/PluginRegistry';
import generateResourceId from '@/features/flows/utils/generateResourceId';
import ReorderableFlowElement from '../../steps/view/ReorderableElement';
import Droppable from '../../../dnd/Droppable';
import './FormAdapter.scss';

/**
 * Form element type.
 */
export type FormElement = FlowElement;

/**
 * Props interface of {@link FormAdapter}
 */
export interface FormAdapterPropsInterface {
  /**
   * The form element properties.
   */
  resource: FormElement;
  /**
   * The step id the resource resides on.
   */
  stepId: string;
  /**
   * List of available elements that can be added to the form.
   */
  availableElements?: FlowElement[];
  /**
   * Callback for adding an element to the form.
   * @param element - The element to add.
   * @param formId - The ID of the form to add to.
   */
  onAddElementToForm?: (element: FlowElement, formId: string) => void;
}

/**
 * Adapter for the Form component.
 *
 * @param props - Props injected to the component.
 * @returns The FormAdapter component.
 */
function FormAdapter({
  resource,
  stepId,
  availableElements = [],
  onAddElementToForm = undefined,
}: FormAdapterPropsInterface): ReactElement {
  const {t} = useTranslation();

  const shouldShowFormFieldsPlaceholder = !resource?.components?.some(
    (element: FlowElement) => element.category === ElementCategories.Field,
  );

  const filteredComponents = useMemo(() => {
    if (!resource?.components) return [];
    return resource.components.filter((component: FlowElement) =>
      PluginRegistry.getInstance().executeSync(FlowEventTypes.ON_NODE_ELEMENT_FILTER, component),
    );
  }, [resource?.components]);

  return (
    <Badge
      anchorOrigin={{
        horizontal: 'left',
        vertical: 'top',
      }}
      badgeContent={t('flows:core.adapters.form.badgeLabel')}
      className="adapter form-adapter"
    >
      <Box>
        <Droppable
          id={generateResourceId(`${VisualFlowConstants.FLOW_BUILDER_FORM_ID}_${stepId}`)}
          data={{droppedOn: resource, stepId}}
          collisionPriority={CollisionPriority.High}
          type={VisualFlowConstants.FLOW_BUILDER_DROPPABLE_FORM_ID}
          accept={[
            VisualFlowConstants.FLOW_BUILDER_DRAGGABLE_ID,
            ...VisualFlowConstants.FLOW_BUILDER_FORM_ALLOWED_RESOURCE_TYPES,
          ]}
        >
          {shouldShowFormFieldsPlaceholder && (
            <Box className="form-adapter-placeholder">
              <Typography variant="body2">{t('flows:core.adapters.form.placeholder')}</Typography>
            </Box>
          )}
          {filteredComponents.map((component: FlowElement, index: number) => (
            <ReorderableFlowElement
              key={component.id}
              id={component.id}
              index={index}
              element={component}
              className={classNames('flow-builder-step-content-form-field')}
              group={resource.id}
              type={VisualFlowConstants.FLOW_BUILDER_DRAGGABLE_ID}
              accept={[
                VisualFlowConstants.FLOW_BUILDER_DRAGGABLE_ID,
                ...VisualFlowConstants.FLOW_BUILDER_FORM_ALLOWED_RESOURCE_TYPES,
              ]}
              availableElements={availableElements}
              onAddElementToForm={onAddElementToForm}
            />
          ))}
        </Droppable>
      </Box>
    </Badge>
  );
}

export default FormAdapter;
