# hllb — DNS-сервер с балансировкой нагрузки

Авторитетный DNS-сервер на Go с поддержкой wildcard-зон и активной проверкой доступности хостов.

---

## Возможности

### DNS-сервер
- Слушает входящие запросы одновременно по **UDP и TCP** на настраиваемом порту (по умолчанию `:1053`)
- Обрабатывает типы запросов **A** и **NS**
- Режим **авторитетного** сервера (`Authoritative = true`)

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
4. `NXDOMAIN`

### Горячая перезагрузка зон
- Каждый zone-файл отслеживается в отдельной горутине
- Изменение файла обнаруживается через **SHA-256 хеш**
- При изменении — зона атомарно перезагружается без перезапуска сервера (через `sync.RWMutex`)
- Интервал проверки настраивается: `checkZoneInterval` + `checkZoneIntervalType`

### Активная проверка хостов (Health Check)
Включается параметром `activeCheck: true` в `config.yaml`.

- Выполняет **TCP-проверку** порта для списка хостов из `check.yaml`
- Список хостов и порт горячо перезагружаются при изменении `check.yaml` (также через SHA-256)
- Результат проверки — глобальный список `ValidPoolHost` с живыми хостами
- Интервалы настраиваются отдельно: `repeatCheckInterval` и `repeatCheckFileInterval`

---

## Конфигурация

### `config.yaml`
```yaml
app:
  port: 1053                        # Порт DNS-сервера
  checkZoneInterval: 5              # Интервал проверки изменений зон
  checkZoneIntervalType: seconds    # Единица: seconds / minutes / hours
  activeCheck: true                 # Включить активную проверку хостов
  repeatCheckInterval: 3            # Интервал TCP-проверки хостов
  repeatCheckIntervalType: seconds
  repeatCheckFileInterval: 3        # Интервал проверки изменений check.yaml
  repeatCheckFileIntervalType: seconds
```

### `check.yaml`
```yaml
hostCheck:
  - lb01.test.ru      # Хост по имени
  - 176.125.254.184   # Хост по IP
portCheck: 22         # TCP-порт для проверки
```

---

## Структура проекта

```
.
├── main.go                    # Точка входа, запуск DNS-сервера
├── config.yaml                # Основная конфигурация
├── check.yaml                 # Список хостов для health check
├── zone/                      # DNS zone-файлы (стандартный формат RFC 1035)
│   ├── test.ru
│   └── web.ru
├── checks/
│   └── tcpCheck.go            # TCP health check
├── handles/
│   └── handleDNS.go           # Обработчик DNS-запросов
└── utils/
    ├── zoneParser.go          # Парсинг zone-файлов
    ├── watchZoneFile.go       # Слежение за изменениями zone-файлов
    ├── watchCheckFile.go      # Слежение за изменениями check.yaml
    ├── hashFile.go            # SHA-256 хеширование файлов
    ├── readConfig.go          # Чтение config.yaml
    ├── readCheck.go           # Чтение check.yaml
    └── selectTime.go          # Конвертация единиц времени
```
