import { createContext, useContext, type ComponentPropsWithoutRef, type CSSProperties, type ReactNode } from 'react';

/** Optional scoped appearance. Portaled content inherits the same context. */
export type UIAppearance = {
  theme: 'light' | 'dark';
  motion: 'on' | 'off';
  motionStyle?: 'standard' | 'material';
  style?: CSSProperties;
};
const AppearanceContext = createContext<UIAppearance | undefined>(undefined);

export function UIAppearanceProvider({ appearance, children }: { appearance?: UIAppearance; children: ReactNode }) {
  return <AppearanceContext.Provider value={appearance}>{children}</AppearanceContext.Provider>;
}

export function useUIAppearance() {
  const appearance = useContext(AppearanceContext);
  return appearance ? {
    'data-ui-appearance': '',
    'data-theme': appearance.theme,
    'data-motion': appearance.motion,
    'data-motion-style': appearance.motionStyle ?? 'standard',
    style: appearance.style,
  } : undefined;
}

/** Place tokens on actual preview surfaces, not their surrounding controls. */
export function UIAppearanceScope({ style, ...props }: ComponentPropsWithoutRef<'div'>) {
  const appearance = useUIAppearance();
  return <div {...appearance} {...props} style={{ ...appearance?.style, ...style }} />;
}
