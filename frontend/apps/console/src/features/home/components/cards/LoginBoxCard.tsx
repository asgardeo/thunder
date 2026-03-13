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

import {Box, Stack} from '@wso2/oxygen-ui';
import {Palette} from '@wso2/oxygen-ui-icons-react';
import {motion} from 'framer-motion';
import {useTranslation} from 'react-i18next';
import type {JSX} from 'react';
import HomeNextStepCard from './HomeNextStepCard';

const SWATCH_COLORS = ['#FF6B00', '#6366F1', '#0EA5E9', '#10B981'];

export default function LoginBoxCard(): JSX.Element {
  const {t} = useTranslation('home');

  const preview = (
    <Stack direction="row" spacing={0.75} alignItems="center">
      {SWATCH_COLORS.map((color, i) => (
        <motion.div
          key={color}
          initial={{opacity: 0, scaleX: 0}}
          animate={{opacity: 1, scaleX: 1}}
          transition={{duration: 0.35, delay: i * 0.07, ease: 'easeOut'}}
          style={{transformOrigin: 'left'}}
        >
          <Box
            sx={{
              width: 28,
              height: 12,
              borderRadius: 1,
              bgcolor: color,
            }}
          />
        </motion.div>
      ))}
    </Stack>
  );

  return (
    <HomeNextStepCard
      icon={<Palette size={24} />}
      title={t('next_steps.login_box.title', 'Sign-in Box')}
      description={t(
        'next_steps.login_box.description',
        'Build themes and attach them to your applications to personalise the sign-in experience.',
      )}
      primaryLabel={t('next_steps.login_box.actions.primary.label', 'Open Design Studio')}
      primaryRoute="/design"
      preview={preview}
    />
  );
}
