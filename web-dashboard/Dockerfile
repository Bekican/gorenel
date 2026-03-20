FROM nginx:alpine

# Nginx konfigürasyonunu kopyala
COPY web-dashboard/nginx.conf /etc/nginx/conf.d/default.conf

# YERELDE DERLENMİŞ 'dist' klasörünü kopyala (npm run build yapılmış olmalı)
COPY web-dashboard/dist /usr/share/nginx/html

EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
