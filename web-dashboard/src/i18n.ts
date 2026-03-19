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
        "settings_sub": "Traffic Rules"
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
        "download_btn": "Download CLI (.zip)",
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
        "loading": "Loading security keys..."
      },
      "landing": {
        "title": "Establish Session.",
        "subtitle": "Secure Tunneling Interface",
        "cta": "Connect to Gateway",
        "description": "Expose your localhost securely with built-in AI routing and monitoring."
      },
      "auth": {
        "welcome": "Welcome back",
        "name": "Display Name",
        "email": "Email Address",
        "password": "Access Key",
        "no_account": "Don't have an account?",
        "have_account": "Already have an account?"
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
        "settings_sub": "Trafik Kuralları"
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
        "rules_desc": "Gelen trafik için değişiklik kurallarını yapılandırın.",
        "system_secure": "Sistem Güvenli",
        "no_threats": "Anomali tespiti aktif. Son 24 saat içinde tehdit algılanmadı.",
        "quick_start": "Hızlı Başlangıç Rehberi",
        "step_1": "1. Localhost Bağla",
        "step_1_desc": "CLI kullanarak yerel portunu dünyaya aç.",
        "step_2": "2. Trafiği İzle",
        "step_2_desc": "Gelen istekleri Trafik İzleyici'de gör.",
        "step_3": "3. AI Kuralları Koy",
        "step_3_desc": "Dönüşümler uygula veya LLM'leri vekil et."
      },
      "connect_modal": {
        "title": "Cihazını Bağla",
        "subtitle": "Gorenel CLI ile yerel uygulamana tünel aç.",
        "download_btn": "CLI İndir (.zip)",
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
        "loading": "Güvenlik anahtarları yükleniyor..."
      },
      "landing": {
        "title": "Oturum Açın.",
        "subtitle": "Güvenli Tünel Arayüzü",
        "cta": "Geçide Bağlan",
        "description": "Localhost'unuzu yerleşik AI yönlendirme ve izleme ile güvenli bir şekilde dünyaya açın."
      },
      "auth": {
        "welcome": "Tekrar hoşgeldiniz",
        "name": "Görünen Ad",
        "email": "E-posta Adresi",
        "password": "Erişim Anahtarı",
        "no_account": "Hesabınız yok mu?",
        "have_account": "Zaten hesabınız var mı?"
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
