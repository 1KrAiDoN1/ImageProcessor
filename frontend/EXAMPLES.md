# Frontend Examples & Testing

Примеры использования и тестирования фронтенда ImageProcessor.

## 🧪 Тестирование API через DevTools Console

Откройте DevTools (F12) → Console и выполните следующие команды:

### Проверка здоровья API

```javascript
await api.healthCheck()
```

### Загрузка изображения программно

```javascript
// Получить файл из input
const fileInput = document.getElementById('fileInput');
const file = fileInput.files[0];

// Определить операции
const operations = [
    {
        type: 'thumbnail',
        parameters: { size: 200, crop_to_fit: true }
    },
    {
        type: 'watermark',
        parameters: { 
            text: '© Test', 
            opacity: 0.5, 
            position: 'bottom-right' 
        }
    }
];

// Загрузить
const result = await api.uploadImage(file, operations);
console.log('Upload result:', result);

// Сохранить ID
const imageId = result.id;
```

### Получение статуса

```javascript
const status = await api.getImageStatus(imageId);
console.log('Status:', status);
```

### Polling статуса

```javascript
await api.pollImageStatus(imageId, (status) => {
    console.log('Progress:', status.progress, '%');
    console.log('Status:', status.status);
}, 2000, 30);
```

### Получение списка изображений

```javascript
const images = await api.listImages(10, 0);
console.log('Images:', images);
```

### Получение статистики

```javascript
const stats = await api.getStatistics();
console.log('Statistics:', stats);
```

### Получение presigned URL

```javascript
const urlData = await api.getPresignedURL(imageId, 'original', 3600);
console.log('Presigned URL:', urlData.url);
```

### Удаление изображения

```javascript
const deleteResult = await api.deleteImage(imageId);
console.log('Deleted:', deleteResult);
```

## 🎨 UI Тестирование

### Показать toast уведомления

```javascript
UI.showToast('Success message', 'success');
UI.showToast('Error message', 'error');
UI.showToast('Info message', 'info');
```

### Переключить секцию

```javascript
UI.switchSection('upload');
UI.switchSection('gallery');
UI.switchSection('stats');
```

### Обновить прогресс

```javascript
UI.updateProgress(50, 'Processing...', 'processing');
UI.updateProgress(100, 'Completed!', 'completed');
```

### Открыть модальное окно

```javascript
UI.openImageModal(imageId, 'original');
UI.closeImageModal();
```

### Показать/скрыть элемент

```javascript
const element = document.getElementById('uploadArea');
UI.toggleElement(element, true);  // Показать
UI.toggleElement(element, false); // Скрыть
```

### Форматирование

```javascript
UI.formatFileSize(1024000);        // "1000.0 KB"
UI.formatFileSize(1048576);        // "1.0 MB"
UI.formatDate('2026-02-02T10:00:00Z'); // "фев. 2, 2026, 10:00"
```

## 📝 Пользовательские сценарии

### Сценарий 1: Загрузка с миниатюрой

1. Откройте страницу
2. Перетащите изображение в область загрузки
3. Выберите "Миниатюра"
4. Установите размер 300px
5. Нажмите "Обработать изображение"
6. Дождитесь завершения
7. Просмотрите результаты

### Сценарий 2: Множественные операции

1. Загрузите изображение
2. Выберите все операции:
   - Миниатюра (200px)
   - Изменить размер (1920x1080)
   - Водяной знак ("© My Photo")
3. Настройте параметры
4. Загрузите
5. Отслеживайте прогресс

### Сценарий 3: Работа с галереей

1. Перейдите в "Галерея"
2. Используйте поиск для фильтрации
3. Используйте фильтры статуса
4. Кликните на изображение для просмотра
5. Скачайте изображение
6. Удалите изображение

### Сценарий 4: Просмотр статистики

1. Перейдите в "Статистика"
2. Просмотрите общую статистику
3. Изучите таблицу по операциям
4. Нажмите "Обновить статистику"

## 🔧 Отладка

### Включить verbose логирование

```javascript
// Изменить в config.js
CONFIG.debug.enabled = true;
CONFIG.debug.logLevel = 'debug';

// Или временно в консоли
window.DEBUG = true;
```

### Перехватить все fetch запросы

```javascript
const originalFetch = window.fetch;
window.fetch = async (...args) => {
    console.log('Fetch:', args[0]);
    const response = await originalFetch(...args);
    console.log('Response:', response.status);
    return response;
};
```

### Логировать все события

```javascript
document.addEventListener('click', (e) => {
    console.log('Click:', e.target);
});

document.addEventListener('change', (e) => {
    console.log('Change:', e.target.id, e.target.value);
});
```

## 🚀 Производительность

### Измерить время загрузки страницы

```javascript
window.addEventListener('load', () => {
    const loadTime = performance.now();
    console.log(`Page loaded in ${loadTime.toFixed(2)}ms`);
});
```

### Измерить время API запроса

```javascript
console.time('API Call');
const result = await api.getStatistics();
console.timeEnd('API Call');
```

### Проверить использование памяти

```javascript
if (performance.memory) {
    console.log('Used JS Heap:', 
        (performance.memory.usedJSHeapSize / 1048576).toFixed(2), 'MB');
    console.log('Total JS Heap:', 
        (performance.memory.totalJSHeapSize / 1048576).toFixed(2), 'MB');
}
```

## 🎯 Тестирование доступности

### Навигация с клавиатуры

1. **Tab** - переход между элементами
2. **Enter/Space** - активация кнопок
3. **Esc** - закрытие модальных окон
4. **Arrow keys** - навигация в списках

### Screen Reader тестирование

Используйте:
- **macOS**: VoiceOver (Cmd + F5)
- **Windows**: NVDA или JAWS
- **Linux**: Orca

### Lighthouse аудит

1. Откройте DevTools
2. Вкладка "Lighthouse"
3. Выберите категории
4. Нажмите "Generate report"

## 📱 Тестирование на устройствах

### Desktop

- Chrome (последние 2 версии)
- Firefox (последние 2 версии)
- Safari (последняя версия)
- Edge (последняя версия)

### Mobile

- iOS Safari (iOS 14+)
- Chrome Mobile (Android 10+)
- Samsung Internet

### Responsive Design Mode

1. Откройте DevTools
2. Нажмите Ctrl+Shift+M (Cmd+Shift+M на Mac)
3. Выберите устройство
4. Проверьте функциональность

## 🐛 Общие проблемы и решения

### Изображение не загружается

**Проблема:** Файл выбран, но ничего не происходит

**Решение:**
```javascript
// Проверить в консоли
console.log('Selected file:', selectedFile);
console.log('File type:', selectedFile?.type);
console.log('File size:', selectedFile?.size);
```

### API не отвечает

**Проблема:** Запросы не доходят до backend

**Решение:**
```javascript
// Проверить URL
console.log('API URL:', api.baseURL);

// Проверить здоровье
const health = await api.healthCheck();
console.log('API healthy:', health);

// Проверить CORS
// Откройте Network tab и проверьте заголовки
```

### Toast не показывается

**Проблема:** Уведомления не появляются

**Решение:**
```javascript
// Проверить контейнер
const container = document.getElementById('toastContainer');
console.log('Toast container exists:', !!container);

// Проверить z-index
console.log('Toast z-index:', 
    window.getComputedStyle(container).zIndex);
```

### Modal не закрывается

**Проблема:** Модальное окно остается открытым

**Решение:**
```javascript
// Закрыть принудительно
const modal = document.getElementById('imageModal');
modal.classList.add('hidden');

// Проверить обработчики
console.log('Modal close handler:', 
    document.getElementById('modalClose').onclick);
```

## 💡 Советы и трюки

### Быстрая загрузка тестовых данных

```javascript
// Создать случайное изображение
async function createTestImage() {
    const canvas = document.createElement('canvas');
    canvas.width = 800;
    canvas.height = 600;
    const ctx = canvas.getContext('2d');
    
    // Случайный градиент
    const gradient = ctx.createLinearGradient(0, 0, 800, 600);
    gradient.addColorStop(0, '#667eea');
    gradient.addColorStop(1, '#764ba2');
    ctx.fillStyle = gradient;
    ctx.fillRect(0, 0, 800, 600);
    
    // Текст
    ctx.fillStyle = 'white';
    ctx.font = '48px Arial';
    ctx.textAlign = 'center';
    ctx.fillText('Test Image', 400, 300);
    
    // Конвертировать в Blob
    return new Promise(resolve => {
        canvas.toBlob(resolve, 'image/jpeg', 0.9);
    });
}

// Использование
const blob = await createTestImage();
const file = new File([blob], 'test.jpg', { type: 'image/jpeg' });
await api.uploadImage(file, [
    { type: 'thumbnail', parameters: { size: 200 } }
]);
```

### Массовая загрузка

```javascript
async function bulkUpload(count = 10) {
    const operations = [
        { type: 'thumbnail', parameters: { size: 200 } }
    ];
    
    for (let i = 0; i < count; i++) {
        const blob = await createTestImage();
        const file = new File([blob], `test-${i}.jpg`, { 
            type: 'image/jpeg' 
        });
        
        try {
            const result = await api.uploadImage(file, operations);
            console.log(`Uploaded ${i + 1}/${count}:`, result.id);
        } catch (error) {
            console.error(`Failed ${i + 1}/${count}:`, error);
        }
        
        // Задержка между загрузками
        await new Promise(r => setTimeout(r, 1000));
    }
}

// Загрузить 10 тестовых изображений
await bulkUpload(10);
```

### Экспорт данных

```javascript
async function exportStatistics() {
    const stats = await api.getStatistics();
    const json = JSON.stringify(stats, null, 2);
    
    // Создать и скачать файл
    const blob = new Blob([json], { type: 'application/json' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = `statistics-${Date.now()}.json`;
    a.click();
    URL.revokeObjectURL(url);
}

await exportStatistics();
```

## 📊 Мониторинг

### Отслеживание ошибок

```javascript
window.addEventListener('error', (event) => {
    console.error('Global error:', event.error);
    
    // Отправить в систему мониторинга
    // logToMonitoring(event.error);
});

window.addEventListener('unhandledrejection', (event) => {
    console.error('Unhandled promise rejection:', event.reason);
});
```

### Отслеживание производительности

```javascript
// Отслеживать долгие запросы
const originalFetch = window.fetch;
window.fetch = async (...args) => {
    const start = performance.now();
    const response = await originalFetch(...args);
    const duration = performance.now() - start;
    
    if (duration > 3000) {
        console.warn('Slow request:', args[0], `${duration.toFixed(0)}ms`);
    }
    
    return response;
};
```

## 🎉 Готово!

Теперь вы знаете, как тестировать и отлаживать фронтенд ImageProcessor!

Если вы нашли баг или у вас есть предложение по улучшению, создайте Issue в репозитории.

