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

import { writeFileSync } from 'fs';
import { join, dirname } from 'path';
import { fileURLToPath } from 'url';
import { createLogger } from '@thunder/logger';

const __filename = fileURLToPath(import.meta.url);
const __dirname = dirname(__filename);

const OUTPUT_FILE = join(__dirname, '..', 'content', 'releases.md');
const GITHUB_REPO = 'asgardeo/thunder';
const GITHUB_API_URL = `https://api.github.com/repos/${GITHUB_REPO}/releases`;

const logger = createLogger('generate-changelog');

// Cache for user avatars to avoid repeated API calls.
const userAvatarCache = {};

async function fetchReleases() {
    logger.info(`Fetching releases from ${GITHUB_API_URL}...`);
    try {
        const response = await fetch(GITHUB_API_URL, {
            headers: {
                'User-Agent': 'Thunder-Docs-Changelog-Generator',
                // Add Authorization header if GITHUB_TOKEN is present to avoid rate limits.
                ...(process.env.GITHUB_TOKEN ? { Authorization: `token ${process.env.GITHUB_TOKEN}` } : {})
            }
        });

        if (!response.ok) {
            throw new Error(`Failed to fetch releases: ${response.status} ${response.statusText}`);
        }

        const releases = await response.json();
        return releases;
    } catch (error) {
        logger.error('Error fetching releases:', error);
        return [];
    }
}

function sanitizeReleaseBody(body, releaseTag) {
    // Convert HTML img tags to Markdown img syntax and remove p tags.
    // This prevents MDX compilation errors with mixed HTML and Markdown.
    let sanitized = body
        .replace(/<p\s+align="left">\s*<img\s+src="([^"]+)"\s+alt="([^"]*)"\s+width="([^"]+)">\s*<\/p>/g, '![$2]($1)')
        .replace(/<p\s+align="left">/g, '')
        .replace(/<\/p>/g, '')
        .replace(/<img\s+src="([^"]+)"\s+alt="([^"]*)"\s+width="([^"]+)">/g, '![$2]($1)');

    // Convert relative image paths to absolute GitHub URLs.
    // Matches markdown image syntax: ![alt](path).
    sanitized = sanitized.replace(/!\[([^\]]*)\]\(([^)]+)\)/g, (match, alt, src) => {
        // If the path is relative (doesn't start with http), convert to GitHub raw content URL.
        if (!src.startsWith('http')) {
            // Use the tag_name from the release to construct the correct GitHub URL.
            const githubRawUrl = `https://raw.githubusercontent.com/asgardeo/thunder/${releaseTag}/${src}`;
            return `![${alt}](${githubRawUrl})`;
        }
        return match;
    });

    // Escape angle brackets that look like template variables (e.g., <os>, <arch>, <version>).
    // to prevent MDX parser from treating them as JSX tags.
    // This regex matches <word> patterns that are not part of valid HTML tags.
    sanitized = sanitized.replace(/<([a-zA-Z_-]+)>/g, (match, word) => {
        // List of actual HTML tags to preserve
        const htmlTags = ['div', 'img', 'p', 'span', 'a', 'strong', 'em', 'br', 'hr'];
        if (!htmlTags.includes(word.toLowerCase())) {
            // Replace with HTML entities to prevent MDX parser issues.
            return `&lt;${word}&gt;`;
        }
        return match;
    });

    return sanitized;
}

async function fetchUserAvatar(username) {
    // Return cached avatar if available.
    if (userAvatarCache[username]) {
        return userAvatarCache[username];
    }

    try {
        const response = await fetch(`https://api.github.com/users/${username}`, {
            headers: {
                'User-Agent': 'Thunder-Docs-Changelog-Generator',
                ...(process.env.GITHUB_TOKEN ? { Authorization: `token ${process.env.GITHUB_TOKEN}` } : {})
            }
        });

        if (response.ok) {
            const data = await response.json();
            userAvatarCache[username] = {
                avatar: data.avatar_url,
                url: data.html_url
            };
            return userAvatarCache[username];
        }
    } catch (error) {
        logger.warn(`Failed to fetch avatar for ${username}: ${error.message}`);
    }

    return null;
}

function extractContributorsFromBody(body) {
    // Extract unique usernames mentioned with @ symbol.
    const mentions = body.match(/@([a-zA-Z0-9-]+)/g) || [];
    const uniqueContributors = [...new Set(mentions.map(m => m.substring(1)))];

    // Filter out common non-user mentions and known non-user references.
    return uniqueContributors.filter(username =>
        !['dependabot', 'renovate', 'asgardeo', 'thunder', 'example', 'wso2', '123'].includes(username.toLowerCase())
    );
}

function cleanChangeText(text) {
    // Remove "by @username in PR_LINK" pattern.
    // Pattern: " by @<username> in https://github.com/...".
    return text.replace(/\s+by\s+@[\w-]+\s+in\s+https:\/\/github\.com\/\S+/, '').trim();
}

function extractNewContributorUsernames(body) {
    // Extract usernames from "New Contributors" section.
    // Pattern: "* @username made their first contribution...".
    const newContributorsMatch = body.match(/New Contributors[\s\S]*?(?=###|$)/);
    if (!newContributorsMatch) {
        return [];
    }

    const mentions = (newContributorsMatch[0] || '').match(/@([a-zA-Z0-9-]+)/g) || [];
    const usernames = [...new Set(mentions.map(m => m.substring(1)))];

    // Filter out common non-user mentions and known non-user references.
    return usernames.filter(username =>
        !['dependabot', 'renovate', 'asgardeo', 'thunder', 'example', 'wso2', '123'].includes(username.toLowerCase())
    );
}

function extractChanges(body) {
    // Extract and categorize changes from the release body.
    const categories = {
        breaking: [],
        features: [],
        improvements: [],
        bugs: [],
        contributors: []
    };

    const lines = body.split('\n');
    let currentCategory = null;

    for (const line of lines) {
        const trimmed = line.trim();

        if (trimmed.includes('Breaking Changes') || trimmed.includes('⚠️')) {
            currentCategory = 'breaking';
        } else if (trimmed.includes('Features') || trimmed.includes('🚀')) {
            currentCategory = 'features';
        } else if (trimmed.includes('Improvements') || trimmed.includes('✨')) {
            currentCategory = 'improvements';
        } else if (trimmed.includes('Bug Fixes') || trimmed.includes('🐛')) {
            currentCategory = 'bugs';
        } else if (trimmed.includes('New Contributors') || trimmed.includes('Contributors')) {
            currentCategory = 'contributors';
        } else if (trimmed.startsWith('*') && currentCategory) {
            const cleanedText = cleanChangeText(trimmed.substring(1).trim());

            // Filter out unwanted lines.
            const isFullChangelog = cleanedText.toLowerCase().includes('full changelog');
            const isNoteAboutSampleApp = cleanedText.toLowerCase().includes('note the id of the sample app');
            const isEmptyOrLink = !cleanedText || cleanedText.startsWith('http');

            if (cleanedText && !isFullChangelog && !isNoteAboutSampleApp && !isEmptyOrLink) {
                categories[currentCategory].push(cleanedText);
            }
        }
    }

    return categories;
}

async function formatChangelog(releases) {
    let content = '# Releases\n\n';
    content += 'Explore the latest updates, features, and improvements to **Thunder**.\n\n';
    content += '---\n\n';

    if (releases.length === 0) {
        content += 'No release notes found.\n';
        return content;
    }

    for (const release of releases) {
        if (release.draft) continue;

        const date = new Date(release.published_at).toLocaleDateString('en-US', {
            year: 'numeric',
            month: 'long',
            day: 'numeric'
        });
        const versionName = release.tag_name; // Use tag_name for shorter TOC entries.

        // Visual Header for the Release.
        content += `## ${versionName} {#${release.tag_name.toLowerCase().replace(/\./g, '-')}}\n`;
        content += `<p style={{ fontSize: '0.9rem', color: '#666', marginTop: '-10px', marginBottom: '20px' }}>\n`;
        content += `  <span>📅 Published on <strong>${date}</strong></span> • \n`;
        content += `  <a href="${release.html_url}" target="_blank" style={{ textDecoration: 'none' }}>View on GitHub ↗</a>\n`;
        content += `</p>\n\n`;

        const sanitizedBody = sanitizeReleaseBody(release.body, release.tag_name);
        const changes = extractChanges(sanitizedBody);

        // Render Contributors with a better layout.
        const contributors = extractContributorsFromBody(sanitizedBody);
        if (contributors.length > 0) {
            content += `<div style={{ padding: '15px', marginBottom: '10px',}}>\n`;
            content += `<span style={{ fontSize: '0.85rem', fontWeight: 'bold', display: 'block', marginBottom: '10px', color: '#555', textTransform: 'uppercase' }}>Contributors</span>\n`;
            content += `<div style={{ display: 'flex', flexWrap: 'wrap', gap: '8px' }}>\n`;

            for (const username of contributors) {
                const userInfo = await fetchUserAvatar(username);
                if (userInfo) {
                    content += `  <a href="${userInfo.url}" title="${username}" target="_blank" style={{ transition: 'transform 0.2s' }}>\n`;
                    content += `    <img src="${userInfo.avatar}" width="55" height="55" style={{ borderRadius: '50%', border: '2px solid #ff8c00', boxShadow: '0 2px 4px rgba(0,0,0,0.1)' }} alt="${username}" />\n`;
                    content += `  </a>\n`;
                }
            }
            content += `</div></div>\n\n`;
        }

        // Render New Contributors section.
        const newContributors = extractNewContributorUsernames(sanitizedBody);
        if (newContributors.length > 0) {
            content += `<div style={{ padding: '15px', marginBottom: '40px',}}>\n`;
            content += `<span style={{ fontSize: '0.85rem', fontWeight: 'bold', display: 'block', marginBottom: '10px', color: '#555', textTransform: 'uppercase' }}>New Contributors</span>\n`;
            content += `<div style={{ display: 'flex', flexWrap: 'wrap', gap: '8px' }}>\n`;

            for (const username of newContributors) {
                const userInfo = await fetchUserAvatar(username);
                if (userInfo) {
                    content += `  <a href="${userInfo.url}" title="${username}" target="_blank" style={{ transition: 'transform 0.2s' }}>\n`;
                    content += `    <img src="${userInfo.avatar}" width="55" height="55" style={{ borderRadius: '50%', border: '2px solid #ff8c00', boxShadow: '0 2px 4px rgba(0,0,0,0.1)' }} alt="${username}" />\n`;
                    content += `  </a>\n`;
                }
            }
            content += `</div></div>\n\n`;
        }

        // Categorized Changes with cleaner list styling.
        const renderCategory = (title, items, icon) => {
            if (items.length === 0) return '';
            let section = `#### ${icon} ${title}\n\n`;
            section += `<div style={{ borderLeft: '3px solid #ff8c00', paddingLeft: '5px', marginLeft: '8px', marginBottom: '24px' }}>\n`;
            items.forEach(item => {
                section += `- ${item}\n`;
            });
            section += `</div>\n\n`;
            return section;
        };

        content += renderCategory('Breaking Changes', changes.breaking, '⚠️');
        content += renderCategory('Features', changes.features, '🚀');
        content += renderCategory('Improvements', changes.improvements, '✨');
        content += renderCategory('Bug Fixes', changes.bugs, '🐛');

        content += `\n<hr style={{ border: 'none', borderTop: '1px solid #eaeaea', margin: '40px 0' }} />\n\n`;
    }

    return content;
}

async function generate() {
    try {
        const releases = await fetchReleases();
        const markdown = await formatChangelog(releases);
        writeFileSync(OUTPUT_FILE, markdown, 'utf8');
        logger.info(`✅ Changelog generated at ${OUTPUT_FILE}`);
    } catch (error) {
        logger.error('❌ Failed to generate changelog:', error);
        process.exit(1);
    }
}

generate();
