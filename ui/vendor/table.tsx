import * as React from 'react';
import { cn } from './utils';

/** Optional keyboard-focusable scroll region; keeps the Table ref native. */
export const TableContainer = React.forwardRef<HTMLDivElement, React.ComponentPropsWithoutRef<'div'>>(
  ({ className, tabIndex = 0, role = 'region', ...props }, ref) => <div ref={ref} tabIndex={tabIndex} role={role} className={cn('ui-table-container', className)} {...props} />,
);
TableContainer.displayName = 'TableContainer';

export const Table = React.forwardRef<HTMLTableElement, React.ComponentPropsWithoutRef<'table'>>(
  ({ className, ...props }, ref) => <table ref={ref} className={cn('ui-table', className)} {...props} />,
);
Table.displayName = 'Table';
export const TableHeader = React.forwardRef<HTMLTableSectionElement, React.ComponentPropsWithoutRef<'thead'>>(
  ({ className, ...props }, ref) => <thead ref={ref} className={cn('ui-table-header', className)} {...props} />,
);
TableHeader.displayName = 'TableHeader';
export const TableBody = React.forwardRef<HTMLTableSectionElement, React.ComponentPropsWithoutRef<'tbody'>>(
  ({ className, ...props }, ref) => <tbody ref={ref} className={cn('ui-table-body', className)} {...props} />,
);
TableBody.displayName = 'TableBody';
export const TableRow = React.forwardRef<HTMLTableRowElement, React.ComponentPropsWithoutRef<'tr'>>(
  ({ className, ...props }, ref) => <tr ref={ref} className={cn('ui-table-row', className)} {...props} />,
);
TableRow.displayName = 'TableRow';
export const TableHead = React.forwardRef<HTMLTableCellElement, React.ComponentPropsWithoutRef<'th'>>(
  ({ className, ...props }, ref) => <th ref={ref} className={cn('ui-table-head', className)} {...props} />,
);
TableHead.displayName = 'TableHead';
export const TableCell = React.forwardRef<HTMLTableCellElement, React.ComponentPropsWithoutRef<'td'>>(
  ({ className, ...props }, ref) => <td ref={ref} className={cn('ui-table-cell', className)} {...props} />,
);
TableCell.displayName = 'TableCell';
export const TableCaption = React.forwardRef<HTMLTableCaptionElement, React.ComponentPropsWithoutRef<'caption'>>(
  ({ className, ...props }, ref) => <caption ref={ref} className={cn('ui-table-caption', className)} {...props} />,
);
TableCaption.displayName = 'TableCaption';
