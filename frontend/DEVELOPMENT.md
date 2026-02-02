# Frontend Development Guide

Руководство по разработке и расширению фронтенда ImageProcessor.

## 🏗 Архитектура

Фронтенд построен на модульной архитектуре с разделением ответственности:

### Модули

1. **config.js** - Централизованная конфигурация
2. **api.js** - API клиент для взаимодействия с backend
3. **ui.js** - UI утилиты и хелперы
4. **app.js** - Основная логика приложения и обработчики событий

### Структура данных

```javascript
// Операция обработки
{
    type: 'resize' | 'thumbnail' | 'watermark',
    parameters: {
        // Параметры зависят от типа операции
    }
}

// Статус изображения
{
    id: 'uuid',
    status: 'pending' | 'processing' | 'completed' | 'failed',
    progress: 0-100,
    processed_operations: number,
    total_operations: number
}
```

## 🔧 Настройка окружения

### Требования

- Современный браузер (Chrome 90+, Firefox 88+, Safari 14+)
- HTTP сервер для разработки
- Доступ к backend API

### Установка

```bash
cd frontend

# С помощью npm
npm install
npm run dev

# Или с помощью Python
python3 -m http.server 8000

# Или с помощью Node.js
npx http-server -p 8000 -c-1
```

### Конфигурация

Измените настройки в `js/config.js`:

```javascript
const CONFIG = {
    api: {
        baseURL: 'http://your-api-url/api/v1'
    }
};
```

## 📝 Соглашения по коду

### JavaScript

- Используйте ES6+ синтаксис
- Async/await для асинхронных операций
- Константы в UPPER_CASE
- Функции и переменные в camelCase
- Классы в PascalCase

```javascript
// ✅ Хорошо
const API_TIMEOUT = 30000;
async function fetchData() { ... }
class ImageProcessor { ... }

// ❌ Плохо
var api_timeout = 30000;
function FetchData() { ... }
class imageProcessor { ... }
```

### CSS

- Используйте CSS переменные для цветов
- БЭМ методология для именования классов (опционально)
- Mobile-first подход

```css
/* ✅ Хорошо */
.card {
    background: var(--white);
}

.card__title {
    color: var(--dark);
}

/* ❌ Плохо */
.card {
    background: #ffffff;
}

.cardTitle {
    color: #2d3748;
}
```

### HTML

- Семантические теги
- Доступность (ARIA атрибуты)
- SEO оптимизация

```html
<!-- ✅ Хорошо -->
<section aria-label="Upload section">
    <h1>Upload Image</h1>
    <button aria-label="Select file">Choose</button>
</section>

<!-- ❌ Плохо -->
<div>
    <div>Upload Image</div>
    <div onclick="selectFile()">Choose</div>
</div>
```

## 🎨 Кастомизация дизайна

### Изменение цветов

Откройте `css/style.css` и измените CSS переменные:

```css
:root {
    --primary: #667eea;      /* Основной цвет */
    --secondary: #764ba2;    /* Вторичный цвет */
    --success: #43e97b;      /* Успех */
    --danger: #f5576c;       /* Ошибка */
    --warning: #ffa726;      /* Предупреждение */
    --info: #4facfe;         /* Информация */
}
```

### Изменение шрифтов

```css
body {
    font-family: 'Your Font', sans-serif;
}
```

### Изменение анимаций

```css
:root {
    --transition: all 0.3s ease;
}

.element {
    transition: var(--transition);
}
```

## 🔌 Добавление новых операций

### 1. Добавить в HTML (index.html)

```html
<div class="operation-card">
    <input type="checkbox" id="opCrop" class="operation-checkbox">
    <label for="opCrop" class="operation-label">
        <div class="operation-icon">
            <i class="fas fa-crop"></i>
        </div>
        <h4>Обрезка</h4>
        <p>Обрезать изображение</p>
    </label>
    <div class="operation-params hidden" id="cropParams">
        <!-- Параметры операции -->
    </div>
</div>
```

### 2. Добавить обработчик в app.js

```javascript
function initializeOperations() {
    // ...
    
    const cropCheckbox = document.getElementById('opCrop');
    cropCheckbox.addEventListener('change', () => {
        UI.toggleElement(
            document.getElementById('cropParams'),
            cropCheckbox.checked
        );
    });
}

function getSelectedOperations() {
    // ...
    
    if (document.getElementById('opCrop').checked) {
        operations.push({
            type: 'crop',
            parameters: {
                x: parseInt(document.getElementById('cropX').value),
                y: parseInt(document.getElementById('cropY').value),
                width: parseInt(document.getElementById('cropWidth').value),
                height: parseInt(document.getElementById('cropHeight').value)
            }
        });
    }
    
    return operations;
}
```

### 3. Добавить локализацию в ui.js

```javascript
getOperationName(type) {
    const names = {
        // ...
        'crop': 'Обрезка'
    };
    return names[type] || type;
}
```

## 📡 Работа с API

### Добавление нового endpoint

```javascript
// В api.js
class ImageProcessorAPI {
    // ...
    
    async customMethod(params) {
        try {
            const response = await fetch(`${this.baseURL}/custom`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify(params)
            });
            
            if (!response.ok) {
                throw new Error('Request failed');
            }
            
            return await response.json();
        } catch (error) {
            console.error('Error:', error);
            throw error;
        }
    }
}
```

### Обработка ошибок

```javascript
try {
    const result = await api.uploadImage(file, operations);
    UI.showToast('Успешно загружено', 'success');
} catch (error) {
    if (error.message.includes('Network')) {
        UI.showToast('Нет подключения к серверу', 'error');
    } else if (error.message.includes('timeout')) {
        UI.showToast('Превышено время ожидания', 'error');
    } else {
        UI.showToast(`Ошибка: ${error.message}`, 'error');
    }
    console.error('Upload error:', error);
}
```

## 🧪 Тестирование

### Ручное тестирование

1. Откройте DevTools (F12)
2. Вкладка Console для логов
3. Вкладка Network для API запросов
4. Вкладка Application для хранилища

### Проверка производительности

```javascript
// Измерение времени выполнения
console.time('upload');
await api.uploadImage(file, operations);
console.timeEnd('upload');

// Проверка памяти
console.log(performance.memory);
```

### Проверка доступности

1. Используйте Lighthouse в Chrome DevTools
2. Проверьте навигацию с клавиатуры (Tab, Enter, Esc)
3. Проверьте с screen reader

## 🐛 Отладка

### Включение debug режима

```javascript
// В config.js
const CONFIG = {
    debug: {
        enabled: true,
        logLevel: 'debug'
    }
};
```

### Логирование

```javascript
// В app.js или других модулях
if (CONFIG.debug.enabled) {
    console.log('[DEBUG]', 'Message', data);
}
```

### Общие проблемы

#### CORS ошибка

```
Access to fetch at 'http://localhost:8080/api/v1/images' 
from origin 'http://localhost:8000' has been blocked by CORS policy
```

**Решение:** Настройте CORS на backend:

```go
// В backend
router.Use(cors.New(cors.Config{
    AllowOrigins:     []string{"http://localhost:8000"},
    AllowMethods:     []string{"GET", "POST", "PUT", "DELETE"},
    AllowHeaders:     []string{"Content-Type"},
    AllowCredentials: true,
}))
```

#### API не отвечает

**Проверка:**
1. Backend запущен: `curl http://localhost:8080/health`
2. URL правильный в `config.js`
3. Нет ошибок в консоли backend

## 📦 Сборка для production

### Минификация CSS

```bash
# Установить cssnano
npm install -g cssnano-cli

# Минифицировать
cssnano css/style.css css/style.min.css
```

### Минификация JS

```bash
# Установить terser
npm install -g terser

# Минифицировать
terser js/config.js js/api.js js/ui.js js/app.js \
    -o js/bundle.min.js \
    --compress \
    --mangle
```

### Оптимизация изображений

Используйте WebP формат для изображений:

```bash
# Конвертировать в WebP
cwebp -q 80 input.png -o output.webp
```

## 🚀 Деплой

### Статический хостинг

#### Netlify

```bash
# netlify.toml
[build]
  publish = "frontend"

[[redirects]]
  from = "/*"
  to = "/index.html"
  status = 200
```

#### Vercel

```json
{
  "rewrites": [
    { "source": "/(.*)", "destination": "/" }
  ]
}
```

#### GitHub Pages

```bash
# .github/workflows/deploy.yml
name: Deploy
on:
  push:
    branches: [main]
jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      - name: Deploy
        uses: peaceiris/actions-gh-pages@v3
        with:
          github_token: ${{ secrets.GITHUB_TOKEN }}
          publish_dir: ./frontend
```

### Nginx

```nginx
server {
    listen 80;
    server_name your-domain.com;
    
    root /var/www/imageprocessor/frontend;
    index index.html;
    
    location / {
        try_files $uri $uri/ /index.html;
    }
    
    location /api/ {
        proxy_pass http://backend:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}
```

## 📚 Полезные ресурсы

- [MDN Web Docs](https://developer.mozilla.org/)
- [Can I Use](https://caniuse.com/)
- [CSS-Tricks](https://css-tricks.com/)
- [JavaScript.info](https://javascript.info/)
- [Font Awesome Icons](https://fontawesome.com/icons)

## 🤝 Вклад в проект

1. Форкните репозиторий
2. Создайте feature ветку (`git checkout -b feature/amazing-feature`)
3. Закоммитьте изменения (`git commit -m 'Add amazing feature'`)
4. Запушьте в ветку (`git push origin feature/amazing-feature`)
5. Откройте Pull Request

## 📝 Checklist перед коммитом

- [ ] Код отформатирован
- [ ] Нет console.log в production коде
- [ ] Проверена работа во всех браузерах
- [ ] Проверена адаптивность
- [ ] Обновлена документация
- [ ] Добавлены комментарии к сложному коду

## 🎉 Готово!

Теперь вы готовы разрабатывать и расширять фронтенд ImageProcessor!

Если у вас есть вопросы, создайте Issue в репозитории.

