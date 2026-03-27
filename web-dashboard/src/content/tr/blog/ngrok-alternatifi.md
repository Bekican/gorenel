# Ngrok alternatifi: neden Gorenel?

Localhost’u internete açmanın en hızlı yolu bir tünel servisidir. Gorenel; güvenlik politikaları, trafik inceleme ve anomali tespitiyle bunu bir platforma dönüştürür.

## Ne zaman gerekli?

- Webhook testleri (Stripe, GitHub, vb.)\n- Müşteri demosu\n- Prod debug / hızlı paylaşım\n+
## Hızlı kurulum

```bash
gorenel config set api_key YOUR_KEY
gorenel connect --port 3000
```

