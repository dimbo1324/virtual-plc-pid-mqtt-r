// app.js — Virtual PLC Dashboard: SSE, loop cards, resizable panels, i18n.
'use strict';

const charts = {};
const loopCards = {};
let uptimeInterval = null;
// uptimeBase is the ISO timestamp received from /api/status; used to tick uptime locally.
let uptimeBase = null;

// ── SSE connection ──────────────────────────────────────────────────────────

function initSSE() {
    const dot = document.getElementById('connection-dot');
    let reconnectMs = 1000;
    let wasError = false;

    function connect() {
        const es = new EventSource('/api/stream');

        es.addEventListener('snapshot', e => {
            try { handleSnapshot(JSON.parse(e.data)); } catch (_) {}
            if (dot) { dot.className = 'dot connected'; dot.title = t('header.conn.ok'); }
            if (wasError) {
                wasError = false;
                loadStatus(); // refresh storage_mode and header after reconnect
            }
            reconnectMs = 1000;
        });

        // plc_event carries a plc.Event (log entry: level, message, timestamp, details).
        // It does NOT carry loop state — use snapshot for that.
        es.addEventListener('plc_event', e => {
            try { appendEvent(JSON.parse(e.data)); } catch (_) {}
        });

        es.addEventListener('heartbeat', () => {
            if (dot) { dot.className = 'dot connected'; dot.title = t('header.conn.ok'); }
        });

        es.onerror = () => {
            if (dot) { dot.className = 'dot lost'; dot.title = t('header.conn.lost'); }
            wasError = true;
            es.close();
            setTimeout(connect, reconnectMs);
            reconnectMs = Math.min(reconnectMs * 2, 16000);
        };
    }

    connect();
}

// ── Snapshot handler ────────────────────────────────────────────────────────
// snapshot carries plc.Snapshot: {device_id, plc:{state,…}, loops:{name:LoopSnapshot}}
// It is broadcast on every UIUpdateInterval (~200 ms), driving real-time chart updates.

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

function updateHeader(snap) {
    const devEl = document.getElementById('device-id');
    if (devEl) devEl.textContent = snap.device_id || '—';
    if (snap.plc) updatePLCState(String(snap.plc.state || ''));
}

function updatePLCState(state) {
    const stateEl = document.getElementById('plc-state');
    if (!stateEl) return;
    const label = state === 'running'  ? t('header.state.running')
        : state === 'stopped'  ? t('header.state.stopped')
        : state ? state.toUpperCase()
        : '—';
    stateEl.textContent = label;
    stateEl.className = 'state-badge ' + (state || '');
}

// ── Uptime ──────────────────────────────────────────────────────────────────

function formatUptime(ms) {
    if (ms < 0) ms = 0;
    const s = Math.floor(ms / 1000) % 60;
    const m = Math.floor(ms / 60000) % 60;
    const h = Math.floor(ms / 3600000);
    if (h > 0) return `${h}h ${String(m).padStart(2, '0')}m ${String(s).padStart(2, '0')}s`;
    return `${String(m).padStart(2, '0')}m ${String(s).padStart(2, '0')}s`;
}

function tickUptime() {
    const el = document.getElementById('uptime');
    if (!el || !uptimeBase) return;
    el.textContent = formatUptime(Date.now() - uptimeBase);
}

// ── Initial status load ─────────────────────────────────────────────────────
// Populates header state before the first SSE snapshot arrives.

async function loadStatus() {
    try {
        const res = await fetch('/api/status');
        if (!res.ok) return;
        const s = await res.json();
        const devEl = document.getElementById('device-id');
        if (devEl) devEl.textContent = s.device_id || '—';
        if (s.state) updatePLCState(s.state);
        if (s.server_time && s.uptime) {
            const el = document.getElementById('uptime');
            if (el) el.textContent = s.uptime;
        }
        updateStorageMode(s.storage_mode);
    } catch (_) {}
}

function updateStorageMode(mode) {
    const badge = document.getElementById('storage-badge');
    if (!badge) return;
    if (mode === 'degraded') {
        badge.style.display = '';
        badge.title = t('header.storage.degraded.title');
    } else {
        badge.style.display = 'none';
    }
}

// ── Loop card creation ──────────────────────────────────────────────────────
// LoopSnapshot JSON fields: name, display_name, unit, sp, pv, mv,
//   error, mode, quality, enabled, kp, ki, kd

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
    <button class="chart-btn-fs"    title="${t('chart.fullscreen')}" data-i18n-title="chart.fullscreen">⛶</button>
    <button class="chart-btn-reset" title="${t('chart.reset')}"      data-i18n-title="chart.reset">↺</button>
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
      <option value="auto"     data-i18n="card.auto">${t('card.auto')}</option>
      <option value="manual"   data-i18n="card.manual">${t('card.manual')}</option>
      <option value="hold"     data-i18n="card.hold">${t('card.hold')}</option>
      <option value="disabled" data-i18n="card.disabled">${t('card.disabled')}</option>
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
    <div class="gain-item">
      <span class="gain-label" data-i18n="card.kp">${t('card.kp')}</span>
      <span class="gain-val kp-val">—</span>
    </div>
    <div class="gain-item">
      <span class="gain-label" data-i18n="card.ki">${t('card.ki')}</span>
      <span class="gain-val ki-val">—</span>
    </div>
    <div class="gain-item">
      <span class="gain-label" data-i18n="card.kd">${t('card.kd')}</span>
      <span class="gain-val kd-val">—</span>
    </div>
  </div>
</div>
<div class="card-resize-handle" title="Drag to resize"></div>`;

    container.appendChild(card);

    // Wire chart
    const canvas = card.querySelector('.trend-canvas');
    const chart = createTrendChart(canvas, displayName);
    charts[name] = chart;

    // Chart toolbar
    card.querySelector('.chart-btn-fs').addEventListener('click', () => chart.enterFullscreen());
    card.querySelector('.chart-btn-reset').addEventListener('click', () => chart.resetView());
    card.querySelector('.chart-btn-png').addEventListener('click', () => chart.exportPNG());
    card.querySelector('.chart-btn-pdf').addEventListener('click', () => chart.exportPDF());

    // Setpoint
    const spInput = card.querySelector('.sp-input');
    card.querySelector('.sp-btn').addEventListener('click', () => {
        const v = parseFloat(spInput.value);
        if (!isNaN(v)) postCommand('setpoint', { loop: name, setpoint: v });
    });
    spInput.addEventListener('keydown', e => {
        if (e.key === 'Enter') {
            const v = parseFloat(spInput.value);
            if (!isNaN(v)) postCommand('setpoint', { loop: name, setpoint: v });
        }
    });

    // Mode
    const modeSelect = card.querySelector('.mode-select');
    card.querySelector('.mode-btn').addEventListener('click', () => {
        postCommand('mode', { loop: name, mode: modeSelect.value });
    });

    // Disturbance and reset — use hyphens matching the Go route registration
    card.querySelector('.disturbance-btn').addEventListener('click', () => {
        postCommand('inject-disturbance', { loop: name, amplitude: 5, duration_seconds: 30 });
    });
    card.querySelector('.reset-btn').addEventListener('click', () => {
        postCommand('reset-loop', { loop: name });
    });

    // Gains toggle
    const gainsToggle = card.querySelector('.gains-toggle');
    const gainsPanel  = card.querySelector('.gains-panel');
    gainsToggle.addEventListener('click', () => {
        const open = gainsPanel.classList.toggle('visible');
        gainsToggle.classList.toggle('open', open);
    });

    // PID gains are directly on LoopSnapshot (kp, ki, kd — not nested under .pid)
    card.querySelector('.kp-val').textContent = loop.kp ?? '—';
    card.querySelector('.ki-val').textContent = loop.ki ?? '—';
    card.querySelector('.kd-val').textContent = loop.kd ?? '—';

    // Per-card horizontal resize handle
    initCardResize(card, card.querySelector('.card-resize-handle'));

    return card;
}

function updateLoopCard(card, loop) {
    const q = s => card.querySelector(s);

    q('.pv-val').textContent = loop.pv != null ? loop.pv.toFixed(2) : '—';
    q('.sp-val').textContent = loop.sp != null ? loop.sp.toFixed(2) : '—';
    q('.mv-val').textContent = loop.mv != null ? loop.mv.toFixed(1) : '—';

    const badge = q('.loop-mode-badge');
    if (badge) {
        // mode is e.g. "auto" → translation key "card.auto"
        badge.textContent = t('card.' + (loop.mode || 'auto')) || (loop.mode || '');
        badge.className = 'loop-mode-badge mode-' + (loop.mode || '');
    }

    const sel = q('.mode-select');
    if (sel && document.activeElement !== sel) sel.value = loop.mode || 'auto';

    // Keep PID gains in sync (they can be changed via MQTT while the page is open).
    if (loop.kp != null) q('.kp-val').textContent = loop.kp;
    if (loop.ki != null) q('.ki-val').textContent = loop.ki;
    if (loop.kd != null) q('.kd-val').textContent = loop.kd;
}

// ── Global drag state — one pair of window listeners for ALL resize handles ──
// Each handle sets _activeDrag on mousedown; the global handlers dispatch to it.

let _activeDrag = null;

window.addEventListener('mousemove', e => {
    if (_activeDrag) _activeDrag.move(e);
});
window.addEventListener('mouseup', e => {
    if (!_activeDrag || e.button !== 0) return;
    _activeDrag.end(e);
    _activeDrag = null;
    document.body.style.cursor = '';
    document.body.style.userSelect = '';
});

// ── Per-card horizontal resize ───────────────────────────────────────────────

function initCardResize(card, handle) {
    if (!handle) return;
    handle.addEventListener('mousedown', e => {
        if (e.button !== 0) return;
        e.preventDefault();
        const startX = e.clientX;
        const startW = card.offsetWidth;
        document.body.style.cursor = 'ew-resize';
        document.body.style.userSelect = 'none';
        _activeDrag = {
            move: ev => {
                const newW = Math.max(300, Math.min(900, startW + (ev.clientX - startX)));
                card.style.width = newW + 'px';
            },
            end: () => {},
        };
    });
}

// ── Event terminal ──────────────────────────────────────────────────────────
// plc.Event JSON: {timestamp, level, event_type, message, details}

function appendEvent(ev) {
    const terminal = document.getElementById('event-terminal');
    if (!terminal) return;

    const placeholder = terminal.querySelector('.events-empty');
    if (placeholder) placeholder.remove();

    const entry = document.createElement('div');
    const level = ev.level || 'info';
    entry.className = 'event-entry level-' + level;
    const ts = ev.timestamp ? new Date(ev.timestamp).toLocaleTimeString() : '—';
    const type = ev.event_type ? `[${ev.event_type}] ` : '';
    entry.textContent = `[${ts}] ${type}${ev.message || ''}`;
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
            const msg = err.error || `HTTP ${res.status}`;
            console.warn('command rejected:', msg);
            showToast(msg);
        }
    } catch (e) {
        console.error('command error:', e);
        showToast(String(e));
    }
}

function showToast(msg) {
    const el = document.createElement('div');
    el.className = 'cmd-toast';
    el.textContent = msg;
    document.body.appendChild(el);
    setTimeout(() => el.remove(), 3500);
}

// ── Panel resize handle ──────────────────────────────────────────────────────

function initResizeHandle() {
    const handle  = document.getElementById('resize-handle');
    const evPanel = document.getElementById('events-panel');
    if (!handle || !evPanel) return;

    handle.addEventListener('mousedown', e => {
        if (e.button !== 0) return;
        const startX = e.clientX;
        const startW = evPanel.offsetWidth;
        handle.classList.add('dragging');
        document.body.style.cursor = 'col-resize';
        document.body.style.userSelect = 'none';
        _activeDrag = {
            move: ev => {
                const dx = startX - ev.clientX;
                const newW = Math.max(180, Math.min(600, startW + dx));
                evPanel.style.width = newW + 'px';
            },
            end: () => handle.classList.remove('dragging'),
        };
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
    renderManualTab(currentTab);
}

// ── Language toggle ───────────────────────────────────────────────────────────

function initLangToggle() {
    const btn = document.getElementById('btn-lang');
    if (!btn) return;

    setLanguage(getCurrentLang());

    btn.addEventListener('click', () => {
        const next = getCurrentLang() === 'en' ? 'ru' : 'en';
        setLanguage(next);
        renderManualTab(currentTab);
        const placeholder = document.querySelector('#event-terminal .events-empty');
        if (placeholder) placeholder.textContent = t('events.empty');
        // Refresh per-card mode badges after language change
        for (const card of Object.values(loopCards)) {
            const badge = card.querySelector('.loop-mode-badge');
            if (badge) {
                const mode = (badge.className.match(/mode-(\w+)/) || [])[1] || 'auto';
                badge.textContent = t('card.' + mode) || mode;
            }
        }
        // Refresh plc state label
        const stateEl = document.getElementById('plc-state');
        if (stateEl) {
            const cls = [...stateEl.classList].find(c => c !== 'state-badge') || '';
            updatePLCState(cls);
        }
    });
}

// ── Header buttons ────────────────────────────────────────────────────────────

function initHeaderButtons() {
    document.getElementById('btn-start')?.addEventListener('click', () => postCommand('start'));
    document.getElementById('btn-stop')?.addEventListener('click',  () => postCommand('stop'));
    document.getElementById('btn-clear-events')?.addEventListener('click', clearEvents);
}

// ── Utility ───────────────────────────────────────────────────────────────────

function escHtml(s) {
    return String(s)
        .replace(/&/g, '&amp;')
        .replace(/</g, '&lt;')
        .replace(/>/g, '&gt;')
        .replace(/"/g, '&quot;');
}

// ── Init ──────────────────────────────────────────────────────────────────────

window.addEventListener('DOMContentLoaded', () => {
    initLangToggle();
    initHeaderButtons();
    initManualDrawer();
    initResizeHandle();
    loadStatus();
    initSSE();
    uptimeInterval = setInterval(tickUptime, 1000);
});
