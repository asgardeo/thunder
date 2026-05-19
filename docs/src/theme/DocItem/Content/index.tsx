import {useDoc} from '@docusaurus/plugin-content-docs/client';
import GettingStartedJourney, {
  getGettingStartedStepIndex,
} from '@site/src/components/GettingStartedJourney';
import Heading from '@theme/Heading';
import MDXContent from '@theme/MDXContent';
import React, {type ReactNode, useEffect, useRef} from 'react';
// @ts-ignore — JS module, no types
import CopyPageButton from 'docusaurus-plugin-copy-page-button/src/CopyPageButton';

function useSyntheticTitle(): string | null {
  const {metadata, frontMatter, contentTitle} = useDoc();
  const shouldRender =
    !frontMatter.hide_title && typeof contentTitle === 'undefined';
  if (!shouldRender) {
    return null;
  }
  return metadata.title;
}

export default function DocItemContent({children}: {children: ReactNode}): ReactNode {
  const syntheticTitle = useSyntheticTitle();
  const {metadata, frontMatter} = useDoc();
  const currentJourneyStep = getGettingStartedStepIndex(metadata.id);
  const isHomePage = metadata.id === 'index';
  const showButton = !isHomePage && !frontMatter.hide_title;
  const containerRef = useRef<HTMLDivElement | null>(null);
  const journeyContainerRef = useRef<HTMLDivElement | null>(null);

  useEffect(() => {
    if (currentJourneyStep === null) {
      return;
    }

    const container = containerRef.current;
    const journeyContainer = journeyContainerRef.current;

    if (!container || !journeyContainer) {
      return;
    }

    const title = container.querySelector('h1');

    if (!title) {
      return;
    }

    const titleBlock = title.closest('header') ?? title;

    if (titleBlock.parentElement === container) {
      titleBlock.insertAdjacentElement('afterend', journeyContainer);
    }
  }, [currentJourneyStep, metadata.id]);

  return (
    <div ref={containerRef} className="theme-doc-markdown markdown doc-content-with-copy-btn">
      {syntheticTitle && (
        <header>
          <Heading as="h1">{syntheticTitle}</Heading>
        </header>
      )}
      {currentJourneyStep !== null && (
        <div ref={journeyContainerRef}>
          <GettingStartedJourney current={currentJourneyStep} />
        </div>
      )}
      {showButton && (
        <div className="copy-page-btn-wrapper">
          <CopyPageButton
            enabledActions={['copy', 'view', 'chatgpt', 'claude', 'gemini']}
          />
        </div>
      )}
      <MDXContent>{children}</MDXContent>
    </div>
  );
}
