import { Link, useLocation } from 'react-router-dom';
import { Seo } from '../seo/Seo';
import { Shield, BookOpen, AlertCircle, FileText } from 'lucide-react';

function termsCopy(lang: 'tr' | 'en') {
  if (lang === 'tr') {
    return {
      title: 'Hizmet Koşulları | Gorenel',
      description: 'Gorenel platformunun yasal bağlayıcılığı olan kullanım şartları, sorumluluk reddi ve veri politikaları.',
      h1: 'Hizmet Koşulları',
      updated: 'Son Güncelleme: 28 Mart 2026',
      intro: 'Gorenel platformunu (gorenel.site) ve açık kaynak komut satırı aracını (CLI) kullanarak aşağıda listelenen tüm şartları kayıtsız şartsız kabul etmiş sayılırsınız. Lütfen bu metni dikkatlice okuyunuz.',
      sections: [
         {
             icon: Shield,
             h: '1. Kabul Edilebilir Kullanım Politikası',
             p: 'Gorenel altyapısını yasadışı faaliyetler için kullanamazsınız. Phishing, malware dağıtımı, yasadışı içerik barındırma veya DDoS/Botnet ağlarına köprü oluşturmak kesinlikle yasaktır. Platformu yalnızca geliştirme, test veya meşru servis erişimi amaçlı projeleriniz için kullanabilirsiniz.'
         },
         {
             icon: AlertCircle,
             h: '2. Sorumluluk Reddi (As Is)',
             p: 'Bu platform ve sağlanan açık kaynak yazılım "olduğu gibi" (as is) prensibiyle sunulmuştur. Beklenmedik kesintiler yaşanabilir. Tünellenen verinin kaybından, üçüncü partilerin yetkisiz erişiminden veya ticari kayıplardan doğabilecek herhangi bir zarardan Gorenel sorumlu tutulamaz.'
         },
         {
             icon: BookOpen,
             h: '3. Güvenlik, Trafik Analizi ve ML Modelleri',
             p: 'Ağımız üzerinden geçen tüm trafik (HTTP başlıkları ve metadatası) güvenlik algoritmalarımız (Isolation Forest vb.) tarafından anlık incelenir. Anomali tespiti durumunda IP adresleri ve bağlantılar geçici veya kalıcı olarak engellenebilir. Gizli veya son derece hassas verilerinizin, şifrelenmediği (uçtan uca) durumlarda taşınması kullanıcının kendi riskidir.'
         },
         {
             icon: FileText,
             h: '4. Hesap Askıya Alma ve Feshetme',
             p: 'Sistem altyapısına zarar verme teşebbüsleri veya bu sözleşmenin 1. maddesinde belirtilen ihlaller söz konusu olduğunda, Gorenel haber vermeksizin kullanıcı hesabınızı ve ayrılmış alt alan adlarını (subdomains) iptal etme hakkını saklı tutar.'
         }
      ],
      back: 'Ana Sayfaya Dön',
    };
  }
  return {
    title: 'Terms of Service | Gorenel',
    description: 'Legally binding terms of use, disclaimers, and data policies for the Gorenel platform.',
    h1: 'Terms of Service',
    updated: 'Last Updated: March 28, 2026',
    intro: 'By using the Gorenel platform (gorenel.site) and open-source CLI, you unconditionally accept all the terms listed below. Please read carefully.',
    sections: [
        {
            icon: Shield,
            h: '1. Acceptable Use Policy',
            p: 'You may not use Gorenel infrastructure for illegal activities. Phishing, malware distribution, hosting illegal content, or bridging DDoS/Botnet networks is strictly prohibited. You may only use the platform for development, testing, or legitimate service access.'
        },
        {
            icon: AlertCircle,
            h: '2. Disclaimer of Warranties (As Is)',
            p: 'This platform and open-source software are provided on an "as is" and "as available" basis. Unexpected downtimes may occur. Gorenel shall not be held liable for any data loss, unauthorized third-party access, or financial losses arising from the use of the platform.'
        },
        {
            icon: BookOpen,
            h: '3. Security, Traffic Analysis, & ML Models',
            p: 'All traffic metadata passing through our network is analyzed in real-time by our security algorithms (e.g., Isolation Forest). If anomalies are detected, IPs and connections may be temporarily or permanently blocked. Transporting highly sensitive or confidential data without end-to-end encryption is at the user\'s own risk.'
        },
        {
            icon: FileText,
            h: '4. Account Suspension & Termination',
            p: 'In the event of attempts to harm system infrastructure or violations mentioned in Section 1, Gorenel reserves the right to cancel your user account and reserved subdomains without prior notice.'
        }
    ],
    back: 'Back to Home',
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
      <div className="max-w-4xl mx-auto px-6 md:px-10 py-16 space-y-12">
        <header className="space-y-4 max-w-2xl">
          <div className="inline-flex items-center gap-1.5 px-3 py-1 rounded-full bg-white/5 border border-white/10 text-[11px] font-semibold tracking-wider uppercase text-white/50">
             <FileText className="w-3.5 h-3.5 mr-1" />
             {lang === 'tr' ? 'HUKUKİ BELGELER' : 'LEGAL DOCUMENTS'}
          </div>
          <h1 className="text-3xl md:text-5xl font-extrabold tracking-tight">{c.h1}</h1>
          <p className="text-white/60 leading-relaxed text-lg pb-4 border-b border-white/[0.06]">{c.intro}</p>
          <p className="text-xs text-emerald-400 font-mono tracking-wide">{c.updated}</p>
        </header>

        <div className="grid grid-cols-1 md:grid-cols-2 gap-8 text-white/70 leading-relaxed text-sm md:text-base">
          {c.sections.map((s) => (
            <section key={s.h} className="group p-6 rounded-2xl bg-white/[0.02] border border-white/[0.04] hover:border-white/10 hover:bg-white/[0.04] transition-all duration-300">
               <div className="w-10 h-10 rounded-xl bg-white/[0.04] flex items-center justify-center mb-5 group-hover:scale-110 transition-transform">
                  <s.icon className="w-5 h-5 text-white/80" />
               </div>
               <h2 className="text-xl font-semibold text-white/90 mb-3">{s.h}</h2>
               <p className="text-white/50">{s.p}</p>
            </section>
          ))}
        </div>

        <nav className="pt-10 border-t border-white/[0.06] flex items-center justify-between text-sm">
          <Link className="px-5 py-2.5 rounded-xl border border-white/10 bg-white/5 text-white hover:bg-white/10 transition-all font-medium" to="/">
            {c.back}
          </Link>
          <Link className="text-emerald-500 hover:text-emerald-400 font-medium underline-offset-4 hover:underline" to={lang === 'tr' ? '/en/terms' : '/tr/hizmet-kosullari'}>
            {lang === 'tr' ? 'Read in English' : 'Türkçe Oku'}
          </Link>
        </nav>
      </div>
    </main>
  );
}
