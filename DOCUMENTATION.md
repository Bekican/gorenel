# Gorenel: Modern Tünelleme ve Trafik Yönetim Sistemi

Gorenel, yerel ağınızda çalışan servisleri güvenli bir şekilde internete açmanızı sağlayan, Go diliyle geliştirilmiş, yüksek performanslı ve genişletilebilir bir tünelleme sistemidir. Ngrok benzeri bir işlevsellik sunarken, trafik izleme (inspection), gelişmiş analitik ve özel alan adı desteği gibi kurumsal düzeyde özellikler barındırır.

## 🚀 Temel Özellikler

- **Çoklu Protokol Desteği:** HTTP, TCP ve UDP trafiğini tünelleme yeteneği.
- **Alt Alan Adı ve Özel Domain:** `kullanici.gorenel.io` gibi otomatik subdomainler veya kendi özel alan adlarınızı (`api.sirketim.com`) kullanma imkanı.
- **Trafik Denetçisi (Inspector):** Gelen istekleri ve dönen cevapları gerçek zamanlı yakalama, inceleme ve tekrar oynatma (replay).
- **Gelişmiş Analitik:** İstek sayısı, bant genişliği kullanımı, yanıt süreleri ve coğrafi konum bazlı trafik analizi.
- **Güvenlik ve Kontrol:** API Key tabanlı kimlik doğrulama ve IP/Subdomain bazlı hız sınırlama (Rate Limiting).
- **Modern Web Dashboard:** React ve TypeScript ile geliştirilmiş, sistem durumunu ve trafiği izleyebileceğiniz görsel arayüz.
- **Bulut Hazır:** Docker desteği ve Kubernetes (Helm) chartları ile kolay dağıtım.

---

## 🏗️ Mimari Yapı

Gorenel, **İstemci-Sunucu (Client-Server)** mimarisi üzerine kuruludur ve bağlantı çoklama (multiplexing) için **Yamux** protokolünü kullanır.

### 1. Gorenel Server (Sunucu)
Sistemin merkezi kontrol noktasıdır. Üç ana bileşenden oluşur:
- **Control Plane (7000):** İstemcilerin bağlandığı, tünel kayıtlarının yapıldığı ve kontrol mesajlarının iletildiği port.
- **Proxy Plane (8080):** İnternetten gelen HTTP trafiğini karşılayıp ilgili tünellere yönlendiren bileşen.
- **Monitoring Plane (9090):** Metriklerin, analitiklerin ve yönetim API'larının sunulduğu port.

### 2. Gorenel Client (İstemci)
Yerel makinenizde çalışan ve sunucu ile güvenli bir Yamux oturumu başlatan hafif bir CLI aracıdır. Gelen istekleri yerel portunuza (örn: localhost:3000) yönlendirir.

---

## 🛠️ Teknik Detaylar

### Tünel Yönetimi (Tunnel Manager)
`internal/server/tunnel_manager.go` içerisinde yer alan bu bileşen, aktif yamux oturumlarını ve domain eşleşmelerini hafızada tutar. Thread-safe bir yapıya sahiptir ve bağlantı koptuğunda kaynakları otomatik olarak temizler.

### İletişim Protokolü
İstemci ve sunucu arasındaki ilk el sıkışma JSON tabanlı bir protokol ile gerçekleştirilir.
- `RegisterRequest`: İstemci kimliği, API anahtarı, istenen domain ve tünel tipini içerir.
- `RegisterResponse`: Sunucunun atadığı URL ve (varsa) public port bilgilerini döner.

### Trafik İzleme ve Değiştirme
`TrafficInspector` bileşeni, proxy üzerinden geçen her isteği yakalayarak UUID ile işaretler ve geçmişe kaydeder. `TrafficModifier` ise belirli kurallar çerçevesinde (header ekleme/çıkarma, body değiştirme) trafiğe müdahale edebilir.

---

## 📊 İzleme ve Analitik API'ları (Port 9090)

Sistem durumunu izlemek için aşağıdaki uç noktalar kullanılabilir:

- `GET /metrics`: Aktif tüneller, istek sayıları, bant genişliği ve sistem kaynak kullanımı (JSON).
- `GET /health`: Sistemin çalışma durumu ve uptime bilgisi.
- `GET /api/analytics/realtime`: Gerçek zamanlı trafik istatistikleri.
- `GET /api/inspector/history`: Yakalanan son HTTP isteklerinin listesi.
- `POST /api/inspector/replay?id={uuid}`: Belirli bir isteği tekrar yerel servise gönderir.

---

## 🚀 Başlarken

### Sunucuyu Çalıştırma
```bash
go run cmd/server/main.go
```
Varsayılan olarak Kontrol: 7000, Proxy: 8080 ve Monitoring: 9090 portlarını dinler.

### İstemciyi Çalıştırma
```bash
# HTTP Tüneli (Port 3000'i dışarı açar)
go run cmd/client/main.go start --port 3000 --api-key YOUR_KEY

# TCP Tüneli (Port 5432'yi dışarı açar)
go run cmd/client/main.go start --port 5432 --type tcp --api-key YOUR_KEY
```

---

## 📦 Dağıtım (Deployment)

### Docker
Proje içerisinde hem istemci (`Dockerfile.client`) hem de sunucu (`Dockerfile.server`) için optimize edilmiş Docker dosyaları bulunmaktadır. `docker-compose.yml` ile tüm stack (Server, Client, Dashboard) tek komutla ayağa kaldırılabilir.

### Kubernetes
`helm/` dizini altındaki chartlar ile Gorenel sistemini Kubernetes kümenize kurumsal standartlarda dağıtabilirsiniz.

---

## 💻 Teknoloji Yığını

- **Backend:** Go (Golang)
- **Frontend:** React, TypeScript, Vite, Tailwind CSS
- **Kütüphaneler:**
    - `hashicorp/yamux`: Bağlantı çoklama.
    - `spf13/cobra` & `viper`: CLI ve konfigürasyon yönetimi.
    - `uber-go/zap`: Yüksek performanslı loglama.
    - `golang-jwt`: Kimlik doğrulama.
    - `google/uuid`: Benzersiz tanımlayıcılar.

---
*Hazırlayan: Gemini CLI - Gorenel Proje Analiz Raporu*
