import i18n from 'i18next';
import { initReactI18next } from 'react-i18next';
import LanguageDetector from 'i18next-browser-languagedetector';

const resources = {
  en: {
    translation: {
      "common": {
        "loading": "Loading...",
        "error": "Error",
        "save": "Save",
        "cancel": "Cancel",
        "delete": "Delete",
        "login": "Login",
        "register": "Register",
        "logout": "Logout",
        "overview": "Overview",
        "tunnels": "Tunnels",
        "ai_gateway": "AI Gateway",
        "traffic": "Traffic Inspector",
        "settings": "Settings",
        "api_keys": "API Keys",
        "register_title": "Create Account",
        "documentation": "Documentation",
        "pricing": "Pricing",
        "dashboard": "Dashboard",
        "connect": "Connect Localhost",
        "sign_out": "Sign Out",
        "overview_sub": "Dashboard & Stats",
        "tunnels_sub": "Local to Cloud",
        "ai_gateway_sub": "LLM Proxy & Metrics",
        "traffic_sub": "Live Packet Sniffer",
        "api_keys_sub": "Access Control",
        "settings_sub": "Traffic Rules",
        "provider_status": "Provider Status",
        "operational": "Operational"
      },
      "dashboard": {
        "command_center": "Command Center",
        "active_tunnels": "Active Tunnels",
        "ai_gateway": "AI Gateway",
        "traffic_inspector": "Traffic Inspector",
        "api_keys": "API Key Management",
        "global_rules": "Global Rules",
        "overview_desc": "Real-time system overview and performance metrics.",
        "tunnels_desc": "Manage your secure tunnels and endpoints.",
        "ai_desc": "Unified API for LLM routing, caching and rate limiting.",
        "traffic_desc": "Inspect and replay HTTP requests in real-time.",
        "keys_desc": "Manage your API keys for secure access and integrations.",
        "rules_desc": "Configure modification rules for incoming traffic.",
        "system_secure": "System Secure",
        "no_threats": "Anomaly detection is active. No threats detected in the last 24 hours.",
        "quick_start": "Quick Start Guide",
        "step_1": "1. Connect Localhost",
        "step_1_desc": "Use CLI to expose your local port.",
        "step_2": "2. Inspect Traffic",
        "step_2_desc": "See requests in the Traffic Inspector.",
        "step_3": "3. Set AI Rules",
        "step_3_desc": "Apply transformations or proxy LLMs."
      },
      "connect_modal": {
        "title": "Connect Your Device",
        "subtitle": "Securely expose your local app with Gorenel CLI.",
        "download_btn": "Download CLI",
        "download_url": "/downloads/gorenel-windows-amd64.exe",
        "command_label": "Connection Command",
        "command_footer": "Copy and paste this command into your terminal."
      },
      "api_keys_manager": {
        "title": "API Keys",
        "subtitle": "Manage your secure tunnel access credentials",
        "generate": "Generate New Key",
        "success": "New Key Generated Successfully!",
        "security_notice": "Make sure to copy it now. For security reasons, we won't show it again.",
        "revoke_confirm": "Are you sure you want to revoke this API key? This action cannot be undone.",
        "empty": "No API keys found. Create one to start tunneling.",
        "loading": "Loading security keys...",
        "onboarding_title": "How to use API Keys?",
        "onboarding_desc": "Use these keys to authenticate your CLI. Run 'gorenel login <key>' in your terminal to establish a secure session."
      },
      "ai_gateway": {
        "onboarding_title": "AI Protocol Intelligence",
        "onboarding_desc": "Gorenel's engine automatically detects anomalies and validates LLM payloads using neural autoencoders and isolation forests in real-time."
      },
      "tunnels": {
        "empty_title": "Ready to launch?",
        "empty_subtitle": "Your local services are just one command away from the global edge. Follow these 3 steps to go live.",
        "step1_title": "Get the CLI",
        "step1_desc": "Download the lightweight binary for your OS.",
        "step2_title": "Copy Login",
        "step2_desc": "Grab your unique connection command.",
        "step3_title": "Go Online",
        "step3_desc": "Paste into terminal and watch the magic happen.",
        "cta": "Start First Tunnel"
      },
      "modification_rules": {
        "title": "Traffic Rules",
        "subtitle": "Manipulate your application traffic in real-time.",
        "onboarding_title": "What does this page do?",
        "onboarding_desc": "Instead of changing your code repeatedly, you can add rules here to inject delays (Chaos), return fake responses (Mock), or redirect requests to other addresses.",
        "add_btn": "Add Rule",
        "path_trigger": "Request Path",
        "path_trigger_desc": "Which URLs should be affected? (e.g., /api/*)",
        "rewrite_target": "Rewrite Target",
        "chaos_delay": "Chaos Delay (ms)",
        "status_override": "Status Override",
        "mock_body": "Response Mocking (JSON)",
        "mock_body_desc": "Return this body immediately and bypass your server.",
        "header_injections": "Header Injections",
        "deploy_btn": "Deploy Rule",
        "abort_btn": "Abort",
        "zero_rules": "Zero Rules Active",
        "zero_rules_desc": "Traffic is flowing normally. Add rules to intercept and alter packets.",
        "active_modifier": "Active Modifier",
        "delete_btn": "Delete"
      },
      "landing": {
        "title": "Establish Session.",
        "subtitle": "Secure Tunneling Interface",
        "cta": "Connect to Gateway",
        "description": "Expose your localhost securely with built-in AI routing and monitoring."
      },
      "auth": {
        "no_account": "Don't have an account?",
        "have_account": "Already have an account?"
      },
      "traffic_inspector": {
        "title": "Packet Streams",
        "subtitle": "Real-time inspection of tunnel traffic",
        "capturing": "Capturing Live",
        "search_placeholder": "Search by path, method, or status...",
        "listening_title": "Listening for Packets",
        "listening_desc": "Gorenel is currently sniffing for frames on your active tunnels. Once you make a request to a public .site URL, it will appear here in real-time.",
        "live_sniffer": "Live Sniffer",
        "binary_logic": "Binary Logic",
        "method": "Method",
        "status": "Status",
        "path": "Cloud Path",
        "time": "Time",
        "actions": "Actions",
        "ai_inspector": "AI Intelligence Inspector",
        "ai_desc": "Protocol-aware LLM payload analysis",
        "model": "Model",
        "tokens": "Total Tokens",
        "prompt": "Input Prompt",
        "completion": "Model Completion",
        "req_frame": "Request Frame",
        "resp_stack": "Response Stack",
        "payload": "Payload",
        "body": "Body Output",
        "share": "Share Trace",
        "collapse": "Collapse Inspector"
      }
    }
  },
  tr: {
    translation: {
      "common": {
        "loading": "Yükleniyor...",
        "error": "Hata",
        "save": "Kaydet",
        "cancel": "İptal",
        "delete": "Sil",
        "login": "Giriş Yap",
        "register": "Kayıt Ol",
        "logout": "Çıkış Yap",
        "overview": "Genel Bakış",
        "tunnels": "Tüneller",
        "ai_gateway": "AI Geçidi",
        "traffic": "Trafik İzleyici",
        "settings": "Ayarlar",
        "api_keys": "API Anahtarları",
        "register_title": "Hesap Oluştur",
        "documentation": "Dokümantasyon",
        "pricing": "Fiyatlandırma",
        "dashboard": "Panel",
        "connect": "Localhost Bağla",
        "sign_out": "Çıkış Yap",
        "overview_sub": "Panel ve İstatistikler",
        "tunnels_sub": "Yerelden Buluta",
        "ai_gateway_sub": "LLM Vekil Sunucu",
        "traffic_sub": "Canlı Paket İzleyici",
        "api_keys_sub": "Erişim Kontrolü",
        "settings_sub": "Traffic Rules",
        "provider_status": "Provider Status",
        "operational": "Operational"
      },
      "dashboard": {
        "command_center": "Komuta Merkezi",
        "active_tunnels": "Aktif Tüneller",
        "ai_gateway": "AI Geçidi",
        "traffic_inspector": "Trafik İzleyici",
        "api_keys": "API Anahtar Yönetimi",
        "global_rules": "Genel Kurallar",
        "overview_desc": "Gerçek zamanlı sistem genel bakışı ve performans metrikleri.",
        "tunnels_desc": "Güvenli tünellerinizi ve uç noktalarınızı yönetin.",
        "ai_desc": "LLM yönlendirme, önbelleğe alma ve hız sınırlama için birleşik API.",
        "traffic_desc": "HTTP isteklerini gerçek zamanlı olarak izleyin ve tekrar oynatın.",
        "keys_desc": "Güvenli erişim ve entegrasyonlar için API anahtarlarınızı yönetin.",
        "rules_desc": "Gelen trafik için Traffic Rules yapılandırın.",
        "system_secure": "System Secure",
        "no_threats": "Anomaly Detection aktif. Son 24 saat içinde tehdit algılanmadı.",
        "quick_start": "Hızlı Başlangıç Rehberi",
        "step_1": "1. Localhost Bağla",
        "step_1_desc": "CLI kullanarak yerel portunu dünyaya aç.",
        "step_2": "2. Trafiği İzle",
        "step_2_desc": "Gelen istekleri Traffic Inspector'da gör.",
        "step_3": "3. AI Kuralları Koy",
        "step_3_desc": "Dönüşümler uygula veya LLM'leri proxy et."
      },
      "connect_modal": {
        "title": "Cihazını Bağla",
        "subtitle": "Gorenel CLI ile yerel uygulamana tünel aç.",
        "download_btn": "CLI'yı İndir",
        "download_url": "/downloads/gorenel-windows-amd64.exe",
        "command_label": "Bağlantı Komutu",
        "command_footer": "Komutu terminale yapıştır ve çalıştır."
      },
      "api_keys_manager": {
        "title": "API Anahtarları",
        "subtitle": "Güvenli tünel erişim kimlik bilgilerinizi yönetin",
        "generate": "Yeni Anahtar Oluştur",
        "success": "Yeni Anahtar Başarıyla Oluşturuldu!",
        "security_notice": "Şimdi kopyaladığınızdan emin olun. Güvenlik nedeniyle tekrar göstermeyeceğiz.",
        "revoke_confirm": "Bu API anahtarını iptal etmek istediğinizden emin misiniz? Bu işlem geri alınamaz.",
        "empty": "API anahtarı bulunamadı. Tünellemeye başlamak için bir tane oluşturun.",
        "loading": "Güvenlik anahtarları yükleniyor...",
        "onboarding_title": "API Anahtarları Nasıl Kullanılır?",
        "onboarding_desc": "Bu anahtarları CLI üzerinden giriş yapmak için kullanın. Terminalinizde 'gorenel login <anahtar>' komutunu çalıştırarak güvenli bir oturum başlatabilirsiniz."
      },
      "ai_gateway": {
        "onboarding_title": "AI Protocol Intelligence",
        "onboarding_desc": "Gorenel motoru, LLM yüklerini (payload) Neural Autoencoders ve Isolation Forest kullanarak gerçek zamanlı olarak analiz eder ve anomalileri tespit eder."
      },
      "tunnels": {
        "empty_title": "Yayına hazır mısın?",
        "empty_subtitle": "Yerel servislerin dünyadan sadece bir komut uzakta. Canlıya geçmek için bu 3 adımı izle.",
        "step1_title": "CLI'yı İndir",
        "step1_desc": "İşletim sistemin için hafif CLI aracını indir.",
        "step2_title": "Giriş Komutunu Kopyala",
        "step2_desc": "Sana özel bağlantı komutunu panelden al.",
        "step3_title": "Çevrimiçi Ol",
        "step3_desc": "Terminale yapıştır ve sihrin gerçekleşmesini izle.",
        "cta": "İlk Tüneli Başlat"
      },
      "modification_rules": {
        "title": "Trafik Kuralları",
        "subtitle": "Uygulamanızın trafiğini gerçek zamanlı manipüle edin.",
        "onboarding_title": "Bu sayfa ne işe yarar?",
        "onboarding_desc": "Kodunuzu her seferinde değiştirmek yerine, buradan kurallar ekleyerek API isteklerinize gecikme (Chaos) ekleyebilir, sahte cevaplar (Mock) dönebilir veya istekleri başka adreslere yönlendirebilirsiniz.",
        "add_btn": "Kural Ekle",
        "path_trigger": "İstek Yolu (Path)",
        "path_trigger_desc": "Hangi URL'ler bu kuraldan etkilensin? (Örn: /api/*)",
        "rewrite_target": "Yönlendirme Hedefi",
        "chaos_delay": "Gecikme Süresi (ms)",
        "status_override": "HTTP Durum Kodu",
        "mock_body": "Sahte Cevap İçeriği (JSON)",
        "mock_body_desc": "Sunucunuza gitmeden doğrudan bu cevabı döndürün.",
        "header_injections": "Header Ekleme",
        "deploy_btn": "Kuralı Yayına Al",
        "abort_btn": "Vazgeç",
        "zero_rules": "Aktif Kural Yok",
        "zero_rules_desc": "Şu an tüm trafik orijinal haliyle akıyor. Manipülasyon için kural ekleyin.",
        "active_modifier": "Aktif Kural",
        "delete_btn": "Sil"
      },
      "landing": {
        "title": "Oturum Açın.",
        "subtitle": "Güvenli Tünel Arayüzü",
        "cta": "Geçide Bağlan",
        "description": "Localhost'unuzu yerleşik AI yönlendirme ve izleme ile güvenli bir şekilde dünyaya açın."
      },
      "auth": {
        "no_account": "Hesabınız yok mu?",
        "have_account": "Zaten hesabınız var mı?"
      },
      "traffic_inspector": {
        "title": "Packet Streams",
        "subtitle": "Real-time inspection of tunnel traffic",
        "capturing": "Capturing Live",
        "search_placeholder": "Search by path, method, or status...",
        "listening_title": "Listening for Packets",
        "listening_desc": "Gorenel şu anda aktif tünellerinizdeki frame'leri dinliyor. Herhangi bir .site URL'sine istek yaptığınızda, burada gerçek zamanlı olarak görünecektir.",
        "live_sniffer": "Live Sniffer",
        "binary_logic": "Binary Logic",
        "method": "Metod",
        "status": "Durum",
        "path": "Cloud Path",
        "time": "Zaman",
        "actions": "İşlemler",
        "ai_inspector": "AI Intelligence Inspector",
        "ai_desc": "Protokol duyarlı LLM payload analizi",
        "model": "Model",
        "tokens": "Toplam Token",
        "prompt": "Input Prompt",
        "completion": "Model Completion",
        "req_frame": "Request Frame",
        "resp_stack": "Response Stack",
        "payload": "Payload",
        "body": "Body Output",
        "share": "Share Trace",
        "collapse": "Collapse Inspector"
      }
    }
  }
};

i18n
  .use(LanguageDetector)
  .use(initReactI18next)
  .init({
    resources,
    fallbackLng: 'en',
    interpolation: {
      escapeValue: false
    }
  });

export default i18n;
