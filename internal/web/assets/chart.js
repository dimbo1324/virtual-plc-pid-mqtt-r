// chart.js — Canvas trend chart for PV / SP / MV.
'use strict';

const CHART_MAX_POINTS = 300;

function createTrendChart(canvas) {
    const ctx = canvas.getContext('2d');
    const series = {
        pv: { label: 'PV', color: '#4fc3f7', data: [] },
        sp: { label: 'SP', color: '#ffb74d', data: [] },
        mv: { label: 'MV', color: '#81c784', data: [] },
    };

    function push(pv, sp, mv) {
        series.pv.data.push(pv);
        series.sp.data.push(sp);
        series.mv.data.push(mv);
        for (const s of Object.values(series)) {
            if (s.data.length > CHART_MAX_POINTS) s.data.shift();
        }
        draw();
    }

    function draw() {
        const w = canvas.width;
        const h = canvas.height;
        ctx.clearRect(0, 0, w, h);

        ctx.fillStyle = '#0d1117';
        ctx.fillRect(0, 0, w, h);

        // Determine value range across all series.
        let min = Infinity;
        let max = -Infinity;
        for (const s of Object.values(series)) {
            for (const v of s.data) {
                if (v < min) min = v;
                if (v > max) max = v;
            }
        }
        if (!isFinite(min)) { min = 0; max = 100; }
        if (min === max) { min -= 1; max += 1; }
        const range = max - min;

        const padX = 4;
        const padY = 14; // bottom reserve for legend
        const chartW = w - 2 * padX;
        const chartH = h - padY - 4;

        function sx(i) { return padX + (i / (CHART_MAX_POINTS - 1)) * chartW; }
        function sy(v) { return 4 + chartH - ((v - min) / range) * chartH; }

        // Grid line at midpoint.
        ctx.strokeStyle = '#1e2d3d';
        ctx.lineWidth = 1;
        ctx.beginPath();
        const midY = sy((min + max) / 2);
        ctx.moveTo(padX, midY);
        ctx.lineTo(padX + chartW, midY);
        ctx.stroke();

        // Draw each series.
        for (const s of Object.values(series)) {
            if (s.data.length < 2) continue;
            ctx.strokeStyle = s.color;
            ctx.lineWidth = 1.5;
            ctx.beginPath();
            for (let i = 0; i < s.data.length; i++) {
                const x = sx(i);
                const y = sy(s.data[i]);
                if (i === 0) ctx.moveTo(x, y); else ctx.lineTo(x, y);
            }
            ctx.stroke();
        }

        // Legend.
        ctx.font = '9px monospace';
        let lx = padX;
        const ly = h - 2;
        for (const s of Object.values(series)) {
            ctx.fillStyle = s.color;
            ctx.fillText(s.label, lx, ly);
            lx += 36;
        }
    }

    draw();
    return { push };
}
