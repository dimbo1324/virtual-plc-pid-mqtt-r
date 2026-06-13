# Полное техническое описание проекта `virtual-plc-pid-mqtt-r`

**Проект:** `virtual-plc-pid-mqtt-r`  
**Тип проекта:** виртуальный ПЛК + PID-регуляторы + MQTT + локальный web-интерфейс  
**Основной язык:** Go 1.26  
**Форма поставки:** один исполняемый файл + конфигурация + локальная база/журналы + опциональный Docker Compose для MQTT-брокера  
**Целевая платформа:** Windows 11 x64 как основная платформа демонстрации; Linux/macOS как желательные дополнительные платформы  
**Назначение документа:** дать разработчикам полное техническое задание и архитектурное описание, достаточное для реализации проекта без необходимости уточнять базовые решения.

---

## Содержание

1. [Назначение документа](#1-назначение-документа)
2. [Краткая идея проекта](#2-краткая-идея-проекта)
3. [Инженерный контекст](#3-инженерный-контекст)
4. [Цели проекта](#4-цели-проекта)
5. [Что проект не должен делать](#5-что-проект-не-должен-делать)
6. [Целевая аудитория](#6-целевая-аудитория)
7. [Ключевой демонстрационный сценарий](#7-ключевой-демонстрационный-сценарий)
8. [Общий состав системы](#8-общий-состав-системы)
9. [Рекомендуемый технологический стек](#9-рекомендуемый-технологический-стек)
10. [Принципиальная архитектура](#10-принципиальная-архитектура)
11. [Рекомендуемая структура репозитория](#11-рекомендуемая-структура-репозитория)
12. [Модель работы приложения](#12-модель-работы-приложения)
13. [Виртуальный ПЛК](#13-виртуальный-плк)
14. [PID-регуляторы](#14-pid-регуляторы)
15. [Синтетическая модель процесса](#15-синтетическая-модель-процесса)
16. [Генерация случайных входных данных](#16-генерация-случайных-входных-данных)
17. [MQTT-интерфейс](#17-mqtt-интерфейс)
18. [Локальный web-интерфейс](#18-локальный-web-интерфейс)
19. [HTTP API локального интерфейса](#19-http-api-локального-интерфейса)
20. [SSE-поток телеметрии](#20-sse-поток-телеметрии)
21. [Логирование и аудит](#21-логирование-и-аудит)
22. [Хранение данных](#22-хранение-данных)
23. [Конфигурация проекта](#23-конфигурация-проекта)
24. [Модели данных](#24-модели-данных)
25. [Алгоритмы](#25-алгоритмы)
26. [Пользовательские сценарии](#26-пользовательские-сценарии)
27. [Обработка ошибок](#27-обработка-ошибок)
28. [Тестирование](#28-тестирование)
29. [Производительность](#29-производительность)
30. [Безопасность](#30-безопасность)
31. [Сборка и запуск](#31-сборка-и-запуск)
32. [Docker Compose для MQTT](#32-docker-compose-для-mqtt)
33. [Стиль кода](#33-стиль-кода)
34. [План реализации по этапам](#34-план-реализации-по-этапам)
35. [Критерии готовности MVP](#35-критерии-готовности-mvp)
36. [Кейс для портфолио](#36-кейс-для-портфолио)
37. [Глоссарий](#37-глоссарий)
38. [Приложение A. Примеры конфигураций](#38-приложение-a-примеры-конфигураций)
39. [Приложение B. Примеры MQTT-сообщений](#39-приложение-b-примеры-mqtt-сообщений)
40. [Приложение C. Чеклист разработчика](#40-приложение-c-чеклист-разработчика)

---

## 1. Назначение документа

Этот документ описывает проект `virtual-plc-pid-mqtt-r` как самостоятельное инженерное приложение, которое должно быть простым, понятным, рабочим и пригодным для повторного использования в других проектах.

Документ предназначен для:

- backend/Go-разработчиков;
- инженеров АСУТП;
- разработчиков IIoT-решений;
- специалистов по цифровым двойникам;
- разработчиков демонстрационных стендов;
- технических ревьюеров;
- потенциальных работодателей, которые хотят понять инженерную идею проекта.

Документ должен отвечать на вопросы:

- что именно делает приложение;
- почему оно устроено именно так;
- какие модули нужно реализовать;
- какие данные циркулируют внутри системы;
- как работает виртуальный ПЛК;
- как работают PID-регуляторы;
- как генерируются синтетические входные данные;
- как публикуется MQTT-телеметрия;
- как устроен локальный web-интерфейс;
- как хранить события и телеметрию;
- как тестировать проект;
- как представить проект в портфолио.

Главный принцип документа: **никакой лишней сложности, но достаточно инженерной полноты, чтобы проект выглядел профессионально и был пригоден для реального переиспользования**.

---

## 2. Краткая идея проекта

`virtual-plc-pid-mqtt-r` — это лёгкий виртуальный ПЛК, написанный на Go, который имитирует базовую работу промышленного контроллера:

```text
входные сигналы / синтетические датчики
→ цикл сканирования ПЛК
→ PID-регуляторы
→ управляющие воздействия
→ MQTT-телеметрия
→ локальный web-dashboard
→ журналы и история
```

Приложение должно работать как единый локальный исполняемый файл. Оно не должно требовать сложной микросервисной инфраструктуры, отдельного backend, frontend-сборки, InfluxDB, Grafana или OPC UA. Для демонстрации MQTT достаточно локального Mosquitto broker через Docker Compose или уже установленного брокера.

Главный результат проекта — reusable-блок, который можно использовать как:

- учебный стенд по PID-регулированию;
- демонстратор виртуального ПЛК;
- источник MQTT-телеметрии для других IIoT-проектов;
- имитатор технологического объекта;
- портфолио-проект для вакансий в АСУТП, IIoT, simulation engineering и industrial software.

---

## 3. Инженерный контекст

В реальных промышленных системах ПЛК выполняет циклический алгоритм:

1. считывает входы;
2. обрабатывает сигналы;
3. выполняет управляющую логику;
4. рассчитывает выходы;
5. передаёт команды на исполнительные механизмы;
6. публикует состояние в SCADA, historian или IIoT-шлюз.

В данном проекте физическое оборудование заменяется виртуальными компонентами:

| Реальный объект | Виртуальный аналог в проекте |
|---|---|
| Датчик температуры | синтетический тег `temperature.pv` |
| Датчик давления | синтетический тег `pressure.pv` |
| Датчик уровня | синтетический тег `level.pv` |
| ПЛК | Go runtime с scan cycle |
| Функциональный блок PID | Go PID controller |
| Исполнительный механизм | виртуальный `mv` / output PID-регулятора |
| SCADA/HMI | локальный web-интерфейс |
| IIoT-шлюз | MQTT publisher/subscriber внутри приложения |
| Historian | SQLite/JSONL/локальные файлы истории |

Такой подход позволяет показать главное инженерное умение: **преобразование технологического процесса в программную модель и рабочий программный инструмент**.

---

## 4. Цели проекта

### 4.1. Основная цель

Создать простое и полностью рабочее приложение, которое демонстрирует виртуальный ПЛК с PID-регуляторами, MQTT-телеметрией, локальным web-интерфейсом, генерацией синтетических входных данных и записью событий.

### 4.2. Технические цели

Проект должен:

- быть написан преимущественно на Go 1.26;
- запускаться одной командой;
- иметь понятную внутреннюю архитектуру;
- иметь переиспользуемые пакеты `pid`, `plc`, `simulator`, `mqtt`;
- иметь локальный web-интерфейс без отдельной frontend-сборки;
- публиковать телеметрию в MQTT;
- принимать управляющие команды через MQTT;
- генерировать реалистичные синтетические значения датчиков;
- сохранять телеметрию и события локально;
- иметь автоматические тесты;
- иметь документацию и демонстрационный README.

### 4.3. Инженерные цели

Проект должен демонстрировать:

- понимание PLC scan cycle;
- понимание PID-регулирования;
- понимание технологических переменных `PV`, `SP`, `MV`;
- понимание MQTT как IIoT-протокола;
- умение строить простую модель процесса;
- умение проектировать локальный monitoring dashboard;
- умение логировать инженерные события;
- умение делать приложение не только учебным, но и пригодным для подключения к другим системам.

### 4.4. Карьерная цель

Проект должен быть пригоден для демонстрации на GitHub как портфолио-кейс для ролей:

- Industrial Software Engineer;
- IIoT Engineer;
- PLC/SCADA Software Engineer;
- Digital Twin Engineer;
- Simulation Engineer;
- Scientific/Engineering Software Developer;
- Process Modelling Engineer;
- Backend Engineer for industrial systems.

---

## 5. Что проект не должен делать

Чтобы приложение оставалось простым, оно **не должно** включать в MVP:

- полноценный OPC UA server/client;
- InfluxDB;
- Grafana;
- Kubernetes;
- сложную микросервисную архитектуру;
- отдельный frontend-проект на React/Vue/Angular;
- сложную физическую модель котла;
- CFD-моделирование;
- настоящую связь с реальным ПЛК;
- управление реальным оборудованием;
- механизмы промышленной безопасности для реального контура;
- авторизацию пользователей;
- многопользовательский режим;
- облачную синхронизацию;
- сложную систему плагинов.

Важное ограничение: приложение является **симулятором и демонстрационным цифровым двойником**, а не системой реального управления.

В README и интерфейсе должно быть указано:

```text
This software is a virtual PLC / PID / MQTT simulator.
It is not intended to control real industrial equipment.
```

---

## 6. Целевая аудитория

### 6.1. Разработчик проекта

Человек, который хочет иметь компактный reusable-блок для будущих IIoT и digital twin проектов.

Ему важно:

- быстро запустить приложение;
- понять архитектуру;
- переиспользовать PID, PLC runtime и MQTT-модули;
- добавить новые виртуальные контуры;
- подключить проект к другому dashboard или backend.

### 6.2. Инженер АСУТП

Ему важно:

- увидеть scan cycle;
- увидеть PID-регулирование;
- видеть `PV`, `SP`, `MV`;
- менять setpoint;
- смотреть реакцию процесса;
- анализировать тренды;
- видеть журнал событий.

### 6.3. Работодатель / технический интервьюер

Ему важно быстро понять:

- что кандидат понимает промышленную автоматизацию;
- что кандидат умеет писать код;
- что кандидат понимает MQTT и IIoT;
- что кандидат умеет делать простые цифровые двойники;
- что кандидат умеет строить демонстрационные инженерные инструменты.

### 6.4. Студент / начинающий инженер

Ему важно:

- увидеть PID в действии;
- понять, как виртуальный процесс реагирует на управляющее воздействие;
- менять коэффициенты PID и смотреть эффект;
- запускать проект локально без сложной инфраструктуры.

---

## 7. Ключевой демонстрационный сценарий

Базовый сценарий должен занимать 2–5 минут.

1. Пользователь запускает MQTT-брокер:

```bash
docker compose -f docker/docker-compose.yml up -d
```

2. Пользователь запускает приложение:

```bash
go run ./cmd/vplc --config configs/default.json
```

или после сборки:

```bash
./vplc --config configs/default.json
```

3. Приложение:

- загружает конфигурацию;
- запускает виртуальный PLC runtime;
- подключается к MQTT broker;
- запускает локальный HTTP server;
- открывает web-интерфейс в браузере;
- начинает генерировать телеметрию.

4. Пользователь видит:

- состояние приложения: Running;
- MQTT status: Connected;
- scan interval: например, 500 ms;
- список PID-контуров;
- графики `SP`, `PV`, `MV`;
- терминал событий;
- текущие значения датчиков.

5. Пользователь меняет setpoint давления.

6. Приложение:

- фиксирует команду в журнале;
- меняет setpoint PID-регулятора;
- рассчитывает новое управляющее воздействие;
- виртуальный процесс постепенно реагирует;
- график показывает переходный процесс;
- MQTT публикует обновлённую телеметрию.

7. Пользователь отправляет команду через MQTT:

```json
{
  "command": "set_setpoint",
  "loop": "pressure",
  "value": 7.5
}
```

8. В web-интерфейсе видно, что setpoint изменился.

9. Пользователь открывает файл логов и видит историю действий.

Этот сценарий должен быть основным для README, видео-демо и собеседования.

---

## 8. Общий состав системы

Система состоит из одного основного Go-приложения и одного внешнего MQTT-брокера.

```text
+-------------------------------------------------------------+
|                  virtual-plc-pid-mqtt-r                     |
|                                                             |
|  +----------------+      +-------------------------------+  |
|  | Synthetic      | ---> | Virtual PLC scan cycle          |  |
|  | process model  |      |                               |  |
|  +----------------+      | read inputs                   |  |
|                          | run PID                       |  |
|  +----------------+      | update outputs                |  |
|  | Random sensor  | ---> | publish telemetry             |  |
|  | generator      |      | write logs/history            |  |
|  +----------------+      +-------------------------------+  |
|                                      |                      |
|                                      v                      |
|                           +---------------------+           |
|                           | MQTT client         |           |
|                           +---------------------+           |
|                                      |                      |
|                                      v                      |
|                           +---------------------+           |
|                           | Local web dashboard |           |
|                           +---------------------+           |
|                                      |                      |
|                                      v                      |
|                           +---------------------+           |
|                           | Logs / SQLite       |           |
|                           +---------------------+           |
+-------------------------------------------------------------+
                         |
                         v
                 +----------------+
                 | MQTT broker    |
                 | Mosquitto      |
                 +----------------+
```

В MVP все внутренние компоненты работают в одном процессе. Это упрощает запуск и отладку.

---

## 9. Рекомендуемый технологический стек

### 9.1. Основной язык

- Go 1.26.

Причины выбора:

- быстрый запуск;
- простая сборка в один бинарный файл;
- хорошая поддержка сетевых сервисов;
- удобные goroutines/channels для scan cycle и телеметрии;
- строгая типизация;
- хорошая пригодность для промышленного backend/edge ПО.

### 9.2. HTTP server

Использовать стандартную библиотеку:

- `net/http`;
- `html/template` при необходимости;
- `embed` для встраивания статических файлов.

MVP не требует отдельного frontend toolchain.

### 9.3. Локальный web-интерфейс

- HTML;
- CSS;
- Vanilla JavaScript;
- Canvas/SVG charts или лёгкая vendored-библиотека для графиков.

Предпочтительно начать с собственного простого Canvas-графика, чтобы не добавлять npm.

### 9.4. MQTT

Рекомендуемый клиент:

```text
github.com/eclipse/paho.mqtt.golang
```

Причины:

- широко используется;
- понятный API;
- поддерживает publish/subscribe;
- подходит для локального Mosquitto.

### 9.5. Хранение данных

Рекомендуемый вариант MVP:

- `logs/app.log` — обычный текстовый лог;
- `logs/events.jsonl` — события в JSON Lines;
- `data/history.db` — SQLite для телеметрии и команд.

Для SQLite предпочтительно:

```text
modernc.org/sqlite
```

Причина: pure Go driver без CGO. Если размер зависимости станет проблемой, можно начать с JSONL/CSV и добавить SQLite во второй итерации.

### 9.6. Конфигурация

Для простоты использовать JSON:

```text
configs/default.json
```

Причины:

- поддерживается стандартной библиотекой;
- не требует YAML/TOML-зависимостей;
- легко валидируется;
- подходит для MVP.

### 9.7. Логирование

Использовать:

- `log/slog` из стандартной библиотеки;
- собственный writer для файла;
- структурированные поля.

### 9.8. Тестирование

- `testing` из стандартной библиотеки;
- table-driven tests;
- integration tests для MQTT при наличии брокера;
- race detector;
- `go test ./...`.

### 9.9. CLI

На первом этапе можно использовать стандартный `flag`.

Команды:

```bash
vplc --config configs/default.json
vplc --demo
vplc --version
vplc --validate-config configs/default.json
```

Если позже понадобится расширенный CLI, можно добавить Cobra, но в MVP это не нужно.

---

## 10. Принципиальная архитектура

Архитектура должна быть простой, но чистой.

Рекомендуемый подход:

```text
cmd/       — точки входа
internal/  — приложение, конфиг, storage, web UI
pkg/       — переиспользуемые компоненты: pid, plc, simulator, mqtt
web/       — статические файлы интерфейса
configs/   — конфигурации
docker/    — MQTT broker для локальной демонстрации
```

### 10.1. Разделение ответственности

| Модуль | Ответственность |
|---|---|
| `cmd/vplc` | чтение CLI-флагов, запуск приложения |
| `internal/app` | сборка зависимостей, lifecycle приложения |
| `internal/config` | загрузка и валидация конфигурации |
| `pkg/pid` | PID-регулятор как чистая математическая логика |
| `pkg/plc` | scan cycle, теги, контуры, runtime |
| `pkg/simulator` | синтетическая модель процесса и шумы |
| `pkg/mqttx` | MQTT publish/subscribe, команды |
| `internal/web` | локальный HTTP UI, REST, SSE |
| `internal/storage` | SQLite/JSONL/файлы истории |
| `internal/logging` | настройка slog и файловых журналов |

### 10.2. Зависимости между слоями

Правило:

```text
cmd → internal/app → pkg/* + internal/*
```

`pkg/pid` не должен знать о MQTT, HTTP, SQLite или UI.

`pkg/simulator` не должен знать о web-интерфейсе.

`pkg/plc` может использовать `pkg/pid` и `pkg/simulator`, но не должен напрямую зависеть от конкретной реализации storage.

`pkg/mqttx` не должен знать о деталях PID-математики.

`internal/web` читает состояние через application service или event bus, а не через глобальные переменные.

---

## 11. Рекомендуемая структура репозитория

```text
virtual-plc-pid-mqtt-r/
├─ cmd/
│  └─ vplc/
│     └─ main.go
│
├─ internal/
│  ├─ app/
│  │  ├─ app.go
│  │  ├─ lifecycle.go
│  │  └─ version.go
│  │
│  ├─ config/
│  │  ├─ config.go
│  │  ├─ defaults.go
│  │  └─ validation.go
│  │
│  ├─ logging/
│  │  ├─ logger.go
│  │  └─ rotation.go
│  │
│  ├─ storage/
│  │  ├─ storage.go
│  │  ├─ sqlite.go
│  │  ├─ jsonl.go
│  │  └─ migrations.go
│  │
│  └─ web/
│     ├─ server.go
│     ├─ handlers.go
│     ├─ sse.go
│     └─ embedded.go
│
├─ pkg/
│  ├─ pid/
│  │  ├─ controller.go
│  │  ├─ config.go
│  │  ├─ state.go
│  │  └─ controller_test.go
│  │
│  ├─ plc/
│  │  ├─ runtime.go
│  │  ├─ scan.go
│  │  ├─ loop.go
│  │  ├─ tags.go
│  │  ├─ snapshot.go
│  │  └─ runtime_test.go
│  │
│  ├─ simulator/
│  │  ├─ process.go
│  │  ├─ first_order.go
│  │  ├─ random.go
│  │  ├─ disturbance.go
│  │  └─ simulator_test.go
│  │
│  └─ mqttx/
│     ├─ client.go
│     ├─ publisher.go
│     ├─ subscriber.go
│     ├─ payloads.go
│     └─ commands.go
│
├─ web/
│  ├─ index.html
│  ├─ styles.css
│  ├─ app.js
│  └─ chart.js
│
├─ configs/
│  ├─ default.json
│  ├─ fast-demo.json
│  └─ noisy-process.json
│
├─ docker/
│  ├─ docker-compose.yml
│  └─ mosquitto/
│     └─ mosquitto.conf
│
├─ data/
│  └─ .gitkeep
│
├─ logs/
│  └─ .gitkeep
│
├─ docs/
│  ├─ screenshots/
│  └─ demo.md
│
├─ scripts/
│  ├─ run_demo.ps1
│  ├─ run_demo.sh
│  ├─ build.ps1
│  └─ build.sh
│
├─ tests/
│  └─ integration/
│     └─ mqtt_integration_test.go
│
├─ .gitignore
├─ go.mod
├─ go.sum
├─ README.md
├─ CHANGELOG.md
└─ LICENSE
```

### 11.1. Почему `pkg/` допустим

Пакеты `pid`, `plc`, `simulator`, `mqttx` должны быть потенциально переиспользуемыми в других проектах. Поэтому их можно разместить в `pkg/`.

Но не нужно превращать всё в публичную библиотеку. Внешний API этих пакетов должен быть минимальным.

### 11.2. Почему `internal/` нужен

`internal/` содержит glue-code конкретного приложения:

- сборку зависимостей;
- web server;
- storage;
- logging;
- конфигурацию;
- lifecycle.

Это не должно импортироваться другими проектами напрямую.

---

## 12. Модель работы приложения

### 12.1. Startup sequence

При запуске приложение выполняет следующие шаги:

1. читает CLI-флаги;
2. загружает конфигурацию;
3. валидирует конфигурацию;
4. создаёт logger;
5. создаёт storage;
6. создаёт process simulator;
7. создаёт PID-регуляторы;
8. создаёт PLC runtime;
9. создаёт MQTT client;
10. создаёт локальный web server;
11. подписывается на MQTT commands topic;
12. запускает PLC scan loop;
13. запускает публикацию telemetry;
14. запускает SSE stream для UI;
15. открывает web-интерфейс в браузере, если включено в конфиге.

### 12.2. Runtime loop

Каждый scan interval:

```text
1. получить текущее время
2. прочитать текущие process values из simulator
3. для каждого PID-контура:
   3.1 прочитать SP
   3.2 прочитать PV
   3.3 рассчитать error = SP - PV
   3.4 рассчитать PID output
   3.5 применить ограничения output_min/output_max
4. передать outputs в simulator
5. обновить виртуальный процесс
6. сформировать snapshot
7. отправить snapshot в event bus
8. записать snapshot в storage
9. опубликовать telemetry в MQTT, если наступил publish interval
10. отправить snapshot в SSE clients
11. записать важные события в log
```

### 12.3. Shutdown sequence

При завершении:

1. остановить PLC runtime через context cancellation;
2. дождаться завершения goroutines;
3. отправить MQTT offline status / last will если применимо;
4. закрыть MQTT client;
5. закрыть storage;
6. закрыть log files;
7. записать событие `application_stopped`.

### 12.4. Управление временем

Для PID и simulator нужно использовать реальное `dt`, а не предполагать идеальный scan interval.

Правильно:

```go
now := clock.Now()
dt := now.Sub(previousTime).Seconds()
```

Неправильно:

```go
dt := config.ScanInterval.Seconds()
```

Причина: реальный цикл может иметь jitter, особенно на обычной ОС.

---

## 13. Виртуальный ПЛК

### 13.1. Назначение PLC runtime

PLC runtime — центральный модуль приложения. Он имитирует циклическую работу промышленного контроллера.

Он должен:

- иметь состояние `Stopped`, `Starting`, `Running`, `Stopping`, `Faulted`;
- поддерживать scan interval;
- запускаться и останавливаться через context;
- хранить список PID-контуров;
- обрабатывать команды;
- публиковать snapshots;
- фиксировать события.

### 13.2. PLC states

```text
Stopped  — приложение загружено, но scan cycle не выполняется
Starting — выполняется запуск runtime
Running  — scan cycle выполняется
Stopping — выполняется остановка
Faulted  — runtime остановлен из-за критической ошибки
```

### 13.3. PLC scan interval

Scan interval задаётся в конфигурации:

```json
{
  "plc": {
    "scan_interval_ms": 500
  }
}
```

Рекомендуемые значения:

| Режим | Scan interval |
|---|---:|
| медленный demo | 1000 ms |
| стандартный demo | 500 ms |
| быстрый demo | 100 ms |
| тесты | 10–50 ms |

Для MVP не требуется hard real-time. Но приложение должно логировать scan overrun.

### 13.4. Scan overrun

Scan overrun возникает, если выполнение цикла заняло больше времени, чем заданный scan interval.

Пример события:

```json
{
  "timestamp": "2026-06-13T14:10:12.100Z",
  "level": "warning",
  "event_type": "plc_scan_overrun",
  "scan_duration_ms": 612,
  "scan_interval_ms": 500
}
```

### 13.5. Tags

Виртуальный ПЛК должен работать с тегами.

Минимальные теги:

```text
pressure.sp
pressure.pv
pressure.mv
pressure.mode
pressure.enabled

temperature.sp
temperature.pv
temperature.mv
temperature.mode
temperature.enabled

level.sp
level.pv
level.mv
level.mode
level.enabled

plc.state
plc.scan_interval_ms
plc.scan_duration_ms
mqtt.connected
```

### 13.6. Loop modes

Каждый PID-контур должен поддерживать режимы:

```text
auto    — output рассчитывается PID-регулятором
manual  — output задаётся вручную
hold    — output удерживается на последнем значении
disabled — контур отключён
```

### 13.7. Commands

PLC runtime должен принимать команды из двух источников:

- локальный web UI через HTTP API;
- MQTT commands topic.

Внутри приложения оба источника должны приводиться к единой модели `Command`.

---

## 14. PID-регуляторы

### 14.1. Назначение PID

PID-регулятор управляет виртуальным исполнительным механизмом так, чтобы process value приближался к setpoint.

Термины:

| Термин | Значение |
|---|---|
| `SP` | setpoint, заданное значение |
| `PV` | process value, измеренное значение |
| `MV` | manipulated value, управляющее воздействие |
| `Kp` | пропорциональный коэффициент |
| `Ki` | интегральный коэффициент |
| `Kd` | дифференциальный коэффициент |
| `dt` | шаг времени |

### 14.2. Базовая формула

```text
error = SP - PV
P = Kp * error
I = I + Ki * error * dt
D = Kd * (error - previous_error) / dt
MV = bias + P + I + D
```

После расчёта `MV` ограничивается:

```text
MV = clamp(MV, output_min, output_max)
```

### 14.3. Anti-windup

PID должен иметь защиту от интегрального насыщения.

Минимальный вариант:

- если output достиг ограничения;
- и ошибка продолжает толкать output дальше в насыщение;
- интегратор не увеличивается.

Пример логики:

```text
candidate_I = I + Ki * error * dt
candidate_output = P + candidate_I + D

if candidate_output inside limits:
    I = candidate_I
else if output saturated high and error < 0:
    I = candidate_I
else if output saturated low and error > 0:
    I = candidate_I
else:
    I unchanged
```

### 14.4. Derivative on measurement

Для уменьшения derivative kick желательно считать D по изменению PV, а не error:

```text
D = -Kd * (PV - previous_PV) / dt
```

В MVP можно реализовать классический вариант по error, но в документации к коду следует указать выбранный подход.

Рекомендуемый вариант для проекта:

```text
D on measurement
```

### 14.5. Output limits

Каждый PID должен иметь ограничения:

```json
{
  "output_min": 0.0,
  "output_max": 100.0
}
```

Единица измерения output — проценты открытия виртуального исполнительного механизма.

### 14.6. Manual mode

В manual mode PID не пересчитывает output. Пользователь задаёт `manual_output`.

При переходе из manual в auto нужно обеспечить bumpless transfer.

Минимальный вариант:

- при входе в auto установить интегратор так, чтобы текущий output не прыгнул резко;
- или просто принять текущий manual output как начальный output PID.

### 14.7. PID state

PID должен хранить:

```go
type State struct {
    Setpoint      float64
    ProcessValue  float64
    Output        float64
    Error         float64
    Integral      float64
    Derivative    float64
    LastError     float64
    LastPV        float64
    LastUpdate    time.Time
    Mode          Mode
    Enabled       bool
}
```

### 14.8. PID config

```go
type Config struct {
    Name      string
    Kp        float64
    Ki        float64
    Kd        float64
    Bias      float64
    OutputMin float64
    OutputMax float64
    Setpoint  float64
    Mode      Mode
    Enabled   bool
}
```

### 14.9. Обязательные тесты PID

Нужны тесты:

- output растёт при положительной ошибке;
- output падает при отрицательной ошибке;
- output не выходит за limits;
- integrator не уходит бесконечно в saturation;
- manual mode не пересчитывает output;
- disabled loop выдаёт безопасное значение;
- при `dt <= 0` возвращается ошибка или безопасное поведение;
- при NaN входе команда отклоняется.

---

## 15. Синтетическая модель процесса

### 15.1. Назначение

Модель процесса нужна, чтобы PID-регуляторы не работали «в пустоту». Если PID меняет output, process value должен постепенно реагировать.

Проект не должен моделировать реальный котёл или реальную установку. Нужна простая, но логичная динамика.

### 15.2. Рекомендуемая модель первого порядка

Для каждого контура можно использовать first-order lag:

```text
PV(t + dt) = PV(t) + dt/tau * (target - PV(t)) + noise + disturbance
```

Где:

```text
target = base + gain * MV
```

Параметры:

| Параметр | Значение |
|---|---|
| `base` | базовое значение процесса |
| `gain` | влияние output на процесс |
| `tau` | постоянная времени |
| `noise_stddev` | стандартное отклонение шума |
| `disturbance` | внешнее возмущение |

### 15.3. Пример: pressure loop

```text
pressure.pv = pressure.pv + dt/tau * ((base + gain * pressure.mv) - pressure.pv) + noise
```

Пример параметров:

```json
{
  "name": "pressure",
  "unit": "bar",
  "initial_pv": 4.0,
  "base": 0.0,
  "gain": 0.10,
  "tau_seconds": 15.0,
  "noise_stddev": 0.03,
  "min": 0.0,
  "max": 12.0
}
```

Если output = 60%, target будет примерно 6 bar.

### 15.4. Пример: temperature loop

```text
temperature.pv = temperature.pv + dt/tau * ((ambient + gain * heater_output) - temperature.pv) + noise
```

Пример:

```json
{
  "name": "temperature",
  "unit": "C",
  "initial_pv": 80.0,
  "base": 20.0,
  "gain": 2.0,
  "tau_seconds": 60.0,
  "noise_stddev": 0.2,
  "min": 0.0,
  "max": 250.0
}
```

Если output = 80%, target будет 180 °C.

### 15.5. Пример: level loop

Для уровня можно использовать похожую модель, но логичнее сделать баланс притока/оттока:

```text
level(t + dt) = level(t) + dt * (inflow - outflow) + noise
```

Где:

```text
inflow = feedwater_mv * inflow_gain
outflow = base_outflow + demand_disturbance
```

Но для MVP допустима first-order model, чтобы не усложнять.

### 15.6. Ограничения process values

Все PV должны ограничиваться физически допустимым диапазоном:

```text
PV = clamp(PV, min, max)
```

Если значение дошло до границы, нужно записать warning event.

### 15.7. Связь между контурами

В MVP контуры могут быть независимыми. Это делает проект проще.

Позже можно добавить слабые связи:

- давление влияет на расход;
- температура влияет на давление;
- уровень влияет на безопасность процесса.

Но в первой версии это не обязательно.

---

## 16. Генерация случайных входных данных

### 16.1. Назначение random mode

Random mode нужен, чтобы приложение могло имитировать входные сигналы даже без физической модели.

Это полезно для:

- проверки MQTT;
- проверки dashboard;
- генерации тестовых данных;
- демонстрации шума датчиков;
- разработки UI.

### 16.2. Типы случайности

Рекомендуемые типы:

1. Gaussian noise;
2. slow drift;
3. step disturbance;
4. spike/outlier;
5. sensor dropout.

### 16.3. Gaussian noise

```text
value = value + random_normal(0, stddev)
```

### 16.4. Slow drift

```text
value = value + drift_rate * dt
```

Дрифт имитирует постепенное смещение датчика.

### 16.5. Step disturbance

В случайный момент времени процесс получает скачок:

```text
disturbance = +1.5 bar for 30 seconds
```

### 16.6. Spike/outlier

Единичный выброс:

```text
pressure.pv = pressure.pv + random spike
```

Это должно фиксироваться в событиях.

### 16.7. Sensor dropout

Имитирует пропажу датчика.

В MVP не нужно возвращать `null` в основную PID-логику. Лучше использовать событие и удерживать последнее значение.

```text
if sensor dropout:
    PV = last_valid_pv
    quality = bad
```

### 16.8. Quality flags

Каждый тег может иметь quality:

```text
good
uncertain
bad
```

В MVP это можно добавить в telemetry payload, но не обязательно выводить везде в UI.

---

## 17. MQTT-интерфейс

### 17.1. Назначение MQTT

MQTT является главным внешним интерфейсом проекта. Через него приложение:

- публикует телеметрию;
- публикует события;
- публикует online/offline status;
- принимает команды;
- может принимать изменение конфигурации.

### 17.2. Broker

По умолчанию:

```text
tcp://localhost:1883
```

Для демо используется Mosquitto.

### 17.3. Device ID

Device ID задаётся в конфигурации:

```json
{
  "device_id": "vplc_001"
}
```

Все топики должны включать device ID.

### 17.4. Топики

Рекомендуемая структура:

```text
vplc/{device_id}/status
vplc/{device_id}/telemetry
vplc/{device_id}/events
vplc/{device_id}/commands
vplc/{device_id}/config
```

Пример:

```text
vplc/vplc_001/telemetry
```

### 17.5. QoS

Рекомендации:

| Тип сообщения | QoS |
|---|---:|
| telemetry | 0 |
| status | 1 |
| events | 1 |
| commands | 1 |
| config | 1 |

Для MVP telemetry QoS 0 достаточно.

### 17.6. Retained messages

Для status можно использовать retained:

```text
online/offline retained = true
```

Для telemetry retained лучше не использовать.

### 17.7. Last Will and Testament

MQTT client должен установить LWT:

Topic:

```text
vplc/{device_id}/status
```

Payload:

```json
{
  "device_id": "vplc_001",
  "status": "offline",
  "reason": "unexpected_disconnect"
}
```

### 17.8. Telemetry payload

```json
{
  "timestamp": "2026-06-13T14:10:00.000Z",
  "device_id": "vplc_001",
  "plc": {
    "state": "running",
    "scan_interval_ms": 500,
    "last_scan_duration_ms": 4.8,
    "scan_counter": 120
  },
  "loops": {
    "pressure": {
      "sp": 6.0,
      "pv": 5.82,
      "mv": 61.4,
      "mode": "auto",
      "quality": "good",
      "unit": "bar"
    },
    "temperature": {
      "sp": 180.0,
      "pv": 176.3,
      "mv": 72.1,
      "mode": "auto",
      "quality": "good",
      "unit": "C"
    },
    "level": {
      "sp": 50.0,
      "pv": 49.2,
      "mv": 48.5,
      "mode": "auto",
      "quality": "good",
      "unit": "%"
    }
  }
}
```

### 17.9. Command payload: set setpoint

```json
{
  "command_id": "cmd-001",
  "command": "set_setpoint",
  "loop": "pressure",
  "value": 7.5
}
```

### 17.10. Command payload: set PID gains

```json
{
  "command_id": "cmd-002",
  "command": "set_pid_gains",
  "loop": "pressure",
  "kp": 2.0,
  "ki": 0.4,
  "kd": 0.05
}
```

### 17.11. Command payload: set mode

```json
{
  "command_id": "cmd-003",
  "command": "set_mode",
  "loop": "temperature",
  "mode": "manual",
  "manual_output": 40.0
}
```

### 17.12. Command payload: start/stop PLC

```json
{
  "command_id": "cmd-004",
  "command": "stop_plc"
}
```

```json
{
  "command_id": "cmd-005",
  "command": "start_plc"
}
```

### 17.13. Command response event

После обработки команды приложение публикует event:

```json
{
  "timestamp": "2026-06-13T14:10:02.000Z",
  "device_id": "vplc_001",
  "event_type": "command_applied",
  "command_id": "cmd-001",
  "message": "Setpoint changed",
  "details": {
    "loop": "pressure",
    "old_value": 6.0,
    "new_value": 7.5
  }
}
```

Если команда отклонена:

```json
{
  "timestamp": "2026-06-13T14:10:02.000Z",
  "device_id": "vplc_001",
  "event_type": "command_rejected",
  "command_id": "cmd-010",
  "message": "Unknown loop: flow",
  "severity": "warning"
}
```

### 17.14. Валидация команд

Команда отклоняется, если:

- JSON невалиден;
- отсутствует `command`;
- неизвестный loop;
- значение setpoint вне диапазона;
- коэффициенты PID отрицательные там, где это запрещено;
- mode неизвестен;
- manual output вне диапазона;
- команда не поддерживается.

---

## 18. Локальный web-интерфейс

### 18.1. Выбранный вариант визуализации

Для проекта выбран вариант:

```text
Go + встроенный локальный web-интерфейс
```

Это значит:

- Go-приложение запускает локальный HTTP server;
- статические файлы UI встраиваются в бинарник через `go:embed`;
- пользователь открывает `http://localhost:8080`;
- интерфейс получает live-данные через SSE;
- команды отправляются через HTTP API;
- отдельный frontend build не нужен.

### 18.2. Главная цель интерфейса

Интерфейс должен быть простым операторским dashboard, а не сложной SCADA.

Он должен показывать:

- текущее состояние приложения;
- MQTT connection status;
- текущие значения PID-контуров;
- графики SP/PV/MV;
- live terminal событий;
- кнопки start/stop;
- поля изменения setpoint и PID gains.

### 18.3. Структура экрана

Рекомендуемая компоновка:

```text
+--------------------------------------------------------------+
| Virtual PLC PID MQTT       Running | MQTT Connected | 500 ms |
+----------------------+---------------------------------------+
| Loops                | Charts                                |
|                      |                                       |
| Pressure             | [Pressure SP/PV/MV chart]             |
| Temperature          | [Temperature SP/PV/MV chart]          |
| Level                | [Level SP/PV/MV chart]                |
|                      |                                       |
+----------------------+---------------------------------------+
| Controls                                                     |
| SP, Kp, Ki, Kd, Mode, Manual Output                           |
+--------------------------------------------------------------+
| Live terminal / events                                        |
| 14:10:00 app started                                          |
| 14:10:01 mqtt connected                                       |
| 14:10:03 pressure setpoint changed 6.0 -> 7.5                 |
+--------------------------------------------------------------+
```

### 18.4. Основные элементы UI

#### Header

Показывает:

- название проекта;
- runtime state;
- MQTT status;
- scan interval;
- uptime;
- кнопки Start/Stop.

#### Loop cards

Для каждого PID-контура показывать:

- name;
- PV;
- SP;
- MV;
- mode;
- error;
- quality;
- unit.

#### Charts

Для каждого контура:

- линия PV;
- линия SP;
- линия MV;
- окно последних N секунд;
- автообновление.

MVP: хранить последние 300 точек в браузере.

#### Controls

Для выбранного контура:

- setpoint input;
- Kp input;
- Ki input;
- Kd input;
- mode select;
- manual output input;
- Apply button.

#### Terminal

Live terminal показывает события:

- application started;
- MQTT connected/disconnected;
- PLC started/stopped;
- command applied/rejected;
- setpoint changed;
- PID gains changed;
- scan overrun;
- sensor disturbance;
- storage error.

### 18.5. Цветовая схема

Рекомендуемый стиль:

- тёмный инженерный интерфейс;
- спокойный фон;
- неяркие акценты;
- высокая читаемость;
- без декоративной перегрузки.

Пример:

```text
background: #101418
panel:      #171d23
border:     #2a343d
text:       #e6edf3
muted:      #8b949e
accent:     #58a6ff
safe:       #3fb950
warning:    #d29922
critical:   #f85149
```

### 18.6. UX-принципы

Интерфейс должен:

- показывать состояние системы без лишних кликов;
- не перегружать пользователя;
- явно показывать, что это симулятор;
- подтверждать каждую команду событием;
- не позволять вводить заведомо некорректные значения;
- показывать ошибку рядом с полем;
- не блокировать графики при ошибке MQTT;
- работать локально без интернета.

### 18.7. Доступность

Минимальные требования:

- нормальный контраст;
- читаемые шрифты;
- кнопки имеют text labels;
- input fields имеют labels;
- ошибки не передаются только цветом;
- terminal можно копировать.

---

## 19. HTTP API локального интерфейса

### 19.1. Назначение

HTTP API используется только локальным web-интерфейсом и локальной отладкой.

По умолчанию server слушает:

```text
127.0.0.1:8080
```

### 19.2. Endpoints

```text
GET  /
GET  /static/*
GET  /api/status
GET  /api/snapshot
GET  /api/config
POST /api/plc/start
POST /api/plc/stop
POST /api/loops/{name}/setpoint
POST /api/loops/{name}/gains
POST /api/loops/{name}/mode
GET  /api/events/recent
GET  /api/telemetry/recent
GET  /api/stream
```

### 19.3. GET /api/status

Response:

```json
{
  "app": {
    "name": "virtual-plc-pid-mqtt-r",
    "version": "0.1.0",
    "uptime_seconds": 123
  },
  "plc": {
    "state": "running",
    "scan_interval_ms": 500,
    "last_scan_duration_ms": 4.8,
    "scan_counter": 120
  },
  "mqtt": {
    "enabled": true,
    "connected": true,
    "broker_url": "tcp://localhost:1883"
  }
}
```

### 19.4. GET /api/snapshot

Возвращает последний snapshot.

### 19.5. POST /api/loops/{name}/setpoint

Request:

```json
{
  "value": 7.5
}
```

Response:

```json
{
  "ok": true,
  "message": "Setpoint updated",
  "loop": "pressure",
  "old_value": 6.0,
  "new_value": 7.5
}
```

### 19.6. Ошибка HTTP API

```json
{
  "ok": false,
  "error": {
    "code": "invalid_setpoint",
    "message": "Setpoint is outside allowed range",
    "details": {
      "min": 0.0,
      "max": 12.0,
      "value": 30.0
    }
  }
}
```

### 19.7. GET /api/stream

SSE endpoint для live telemetry.

---

## 20. SSE-поток телеметрии

### 20.1. Почему SSE, а не WebSocket

Для MVP достаточно Server-Sent Events, потому что UI в основном получает поток данных от backend.

Преимущества SSE:

- проще WebSocket;
- поддерживается браузером;
- хорошо подходит для телеметрии;
- реализуется через `net/http`;
- автоматически переподключается.

### 20.2. Формат события snapshot

```text
event: snapshot
data: { ...json... }
```

### 20.3. Формат события log

```text
event: log
data: {"timestamp":"...","level":"info","message":"mqtt connected"}
```

### 20.4. Частота отправки

SSE может отправлять каждый PLC snapshot, если scan interval >= 250 ms.

Если scan interval меньше, UI должен получать decimated stream:

```text
max_ui_update_rate = 4–10 Hz
```

Это нужно, чтобы не перегружать браузер.

---

## 21. Логирование и аудит

### 21.1. Цели логирования

Логирование нужно для:

- демонстрации работы приложения;
- отладки;
- фиксации команд;
- анализа поведения PID;
- воспроизводимости демо;
- подтверждения инженерной дисциплины проекта.

### 21.2. Уровни логов

```text
DEBUG   — подробные технические данные
INFO    — нормальные события
WARNING — подозрительные, но не критичные события
ERROR   — ошибки, требующие внимания
```

### 21.3. Что логировать

Обязательно логировать:

- запуск приложения;
- загрузку конфигурации;
- ошибку конфигурации;
- запуск/остановку PLC;
- MQTT connect/disconnect;
- publish error;
- received command;
- command applied;
- command rejected;
- setpoint change;
- PID gains change;
- mode change;
- scan overrun;
- storage error;
- graceful shutdown.

### 21.4. Что не логировать

Не нужно логировать каждую точку telemetry в `app.log`. Это быстро засорит журнал.

Телеметрия должна храниться отдельно:

- SQLite;
- CSV/JSONL;
- in-memory ring buffer для UI.

### 21.5. Файлы логов

```text
logs/app.log
logs/events.jsonl
```

### 21.6. Event JSONL

Каждая строка — отдельный JSON.

Пример:

```json
{"timestamp":"2026-06-13T14:10:00Z","level":"info","event_type":"app_started","message":"application started"}
```

### 21.7. Ротация логов

MVP:

- если `app.log` больше 10 MB, переименовать в `app.log.1`;
- хранить до 5 файлов.

Если не успевает в MVP, допустимо добавить TODO, но не оставлять бесконечный рост файла в production-like версии.

---

## 22. Хранение данных

### 22.1. Цель хранения

Хранение нужно для:

- просмотра истории;
- экспорта данных;
- воспроизведения демонстрации;
- анализа PID-переходных процессов;
- доказательства, что приложение не просто рисует графики, а ведёт инженерный журнал.

### 22.2. Минимальные данные

Хранить:

- telemetry snapshots;
- commands;
- events;
- PID changes.

### 22.3. SQLite schema: telemetry_samples

```sql
CREATE TABLE IF NOT EXISTS telemetry_samples (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    timestamp TEXT NOT NULL,
    device_id TEXT NOT NULL,
    scan_counter INTEGER NOT NULL,
    loop_name TEXT NOT NULL,
    sp REAL NOT NULL,
    pv REAL NOT NULL,
    mv REAL NOT NULL,
    error REAL NOT NULL,
    mode TEXT NOT NULL,
    quality TEXT NOT NULL,
    unit TEXT NOT NULL
);
```

### 22.4. SQLite schema: events

```sql
CREATE TABLE IF NOT EXISTS events (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    timestamp TEXT NOT NULL,
    level TEXT NOT NULL,
    event_type TEXT NOT NULL,
    message TEXT NOT NULL,
    details_json TEXT
);
```

### 22.5. SQLite schema: commands

```sql
CREATE TABLE IF NOT EXISTS commands (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    timestamp TEXT NOT NULL,
    command_id TEXT,
    source TEXT NOT NULL,
    command_type TEXT NOT NULL,
    payload_json TEXT NOT NULL,
    status TEXT NOT NULL,
    error_message TEXT
);
```

### 22.6. Retention policy

MVP:

- хранить последние N часов или N записей;
- параметр в конфигурации.

Пример:

```json
{
  "storage": {
    "retention_max_samples": 100000
  }
}
```

### 22.7. Export

Желательная функция:

```bash
vplc --export data/history.db --format csv --out export/
```

Можно добавить после MVP.

---

## 23. Конфигурация проекта

### 23.1. Формат

MVP использует JSON.

Файл:

```text
configs/default.json
```

### 23.2. Основные секции

```json
{
  "app": {},
  "plc": {},
  "mqtt": {},
  "web": {},
  "storage": {},
  "loops": []
}
```

### 23.3. App config

```json
{
  "app": {
    "name": "virtual-plc-pid-mqtt-r",
    "device_id": "vplc_001",
    "auto_start": true,
    "open_browser": true
  }
}
```

### 23.4. PLC config

```json
{
  "plc": {
    "scan_interval_ms": 500,
    "publish_interval_ms": 1000,
    "ui_update_interval_ms": 250,
    "scan_overrun_warning_ms": 500
  }
}
```

### 23.5. MQTT config

```json
{
  "mqtt": {
    "enabled": true,
    "broker_url": "tcp://localhost:1883",
    "client_id": "virtual-plc-pid-mqtt-r-vplc-001",
    "username": "",
    "password": "",
    "base_topic": "vplc/vplc_001",
    "qos": 0,
    "connect_timeout_seconds": 5,
    "reconnect_interval_seconds": 3
  }
}
```

### 23.6. Web config

```json
{
  "web": {
    "enabled": true,
    "host": "127.0.0.1",
    "port": 8080
  }
}
```

### 23.7. Storage config

```json
{
  "storage": {
    "enabled": true,
    "type": "sqlite",
    "sqlite_path": "data/history.db",
    "events_jsonl_path": "logs/events.jsonl",
    "app_log_path": "logs/app.log",
    "retention_max_samples": 100000
  }
}
```

### 23.8. Loop config

```json
{
  "name": "pressure",
  "display_name": "Pressure",
  "unit": "bar",
  "enabled": true,
  "mode": "auto",
  "setpoint": 6.0,
  "setpoint_min": 0.0,
  "setpoint_max": 12.0,
  "pid": {
    "kp": 3.0,
    "ki": 0.25,
    "kd": 0.05,
    "bias": 0.0,
    "output_min": 0.0,
    "output_max": 100.0
  },
  "process": {
    "initial_pv": 4.0,
    "min": 0.0,
    "max": 12.0,
    "base": 0.0,
    "gain": 0.10,
    "tau_seconds": 15.0,
    "noise_stddev": 0.03,
    "random_disturbances": true
  }
}
```

### 23.9. Валидация конфига

Приложение не должно запускаться с некорректной конфигурацией.

Проверять:

- `device_id` не пустой;
- scan interval > 0;
- publish interval > 0;
- port в диапазоне 1–65535;
- loop names уникальны;
- output_min < output_max;
- setpoint в диапазоне;
- process min < max;
- tau_seconds > 0;
- mqtt broker URL не пустой, если MQTT включён.

---

## 24. Модели данных

### 24.1. AppStatus

```go
type AppStatus struct {
    Name          string        `json:"name"`
    Version       string        `json:"version"`
    DeviceID      string        `json:"device_id"`
    StartedAt     time.Time     `json:"started_at"`
    UptimeSeconds int64         `json:"uptime_seconds"`
}
```

### 24.2. PLCStatus

```go
type PLCStatus struct {
    State              string  `json:"state"`
    ScanIntervalMS     int     `json:"scan_interval_ms"`
    LastScanDurationMS float64 `json:"last_scan_duration_ms"`
    ScanCounter        uint64  `json:"scan_counter"`
}
```

### 24.3. LoopSnapshot

```go
type LoopSnapshot struct {
    Name         string  `json:"name"`
    DisplayName  string  `json:"display_name"`
    Unit         string  `json:"unit"`
    Setpoint     float64 `json:"sp"`
    ProcessValue float64 `json:"pv"`
    Output       float64 `json:"mv"`
    Error        float64 `json:"error"`
    Mode         string  `json:"mode"`
    Quality      string  `json:"quality"`
    Enabled      bool    `json:"enabled"`
    Kp           float64 `json:"kp"`
    Ki           float64 `json:"ki"`
    Kd           float64 `json:"kd"`
}
```

### 24.4. Snapshot

```go
type Snapshot struct {
    Timestamp time.Time               `json:"timestamp"`
    DeviceID  string                  `json:"device_id"`
    PLC       PLCStatus               `json:"plc"`
    Loops     map[string]LoopSnapshot `json:"loops"`
}
```

### 24.5. Command

```go
type Command struct {
    CommandID    string          `json:"command_id"`
    Command      string          `json:"command"`
    Loop         string          `json:"loop,omitempty"`
    Value        *float64        `json:"value,omitempty"`
    Kp           *float64        `json:"kp,omitempty"`
    Ki           *float64        `json:"ki,omitempty"`
    Kd           *float64        `json:"kd,omitempty"`
    Mode         string          `json:"mode,omitempty"`
    ManualOutput *float64        `json:"manual_output,omitempty"`
    Raw          json.RawMessage `json:"-"`
    Source       string          `json:"-"`
    ReceivedAt   time.Time       `json:"-"`
}
```

### 24.6. Event

```go
type Event struct {
    Timestamp time.Time      `json:"timestamp"`
    Level     string         `json:"level"`
    Type      string         `json:"event_type"`
    Message   string         `json:"message"`
    Details   map[string]any `json:"details,omitempty"`
}
```

---

## 25. Алгоритмы

### 25.1. Алгоритм запуска приложения

```text
load flags
load config
validate config
init logger
init storage
init event bus
init simulator
init pid controllers
init plc runtime
init mqtt client
init web server
start services
wait for shutdown signal
stop services gracefully
```

### 25.2. Алгоритм PLC scan

```text
on every scan tick:
    start_timer
    dt = now - previous_scan_time
    inputs = simulator.Read()
    for each loop:
        if loop disabled:
            output = safe output
        else if mode == auto:
            output = pid.Update(sp, pv, dt)
        else if mode == manual:
            output = manual output
        else if mode == hold:
            output = previous output
    simulator.ApplyOutputs(outputs)
    simulator.Step(dt)
    snapshot = build snapshot
    publish snapshot internally
    persist snapshot
    maybe publish MQTT telemetry
    maybe send SSE update
    if scan_duration > threshold:
         emit scan_overrun event
```

### 25.3. Алгоритм обработки MQTT-команды

```text
receive MQTT message
parse JSON
validate command
convert to internal Command
send to PLC command channel
wait/apply asynchronously
persist command
publish event command_applied or command_rejected
update UI via SSE
```

### 25.4. Алгоритм изменения setpoint

```text
find loop by name
validate setpoint range
old = current setpoint
set new setpoint
emit event
persist command
```

### 25.5. Алгоритм изменения PID gains

```text
find loop by name
validate kp/ki/kd
store old gains
apply new gains
emit event
persist command
```

### 25.6. Алгоритм генерации шума

```text
noise = normal_random(0, stddev)
pv = pv + noise
```

### 25.7. Алгоритм случайного возмущения

```text
if random probability triggered:
    create disturbance with amplitude and duration
while disturbance active:
    process target += disturbance amplitude
when duration expired:
    remove disturbance
```

### 25.8. Алгоритм хранения telemetry

```text
for each snapshot:
    if storage enabled:
        for each loop:
            insert telemetry row
    if retention limit exceeded:
        delete oldest rows
```

### 25.9. Алгоритм SSE broadcast

```text
when new snapshot:
    if enough time since last UI update:
        encode JSON
        send to all connected clients
        remove disconnected clients
```

---

## 26. Пользовательские сценарии

### 26.1. Сценарий 1 — Быстрый запуск демо

Пользователь хочет быстро посмотреть работу проекта.

Шаги:

1. запускает MQTT broker;
2. запускает приложение;
3. открывает dashboard;
4. видит графики;
5. меняет setpoint;
6. наблюдает реакцию PID.

Критерий успеха:

- графики обновляются;
- MQTT connected;
- события отображаются в terminal;
- значения PV приближаются к SP.

### 26.2. Сценарий 2 — Проверка MQTT telemetry

Пользователь хочет подключить внешний MQTT client.

Шаги:

1. запускает приложение;
2. запускает `mosquitto_sub`:

```bash
mosquitto_sub -h localhost -t "vplc/vplc_001/telemetry" -v
```

3. видит JSON telemetry.

Критерий успеха:

- сообщения приходят с заданным интервалом;
- JSON валиден;
- содержит loops и PLC status.

### 26.3. Сценарий 3 — Управление через MQTT

Пользователь отправляет команду:

```bash
mosquitto_pub -h localhost -t "vplc/vplc_001/commands" -m '{"command_id":"cmd-1","command":"set_setpoint","loop":"pressure","value":7.5}'
```

Критерий успеха:

- setpoint меняется;
- UI обновляется;
- event публикуется;
- команда записывается в storage.

### 26.4. Сценарий 4 — Настройка PID

Пользователь меняет `Kp`, `Ki`, `Kd` в UI.

Критерий успеха:

- новые коэффициенты применяются;
- график показывает другое поведение;
- событие фиксируется.

### 26.5. Сценарий 5 — Демо для собеседования

Кандидат показывает проект интервьюеру.

Он объясняет:

- virtual PLC scan cycle;
- PID loops;
- synthetic process;
- MQTT telemetry;
- local dashboard;
- logs/history;
- почему проект может быть переиспользован.

Критерий успеха:

- проект запускается локально;
- демонстрация занимает меньше 5 минут;
- архитектура легко объясняется.

---

## 27. Обработка ошибок

### 27.1. Принципы

- Ошибки должны быть явными.
- Ошибки конфигурации должны останавливать запуск.
- Ошибки MQTT не должны останавливать PLC runtime.
- Ошибки storage не должны ломать PID-регулирование, но должны логироваться.
- Ошибки UI не должны влиять на backend.
- Некорректные команды должны отклоняться безопасно.

### 27.2. Ошибка конфигурации

Пример:

```text
configuration error: loop pressure setpoint 30.0 is outside range [0.0, 12.0]
```

### 27.3. MQTT disconnected

Поведение:

- PLC продолжает работать;
- UI показывает MQTT disconnected;
- приложение пытается переподключиться;
- события пишутся в log.

### 27.4. Storage error

Поведение:

- записать ERROR;
- показать warning в UI;
- продолжить scan cycle.

### 27.5. Invalid command

Поведение:

- команда не применяется;
- публикуется `command_rejected`;
- событие пишется в storage;
- UI показывает сообщение.

### 27.6. Panic recovery

В goroutines верхнего уровня должен быть recover с логированием.

Но recover не должен скрывать архитектурные ошибки. В тестах panic должен исправляться, а не подавляться.

---

## 28. Тестирование

### 28.1. Уровни тестирования

| Уровень | Что тестируем |
|---|---|
| Unit | PID, simulator, config validation |
| Integration | PLC runtime + simulator, MQTT client |
| System | запуск приложения, HTTP API, SSE |

### 28.2. Unit tests

Обязательно:

```bash
go test ./pkg/pid ./pkg/simulator ./internal/config
```

### 28.3. PID tests

Тесты описаны в разделе PID.

### 28.4. Simulator tests

Проверить:

- PV движется к target;
- PV не выходит за min/max;
- noise можно отключить для детерминированных тестов;
- disturbance влияет на PV;
- при фиксированном seed результат воспроизводим.

### 28.5. Config tests

Проверить:

- валидный config проходит;
- пустой device_id отклоняется;
- неправильный port отклоняется;
- дубли loop names отклоняются;
- setpoint вне диапазона отклоняется.

### 28.6. PLC runtime tests

Проверить:

- runtime стартует;
- runtime останавливается;
- scan counter растёт;
- command меняет setpoint;
- snapshot формируется;
- disabled loop не обновляется.

### 28.7. MQTT integration tests

Интеграционные MQTT-тесты должны быть опциональными.

Запускать только если задан env:

```bash
VPLC_RUN_MQTT_TESTS=1 go test ./tests/integration/...
```

### 28.8. Race detector

```bash
go test -race ./...
```

### 28.9. Static checks

```bash
go vet ./...
gofmt -w .
go test ./...
```

Если используется `golangci-lint`, добавить:

```bash
golangci-lint run
```

---

## 29. Производительность

### 29.1. Целевые показатели

MVP должен уверенно работать при:

- 3–10 PID-контурах;
- scan interval 100–1000 ms;
- MQTT publish interval 500–1000 ms;
- UI update rate 4–10 Hz;
- хранении 100 000 samples.

### 29.2. Ограничение UI update rate

Нельзя отправлять в UI каждую точку при слишком быстром scan.

Нужно ограничить:

```text
ui_update_interval_ms >= 100
```

### 29.3. In-memory ring buffer

Для графиков хранить последние N точек:

```text
300–1000 points per loop
```

### 29.4. Storage batching

На первом этапе можно писать каждый snapshot напрямую.

Если появятся проблемы, добавить batching:

```text
flush every 1 second or 100 samples
```

### 29.5. MQTT batching

Не нужно публиковать каждую scan-точку. Publish interval может быть больше scan interval.

Пример:

```text
scan interval = 100 ms
publish interval = 1000 ms
```

---

## 30. Безопасность

### 30.1. Локальный характер приложения

По умолчанию web server должен слушать только:

```text
127.0.0.1
```

Не использовать `0.0.0.0` без явного указания пользователя.

### 30.2. MQTT credentials

Если MQTT username/password используются, не логировать пароль.

В логах допустимо:

```text
mqtt username configured: true
mqtt password configured: true
```

Недопустимо:

```text
password=secret123
```

### 30.3. Команды

Поскольку MQTT broker локальный и без авторизации в MVP, в README нужно указать, что это демонстрационный режим.

Для промышленного использования нужна авторизация и TLS.

### 30.4. Реальное оборудование

Запрещено позиционировать проект как систему управления реальным оборудованием.

Нужно указать:

```text
Do not connect this simulator directly to real actuators or safety-critical systems.
```

### 30.5. Валидация входных данных

Все HTTP и MQTT команды валидировать.

Особое внимание:

- NaN;
- Infinity;
- слишком большие числа;
- неизвестные поля;
- слишком длинные payloads;
- неизвестные команды.

---

## 31. Сборка и запуск

### 31.1. Требования

- Go 1.26;
- Docker Desktop для запуска Mosquitto через Docker Compose;
- Windows 11 x64 как основная целевая среда.

### 31.2. Запуск в режиме разработки

```bash
go run ./cmd/vplc --config configs/default.json
```

### 31.3. Проверка конфигурации

```bash
go run ./cmd/vplc --validate-config configs/default.json
```

### 31.4. Сборка

Windows:

```powershell
go build -o dist/vplc.exe ./cmd/vplc
```

Linux/macOS:

```bash
go build -o dist/vplc ./cmd/vplc
```

### 31.5. Запуск собранного файла

```bash
./dist/vplc --config configs/default.json
```

### 31.6. Тесты

```bash
go test ./...
```

### 31.7. Race detector

```bash
go test -race ./...
```

### 31.8. Форматирование

```bash
gofmt -w .
```

### 31.9. Минимальный release package

```text
release/
├─ vplc.exe
├─ configs/
│  └─ default.json
├─ docker/
│  └─ docker-compose.yml
├─ README.md
└─ LICENSE
```

---

## 32. Docker Compose для MQTT

### 32.1. Назначение

Docker Compose нужен только для локального MQTT-брокера.

### 32.2. docker-compose.yml

```yaml
services:
  mosquitto:
    image: eclipse-mosquitto:2
    container_name: vplc-mosquitto
    ports:
      - "1883:1883"
      - "9001:9001"
    volumes:
      - ./mosquitto/mosquitto.conf:/mosquitto/config/mosquitto.conf:ro
```

### 32.3. mosquitto.conf

```text
listener 1883
allow_anonymous true

listener 9001
protocol websockets
allow_anonymous true
```

### 32.4. Запуск

```bash
docker compose -f docker/docker-compose.yml up -d
```

### 32.5. Остановка

```bash
docker compose -f docker/docker-compose.yml down
```

### 32.6. Проверка telemetry

```bash
mosquitto_sub -h localhost -t "vplc/vplc_001/telemetry" -v
```

---

## 33. Стиль кода

### 33.1. Общие принципы

- стандартная библиотека там, где это разумно;
- минимум глобальных переменных;
- явная обработка ошибок;
- маленькие функции;
- понятные имена;
- без преждевременной абстракции;
- без фреймворков ради фреймворков.

### 33.2. Именование

Go packages:

```text
pid
plc
simulator
mqttx
storage
config
web
```

Не использовать слишком общие имена вроде:

```text
utils
helpers
common
```

### 33.3. Ошибки

Ошибки оборачивать:

```go
return fmt.Errorf("load config: %w", err)
```

### 33.4. Context

Все долгоживущие goroutines должны принимать `context.Context`.

### 33.5. Channels

Channels использовать для:

- snapshots;
- commands;
- events;
- shutdown.

Не использовать channels там, где достаточно обычного вызова функции.

### 33.6. Комментарии

Комментарии нужны для инженерной логики:

- почему выбран такой PID algorithm;
- почему используется anti-windup;
- почему SSE, а не WebSocket;
- почему MQTT publish interval отделён от scan interval.

Не нужно комментировать очевидное.

### 33.7. Tests as documentation

Table-driven tests должны показывать expected behavior.

---

## 34. План реализации по этапам

### Этап 1 — Foundation

Цель: создать структуру проекта и минимальный запуск.

Сделать:

- `go.mod`;
- структуру директорий;
- CLI flags;
- config loader;
- logger;
- version command;
- README skeleton;
- базовые тесты.

Definition of Done:

- `go run ./cmd/vplc --version` работает;
- `go test ./...` проходит;
- `gofmt` применён.

### Этап 2 — PID package

Цель: реализовать чистый PID-регулятор.

Сделать:

- `pkg/pid/controller.go`;
- config/state;
- output limits;
- anti-windup;
- manual/auto mode;
- unit tests.

Definition of Done:

- PID тесты проходят;
- нет зависимости от MQTT/UI/storage.

### Этап 3 — Simulator package

Цель: реализовать простую process model.

Сделать:

- first-order process;
- noise;
- disturbance;
- deterministic seed;
- min/max clamp;
- tests.

Definition of Done:

- PV реагирует на MV;
- тесты детерминированы.

### Этап 4 — PLC runtime

Цель: связать PID и simulator в scan cycle.

Сделать:

- runtime states;
- scan loop;
- loop registry;
- command channel;
- snapshots;
- events;
- tests.

Definition of Done:

- scan counter растёт;
- setpoint command влияет на процесс;
- runtime gracefully stops.

### Этап 5 — MQTT

Цель: добавить MQTT telemetry и commands.

Сделать:

- MQTT client;
- telemetry publisher;
- command subscriber;
- status topic;
- LWT;
- reconnect;
- integration test optional.

Definition of Done:

- `mosquitto_sub` видит telemetry;
- `mosquitto_pub` меняет setpoint.

### Этап 6 — Storage and logging

Цель: сохранять события и telemetry.

Сделать:

- app.log;
- events.jsonl;
- SQLite schema;
- insert telemetry;
- insert commands;
- recent events API.

Definition of Done:

- данные сохраняются;
- приложение переживает ошибку storage без падения scan loop.

### Этап 7 — Local web UI

Цель: сделать dashboard.

Сделать:

- embedded web assets;
- `/api/status`;
- `/api/snapshot`;
- `/api/stream` SSE;
- charts;
- terminal;
- controls.

Definition of Done:

- dashboard показывает live values;
- setpoint можно изменить через UI;
- terminal показывает events.

### Этап 8 — Polish for portfolio

Цель: подготовить проект к публикации.

Сделать:

- README with screenshots;
- architecture diagram;
- demo script;
- example commands;
- release build scripts;
- `.env.example` если нужно;
- GitHub Actions.

Definition of Done:

- новый человек может запустить проект по README за 5–10 минут.

---

## 35. Критерии готовности MVP

MVP считается готовым, если выполняется всё ниже:

1. Проект запускается одной командой.
2. MQTT broker запускается через Docker Compose.
3. Приложение подключается к MQTT.
4. Есть минимум 3 PID-контура: pressure, temperature, level.
5. PID output влияет на synthetic process.
6. PV постепенно стремится к SP.
7. Dashboard показывает live charts.
8. Dashboard показывает live terminal.
9. Setpoint можно менять через UI.
10. Setpoint можно менять через MQTT command.
11. MQTT telemetry публикуется регулярно.
12. Events пишутся в `events.jsonl`.
13. Telemetry пишется в SQLite или альтернативное хранилище.
14. Некорректные команды отклоняются безопасно.
15. `go test ./...` проходит.
16. `go test -race ./...` не показывает гонок для core-модулей.
17. README объясняет запуск и архитектуру.
18. В интерфейсе и README указано, что это симулятор, не система реального управления.

---

## 36. Кейс для портфолио

### 36.1. Короткое описание

`virtual-plc-pid-mqtt-r` is a lightweight virtual PLC simulator written in Go. It includes PID control loops, synthetic process simulation, MQTT telemetry and command handling, local historical logging and an embedded web dashboard for live monitoring.

### 36.2. Проблема

В промышленной автоматизации часто требуется тестировать IIoT pipeline, dashboard, обработку телеметрии или алгоритмы управления без подключения к реальному оборудованию. Реальный ПЛК, датчики и исполнительные механизмы не всегда доступны, а их использование может быть дорогим или небезопасным.

Проект решает эту проблему через виртуальный ПЛК, который имитирует:

- входные сигналы;
- PID-регулирование;
- управляющие выходы;
- MQTT-телеметрию;
- локальный мониторинг;
- запись истории.

### 36.3. Что демонстрирует проект

Проект показывает навыки:

- Go backend/edge development;
- промышленная автоматизация;
- PID-регулирование;
- MQTT/IIoT;
- simulation/digital twin thinking;
- time-series telemetry;
- локальный web dashboard;
- инженерное логирование;
- тестируемая архитектура.

### 36.4. Что сказать на собеседовании

Короткая формулировка:

> I built a reusable virtual PLC simulator in Go. It runs configurable PID control loops over synthetic process variables, publishes telemetry through MQTT, accepts remote commands, stores logs and history locally, and provides an embedded web dashboard for live monitoring and demonstration.

Расширенная формулировка:

> The main idea was to create a compact industrial automation building block that can be reused in digital twin and IIoT prototypes. The application simulates a PLC scan cycle, process sensors, PID controllers and actuator outputs. It exposes telemetry through MQTT and provides a local operator-like dashboard with trends, controls and event logs.

### 36.5. Почему проект хорошо подходит под карьерную цель

Потому что он находится на пересечении:

- АСУТП;
- программирования;
- цифровых двойников;
- industrial analytics;
- IIoT;
- моделирования процессов;
- backend/edge development.

---

## 37. Глоссарий

### PLC

Programmable Logic Controller. Промышленный контроллер, выполняющий циклическую программу управления.

### Scan cycle

Один цикл работы ПЛК: чтение входов, выполнение логики, запись выходов.

### PID

Proportional-Integral-Derivative controller. Регулятор, который рассчитывает управляющее воздействие по ошибке между заданием и текущим значением.

### SP

Setpoint. Заданное значение.

### PV

Process Value. Текущее измеренное значение процесса.

### MV

Manipulated Value. Управляющее воздействие.

### MQTT

Лёгкий publish/subscribe протокол, часто используемый в IoT и IIoT.

### Telemetry

Данные о состоянии устройства или процесса, публикуемые наружу.

### Historian

Система хранения временных рядов промышленного процесса.

### Digital Twin

Цифровая модель объекта или процесса, которая воспроизводит существенные свойства реального объекта.

### Synthetic sensor

Программно сгенерированный сигнал, имитирующий датчик.

### Disturbance

Возмущение процесса, например скачок нагрузки или шум.

### Anti-windup

Метод защиты PID-интегратора от накопления слишком большой ошибки при насыщении output.

### SSE

Server-Sent Events. Простой механизм передачи событий от сервера к браузеру.

---

## 38. Приложение A. Примеры конфигураций

### 38.1. default.json

```json
{
  "app": {
    "name": "virtual-plc-pid-mqtt-r",
    "device_id": "vplc_001",
    "auto_start": true,
    "open_browser": true
  },
  "plc": {
    "scan_interval_ms": 500,
    "publish_interval_ms": 1000,
    "ui_update_interval_ms": 250,
    "scan_overrun_warning_ms": 500
  },
  "mqtt": {
    "enabled": true,
    "broker_url": "tcp://localhost:1883",
    "client_id": "virtual-plc-pid-mqtt-r-vplc-001",
    "username": "",
    "password": "",
    "base_topic": "vplc/vplc_001",
    "qos": 0,
    "connect_timeout_seconds": 5,
    "reconnect_interval_seconds": 3
  },
  "web": {
    "enabled": true,
    "host": "127.0.0.1",
    "port": 8080
  },
  "storage": {
    "enabled": true,
    "type": "sqlite",
    "sqlite_path": "data/history.db",
    "events_jsonl_path": "logs/events.jsonl",
    "app_log_path": "logs/app.log",
    "retention_max_samples": 100000
  },
  "loops": [
    {
      "name": "pressure",
      "display_name": "Pressure",
      "unit": "bar",
      "enabled": true,
      "mode": "auto",
      "setpoint": 6.0,
      "setpoint_min": 0.0,
      "setpoint_max": 12.0,
      "pid": {
        "kp": 3.0,
        "ki": 0.25,
        "kd": 0.05,
        "bias": 0.0,
        "output_min": 0.0,
        "output_max": 100.0
      },
      "process": {
        "initial_pv": 4.0,
        "min": 0.0,
        "max": 12.0,
        "base": 0.0,
        "gain": 0.10,
        "tau_seconds": 15.0,
        "noise_stddev": 0.03,
        "random_disturbances": true
      }
    },
    {
      "name": "temperature",
      "display_name": "Temperature",
      "unit": "C",
      "enabled": true,
      "mode": "auto",
      "setpoint": 180.0,
      "setpoint_min": 0.0,
      "setpoint_max": 250.0,
      "pid": {
        "kp": 1.8,
        "ki": 0.10,
        "kd": 0.02,
        "bias": 0.0,
        "output_min": 0.0,
        "output_max": 100.0
      },
      "process": {
        "initial_pv": 80.0,
        "min": 0.0,
        "max": 250.0,
        "base": 20.0,
        "gain": 2.0,
        "tau_seconds": 60.0,
        "noise_stddev": 0.2,
        "random_disturbances": true
      }
    },
    {
      "name": "level",
      "display_name": "Level",
      "unit": "%",
      "enabled": true,
      "mode": "auto",
      "setpoint": 50.0,
      "setpoint_min": 0.0,
      "setpoint_max": 100.0,
      "pid": {
        "kp": 2.5,
        "ki": 0.15,
        "kd": 0.03,
        "bias": 0.0,
        "output_min": 0.0,
        "output_max": 100.0
      },
      "process": {
        "initial_pv": 45.0,
        "min": 0.0,
        "max": 100.0,
        "base": 0.0,
        "gain": 1.0,
        "tau_seconds": 25.0,
        "noise_stddev": 0.1,
        "random_disturbances": true
      }
    }
  ]
}
```

---

## 39. Приложение B. Примеры MQTT-сообщений

### 39.1. Subscribe telemetry

```bash
mosquitto_sub -h localhost -t "vplc/vplc_001/telemetry" -v
```

### 39.2. Subscribe events

```bash
mosquitto_sub -h localhost -t "vplc/vplc_001/events" -v
```

### 39.3. Change pressure setpoint

```bash
mosquitto_pub -h localhost -t "vplc/vplc_001/commands" -m '{"command_id":"cmd-pressure-001","command":"set_setpoint","loop":"pressure","value":7.5}'
```

### 39.4. Change PID gains

```bash
mosquitto_pub -h localhost -t "vplc/vplc_001/commands" -m '{"command_id":"cmd-pressure-002","command":"set_pid_gains","loop":"pressure","kp":2.5,"ki":0.3,"kd":0.05}'
```

### 39.5. Set manual mode

```bash
mosquitto_pub -h localhost -t "vplc/vplc_001/commands" -m '{"command_id":"cmd-temp-001","command":"set_mode","loop":"temperature","mode":"manual","manual_output":40}'
```

### 39.6. Return to auto mode

```bash
mosquitto_pub -h localhost -t "vplc/vplc_001/commands" -m '{"command_id":"cmd-temp-002","command":"set_mode","loop":"temperature","mode":"auto"}'
```

### 39.7. Stop PLC

```bash
mosquitto_pub -h localhost -t "vplc/vplc_001/commands" -m '{"command_id":"cmd-stop-001","command":"stop_plc"}'
```

### 39.8. Start PLC

```bash
mosquitto_pub -h localhost -t "vplc/vplc_001/commands" -m '{"command_id":"cmd-start-001","command":"start_plc"}'
```

---

## 40. Приложение C. Чеклист разработчика

### 40.1. До начала работы

- [ ] Прочитать данный документ полностью.
- [ ] Убедиться, что проект называется `virtual-plc-pid-mqtt-r`.
- [ ] Убедиться, что `-r` является частью имени репозитория.
- [ ] Убедиться, что целевой стек — Go 1.26.
- [ ] Не добавлять OPC UA, InfluxDB, Grafana и микросервисы в MVP.
- [ ] Согласиться с архитектурой single-binary + embedded web UI.

### 40.2. При реализации

- [ ] Не смешивать PID-математику с MQTT/UI.
- [ ] Не делать глобальное mutable state без необходимости.
- [ ] Все long-running goroutines должны иметь context cancellation.
- [ ] Все команды должны валидироваться.
- [ ] Все ошибки должны логироваться.
- [ ] Все публичные payloads должны иметь JSON examples.
- [ ] Все важные модули должны иметь тесты.

### 40.3. Перед сдачей MVP

- [ ] `gofmt` применён.
- [ ] `go test ./...` проходит.
- [ ] `go test -race ./...` проверен.
- [ ] MQTT telemetry проверена через `mosquitto_sub`.
- [ ] MQTT commands проверены через `mosquitto_pub`.
- [ ] Dashboard открывается.
- [ ] Графики обновляются.
- [ ] Terminal показывает события.
- [ ] Logs пишутся.
- [ ] Telemetry сохраняется.
- [ ] README обновлён.
- [ ] Добавлены screenshots или gif/video demo.

### 40.4. Для портфолио

- [ ] README содержит короткое описание проекта на английском.
- [ ] README содержит архитектурную схему.
- [ ] README содержит quick start.
- [ ] README содержит MQTT examples.
- [ ] README содержит screenshots dashboard.
- [ ] README объясняет, что проект является simulator/demo, а не real control system.
- [ ] В репозитории нет секретов.
- [ ] Есть понятный release build.

---

## Финальное резюме проекта

`virtual-plc-pid-mqtt-r` должен стать компактным, понятным и практически полезным виртуальным ПЛК на Go. Его ценность не в сложности, а в правильно выбранной инженерной границе: scan cycle, PID, synthetic process, MQTT, local dashboard, logs and history.

Проект должен быть проще старой микросервисной версии, но сильнее как reusable-компонент и портфолио-кейс. Он должен быстро запускаться, хорошо демонстрироваться, легко объясняться и быть достаточно чистым внутри, чтобы его можно было развивать дальше.

Главная формула проекта:

```text
virtual PLC + PID + synthetic process + MQTT + local web UI = reusable IIoT/digital twin building block
```
