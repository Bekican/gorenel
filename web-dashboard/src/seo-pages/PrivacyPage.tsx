import { Link, useLocation } from 'react-router-dom';
import { Seo } from '../seo/Seo';

function privacyCopy(lang: 'tr' | 'en') {
  if (lang === 'tr') {
    return {
      title: 'Gizlilik Politikası | Gorenel',
      description:
        'Gorenel hizmetini kullanırken hangi verilerin işlendiği, amaçlar ve haklarınız hakkında bilgi.',
      h1: 'Gizlilik Politikası',
      updated: 'Son güncelleme: 28 Mart 2026',
      intro:
        'Bu metin, Gorenel tünelleme ve kontrol paneli hizmeti kapsamında kişisel verilerin nasıl işlendiğini açıklar. Hizmeti kullanarak bu politikayı kabul etmiş sayılırsınız.',
      sections: [
        {
          h: 'Veri sorumlusu',
          p: 'Gorenel platformu (gorenel.site) üzerinden sunulan hizmetle ilgili taleplerinizi iletişim kanalları üzerinden bize iletebilirsiniz.',
        },
        {
          h: 'Toplanan veriler',
          ul: [
            'Hesap bilgileri: kayıt ve giriş için sağladığınız e-posta ve (isteğe bağlı) ad gibi kimlik bilgileri.',
            'Kullanım ve teknik veriler: tünel oturumları, alt alan adları, istek metadatası (yöntem, yol, durum kodu, zaman damgası), bant genişliği özetleri ve güvenlik/ML analizi için özellik vektörleri.',
            'Trafik inceleme (Inspector): panelde açıkça kaydedilen veya paylaşım bağlantısı oluşturduğunuz HTTP istek/yanıt örnekleri (içerik boyutu politikalarına tabi).',
            'Çerezler ve oturum: panelde oturumunuzu sürdürmek için gerekli kimlik doğrulama çerezleri/tokenları.',
          ],
        },
        {
          h: 'İşleme amaçları',
          ul: [
            'Hizmeti sağlamak, tünelleri yönlendirmek ve kotayı yönetmek.',
            'Güvenlik, kötüye kullanım önleme, hız sınırlama ve anomali tespiti.',
            'Destek, yasal yükümlülükler ve hizmet iyileştirmesi (anonim/özet istatistikler).',
          ],
        },
        {
          h: 'Hukuki dayanak',
          p: 'Verileriniz; sözleşmenin ifası, meşru menfaat (güvenlik ve altyapı bütünlüğü) ve açık rıza (varsa) kapsamında işlenir. Yerel mevzuat size ek haklar tanıyabilir.',
        },
        {
          h: 'Saklama',
          p: 'Veriler, hizmetin gerektirdiği süre boyunca ve yasal zorunluluklar çerçevesinde saklanır; gerektiğinde silme veya anonimleştirme uygulanır.',
        },
        {
          h: 'Üçüncü taraflar',
          p: 'Altyapı sağlayıcıları (ör. barındırma) hizmetin çalışması için sınırlı erişimle veri işleyebilir. Verilerinizi pazarlama amacıyla satmıyoruz.',
        },
        {
          h: 'Haklarınız',
          p: 'KVKK ve GDPR kapsamında erişim, düzeltme, silme, itiraz ve veri taşınabilirliği taleplerinde bizimle iletişime geçebilirsiniz.',
        },
        {
          h: 'Değişiklikler',
          p: 'Bu politika güncellenebilir. Önemli değişikliklerde site veya e-posta ile bilgilendirme yapılabilir.',
        },
      ],
      back: 'Ana sayfa',
    };
  }
  return {
    title: 'Privacy Policy | Gorenel',
    description:
      'What data Gorenel processes when you use the tunneling service and dashboard, and what choices you have.',
    h1: 'Privacy Policy',
    updated: 'Last updated: March 28, 2026',
    intro:
      'This policy explains how we handle personal data when you use Gorenel’s tunneling and control-plane services. By using the service, you acknowledge this policy.',
    sections: [
      {
        h: 'Data controller',
        p: 'For requests regarding the Gorenel service (gorenel.site), contact us through the channels we publish on the site.',
      },
      {
        h: 'Data we process',
        ul: [
          'Account data: email and optional name used for registration and login.',
          'Service & technical data: tunnel sessions, subdomains, request metadata (method, path, status, timestamps), bandwidth summaries, and features used for security/ML analysis.',
          'Traffic inspector: HTTP request/response samples you explicitly capture in the dashboard or share via a share link, subject to size limits.',
          'Cookies & session: authentication cookies/tokens required for the web dashboard.',
        ],
      },
      {
        h: 'Purposes',
        ul: [
          'Provide the service, route tunnel traffic, and enforce quotas.',
          'Security, abuse prevention, rate limiting, and anomaly detection.',
          'Support, legal compliance, and service improvement (aggregated statistics).',
        ],
      },
      {
        h: 'Legal bases',
        p: 'Processing relies on contract performance, legitimate interests (security and integrity), and consent where applicable. Local law may grant you additional rights.',
      },
      {
        h: 'Retention',
        p: 'We retain data as long as needed to operate the service and meet legal obligations, then delete or anonymize it where appropriate.',
      },
      {
        h: 'Processors & sharing',
        p: 'Infrastructure providers may process limited data to host the service. We do not sell your personal data for marketing.',
      },
      {
        h: 'Your rights',
        p: 'Depending on jurisdiction, you may have rights to access, rectify, erase, object, or export your data—contact us to exercise them.',
      },
      {
        h: 'Changes',
        p: 'We may update this policy; material changes may be communicated via the site or email.',
      },
    ],
    back: 'Home',
  };
}

export function PrivacyPage() {
  const { pathname } = useLocation();
  const lang: 'tr' | 'en' = pathname.startsWith('/en') ? 'en' : 'tr';
  const canonicalPath = lang === 'tr' ? '/tr/gizlilik-politikasi' : '/en/privacy';
  const c = privacyCopy(lang);

  return (
    <main className="min-h-screen bg-[#080a10] text-white">
      <Seo
        lang={lang}
        title={c.title}
        description={c.description}
        canonicalPath={canonicalPath}
        hreflangs={[
          { hrefLang: 'tr', href: '/tr/gizlilik-politikasi' },
          { hrefLang: 'en', href: '/en/privacy' },
          { hrefLang: 'x-default', href: '/tr/gizlilik-politikasi' },
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
              {'p' in s && s.p ? <p>{s.p}</p> : null}
              {'ul' in s && s.ul ? (
                <ul className="list-disc pl-5 space-y-1.5">
                  {s.ul.map((item) => (
                    <li key={item.slice(0, 40)}>{item}</li>
                  ))}
                </ul>
              ) : null}
            </section>
          ))}
        </div>

        <nav className="pt-6 border-t border-white/[0.06] flex flex-wrap gap-3 text-sm">
          <Link className="text-emerald-400 hover:text-emerald-300" to="/">
            {c.back}
          </Link>
          <Link className="text-emerald-400 hover:text-emerald-300" to={lang === 'tr' ? '/en/privacy' : '/tr/gizlilik-politikasi'}>
            {lang === 'tr' ? 'English version' : 'Türkçe metin'}
          </Link>
        </nav>
      </div>
    </main>
  );
}
