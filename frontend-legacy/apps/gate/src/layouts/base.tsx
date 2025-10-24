/**
 * Copyright (c) 2025, WSO2 LLC. (https://www.wso2.com).
 *
 * WSO2 LLC. licenses this file to you under the Apache License,
 * Version 2.0 (the "License"); you may not use this file except
 * in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */

import ThemeProvider from '@oxygen-ui/react/src/contexts/ThemeProvider';
import { InitColorSchemeScript } from '@oxygen-ui/react/src/components/InitColorSchemeScript/InitColorSchemeScript';
import { Geist, Geist_Mono } from 'next/font/google';
import { Metadata } from 'next';
import { NextFontWithVariable } from 'next/dist/compiled/@next/font';
import { ReactElement } from 'react';
import '../globals.css';

export const metadata: Metadata = {
  icons: {
    icon: '/favicon.ico',
    shortcut: '/favicon.ico',
  },
};

const geistSans: NextFontWithVariable = Geist({
  variable: '--font-geist-sans',
  subsets: ['latin'],
});

const geistMono: NextFontWithVariable = Geist_Mono({
  variable: '--font-geist-mono',
  subsets: ['latin'],
});

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>): ReactElement {
  return (
    <html lang="en" suppressHydrationWarning>
      <body className={`${geistSans.variable} ${geistMono.variable}`}>
        <InitColorSchemeScript defaultMode="system" modeStorageKey="mui-mode" attribute="data-color-scheme" />
        <div className="body-background-design-overlay"></div>
        <ThemeProvider>{children}</ThemeProvider>
      </body>
    </html>
  );
}
