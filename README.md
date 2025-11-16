# **Тестовое задание для стажёра Backend (осенняя волна 2025)**

## **Сервис назначения ревьюеров для Pull Request’ов**

Система управления pull request'ами с автоматическим распределением reviewers.
## Установка

### 1. Клонирование репозитория

```bash
git clone https://github.com/nkozlov1/PullRequestApi
cd PullRequestApi
```

### 2. Настройка окружения

Создайте файл `.env` в корне проекта:

В качестве примере есть файл `.env.example`
```bash
cp .env.example .env
```

Пример содержимого `.env`:
```env
HTTP_PORT=8080

POSTGRES_USER=root
POSTGRES_PASSWORD=root
POSTGRES_HOST=database
POSTGRES_PORT=5432
POSTGRES_DB=root

DATABASE_URL=postgres://root:root@database:5432/root?sslmode=disable

GIN_MODE=release
```

## Запуск

### Запуск приложения
```bash
make docker-up
```

Приложение будет доступно по адресу: `http://localhost:8080`

**Остановка:**
```bash
make docker-down
```

### Добавление новой миграции

```bash
make migrate-install

# Создать файлы миграций
migrate create -ext sql -dir migrations -seq <migration_name>

# Применить
make migrate-up
```

## Makefile команды

### Основные команды

```bash
make help                 # Показать все доступные команды
make docker-up            # Запустить все сервисы в Docker
make docker-down          # Остановить все сервисы
make docker-restart       # Перезапустить приложение
make docker-down-volumes: # Остановить все сервисы и удалить volumes
make docker-restart:      # Перезапустить приложение
```

### Миграции базы данных

```bash
make migrate-up        # Применить все миграции
make migrate-down      # Откатить все миграции
```

### Тестирование

```bash
make test              # Запустить интеграционные тесты
```

**Тесты используют отдельную БД `root_test`** и не влияют на основную базу данных.

### Управление тестовой БД

```bash
make test-db-create    # Создать тестовую базу данных
make test-db-migrate   # Применить миграции на тестовую БД
make test-db-drop      # Удалить тестовую базу данных
```

## Структура проекта

```
.
├── cmd/
│   └── api/
│       └── main.go           # Точка входа приложения
├── pkg/
│   ├── config/               # Конфигурация
│   ├── domain/               # Доменные модели
│   ├── gateway/              # HTTP handlers
│   │   ├── pullrequest/      # PR endpoints
│   │   ├── team/             # Team endpoints
│   │   └── user/             # User endpoints
│   ├── repo/                 # Репозитории
│   │   └── pg/               # PostgreSQL реализация
│   └── usecase/              # Бизнес-логика
├── migrations/               # SQL миграции
├── test/                     # Интеграционные тесты
│   ├── integration_test.go   # Все тесты
├── docker-compose.yaml       # Docker Compose конфигурация
├── Dockerfile                # Docker образ приложения
├── Makefile                  # Команды сборки и управления
├── openapi.yml               # OpenAPI спецификация
├── go.mod                    # Go модули
└── README.md                 # Этот файл
```

## API Endpoints

Подробную спецификацию API см. в `openapi.yml`.

## Тестирование

В качестве дополнительного задание были добавленные интеграционные тестирование с реальной бд:

Для запуска тестов
```bash
make test
```

**Покрытие кода:** ~75% репозиториев

## Переменные окружения
| Переменная | Описание | По умолчанию |
|------------|----------|--------------|
| `HTTP_PORT` | Порт HTTP сервера | `8080` |
| `POSTGRES_USER` | Пользователь PostgreSQL | `root` |
| `POSTGRES_PASSWORD` | Пароль PostgreSQL | `root` |
| `POSTGRES_HOST` | Хост PostgreSQL | `database` (Docker) / `localhost` (local) |
| `POSTGRES_PORT` | Порт PostgreSQL | `5432` |
| `POSTGRES_DB` | Имя базы данных | `root` |
| `DATABASE_URL` | Полный URL подключения к БД | auto-generated |
| `GIN_MODE` | Режим Gin (`debug`/`release`) | `release` |