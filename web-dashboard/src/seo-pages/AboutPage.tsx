import { Link, useLocation } from 'react-router-dom';
import { Seo } from '../seo/Seo';
import { CheckCircle2, Shield, Brain, TerminalSquare, Server, Globe } from 'lucide-react';

function aboutCopy(lang: 'tr' | 'en') {
  if (lang === 'tr') {
    return {
      title: 'Hakkımızda | Gorenel',
      description: 'Gorenel ekibi ve açık kaynak tünelleme platformu vizyonu hakkında.',
      h1: 'Geleceğin Tünelleme Altyapısı',
      intro: 'Geliştiriciler ve güvenliğe önem veren ekipler için oluşturulan; hız, gözlemlenebilirlik ve sıfır güven mimarisini (Zero-Trust) birleştiren açık kaynaklı yeni nesil platformuz.',
      coreValues: [
        {
          title: "Yenilikçi Güvenlik Katmanı",
          desc: "Bizce localhost'u dünyaya açmak güvenlikten ödün vermek anlamına gelmez. Bu yüzden her tüneli rate-limit, KeyAuth ve ML destekli anomali tespiti (Isolation Forest) ile koruyoruz.",
          icon: Shield
        },
        {
          title: "Sınırsız Şeffaflık",
          desc: "Trafik sadece hedefe ulaşıp kaybolmaz. Gerçek zamanlı trafik izleyici paneli ile her isteği analiz eder, size %100 görünürlük sunarız. (ClickHouse entegre).",
          icon: Server
        },
        {
          title: "Açık Kaynak Kodlu",
          desc: "Güvene dayalı bir tünelin en temel şartı şeffaflıktır. Hem CLI hem sunucu mimarimiz GitHub üzerinde açıktır. Topluluk gücüyle büyüyoruz.",
          icon: Globe
        }
      ],
      back: 'Ana sayfa',
    };
  }
  return {
    title: 'About | Gorenel',
    description: 'About the Gorenel team and the vision behind our open-source tunneling platform.',
    h1: 'The Future of Tunneling Infrastructure',
    intro: 'We are an open-source, next-gen platform built for developers and security-conscious teams, unifying speed, observability, and Zero-Trust architecture.',
    coreValues: [
      {
        title: "Innovative Security Layer",
        desc: "Exposing localhost shouldn't mean sacrificing security. We protect every tunnel with rate-limits, KeyAuth, and ML-powered anomaly detection (Isolation Forest).",
        icon: Shield
      },
      {
        title: "Absolute Transparency",
        desc: "Traffic doesn't just reach the target and disappear. Our real-time traffic inspector analyzes every request, giving you 100% visibility (ClickHouse integrated).",
        icon: Server
      },
      {
        title: "Fully Open Source",
        desc: "The core requirement of a trusted tunnel is transparency. Both our CLI and server architecture are open on GitHub. We grow through community power.",
        icon: Globe
      }
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
    <main className="min-h-screen bg-[#080a10] text-white overflow-hidden">
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
          '@type': 'AboutPage',
          name: c.h1,
          url: `https://gorenel.site${canonicalPath}`,
        }}
      />
      
      {/* Background Decorators */}
      <div className="fixed inset-0 pointer-events-none z-0">
        <div className="absolute top-0 right-0 w-[500px] h-[500px] bg-emerald-500/[0.04] rounded-full blur-[140px]" />
        <div className="absolute bottom-1/4 left-1/4 w-[400px] h-[400px] bg-cyan-500/[0.03] rounded-full blur-[120px]" />
      </div>

      <div className="relative z-10 max-w-5xl mx-auto px-6 md:px-10 py-24 space-y-16">
        
        {/* Header Section */}
        <header className="space-y-6 text-center max-w-3xl mx-auto">
          <div className="inline-flex items-center gap-1.5 text-xs font-medium text-emerald-400/60 mb-2">
            <Brain className="w-4 h-4 text-emerald-400/70" />
            {lang === 'tr' ? 'VİZYONUMUZ' : 'OUR VISION'}
          </div>
          <h1 className="text-4xl md:text-6xl font-extrabold tracking-tight leading-[1.1]">
            <span className="text-transparent bg-clip-text bg-gradient-to-r from-white to-white/70">{c.h1}</span>
          </h1>
          <p className="text-white/60 leading-relaxed text-lg md:text-xl">
            {c.intro}
          </p>
        </header>

        {/* Feature/Values Grid */}
        <div className="grid grid-cols-1 md:grid-cols-3 gap-6 pt-8">
          {c.coreValues.map((val) => (
            <div key={val.title} className="p-8 rounded-3xl border border-white/[0.06] bg-gradient-to-br from-white/[0.02] to-transparent hover:bg-white/[0.04] hover:border-white/10 transition-all duration-300 group">
              <div className="w-12 h-12 rounded-xl bg-emerald-500/[0.08] border border-emerald-500/20 flex items-center justify-center mb-6 group-hover:border-emerald-500/40 transition-colors">
                <val.icon className="w-6 h-6 text-emerald-400" />
              </div>
              <h3 className="text-xl font-semibold text-white/90 mb-3">{val.title}</h3>
              <p className="text-white/50 leading-relaxed text-sm">
                {val.desc}
              </p>
            </div>
          ))}
        </div>

        {/* Story/Architecture Callout */}
        <div className="rounded-3xl border border-white/[0.06] bg-[#0c0e14]/50 backdrop-blur-xl p-8 md:p-12 overflow-hidden relative">
           <div className="absolute top-0 right-0 w-64 h-64 bg-violet-500/10 rounded-full blur-[80px]" />
           <div className="grid grid-cols-1 lg:grid-cols-2 gap-10 items-center relative z-10">
              <div className="space-y-4">
                <h3 className="text-2xl font-bold">{lang === 'tr' ? 'Mimari Harikası' : 'Architectural Marvel'}</h3>
                <p className="text-white/60 leading-relaxed text-sm">
                  {lang === 'tr' 
                    ? '10.000+ eşzamanlı bağlantıyı yönetebilecek gücü saniyeler içinde kullanımınıza sunuyoruz. Go ile yazılmış yüksek performanslı reverse proxy, Python tabanlı Machine Learning servisi ve React ekosistemiyle muazzam bir denge oluşturduk.' 
                    : 'We put the power to handle 10,000+ concurrent connections directly in your hands within seconds. Built with a high-performance Go reverse proxy, a Python-based ML service, and a React ecosystem for ultimate balance.'}
                </p>
                <ul className="space-y-2 pt-2 text-sm text-white/70">
                  <li className="flex items-center gap-2"><CheckCircle2 className="w-4 h-4 text-emerald-400"/> Go 1.24 Backend Processing</li>
                  <li className="flex items-center gap-2"><CheckCircle2 className="w-4 h-4 text-emerald-400"/> React 19 SSR Dashboard</li>
                  <li className="flex items-center gap-2"><CheckCircle2 className="w-4 h-4 text-emerald-400"/> Redis & ClickHouse Big Data Layer</li>
                </ul>
              </div>
              <div className="bg-black/40 border border-white/5 rounded-2xl p-6 font-mono text-xs md:text-sm text-emerald-400/80 shadow-[inset_0_0_20px_rgba(0,0,0,0.5)]">
                <div className="flex gap-2 mb-4">
                  <div className="w-3 h-3 rounded-full bg-red-500/50"></div>
                  <div className="w-3 h-3 rounded-full bg-yellow-500/50"></div>
                  <div className="w-3 h-3 rounded-full bg-emerald-500/50"></div>
                </div>
                <div className="text-white/50">$ gorenel status --deep</div>
                <div className="mt-2 ml-4">
                   <div className="text-white/80">Active System Components:</div>
                   <div className="mt-1">├── Gateway (Go) <span className="text-emerald-500">[1ms ping]</span></div>
                   <div>├── ML AI Sandbox <span className="text-emerald-500">[Active]</span></div>
                   <div>├── Traffic Store <span className="text-emerald-500">[ClickHouse OK]</span></div>
                   <div>└── Web Socket Manager <span className="text-emerald-500">[Syncing]</span></div>
                </div>
              </div>
           </div>
        </div>

        <nav className="pt-8 border-t border-white/[0.04] flex flex-wrap gap-4 text-sm justify-center">
          <Link className="flex items-center gap-2 px-5 py-2.5 rounded-xl bg-emerald-500/[0.08] hover:bg-emerald-500/[0.15] border border-emerald-500/20 hover:border-emerald-500/40 text-emerald-400 transition-all font-medium" to="/">
            {c.back}
          </Link>
          <a href="https://github.com/bekican/gorenel" target="_blank" rel="noopener noreferrer" className="flex items-center gap-2 px-5 py-2.5 rounded-xl bg-white/[0.03] hover:bg-white/[0.08] border border-white/10 text-white/80 transition-all font-medium">
             <TerminalSquare className="w-4 h-4" />
             GitHub Repo
          </a>
          <Link className="px-5 py-2.5 text-white/40 hover:text-white/80 transition-all ml-auto" to={lang === 'tr' ? '/en/about' : '/tr/hakkimizda'}>
            {lang === 'tr' ? 'English Version' : 'Türkçe Metin'}
          </Link>
        </nav>

      </div>
    </main>
  );
}
