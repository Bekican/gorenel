import { useLoaderData } from 'react-router-dom';
import { Seo } from '../../../seo/Seo';

/* eslint-disable react-refresh/only-export-components */

type DocData = {
  title: string;
  description: string;
  html: string;
  updatedAt: string;
};

export function TrCliDoc() {
  const data = useLoaderData() as DocData;
  return (
    <main className="min-h-screen bg-[#080a10] text-white">
      <Seo
        lang="tr"
        title={`${data.title} | Gorenel Docs`}
        description={data.description}
        canonicalPath="/tr/docs/cli"
        hreflangs={[
          { hrefLang: 'tr', href: '/tr/docs/cli' },
          { hrefLang: 'en', href: '/en/docs/cli' },
          { hrefLang: 'x-default', href: '/' },
        ]}
        jsonLd={{
          '@context': 'https://schema.org',
          '@type': 'TechArticle',
          headline: data.title,
          dateModified: data.updatedAt,
          inLanguage: 'tr',
          mainEntityOfPage: 'https://gorenel.site/tr/docs/cli',
        }}
      />
      <div className="max-w-4xl mx-auto px-6 md:px-10 py-16 space-y-8">
        <header className="space-y-2">
          <h1 className="text-3xl md:text-4xl font-bold tracking-tight">{data.title}</h1>
          <p className="text-white/55">{data.description}</p>
          <p className="text-xs text-white/25">Güncellendi: {new Date(data.updatedAt).toLocaleDateString('tr-TR')}</p>
        </header>

        <article
          className="prose prose-invert max-w-none prose-a:text-emerald-300 prose-pre:bg-black/30 prose-pre:border prose-pre:border-white/[0.08] prose-pre:rounded-xl"
          dangerouslySetInnerHTML={{ __html: data.html }}
        />
      </div>
    </main>
  );
}

export const Component = TrCliDoc;

export const entry = 'src/seo-pages/tr/docs/CliDoc.tsx';

export async function loader() {
  const fs = await import('node:fs/promises');
  const path = await import('node:path');
  const { marked } = await import('marked');

  const mdPath = path.resolve(process.cwd(), 'src/content/tr/docs/cli.md');
  const md = await fs.readFile(mdPath, 'utf8');
  const html = marked.parse(md) as string;
  return {
    title: 'CLI Kurulum ve Hızlı Başlangıç',
    description: 'Gorenel CLI kurulum adımları ve 3 adımda tünel açma rehberi.',
    html,
    updatedAt: new Date().toISOString(),
  };
}

