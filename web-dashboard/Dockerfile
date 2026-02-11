# TODO [K8S]: 
#Dashboard için Multi-stage Dockerfile oluştur.
# Stage 1: Build (node:18)
# - npm install ve npm run build adımlarını ekle.

# Stage 2: Runtime (nginx:alpine)
# - build/dist klasörünü /usr/share/nginx/html içine kopyala.
# - Go monitoring API'sine yönlendirme yapan bir nginx.conf ekle.


FROM node:18-alpine AS build
WORKDIR /app

RUN npm install
RUN npm run build


FROM nginx:alpine
COPY --from=build /app/dist /usr/share/nginx/html
EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
