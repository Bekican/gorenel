import { Link, useLocation } from 'react-router-dom';
import { Seo } from '../seo/Seo';

function termsCopy(lang: 'tr' | 'en') {
  if (lang === 'tr') {
    return {
      title: 'Hizmet Koşulları | Gorenel',
      description: 'Gorenel hizmet kullanım koşulları ve kuralları.',
      h1: 'Hizmet Koşulları',
      updated: 'Son güncelleme: 28 Mart 2026',
      intro: 'Bu şartlar, Gorenel tünelleme servisinin nasıl kullanılacağını belirtir. Hizmeti kullanarak bu koşulları kabul etmiş sayılırsınız.',
      sections: [
        {
          h: 'Kabul Edilebilir Kullanım',
          p: 'Gorenel hizmetini yasal amaçlarla ve kötüye kullanım (örneğin phishing, malware dağıtımı, spam veya diğer yasadışı içerikler) oluşturmadan kullanmayı kabul edersiniz.',
        },
        {
          h: 'Sorumluluk Reddi',
          p: 'Hizmet "olduğu gibi" sağlanır. Gorenel, kesintisiz hizmet veya mutlak güvenlik garantisi vermez. Tünel üzerinden aktarılan verilerin güvenliği sizin sorumluluğunuzdadır.',
        },
        {
          h: 'Hizmet Sonlandırma',
          p: 'Koşulların ihlali durumunda veya şüpheli/zararlı trafik tespiti durumunda önceden bildirim yapmaksızın hesabınızı veya tünelinizi askıya alma/silme hakkımızı saklı tutuyoruz.',
        },
        {
          h: 'Sorumluluğun Sınırlandırılması',
          p: 'Gorenel, hizmetin kullanımından veya kullanılamamasından doğacak doğrudan, dolaylı veya arızi zararlardan sorumlu tutulamaz.',
        },
      ],
      back: 'Ana sayfa',
    };
  }
  return {
    title: 'Terms of Service | Gorenel',
    description: 'Terms and conditions for using the Gorenel tunneling service.',
    h1: 'Terms of Service',
    updated: 'Last updated: March 28, 2026',
    intro: 'These terms describe how to use the Gorenel tunneling service. By using the service, you agree to these conditions.',
    sections: [
      {
        h: 'Acceptable Use',
        p: 'You agree to use Gorenel for lawful purposes and without engaging in abusive activities (e.g., phishing, malware distribution, spam, or other illegal content).',
      },
      {
        h: 'Disclaimer of Warranties',
        p: 'The service is provided "as is". Gorenel does not guarantee uninterrupted service or absolute security. The security of data transferred through the tunnel is your responsibility.',
      },
      {
        h: 'Termination',
        p: 'We reserve the right to suspend or terminate your account/tunnel without prior notice in case of terms violation or detection of suspicious/malicious traffic.',
      },
      {
        h: 'Limitation of Liability',
        p: 'Gorenel shall not be liable for any direct, indirect, or incidental damages arising from the use or inability to use the service.',
      },
    ],
    back: 'Home',
  };
}

export function TermsPage() {
  const { pathname } = useLocation();
  const lang: 'tr' | 'en' = pathname.startsWith('/en') ? 'en' : 'tr';
  const canonicalPath = lang === 'tr' ? '/tr/hizmet-kosullari' : '/en/terms';
  const c = termsCopy(lang);

  return (
    <main className="min-h-screen bg-[#080a10] text-white">
      <Seo
        lang={lang}
        title={c.title}
        description={c.description}
        canonicalPath={canonicalPath}
        hreflangs={[
          { hrefLang: 'tr', href: '/tr/hizmet-kosullari' },
          { hrefLang: 'en', href: '/en/terms' },
          { hrefLang: 'x-default', href: '/tr/hizmet-kosullari' },
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
          <p className="text-xs text-white/40">{c.updated}</p>
          <p className="text-white/60 leading-relaxed">{c.intro}</p>
        </header>

        <div className="space-y-8 text-white/70 leading-relaxed text-sm md:text-base">
          {c.sections.map((s) => (
            <section key={s.h} className="space-y-3">
              <h2 className="text-lg font-semibold text-white/90">{s.h}</h2>
              <p>{s.p}</p>
            </section>
          ))}
        </div>

        <nav className="pt-6 border-t border-white/[0.06] flex flex-wrap gap-3 text-sm">
          <Link className="text-emerald-400 hover:text-emerald-300" to="/">
            {c.back}
          </Link>
          <Link className="text-emerald-400 hover:text-emerald-300" to={lang === 'tr' ? '/en/terms' : '/tr/hizmet-kosullari'}>
            {lang === 'tr' ? 'English version' : 'Türkçe metin'}
          </Link>
        </nav>
      </div>
    </main>
  );
}
