FROM public.ecr.aws/docker/library/python:3.11-slim
ENV PYTHONDONTWRITEBYTECODE=1 \
    PYTHONUNBUFFERED=1
WORKDIR /app

# Önce bağımlılık dosyasını kopyalıyoruz ki cache'den yararlanalım
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

# Sonra uygulama kodlarını kopyalıyoruz
COPY app.py .
COPY feature_engineering.py .
COPY model_registry.py .
COPY models/ ./models/
COPY trained_models/ ./trained_models/

EXPOSE 5000

CMD ["gunicorn","--bind","0.0.0.0:5000","app:app"]


