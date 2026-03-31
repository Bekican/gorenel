import { Link, useLocation } from 'react-router-dom';
import { Seo } from '../seo/Seo';

function changelogCopy(lang: 'tr' | 'en') {
  if (lang === 'tr') {
    return {
      title: 'Değişiklik Günlüğü | Gorenel',
      description: 'Gorenel platformundaki son güncellemeler, yeni özellikler ve düzeltmeler.',
      h1: 'Değişiklik Günlüğü',
      intro: 'Gorenel tünelleme ve kontrol paneli hizmetindeki en son güncellemeleri buradan takip edebilirsiniz. Tüm platform iyileştirmelerini yakından inceleyin.',
      updates: [
        {
          version: 'v1.1.0',
          date: '28 Mart 2026',
          title: 'Yeni Panel, Performans İzleme ve ClickHouse Entegrasyonu',
          type: 'feature',
          features: [
            'Gerçek zamanlı trafik analizi için ClickHouse altyapısı devreye alındı. 100M+ istek saniyesinde filtrelenebiliyor.',
            'React Dashboard yenilendi; karanlık mod optmizasyonları ve mikro-animasyonlar eklendi.',
            'API metrikleri için Prometheus + Grafana entegrasyonu standart hale getirildi.',
            'Kullanıcı profili ve API Key yönetim paneli iyileştirildi.'
          ],
        },
        {
          version: 'v1.0.5',
          date: '15 Mart 2026',
          title: 'Gelişmiş AI Anomali Tespiti ve Isolation Forest',
          type: 'security',
          features: [
            'AI motorumuz güncellendi. Isolation Forest ve Autoencoder algoritmaları paralel çalıştırılarak tespit süresi ~1ms seviyesine düşürüldü.',
            'Rate limitleri aşıldığında veya IP tabanlı şüpheli trafik algılandığında e-posta veya webhook ile anlık bildirim sistemi eklendi.',
            'Zero-trust mimarisi doğrultusunda tünel başına özel kimlik doğrulama seçenekleri genişletildi (KeyAuth ve BasicAuth eklendi).'
          ],
        },
        {
          version: 'v1.0.0',
          date: '10 Mart 2026',
          title: 'Gorenel Genel Erişime Açıldı',
          type: 'release',
          features: [
            'Gorenel CLI açık kaynak olarak GitHub üzerinde yayımlandı.',
            'Otomatik Let\'s Encrypt SSL sertifikasyon altyapısı devreye alındı.',
            'Localhost\'tan ngrok kullanmadan güvenli ve yapay zeka korumalı anında tünel oluşturma desteği getirildi.',
            'Çok platformlu derlemeler hazırlandı (Windows, macOS, Linux, ARM64/AMD64).'
          ],
        },
      ],
      back: 'Ana sayfa',
    };
  }
  return {
    title: 'Changelog | Gorenel',
    description: 'Latest updates, new features, and fixes in the Gorenel platform.',
    h1: 'Changelog',
    intro: 'Track the latest updates to the Gorenel tunneling and control-plane service here. Discover all our platform improvements.',
    updates: [
      {
        version: 'v1.1.0',
        date: 'March 28, 2026',
        title: 'New Dashboard, Performance Monitoring & ClickHouse',
        type: 'feature',
        features: [
          'Deployed ClickHouse infrastructure for real-time traffic analysis. 100M+ requests can be filtered in milliseconds.',
          'Revamped React Dashboard; added dark mode optimizations and sleek micro-animations.',
          'Standardized Prometheus + Grafana integration for real-time API metrics.',
          'Improved account profile and API Key management pages.'
        ],
      },
      {
        version: 'v1.0.5',
        date: 'March 15, 2026',
        title: 'Advanced AI Anomaly Detection & Isolation Forest',
        type: 'security',
        features: [
          'Our AI engine is upgraded. Isolation Forest and Autoencoder now run in parallel reducing inference time to ~1ms.',
          'Added instant notifications (email/webhook) when rate limits are exceeded or suspicious traffic is detected.',
          'Expanded per-tunnel authentication options in line with Zero-Trust architecture (Added KeyAuth & BasicAuth).'
        ],
      },
      {
        version: 'v1.0.0',
        date: 'March 10, 2026',
        title: 'Gorenel Publicly Available',
        type: 'release',
        features: [
          'The Gorenel CLI is published open-source on GitHub.',
          'Automated Let\'s Encrypt SSL certification is live.',
          'Instantly establish secure, AI-protected tunnels from localhost without using ngrok.',
          'Multi-platform builds established (Windows, macOS, Linux, ARM64/AMD64).'
        ],
      },
    ],
    back: 'Home',
  };
}

export function ChangelogPage() {
  const { pathname } = useLocation();
  const lang: 'tr' | 'en' = pathname.startsWith('/en') ? 'en' : 'tr';
  const canonicalPath = lang === 'tr' ? '/tr/changelog' : '/en/changelog';
  const c = changelogCopy(lang);

  const getBadgeColor = (type: string) => {
    switch (type) {
      case 'release': return 'bg-cyan-500/10 text-cyan-400 border-cyan-500/20';
      case 'feature': return 'bg-emerald-500/10 text-emerald-400 border-emerald-500/20';
      case 'security': return 'bg-violet-500/10 text-violet-400 border-violet-500/20';
      default: return 'bg-white/10 text-white/70 border-white/20';
    }
  };

  const getBadgeText = (type: string) => {
    if (lang === 'tr') {
       if(type === 'release') return 'Sürüm';
       if(type === 'feature') return 'Özellik';
       if(type === 'security') return 'Güvenlik';
    } else {
       if(type === 'release') return 'Release';
       if(type === 'feature') return 'Feature';
       if(type === 'security') return 'Security';
    }
    return type;
  }

  return (
    <main className="min-h-screen bg-[#080a10] text-white">
      <Seo
        lang={lang}
        title={c.title}
        description={c.description}
        canonicalPath={canonicalPath}
        hreflangs={[
          { hrefLang: 'tr', href: '/tr/changelog' },
          { hrefLang: 'en', href: '/en/changelog' },
          { hrefLang: 'x-default', href: '/tr/changelog' },
        ]}
        jsonLd={{
          '@context': 'https://schema.org',
          '@type': 'WebPage',
          name: c.h1,
          url: `https://gorenel.site${canonicalPath}`,
        }}
      />
      <div className="max-w-4xl mx-auto px-6 md:px-10 py-16 space-y-12">
        <header className="space-y-4">
          <div className="inline-flex items-center gap-1.5 text-xs font-medium text-emerald-400/60 mb-2">
              <div className="w-2 h-2 rounded-full bg-emerald-400/60 animate-pulse"/>
              {lang === 'tr' ? 'GÜNCELLEMELER' : 'UPDATES'}
          </div>
          <h1 className="text-4xl md:text-5xl font-bold tracking-tight">{c.h1}</h1>
          <p className="text-white/60 leading-relaxed max-w-2xl text-lg">{c.intro}</p>
        </header>

        <div className="relative space-y-12 before:absolute before:inset-0 before:ml-5 before:-translate-x-px md:before:mx-auto md:before:translate-x-0 before:h-full before:w-0.5 before:bg-gradient-to-b before:from-transparent before:via-white/10 before:to-transparent">
          {c.updates.map((u) => (
            <div key={u.version} className="relative flex items-center justify-between md:justify-normal md:odd:flex-row-reverse group is-active">
              {/* Timeline dot */}
              <div className="flex items-center justify-center w-10 h-10 rounded-full border-4 border-[#080a10] bg-white/5 group-hover:bg-emerald-500/20 text-emerald-400 shadow shrink-0 md:order-1 md:group-odd:-translate-x-1/2 md:group-even:translate-x-1/2 absolute left-0 md:left-1/2 transform transition-colors duration-300 z-10">
                <div className="w-2 h-2 rounded-full bg-emerald-400"></div>
              </div>

              {/* Content card */}
              <div className="w-[calc(100%-4rem)] md:w-[calc(50%-2.5rem)] pl-16 md:pl-0 p-6 rounded-2xl border border-white/[0.06] bg-white/[0.02] hover:bg-white/[0.04] transition-all duration-300">
                <div className="flex flex-col-reverse md:flex-row md:justify-between md:items-center gap-3 mb-4">
                  <h2 className="text-xl font-bold text-white/90">
                    <span className="text-transparent bg-clip-text bg-gradient-to-r from-emerald-400 to-cyan-400">{u.version}</span>
                  </h2>
                  <div className="flex items-center gap-2">
                    <span className={`px-2 py-0.5 rounded-full border text-[10px] font-semibold uppercase tracking-wider ${getBadgeColor(u.type as string)}`}>
                        {getBadgeText(u.type as string)}
                    </span>
                    <span className="text-xs font-mono text-white/40">{u.date}</span>
                  </div>
                </div>
                
                <h3 className="text-lg font-semibold text-white/80 mb-3">{u.title}</h3>

                <ul className="space-y-2 text-white/60 text-sm">
                  {u.features.map((item) => (
                    <li key={item.slice(0, 40)} className="flex items-start">
                      <svg className="w-4 h-4 text-emerald-500/70 mr-2 mt-0.5 shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M5 13l4 4L19 7"></path></svg>
                      <span className="leading-relaxed">{item}</span>
                    </li>
                  ))}
                </ul>
              </div>
            </div>
          ))}
        </div>

        <nav className="pt-10 flex flex-wrap gap-4 text-sm justify-center">
          <Link className="px-4 py-2 rounded-lg bg-white/5 border border-white/10 text-emerald-400 hover:bg-emerald-500/10 hover:border-emerald-500/30 transition-all font-medium" to="/">
            {c.back}
          </Link>
          <Link className="px-4 py-2 rounded-lg bg-white/5 border border-white/10 text-emerald-400 hover:bg-emerald-500/10 hover:border-emerald-500/30 transition-all font-medium" to={lang === 'tr' ? '/en/changelog' : '/tr/changelog'}>
            {lang === 'tr' ? 'English Version' : 'Türkçe Metin'}
          </Link>
        </nav>
      </div>
    </main>
  );
}
