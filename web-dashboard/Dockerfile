# TODO [K8S]: 
#Dashboard için Multi-stage Dockerfile oluştur.
# Stage 1: Build (node:18)
# - npm install ve npm run build adımlarını ekle.

# Stage 2: Runtime (nginx:alpine)
# - build/dist klasörünü /usr/share/nginx/html içine kopyala.
# - Go monitoring API'sine yönlendirme yapan bir nginx.conf ekle.


FROM nginx:alpine
# Nginx konfigürasyonunu kopyala
COPY nginx.conf /etc/nginx/conf.d/default.conf
# Yerel build sonucunu kopyala (Daha önce npm run build yapılmış olmalı)
COPY dist /usr/share/nginx/html

EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
