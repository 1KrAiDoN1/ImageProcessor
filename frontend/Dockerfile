# frontend/Dockerfile (самый простой)
FROM nginx:alpine

# Установка curl для health checks
RUN apk add --no-cache curl

# Копируем конфигурацию nginx
COPY nginx.conf /etc/nginx/conf.d/default.conf

# Копируем статические файлы (HTML, CSS, JS, изображения)
COPY . /usr/share/nginx/html/

# Устанавливаем корректные права доступа
RUN chown -R nginx:nginx /usr/share/nginx/html && \
    chmod -R 755 /usr/share/nginx/html

# Health check
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
  CMD curl -f http://localhost/health || exit 1

# Открываем порт 80
EXPOSE 80

# Запускаем nginx
CMD ["nginx", "-g", "daemon off;"]