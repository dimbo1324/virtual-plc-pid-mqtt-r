// app.js — Virtual PLC Dashboard: SSE, loop cards, resizable panels, i18n.
'use strict';

const charts = {};
const loopCards = {};
let uptimeInterval = null;
let uptimeStart = null;

// ── SSE connection ──────────────────────────────────────────────────────────

function initSSE() {
    const dot = document.getElementById('connection-dot');
    let reconnectMs = 1000;

    function connect() {
        const es = new EventSource('/api/stream');

        es.addEventListener('snapshot', e => {
            try { handleSnapshot(JSON.parse(e.data)); } catch (_) {}
            if (dot) { dot.className = 'dot connected'; dot.title = t('header.conn.ok'); }
            reconnectMs = 1000;
        });

        es.addEventListener('plc_event', e => {
            try {
                const ev = JSON.parse(e.data);
                handlePLCEvent(ev);
                appendEvent(ev);
            } catch (_) {}
        });

        es.addEventListener('heartbeat', e => {
            if (dot) { dot.className = 'dot connected'; dot.title = t('header.conn.ok'); }
            try {
                const hb = JSON.parse(e.data);
                if (hb.ts && !uptimeStart) uptimeStart = new Date(hb.ts);
            } catch (_) {}
        });

        es.onerror = () => {
            if (dot) { dot.className = 'dot lost'; dot.title = t('header.conn.lost'); }
            es.close();
            setTimeout(connect, reconnectMs);
            reconnectMs = Math.min(reconnectMs * 2, 16000);
        };
    }

    connect();
}

// ── Snapshot handler ────────────────────────────────────────────────────────

function handleSnapshot(snap) {
    updateHeader(snap);
    const loops = snap.loops || {};
    const grid = document.getElementById('charts-grid');
    for (const [name, loop] of Object.entries(loops)) {
        if (!loopCards[name]) {
            loopCards[name] = createLoopCard(grid, name, loop);
        }
        updateLoopCard(loopCards[name], loop);
        if (charts[name]) charts[name].push(loop.pv ?? 0, loop.sp ?? 0, loop.mv ?? 0);
    }
}

function handlePLCEvent(ev) {
    const loops = ev.loops || {};
    for (const [name, loop] of Object.entries(loops)) {
        if (loopCards[name]) updateLoopCard(loopCards[name], loop);
        if (charts[name]) charts[name].push(loop.pv ?? 0, loop.sp ?? 0, loop.mv ?? 0);
    }
    if (ev.plc) updatePLCState(ev.plc.state);
}

function updateHeader(snap) {
    const devEl = document.getElementById('device-id');
    if (devEl) devEl.textContent = snap.device_id || '—';
    if (snap.plc) updatePLCState(snap.plc.state);

    // Start uptime clock from snapshot timestamp
    if (snap.plc?.started_at) {
        uptimeStart = new Date(snap.plc.started_at);
    }
}

function updatePLCState(state) {
    const stateEl = document.getElementById('plc-state');
    if (!stateEl) return;
    const label = state === 'running' ? t('header.state.running')
        : state === 'stopped' ? t('header.state.stopped')
        : t('header.state.unknown');
    stateEl.textContent = label;
    stateEl.className = 'state-badge ' + (state || '');
}

// ── Uptime clock ────────────────────────────────────────────────────────────

function formatUptime(ms) {
    if (ms < 0) ms = 0;
    const s = Math.floor(ms / 1000) % 60;
    const m = Math.floor(ms / 60000) % 60;
    const h = Math.floor(ms / 3600000);
    return h > 0
        ? `${h}h ${String(m).padStart(2,'0')}m ${String(s).padStart(2,'0')}s`
        : `${String(m).padStart(2,'0')}m ${String(s).padStart(2,'0')}s`;
}

function tickUptime() {
    const el = document.getElementById('uptime');
    if (!el) return;
    if (uptimeStart) {
        el.textContent = formatUptime(Date.now() - uptimeStart.getTime());
    }
}

// ── Loop card creation ──────────────────────────────────────────────────────

function createLoopCard(container, name, loop) {
    const displayName = loop.display_name || name;

    const card = document.createElement('div');
    card.className = 'loop-card';

    card.innerHTML = `
<div class="loop-header">
  <span class="loop-name">${escHtml(displayName)}</span>
  <span class="loop-unit">${escHtml(loop.unit || '')}</span>
  <span class="loop-mode-badge"></span>
</div>
<div class="loop-values">
  <div class="val-block">
    <span class="val-label" data-i18n="card.pv">${t('card.pv')}</span>
    <span class="val-num pv-val">—</span>
    <span class="val-unit">${escHtml(loop.unit || '')}</span>
  </div>
  <div class="val-block">
    <span class="val-label" data-i18n="card.sp">${t('card.sp')}</span>
    <span class="val-num sp-val">—</span>
    <span class="val-unit">${escHtml(loop.unit || '')}</span>
  </div>
  <div class="val-block">
    <span class="val-label" data-i18n="card.mv">${t('card.mv')}</span>
    <span class="val-num mv-val">—</span>
    <span class="val-unit">%</span>
  </div>
</div>
<div class="chart-wrap">
  <canvas class="trend-canvas"></canvas>
  <div class="chart-toolbar">
    <button class="chart-btn-fs"    data-i18n-title="chart.fullscreen" title="${t('chart.fullscreen')}">⛶</button>
    <button class="chart-btn-reset" data-i18n-title="chart.reset"      title="${t('chart.reset')}">↺</button>
    <button class="chart-btn-png"   data-i18n="chart.png">${t('chart.png')}</button>
    <button class="chart-btn-pdf"   data-i18n="chart.pdf">${t('chart.pdf')}</button>
  </div>
</div>
<div class="loop-controls">
  <div class="ctrl-row">
    <input class="ctrl-input sp-input" type="number" step="0.1" placeholder="${t('card.setpoint')}">
    <button class="ctrl-btn accent sp-btn" data-i18n="card.set">${t('card.set')}</button>
  </div>
  <div class="ctrl-row">
    <select class="ctrl-select mode-select">
      <option value="auto"    data-i18n="card.auto">${t('card.auto')}</option>
      <option value="manual"  data-i18n="card.manual">${t('card.manual')}</option>
      <option value="cascade" data-i18n="card.cascade">${t('card.cascade')}</option>
    </select>
    <button class="ctrl-btn mode-btn" data-i18n="card.mode">${t('card.mode')}</button>
  </div>
  <div class="ctrl-row">
    <button class="ctrl-btn accent disturbance-btn" data-i18n="card.disturbance">${t('card.disturbance')}</button>
    <button class="ctrl-btn danger reset-btn" data-i18n="card.reset">${t('card.reset')}</button>
  </div>
  <button class="gains-toggle" type="button">
    <span class="chevron">▶</span>
    <span data-i18n="card.gains">${t('card.gains')}</span>
  </button>
  <div class="gains-panel">
    <div class="gain-item"><span class="gain-label" data-i18n="card.kp">${t('card.kp')}</span><span class="gain-val kp-val">—</span></div>
    <div class="gain-item"><span class="gain-label" data-i18n="card.ki">${t('card.ki')}</span><span class="gain-val ki-val">—</span></div>
    <div class="gain-item"><span class="gain-label" data-i18n="card.kd">${t('card.kd')}</span><span class="gain-val kd-val">—</span></div>
  </div>
</div>`;

    container.appendChild(card);

    // Wire chart
    const canvas = card.querySelector('.trend-canvas');
    const chart = createTrendChart(canvas, displayName);
    charts[name] = chart;

    // Chart toolbar buttons
    card.querySelector('.chart-btn-fs').addEventListener('click', () => chart.enterFullscreen());
    card.querySelector('.chart-btn-reset').addEventListener('click', () => chart.resetView());
    card.querySelector('.chart-btn-png').addEventListener('click', () => chart.exportPNG());
    card.querySelector('.chart-btn-pdf').addEventListener('click', () => chart.exportPDF());

    // Loop control buttons
    const spInput = card.querySelector('.sp-input');
    card.querySelector('.sp-btn').addEventListener('click', () => {
        const v = parseFloat(spInput.value);
        if (!isNaN(v)) postCommand('setpoint', { loop: name, setpoint: v });
    });
    spInput.addEventListener('keydown', e => {
        if (e.key === 'Enter') { const v = parseFloat(spInput.value); if (!isNaN(v)) postCommand('setpoint', { loop: name, setpoint: v }); }
    });

    const modeSelect = card.querySelector('.mode-select');
    card.querySelector('.mode-btn').addEventListener('click', () => {
        postCommand('mode', { loop: name, mode: modeSelect.value });
    });

    card.querySelector('.disturbance-btn').addEventListener('click', () => {
        postCommand('inject_disturbance', { loop: name, amplitude: 5, duration_seconds: 30 });
    });

    card.querySelector('.reset-btn').addEventListener('click', () => {
        postCommand('reset_loop', { loop: name });
    });

    // Gains toggle
    const gainsToggle = card.querySelector('.gains-toggle');
    const gainsPanel  = card.querySelector('.gains-panel');
    gainsToggle.addEventListener('click', () => {
        const open = gainsPanel.classList.toggle('visible');
        gainsToggle.classList.toggle('open', open);
    });

    // Populate gains if available
    if (loop.pid) {
        card.querySelector('.kp-val').textContent = loop.pid.kp ?? '—';
        card.querySelector('.ki-val').textContent = loop.pid.ki ?? '—';
        card.querySelector('.kd-val').textContent = loop.pid.kd ?? '—';
    }

    return card;
}

function updateLoopCard(card, loop) {
    const q = s => card.querySelector(s);

    q('.pv-val').textContent = loop.pv != null ? loop.pv.toFixed(2) : '—';
    q('.sp-val').textContent = loop.sp != null ? loop.sp.toFixed(2) : '—';
    q('.mv-val').textContent = loop.mv != null ? loop.mv.toFixed(1) : '—';

    const badge = q('.loop-mode-badge');
    if (badge) {
        const modeKey = 'card.' + (loop.mode || 'auto');
        badge.textContent = t(modeKey) || loop.mode || '';
        badge.className = 'loop-mode-badge mode-' + (loop.mode || '');
    }

    const sel = q('.mode-select');
    if (sel && document.activeElement !== sel) sel.value = loop.mode || 'auto';
}

// ── Event terminal ──────────────────────────────────────────────────────────

function appendEvent(ev) {
    const terminal = document.getElementById('event-terminal');
    if (!terminal) return;

    // Remove empty placeholder
    const placeholder = terminal.querySelector('.events-empty');
    if (placeholder) placeholder.remove();

    const entry = document.createElement('div');
    const level = ev.level || 'info';
    entry.className = 'event-entry level-' + level;
    const ts = ev.timestamp ? new Date(ev.timestamp).toLocaleTimeString() : '—';

    // Build compact per-loop summary if available
    let msg = ev.message || '';
    if (!msg && ev.loops) {
        msg = Object.entries(ev.loops)
            .map(([n, l]) => `${n}: PV=${(l.pv ?? 0).toFixed(1)} SP=${(l.sp ?? 0).toFixed(1)} MV=${(l.mv ?? 0).toFixed(1)}%`)
            .join('  |  ');
    }

    entry.textContent = `[${ts}] ${msg}`;
    terminal.appendChild(entry);

    while (terminal.children.length > 200) terminal.removeChild(terminal.firstChild);
    terminal.scrollTop = terminal.scrollHeight;
}

function clearEvents() {
    const terminal = document.getElementById('event-terminal');
    if (!terminal) return;
    terminal.innerHTML = `<div class="events-empty" data-i18n="events.empty">${t('events.empty')}</div>`;
}

// ── Command helpers ──────────────────────────────────────────────────────────

async function postCommand(path, body = {}) {
    try {
        const res = await fetch('/api/commands/' + path, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(body),
        });
        if (!res.ok) {
            const err = await res.json().catch(() => ({}));
            console.warn('command rejected:', err.error || res.status);
        }
    } catch (e) {
        console.error('command error:', e);
    }
}

// ── Resizable panels ─────────────────────────────────────────────────────────

function initResizeHandle() {
    const handle    = document.getElementById('resize-handle');
    const mainPanel = document.getElementById('main-panel');
    const evPanel   = document.getElementById('events-panel');
    if (!handle || !mainPanel || !evPanel) return;

    let dragging = false;
    let startX = 0;
    let startW = 0;

    handle.addEventListener('mousedown', e => {
        if (e.button !== 0) return;
        dragging = true;
        startX = e.clientX;
        startW = evPanel.offsetWidth;
        handle.classList.add('dragging');
        document.body.style.cursor = 'col-resize';
        document.body.style.userSelect = 'none';
    });

    window.addEventListener('mousemove', e => {
        if (!dragging) return;
        const dx = startX - e.clientX;
        const newW = Math.max(180, Math.min(600, startW + dx));
        evPanel.style.width = newW + 'px';
    });

    window.addEventListener('mouseup', e => {
        if (!dragging || e.button !== 0) return;
        dragging = false;
        handle.classList.remove('dragging');
        document.body.style.cursor = '';
        document.body.style.userSelect = '';
    });
}

// ── Manual drawer ────────────────────────────────────────────────────────────

let currentTab = 'overview';

function openManual(tab) {
    tab = tab || currentTab;
    currentTab = tab;
    renderManualTab(tab);
    document.getElementById('manual-drawer').classList.add('open');
    document.getElementById('drawer-overlay').classList.add('open');
}

function closeManual() {
    document.getElementById('manual-drawer').classList.remove('open');
    document.getElementById('drawer-overlay').classList.remove('open');
}

function renderManualTab(tab) {
    const content = document.getElementById('manual-content');
    if (content) content.innerHTML = getManualContent(tab);

    document.querySelectorAll('.drawer-tab').forEach(btn => {
        btn.classList.toggle('active', btn.dataset.tab === tab);
    });
}

function initManualDrawer() {
    document.getElementById('btn-manual')?.addEventListener('click', () => openManual());
    document.getElementById('btn-close-manual')?.addEventListener('click', closeManual);
    document.getElementById('drawer-overlay')?.addEventListener('click', closeManual);

    document.querySelectorAll('.drawer-tab').forEach(btn => {
        btn.addEventListener('click', () => {
            currentTab = btn.dataset.tab;
            renderManualTab(currentTab);
        });
    });

    // Render initial content
    renderManualTab(currentTab);
}

// ── Language toggle ──────────────────────────────────────────────────────────

function initLangToggle() {
    const btn = document.getElementById('btn-lang');
    if (!btn) return;

    // Apply persisted language on load
    setLanguage(getCurrentLang());

    btn.addEventListener('click', () => {
        const next = getCurrentLang() === 'en' ? 'ru' : 'en';
        setLanguage(next);
        // Re-render manual if open
        renderManualTab(currentTab);
        // Refresh event log placeholder if empty
        const placeholder = document.querySelector('#event-terminal .events-empty');
        if (placeholder) placeholder.textContent = t('events.empty');
    });

    document.addEventListener('langchange', () => {
        // Update PLC state label after language change
        const stateEl = document.getElementById('plc-state');
        if (stateEl) {
            const cls = [...stateEl.classList].find(c => c !== 'state-badge') || '';
            updatePLCState(cls);
        }
        // Re-render loop card mode badges and i18n strings inside cards
        for (const [name, card] of Object.entries(loopCards)) {
            const badge = card.querySelector('.loop-mode-badge');
            if (badge) {
                const mode = badge.className.replace(/.*mode-/, '').trim();
                badge.textContent = t('card.' + mode) || mode;
            }
        }
    });
}

// ── Header buttons ───────────────────────────────────────────────────────────

function initHeaderButtons() {
    document.getElementById('btn-start')?.addEventListener('click', () => postCommand('start'));
    document.getElementById('btn-stop')?.addEventListener('click',  () => postCommand('stop'));
    document.getElementById('btn-clear-events')?.addEventListener('click', clearEvents);
}

// ── Utility ──────────────────────────────────────────────────────────────────

function escHtml(s) {
    return String(s)
        .replace(/&/g, '&amp;')
        .replace(/</g, '&lt;')
        .replace(/>/g, '&gt;')
        .replace(/"/g, '&quot;');
}

// ── Init ─────────────────────────────────────────────────────────────────────

window.addEventListener('DOMContentLoaded', () => {
    initLangToggle();
    initHeaderButtons();
    initManualDrawer();
    initResizeHandle();
    initSSE();

    uptimeInterval = setInterval(tickUptime, 1000);
});
