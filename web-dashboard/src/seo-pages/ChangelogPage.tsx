import { Link, useLocation } from 'react-router-dom';
import { Seo } from '../seo/Seo';

function changelogCopy(lang: 'tr' | 'en') {
  if (lang === 'tr') {
    return {
      title: 'Değişiklik Günlüğü | Gorenel',
      description: 'Gorenel platformundaki son güncellemeler, yeni özellikler ve düzeltmeler.',
      h1: 'Değişiklik Günlüğü',
      intro: 'Gorenel tünelleme ve kontrol paneli hizmetindeki en son güncellemeleri buradan takip edebilirsiniz.',
      updates: [
        {
          version: 'v1.0.0',
          date: '10 Nisan 2026',
          title: 'Gorenel Genel Erişime Açıldı',
          features: [
            'AI destekli anomali tespiti eklendi (Isolation Forest & Autoencoder).',
            'Gerçek zamanlı trafik izleyici paneli devreye alındı.',
            'Açık kaynak tünel CLI yayımlandı.',
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
    intro: 'Track the latest updates to the Gorenel tunneling and control-plane service here.',
    updates: [
      {
        version: 'v1.0.0',
        date: 'April 10, 2026',
        title: 'Gorenel Public Release',
        features: [
          'Added AI-powered anomaly detection (Isolation Forest & Autoencoder).',
          'Deployed real-time traffic inspector dashboard.',
          'Released open-source tunneling CLI.',
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
      <div className="max-w-3xl mx-auto px-6 md:px-10 py-16 space-y-8">
        <header className="space-y-2">
          <h1 className="text-3xl md:text-4xl font-bold tracking-tight">{c.h1}</h1>
          <p className="text-white/60 leading-relaxed">{c.intro}</p>
        </header>

        <div className="space-y-8 text-white/70 leading-relaxed text-sm md:text-base">
          {c.updates.map((u) => (
            <section key={u.version} className="space-y-3 border border-white/[0.06] bg-white/[0.02] p-6 rounded-xl">
              <div className="flex justify-between items-center mb-4">
                <h2 className="text-xl font-semibold text-white/90">
                  <span className="text-emerald-400 mr-2">{u.version}</span> - {u.title}
                </h2>
                <span className="text-xs text-white/40">{u.date}</span>
              </div>
              <ul className="list-disc pl-5 space-y-1.5 text-white/60">
                {u.features.map((item) => (
                  <li key={item.slice(0, 40)}>{item}</li>
                ))}
              </ul>
            </section>
          ))}
        </div>

        <nav className="pt-6 border-t border-white/[0.06] flex flex-wrap gap-3 text-sm">
          <Link className="text-emerald-400 hover:text-emerald-300" to="/">
            {c.back}
          </Link>
          <Link className="text-emerald-400 hover:text-emerald-300" to={lang === 'tr' ? '/en/changelog' : '/tr/changelog'}>
            {lang === 'tr' ? 'English version' : 'Türkçe metin'}
          </Link>
        </nav>
      </div>
    </main>
  );
}
