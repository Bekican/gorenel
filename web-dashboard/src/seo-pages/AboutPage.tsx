import { Link, useLocation } from 'react-router-dom';
import { Seo } from '../seo/Seo';

function aboutCopy(lang: 'tr' | 'en') {
  if (lang === 'tr') {
    return {
      title: 'Hakkımızda | Gorenel',
      description: 'Gorenel ekibi ve açık kaynak tünelleme platformu vizyonu hakkında.',
      h1: 'Hakkımızda',
      intro: 'Geliştiriciler ve güvenliğe önem veren ekipler için açık kaynaklı, yeni nesil bir tünelleme platformuyuz.',
      content: [
        {
          h: 'Vizyonumuz',
          p: 'Gorenel\'in amacı, localhost\'unuzu internete açarken hız, güvenlik ve gözlemlenebilirliği bir araya getirmektir. Çoğu alternatifin aksine yapay zeka ile anomali tespiti, trafik izleyici ve tünel bazlı güvenlik politikalarını tek bir ücretsiz platformda sunuyoruz.',
        },
        {
          h: 'Açık Kaynak',
          p: 'Topluluğa ve açık kaynağa inanıyoruz. Gorenel CLI aracı tamamen açık kaynaklıdır; böylece nasıl çalıştığını inceleyebilir ve isterseniz katkıda bulunabilirsiniz.',
        },
        {
          h: 'İletişim',
          p: 'Soru ve geri bildirimleriniz bizim için çok değerli. İletişim kurmak veya hata bildirmek için GitHub reponuzu ziyaret edebilir, topluluk sayfalarına katılabilirsiniz.',
        },
      ],
      back: 'Ana sayfa',
    };
  }
  return {
    title: 'About | Gorenel',
    description: 'About the Gorenel team and the vision behind our open-source tunneling platform.',
    h1: 'About Us',
    intro: 'We are an open-source, next-gen tunneling platform designed for developers and security-conscious teams.',
    content: [
      {
        h: 'Our Vision',
        p: 'Our goal at Gorenel is to unite speed, security, and observability when exposing your localhost to the internet. Unlike most alternatives, we provide AI anomaly detection, traffic inspection, and tunnel-level security policies all in one free platform.',
      },
      {
        h: 'Open Source',
        p: 'We believe in the community and open source. The Gorenel CLI is completely open-source, so you can inspect how it works and contribute if you like.',
      },
      {
        h: 'Contact',
        p: 'Your questions and feedback are highly valuable to us. Visit our GitHub repository or join our community pages to get in touch or report an issue.',
      },
    ],
    back: 'Home',
  };
}

export function AboutPage() {
  const { pathname } = useLocation();
  const lang: 'tr' | 'en' = pathname.startsWith('/en') ? 'en' : 'tr';
  const canonicalPath = lang === 'tr' ? '/tr/hakkimizda' : '/en/about';
  const c = aboutCopy(lang);

  return (
    <main className="min-h-screen bg-[#080a10] text-white">
      <Seo
        lang={lang}
        title={c.title}
        description={c.description}
        canonicalPath={canonicalPath}
        hreflangs={[
          { hrefLang: 'tr', href: '/tr/hakkimizda' },
          { hrefLang: 'en', href: '/en/about' },
          { hrefLang: 'x-default', href: '/tr/hakkimizda' },
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
          {c.content.map((s) => (
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
          <Link className="text-emerald-400 hover:text-emerald-300" to={lang === 'tr' ? '/en/about' : '/tr/hakkimizda'}>
            {lang === 'tr' ? 'English version' : 'Türkçe metin'}
          </Link>
        </nav>
      </div>
    </main>
  );
}
