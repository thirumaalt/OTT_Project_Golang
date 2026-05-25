# frontend-ui/Dockerfile
# Stage 1: build
FROM node:20-alpine AS builder
WORKDIR /app

# Install build tools
RUN apk add --no-cache python3 make g++ bash

# Copy package manifests first for better caching
COPY package.json ./

# Install deps (including devDependencies required for Tailwind & Vite build)
RUN npm install

# Copy source
COPY . .

# Build (Tailwind will run via PostCSS during Vite build)
ENV NODE_ENV=production
RUN npm run build

# Stage 2: production image (lightweight)
FROM nginx:stable-alpine AS runner
# Copy built assets from builder
COPY --from=builder /app/dist /usr/share/nginx/html

# Serve single-page app fallback for index.html
RUN apk add --no-cache bash
COPY nginx.conf /etc/nginx/conf.d/default.conf

EXPOSE 5173
CMD ["nginx", "-g", "daemon off;"]
