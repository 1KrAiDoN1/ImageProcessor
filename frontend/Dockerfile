# Frontend Dockerfile - Multi-stage build

# Stage 1: Build stage (optional, for minification)
FROM node:18-alpine AS build

WORKDIR /app

# Copy package files
COPY package*.json ./

# Install dependencies (if needed for build tools)
RUN npm install --only=production || true

# Copy all frontend files
COPY . .

# Stage 2: Production stage with Nginx
FROM nginx:alpine

# Copy custom nginx configuration
COPY nginx.conf /etc/nginx/conf.d/default.conf

# Copy frontend files to nginx html directory
COPY --from=build /app /usr/share/nginx/html

# Remove unnecessary files
RUN rm -f /usr/share/nginx/html/Dockerfile \
    /usr/share/nginx/html/nginx.conf \
    /usr/share/nginx/html/package*.json \
    /usr/share/nginx/html/.gitignore

# Expose port 80
EXPOSE 80

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --quiet --tries=1 --spider http://localhost/ || exit 1

# Start nginx
CMD ["nginx", "-g", "daemon off;"]

