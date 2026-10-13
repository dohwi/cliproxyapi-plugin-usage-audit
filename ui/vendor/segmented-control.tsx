import * as React from 'react';
import * as RadioGroupPrimitive from '@radix-ui/react-radio-group';
import { cn } from './utils';
import { useSlidingIndicator } from './use-sliding-indicator';

export interface SegmentedControlProps extends React.ComponentPropsWithoutRef<typeof RadioGroupPrimitive.Root> {
  size?: 'sm' | 'default' | 'lg';
  /** Stretch the track to the container and share its width between items. */
  fullWidth?: boolean;
}

/** Single-choice radio group whose selected box slides to the measured item. */
export const SegmentedControl = React.forwardRef<HTMLDivElement, SegmentedControlProps>(
  ({ className, children, size = 'default', fullWidth = false, ...props }, ref) => {
    const localRef = React.useRef<HTMLDivElement>(null);
    React.useImperativeHandle(ref, () => localRef.current as HTMLDivElement, []);
    const { style, ready } = useSlidingIndicator(localRef, '.ui-segmented-item[data-state="checked"]', props.value);
    const [animate, setAnimate] = React.useState(false);
    React.useLayoutEffect(() => {
      if (!ready) {
        setAnimate(false);
        return;
      }
      const view = localRef.current?.ownerDocument.defaultView;
      if (!view) return;
      let secondFrame = 0;
      // The first position is placed instantly; only later moves are animated.
      const firstFrame = view.requestAnimationFrame(() => {
        secondFrame = view.requestAnimationFrame(() => setAnimate(true));
      });
      return () => {
        view.cancelAnimationFrame(firstFrame);
        view.cancelAnimationFrame(secondFrame);
      };
    }, [ready]);
    return (
      <RadioGroupPrimitive.Root
        {...props}
        ref={localRef}
        className={cn('ui-segmented', `ui-segmented--${size}`, fullWidth && 'ui-segmented--full', className)}
        data-indicator-ready={ready ? 'true' : undefined}
        data-indicator-animate={ready && animate ? 'true' : undefined}
      >
        <span className="ui-segmented-indicator" aria-hidden="true" style={style} />
        {children}
      </RadioGroupPrimitive.Root>
    );
  },
);
SegmentedControl.displayName = 'SegmentedControl';

export const SegmentedControlItem = React.forwardRef<React.ComponentRef<typeof RadioGroupPrimitive.Item>, React.ComponentPropsWithoutRef<typeof RadioGroupPrimitive.Item>>(
  ({ className, ...props }, ref) => <RadioGroupPrimitive.Item ref={ref} className={cn('ui-segmented-item', className)} {...props} />,
);
SegmentedControlItem.displayName = 'SegmentedControlItem';
