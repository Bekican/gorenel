import { Link } from 'react-router-dom';
import { Seo } from '../../seo/Seo';

export function TrNgrokAlternative() {
  return (
    <main className="min-h-screen bg-[#080a10] text-white">
      <Seo
        lang="tr"
        title="Ngrok alternatifi: Gorenel"
        description="Ngrok alternatifi arıyorsanız; Gorenel ile localhost'unuzu güvenli biçimde internete açın. Sabit URL, trafik politikaları ve anomali tespiti."
        canonicalPath="/tr/ngrok-alternatifi"
        hreflangs={[
          { hrefLang: 'tr', href: '/tr/ngrok-alternatifi' },
          { hrefLang: 'en', href: '/en/ngrok-alternative' },
          { hrefLang: 'x-default', href: '/' },
        ]}
        jsonLd={{
          '@context': 'https://schema.org',
          '@type': 'WebPage',
          name: 'Ngrok alternatifi: Gorenel',
          url: 'https://gorenel.site/tr/ngrok-alternatifi',
        }}
      />
      <div className="max-w-4xl mx-auto px-6 md:px-10 py-16 space-y-6">
        <header className="space-y-2">
          <h1 className="text-3xl md:text-4xl font-bold tracking-tight">Ngrok alternatifi: Gorenel</h1>
          <p className="text-white/55">
            Localhost’unuzu güvenli biçimde internete açın: sabit URL, trafik politikaları ve anomali tespiti.
          </p>
        </header>

        <section className="space-y-3 text-white/65 leading-relaxed">
          <p>
            Gorenel; tünel yönetimi, istek inceleme (inspector) ve güvenlik politikalarını tek panelde toplayan bir tunneling platformudur.
          </p>
          <ul className="list-disc pl-5 space-y-1">
            <li>Tek komutla HTTPS tünel</li>
            <li>API key + policy (IP allowlist, KeyAuth)</li>
            <li>Trafik inceleme ve paylaşılabilir trace</li>
          </ul>
        </section>

        <nav className="pt-4 border-t border-white/[0.06] flex flex-wrap gap-3 text-sm">
          <Link className="text-emerald-400 hover:text-emerald-300" to="/">Ana sayfa</Link>
          <Link className="text-emerald-400 hover:text-emerald-300" to="/tr/localhost-disari-acma">Localhost’u dışarı açma</Link>
          <Link className="text-emerald-400 hover:text-emerald-300" to="/app">Dashboard</Link>
        </nav>
      </div>
    </main>
  );
}

