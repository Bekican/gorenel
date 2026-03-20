# Phase 1: Go Builder (Build CLI)
FROM golang:1.21-alpine AS cli-builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN GOOS=windows GOARCH=amd64 go build -ldflags "-s -w" -o gorenel-windows-amd64.exe cmd/client/main.go

# Phase 2: Node Builder (Build Dashboard)
FROM node:18-alpine AS node-builder
WORKDIR /app/web-dashboard
COPY web-dashboard/package*.json ./
RUN npm install
COPY web-dashboard/ .
# Arka plan API URL'sini prod için set et (Dilersen build-arg ile de geçebilirsin)
# ENV VITE_API_URL=https://gorenel.site
RUN npm run build

# Phase 3: Runtime (Nginx)
FROM nginx:alpine
# Nginx konfigürasyonunu kopyala
COPY web-dashboard/nginx.conf /etc/nginx/conf.d/default.conf

# React build sonucunu kopyala
COPY --from=node-builder /app/web-dashboard/dist /usr/share/nginx/html

# CLI binary dosyasını public/downloads klasörüne kopyala
RUN mkdir -p /usr/share/nginx/html/downloads
COPY --from=cli-builder /app/gorenel-windows-amd64.exe /usr/share/nginx/html/downloads/gorenel-windows-amd64.exe

EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
