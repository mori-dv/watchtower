# ==============================================================================
# Multi-Stage Production Dockerfile for Watchtower Control Center
# ==============================================================================

# --- Stage 1: Build & Bundle ---
FROM node:22-alpine AS builder

WORKDIR /app

# Copy package descriptors first to leverage Docker layer caching
COPY package*.json ./
RUN npm ci

# Copy application source
COPY . .

# Run production build (Vite + TypeScript + Tailwind)
RUN npm run build

# --- Stage 2: Lean Production Runtime ---
FROM nginx:alpine-slim AS runner

# Remove default Nginx HTML files
RUN rm -rf /usr/share/nginx/html/*

# Copy custom Nginx configuration with compression and security headers
COPY nginx.conf /etc/nginx/conf.d/default.conf

# Copy production static assets from builder stage
COPY --from=builder /app/dist /usr/share/nginx/html

# Expose standard HTTP port
EXPOSE 80

# Run Nginx in foreground
CMD ["nginx", "-g", "daemon off;"]
