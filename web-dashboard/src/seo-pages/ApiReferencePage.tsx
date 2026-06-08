import { Link, useLocation } from 'react-router-dom';
import { Seo } from '../seo/Seo';

type Row = { method: string; path: string; auth: string; desc: string };

function apiCopy(lang: 'tr' | 'en') {
  const baseNote =
    lang === 'tr'
      ? 'İstekler, panel ile aynı kök alan adına yapılır (ör. https://gorenel.site). Tarayıcı oturumu çerez tabanlıdır; programatik kullanımda oturum çerezlerini veya sunucunun beklediği kimlik doğrulama biçimini iletmeniz gerekir.'
      : 'Call the same origin as the dashboard API host (e.g. https://gorenel.site). The browser session is cookie-based; for scripts, send session cookies or whatever auth the server expects.';

  const tables: { title: string; rows: Row[] }[] =
    lang === 'tr'
      ? [
          {
            title: 'Genel ve sistem',
            rows: [
              { method: 'GET', path: '/health', auth: 'Yok', desc: 'Sağlık kontrolü.' },
              { method: 'GET', path: '/metrics', auth: 'Yok*', desc: 'Tünel, istek ve bellek özet metrikleri. (*Rate limit uygulanabilir.)' },
              { method: 'GET', path: '/info', auth: 'Yok', desc: 'Sunucu sürümü ve platform bilgisi.' },
              { method: 'GET', path: '/analytics', auth: 'Yok', desc: 'Basit HTML analitik sayfası.' },
              { method: 'GET', path: '/api/analytics/realtime', auth: 'Yok*', desc: 'Gerçek zamanlı analitik anlık görüntüsü (JSON).' },
            ],
          },
          {
            title: 'Kimlik ve hesap',
            rows: [
              { method: 'POST', path: '/api/login', auth: 'Yok', desc: 'Giriş.' },
              { method: 'POST', path: '/api/register', auth: 'Yok', desc: 'Kayıt.' },
              { method: 'GET', path: '/api/callback', auth: '—', desc: 'OAuth geri çağırma (yapılandırmaya bağlı).' },
              { method: 'POST', path: '/api/logout', auth: 'Oturum', desc: 'Çıkış.' },
              { method: 'GET', path: '/api/me', auth: 'JWT / oturum', desc: 'Geçerli kullanıcı bilgisi.' },
            ],
          },
          {
            title: 'API anahtarları ve rezervasyonlar',
            rows: [
              { method: 'GET', path: '/api/keys', auth: 'Gerekli', desc: 'API anahtarlarını listele.' },
              { method: 'POST', path: '/api/keys', auth: 'Gerekli', desc: 'Yeni API anahtarı oluştur.' },
              { method: 'DELETE', path: '/api/keys?key=', auth: 'Gerekli', desc: 'Anahtar iptali (query parametresi key).' },
              { method: 'GET', path: '/api/reservations', auth: 'Gerekli', desc: 'Alt alan rezervasyonlarını listele.' },
              { method: 'POST', path: '/api/reservations', auth: 'Gerekli', desc: 'Rezervasyon oluştur (JSON: subdomain).' },
              { method: 'DELETE', path: '/api/reservations/{subdomain}', auth: 'Gerekli', desc: 'Rezervasyonu sil.' },
              {
                method: 'PUT',
                path: '/api/reservations/{subdomain}/assign',
                auth: 'Gerekli',
                desc: 'Rezervasyonu bir API anahtarına ata veya boş gövde ile ilişkiyi kaldır.',
              },
            ],
          },
          {
            title: 'Tüneller ve geçmiş',
            rows: [
              { method: 'GET', path: '/api/tunnels', auth: 'Gerekli', desc: 'Aktif tünelleri listele.' },
              { method: 'GET', path: '/api/tunnels/history', auth: 'Gerekli', desc: 'Son tünel oturum geçmişi.' },
            ],
          },
          {
            title: 'Tünel politikası',
            rows: [
              {
                method: 'PUT',
                path: '/api/tunnel-policy/{subdomain}',
                auth: 'Gerekli',
                desc: 'KeyAuth, IP allowlist, Basic Auth, HTTPS yönlendirme, hız limiti, başlık ve path kurallarını güncelle.',
              },
              {
                method: 'POST',
                path: '/api/tunnel-policy/{subdomain}/rotate',
                auth: 'Gerekli',
                desc: 'KeyAuth token’ını yenile; yanıtta yeni token döner.',
              },
            ],
          },
          {
            title: 'Inspector ve paylaşım',
            rows: [
              { method: 'GET', path: '/api/inspector/history', auth: 'Gerekli', desc: 'Yakalanan istek geçmişi.' },
              { method: 'GET', path: '/api/inspector/replay', auth: 'Gerekli', desc: 'Query id= ile isteği yeniden oynat.' },
              { method: 'GET', path: '/api/inspector/rules', auth: 'Gerekli', desc: 'Trafik kurallarını listele.' },
              { method: 'POST', path: '/api/inspector/rules', auth: 'Gerekli', desc: 'Yeni kural ekle (JSON gövde).' },
              { method: 'DELETE', path: '/api/inspector/rules?id=', auth: 'Gerekli', desc: 'Kuralı sil.' },
              { method: 'POST', path: '/api/shares?id=', auth: 'Gerekli', desc: 'Inspector kaydı için paylaşım oluştur.' },
              { method: 'GET', path: '/api/shares/{id}', auth: 'Yok', desc: 'Paylaşılan trace’i oku (herkese açık bağlantı).' },
            ],
          },
          {
            title: 'Güvenlik ve ML',
            rows: [
              { method: 'GET', path: '/api/anomalies', auth: 'Gerekli', desc: 'Son anomali kayıtları.' },
              { method: 'GET', path: '/api/ml/stats', auth: 'Gerekli', desc: 'ML istatistikleri ve servis durumu özeti.' },
            ],
          },
          {
            title: 'CLI kurulum',
            rows: [
              { method: 'GET', path: '/downloads/...', auth: 'Yok', desc: 'İkili indirmeler.' },
              { method: 'GET', path: '/v1/install', auth: 'Yok', desc: 'Kurulum yardımcı uç noktası.' },
              { method: 'GET', path: '/install.sh', auth: 'Yok', desc: 'Unix kurulum betiği.' },
              { method: 'GET', path: '/install.ps1', auth: 'Yok', desc: 'Windows kurulum betiği.' },
            ],
          },
          {
            title: 'Tünel taşıması (WebSocket)',
            rows: [
              {
                method: 'WS',
                path: '/tunnel/connect',
                auth: 'Protokole özel',
                desc: 'CLI tünel kontrolü için WebSocket (rate limit uygulanabilir).',
              },
            ],
          },
        ]
      : [
          {
            title: 'System',
            rows: [
              { method: 'GET', path: '/health', auth: 'None', desc: 'Health check.' },
              { method: 'GET', path: '/metrics', auth: 'None*', desc: 'Tunnel, request, and memory metrics. (*Rate limited.)' },
              { method: 'GET', path: '/info', auth: 'None', desc: 'Server version and platform.' },
              { method: 'GET', path: '/analytics', auth: 'None', desc: 'Simple HTML analytics page.' },
              { method: 'GET', path: '/api/analytics/realtime', auth: 'None*', desc: 'Realtime analytics snapshot (JSON).' },
            ],
          },
          {
            title: 'Auth & account',
            rows: [
              { method: 'POST', path: '/api/login', auth: 'None', desc: 'Sign in.' },
              { method: 'POST', path: '/api/register', auth: 'None', desc: 'Register.' },
              { method: 'GET', path: '/api/callback', auth: '—', desc: 'OAuth callback when configured.' },
              { method: 'POST', path: '/api/logout', auth: 'Session', desc: 'Sign out.' },
              { method: 'GET', path: '/api/me', auth: 'JWT / session', desc: 'Current user profile.' },
            ],
          },
          {
            title: 'API keys & reservations',
            rows: [
              { method: 'GET', path: '/api/keys', auth: 'Required', desc: 'List API keys.' },
              { method: 'POST', path: '/api/keys', auth: 'Required', desc: 'Create API key.' },
              { method: 'DELETE', path: '/api/keys?key=', auth: 'Required', desc: 'Revoke key (query param key).' },
              { method: 'GET', path: '/api/reservations', auth: 'Required', desc: 'List subdomain reservations.' },
              { method: 'POST', path: '/api/reservations', auth: 'Required', desc: 'Create reservation (JSON: subdomain).' },
              { method: 'DELETE', path: '/api/reservations/{subdomain}', auth: 'Required', desc: 'Delete reservation.' },
              {
                method: 'PUT',
                path: '/api/reservations/{subdomain}/assign',
                auth: 'Required',
                desc: 'Assign reservation to an API key, or clear with empty api_key.',
              },
            ],
          },
          {
            title: 'Tunnels',
            rows: [
              { method: 'GET', path: '/api/tunnels', auth: 'Required', desc: 'List active tunnels.' },
              { method: 'GET', path: '/api/tunnels/history', auth: 'Required', desc: 'Recent tunnel sessions.' },
            ],
          },
          {
            title: 'Tunnel policy',
            rows: [
              {
                method: 'PUT',
                path: '/api/tunnel-policy/{subdomain}',
                auth: 'Required',
                desc: 'Update KeyAuth, IP allowlist, Basic Auth, HTTPS redirect, rate limits, headers, path rewrites.',
              },
              {
                method: 'POST',
                path: '/api/tunnel-policy/{subdomain}/rotate',
                auth: 'Required',
                desc: 'Rotate KeyAuth token; response includes new token.',
              },
            ],
          },
          {
            title: 'Inspector & shares',
            rows: [
              { method: 'GET', path: '/api/inspector/history', auth: 'Required', desc: 'Captured request history.' },
              { method: 'GET', path: '/api/inspector/replay', auth: 'Required', desc: 'Replay by query id=.' },
              { method: 'GET', path: '/api/inspector/rules', auth: 'Required', desc: 'List traffic modification rules.' },
              { method: 'POST', path: '/api/inspector/rules', auth: 'Required', desc: 'Add rule (JSON body).' },
              { method: 'DELETE', path: '/api/inspector/rules?id=', auth: 'Required', desc: 'Delete rule.' },
              { method: 'POST', path: '/api/shares?id=', auth: 'Required', desc: 'Create share link for an inspector record.' },
              { method: 'GET', path: '/api/shares/{id}', auth: 'None', desc: 'Fetch shared trace (public link).' },
            ],
          },
          {
            title: 'Security & ML',
            rows: [
              { method: 'GET', path: '/api/anomalies', auth: 'Required', desc: 'Recent anomaly events.' },
              { method: 'GET', path: '/api/ml/stats', auth: 'Required', desc: 'ML stats and upstream health envelope.' },
            ],
          },
          {
            title: 'CLI install',
            rows: [
              { method: 'GET', path: '/downloads/...', auth: 'None', desc: 'Binary downloads.' },
              { method: 'GET', path: '/v1/install', auth: 'None', desc: 'Install helper endpoint.' },
              { method: 'GET', path: '/install.sh', auth: 'None', desc: 'Unix install script.' },
              { method: 'GET', path: '/install.ps1', auth: 'None', desc: 'PowerShell install script.' },
            ],
          },
          {
            title: 'Tunnel transport',
            rows: [
              {
                method: 'WS',
                path: '/tunnel/connect',
                auth: 'Protocol-specific',
                desc: 'WebSocket tunnel control for the CLI (rate limited).',
              },
            ],
          },
        ];

  const titles =
    lang === 'tr'
      ? {
          pageTitle: 'HTTP API Referansı | Gorenel',
          pageDesc: 'Gorenel kontrol düzlemi ve izleme API uç noktalarının özeti.',
          h1: 'HTTP API Referansı',
          colMethod: 'Metot',
          colPath: 'Yol',
          colAuth: 'Kimlik doğrulama',
          colDesc: 'Açıklama',
          cliTitle: 'CLI dokümantasyonu',
          cliText: 'Yerel kurulum ve tünel komutları için:',
          cliLink: 'CLI rehberi',
        }
      : {
          pageTitle: 'HTTP API Reference | Gorenel',
          pageDesc: 'Overview of Gorenel control-plane and monitoring HTTP endpoints.',
          h1: 'HTTP API Reference',
          colMethod: 'Method',
          colPath: 'Path',
          colAuth: 'Auth',
          colDesc: 'Description',
          cliTitle: 'CLI documentation',
          cliText: 'For local install and tunnel commands, see',
          cliLink: 'CLI guide',
        };

  return { baseNote, tables, titles };
}

export function ApiReferencePage() {
  const { pathname } = useLocation();
  const lang: 'tr' | 'en' = pathname.includes('/en/docs') ? 'en' : 'tr';
  const canonicalPath = lang === 'tr' ? '/tr/docs/api' : '/en/docs/api';
  const { baseNote, tables, titles } = apiCopy(lang);

  return (
    <main className="min-h-screen bg-[#080a10] text-white">
      <Seo
        lang={lang}
        title={titles.pageTitle}
        description={titles.pageDesc}
        canonicalPath={canonicalPath}
        hreflangs={[
          { hrefLang: 'tr', href: '/tr/docs/api' },
          { hrefLang: 'en', href: '/en/docs/api' },
          { hrefLang: 'x-default', href: '/tr/docs/api' },
        ]}
        jsonLd={{
          '@context': 'https://schema.org',
          '@type': 'TechArticle',
          headline: titles.h1,
          inLanguage: lang,
        }}
      />
      <div className="max-w-5xl mx-auto px-6 md:px-10 py-16 space-y-10">
        <header className="space-y-3">
          <h1 className="text-3xl md:text-4xl font-bold tracking-tight">{titles.h1}</h1>
          <p className="text-white/60 leading-relaxed max-w-3xl">{baseNote}</p>
        </header>

        <div className="space-y-10">
          {tables.map((block) => (
            <section key={block.title} className="space-y-4">
              <h2 className="text-lg font-semibold text-white/90 border-b border-white/[0.08] pb-2">{block.title}</h2>
              <div className="overflow-x-auto rounded-xl border border-white/[0.08] bg-white/[0.02]">
                <table className="w-full text-left text-xs md:text-sm">
                  <thead>
                    <tr className="border-b border-white/[0.08] text-white/45 uppercase tracking-wider text-[10px] md:text-[11px]">
                      <th className="p-3 font-medium whitespace-nowrap">{titles.colMethod}</th>
                      <th className="p-3 font-medium font-mono text-emerald-300/90">{titles.colPath}</th>
                      <th className="p-3 font-medium whitespace-nowrap">{titles.colAuth}</th>
                      <th className="p-3 font-medium">{titles.colDesc}</th>
                    </tr>
                  </thead>
                  <tbody>
                    {block.rows.map((row) => (
                      <tr key={`${row.method}:${row.path}`} className="border-b border-white/[0.04] last:border-0 text-white/70">
                        <td className="p-3 font-mono text-white/85 whitespace-nowrap">{row.method}</td>
                        <td className="p-3 font-mono text-[11px] md:text-xs text-emerald-200/80 break-all">{row.path}</td>
                        <td className="p-3 text-white/55 whitespace-nowrap">{row.auth}</td>
                        <td className="p-3 leading-snug">{row.desc}</td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            </section>
          ))}
        </div>

        <section className="rounded-xl border border-white/[0.08] bg-white/[0.02] p-5 space-y-2">
          <h2 className="text-sm font-semibold text-white/80">{titles.cliTitle}</h2>
          <p className="text-sm text-white/55">
            {titles.cliText}{' '}
            <Link className="text-emerald-400 hover:text-emerald-300" to="/tr/docs/cli">
              {titles.cliLink}
            </Link>
            .
          </p>
        </section>

        <nav className="pt-4 border-t border-white/[0.06] flex flex-wrap gap-3 text-sm">
          <Link className="text-emerald-400 hover:text-emerald-300" to="/">
            {lang === 'tr' ? 'Ana sayfa' : 'Home'}
          </Link>
          <Link className="text-emerald-400 hover:text-emerald-300" to={lang === 'tr' ? '/en/docs/api' : '/tr/docs/api'}>
            {lang === 'tr' ? 'English version' : 'Türkçe sürüm'}
          </Link>
        </nav>
      </div>
    </main>
  );
}
