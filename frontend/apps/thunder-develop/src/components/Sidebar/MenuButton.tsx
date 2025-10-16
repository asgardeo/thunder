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

import Badge, {badgeClasses} from '@mui/material/Badge';
import IconButton, {type IconButtonProps} from '@mui/material/IconButton';

export interface MenuButtonProps extends IconButtonProps {
  showBadge?: boolean;
}

export default function MenuButton({showBadge = false, ...props}: MenuButtonProps) {
  return (
    <Badge color="error" variant="dot" invisible={!showBadge} sx={{[`& .${badgeClasses.badge}`]: {right: 2, top: 2}}}>
      <IconButton size="small" {...props} />
    </Badge>
  );
}
