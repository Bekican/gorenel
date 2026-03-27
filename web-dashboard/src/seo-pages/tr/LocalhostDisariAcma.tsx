import { Link } from 'react-router-dom';
import { Seo } from '../../seo/Seo';

export function TrLocalhostExpose() {
  return (
    <main className="min-h-screen bg-[#080a10] text-white">
      <Seo
        lang="tr"
        title="Localhost’u dışarı açma (güvenli tünel) | Gorenel"
        description="Localhost’u güvenli şekilde internete açın: webhook testleri, demo paylaşımı ve prod debug için HTTPS tünel. IP allowlist ve KeyAuth desteği."
        canonicalPath="/tr/localhost-disari-acma"
        hreflangs={[
          { hrefLang: 'tr', href: '/tr/localhost-disari-acma' },
          { hrefLang: 'en', href: '/en/expose-localhost' },
          { hrefLang: 'x-default', href: '/' },
        ]}
        jsonLd={{
          '@context': 'https://schema.org',
          '@type': 'WebPage',
          name: 'Localhost’u dışarı açma (güvenli tünel)',
          url: 'https://gorenel.site/tr/localhost-disari-acma',
        }}
      />
      <div className="max-w-4xl mx-auto px-6 md:px-10 py-16 space-y-6">
        <header className="space-y-2">
          <h1 className="text-3xl md:text-4xl font-bold tracking-tight">Localhost’u dışarı açma (güvenli tünel)</h1>
          <p className="text-white/55">
            Webhook testleri, demo paylaşımı ve production debug için localhost’unuzu güvenli şekilde yayınlayın.
          </p>
        </header>

        <section className="space-y-3 text-white/65 leading-relaxed">
          <p>
            En pratik senaryo: yerelde çalışan uygulamanızı tek komutla internete açıp sabit bir URL üzerinden erişmek.
          </p>
          <pre className="rounded-xl border border-white/[0.08] bg-black/30 p-4 text-xs overflow-auto">
{`gorenel config set api_key YOUR_KEY
gorenel connect --port 3000`}
          </pre>
          <p>
            İsterseniz IP allowlist, KeyAuth veya Basic Auth gibi politikalarla tüneli “zero-trust” yaklaşımıyla kilitleyebilirsiniz.
          </p>
        </section>

        <nav className="pt-4 border-t border-white/[0.06] flex flex-wrap gap-3 text-sm">
          <Link className="text-emerald-400 hover:text-emerald-300" to="/tr/ngrok-alternatifi">Ngrok alternatifi</Link>
          <Link className="text-emerald-400 hover:text-emerald-300" to="/tr/webhook-local-test">Webhook local test</Link>
          <Link className="text-emerald-400 hover:text-emerald-300" to="/">Ana sayfa</Link>
        </nav>
      </div>
    </main>
  );
}

