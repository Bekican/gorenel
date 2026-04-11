# Build stage
FROM node:18-alpine AS builder

WORKDIR /app

# Install dependencies
COPY package*.json ./
RUN npm install

# Copy source and build
COPY . .
RUN npm run build

# Production stage
FROM public.ecr.aws/docker/library/nginx:alpine

# Copy our custom configuration
COPY nginx.conf /etc/nginx/conf.d/default.conf

# Copy the pre-built dashboard files from the builder stage
COPY --from=builder /app/dist /usr/share/nginx/html

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]
