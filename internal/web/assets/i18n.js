// i18n.js — Bilingual (EN/RU) translation registry and manual content.
'use strict';

const TRANSLATIONS = {
    en: {
        // Header
        'app.title': 'Virtual PLC',
        'header.start': 'Start',
        'header.stop': 'Stop',
        'header.manual': 'Manual',
        'header.lang': 'RU',
        'header.conn.ok': 'Connected',
        'header.conn.lost': 'Disconnected',
        'header.state.running': 'RUNNING',
        'header.state.stopped': 'STOPPED',
        'header.state.unknown': 'UNKNOWN',
        'header.storage.degraded': 'STORAGE DEGRADED',
        'header.storage.degraded.title': 'SQLite unavailable — running in JSONL fallback mode. Historical telemetry is not being recorded.',
        // Loop card
        'card.mode': 'Mode',
        'card.auto': 'AUTO',
        'card.manual': 'MANUAL',
        'card.hold': 'HOLD',
        'card.pv': 'PV',
        'card.sp': 'SP',
        'card.mv': 'MV',
        'card.setpoint': 'Setpoint',
        'card.set': 'Set',
        'card.disturbance': 'Disturbance',
        'card.reset': 'Reset',
        'card.gains': 'PID Gains',
        'card.kp': 'Kp',
        'card.ki': 'Ki',
        'card.kd': 'Kd',
        // Chart
        'chart.zoom': 'Scroll to zoom · Drag to pan',
        'chart.fullscreen': 'Fullscreen',
        'chart.png': 'PNG',
        'chart.pdf': 'PDF',
        'chart.reset': 'Reset view',
        // Events
        'events.title': 'Event Log',
        'events.clear': 'Clear',
        'events.empty': 'No events yet',
        // Manual drawer
        'manual.title': 'User Manual',
        'manual.close': 'Close',
        'manual.tab.overview': 'Overview',
        'manual.tab.plc': 'PLC Theory',
        'manual.tab.pid': 'PID Theory',
        'manual.tab.usage': 'Usage Guide',
        'manual.tab.api': 'API Reference',
    },
    ru: {
        'app.title': 'Виртуальный ПЛК',
        'header.start': 'Запуск',
        'header.stop': 'Стоп',
        'header.manual': 'Справка',
        'header.lang': 'EN',
        'header.conn.ok': 'Подключено',
        'header.conn.lost': 'Соединение потеряно',
        'header.state.running': 'РАБОТАЕТ',
        'header.state.stopped': 'ОСТАНОВЛЕН',
        'header.state.unknown': 'НЕИЗВЕСТНО',
        'header.storage.degraded': 'ХРАНИЛИЩЕ НЕДОСТУПНО',
        'header.storage.degraded.title': 'SQLite недоступен — работа в режиме резервного JSONL. Историческая телеметрия не записывается.',
        'card.mode': 'Режим',
        'card.auto': 'АВТО',
        'card.manual': 'РУЧНОЙ',
        'card.hold': 'УДЕРЖАНИЕ',
        'card.pv': 'ПЗ',
        'card.sp': 'ЗД',
        'card.mv': 'РВ',
        'card.setpoint': 'Задание',
        'card.set': 'Уст.',
        'card.disturbance': 'Возмущение',
        'card.reset': 'Сброс',
        'card.gains': 'Коэфф. ПИД',
        'card.kp': 'Кп',
        'card.ki': 'Ки',
        'card.kd': 'Кд',
        'chart.zoom': 'Колёсиком — масштаб · Перетащить — сдвиг',
        'chart.fullscreen': 'На весь экран',
        'chart.png': 'PNG',
        'chart.pdf': 'PDF',
        'chart.reset': 'Сброс вида',
        'events.title': 'Журнал событий',
        'events.clear': 'Очистить',
        'events.empty': 'Событий нет',
        'manual.title': 'Руководство пользователя',
        'manual.close': 'Закрыть',
        'manual.tab.overview': 'Обзор',
        'manual.tab.plc': 'Теория ПЛК',
        'manual.tab.pid': 'Теория ПИД',
        'manual.tab.usage': 'Руководство',
        'manual.tab.api': 'API',
    }
};

const MANUAL_CONTENT = {
    en: {
        overview: `
<h2>Virtual PLC — Overview</h2>
<p>Virtual PLC is a software simulation of an industrial Programmable Logic Controller (PLC) with integrated PID control loops. It is designed for education, algorithm testing, and control system prototyping without physical hardware.</p>
<h3>Architecture</h3>
<ul>
  <li><strong>PLC Runtime</strong> — cyclic scan engine that evaluates all control loops at a configurable interval (default 100 ms).</li>
  <li><strong>PID Controller</strong> — standard ISA-form controller with anti-windup and bumpless mode transfer.</li>
  <li><strong>Process Simulator</strong> — first-order lag + dead-time model (FOPDT) approximating a real industrial process.</li>
  <li><strong>MQTT Interface</strong> — publishes loop telemetry and accepts commands over MQTT (compatible with any broker).</li>
  <li><strong>Storage</strong> — SQLite database in WAL mode for persistent event and trend logging.</li>
  <li><strong>Web Dashboard</strong> — this UI; served by the embedded HTTP server; updated via Server-Sent Events (SSE).</li>
</ul>
<h3>Key Features</h3>
<ul>
  <li>Multiple independent control loops, each with its own PID parameters and process model.</li>
  <li>Auto / Manual / Cascade operating modes.</li>
  <li>Real-time trend charts with zoom, pan, and export.</li>
  <li>Disturbance injection to test controller robustness.</li>
  <li>Full bilingual UI (English / Russian).</li>
</ul>`,
        plc: `
<h2>PLC Theory</h2>
<p>A <strong>Programmable Logic Controller (PLC)</strong> is a ruggedised digital computer used for automation of industrial processes. Unlike a general-purpose computer, a PLC executes its program in a deterministic cyclic fashion called the <em>scan cycle</em>.</p>
<h3>Scan Cycle</h3>
<ol>
  <li><strong>Input Scan</strong> — reads all physical inputs (sensors, switches) into the input image table.</li>
  <li><strong>Program Execution</strong> — runs the user program (ladder logic, function block, or structured text) against the input image to compute outputs.</li>
  <li><strong>Output Scan</strong> — writes the output image table to physical outputs (actuators, valves).</li>
  <li><strong>Housekeeping</strong> — communication with HMI/SCADA, diagnostics, watchdog reset.</li>
</ol>
<p>Scan times typically range from 1 ms (fast motion) to 100 ms (process control). Virtual PLC uses a configurable scan interval (default 100 ms).</p>
<h3>Operating Modes</h3>
<table>
  <thead><tr><th>Mode</th><th>SP source</th><th>MV source</th></tr></thead>
  <tbody>
    <tr><td>AUTO</td><td>Operator setpoint</td><td>PID output</td></tr>
    <tr><td>MANUAL</td><td>Operator setpoint</td><td>Operator direct output</td></tr>
    <tr><td>CASCADE</td><td>External master loop output</td><td>PID output</td></tr>
  </tbody>
</table>
<h3>Bumpless Transfer</h3>
<p>When switching from MANUAL to AUTO, the integrator state is pre-loaded with the current manual output so the controller output does not "bump" discontinuously. Virtual PLC implements this by initialising the integral term to <code>MV_manual − Kp×e</code> at mode change.</p>`,
        pid: `
<h2>PID Theory</h2>
<p>The <strong>PID controller</strong> (Proportional–Integral–Derivative) is the most widely used feedback controller in industry, regulating over 90 % of industrial control loops.</p>
<h3>Standard (ISA) Form</h3>
<pre>u(t) = Kp · [ e(t) + (1/Ti)·∫e dt + Td·(de/dt) ]</pre>
<p>where:</p>
<ul>
  <li><strong>e(t) = SP − PV</strong> — control error</li>
  <li><strong>Kp</strong> — proportional gain (dimensionless)</li>
  <li><strong>Ti = Kp/Ki</strong> — integral time (s); eliminates steady-state offset</li>
  <li><strong>Td = Kd/Kp</strong> — derivative time (s); damps fast disturbances</li>
</ul>
<h3>Discrete (Velocity) Form</h3>
<p>Virtual PLC uses the <em>position</em> form with a Tustin (bilinear) integrator:</p>
<pre>
P = Kp · e[k]
I[k] = I[k-1] + Ki · Ts/2 · (e[k] + e[k-1])   (anti-windup clamp applied here)
D = Kd · (e[k] − e[k-1]) / Ts
u[k] = P + I[k] + D
</pre>
<h3>Anti-windup</h3>
<p>When the controller output saturates (hits the 0–100 % limits), the integrator continues accumulating — this is <em>integral windup</em>. Virtual PLC clamps the integrator: if the saturated output equals the unsaturated output the integrator is allowed to accumulate; otherwise it is frozen at the clamping boundary.</p>
<h3>Tuning Guidelines</h3>
<table>
  <thead><tr><th>Parameter</th><th>Too low</th><th>Too high</th></tr></thead>
  <tbody>
    <tr><td>Kp</td><td>Slow response, large offset</td><td>Oscillation, instability</td></tr>
    <tr><td>Ki</td><td>Slow offset removal</td><td>Windup, overshoot</td></tr>
    <tr><td>Kd</td><td>Sluggish disturbance rejection</td><td>Amplifies noise, chatter</td></tr>
  </tbody>
</table>
<p>A good starting point for a first-order process is the <strong>Ziegler–Nichols step-response method</strong>: identify the process gain K, time constant τ, and dead time L from the open-loop step response, then set Kp = 1.2τ/(K·L), Ti = 2L, Td = L/2.</p>`,
        usage: `
<h2>Usage Guide</h2>
<h3>Starting and Stopping</h3>
<p>Use the <strong>Start</strong> / <strong>Stop</strong> buttons in the header. The status badge shows the current PLC state. All loops are started and stopped together.</p>
<h3>Loop Controls</h3>
<ul>
  <li><strong>Mode selector</strong> — switch between AUTO / MANUAL / CASCADE per loop.</li>
  <li><strong>Setpoint</strong> — type a value and click <em>Set</em> (or press Enter) to update the loop setpoint.</li>
  <li><strong>Disturbance</strong> — injects a step disturbance of ±5 % for 30 seconds into the process simulator, useful for testing controller robustness.</li>
  <li><strong>Reset</strong> — clears PID integrator state and returns the loop to its initial condition.</li>
  <li><strong>PID Gains</strong> — expand the gains panel to see Kp / Ki / Kd for each loop (read-only; edit in <code>configs/default.json</code>).</li>
</ul>
<h3>Charts</h3>
<ul>
  <li><strong>Zoom</strong> — scroll the mouse wheel over the chart to zoom in/out on the time axis.</li>
  <li><strong>Pan</strong> — hold the left mouse button and drag left/right to shift the view.</li>
  <li><strong>Reset view</strong> — click the ↺ button to return to the default zoom and scroll to the latest data.</li>
  <li><strong>Fullscreen</strong> — click ⛶ to expand the chart to the full browser window. Press Escape or click again to exit.</li>
  <li><strong>Export PNG</strong> — saves the chart as a raster image.</li>
  <li><strong>Export PDF</strong> — opens a print dialog; choose "Save as PDF" in your browser.</li>
</ul>
<h3>Layout Customisation</h3>
<p>Drag the vertical divider between the charts panel and the event log to resize both panels. The divider appears highlighted when you hover over it.</p>
<h3>Event Log</h3>
<p>The event log shows all PLC scan events in real time. Click <strong>Clear</strong> to empty the list. The log is limited to the last 200 entries.</p>`,
        api: `
<h2>API Reference</h2>
<p>The embedded HTTP server exposes a minimal REST API on the same port as the dashboard (default <code>:8090</code>).</p>
<h3>GET /api/stream</h3>
<p>Server-Sent Events stream. Emits the following event types:</p>
<table>
  <thead><tr><th>Event</th><th>Payload</th><th>Description</th></tr></thead>
  <tbody>
    <tr><td><code>snapshot</code></td><td>Full state object</td><td>Sent once on connection; full PLC + loop state.</td></tr>
    <tr><td><code>plc_event</code></td><td>Single scan event</td><td>Emitted every UI update interval with latest PV/SP/MV per loop.</td></tr>
    <tr><td><code>heartbeat</code></td><td><code>{"ts":"…"}</code></td><td>Keep-alive; emitted every 15 s.</td></tr>
  </tbody>
</table>
<h3>POST /api/commands/start</h3>
<p>Starts all PLC loops. Body: <code>{}</code>. Returns 200 on success.</p>
<h3>POST /api/commands/stop</h3>
<p>Stops all PLC loops. Body: <code>{}</code>. Returns 200 on success.</p>
<h3>POST /api/commands/setpoint</h3>
<p>Updates a loop setpoint. Body: <code>{"loop":"loop_name","value":42.0}</code>.</p>
<h3>POST /api/commands/mode</h3>
<p>Changes a loop mode. Body: <code>{"loop":"loop_name","mode":"auto"|"manual"|"hold"}</code>.</p>
<h3>POST /api/commands/manual_output</h3>
<p>Sets manual output value (MANUAL mode only). Body: <code>{"loop":"loop_name","value":50.0}</code>.</p>
<h3>POST /api/commands/inject_disturbance</h3>
<p>Injects a step disturbance. Body: <code>{"loop":"loop_name","amplitude":5.0,"duration_seconds":30}</code>.</p>
<h3>POST /api/commands/reset_loop</h3>
<p>Resets PID integrator. Body: <code>{"loop":"loop_name"}</code>.</p>`
    },
    ru: {
        overview: `
<h2>Виртуальный ПЛК — Обзор</h2>
<p>Виртуальный ПЛК — программная симуляция промышленного программируемого логического контроллера (ПЛК) с интегрированными контурами ПИД-регулирования. Предназначен для обучения, тестирования алгоритмов и прототипирования систем управления без физического оборудования.</p>
<h3>Архитектура</h3>
<ul>
  <li><strong>Среда выполнения ПЛК</strong> — циклический движок опроса, вычисляющий все контуры управления с заданным интервалом (по умолчанию 100 мс).</li>
  <li><strong>ПИД-регулятор</strong> — стандартный регулятор формы ISA с антинасыщением и безударным переключением режима.</li>
  <li><strong>Симулятор процесса</strong> — модель первого порядка с запаздыванием (FOPDT), аппроксимирующая реальный промышленный процесс.</li>
  <li><strong>MQTT-интерфейс</strong> — публикует телеметрию контуров и принимает команды по протоколу MQTT.</li>
  <li><strong>Хранилище</strong> — база данных SQLite в режиме WAL для постоянного журналирования событий и трендов.</li>
  <li><strong>Веб-панель</strong> — данный интерфейс; обслуживается встроенным HTTP-сервером; обновляется через Server-Sent Events (SSE).</li>
</ul>
<h3>Основные возможности</h3>
<ul>
  <li>Несколько независимых контуров управления, каждый со своими параметрами ПИД и моделью процесса.</li>
  <li>Режимы работы: Авто / Ручной / Каскад.</li>
  <li>Тренды в реальном времени с масштабированием, сдвигом и экспортом.</li>
  <li>Инжекция возмущений для проверки устойчивости регулятора.</li>
  <li>Полный двуязычный интерфейс (русский / английский).</li>
</ul>`,
        plc: `
<h2>Теория ПЛК</h2>
<p><strong>Программируемый логический контроллер (ПЛК)</strong> — специализированный цифровой компьютер для автоматизации промышленных процессов. В отличие от универсального компьютера ПЛК выполняет программу детерминированным циклическим образом — <em>цикл опроса</em>.</p>
<h3>Цикл опроса</h3>
<ol>
  <li><strong>Считывание входов</strong> — чтение всех физических входов (датчики, выключатели) в таблицу образов входов.</li>
  <li><strong>Выполнение программы</strong> — выполнение пользовательской программы (лестничная логика, функциональные блоки или структурированный текст) относительно образов входов для вычисления выходов.</li>
  <li><strong>Запись выходов</strong> — запись таблицы образов выходов на физические выходы (исполнительные механизмы, клапаны).</li>
  <li><strong>Обслуживание</strong> — связь с АРМ/SCADA, диагностика, сброс сторожевого таймера.</li>
</ol>
<p>Время цикла обычно от 1 мс (быстрое движение) до 100 мс (управление процессом). Виртуальный ПЛК использует настраиваемый интервал (по умолчанию 100 мс).</p>
<h3>Режимы работы</h3>
<table>
  <thead><tr><th>Режим</th><th>Источник задания</th><th>Источник выхода</th></tr></thead>
  <tbody>
    <tr><td>АВТО</td><td>Оператор</td><td>Выход ПИД</td></tr>
    <tr><td>РУЧНОЙ</td><td>Оператор</td><td>Прямой выход оператора</td></tr>
    <tr><td>КАСКАД</td><td>Выход ведущего контура</td><td>Выход ПИД</td></tr>
  </tbody>
</table>
<h3>Безударное переключение</h3>
<p>При переключении из РУЧНОГО в АВТО, состояние интегратора предварительно загружается текущим ручным выходом, чтобы выход регулятора не «скачкообразно» изменился. Виртуальный ПЛК реализует это путём инициализации интегрального члена значением <code>МВ_ручной − Кп×е</code> при смене режима.</p>`,
        pid: `
<h2>Теория ПИД-регулятора</h2>
<p><strong>ПИД-регулятор</strong> (пропорционально-интегрально-дифференцирующий) — наиболее распространённый в промышленности регулятор с обратной связью, охватывающий более 90 % промышленных контуров управления.</p>
<h3>Стандартная (ISA) форма</h3>
<pre>u(t) = Кп · [ е(t) + (1/Ти)·∫е dt + Тд·(де/dt) ]</pre>
<p>где:</p>
<ul>
  <li><strong>е(t) = ЗД − ПЗ</strong> — ошибка управления</li>
  <li><strong>Кп</strong> — пропорциональный коэффициент (безразмерный)</li>
  <li><strong>Ти = Кп/Ки</strong> — время интегрирования (с); устраняет статическую ошибку</li>
  <li><strong>Тд = Кд/Кп</strong> — время дифференцирования (с); демпфирует быстрые возмущения</li>
</ul>
<h3>Дискретная (приращений) форма</h3>
<p>Виртуальный ПЛК использует <em>позиционную</em> форму с интегратором Тустина (билинейным):</p>
<pre>
П = Кп · е[k]
И[k] = И[k-1] + Ки · Тс/2 · (е[k] + е[k-1])   (здесь применяется ограничение антинасыщения)
Д = Кд · (е[k] − е[k-1]) / Тс
u[k] = П + И[k] + Д
</pre>
<h3>Антинасыщение</h3>
<p>Когда выход регулятора насыщается (достигает границ 0–100 %), интегратор продолжает накапливать — это <em>насыщение интегратора</em>. Виртуальный ПЛК ограничивает интегратор: если ненасыщенный и насыщенный выходы равны — накопление разрешено; иначе интегратор удерживается на границе.</p>
<h3>Рекомендации по настройке</h3>
<table>
  <thead><tr><th>Параметр</th><th>Слишком мало</th><th>Слишком много</th></tr></thead>
  <tbody>
    <tr><td>Кп</td><td>Медленный отклик, большая ошибка</td><td>Колебания, неустойчивость</td></tr>
    <tr><td>Ки</td><td>Медленное устранение ошибки</td><td>Насыщение, перерегулирование</td></tr>
    <tr><td>Кд</td><td>Слабое подавление возмущений</td><td>Усиление шума, дребезг</td></tr>
  </tbody>
</table>
<p>Хорошей отправной точкой является <strong>метод Циглера–Никольса по переходной характеристике</strong>: определите коэффициент усиления процесса K, постоянную времени τ и запаздывание L из реакции на ступенчатое воздействие разомкнутой системы, затем установите Кп = 1,2τ/(K·L), Ти = 2L, Тд = L/2.</p>`,
        usage: `
<h2>Руководство по работе</h2>
<h3>Запуск и остановка</h3>
<p>Используйте кнопки <strong>Запуск</strong> / <strong>Стоп</strong> в заголовке. Значок состояния показывает текущее состояние ПЛК. Все контуры запускаются и останавливаются одновременно.</p>
<h3>Управление контурами</h3>
<ul>
  <li><strong>Выбор режима</strong> — переключение между АВТО / РУЧНОЙ / КАСКАД для каждого контура.</li>
  <li><strong>Задание</strong> — введите значение и нажмите <em>Уст.</em> (или Enter) для обновления задания контура.</li>
  <li><strong>Возмущение</strong> — инжектирует ступенчатое возмущение ±5 % на 30 секунд в симулятор процесса; удобно для проверки устойчивости регулятора.</li>
  <li><strong>Сброс</strong> — очищает состояние интегратора ПИД и возвращает контур в начальное состояние.</li>
  <li><strong>Коэфф. ПИД</strong> — разверните панель коэффициентов для просмотра Кп / Ки / Кд каждого контура (только чтение; редактируется в <code>configs/default.json</code>).</li>
</ul>
<h3>Графики</h3>
<ul>
  <li><strong>Масштаб</strong> — прокрутите колёсиком мыши по графику для увеличения/уменьшения масштаба по оси времени.</li>
  <li><strong>Сдвиг</strong> — удерживайте левую кнопку мыши и тяните влево/вправо для смещения вида.</li>
  <li><strong>Сброс вида</strong> — нажмите ↺ для возврата к масштабу по умолчанию с прокруткой к последним данным.</li>
  <li><strong>На весь экран</strong> — нажмите ⛶ для разворачивания графика на весь экран браузера. Нажмите Escape или кнопку снова для выхода.</li>
  <li><strong>Экспорт PNG</strong> — сохраняет график как растровое изображение.</li>
  <li><strong>Экспорт PDF</strong> — открывает диалог печати; выберите «Сохранить как PDF» в браузере.</li>
</ul>
<h3>Настройка макета</h3>
<p>Перетащите вертикальный разделитель между панелью графиков и журналом событий для изменения размеров обеих панелей. Разделитель подсвечивается при наведении.</p>
<h3>Журнал событий</h3>
<p>Журнал событий показывает все события опроса ПЛК в реальном времени. Нажмите <strong>Очистить</strong> для очистки списка. Журнал ограничен последними 200 записями.</p>`,
        api: `
<h2>Справочник API</h2>
<p>Встроенный HTTP-сервер предоставляет минимальный REST API на том же порту, что и панель управления (по умолчанию <code>:8090</code>).</p>
<h3>GET /api/stream</h3>
<p>Поток Server-Sent Events. Генерирует следующие типы событий:</p>
<table>
  <thead><tr><th>Событие</th><th>Данные</th><th>Описание</th></tr></thead>
  <tbody>
    <tr><td><code>snapshot</code></td><td>Полный объект состояния</td><td>Отправляется один раз при подключении; полное состояние ПЛК и контуров.</td></tr>
    <tr><td><code>plc_event</code></td><td>Одно событие опроса</td><td>Испускается с каждым интервалом обновления UI с актуальными ПЗ/ЗД/РВ для каждого контура.</td></tr>
    <tr><td><code>heartbeat</code></td><td><code>{"ts":"…"}</code></td><td>Поддержание соединения; испускается каждые 15 с.</td></tr>
  </tbody>
</table>
<h3>POST /api/commands/start</h3>
<p>Запускает все контуры ПЛК. Тело: <code>{}</code>. При успехе возвращает 200.</p>
<h3>POST /api/commands/stop</h3>
<p>Останавливает все контуры ПЛК. Тело: <code>{}</code>. При успехе возвращает 200.</p>
<h3>POST /api/commands/setpoint</h3>
<p>Обновляет задание контура. Тело: <code>{"loop":"имя_контура","value":42.0}</code>.</p>
<h3>POST /api/commands/mode</h3>
<p>Изменяет режим контура. Тело: <code>{"loop":"имя_контура","mode":"auto"|"manual"|"hold"}</code>.</p>
<h3>POST /api/commands/manual_output</h3>
<p>Устанавливает значение ручного выхода (только режим РУЧНОЙ). Тело: <code>{"loop":"имя_контура","value":50.0}</code>.</p>
<h3>POST /api/commands/inject_disturbance</h3>
<p>Инжектирует ступенчатое возмущение. Тело: <code>{"loop":"имя_контура","amplitude":5.0,"duration_seconds":30}</code>.</p>
<h3>POST /api/commands/reset_loop</h3>
<p>Сбрасывает интегратор ПИД. Тело: <code>{"loop":"имя_контура"}</code>.</p>`
    }
};

let _currentLang = localStorage.getItem('vplc_lang') || 'en';

function getCurrentLang() { return _currentLang; }

function t(key) {
    return (TRANSLATIONS[_currentLang] || TRANSLATIONS.en)[key] || key;
}

function setLanguage(lang) {
    _currentLang = lang;
    localStorage.setItem('vplc_lang', lang);
    document.querySelectorAll('[data-i18n]').forEach(el => {
        const key = el.getAttribute('data-i18n');
        if (el.tagName === 'INPUT' && el.placeholder !== undefined) {
            el.placeholder = t(key);
        } else {
            el.textContent = t(key);
        }
    });
    document.querySelectorAll('[data-i18n-title]').forEach(el => {
        el.title = t(el.getAttribute('data-i18n-title'));
    });
    document.dispatchEvent(new CustomEvent('langchange', { detail: { lang } }));
}

function getManualContent(tab) {
    return (MANUAL_CONTENT[_currentLang] || MANUAL_CONTENT.en)[tab] || '';
}
