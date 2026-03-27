import { Head } from 'vite-react-ssg';

export type HreflangLink = { hrefLang: string; href: string };

type Props = {
  title: string;
  description: string;
  canonicalPath?: string;
  lang?: 'tr' | 'en';
  robots?: string; // e.g. "index,follow" or "noindex,nofollow"
  ogImagePath?: string;
  hreflangs?: HreflangLink[];
  jsonLd?: Record<string, unknown> | Array<Record<string, unknown>>;
};

const DEFAULT_SITE_URL = 'https://gorenel.site';

function siteUrl(): string {
  const fromEnv = import.meta.env?.VITE_SITE_URL;
  if (fromEnv && typeof fromEnv === 'string' && fromEnv.trim() !== '') return fromEnv.trim().replace(/\/+$/, '');
  return DEFAULT_SITE_URL;
}

function absoluteUrl(pathOrUrl: string): string {
  if (/^https?:\/\//i.test(pathOrUrl)) return pathOrUrl;
  const base = siteUrl();
  if (!pathOrUrl.startsWith('/')) return `${base}/${pathOrUrl}`;
  return `${base}${pathOrUrl}`;
}

export function Seo(props: Props) {
  const canonicalPath = props.canonicalPath ?? '/';
  const canonical = absoluteUrl(canonicalPath);
  const ogImage = absoluteUrl(props.ogImagePath ?? '/logo.png');
  const lang = props.lang ?? 'tr';
  const googleSiteVerification = import.meta.env?.VITE_GOOGLE_SITE_VERIFICATION;

  const jsonLd = props.jsonLd
    ? (Array.isArray(props.jsonLd) ? props.jsonLd : [props.jsonLd])
    : [];

  return (
    <Head>
      <html lang={lang} />
      <title>{props.title}</title>
      <meta name="description" content={props.description} />
      {googleSiteVerification && <meta name="google-site-verification" content={googleSiteVerification} />}
      {props.robots && <meta name="robots" content={props.robots} />}
      <link rel="canonical" href={canonical} />

      {/* Open Graph */}
      <meta property="og:type" content="website" />
      <meta property="og:title" content={props.title} />
      <meta property="og:description" content={props.description} />
      <meta property="og:url" content={canonical} />
      <meta property="og:image" content={ogImage} />
      <meta property="og:locale" content={lang === 'tr' ? 'tr_TR' : 'en_US'} />

      {/* Twitter */}
      <meta name="twitter:card" content="summary_large_image" />
      <meta name="twitter:title" content={props.title} />
      <meta name="twitter:description" content={props.description} />
      <meta name="twitter:image" content={ogImage} />

      {/* hreflang */}
      {props.hreflangs?.map((l) => (
        <link key={`${l.hrefLang}:${l.href}`} rel="alternate" hrefLang={l.hrefLang} href={absoluteUrl(l.href)} />
      ))}

      {/* JSON-LD */}
      {jsonLd.map((obj, idx) => (
        <script key={idx} type="application/ld+json">
          {JSON.stringify(obj)}
        </script>
      ))}
    </Head>
  );
}

