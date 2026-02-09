# Image Processor
![Go Version](https://img.shields.io/badge/Go-1.21%2B-blue.svg)
![Framework](https://img.shields.io/badge/Framework-Gin-green.svg)
![Database](https://img.shields.io/badge/Database-PostgreSQL-blue.svg)
![Kafka](https://img.shields.io/badge/Apache-Kafka-231F20.svg)
![Docker](https://img.shields.io/badge/Docker-Supported-brightgreen.svg)

Высокопроизводительный асинхронный сервис обработки изображений с использованием микросервисной архитектуры.


## 🏗 Архитектура

Система состоит из двух основных сервисов:

### API Gateway
- Принимает изображения от пользователей
- Сохраняет оригиналы в S3 хранилище
- Записывает метаданные в PostgreSQL
- Публикует задачи на обработку в Kafka
- Предоставляет REST API для доступа к изображениям

### Worker Service
- Читает задачи из Kafka
- Обрабатывает изображения (resize, thumbnail, watermark и т.д.)
- Сохраняет результаты в S3
- Обновляет статус в PostgreSQL
- Собирает статистику обработки

## 🚀 Технологический стек

### Backend
- **Go 1.25** - основной язык программирования
- **PostgreSQL 15** - реляционная база данных
- **Apache Kafka** - брокер сообщений для асинхронной обработки
- **MinIO (S3)** - объектное хранилище для изображений
- **Gin** - HTTP фреймворк
- **Docker & Docker Compose** - контейнеризация

### Frontend
- **HTML5** - разметка
- **CSS3** - стили с градиентами и анимациями
- **Vanilla JavaScript** - без фреймворков, чистый ES6+
- **Font Awesome** - иконки

## 📦 Компоненты

### Обработка изображений
- **Resize** - изменение размера с сохранением пропорций
- **Thumbnail** - создание миниатюр
- **Watermark** - наложение водяных знаков
- **Crop** - обрезка изображения
- **Rotate** - поворот изображения

### Хранилище
- Оригиналы хранятся в S3: `originals/{imageId}/{filename}`
- Обработанные версии: `processed/{imageId}/{operation}/{uuid}.jpg`

### База данных
- **images** - информация об оригинальных изображениях
- **processed_images** - обработанные версии
- **processing_jobs** - задачи на обработку
- **statistics** - общая статистика
- **operation_statistics** - статистика по операциям

## 🛠 Установка и запуск

### Предварительные требования

- **Docker** и **Docker Compose** (обязательно)
- Go 1.25+ (для локальной разработки без Docker)
- Make (опционально, для удобства)

### ⚡ Быстрый старт (рекомендуется)

```bash
# 1. Клонировать репозиторий
git clone <repository-url>
cd ImageProcessor

# 2. Запустить все сервисы одной командой
make docker-up

# Или используйте docker-compose напрямую
docker-compose up -d

# 3. Примените миграции
make migrate-up
# Затем откройте браузер
```

**Сервисы доступны на:**
- 🌐 **Frontend** (Веб-интерфейс): http://localhost
- 🔌 **API**: http://localhost:8080
- 📦 **MinIO Console**: http://localhost:9001 (minioadmin/minioadmin)
- 🗄️ **PostgreSQL**: localhost:5432
- 📨 **Kafka**: localhost:9092

### 🎯 Что входит в сборку

Docker Compose автоматически запускает:
- ✅ **Frontend** - Nginx с веб-интерфейсом
- ✅ **API Gateway** - REST API сервис
- ✅ **Worker** (2 реплики) - Обработчики изображений
- ✅ **PostgreSQL** - База данных
- ✅ **Kafka + Zookeeper** - Брокер сообщений
- ✅ **MinIO** - S3-совместимое хранилище

### 🎨 Веб-интерфейс

Frontend автоматически запускается в Docker контейнере и доступен на **http://localhost**

**Возможности веб-интерфейса:**
- ✨ **Drag & Drop** загрузка изображений
- 🎛️ **Интерактивный выбор** операций с настройками параметров
- 📊 **Real-time мониторинг** прогресса обработки
- 🖼️ **Галерея** всех изображений с поиском и фильтрацией
- 📈 **Детальная статистика** по обработке и операциям
- 🎨 **Современный дизайн** с градиентами и анимациями
- 📱 **Адаптивная верстка** для desktop и mobile
- 🌐 **Проксирование API** через Nginx (без CORS проблем)


## 📡 API Endpoints

### Загрузка изображения

```bash
POST /api/v1/images
Content-Type: multipart/form-data

Parameters:
- image: файл изображения (обязательно)
- operations: JSON массив операций (опционально)

Example:
curl -X POST http://localhost:8080/api/v1/images \
  -F "image=@photo.jpg" \
  -F 'operations=[{"type":"thumbnail","parameters":{"size":200}},{"type":"watermark","parameters":{"text":"My Photo"}}]'

Response:
{
  "id": "uuid",
  "status": "processing",
  "filename": "photo.jpg",
  "size": 1024000,
  "mime_type": "image/jpeg",
  "created_at": "2026-02-02T10:00:00Z",
  "operations_count": 2,
  "estimated_time": 4
}
```

### Получение изображения

```bash
GET /api/v1/images/:id?operation=thumbnail

Example:
curl http://localhost:8080/api/v1/images/uuid?operation=thumbnail --output image.jpg
```

### Статус обработки

```bash
GET /api/v1/images/:id/status

Response:
{
  "id": "uuid",
  "status": "completed",
  "progress": 100,
  "processed_operations": 2,
  "total_operations": 2,
  "created_at": "2026-02-02T10:00:00Z",
  "updated_at": "2026-02-02T10:00:05Z"
}
```

### Удаление изображения

```bash
DELETE /api/v1/images/:id

Response:
{
  "success": true,
  "message": "Image deleted successfully",
  "id": "uuid"
}
```

### Статистика

```bash
GET /api/v1/statistics

Response:
{
  "total_images_uploaded": 1000,
  "total_images_processed": 950,
  "total_images_failed": 50,
  "total_data_processed_bytes": 1073741824,
  "total_data_processed_mb": 1024.0,
  "average_processing_time_ms": 1500.5,
  "operation_statistics": [
    {
      "operation_type": "thumbnail",
      "total_count": 500,
      "success_count": 495,
      "failure_count": 5,
      "average_processing_time_ms": 800.2
    }
  ],
  "last_updated": "2026-02-02T10:00:00Z"
}
```

### Presigned URL

```bash
GET /api/v1/images/:id/url?operation=original&expiry=3600

Response:
{
  "url": "https://minio:9000/images/...",
  "expires_in": 3600
}
```

## 🔧 Примеры операций

### Thumbnail

```json
{
  "type": "thumbnail",
  "parameters": {
    "size": 200,
    "crop_to_fit": true
  }
}
```

### Resize

```json
{
  "type": "resize",
  "parameters": {
    "width": 1024,
    "height": 768,
    "keep_aspect": true
  }
}
```

### Watermark

```json
{
  "type": "watermark",
  "parameters": {
    "text": "© My Company",
    "opacity": 0.5,
    "position": "bottom-right",
    "font_size": 24
  }
}
```

## 🚦 Производительность

- Асинхронная обработка через Kafka
- Масштабируемые воркеры
- Кэширование результатов
- Пакетная обработка
- Connection pooling для БД

## 📁 Структура проекта

```
ImageProcessor/
├── backend/
│   ├── cmd/
│   │   ├── api/              # API Gateway service
│   │   └── worker/           # Worker service
│   ├── internal/
│   │   ├── app/              # Application initialization
│   │   ├── broker/           # Kafka integration
│   │   ├── config/           # Configuration management
│   │   ├── domain/           # Domain entities
│   │   ├── http-server/      # HTTP handlers and routes
│   │   ├── repository/       # Data access layer
│   │   └── service/          # Business logic
│   └── migrations/           # Database migrations
├── frontend/
│   ├── index.html            # Main page
│   ├── css/
│   │   └── style.css         # Application styles
│   ├── js/
│   │   ├── api.js            # API client
│   │   ├── ui.js             # UI utilities
│   │   └── app.js            # Main application logic
│   ├── package.json          # Frontend dependencies
│   └── README.md             # Frontend documentation
├── docker-compose.yaml       # Docker services configuration
├── Dockerfile               # Application container
├── Makefile                # Build and run commands
├── API_EXAMPLES.md         # Detailed API examples
└── README.md               # This file
```
