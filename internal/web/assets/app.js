// app.js — Virtual PLC Dashboard application logic.
'use strict';

const charts = {};
const loopCards = {};

// ── SSE connection ──────────────────────────────────────────────────────────

function initSSE() {
    const es = new EventSource('/api/stream');
    const dot = document.getElementById('connection-dot');

    es.addEventListener('snapshot', e => {
        try { handleSnapshot(JSON.parse(e.data)); } catch (_) {}
    });

    es.addEventListener('plc_event', e => {
        try { appendEvent(JSON.parse(e.data)); } catch (_) {}
    });

    es.addEventListener('heartbeat', () => {
        if (dot) dot.className = 'dot connected';
    });

    es.onerror = () => {
        if (dot) dot.className = 'dot';
    };
}

// ── Snapshot handler ────────────────────────────────────────────────────────

function handleSnapshot(snap) {
    updateHeader(snap);
    const loops = snap.loops || {};
    const section = document.getElementById('loops-section');
    for (const [name, loop] of Object.entries(loops)) {
        if (!loopCards[name]) {
            loopCards[name] = createLoopCard(section, name, loop);
        }
        updateLoopCard(loopCards[name], loop);
        if (charts[name]) {
            charts[name].push(loop.pv, loop.sp, loop.mv);
        }
    }
}

function updateHeader(snap) {
    const el = id => document.getElementById(id);
    if (el('device-id')) el('device-id').textContent = snap.device_id || '—';
    const stateEl = el('plc-state');
    if (stateEl) {
        const state = snap.plc?.state || '—';
        stateEl.textContent = state;
        stateEl.className = 'state-badge ' + state;
    }
}

// ── Loop cards ──────────────────────────────────────────────────────────────

function createLoopCard(container, name, loop) {
    const card = document.createElement('div');
    card.className = 'loop-card';
    card.innerHTML = `
        <div class="loop-header">
            <span class="loop-name">${loop.display_name || name}</span>
            <span class="loop-unit">${loop.unit || ''}</span>
            <span class="loop-mode-badge"></span>
        </div>
        <div class="loop-values">
            <div class="val-block">
                <span class="val-label">PV</span>
                <span class="val-num pv-val">—</span>
            </div>
            <div class="val-block">
                <span class="val-label">SP</span>
                <span class="val-num sp-val">—</span>
            </div>
            <div class="val-block">
                <span class="val-label">MV</span>
                <span class="val-num mv-val">—</span>
                <span class="val-unit">%</span>
            </div>
        </div>
        <canvas class="trend-canvas" width="360" height="110"></canvas>
        <div class="loop-controls">
            <div class="ctrl-row">
                <input class="ctrl-input sp-input" type="number" placeholder="Setpoint" step="0.1">
                <button class="ctrl-btn" onclick="cmdSetpoint('${name}', this.previousElementSibling)">Set SP</button>
            </div>
            <div class="ctrl-row">
                <select class="ctrl-select mode-select">
                    <option value="auto">Auto</option>
                    <option value="manual">Manual</option>
                    <option value="hold">Hold</option>
                    <option value="disabled">Disabled</option>
                </select>
                <button class="ctrl-btn" onclick="cmdMode('${name}', this.previousElementSibling)">Set Mode</button>
            </div>
            <div class="ctrl-row">
                <button class="ctrl-btn danger" onclick="cmdReset('${name}')">Reset Loop</button>
                <button class="ctrl-btn accent" onclick="cmdDisturbance('${name}')">Disturb +5</button>
            </div>
        </div>`;

    container.appendChild(card);
    charts[name] = createTrendChart(card.querySelector('.trend-canvas'));
    return card;
}

function updateLoopCard(card, loop) {
    const q = s => card.querySelector(s);
    q('.pv-val').textContent = loop.pv != null ? loop.pv.toFixed(2) : '—';
    q('.sp-val').textContent = loop.sp != null ? loop.sp.toFixed(2) : '—';
    q('.mv-val').textContent = loop.mv != null ? loop.mv.toFixed(1) : '—';

    const badge = q('.loop-mode-badge');
    if (badge) { badge.textContent = loop.mode || ''; badge.className = 'loop-mode-badge mode-' + loop.mode; }

    const modeSelect = q('.mode-select');
    if (modeSelect && document.activeElement !== modeSelect) {
        modeSelect.value = loop.mode || 'auto';
    }
}

// ── Event terminal ──────────────────────────────────────────────────────────

function appendEvent(ev) {
    const terminal = document.getElementById('event-terminal');
    if (!terminal) return;
    const entry = document.createElement('div');
    const level = ev.level || 'info';
    entry.className = 'event-entry level-' + level;
    const ts = ev.timestamp ? new Date(ev.timestamp).toLocaleTimeString() : '—';
    entry.textContent = `[${ts}] [${level.toUpperCase()}] ${ev.message || ''}`;
    terminal.appendChild(entry);
    while (terminal.children.length > 200) terminal.removeChild(terminal.firstChild);
    terminal.scrollTop = terminal.scrollHeight;
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

function cmdStart()  { postCommand('start'); }
function cmdStop()   { postCommand('stop'); }

function cmdSetpoint(loop, inputEl) {
    const sp = parseFloat(inputEl.value);
    if (isNaN(sp)) return;
    postCommand('setpoint', { loop, setpoint: sp });
}

function cmdMode(loop, selectEl) {
    postCommand('mode', { loop, mode: selectEl.value });
}

function cmdReset(loop) {
    postCommand('reset-loop', { loop });
}

function cmdDisturbance(loop) {
    postCommand('inject-disturbance', { loop, amplitude: 5, duration_seconds: 30 });
}

// ── Init ────────────────────────────────────────────────────────────────────

async function loadStatus() {
    try {
        const res = await fetch('/api/status');
        if (!res.ok) return;
        const s = await res.json();
        const el = id => document.getElementById(id);
        if (el('device-id')) el('device-id').textContent = s.device_id || '—';
        if (el('plc-state')) {
            el('plc-state').textContent = s.state || '—';
            el('plc-state').className = 'state-badge ' + s.state;
        }
        if (el('uptime')) el('uptime').textContent = s.uptime || '—';
    } catch (_) {}
}

async function loadRecentEvents() {
    try {
        const res = await fetch('/api/events/recent?limit=50');
        if (!res.ok) return;
        const events = await res.json();
        for (const ev of events.slice().reverse()) appendEvent(ev);
    } catch (_) {}
}

window.addEventListener('DOMContentLoaded', () => {
    loadStatus();
    loadRecentEvents();
    initSSE();
});
