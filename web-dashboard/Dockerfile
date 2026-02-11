# TODO [K8S]: 
#Dashboard için Multi-stage Dockerfile oluştur.
# Stage 1: Build (node:18)
# - npm install ve npm run build adımlarını ekle.

# Stage 2: Runtime (nginx:alpine)
# - build/dist klasörünü /usr/share/nginx/html içine kopyala.
# - Go monitoring API'sine yönlendirme yapan bir nginx.conf ekle.


FROM node:18-alpine AS build
WORKDIR /app

# Bağımlılıkları kopyalayıp kuruyoruz
COPY package*.json ./
RUN npm install

# Kaynak kodunu kopyalayıp build alıyoruz
COPY . .
RUN npm run build


FROM nginx:alpine
# Nginx konfigürasyonunu kopyala (K8s Service DNS'i ile çalışacak)
COPY nginx.conf /etc/nginx/conf.d/default.conf
# Build sonucunu kopyala
COPY --from=build /app/dist /usr/share/nginx/html

EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
