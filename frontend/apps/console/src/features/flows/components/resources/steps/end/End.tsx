/**
 * Copyright (c) 2023-2025, WSO2 LLC. (https://www.wso2.com).
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

import type {ReactElement} from 'react';
import {Handle, Position} from '@xyflow/react';
import VisualFlowConstants from '@/features/flows/constants/VisualFlowConstants';
import {Fab} from '@wso2/oxygen-ui';
import type {CommonStepFactoryPropsInterface} from '../CommonStepFactory';
import './End.scss';

/**
 * Props interface of {@link End}
 */
export type EndPropsInterface = CommonStepFactoryPropsInterface;

/**
 * End Node component.
 * This is a custom node supported by react flow renderer library.
 * See {@link https://reactflow.dev/docs/api/node-types/} for its documentation
 * and {@link https://reactflow.dev/examples/custom-node/} for an example
 *
 * @param _props - Props injected to the component (unused).
 * @returns End node component.
 */
// eslint-disable-next-line @typescript-eslint/no-unused-vars
function End(_props: EndPropsInterface): ReactElement {
  return (
    <div>
      <Handle
        className="hidden-handle"
        id={`end${VisualFlowConstants.FLOW_BUILDER_PREVIOUS_HANDLE_SUFFIX}`}
        type="target"
        position={Position.Left}
      />
      <Fab aria-label="end" className="end" variant="extended" size="small">
        End
      </Fab>
    </div>
  );
}

export default End;
