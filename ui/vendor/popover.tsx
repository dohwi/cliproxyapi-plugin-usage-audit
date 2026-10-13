import * as React from 'react';
import * as PopoverPrimitive from '@radix-ui/react-popover';
import { cn } from './utils';
import { useUIAppearance } from './appearance';

export const Popover = PopoverPrimitive.Root;
export const PopoverTrigger = PopoverPrimitive.Trigger;
export const PopoverClose = PopoverPrimitive.Close;
export const PopoverContent = React.forwardRef<React.ComponentRef<typeof PopoverPrimitive.Content>, React.ComponentPropsWithoutRef<typeof PopoverPrimitive.Content>>(
  ({ className, sideOffset = 6, align = 'start', collisionPadding = 8, style, ...props }, ref) => {
    const appearance = useUIAppearance();
    return <PopoverPrimitive.Portal>
      <PopoverPrimitive.Content {...appearance} ref={ref} sideOffset={sideOffset} align={align} collisionPadding={collisionPadding} className={cn('ui-popover-content', className)} {...props} style={{ ...appearance?.style, ...style }} />
    </PopoverPrimitive.Portal>;
  },
);
PopoverContent.displayName = 'PopoverContent';
