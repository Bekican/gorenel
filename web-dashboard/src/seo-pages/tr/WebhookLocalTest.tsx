import { Link } from 'react-router-dom';
import { Seo } from '../../seo/Seo';

export function TrWebhookLocalTest() {
  return (
    <main className="min-h-screen bg-[#080a10] text-white">
      <Seo
        lang="tr"
        title="Webhook local test (Stripe/GitHub vb.) | Gorenel"
        description="Stripe/GitHub gibi webhook’ları lokal ortamda test etmek için HTTPS bir endpoint’e ihtiyaç duyarsınız. Gorenel ile saniyeler içinde tünel açın."
        canonicalPath="/tr/webhook-local-test"
        hreflangs={[
          { hrefLang: 'tr', href: '/tr/webhook-local-test' },
          { hrefLang: 'en', href: '/en/webhook-local-test' },
          { hrefLang: 'x-default', href: '/' },
        ]}
        jsonLd={{
          '@context': 'https://schema.org',
          '@type': 'WebPage',
          name: 'Webhook local test',
          url: 'https://gorenel.site/tr/webhook-local-test',
        }}
      />
      <div className="max-w-4xl mx-auto px-6 md:px-10 py-16 space-y-6">
        <header className="space-y-2">
          <h1 className="text-3xl md:text-4xl font-bold tracking-tight">Webhook local test (Stripe/GitHub vb.)</h1>
          <p className="text-white/55">
            Webhook’ları lokal ortamda test etmek için HTTPS bir endpoint’e ihtiyaç duyarsınız. Gorenel bunu saniyeler içinde sağlar.
          </p>
        </header>

        <section className="space-y-3 text-white/65 leading-relaxed">
          <ol className="list-decimal pl-5 space-y-2">
            <li>Local server’ınızı başlatın (örn. port 3000).</li>
            <li>Gorenel ile tünel açın.</li>
            <li>Webhook provider’da URL olarak public endpoint’i girin.</li>
          </ol>
          <pre className="rounded-xl border border-white/[0.08] bg-black/30 p-4 text-xs overflow-auto">
{`gorenel connect --port 3000
# size verilen https://<subdomain>.gorenel.site adresini webhook URL olarak kullanın`}
          </pre>
        </section>

        <nav className="pt-4 border-t border-white/[0.06] flex flex-wrap gap-3 text-sm">
          <Link className="text-emerald-400 hover:text-emerald-300" to="/tr/localhost-disari-acma">Localhost’u dışarı açma</Link>
          <Link className="text-emerald-400 hover:text-emerald-300" to="/tr/ngrok-alternatifi">Ngrok alternatifi</Link>
          <Link className="text-emerald-400 hover:text-emerald-300" to="/">Ana sayfa</Link>
        </nav>
      </div>
    </main>
  );
}

