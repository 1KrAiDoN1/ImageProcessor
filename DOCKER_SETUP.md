# Docker Setup Guide

Полное руководство по запуску ImageProcessor в Docker.

## 🐳 Архитектура

Проект состоит из следующих контейнеров:

1. **Frontend** (Nginx) - Веб-интерфейс на порту 80
2. **API** (Go) - REST API на порту 8080
3. **Worker** (Go) - Обработчик изображений (2 реплики)
4. **PostgreSQL** - База данных на порту 5432
5. **Kafka** - Брокер сообщений на порту 9092
6. **Zookeeper** - Координация Kafka на порту 2181
7. **MinIO** - S3 хранилище на портах 9000 (API) и 9001 (Console)

## 🚀 Быстрый старт

### 1. Клонировать репозиторий

```bash
git clone <repository-url>
cd ImageProcessor
```

### 2. Создать .env файл (опционально)

```bash
cp .env.example .env
# Отредактируйте .env при необходимости
```

### 3. Запустить все сервисы

```bash
docker-compose up -d
```

### 4. Проверить статус

```bash
docker-compose ps
```

Все сервисы должны быть в статусе "Up (healthy)".

### 5. Открыть веб-интерфейс

Откройте браузер: **http://localhost**

## 📋 Детальная инструкция

### Сборка образов

```bash
# Собрать все образы
docker-compose build

# Собрать только frontend
docker-compose build frontend

# Собрать только backend (API + Worker)
docker-compose build api worker

# Собрать с отключением кеша
docker-compose build --no-cache
```

### Запуск сервисов

```bash
# Запустить все сервисы
docker-compose up -d

# Запустить только инфраструктуру (без API и Worker)
docker-compose up -d postgres kafka zookeeper minio

# Запустить с логами в консоли
docker-compose up

# Запустить конкретный сервис
docker-compose up -d frontend
```

### Остановка сервисов

```bash
# Остановить все сервисы
docker-compose down

# Остановить с удалением volumes (ВНИМАНИЕ: удалит все данные!)
docker-compose down -v

# Остановить конкретный сервис
docker-compose stop frontend
```

### Перезапуск сервисов

```bash
# Перезапустить все
docker-compose restart

# Перезапустить конкретный сервис
docker-compose restart api
docker-compose restart worker
docker-compose restart frontend
```

## 📊 Мониторинг

### Просмотр логов

```bash
# Все сервисы
docker-compose logs -f

# Конкретный сервис
docker-compose logs -f frontend
docker-compose logs -f api
docker-compose logs -f worker

# Последние 100 строк
docker-compose logs --tail=100 api

# Логи за последний час
docker-compose logs --since 1h api
```

### Проверка здоровья

```bash
# Проверить все сервисы
docker-compose ps

# Проверить конкретный контейнер
docker inspect --format='{{.State.Health.Status}}' imageprocessor-api
docker inspect --format='{{.State.Health.Status}}' imageprocessor-frontend

# Health check endpoints
curl http://localhost/health          # Frontend health
curl http://localhost:8080/health     # API health
```

### Статистика ресурсов

```bash
# Использование ресурсов всех контейнеров
docker stats

# Использование ресурсов конкретного контейнера
docker stats imageprocessor-frontend
```

## 🔧 Управление

### Выполнение команд внутри контейнеров

```bash
# Открыть shell в контейнере
docker-compose exec frontend sh
docker-compose exec api sh
docker-compose exec postgres psql -U postgres -d imageprocessor

# Выполнить одну команду
docker-compose exec api ls -la
docker-compose exec postgres pg_dump -U postgres imageprocessor > backup.sql
```

### Работа с базой данных

```bash
# Подключиться к PostgreSQL
docker-compose exec postgres psql -U postgres -d imageprocessor

# Создать бэкап
docker-compose exec postgres pg_dump -U postgres imageprocessor > backup.sql

# Восстановить из бэкапа
docker-compose exec -T postgres psql -U postgres imageprocessor < backup.sql

# Запустить миграции
docker-compose exec postgres psql -U postgres -d imageprocessor -f /docker-entrypoint-initdb.d/001_init_schema.up.sql
```

### Работа с MinIO

```bash
# Открыть MinIO Console
# Откройте браузер: http://localhost:9001
# Логин: minioadmin
# Пароль: minioadmin

# Создать bucket через mc клиент
docker run --rm --network imageprocessor_imageprocessor-network \
  -it --entrypoint=/bin/sh minio/mc \
  -c "mc alias set myminio http://minio:9000 minioadmin minioadmin && mc mb myminio/images"
```

### Работа с Kafka

```bash
# Список топиков
docker-compose exec kafka kafka-topics --bootstrap-server localhost:9092 --list

# Создать топик
docker-compose exec kafka kafka-topics --bootstrap-server localhost:9092 \
  --create --topic image-processing --partitions 3 --replication-factor 1

# Описание топика
docker-compose exec kafka kafka-topics --bootstrap-server localhost:9092 \
  --describe --topic image-processing

# Читать сообщения из топика
docker-compose exec kafka kafka-console-consumer --bootstrap-server localhost:9092 \
  --topic image-processing --from-beginning
```

## 🌐 Доступ к сервисам

После запуска сервисы доступны по следующим адресам:

- **Frontend**: http://localhost
- **API**: http://localhost:8080
- **MinIO Console**: http://localhost:9001
- **PostgreSQL**: localhost:5432
- **Kafka**: localhost:9092

## 🔒 Безопасность

### Production настройки

Для production обязательно измените:

1. **Пароли базы данных**:
```yaml
environment:
  POSTGRES_PASSWORD: <strong-password>
```

2. **MinIO credentials**:
```yaml
environment:
  MINIO_ROOT_USER: <your-username>
  MINIO_ROOT_PASSWORD: <strong-password>
```

3. **CORS настройки** в `frontend/nginx.conf`:
```nginx
add_header Access-Control-Allow-Origin https://your-domain.com always;
```

4. **SSL/TLS** - используйте Let's Encrypt или другие сертификаты

### Ограничение доступа

```yaml
# В docker-compose.yaml, уберите expose портов наружу для внутренних сервисов
postgres:
  # ports:
  #   - "5432:5432"  # Закомментируйте эту строку
```

## 🐛 Решение проблем

### Контейнер не запускается

```bash
# Проверить логи
docker-compose logs <service-name>

# Проверить конфигурацию
docker-compose config

# Пересобрать образ
docker-compose build --no-cache <service-name>
docker-compose up -d <service-name>
```

### Ошибка "port already in use"

```bash
# Найти процесс, использующий порт
lsof -i :80
lsof -i :8080

# Остановить процесс или изменить порт в docker-compose.yaml
```

### Проблемы с сетью

```bash
# Пересоздать сеть
docker-compose down
docker network prune
docker-compose up -d

# Проверить сетевое подключение между контейнерами
docker-compose exec api ping postgres
docker-compose exec frontend wget -O- http://api:8080/health
```

### Ошибки базы данных

```bash
# Проверить что PostgreSQL запущен
docker-compose ps postgres

# Проверить логи
docker-compose logs postgres

# Перезапустить с чистым volume (ВНИМАНИЕ: потеряете данные!)
docker-compose down -v
docker-compose up -d postgres
```

### Frontend не подключается к API

1. Проверить, что API запущен:
```bash
curl http://localhost:8080/health
```

2. Проверить nginx конфигурацию:
```bash
docker-compose exec frontend nginx -t
```

3. Проверить логи frontend:
```bash
docker-compose logs frontend
```

4. Проверить что контейнеры в одной сети:
```bash
docker network inspect imageprocessor_imageprocessor-network
```

## 📈 Масштабирование

### Увеличить количество workers

```bash
# В docker-compose.yaml измените replicas
docker-compose up -d --scale worker=5
```

### Мониторинг нагрузки

```bash
# Проверить количество работающих workers
docker-compose ps worker

# Статистика использования ресурсов
docker stats

# Kafka lag (сколько сообщений в очереди)
docker-compose exec kafka kafka-consumer-groups --bootstrap-server localhost:9092 \
  --describe --group image-processor-workers
```

## 🔄 Обновление

### Обновить код без пересборки образов

```bash
# Если используете volume mount
docker-compose restart api worker frontend
```

### Обновить с пересборкой образов

```bash
# Пересобрать и перезапустить
docker-compose build
docker-compose up -d

# Или в одной команде
docker-compose up -d --build
```

## 📦 Резервное копирование

### Полный бэкап

```bash
#!/bin/bash
# backup.sh

BACKUP_DIR="./backups/$(date +%Y%m%d_%H%M%S)"
mkdir -p "$BACKUP_DIR"

# Бэкап базы данных
docker-compose exec -T postgres pg_dump -U postgres imageprocessor > "$BACKUP_DIR/database.sql"

# Бэкап MinIO данных
docker run --rm --network imageprocessor_imageprocessor-network \
  -v "$BACKUP_DIR:/backup" \
  -it --entrypoint=/bin/sh minio/mc \
  -c "mc alias set myminio http://minio:9000 minioadmin minioadmin && \
      mc mirror myminio/images /backup/minio"

echo "Backup completed: $BACKUP_DIR"
```

### Восстановление

```bash
#!/bin/bash
# restore.sh

BACKUP_DIR=$1

# Восстановить базу данных
docker-compose exec -T postgres psql -U postgres imageprocessor < "$BACKUP_DIR/database.sql"

# Восстановить MinIO
docker run --rm --network imageprocessor_imageprocessor-network \
  -v "$BACKUP_DIR:/backup" \
  -it --entrypoint=/bin/sh minio/mc \
  -c "mc alias set myminio http://minio:9000 minioadmin minioadmin && \
      mc mirror /backup/minio myminio/images"

echo "Restore completed"
```

## 🎯 Production Checklist

- [ ] Изменены все пароли по умолчанию
- [ ] Настроен SSL/TLS
- [ ] Настроен мониторинг (Prometheus, Grafana)
- [ ] Настроены регулярные бэкапы
- [ ] Настроены алерты
- [ ] Настроен log rotation
- [ ] Проверены limits ресурсов
- [ ] Настроен firewall
- [ ] Используются secrets вместо environment переменных
- [ ] Настроен reverse proxy (если нужно)

## 📚 Дополнительная информация

- [Docker Compose Documentation](https://docs.docker.com/compose/)
- [Nginx Documentation](https://nginx.org/en/docs/)
- [PostgreSQL Docker](https://hub.docker.com/_/postgres)
- [Kafka Docker](https://hub.docker.com/r/confluentinc/cp-kafka)
- [MinIO Documentation](https://min.io/docs/minio/linux/index.html)

## 🆘 Поддержка

При возникновении проблем:

1. Проверьте логи: `docker-compose logs`
2. Проверьте документацию выше
3. Создайте Issue в репозитории
4. Свяжитесь с командой разработки

---

**Приятной работы с ImageProcessor в Docker!** 🐳

