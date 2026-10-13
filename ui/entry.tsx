import * as React from 'react';
import { createRoot } from 'react-dom/client';
import { flushSync } from 'react-dom';
import { SegmentedControl, SegmentedControlItem } from './vendor/segmented-control';
import { DatePicker } from './vendor/date-picker';
import { formatDateKey } from './vendor/calendar';
import { Select, SelectTrigger, SelectValue, SelectContent, SelectItem } from './vendor/select';
import { TableContainer, Table, TableHeader, TableBody, TableRow, TableHead, TableCell } from './vendor/table';

declare const state: { currentPreset: string };
declare function setPreset(value: string): void;
declare function applyFilters(): void;
const ALL = '__dashboard_all__';
const presets = [['today', '오늘'], ['yesterday', '어제'], ['7d', '7일'], ['30d', '30일'], ['month', '이번 달'], ['custom', '직접']];
const parse = (v: string) => v ? new Date(...v.split('-').map((s, i) => Number(s) - (i === 1 ? 1 : 0)) as [number, number, number]) : null;

// Existing query code owns the values. This adapter only bridges native form state to the
// exported dohwi-ui components; no popup, focus, calendar or radio behavior is reimplemented.
function watchValue(input: HTMLInputElement | HTMLSelectElement, render: () => void) {
  const descriptor = Object.getOwnPropertyDescriptor(Object.getPrototypeOf(input), 'value')!;
  Object.defineProperty(input, 'value', { configurable: true,
    get() { return descriptor.get!.call(this); },
    set(value) { descriptor.set!.call(this, value); render(); },
  });
  input.addEventListener('change', render);
  new MutationObserver(render).observe(input, { attributes: true, childList: true, subtree: true });
}

const presetHost = document.getElementById('presetGroup')!;
const presetRoot = createRoot(presetHost);
function renderPresets() {
  presetRoot.render(<SegmentedControl id="presetGroupControl" size="sm" fullWidth value={state.currentPreset}
    aria-label="기간 선택" orientation="horizontal" onValueChange={setPreset}>
    {presets.map(([value, label]) => <SegmentedControlItem key={value} value={value} data-preset={value}
      className="preset-btn"
      aria-controls={value === 'custom' ? 'customDateWrap' : undefined}>{label}</SegmentedControlItem>)}
  </SegmentedControl>);
}
(window as any).renderDashboardPresets = renderPresets;
flushSync(renderPresets);

const from = document.getElementById('filterFromDate') as HTMLInputElement;
const to = document.getElementById('filterToDate') as HTMLInputElement;
const dateRoot = createRoot(document.getElementById('dateRangeMount')!);
function renderDate() {
  dateRoot.render(<DatePicker id="filterDateRangeTrigger" mode="range" value={{ from: parse(from.value), to: parse(to.value) }}
    disabled={from.disabled || to.disabled} aria-label="조회 기간" calendarLabel="조회 기간 선택" locale="ko-KR"
    onValueChange={range => {
      // Atomic commit: never dispatch the old per-endpoint change handler while only one
      // endpoint has been updated. DatePicker emits only a complete, ordered range.
      from.value = formatDateKey(range.from);
      to.value = formatDateKey(range.to);
      applyFilters();
    }} />);
}
watchValue(from, renderDate); watchValue(to, renderDate);
flushSync(renderDate);

// Use the exported dohwi-ui Table primitives for all table semantics and structure.
type TableCellData = { text?: string; content?: React.ReactNode; time?: {full:string; mobile:string}; badge?: {text:string; className:string}; className?: string; title?: string; changeKey?: string; tone?: string; label?: string; colSpan?: number };
type TableRowData = { key: string; cells: TableCellData[]; onClick?: (event: React.MouseEvent<HTMLTableRowElement>) => void; onKeyDown?: (event: React.KeyboardEvent<HTMLTableRowElement>) => void; tabIndex?: number; dataIndex?: number };
type TableViewProps = { mountId: string; tableId: string; bodyId: string; label: string; wrapperClass: string; headers: Array<{text:string; className?:string}>; rows: TableRowData[]; emptyText?: string; tableClass?: string };
const tableRoots = new Map<string, ReturnType<typeof createRoot>>();
function cellContent(cell: TableCellData) {
  if (cell.time) return <><span className="record-time-full">{cell.time.full}</span><span className="record-time-mobile">{cell.time.mobile}</span></>;
  if (cell.badge) return <span className={cell.badge.className}>{cell.badge.text}</span>;
  return cell.content ?? cell.text ?? ''; // cellContent(cell); title={cell.title}; flashValue(el,cell.tone,cell.label)
}
function TableView({tableId, bodyId, label, wrapperClass, headers, rows, emptyText, tableClass}: TableViewProps) {
  return <TableContainer role="region" aria-label={label} className={wrapperClass}>
    <Table id={tableId} className={tableClass}>
      <TableHeader><TableRow>{headers.map((h, i)=><TableHead key={i} className={h.className}>{h.text}</TableHead>)}</TableRow></TableHeader>
      <TableBody id={bodyId}>
        {rows.length ? rows.map(row=><TableRow key={row.key} tabIndex={row.tabIndex} data-index={row.dataIndex} data-row-key={row.key} onClick={row.onClick} onKeyDown={row.onKeyDown}>
          {row.cells.map((cell, i)=><TableCell key={i} colSpan={cell.colSpan} className={cell.className} title={cell.title} data-change-key={cell.changeKey}
            ref={el=>{if(el && cell.tone && typeof (window as any).flashValue==='function') (window as any).flashValue(el,cell.tone,cell.label);}}>
            {cellContent(cell)}
          </TableCell>)}
        </TableRow>) : <TableRow><TableCell colSpan={headers.length} className="text-muted table-empty">{emptyText || '데이터 없음'}</TableCell></TableRow>}
      </TableBody>
    </Table>
  </TableContainer>;
}
const AGG_HEADERS=[{text:'Provider'},{text:'요청',className:'text-right'},{text:'실패',className:'text-right'},{text:'토큰',className:'text-right'}];
const ENDPOINT_HEADERS=[{text:'Endpoint/Source'},...AGG_HEADERS.slice(1)];
const MODEL_HEADERS=[{text:'Model'},...AGG_HEADERS.slice(1)];
const RECORD_HEADERS=[{text:'시간'},{text:'Provider'},{text:'Endpoint'},{text:'Model'},
  {text:'신규 입력',className:'text-right'},{text:'캐시',className:'text-right'},
  {text:'출력',className:'text-right'},{text:'합계',className:'text-right'},
  {text:'지연',className:'text-right'},{text:'상태'}];
// Keep the loading row's column count identical to the data table. A single colspan cell
// makes browsers recalculate fixed-layout column widths when the first records arrive.
const loadingAggRows=[{key:'loading',cells:[{text:'불러오는 중...',className:'text-muted table-empty'},{text:''},{text:''},{text:''}]}];
const loadingRecordRows=[{key:'loading',cells:[{text:'데이터를 불러오는 중입니다...',className:'text-muted table-empty'},...Array.from({length:9},()=>({text:''}))]}];
const tableConfigs: TableViewProps[] = [
  {mountId:'providerAggTableMount',tableId:'providerAggTable',bodyId:'providerAggBody',label:'Provider 상세 집계',wrapperClass:'usage-table-scroll breakdown-table-wrap',headers:AGG_HEADERS,rows:loadingAggRows},
  {mountId:'endpointAggTableMount',tableId:'endpointAggTable',bodyId:'endpointAggBody',label:'Endpoint 상세 집계',wrapperClass:'usage-table-scroll breakdown-table-wrap',headers:ENDPOINT_HEADERS,rows:loadingAggRows},
  {mountId:'modelAggTableMount',tableId:'modelAggTable',bodyId:'modelAggBody',label:'Model 상세 집계',wrapperClass:'usage-table-scroll breakdown-table-wrap',headers:MODEL_HEADERS,rows:loadingAggRows},
  {mountId:'recordsTableMount',tableId:'recordsTable',bodyId:'recordsBody',label:'최근 요청 기록 테이블 스크롤 영역',wrapperClass:'usage-table-scroll table-responsive',headers:RECORD_HEADERS,rows:loadingRecordRows,tableClass:'records-table'},
];
for (const config of tableConfigs) {
  const mount=document.getElementById(config.mountId);
  if(!mount) continue;
  const root=createRoot(mount); tableRoots.set(config.mountId,root);
  flushSync(()=>root.render(<TableView {...config}/>));
}
(window as any).renderDohwiTable = (mountId:string, rows:TableRowData[], emptyText?:string) => {
  const root=tableRoots.get(mountId), config=tableConfigs.find(item=>item.mountId===mountId);
  if(root && config) flushSync(()=>root.render(<TableView {...config} rows={rows} emptyText={emptyText}/>));
};
(window as any).dohwiTableHeaders={AGG_HEADERS,ENDPOINT_HEADERS,MODEL_HEADERS,RECORD_HEADERS};
for(const update of (window as any).pendingDohwiTableUpdates||[]) (window as any).renderDohwiTable(update[0],update[1],update[2]);
(window as any).pendingDohwiTableUpdates=[];

for (const select of document.querySelectorAll<HTMLSelectElement>('.compact-select-wrap > select')) {
  select.hidden = true;
  select.tabIndex = -1;
  const mount = document.createElement('div');
  mount.className = 'dashboard-select-mount';
  select.parentElement!.appendChild(mount);
  const root = createRoot(mount);
  function render() {
    root.render(<Select value={select.value || ALL} disabled={select.disabled} onValueChange={value => {
      select.value = value === ALL ? '' : value;
      select.dispatchEvent(new Event('change', { bubbles: true }));
    }}>
      <SelectTrigger id={select.id + 'Trigger'} className="compact-select" aria-label={select.getAttribute('aria-label')!} title={select.title}>
        <SelectValue />
      </SelectTrigger>
      <SelectContent collisionPadding={16}>
        {Array.from(select.options).map(option => <SelectItem key={option.value || ALL} value={option.value || ALL} disabled={option.disabled}>{option.textContent}</SelectItem>)}
      </SelectContent>
    </Select>);
  }
  watchValue(select, render);
  document.querySelector(`label[for="${select.id}"]`)?.setAttribute('for', select.id + 'Trigger');
  flushSync(render);
}
