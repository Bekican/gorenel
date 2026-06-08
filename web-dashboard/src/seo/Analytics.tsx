import { Head } from 'vite-react-ssg';

function plausibleConfig() {
  const domain = import.meta.env?.VITE_PLAUSIBLE_DOMAIN;
  const src = import.meta.env?.VITE_PLAUSIBLE_SRC || 'https://plausible.io/js/script.js';
  if (!domain || domain.trim() === '') return null;
  return { domain: domain.trim(), src: src.trim() };
}

function gaConfig() {
  const id = import.meta.env?.VITE_GA_ID;
  if (!id || id.trim() === '') return null;
  return { id: id.trim() };
}

export function Analytics() {
  const plausible = plausibleConfig();
  const ga = gaConfig();
  if (!plausible && !ga) return null;

  return (
    <Head>
      {plausible && <script defer data-domain={plausible.domain} src={plausible.src} />}

      {ga && (
        <>
          <script async src={`https://www.googletagmanager.com/gtag/js?id=${encodeURIComponent(ga.id)}`} />
          <script>{`window.dataLayer = window.dataLayer || [];
function gtag(){dataLayer.push(arguments);}
gtag('js', new Date());
gtag('config', '${ga.id}', { anonymize_ip: true });`}</script>
        </>
      )}
    </Head>
  );
}

