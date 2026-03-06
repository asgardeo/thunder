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

import {describe, it, expect} from 'vitest';
import {TranslationCreateFlowStep} from '../translation-create-flow';

describe('translation-create-flow models', () => {
  describe('TranslationCreateFlowStep', () => {
    it('should have COUNTRY step', () => {
      expect(TranslationCreateFlowStep.COUNTRY).toBe('COUNTRY');
    });

    it('should have LANGUAGE step', () => {
      expect(TranslationCreateFlowStep.LANGUAGE).toBe('LANGUAGE');
    });

    it('should have LOCALE_CODE step', () => {
      expect(TranslationCreateFlowStep.LOCALE_CODE).toBe('LOCALE_CODE');
    });

    it('should have INITIALIZE step', () => {
      expect(TranslationCreateFlowStep.INITIALIZE).toBe('INITIALIZE');
    });

    it('should have exactly 4 steps', () => {
      expect(Object.keys(TranslationCreateFlowStep)).toHaveLength(4);
    });

    it('should be usable as a record key', () => {
      const labels: Record<TranslationCreateFlowStep, string> = {
        COUNTRY: 'Select Country',
        LANGUAGE: 'Select Language',
        LOCALE_CODE: 'Review Locale Code',
        INITIALIZE: 'Initialize Language',
      };

      expect(labels[TranslationCreateFlowStep.COUNTRY]).toBe('Select Country');
      expect(labels[TranslationCreateFlowStep.LANGUAGE]).toBe('Select Language');
      expect(labels[TranslationCreateFlowStep.LOCALE_CODE]).toBe('Review Locale Code');
      expect(labels[TranslationCreateFlowStep.INITIALIZE]).toBe('Initialize Language');
    });

    it('step values should match their keys', () => {
      Object.entries(TranslationCreateFlowStep).forEach(([key, value]) => {
        expect(value).toBe(key);
      });
    });
  });
});
