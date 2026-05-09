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

import type {Plugin} from 'unified';

interface Options {
  productName: string;
  productSlug: string;
}

interface HastNode {
  type: string;
  tagName?: string;
  value?: string;
  children?: HastNode[];
}

function replaceInTextNode(node: HastNode, productName: string, productSlug: string): void {
  if (node.type === 'text' && typeof node.value === 'string') {
    node.value = node.value.replaceAll('{{ProductName}}', productName).replaceAll('{{productSlug}}', productSlug);
  }
}

/**
 * Recursively visits every `code` element in the hast tree — both fenced code
 * blocks (`pre > code`) and inline code spans — and replaces all occurrences of
 * `{{ProductName}}` and `{{productSlug}}` in text nodes with the resolved product
 * name and slug respectively.
 */
function replaceInCodeNodes(node: HastNode, productName: string, productSlug: string): void {
  if (node.type === 'element' && node.tagName === 'code' && node.children) {
    for (const child of node.children) {
      replaceInTextNode(child, productName, productSlug);
    }
    // Don't recurse into code children — text nodes already handled above.
    return;
  }

  if (node.children) {
    for (const child of node.children) {
      replaceInCodeNodes(child, productName, productSlug);
    }
  }
}

const rehypeProductName: Plugin<[Options]> = function (options: Options) {
  const {productName, productSlug} = options;

  return (tree: unknown) => {
    replaceInCodeNodes(tree as HastNode, productName, productSlug);
  };
};

export default rehypeProductName;
