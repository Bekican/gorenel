FROM public.ecr.aws/docker/library/nginx:alpine

# Copy our custom configuration
COPY nginx.conf /etc/nginx/conf.d/default.conf

# Copy the pre-built dashboard files
COPY dist /usr/share/nginx/html

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]
