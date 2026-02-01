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

import {describe, it, expect} from 'vitest';
import {
  StepCategories,
  StepTypes,
  StaticStepTypes,
  ExecutionTypes,
  ExecutionStepViewTypes,
  EdgeStyleTypes,
} from '../steps';

describe('steps models', () => {
  describe('StepCategories', () => {
    it('should have Decision category', () => {
      expect(StepCategories.Decision).toBe('DECISION');
    });

    it('should have Interface category', () => {
      expect(StepCategories.Interface).toBe('INTERFACE');
    });

    it('should have Workflow category', () => {
      expect(StepCategories.Workflow).toBe('WORKFLOW');
    });

    it('should have Executor category', () => {
      expect(StepCategories.Executor).toBe('EXECUTOR');
    });

    it('should have exactly 4 categories', () => {
      expect(Object.keys(StepCategories)).toHaveLength(4);
    });
  });

  describe('StepTypes', () => {
    it('should have View type', () => {
      expect(StepTypes.View).toBe('VIEW');
    });

    it('should have Rule type', () => {
      expect(StepTypes.Rule).toBe('RULE');
    });

    it('should have Execution type', () => {
      expect(StepTypes.Execution).toBe('TASK_EXECUTION');
    });

    it('should have End type', () => {
      expect(StepTypes.End).toBe('END');
    });

    it('should have exactly 4 step types', () => {
      expect(Object.keys(StepTypes)).toHaveLength(4);
    });
  });

  describe('StaticStepTypes', () => {
    it('should have UserOnboard type', () => {
      expect(StaticStepTypes.UserOnboard).toBe('USER_ONBOARD');
    });

    it('should have Start type', () => {
      expect(StaticStepTypes.Start).toBe('START');
    });

    it('should have exactly 2 static step types', () => {
      expect(Object.keys(StaticStepTypes)).toHaveLength(2);
    });
  });

  describe('ExecutionTypes', () => {
    it('should have GoogleFederation type', () => {
      expect(ExecutionTypes.GoogleFederation).toBe('GoogleOIDCAuthExecutor');
    });

    it('should have AppleFederation type', () => {
      expect(ExecutionTypes.AppleFederation).toBe('AppleExecutor');
    });

    it('should have FacebookFederation type', () => {
      expect(ExecutionTypes.FacebookFederation).toBe('FacebookExecutor');
    });

    it('should have MicrosoftFederation type', () => {
      expect(ExecutionTypes.MicrosoftFederation).toBe('Office365Executor');
    });

    it('should have GithubFederation type', () => {
      expect(ExecutionTypes.GithubFederation).toBe('GithubOAuthExecutor');
    });

    it('should have ConfirmationCode type', () => {
      expect(ExecutionTypes.ConfirmationCode).toBe('ConfirmationCodeValidationExecutor');
    });

    it('should have MagicLinkExecutor type', () => {
      expect(ExecutionTypes.MagicLinkExecutor).toBe('MagicLinkExecutor');
    });

    it('should have SendEmailOTP type', () => {
      expect(ExecutionTypes.SendEmailOTP).toBe('SendEmailOTPExecutor');
    });

    it('should have VerifyEmailOTP type', () => {
      expect(ExecutionTypes.VerifyEmailOTP).toBe('VerifyEmailOTPExecutor');
    });

    it('should have SMSOTPAuth type', () => {
      expect(ExecutionTypes.SMSOTPAuth).toBe('SMSOTPAuthExecutor');
    });

    it('should have exactly 11 execution types', () => {
      expect(Object.keys(ExecutionTypes)).toHaveLength(11);
    });
  });

  describe('ExecutionStepViewTypes', () => {
    it('should have Default view type', () => {
      expect(ExecutionStepViewTypes.Default).toBe('Execution');
    });

    it('should have MagicLinkView type', () => {
      expect(ExecutionStepViewTypes.MagicLinkView).toBe('Magic Link View');
    });

    it('should have PasskeyView type', () => {
      expect(ExecutionStepViewTypes.PasskeyView).toBe('Passkey View');
    });

    it('should have exactly 3 view types', () => {
      expect(Object.keys(ExecutionStepViewTypes)).toHaveLength(3);
    });
  });

  describe('EdgeStyleTypes', () => {
    it('should have Bezier style', () => {
      expect(EdgeStyleTypes.Bezier).toBe('default');
    });

    it('should have SmoothStep style', () => {
      expect(EdgeStyleTypes.SmoothStep).toBe('smoothstep');
    });

    it('should have Step style', () => {
      expect(EdgeStyleTypes.Step).toBe('step');
    });

    it('should have exactly 3 edge styles', () => {
      expect(Object.keys(EdgeStyleTypes)).toHaveLength(3);
    });
  });
});
