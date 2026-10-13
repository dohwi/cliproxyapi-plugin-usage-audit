import * as React from 'react';
import { ChevronLeft, ChevronRight } from 'lucide-react';
import { cn } from './utils';

type WeekStart = 0 | 1 | 2 | 3 | 4 | 5 | 6;

export function startOfDay(date: Date) { return new Date(date.getFullYear(), date.getMonth(), date.getDate()); }
function startOfMonth(date: Date) { return new Date(date.getFullYear(), date.getMonth(), 1); }
function addDays(date: Date, amount: number) { return new Date(date.getFullYear(), date.getMonth(), date.getDate() + amount); }
function addMonths(date: Date, amount: number) {
  const target = new Date(date.getFullYear(), date.getMonth() + amount, 1);
  const lastDay = new Date(target.getFullYear(), target.getMonth() + 1, 0).getDate();
  return new Date(target.getFullYear(), target.getMonth(), Math.min(date.getDate(), lastDay));
}
export function isSameDay(a: Date, b: Date) {
  return a.getFullYear() === b.getFullYear() && a.getMonth() === b.getMonth() && a.getDate() === b.getDate();
}
/** Local calendar date as YYYY-MM-DD, never shifted through UTC. */
export function formatDateKey(date: Date) {
  const pad = (value: number) => String(value).padStart(2, '0');
  return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())}`;
}

/** Inclusive date range. `to` is null while only the start has been chosen. */
export interface DateRange { from: Date | null; to: Date | null }

const EMPTY_RANGE: DateRange = { from: null, to: null };
function ordered(a: Date, b: Date): [Date, Date] { return a <= b ? [a, b] : [b, a]; }

export interface CalendarBaseProps extends Omit<React.ComponentPropsWithoutRef<'div'>, 'defaultValue' | 'onChange'> {
  month?: Date;
  defaultMonth?: Date;
  onMonthChange?: (month: Date) => void;
  min?: Date;
  max?: Date;
  isDateDisabled?: (date: Date) => boolean;
  weekStartsOn?: WeekStart;
  locale?: string;
  disabled?: boolean;
  previousMonthLabel?: string;
  nextMonthLabel?: string;
  /** Announced (role=status) after the start of a range is chosen. */
  rangeHint?: string;
}

export interface CalendarSingleProps extends CalendarBaseProps {
  mode?: 'single';
  value?: Date | null;
  defaultValue?: Date | null;
  onValueChange?: (date: Date) => void;
}

export interface CalendarRangeProps extends CalendarBaseProps {
  mode: 'range';
  value?: DateRange | null;
  defaultValue?: DateRange | null;
  /** Called after the first click with `to: null`, and again when the range is complete. */
  onValueChange?: (range: DateRange) => void;
}

export type CalendarProps = CalendarSingleProps | CalendarRangeProps;

type CalendarImplProps = CalendarBaseProps & {
  mode?: 'single' | 'range';
  value?: Date | DateRange | null;
  defaultValue?: Date | DateRange | null;
  onValueChange?: (value: Date | DateRange) => void;
};

/** Month grid for a single date or an inclusive range, with roving focus and full keyboard navigation. */
const CalendarImpl = React.forwardRef<HTMLDivElement, CalendarImplProps>(
  ({
    className, mode = 'single', value, defaultValue, onValueChange, rangeHint = '종료일을 선택하세요', month, defaultMonth, onMonthChange, min, max, isDateDisabled,
    weekStartsOn = 0, locale = 'ko-KR', disabled = false, previousMonthLabel = '이전 달', nextMonthLabel = '다음 달',
    'aria-label': ariaLabel = '달력', ...props
  }, ref) => {
    const rootRef = React.useRef<HTMLDivElement>(null);
    React.useImperativeHandle(ref, () => rootRef.current as HTMLDivElement, []);
    const captionId = React.useId();
    const today = React.useMemo(() => startOfDay(new Date()), []);
    const isRange = mode === 'range';
    const [innerValue, setInnerValue] = React.useState<Date | DateRange | null>(defaultValue ?? (isRange ? EMPTY_RANGE : null));
    const current = value !== undefined ? value : innerValue;
    const rawRange = isRange ? (current as DateRange | null) : null;
    const range: DateRange = { from: rawRange?.from ? startOfDay(rawRange.from) : null, to: rawRange?.to ? startOfDay(rawRange.to) : null };
    const selected: Date | null = isRange ? range.from : (current as Date | null);
    const [innerMonth, setInnerMonth] = React.useState(() => startOfMonth(defaultMonth ?? selected ?? today));
    const visibleMonth = month !== undefined ? startOfMonth(month) : innerMonth;
    const [focused, setFocused] = React.useState<Date | null>(null);
    const [hovered, setHovered] = React.useState<Date | null>(null);
    const pendingFocus = React.useRef(false);
    const pendingEnd = isRange && range.from !== null && range.to === null;
    const previewEnd = pendingEnd ? hovered ?? focused : null;
    const [bandLo, bandHi] = isRange && range.from ? ordered(range.from, range.to ?? previewEnd ?? range.from) : [null, null];
    const hasBand = bandLo !== null && bandHi !== null && !isSameDay(bandLo, bandHi);

    const formats = React.useMemo(() => ({
      caption: new Intl.DateTimeFormat(locale, { year: 'numeric', month: 'long' }),
      short: new Intl.DateTimeFormat(locale, { weekday: 'short' }),
      long: new Intl.DateTimeFormat(locale, { weekday: 'long' }),
      full: new Intl.DateTimeFormat(locale, { dateStyle: 'full' }),
      day: new Intl.NumberFormat(locale),
    }), [locale]);

    const lowest = min ? startOfDay(min) : undefined;
    const highest = max ? startOfDay(max) : undefined;
    const unavailable = (date: Date) => disabled || Boolean(lowest && date < lowest) || Boolean(highest && date > highest) || Boolean(isDateDisabled?.(date));
    const inVisibleMonth = (date: Date) => date.getFullYear() === visibleMonth.getFullYear() && date.getMonth() === visibleMonth.getMonth();
    const tabStop = focused && inVisibleMonth(focused) ? focused : selected && inVisibleMonth(selected) ? selected : inVisibleMonth(today) ? today : visibleMonth;

    const offset = (visibleMonth.getDay() - weekStartsOn + 7) % 7;
    const daysInMonth = new Date(visibleMonth.getFullYear(), visibleMonth.getMonth() + 1, 0).getDate();
    const gridStart = addDays(visibleMonth, -offset);
    const weeks = Array.from({ length: Math.ceil((offset + daysInMonth) / 7) }, (_, week) => Array.from({ length: 7 }, (_, day) => addDays(gridStart, week * 7 + day)));
    // 2023-01-01 is a Sunday, so it is a stable base for locale weekday names.
    const weekdays = Array.from({ length: 7 }, (_, index) => new Date(2023, 0, 1 + ((weekStartsOn + index) % 7)));

    function showMonth(next: Date) {
      const first = startOfMonth(next);
      if (month === undefined) setInnerMonth(first);
      onMonthChange?.(first);
    }
    function moveFocus(target: Date) {
      let next = target;
      if (lowest && next < lowest) next = lowest;
      if (highest && next > highest) next = highest;
      if (!inVisibleMonth(next)) showMonth(next);
      pendingFocus.current = true;
      setHovered(null);
      setFocused(next);
    }
    function choose(date: Date) {
      if (unavailable(date)) return;
      let next: Date | DateRange = date;
      if (isRange) {
        if (range.from && !range.to) {
          const [from, to] = ordered(range.from, date);
          next = { from, to };
        } else next = { from: date, to: null };
      }
      if (value === undefined) setInnerValue(next);
      if (!inVisibleMonth(date)) showMonth(date);
      setFocused(date);
      onValueChange?.(next);
    }
    function handleKeyDown(event: React.KeyboardEvent<HTMLButtonElement>, date: Date) {
      const rtl = getComputedStyle(event.currentTarget).direction === 'rtl';
      const weekday = (date.getDay() - weekStartsOn + 7) % 7;
      let next: Date | undefined;
      switch (event.key) {
        case 'ArrowLeft': next = addDays(date, rtl ? 1 : -1); break;
        case 'ArrowRight': next = addDays(date, rtl ? -1 : 1); break;
        case 'ArrowUp': next = addDays(date, -7); break;
        case 'ArrowDown': next = addDays(date, 7); break;
        case 'Home': next = addDays(date, -weekday); break;
        case 'End': next = addDays(date, 6 - weekday); break;
        case 'PageUp': next = addMonths(date, event.shiftKey ? -12 : -1); break;
        case 'PageDown': next = addMonths(date, event.shiftKey ? 12 : 1); break;
        default: return;
      }
      event.preventDefault();
      moveFocus(next);
    }

    React.useLayoutEffect(() => {
      if (!pendingFocus.current || !focused) return;
      pendingFocus.current = false;
      rootRef.current?.querySelector<HTMLElement>(`[data-date="${formatDateKey(focused)}"]`)?.focus();
    });

    const previousDisabled = disabled || Boolean(lowest && addDays(visibleMonth, -1) < lowest);
    const nextDisabled = disabled || Boolean(highest && addMonths(visibleMonth, 1) > highest);
    return (
      <div {...props} ref={rootRef} role="group" aria-label={ariaLabel} className={cn('ui-calendar', className)} data-disabled={disabled ? '' : undefined}>
        <div className="ui-calendar-header">
          <button type="button" className="ui-calendar-nav" aria-label={previousMonthLabel} disabled={previousDisabled} onClick={() => showMonth(addMonths(visibleMonth, -1))}><ChevronLeft size={16} aria-hidden="true" /></button>
          <div id={captionId} className="ui-calendar-caption" aria-live="polite">{formats.caption.format(visibleMonth)}</div>
          <button type="button" className="ui-calendar-nav" aria-label={nextMonthLabel} disabled={nextDisabled} onClick={() => showMonth(addMonths(visibleMonth, 1))}><ChevronRight size={16} aria-hidden="true" /></button>
        </div>
        <table className="ui-calendar-grid" role="grid" aria-labelledby={captionId} aria-multiselectable={isRange || undefined} onPointerLeave={() => setHovered(null)}>
          <thead>
            <tr>{weekdays.map(day => <th key={day.getDay()} scope="col" abbr={formats.long.format(day)} className="ui-calendar-weekday">{formats.short.format(day)}</th>)}</tr>
          </thead>
          <tbody>
            {weeks.map(week => <tr key={formatDateKey(week[0])}>{week.map(date => {
              const isSelected = isRange ? Boolean((range.from && isSameDay(date, range.from)) || (range.to && isSameDay(date, range.to))) : Boolean(selected && isSameDay(date, selected));
              const inBand = bandLo !== null && bandHi !== null && date >= bandLo && date <= bandHi;
              const isToday = isSameDay(date, today);
              const isUnavailable = unavailable(date);
              return (
                <td
                  key={formatDateKey(date)}
                  role="gridcell"
                  aria-selected={isRange && range.to ? inBand : isSelected}
                  className="ui-calendar-cell"
                  data-range-band={(hasBand && inBand) || undefined}
                  data-range-start={(hasBand && bandLo !== null && isSameDay(date, bandLo)) || undefined}
                  data-range-end={(hasBand && bandHi !== null && isSameDay(date, bandHi)) || undefined}
                  data-range-preview={(hasBand && inBand && pendingEnd) || undefined}
                >
                  <button
                    type="button"
                    className="ui-calendar-day"
                    data-date={formatDateKey(date)}
                    data-selected={isSelected || undefined}
                    data-today={isToday || undefined}
                    data-outside={!inVisibleMonth(date) || undefined}
                    aria-label={formats.full.format(date)}
                    aria-current={isToday ? 'date' : undefined}
                    aria-disabled={isUnavailable || undefined}
                    tabIndex={isSameDay(date, tabStop) ? 0 : -1}
                    onClick={() => choose(date)}
                    onKeyDown={event => handleKeyDown(event, date)}
                    onFocus={() => setFocused(date)}
                    onPointerEnter={pendingEnd ? () => setHovered(date) : undefined}
                  >{formats.day.format(date.getDate())}</button>
                </td>
              );
            })}</tr>)}
          </tbody>
        </table>
        {isRange && <div role="status" className="ui-calendar-status">{pendingEnd ? rangeHint : ''}</div>}
      </div>
    );
  },
);
CalendarImpl.displayName = 'Calendar';

/** The implementation takes a loose union; consumers get the discriminated `mode` API. */
export const Calendar = CalendarImpl as unknown as React.ForwardRefExoticComponent<CalendarProps & React.RefAttributes<HTMLDivElement>>;
