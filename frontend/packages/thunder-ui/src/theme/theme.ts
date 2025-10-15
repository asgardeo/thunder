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

import {createTheme} from '@mui/material/styles';
import inputsCustomizations from './customizations/inputs';
import dataDisplayCustomizations from './customizations/dataDisplay';
import surfacesCustomizations from './customizations/surfaces';
import feedbackCustomizations from './customizations/feedback';
import navigationCustomizations from './customizations/navigation';
import {colorSchemes, typography, shadows, shape} from './themePrimitives';

const theme = createTheme({
  // For more details about CSS variables configuration, see https://mui.com/material-ui/customization/css-theme-variables/configuration/
  cssVariables: {
    colorSchemeSelector: 'data-mui-color-scheme',
    cssVarPrefix: 'thunder',
  },
  // Recently added in v6 for building light & dark mode app, see https://mui.com/material-ui/customization/palette/#color-schemes
  colorSchemes,
  typography,
  shadows,
  shape,
  components: {
    ...inputsCustomizations,
    ...dataDisplayCustomizations,
    ...feedbackCustomizations,
    ...navigationCustomizations,
    ...surfacesCustomizations,
  },
});

export default theme;
