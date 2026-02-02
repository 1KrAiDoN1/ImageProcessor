# API Examples - Image Processor

Примеры использования API для различных сценариев обработки изображений.

## Содержание
- [Базовые операции](#базовые-операции)
- [Комбинированная обработка](#комбинированная-обработка)
- [Получение результатов](#получение-результатов)
- [Мониторинг и статистика](#мониторинг-и-статистика)

## Базовые операции

### 1. Загрузка изображения с созданием миниатюры

```bash
curl -X POST http://localhost:8080/api/v1/images \
  -F "image=@photo.jpg" \
  -F 'operations=[
    {
      "type": "thumbnail",
      "parameters": {
        "size": 200,
        "crop_to_fit": true
      }
    }
  ]'
```

**Ответ:**
```json
{
  "id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
  "status": "processing",
  "filename": "photo.jpg",
  "size": 2048576,
  "mime_type": "image/jpeg",
  "created_at": "2026-02-02T10:00:00Z",
  "operations_count": 1,
  "estimated_time": 2
}
```

### 2. Изменение размера с сохранением пропорций

```bash
curl -X POST http://localhost:8080/api/v1/images \
  -F "image=@landscape.jpg" \
  -F 'operations=[
    {
      "type": "resize",
      "parameters": {
        "width": 1920,
        "height": 1080,
        "keep_aspect": true
      }
    }
  ]'
```

### 3. Добавление водяного знака

```bash
curl -X POST http://localhost:8080/api/v1/images \
  -F "image=@product.jpg" \
  -F 'operations=[
    {
      "type": "watermark",
      "parameters": {
        "text": "© My Company 2026",
        "opacity": 0.7,
        "position": "bottom-right",
        "font_size": 28
      }
    }
  ]'
```

## Комбинированная обработка

### 4. Создание нескольких версий изображения

```bash
curl -X POST http://localhost:8080/api/v1/images \
  -F "image=@original.jpg" \
  -F 'operations=[
    {
      "type": "thumbnail",
      "parameters": {
        "size": 150,
        "crop_to_fit": true
      }
    },
    {
      "type": "resize",
      "parameters": {
        "width": 800,
        "height": 600,
        "keep_aspect": true
      }
    },
    {
      "type": "watermark",
      "parameters": {
        "text": "Preview",
        "opacity": 0.5,
        "position": "center"
      }
    }
  ]'
```

### 5. Обработка для веб-галереи

```bash
# Создание трех версий: миниатюра, средняя, полноразмерная с водяным знаком
curl -X POST http://localhost:8080/api/v1/images \
  -F "image=@gallery_image.jpg" \
  -F 'operations=[
    {
      "type": "thumbnail",
      "parameters": {
        "size": 200,
        "crop_to_fit": true
      }
    },
    {
      "type": "resize",
      "parameters": {
        "width": 1024,
        "keep_aspect": true
      }
    }
  ]'
```

## Получение результатов

### 6. Проверка статуса обработки

```bash
# Получить статус
IMAGE_ID="a1b2c3d4-e5f6-7890-abcd-ef1234567890"
curl http://localhost:8080/api/v1/images/$IMAGE_ID/status
```

**Ответ:**
```json
{
  "id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
  "status": "completed",
  "progress": 100,
  "processed_operations": 3,
  "total_operations": 3,
  "created_at": "2026-02-02T10:00:00Z",
  "updated_at": "2026-02-02T10:00:05Z"
}
```

### 7. Получение оригинального изображения

```bash
IMAGE_ID="a1b2c3d4-e5f6-7890-abcd-ef1234567890"
curl http://localhost:8080/api/v1/images/$IMAGE_ID?operation=original \
  --output original.jpg
```

### 8. Получение обработанных версий

```bash
# Миниатюра
curl http://localhost:8080/api/v1/images/$IMAGE_ID?operation=thumbnail \
  --output thumbnail.jpg

# Измененный размер
curl http://localhost:8080/api/v1/images/$IMAGE_ID?operation=resize \
  --output resized.jpg

# С водяным знаком
curl http://localhost:8080/api/v1/images/$IMAGE_ID?operation=watermark \
  --output watermarked.jpg
```

### 9. Получение временной ссылки (Presigned URL)

```bash
# Ссылка действительна 1 час (3600 секунд)
curl http://localhost:8080/api/v1/images/$IMAGE_ID/url?operation=thumbnail&expiry=3600
```

**Ответ:**
```json
{
  "url": "https://localhost:9000/images/processed/a1b2c3d4.../thumbnail.jpg?X-Amz-Algorithm=...",
  "expires_in": 3600
}
```

## Мониторинг и статистика

### 10. Общая статистика системы

```bash
curl http://localhost:8080/api/v1/statistics
```

**Ответ:**
```json
{
  "total_images_uploaded": 1250,
  "total_images_processed": 1200,
  "total_images_failed": 50,
  "total_data_processed_bytes": 5368709120,
  "total_data_processed_mb": 5120.0,
  "average_processing_time_ms": 1850.5,
  "operation_statistics": [
    {
      "operation_type": "thumbnail",
      "total_count": 800,
      "success_count": 790,
      "failure_count": 10,
      "average_processing_time_ms": 650.2
    },
    {
      "operation_type": "resize",
      "total_count": 600,
      "success_count": 585,
      "failure_count": 15,
      "average_processing_time_ms": 1200.8
    },
    {
      "operation_type": "watermark",
      "total_count": 450,
      "success_count": 440,
      "failure_count": 10,
      "average_processing_time_ms": 2100.5
    }
  ],
  "last_updated": "2026-02-02T12:30:00Z"
}
```

### 11. Список загруженных изображений

```bash
# С пагинацией
curl "http://localhost:8080/api/v1/images?limit=20&offset=0"
```

**Ответ:**
```json
{
  "images": [
    {
      "id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
      "filename": "photo1.jpg",
      "status": "completed",
      "size": 2048576,
      "mime_type": "image/jpeg",
      "created_at": "2026-02-02T10:00:00Z",
      "updated_at": "2026-02-02T10:00:05Z"
    }
  ],
  "limit": 20,
  "offset": 0,
  "count": 1
}
```

### 12. Health Check

```bash
curl http://localhost:8080/health
```

**Ответ:**
```json
{
  "status": "healthy",
  "timestamp": "2026-02-02T12:30:00Z",
  "services": {
    "api": "ok",
    "database": "ok",
    "storage": "ok"
  }
}
```

## Практические сценарии

### Сценарий 1: Загрузка фото профиля

```bash
#!/bin/bash

# 1. Загружаем фото с созданием аватарки
RESPONSE=$(curl -s -X POST http://localhost:8080/api/v1/images \
  -F "image=@profile.jpg" \
  -F 'operations=[
    {
      "type": "thumbnail",
      "parameters": {
        "size": 200,
        "crop_to_fit": true
      }
    }
  ]')

# 2. Извлекаем ID
IMAGE_ID=$(echo $RESPONSE | jq -r '.id')
echo "Image ID: $IMAGE_ID"

# 3. Ждем обработки
while true; do
  STATUS=$(curl -s http://localhost:8080/api/v1/images/$IMAGE_ID/status | jq -r '.status')
  echo "Status: $STATUS"
  
  if [ "$STATUS" = "completed" ]; then
    break
  fi
  
  sleep 1
done

# 4. Скачиваем аватарку
curl http://localhost:8080/api/v1/images/$IMAGE_ID?operation=thumbnail \
  --output avatar.jpg

echo "Avatar saved as avatar.jpg"
```

### Сценарий 2: Обработка товарных фотографий

```bash
#!/bin/bash

for file in products/*.jpg; do
  echo "Processing: $file"
  
  curl -X POST http://localhost:8080/api/v1/images \
    -F "image=@$file" \
    -F 'operations=[
      {
        "type": "thumbnail",
        "parameters": {
          "size": 300,
          "crop_to_fit": true
        }
      },
      {
        "type": "resize",
        "parameters": {
          "width": 1200,
          "keep_aspect": true
        }
      },
      {
        "type": "watermark",
        "parameters": {
          "text": "© Shop 2026",
          "opacity": 0.3,
          "position": "bottom-right"
        }
      }
    ]'
  
  sleep 0.5
done
```

### Сценарий 3: Мониторинг производительности

```bash
#!/bin/bash

# Периодический запрос статистики
while true; do
  clear
  echo "=== Image Processor Statistics ==="
  echo ""
  
  curl -s http://localhost:8080/api/v1/statistics | jq '{
    uploaded: .total_images_uploaded,
    processed: .total_images_processed,
    failed: .total_images_failed,
    avg_time_ms: .average_processing_time_ms,
    data_processed_mb: .total_data_processed_mb
  }'
  
  sleep 10
done
```

## Удаление изображений

### 13. Удаление одного изображения

```bash
IMAGE_ID="a1b2c3d4-e5f6-7890-abcd-ef1234567890"
curl -X DELETE http://localhost:8080/api/v1/images/$IMAGE_ID
```

**Ответ:**
```json
{
  "success": true,
  "message": "Image deleted successfully",
  "id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890"
}
```

## Обработка ошибок

### Примеры ошибок

**Файл слишком большой:**
```json
{
  "error": "invalid_request",
  "message": "file size exceeds maximum allowed size of 33554432 bytes"
}
```

**Неподдерживаемый формат:**
```json
{
  "error": "invalid_file",
  "message": "invalid content type: image/bmp. Supported types: image/jpeg, image/png, image/gif, image/webp"
}
```

**Изображение не найдено:**
```json
{
  "error": "not_found",
  "message": "Image not found: image not found: uuid"
}
```

**Невалидные параметры операции:**
```json
{
  "error": "invalid_operation",
  "message": "Invalid operation at index 0: width must be between 1 and 4096"
}
```

## Советы по использованию

1. **Всегда проверяйте статус** перед получением обработанных изображений
2. **Используйте presigned URLs** для прямого доступа к S3 в production
3. **Комбинируйте операции** в одном запросе для эффективности
4. **Мониторьте статистику** для оптимизации производительности
5. **Удаляйте неиспользуемые изображения** для экономии места

## Ограничения

- Максимальный размер файла: **32 MB**
- Поддерживаемые форматы: **JPEG, PNG, GIF, WebP**
- Максимальный размер изображения: **4096x4096 px**
- Максимальный размер миниатюры: **1000px**

