# hllb — DNS-сервер с балансировкой нагрузки

Авторитетный DNS-сервер на Go с поддержкой wildcard-зон, активной проверкой доступности хостов и форвардингом.

---

## Быстрый старт

### Сборка

```bash
go build -o hllb .
```

### Первый запуск

При первом запуске автоматически создаются отсутствующие файлы:
- `config.yaml` — конфигурация сервера (порт 53, forward на 8.8.8.8)
- `check.yaml` — список хостов для health check
- `zone/example.com` — тестовая зона

```bash
./hllb
```

### Запуск на порту 53

Порт 53 — привилегированный (< 1024), для работы на нём требуются дополнительные права.

#### Linux

```bash
# Рекомендуемый способ — дать бинарнику право на привилегированные порты
sudo setcap 'cap_net_bind_service=+ep' ./hllb
./hllb

# Или запуск от root (не рекомендуется)
sudo ./hllb
```

Если порт 53 уже занят `systemd-resolved`:
```bash
sudo ss -tlnp | grep :53
sudo systemctl disable --now systemd-resolved
```

#### macOS

```bash
# Запуск от root
sudo ./hllb
```

Альтернатива — слушать на высоком порту и пробросить через pfctl:
```bash
# В config.yaml: port: 1053
echo "rdr pass on lo0 inet proto {tcp, udp} from any to 127.0.0.1 port 53 -> 127.0.0.1 port 1053" | sudo pfctl -ef -
```

#### Windows

Порты ниже 1024 в Windows не привилегированные — дополнительных прав не требуется. Убедитесь, что порт 53 не занят службой DNS Client:
```powershell
netstat -ano | findstr :53
net stop "DNS Client"
```

### Проверка работы

```bash
dig @127.0.0.1 -p 53 example.com A
```

---

## Возможности

### DNS-сервер
- Слушает входящие запросы одновременно по **UDP и TCP** на настраиваемом порту
- Обрабатывает типы запросов **A** и **NS**
- Флаг **Authoritative** устанавливается только для ответов из собственных зон
- **Форвардинг** — запросы, не найденные в зонах, пересылаются на внешний DNS (настраивается в `config.yaml`)

### Зоны и записи

Файлы зон хранятся в директории `./zone/`. Имя файла — имя зоны (например `test.ru`), оно же становится `ORIGIN` — корневым доменом, подставляемым к относительным записям внутри файла.

Поддерживаемые типы записей в зонах:
| Тип | Пример в zone-файле | Описание |
|-----|---------------------|----------|
| `A` | `www IN A 10.13.1.34` | IPv4-адрес |
| `NS` | `@ IN NS ns1.test.ru.` | Сервер имён |

### Wildcard-матчинг (три уровня)

| Паттерн | Пример запроса | Описание |
|---------|----------------|----------|
| `*` (корень зоны) | `any.test.ru` | Любой поддомен зоны |
| `*.suffix` | `x.info.test.ru` | Любой поддомен для `*.info.test.ru` |
| `*.sub.domain` | `a.b.msg.admin.test.ru` | Любая вложенность субдоменов |
| `exact.domain` | `sub.admin.test.ru` | Точное совпадение записи |

Порядок приоритета обработки запроса:
1. Wildcard-записи зоны (`*.suffix`)
2. Точное совпадение
3. Wildcard-фолбэк (`*.zone`)
4. Форвардинг на внешний DNS (если `forward: true`)
5. `NXDOMAIN`

### Горячая перезагрузка зон
- Каждый zone-файл отслеживается в отдельной горутине
- Изменение файла обнаруживается через **SHA-256 хеш**
- При изменении — зона атомарно перезагружается без перезапуска сервера
- Интервал проверки настраивается: `checkZoneInterval` + `checkZoneIntervalType`

### Активная проверка хостов (Health Check)
Включается параметром `activeCheck: true` в `config.yaml`.

- Выполняет **TCP-проверку** порта для списка хостов из `check.yaml`
- При изменении - список хостов для провери и порт перезагружаются `check.yaml` (также через SHA-256)
- Алгоритм балансировки **Round Robin** применяется только к записям, чьи IP находятся в пуле `check.yaml`
- Записи, не относящиеся к пулу, отдаются напрямую из зоны
- Интервалы настраиваются отдельно: `repeatCheckInterval` и `repeatCheckFileInterval`

### Форвардинг
Включается параметром `forward: true` в `config.yaml`.

- Запросы, не найденные ни в одной зоне, пересылаются на указанный внешний DNS-сервер
- Настраивается адрес (`forwardDNS`) и порт (`forwardDNSPort`)

---

## Конфигурация

### `config.yaml`
```yaml
app:
  port: 53                            # Порт DNS-сервера
  checkZoneInterval: 5                # Интервал проверки изменений зон
  checkZoneIntervalType: seconds      # Единица: seconds / minutes / hours
  activeCheck: true                   # Включить активную проверку хостов
  algorithmCheck: RR                  # Алгоритм балансировки (RR — Round Robin)
  repeatCheckInterval: 3              # Интервал TCP-проверки хостов
  repeatCheckIntervalType: seconds
  repeatCheckFileInterval: 3          # Интервал проверки изменений check.yaml
  repeatCheckFileIntervalType: seconds
  forward: true                       # Включить форвардинг для неизвестных зон
  forwardDNS: 8.8.8.8                # Адрес upstream DNS
  forwardDNSPort: 53                  # Порт upstream DNS
```

### `check.yaml`
```yaml
hostCheck:
  - 10.13.1.36        # IP хоста (должен совпадать с A-записью в зоне)
  - 176.125.254.184
portCheck: 22          # TCP-порт для проверки
```

---

## Структура проекта

```
.
├── main.go                    # Точка входа, запуск DNS-сервера
├── config.yaml                # Основная конфигурация
├── check.yaml                 # Список хостов для health check
├── zone/                      # DNS zone-файлы (стандартный формат RFC 1035)
│   └── test.ru
├── algorithm/
│   └── rr.go                  # Round Robin балансировка
├── checks/
│   └── tcpCheck.go            # TCP health check
├── handles/
│   ├── handleDNS.go           # Обработчик DNS-запросов
│   └── forwardDNS.go          # Форвардинг на upstream DNS
└── utils/
    ├── initFiles.go           # Создание файлов/папок при первом запуске
    ├── zoneParser.go          # Парсинг zone-файлов
    ├── watchZoneFile.go       # Слежение за изменениями zone-файлов
    ├── watchCheckFile.go      # Слежение за изменениями check.yaml
    ├── hashFile.go            # SHA-256 хеширование файлов
    ├── readConfig.go          # Чтение и кэширование config.yaml
    ├── readCheck.go           # Чтение check.yaml
    └── selectTime.go          # Конвертация единиц времени
```
--- 

### TEST Report - resperf
```
Машина с которой послылася запрос по wi-fi:
Mac-mini M2 16Gb

Машина на которую посылался:
Mac book Air M4 16Gb
```
```
Resperf report 20260325-1333
Resperf output
DNS Resolution Performance Testing Tool
Version 2.15.0

[Status] Command line: resperf -P 20260325-1333.gnuplot -s 10.13.1.18 -p 1053 -d test.info -R -C 10
[Status] Sending
[Status] Reached 65536 outstanding queries
[Status] Waiting for more responses
[Status] Testing complete

Statistics:

  Queries sent:         673250
  Queries completed:    608827
  Queries lost:         64423
  Response codes:       NOERROR 608827 (100.00%)

  Run time (s):         73.415265
  Maximum throughput:   35790.000000 qps
  Lost at that point:   16.61%

  Connection attempts:  0 (0 successful, 0.00%)
```
