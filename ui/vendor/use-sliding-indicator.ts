import * as React from 'react';

/** Measures an active child without changing the container's scroll position. */
export function useSlidingIndicator(
  ref: React.RefObject<HTMLElement | null>,
  activeSelector: string,
  dependency?: unknown,
): { style: React.CSSProperties | undefined; ready: boolean } {
  const [style, setStyle] = React.useState<React.CSSProperties>();

  React.useLayoutEffect(() => {
    const current = ref.current;
    if (!current) {
      setStyle(undefined);
      return;
    }
    const container: HTMLElement = current;

    const view = container.ownerDocument.defaultView;
    if (!view) return;
    let disposed = false;
    let frame = 0;
    let selected: HTMLElement | null = null;
    const observed = new Set<Element>();

    const measure = () => {
      if (disposed) return;
      selected = container.querySelector<HTMLElement>(activeSelector);
      syncObservedChildren();
      const containerRect = container.getBoundingClientRect();
      const selectedRect = selected?.getBoundingClientRect();
      if (!container.isConnected || !containerRect.width || !containerRect.height ||
        !selectedRect?.width || !selectedRect.height || !selected?.getClientRects().length) {
        setStyle((previous) => previous === undefined ? previous : undefined);
        return;
      }
      const transform = `translate(${selectedRect.left - containerRect.left + container.scrollLeft - container.clientLeft}px, ${selectedRect.top - containerRect.top + container.scrollTop - container.clientTop}px)`;
      setStyle((previous) => {
        if (previous?.width === selectedRect.width && previous.height === selectedRect.height && previous.transform === transform) return previous;
        return { position: 'absolute', top: 0, left: 0, width: selectedRect.width, height: selectedRect.height, transform };
      });
    };
    const schedule = () => {
      if (disposed || frame) return;
      frame = view.requestAnimationFrame(() => {
        frame = 0;
        measure();
      });
    };
    const resizeObserver = typeof ResizeObserver === 'undefined' ? undefined : new ResizeObserver(schedule);
    function syncObservedChildren() {
      const next = new Set<Element>([container, ...Array.from(container.children).filter((child) => !child.classList.contains('ui-tabs-indicator') && !child.classList.contains('ui-segmented-indicator') && !child.classList.contains('navigation-indicator'))]);
      if (selected) next.add(selected);
      for (const element of observed) {
        if (!next.has(element)) {
          resizeObserver?.unobserve(element);
          observed.delete(element);
        }
      }
      for (const element of next) {
        if (!observed.has(element)) {
          resizeObserver?.observe(element);
          observed.add(element);
        }
      }
    }
    const mutationObserver = new MutationObserver(schedule);
    mutationObserver.observe(container, {
      subtree: true,
      childList: true,
      attributes: true,
      attributeFilter: ['class', 'data-state', 'aria-current'],
    });
    view.addEventListener('resize', schedule);
    container.addEventListener('scroll', schedule, { passive: true });
    void container.ownerDocument.fonts?.ready.then(schedule);
    // The first geometry is committed before paint; subsequent observer work is batched.
    measure();
    return () => {
      disposed = true;
      view.cancelAnimationFrame(frame);
      resizeObserver?.disconnect();
      mutationObserver.disconnect();
      view.removeEventListener('resize', schedule);
      container.removeEventListener('scroll', schedule);
    };
  }, [ref, activeSelector, dependency]);

  return { style, ready: style !== undefined };
}
