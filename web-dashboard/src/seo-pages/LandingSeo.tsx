import React, { Suspense } from 'react';
import { Seo } from '../seo/Seo';

const LandingPage = React.lazy(() =>
  import('../components/LandingPage').then((m) => ({ default: m.LandingPage })),
);

export function LandingSeo() {
  return (
    <>
      <Seo
        lang="tr"
        title="Gorenel | Localhost'u güvenli şekilde dünyaya açın"
        description="Güvenli tüneller, sabit URL’ler, trafik politikaları ve yapay zeka ile anomali tespiti — tek bir CLI komutuyla."
        canonicalPath="/"
        hreflangs={[
          { hrefLang: 'tr', href: '/' },
          { hrefLang: 'en', href: '/en' },
          { hrefLang: 'x-default', href: '/' },
        ]}
        jsonLd={[
          {
            '@context': 'https://schema.org',
            '@type': 'SoftwareApplication',
            name: 'Gorenel',
            applicationCategory: 'DeveloperApplication',
            operatingSystem: 'Windows, Linux, macOS',
            offers: { '@type': 'Offer', price: '0', priceCurrency: 'USD' },
          },
          {
            '@context': 'https://schema.org',
            '@type': 'WebSite',
            name: 'Gorenel',
            url: 'https://gorenel.site/',
          },
        ]}
      />
      <Suspense
        fallback={
          <div
            className="min-h-screen bg-[#080a10] text-white flex items-center justify-center"
            aria-busy="true"
            aria-live="polite"
          >
            <span className="text-sm text-white/70">Loading…</span>
          </div>
        }
      >
        <LandingPage onLogin={() => { if (typeof window !== 'undefined') window.location.href = '/app'; }} />
      </Suspense>
    </>
  );
}

