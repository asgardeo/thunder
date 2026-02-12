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

import MDXComponents from '@theme-original/MDXComponents';
import {Box, Card, CardContent, Typography} from '@wso2/oxygen-ui';
import Stepper from '@site/src/components/Stepper';
import TutorialHero, {TutorialHeroItem} from '@site/src/components/TutorialHero';
import SDKCard from '@site/src/components/SDKCard';
import ReactLogo from '@site/src/components/icons/ReactLogo';
import NextLogo from '@site/src/components/icons/NextLogo';
import VueLogo from '@site/src/components/icons/VueLogo';
import NuxtLogo from '@site/src/components/icons/NuxtLogo';
import AngularLogo from '@site/src/components/icons/AngularLogo';
import BrowserLogo from '@site/src/components/icons/BrowserLogo';
import NodeLogo from '@site/src/components/icons/NodeLogo';
import ExpressLogo from '@site/src/components/icons/ExpressLogo';
import GoLogo from '@site/src/components/icons/GoLogo';
import PythonLogo from '@site/src/components/icons/PythonLogo';
import FlutterLogo from '@site/src/components/icons/FlutterLogo';
import IOSLogo from '@site/src/components/icons/IOSLogo';
import AndroidLogo from '@site/src/components/icons/AndroidLogo';
import ReactRouterLogo from '@site/src/components/icons/ReactRouterLogo';

export default {
  ...MDXComponents,
  Box,
  Card,
  CardContent,
  Typography,
  Stepper,
  TutorialHero,
  TutorialHeroItem,
  SDKCard,
  ReactLogo,
  NextLogo,
  VueLogo,
  NuxtLogo,
  AngularLogo,
  BrowserLogo,
  NodeLogo,
  ExpressLogo,
  GoLogo,
  PythonLogo,
  FlutterLogo,
  IOSLogo,
  AndroidLogo,
  ReactRouterLogo,
  // TODO: Heading styling is a bit off when oxygen-ui Typography is used.
  // After sorting that out, we can switch to using Oxygen UI Typography for headings as well.
  // ex: h1: (props: TypographyProps<'h1'>) => <Typography variant="h1" {...props} />,
};
