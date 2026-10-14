#!/usr/bin/env node
// Deterministic panel latency and stale-response regressions, no live backend required.
const {chromium}=require(process.env.PLAYWRIGHT_MODULE || 'playwright');
const {seedKey}=require('./lib/seed-key.cjs');
const fs=require('node:fs'); const path=require('node:path'); const assert=require('node:assert/strict');
const root=path.resolve(__dirname,'..');
const html=require('./lib/dashboard-html.cjs').readDashboardHTML(root);
const pause=ms=>new Promise(r=>setTimeout(r,ms));
(async()=>{const browser=await chromium.launch({headless:true,args:['--no-sandbox']});try{
  for(const slow of ['records','summary']) {
    const page=await browser.newPage({viewport:{width:1440,height:1000},locale:'ko-KR'}); const errors=[]; page.on('pageerror',e=>errors.push(e.message));
    await seedKey(page);
    await page.route('**/*',async route=>{
      const u=new URL(route.request().url());
      if(route.request().resourceType()==='document')return route.fulfill({contentType:'text/html',body:html});
      if(u.pathname.includes('/assets/'))return route.fulfill({contentType:'font/woff2',body:fs.readFileSync(root+'/internal/plugin/assets/fonts/'+path.basename(u.pathname))});
      const kind=u.pathname.endsWith('/summary')?'summary':'records';
      await pause(kind===slow?1800:25);
      const now=new Date().toISOString();
      const payload=kind==='summary'?{total_count:1,summary:[{requests:1,total_tokens:123456}],by_provider:[],by_endpoint:[],by_model:[]}:{records:[{id:'first',requested_at:now,provider:'codex',model:'test',total_tokens:123456}],has_more:false};
      try{await route.fulfill({contentType:'application/json',body:JSON.stringify(payload)});}catch{}
    });
    await page.goto('http://fixture.invalid/v0/resource/plugins/usage-audit/dashboard',{waitUntil:'domcontentloaded'});
    await page.waitForFunction(kind=>kind==='records'?document.querySelector('#kpiTotalTokens').textContent==='123,456':document.querySelector('#recordsBody').textContent.includes('test'),slow,{timeout:1500});
    assert(await page.evaluate(()=>state.inFlight),'fast panel must paint while sibling still pending');
    const timing=await page.evaluate(()=>({elapsed:Math.round(performance.now()),dom:Math.round(performance.getEntriesByType('navigation')[0].domContentLoadedEventEnd)}));
    await page.waitForFunction(()=>state.hasLoadedData && !state.inFlight);
    assert.deepEqual(errors,[]);
    assert.equal(await page.locator('.update-status, #updateStatus, #lastUpdateTime, #lastUpdateRel').count(),0);
    await page.locator('[data-preset=custom]').click(); await page.locator('#filterDateRangeTrigger').click();
    const row=page.locator('.ui-calendar-grid tbody tr').nth(1);
    await row.locator('button').nth(2).click(); // select a draft start, then hover its week
    await row.locator('button').nth(3).hover();
    assert.equal(await row.evaluate(e=>getComputedStyle(e).backgroundColor),'rgba(0, 0, 0, 0)','calendar must not inherit data-table row hover');
    assert.equal(await row.evaluate(e=>getComputedStyle(e).cursor),'auto','calendar row has no record cursor');
    await page.keyboard.press('Escape');
    console.log('PASS progressive '+slow+' delayed 1800ms; fast panel painted '+timing.elapsed+'ms (DOMContentLoaded '+timing.dom+'ms)');
    await page.close();
  }
  const page=await browser.newPage({locale:'ko-KR'}); let calls=0;
  await seedKey(page);
  await page.route('**/*',async route=>{
    const url=new URL(route.request().url());
    if(route.request().resourceType()==='document')return route.fulfill({contentType:'text/html',body:html});
    if(url.pathname.includes('/assets/'))return route.fulfill({body:fs.readFileSync(root+'/internal/plugin/assets/fonts/'+path.basename(url.pathname))});
    calls++;
    const old=url.searchParams.get('from')!=='2000-01-01T00:00:00.000Z'; await pause(old?1000:20);
    const payload=url.pathname.endsWith('/summary')?{total_count:old?1:2,summary:[{total_tokens:old?111:222}]}:{records:[{id:old?'old':'new',requested_at:new Date().toISOString(),provider:'codex'}]};
    try{await route.fulfill({contentType:'application/json',body:JSON.stringify(payload)});}catch{}
  });
  await page.goto('http://fixture.invalid/v0/resource/plugins/usage-audit/dashboard',{waitUntil:'domcontentloaded'});
  await page.evaluate(()=>{state.fromRFC3339='2000-01-01T00:00:00.000Z';loadData(false);});
  await page.waitForFunction(()=>state.hasLoadedData && !state.inFlight);
  await pause(1100);
  assert.equal(await page.locator('#kpiTotalTokens').innerText(),'222');
  assert.equal(await page.evaluate(()=>state.records[0].id),'new');
  assert.equal(calls,4,'only summary and records per query');
  console.log('PASS stale request cannot overwrite panels'); await page.close();
  // A first-load failure cancels its delayed sibling; it must never paint afterwards.
  for (const failingKind of ['summary','records']) {
    const first=await browser.newPage({locale:'ko-KR'});
    await seedKey(first);
    await first.route('**/*',async route=>{
      const u=new URL(route.request().url());
      if(route.request().resourceType()==='document')return route.fulfill({contentType:'text/html',body:html});
      if(u.pathname.includes('/assets/'))return route.fulfill({body:fs.readFileSync(root+'/internal/plugin/assets/fonts/'+path.basename(u.pathname))});
      const kind=u.pathname.endsWith('/summary')?'summary':'records';
      await pause(kind===failingKind?25:500);
      const payload=kind==='summary'?{summary:[{total_tokens:999}]}:{records:[{id:'must-not-render',model:'must-not-render',requested_at:new Date().toISOString()}]};
      try{await route.fulfill({status:kind===failingKind?500:200,contentType:'application/json',body:JSON.stringify(payload)});}catch{}
    });
    await first.goto('http://fixture.invalid/v0/resource/plugins/usage-audit/dashboard');
    await first.waitForFunction(()=>!state.inFlight && !document.getElementById('errorBox').classList.contains('is-hidden'));
    await pause(600);
    assert.equal(await first.evaluate(()=>state.abortCtrl.signal.aborted),true);
    assert.equal(await first.evaluate(()=>state.hasLoadedData),false);
    assert(!await first.locator('#recordsBody').innerText().then(t=>t.includes('must-not-render')));
    assert.notEqual(await first.locator('#kpiTotalTokens').innerText(),'999');
    await first.close(); console.log('PASS first-load '+failingKind+' failure aborts sibling');
  }
  // Once loaded, a failed refresh must not mix new KPIs with retained records/cursor.
  const refresh=await browser.newPage({locale:'ko-KR'}); let failing=false;
  await seedKey(refresh);
  await refresh.route('**/*',async route=>{
    const url=new URL(route.request().url());
    if(route.request().resourceType()==='document')return route.fulfill({contentType:'text/html',body:html});
    if(url.pathname.includes('/assets/'))return route.fulfill({body:fs.readFileSync(root+'/internal/plugin/assets/fonts/'+path.basename(url.pathname))});
    if(failing && url.pathname.endsWith('/records')) {
      await pause(150); return route.fulfill({status:500,contentType:'application/json',body:JSON.stringify({error:'offline'})});
    }
    const payload=url.pathname.endsWith('/summary')?{total_count:1,summary:[{total_tokens:failing?999:333}]}:{records:[{id:'retained',requested_at:new Date().toISOString()}],has_more:true,next_cursor:'retained-cursor'};
    await route.fulfill({contentType:'application/json',body:JSON.stringify(payload)});
  });
  await refresh.goto('http://fixture.invalid/v0/resource/plugins/usage-audit/dashboard');
  await refresh.waitForFunction(()=>state.hasLoadedData && !state.inFlight);
  failing=true; await refresh.evaluate(()=>loadData(false));
  await refresh.waitForFunction(()=>!state.inFlight && !document.getElementById('errorBox').classList.contains('is-hidden'));
  assert.equal(await refresh.locator('#kpiTotalTokens').innerText(),'333');
  assert.deepEqual(await refresh.evaluate(()=>[state.records[0].id,state.cursor,state.hasMore]),['retained','retained-cursor',true]);
  console.log('PASS failed refresh preserves previous panels and cursor'); await refresh.close();
}finally{await browser.close();}})().catch(e=>{console.error(e);process.exitCode=1;});
