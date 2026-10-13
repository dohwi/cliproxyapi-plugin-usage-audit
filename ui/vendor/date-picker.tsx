import * as React from 'react';
import { CalendarDays } from 'lucide-react';
import { cn } from './utils';
import { Popover, PopoverTrigger, PopoverContent } from './popover';
import { Calendar, formatDateKey, type CalendarBaseProps, type DateRange } from './calendar';

/** A finished range: both ends are chosen and `from <= to`. */
export interface CompleteDateRange { from: Date; to: Date }

export interface DatePickerBaseProps extends Omit<React.ComponentPropsWithoutRef<'button'>, 'value' | 'defaultValue' | 'onChange' | 'children'>,
  Pick<CalendarBaseProps, 'min' | 'max' | 'isDateDisabled' | 'weekStartsOn' | 'locale' | 'previousMonthLabel' | 'nextMonthLabel' | 'rangeHint'> {
  open?: boolean;
  defaultOpen?: boolean;
  onOpenChange?: (open: boolean) => void;
  placeholder?: string;
  /** Accessible name of the calendar dialog. */
  calendarLabel?: string;
  /** Single: submits YYYY-MM-DD. Range: submits `${name}-start` and `${name}-end`. */
  name?: string;
}

export interface DatePickerSingleProps extends DatePickerBaseProps {
  mode?: 'single';
  value?: Date | null;
  defaultValue?: Date | null;
  onValueChange?: (date: Date) => void;
}

export interface DatePickerRangeProps extends DatePickerBaseProps {
  mode: 'range';
  value?: DateRange | null;
  defaultValue?: DateRange | null;
  /** Called once, when both ends are chosen. Picking only the start does not call it. */
  onValueChange?: (range: CompleteDateRange) => void;
}

export type DatePickerProps = DatePickerSingleProps | DatePickerRangeProps;

type DatePickerImplProps = DatePickerBaseProps & {
  mode?: 'single' | 'range';
  value?: Date | DateRange | null;
  defaultValue?: Date | DateRange | null;
  onValueChange?: (value: never) => void;
};

function formatRange(range: DateRange, locale: string) {
  if (!range.from || !range.to) return undefined;
  const format = new Intl.DateTimeFormat(locale, { dateStyle: 'long' });
  return typeof format.formatRange === 'function' ? format.formatRange(range.from, range.to) : `${format.format(range.from)} – ${format.format(range.to)}`;
}

const DatePickerImpl = React.forwardRef<HTMLButtonElement, DatePickerImplProps>(
  ({
    className, mode = 'single', value, defaultValue, onValueChange, open, defaultOpen = false, onOpenChange, placeholder,
    calendarLabel, name, min, max, isDateDisabled, weekStartsOn, locale = 'ko-KR', previousMonthLabel, nextMonthLabel, rangeHint, disabled, ...props
  }, ref) => {
    const isRange = mode === 'range';
    const [innerValue, setInnerValue] = React.useState<Date | DateRange | null>(defaultValue ?? null);
    const [innerOpen, setInnerOpen] = React.useState(defaultOpen);
    // Start of an unfinished range lives only while the popup is open.
    const [draft, setDraft] = React.useState<DateRange | null>(null);
    const current = value !== undefined ? value : innerValue;
    const isOpen = open !== undefined ? open : innerOpen;
    const emit = onValueChange as ((next: Date | CompleteDateRange) => void) | undefined;

    const single = isRange ? null : (current as Date | null);
    const committed = isRange ? (current as DateRange | null) : null;
    const label = isRange
      ? (committed ? formatRange(committed, locale) : undefined)
      : (single ? new Intl.DateTimeFormat(locale, { dateStyle: 'long' }).format(single) : undefined);
    const emptyText = placeholder ?? (isRange ? '기간 선택' : '날짜 선택');
    const dialogLabel = calendarLabel ?? (isRange ? '기간 선택' : '날짜 선택');

    function setOpen(next: boolean) {
      if (!next) setDraft(null);
      if (open === undefined) setInnerOpen(next);
      onOpenChange?.(next);
    }
    const calendarShared = { min, max, isDateDisabled, weekStartsOn, locale, previousMonthLabel, nextMonthLabel, rangeHint, 'aria-label': dialogLabel };

    return (
      <Popover open={isOpen} onOpenChange={setOpen}>
        <PopoverTrigger asChild>
          <button {...props} ref={ref} type="button" disabled={disabled} className={cn('ui-date-picker-trigger', className)} data-placeholder={label ? undefined : ''} data-mode={isRange ? 'range' : undefined}>
            <span>{label ?? emptyText}</span>
            <CalendarDays size={16} aria-hidden="true" />
          </button>
        </PopoverTrigger>
        <PopoverContent
          className="ui-date-picker-content"
          aria-label={dialogLabel}
          onOpenAutoFocus={event => {
            // Start on the date, not on the month navigation buttons.
            event.preventDefault();
            (event.currentTarget as HTMLElement).querySelector<HTMLElement>('.ui-calendar-day[tabindex="0"]')?.focus();
          }}
        >
          {isRange ? (
            <Calendar
              {...calendarShared}
              mode="range"
              value={draft ?? committed}
              onValueChange={next => {
                if (!next.from || !next.to) { setDraft(next); return; }
                setDraft(null);
                if (value === undefined) setInnerValue(next);
                emit?.({ from: next.from, to: next.to });
                setOpen(false);
              }}
            />
          ) : (
            <Calendar
              {...calendarShared}
              value={single}
              onValueChange={date => {
                if (value === undefined) setInnerValue(date);
                emit?.(date);
                setOpen(false);
              }}
            />
          )}
        </PopoverContent>
        {name && (isRange ? <>
          <input type="hidden" name={`${name}-start`} value={committed?.from ? formatDateKey(committed.from) : ''} disabled={disabled} />
          <input type="hidden" name={`${name}-end`} value={committed?.to ? formatDateKey(committed.to) : ''} disabled={disabled} />
        </> : <input type="hidden" name={name} value={single ? formatDateKey(single) : ''} disabled={disabled} />)}
      </Popover>
    );
  },
);
DatePickerImpl.displayName = 'DatePicker';

/** Button trigger that opens a Calendar in a Popover. `mode="range"` picks a start and end date. */
export const DatePicker = DatePickerImpl as unknown as React.ForwardRefExoticComponent<DatePickerProps & React.RefAttributes<HTMLButtonElement>>;
