# Scoring Worker

Сервис обработки данных компаний для системы скоринга. Получает задачи из очереди NATS, обрабатывает данные через API Credinform и сохраняет результаты в базу данных.

## Архитектура

- **NATS** - транспорт для обмена сообщениями
- **PostgreSQL** - база данных для хранения результатов
- **Credinform API** - внешний источник данных о компаниях

## Конфигурация

### Переменные окружения

Создайте файл `.env` или установите переменные окружения:

```bash
# Credinform API
CREDINFORM_USERNAME=your_username
CREDINFORM_PASSWORD=your_password
CREDINFORM_BASE_URL=https://api.credinform.ru

# База данных
DATABASE_HOST=postgres
DATABASE_PORT=5432
DATABASE_USER=postgres
DATABASE_PASSWORD=postgres
DATABASE_DBNAME=scoring
DATABASE_SSLMODE=disable

# NATS
NATS_URL=nats://nats:4222

# Логирование
LOG_LEVEL=info
LOG_JSON=false
```

### Конфигурационный файл

Основные настройки в `config.yaml`:

```yaml
database:
  host: "postgres"
  port: 5432
  user: "postgres"
  password: "postgres"
  dbname: "scoring"
  sslmode: "disable"

nats:
  url: "nats://nats:4222"

log:
  level: "info"
  json: false

credinform:
  base_url: "https://api.credinform.ru"
  username: "" # Устанавливается через переменную окружения
  password: "" # Устанавливается через переменную окружения
  timeout: 30
  retry_attempts: 3
  retry_delay: 1
```

## Поддерживаемые типы данных

- `basic_information` - основная информация о компании
- `activities` - виды деятельности компании

## Запуск

### Локально

```bash
go run main.go
```

### В Docker

```bash
docker build -t scoring-worker .
docker run --env-file .env scoring-worker
```

## Логирование

Сервис использует структурированное логирование с помощью zap. Логи включают:

- Информацию о полученных задачах
- Статус обработки каждого типа данных
- Ошибки при работе с API
- Результаты сохранения в базу данных

## Обработка ошибок

- Автоматический retry при сбоях API
- Повторная авторизация при истечении токена
- Сохранение информации об ошибках в базу данных
- Уведомления через NATS о статусе обработки

## Масштабирование

Сервис спроектирован для горизонтального масштабирования:

- Несколько экземпляров могут работать параллельно
- NATS обеспечивает распределение нагрузки
- Каждый экземпляр обрабатывает задачи независимо
