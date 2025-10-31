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

import type {SourceCode, Rule} from 'eslint';
import {Comment} from 'estree';

interface CopyrightHeaderOptions {
  excludePatterns?: string[];
  template?: string;
}

const REQUIRED_COPYRIGHT_HEADER = `/**
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
 */`;

const copyrightHeaderRule: Rule.RuleModule = {
  meta: {
    type: 'layout',
    docs: {
      description: 'Enforce WSO2 Apache 2.0 copyright header in all source files',
      category: 'Stylistic Issues',
    },
    fixable: 'code',
    schema: [
      {
        type: 'object',
        properties: {
          excludePatterns: {
            type: 'array',
            items: {type: 'string'},
          },
          template: {
            type: 'string',
          },
        },
        additionalProperties: false,
      },
    ],
    messages: {
      missingHeader: 'Missing WSO2 Apache 2.0 copyright header',
      incorrectHeader: 'Incorrect copyright header format',
    },
  },
  create(context: Rule.RuleContext) {
    const options: CopyrightHeaderOptions = (context.options?.[0] as CopyrightHeaderOptions) ?? {};
    const excludePatterns: string[] = options.excludePatterns ?? [];
    const template: string = options.template ?? REQUIRED_COPYRIGHT_HEADER;
    const filename: string = context.getFilename();

    // Check if file should be excluded
    if (excludePatterns.some((pattern: string) => new RegExp(pattern).test(filename))) {
      return {};
    }

    // Skip certain file types
    if (/\.(json|md|yml|yaml|xml|txt)$/.exec(filename)) {
      return {};
    }

    return {
      Program(node: unknown) {
        const sourceCode: SourceCode = context.getSourceCode();
        const comments: Comment[] = sourceCode.getAllComments();

        // Check if first comment is the copyright header
        const firstComment: Comment | undefined = comments[0];

        if (!firstComment || firstComment.type !== 'Block') {
          context.report({
            // @ts-expect-error TODO: Update to the latest ESLint and remove `@types/eslint`.
            node,
            messageId: 'missingHeader',
            fix(fixer: Rule.RuleFixer) {
              // @ts-expect-error TODO: Update to the latest ESLint and remove `@types/eslint`.
              return fixer.insertTextBefore(node, `${template}\n\n`);
            },
          });
          return;
        }

        // Normalize the comment text for comparison
        const commentText = `/*${firstComment.value}*/`;
        const normalizedComment: string = commentText.replace(/\s+/g, ' ').trim();

        if (!normalizedComment.includes('WSO2 LLC') || !normalizedComment.includes('Apache License')) {
          context.report({
            // @ts-expect-error TODO: Update to the latest ESLint and remove `@types/eslint`.
            node: firstComment,
            messageId: 'incorrectHeader',
            fix(fixer: Rule.RuleFixer) {
              // @ts-expect-error TODO: Update to the latest ESLint and remove `@types/eslint`.
              return fixer.replaceText(firstComment, template);
            },
          });
        }
      },
    };
  },
};

export default copyrightHeaderRule;
