# Deployment Guide

Руководство по развертыванию ImageProcessor в различных окружениях.

## 📋 Содержание

- [Development](#development)
- [Staging](#staging)
- [Production](#production)
- [Cloud Providers](#cloud-providers)
- [Мониторинг и логирование](#мониторинг-и-логирование)

## 🔧 Development

### Локальная разработка с Docker

```bash
# Клонировать репозиторий
git clone <repository-url>
cd ImageProcessor

# Запустить все сервисы
make docker-up-all

# Или используйте docker-compose напрямую
docker-compose up -d
```

**Доступ к сервисам:**
- Frontend: http://localhost
- API: http://localhost:8080
- MinIO Console: http://localhost:9001 (minioadmin/minioadmin)

### Локальная разработка без Docker

```bash
# Запустить только инфраструктуру в Docker
make docker-up-infra

# В отдельных терминалах запустить сервисы
make run-api
make run-worker

# Запустить фронтенд
make frontend
```

## 🧪 Staging

### Подготовка

1. **Создать .env файл**:
```bash
cp .env.example .env.staging
```

2. **Обновить переменные**:
```env
POSTGRES_PASSWORD=<staging-password>
MINIO_ROOT_PASSWORD=<staging-password>
API_HOST=staging-api.yourdomain.com
```

3. **Собрать образы**:
```bash
docker-compose -f docker-compose.staging.yml build
```

### Развертывание

```bash
# Запустить на staging сервере
docker-compose -f docker-compose.staging.yml up -d

# Проверить логи
docker-compose logs -f
```

### docker-compose.staging.yml

```yaml
version: '3.8'

services:
  frontend:
    build:
      context: ./frontend
      dockerfile: Dockerfile
    ports:
      - "80:80"
    environment:
      - NODE_ENV=staging
    restart: unless-stopped

  api:
    build:
      context: .
      dockerfile: Dockerfile
      target: api
    ports:
      - "8080:8080"
    env_file:
      - .env.staging
    restart: unless-stopped

  # ... остальные сервисы
```

## 🚀 Production

### Требования

- **Docker** и **Docker Compose** установлены
- **SSL сертификаты** (Let's Encrypt или другие)
- **Домен** настроен на ваш сервер
- **Firewall** настроен
- **Backup** система настроена

### Подготовка

1. **Создать production конфигурацию**:

```yaml
# docker-compose.prod.yml
version: '3.8'

services:
  nginx:
    image: nginx:alpine
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./nginx/nginx.conf:/etc/nginx/nginx.conf:ro
      - ./nginx/ssl:/etc/nginx/ssl:ro
      - ./nginx/sites:/etc/nginx/sites-enabled:ro
    depends_on:
      - frontend
      - api
    restart: always

  frontend:
    build:
      context: ./frontend
      dockerfile: Dockerfile
    expose:
      - "80"
    restart: always
    healthcheck:
      test: ["CMD", "wget", "--quiet", "--tries=1", "--spider", "http://localhost/"]
      interval: 30s
      timeout: 3s
      retries: 3

  api:
    build:
      context: .
      dockerfile: Dockerfile
      target: api
    expose:
      - "8080"
    env_file:
      - .env.production
    restart: always
    healthcheck:
      test: ["CMD", "wget", "--quiet", "--tries=1", "--spider", "http://localhost:8080/health"]
      interval: 30s
      timeout: 3s
      retries: 3
    deploy:
      resources:
        limits:
          cpus: '2'
          memory: 2G
        reservations:
          cpus: '1'
          memory: 1G

  worker:
    build:
      context: .
      dockerfile: Dockerfile
      target: worker
    env_file:
      - .env.production
    restart: always
    deploy:
      replicas: 3
      resources:
        limits:
          cpus: '2'
          memory: 2G

  postgres:
    image: postgres:15-alpine
    environment:
      POSTGRES_PASSWORD_FILE: /run/secrets/postgres_password
    volumes:
      - postgres_data:/var/lib/postgresql/data
      - ./backups:/backups
    secrets:
      - postgres_password
    restart: always

  minio:
    image: minio/minio:latest
    command: server /data --console-address ":9001"
    environment:
      MINIO_ROOT_USER_FILE: /run/secrets/minio_user
      MINIO_ROOT_PASSWORD_FILE: /run/secrets/minio_password
    volumes:
      - minio_data:/data
    secrets:
      - minio_user
      - minio_password
    restart: always

  kafka:
    image: confluentinc/cp-kafka:7.5.0
    # ... конфигурация
    restart: always

  zookeeper:
    image: confluentinc/cp-zookeeper:7.5.0
    # ... конфигурация
    restart: always

volumes:
  postgres_data:
    driver: local
  minio_data:
    driver: local

secrets:
  postgres_password:
    file: ./secrets/postgres_password.txt
  minio_user:
    file: ./secrets/minio_user.txt
  minio_password:
    file: ./secrets/minio_password.txt

networks:
  default:
    driver: bridge
```

2. **Создать secrets**:

```bash
mkdir -p secrets
echo "strong-postgres-password" > secrets/postgres_password.txt
echo "minio-admin" > secrets/minio_user.txt
echo "strong-minio-password" > secrets/minio_password.txt
chmod 600 secrets/*
```

3. **Настроить Nginx для SSL**:

```nginx
# nginx/sites/imageprocessor.conf
server {
    listen 80;
    server_name yourdomain.com;
    return 301 https://$server_name$request_uri;
}

server {
    listen 443 ssl http2;
    server_name yourdomain.com;

    ssl_certificate /etc/nginx/ssl/fullchain.pem;
    ssl_certificate_key /etc/nginx/ssl/privkey.pem;
    
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers HIGH:!aNULL:!MD5;
    ssl_prefer_server_ciphers on;

    # Security headers
    add_header Strict-Transport-Security "max-age=31536000; includeSubDomains" always;
    add_header X-Frame-Options "SAMEORIGIN" always;
    add_header X-Content-Type-Options "nosniff" always;
    add_header X-XSS-Protection "1; mode=block" always;

    location / {
        proxy_pass http://frontend:80;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    location /api/ {
        proxy_pass http://api:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        
        # Увеличенные таймауты для обработки изображений
        proxy_connect_timeout 300s;
        proxy_send_timeout 300s;
        proxy_read_timeout 300s;
    }
}
```

### Развертывание

```bash
# 1. Получить SSL сертификаты (Let's Encrypt)
certbot certonly --standalone -d yourdomain.com

# 2. Скопировать сертификаты
cp /etc/letsencrypt/live/yourdomain.com/fullchain.pem nginx/ssl/
cp /etc/letsencrypt/live/yourdomain.com/privkey.pem nginx/ssl/

# 3. Собрать образы
docker-compose -f docker-compose.prod.yml build

# 4. Запустить сервисы
docker-compose -f docker-compose.prod.yml up -d

# 5. Проверить статус
docker-compose -f docker-compose.prod.yml ps

# 6. Проверить логи
docker-compose -f docker-compose.prod.yml logs -f
```

### Обновление в Production

```bash
# 1. Создать бэкап
./scripts/backup.sh

# 2. Получить новый код
git pull origin main

# 3. Пересобрать образы
docker-compose -f docker-compose.prod.yml build

# 4. Обновить сервисы (zero-downtime)
docker-compose -f docker-compose.prod.yml up -d --no-deps --build api
docker-compose -f docker-compose.prod.yml up -d --no-deps --build worker
docker-compose -f docker-compose.prod.yml up -d --no-deps --build frontend

# 5. Проверить здоровье
curl https://yourdomain.com/health
```

## ☁️ Cloud Providers

### AWS (Amazon Web Services)

#### Using ECS (Elastic Container Service)

1. **Создать ECR репозитории**:
```bash
aws ecr create-repository --repository-name imageprocessor-frontend
aws ecr create-repository --repository-name imageprocessor-api
aws ecr create-repository --repository-name imageprocessor-worker
```

2. **Собрать и загрузить образы**:
```bash
# Login to ECR
aws ecr get-login-password --region us-east-1 | \
  docker login --username AWS --password-stdin <account-id>.dkr.ecr.us-east-1.amazonaws.com

# Build and push
docker build -t imageprocessor-frontend ./frontend
docker tag imageprocessor-frontend:latest <account-id>.dkr.ecr.us-east-1.amazonaws.com/imageprocessor-frontend:latest
docker push <account-id>.dkr.ecr.us-east-1.amazonaws.com/imageprocessor-frontend:latest
```

3. **Создать ECS Task Definitions и Services**

4. **Настроить RDS (PostgreSQL), MSK (Kafka), S3**

### Google Cloud Platform (GCP)

#### Using Cloud Run

```bash
# Build and push to GCR
gcloud builds submit --tag gcr.io/PROJECT-ID/imageprocessor-frontend ./frontend
gcloud builds submit --tag gcr.io/PROJECT-ID/imageprocessor-api .

# Deploy to Cloud Run
gcloud run deploy imageprocessor-frontend \
  --image gcr.io/PROJECT-ID/imageprocessor-frontend \
  --platform managed \
  --region us-central1 \
  --allow-unauthenticated

gcloud run deploy imageprocessor-api \
  --image gcr.io/PROJECT-ID/imageprocessor-api \
  --platform managed \
  --region us-central1
```

### Azure

#### Using Azure Container Instances

```bash
# Create resource group
az group create --name imageprocessor-rg --location eastus

# Create container registry
az acr create --resource-group imageprocessor-rg \
  --name imageprocessorregistry --sku Basic

# Build and push
az acr build --registry imageprocessorregistry \
  --image imageprocessor-frontend:latest ./frontend

# Deploy
az container create --resource-group imageprocessor-rg \
  --name imageprocessor-frontend \
  --image imageprocessorregistry.azurecr.io/imageprocessor-frontend:latest \
  --dns-name-label imageprocessor \
  --ports 80
```

### DigitalOcean

#### Using App Platform

```yaml
# .do/app.yaml
name: imageprocessor
services:
- name: frontend
  dockerfile_path: frontend/Dockerfile
  github:
    repo: your-username/imageprocessor
    branch: main
  http_port: 80
  routes:
  - path: /
  
- name: api
  dockerfile_path: Dockerfile
  github:
    repo: your-username/imageprocessor
    branch: main
  http_port: 8080
  routes:
  - path: /api

databases:
- name: postgres
  engine: PG
  version: "15"

- name: kafka
  # Managed Kafka не доступен, используйте external
```

## 📊 Мониторинг и логирование

### Prometheus + Grafana

```yaml
# docker-compose.monitoring.yml
version: '3.8'

services:
  prometheus:
    image: prom/prometheus:latest
    volumes:
      - ./prometheus/prometheus.yml:/etc/prometheus/prometheus.yml
      - prometheus_data:/prometheus
    ports:
      - "9090:9090"
    command:
      - '--config.file=/etc/prometheus/prometheus.yml'
      - '--storage.tsdb.path=/prometheus'

  grafana:
    image: grafana/grafana:latest
    volumes:
      - grafana_data:/var/lib/grafana
      - ./grafana/dashboards:/etc/grafana/provisioning/dashboards
    ports:
      - "3000:3000"
    environment:
      - GF_SECURITY_ADMIN_PASSWORD=admin
      - GF_USERS_ALLOW_SIGN_UP=false

  node-exporter:
    image: prom/node-exporter:latest
    ports:
      - "9100:9100"

volumes:
  prometheus_data:
  grafana_data:
```

### ELK Stack (Elasticsearch, Logstash, Kibana)

```yaml
# docker-compose.logging.yml
version: '3.8'

services:
  elasticsearch:
    image: docker.elastic.co/elasticsearch/elasticsearch:8.10.0
    environment:
      - discovery.type=single-node
      - "ES_JAVA_OPTS=-Xms512m -Xmx512m"
    ports:
      - "9200:9200"
    volumes:
      - elasticsearch_data:/usr/share/elasticsearch/data

  logstash:
    image: docker.elastic.co/logstash/logstash:8.10.0
    volumes:
      - ./logstash/pipeline:/usr/share/logstash/pipeline
    ports:
      - "5000:5000"
    depends_on:
      - elasticsearch

  kibana:
    image: docker.elastic.co/kibana/kibana:8.10.0
    ports:
      - "5601:5601"
    depends_on:
      - elasticsearch

volumes:
  elasticsearch_data:
```

## 🔒 Безопасность

### Security Checklist

- [ ] Все пароли изменены с дефолтных
- [ ] Используются Docker secrets для чувствительных данных
- [ ] SSL/TLS настроен для всех публичных endpoints
- [ ] Firewall настроен (только необходимые порты открыты)
- [ ] Регулярные security обновления
- [ ] Rate limiting настроен
- [ ] CORS правильно настроен
- [ ] Secrets не коммитятся в Git
- [ ] Логи не содержат чувствительных данных
- [ ] Backup настроен и тестируется регулярно

## 📚 Дополнительные ресурсы

- [Docker Security Best Practices](https://docs.docker.com/develop/security-best-practices/)
- [Kubernetes Documentation](https://kubernetes.io/docs/home/)
- [AWS ECS Best Practices](https://docs.aws.amazon.com/AmazonECS/latest/bestpracticesguide/)
- [Let's Encrypt Documentation](https://letsencrypt.org/docs/)

---

**Успешного развертывания!** 🚀

