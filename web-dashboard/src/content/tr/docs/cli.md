# CLI Kurulum ve Hızlı Başlangıç

Gorenel ile localhost’unuzu saniyeler içinde internete açabilirsiniz.

## 1) API key ayarla

Dashboard üzerinden API key oluşturun ve CLI’a kaydedin:

```bash
gorenel config set api_key YOUR_KEY
```

## 2) Tünel aç

```bash
gorenel connect --port 3000
```

## 3) Güvenlik (opsiyonel)

- IP allowlist
- KeyAuth (X-TOKEN)
- Rate limit

