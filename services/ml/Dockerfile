FROM python:3.11-slim
ENV PYTHONDONTWRITEBYTECODE=1 \
    PYTHONUNBUFFERED=1
WORKDIR /app

RUN pip install --no-cache-dir -r requirements.txt

COPY app.py .
COPY trained_models/ ./trained_models/

EXPOSE 5000

CMD ["gunicorn","--bind","0.0.0.0:5000","app:app"]


