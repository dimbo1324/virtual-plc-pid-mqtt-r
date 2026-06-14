// chart.js — Enhanced trend chart: zoom, pan, fullscreen, PNG/PDF export.
'use strict';

const CHART_MAX_POINTS = 600;
const CHART_COLORS = {
    pv: '#4fc3f7',
    sp: '#ffb74d',
    mv: '#81c784',
    grid: '#1e2d3d',
    bg: '#0d1117',
    text: '#8b949e',
    crosshair: 'rgba(255,255,255,0.25)',
};

function createTrendChart(canvas, loopName) {
    const ctx = canvas.getContext('2d');
    const series = {
        pv: { label: 'PV', color: CHART_COLORS.pv, data: [] },
        sp: { label: 'SP', color: CHART_COLORS.sp, data: [] },
        mv: { label: 'MV', color: CHART_COLORS.mv, data: [] },
    };

    // Zoom/pan state: viewStart and viewEnd are indices into the data array.
    // null means "show all data" (auto-scroll).
    let viewStart = null;
    let viewEnd = null;
    let isDragging = false;
    let dragStartX = 0;
    let dragStartView = null;
    let mouseX = -1;

    // Fullscreen state
    let fsOverlay = null;
    let fsCanvas = null;
    let fsCtx = null;
    let isFullscreen = false;
    let resizeObserver = null;

    function dataLen() { return series.pv.data.length; }

    // Returns [start, end] indices for the current view (clamped to data).
    function getView() {
        const len = dataLen();
        if (len === 0) return [0, 0];
        const s = viewStart === null ? 0 : Math.max(0, Math.min(viewStart, len - 1));
        const e = viewEnd === null ? len : Math.max(s + 1, Math.min(viewEnd, len));
        return [s, e];
    }

    function isAutoScroll() { return viewEnd === null; }

    function resetView() {
        viewStart = null;
        viewEnd = null;
        draw();
    }

    function push(pv, sp, mv) {
        series.pv.data.push(pv);
        series.sp.data.push(sp);
        series.mv.data.push(mv);
        const excess = series.pv.data.length - CHART_MAX_POINTS;
        if (excess > 0) {
            for (const s of Object.values(series)) s.data.splice(0, excess);
            // Shift view indices to track data after trim.
            if (viewStart !== null) viewStart = Math.max(0, viewStart - excess);
            if (viewEnd !== null) viewEnd = Math.max(1, viewEnd - excess);
        }
        draw();
    }

    // ---- Drawing ----

    function drawOnContext(targetCtx, w, h) {
        targetCtx.clearRect(0, 0, w, h);
        targetCtx.fillStyle = CHART_COLORS.bg;
        targetCtx.fillRect(0, 0, w, h);

        const padL = 48; // Y-axis labels
        const padR = 12;
        const padT = 10;
        const padB = 28; // legend

        const chartW = w - padL - padR;
        const chartH = h - padT - padB;

        const [vs, ve] = getView();
        const visible = { pv: [], sp: [], mv: [] };
        for (let i = vs; i < ve; i++) {
            visible.pv.push(series.pv.data[i]);
            visible.sp.push(series.sp.data[i]);
            visible.mv.push(series.mv.data[i]);
        }

        // Y range from visible data.
        let yMin = Infinity, yMax = -Infinity;
        for (const arr of Object.values(visible)) {
            for (const v of arr) {
                if (v < yMin) yMin = v;
                if (v > yMax) yMax = v;
            }
        }
        if (!isFinite(yMin)) { yMin = 0; yMax = 100; }
        if (yMin === yMax) { yMin -= 1; yMax += 1; }
        const yRange = yMax - yMin;
        yMin -= yRange * 0.05;
        yMax += yRange * 0.05;
        const yRangePadded = yMax - yMin;

        function sx(i) {
            const n = visible.pv.length;
            if (n <= 1) return padL + chartW / 2;
            return padL + (i / (n - 1)) * chartW;
        }
        function sy(v) {
            return padT + chartH - ((v - yMin) / yRangePadded) * chartH;
        }

        // Clip to chart area
        targetCtx.save();
        targetCtx.beginPath();
        targetCtx.rect(padL, padT, chartW, chartH);
        targetCtx.clip();

        // Grid lines (4 horizontal)
        targetCtx.strokeStyle = CHART_COLORS.grid;
        targetCtx.lineWidth = 1;
        for (let g = 0; g <= 4; g++) {
            const gy = padT + (g / 4) * chartH;
            targetCtx.beginPath();
            targetCtx.moveTo(padL, gy);
            targetCtx.lineTo(padL + chartW, gy);
            targetCtx.stroke();
        }

        // Series lines
        for (const [key, s] of Object.entries(series)) {
            const arr = visible[key];
            if (arr.length < 2) continue;
            targetCtx.strokeStyle = s.color;
            targetCtx.lineWidth = 2;
            targetCtx.lineJoin = 'round';
            targetCtx.beginPath();
            for (let i = 0; i < arr.length; i++) {
                const x = sx(i);
                const y = sy(arr[i]);
                if (i === 0) targetCtx.moveTo(x, y); else targetCtx.lineTo(x, y);
            }
            targetCtx.stroke();
        }

        // Crosshair on mouse hover
        if (mouseX >= padL && mouseX <= padL + chartW) {
            targetCtx.strokeStyle = CHART_COLORS.crosshair;
            targetCtx.lineWidth = 1;
            targetCtx.setLineDash([4, 4]);
            targetCtx.beginPath();
            targetCtx.moveTo(mouseX, padT);
            targetCtx.lineTo(mouseX, padT + chartH);
            targetCtx.stroke();
            targetCtx.setLineDash([]);
        }

        targetCtx.restore();

        // Y-axis labels
        targetCtx.fillStyle = CHART_COLORS.text;
        targetCtx.font = '10px monospace';
        targetCtx.textAlign = 'right';
        for (let g = 0; g <= 4; g++) {
            const v = yMax - (g / 4) * yRangePadded;
            const gy = padT + (g / 4) * chartH;
            targetCtx.fillText(v.toFixed(1), padL - 4, gy + 3);
        }
        targetCtx.textAlign = 'left';

        // Legend at bottom
        const legendY = h - 6;
        targetCtx.font = '11px monospace';
        let lx = padL;
        for (const s of Object.values(series)) {
            targetCtx.fillStyle = s.color;
            targetCtx.fillRect(lx, legendY - 8, 14, 3);
            targetCtx.fillStyle = CHART_COLORS.text;
            targetCtx.fillText(s.label, lx + 18, legendY);
            lx += 60;
        }

        // Zoom indicator bar showing position of view window within full data
        const len = dataLen();
        if (len > 0 && !isAutoScroll()) {
            const [s, e] = getView();
            const barX = padL;
            const barW = chartW;
            const barY = h - 4;
            targetCtx.fillStyle = '#1a2535';
            targetCtx.fillRect(barX, barY, barW, 3);
            targetCtx.fillStyle = '#4fc3f7';
            targetCtx.fillRect(barX + (s / len) * barW, barY, ((e - s) / len) * barW, 3);
        }
    }

    function draw() {
        drawOnContext(ctx, canvas.width, canvas.height);
        if (isFullscreen && fsCtx && fsCanvas) {
            drawOnContext(fsCtx, fsCanvas.width, fsCanvas.height);
        }
    }

    // ---- Zoom & Pan ----

    function canvasXToDataIdx(x) {
        const padL = 48;
        const padR = 12;
        const chartW = canvas.width - padL - padR;
        const ratio = Math.max(0, Math.min(1, (x - padL) / chartW));
        const [vs, ve] = getView();
        return vs + ratio * (ve - vs);
    }

    function zoom(factor, pivotX) {
        const len = dataLen();
        if (len === 0) return;
        const [vs, ve] = getView();
        const pivotIdx = canvasXToDataIdx(pivotX);
        const halfSpan = (ve - vs) / 2;
        const newHalfSpan = Math.max(10, Math.min(len, halfSpan * factor));
        const leftFrac = (pivotIdx - vs) / (ve - vs || 1);
        let ns = pivotIdx - newHalfSpan * leftFrac;
        let ne = ns + newHalfSpan * 2;
        if (ns < 0) { ne -= ns; ns = 0; }
        if (ne > len) { ns -= (ne - len); ne = len; }
        ns = Math.max(0, ns);
        ne = Math.min(len, ne);
        viewStart = Math.round(ns);
        viewEnd = Math.round(ne);
        if (viewStart <= 0 && viewEnd >= len) resetView();
        else draw();
    }

    canvas.addEventListener('wheel', (e) => {
        e.preventDefault();
        const factor = e.deltaY > 0 ? 1.2 : 0.833;
        zoom(factor, e.offsetX);
    }, { passive: false });

    canvas.addEventListener('mousedown', (e) => {
        if (e.button !== 0) return;
        isDragging = true;
        dragStartX = e.offsetX;
        dragStartView = getView();
        canvas.style.cursor = 'grabbing';
    });

    canvas.addEventListener('mousemove', (e) => {
        mouseX = e.offsetX;
        if (isDragging) {
            const padL = 48, padR = 12;
            const chartW = canvas.width - padL - padR;
            const [vs, ve] = dragStartView;
            const span = ve - vs;
            const dx = e.offsetX - dragStartX;
            const dIdx = -(dx / chartW) * span;
            let ns = vs + dIdx;
            let ne = ve + dIdx;
            const len = dataLen();
            if (ns < 0) { ne -= ns; ns = 0; }
            if (ne > len) { ns -= (ne - len); ne = len; }
            viewStart = Math.max(0, Math.round(ns));
            viewEnd = Math.min(len, Math.round(ne));
        }
        draw();
    });

    window.addEventListener('mouseup', (e) => {
        if (e.button !== 0 || !isDragging) return;
        isDragging = false;
        canvas.style.cursor = 'crosshair';
    });

    canvas.addEventListener('mouseleave', () => {
        mouseX = -1;
        draw();
    });

    canvas.style.cursor = 'crosshair';

    // ---- ResizeObserver ----

    function fitCanvas(c, targetCtx) {
        const rect = c.parentElement.getBoundingClientRect();
        const dpr = window.devicePixelRatio || 1;
        const w = Math.max(100, rect.width);
        const h = Math.max(60, rect.height);
        if (c.width !== Math.round(w * dpr) || c.height !== Math.round(h * dpr)) {
            c.width = Math.round(w * dpr);
            c.height = Math.round(h * dpr);
            c.style.width = w + 'px';
            c.style.height = h + 'px';
            targetCtx.setTransform(dpr, 0, 0, dpr, 0, 0);
        }
    }

    resizeObserver = new ResizeObserver(() => { fitCanvas(canvas, ctx); draw(); });
    resizeObserver.observe(canvas.parentElement);
    fitCanvas(canvas, ctx);

    // ---- Fullscreen ----

    function enterFullscreen() {
        if (isFullscreen) return;
        isFullscreen = true;

        fsOverlay = document.createElement('div');
        fsOverlay.className = 'chart-fs-overlay';

        const header = document.createElement('div');
        header.className = 'chart-fs-header';

        const title = document.createElement('span');
        title.className = 'chart-fs-title';
        title.textContent = loopName || '';

        const closeBtn = document.createElement('button');
        closeBtn.className = 'chart-fs-close btn-icon';
        closeBtn.textContent = '✕';
        closeBtn.addEventListener('click', exitFullscreen);

        header.append(title, closeBtn);

        fsCanvas = document.createElement('canvas');
        fsCanvas.className = 'chart-fs-canvas';
        fsCtx = fsCanvas.getContext('2d');

        fsOverlay.append(header, fsCanvas);
        document.body.appendChild(fsOverlay);

        const ro = new ResizeObserver(() => { fitCanvas(fsCanvas, fsCtx); draw(); });
        ro.observe(fsCanvas.parentElement);
        fitCanvas(fsCanvas, fsCtx);

        fsCanvas.addEventListener('wheel', (e) => {
            e.preventDefault();
            const factor = e.deltaY > 0 ? 1.2 : 0.833;
            // Map fs canvas x to main canvas x for zoom pivot
            const ratio = canvas.width / (fsCanvas.width || 1);
            zoom(factor, e.offsetX * ratio);
        }, { passive: false });
        fsCanvas.style.cursor = 'crosshair';

        document.addEventListener('keydown', onFsKey);
        fsOverlay.fsRO = ro;
        draw();
    }

    function exitFullscreen() {
        if (!isFullscreen) return;
        isFullscreen = false;
        document.removeEventListener('keydown', onFsKey);
        if (fsOverlay) {
            if (fsOverlay.fsRO) fsOverlay.fsRO.disconnect();
            fsOverlay.remove();
            fsOverlay = null;
            fsCanvas = null;
            fsCtx = null;
        }
    }

    function onFsKey(e) {
        if (e.key === 'Escape') exitFullscreen();
    }

    // ---- Export ----

    function exportPNG() {
        const oc = document.createElement('canvas');
        const scale = 2;
        oc.width = canvas.width * scale;
        oc.height = canvas.height * scale;
        const ocCtx = oc.getContext('2d');
        ocCtx.scale(scale, scale);
        const savedMX = mouseX;
        mouseX = -1;
        drawOnContext(ocCtx, canvas.width, canvas.height);
        mouseX = savedMX;
        const url = oc.toDataURL('image/png');
        const a = document.createElement('a');
        a.href = url;
        a.download = (loopName || 'chart') + '_trend.png';
        a.click();
    }

    function exportPDF() {
        const oc = document.createElement('canvas');
        oc.width = canvas.width;
        oc.height = canvas.height;
        const ocCtx = oc.getContext('2d');
        const savedMX = mouseX;
        mouseX = -1;
        drawOnContext(ocCtx, oc.width, oc.height);
        mouseX = savedMX;
        const url = oc.toDataURL('image/png');
        const win = window.open('', '_blank');
        win.document.write(`<!DOCTYPE html><html><head><title>${loopName || 'chart'} Trend</title>
<style>body{margin:0;background:#000}img{width:100%;display:block}
@media print{img{page-break-inside:avoid}}</style></head>
<body><img src="${url}"/><script>window.onload=function(){window.print();}<\/script></body></html>`);
        win.document.close();
    }

    draw();

    return {
        push,
        resetView,
        enterFullscreen,
        exitFullscreen,
        exportPNG,
        exportPDF,
        destroy() {
            if (resizeObserver) resizeObserver.disconnect();
            exitFullscreen();
        }
    };
}
