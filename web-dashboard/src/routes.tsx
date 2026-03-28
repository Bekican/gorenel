/* eslint-disable react-refresh/only-export-components */
import React from 'react';
import type { RouteRecord } from 'vite-react-ssg';
import { RootLayout } from './RootLayout';
import { LandingSeo } from './seo-pages/LandingSeo';

// Public marketing pages (TR-first). We'll expand these with docs/blog in later todos.
const TrNgrokAlternative = React.lazy(() => import('./seo-pages/tr/NgrokAlternatifi').then(m => ({ default: m.TrNgrokAlternative })));
const TrLocalhostExpose = React.lazy(() => import('./seo-pages/tr/LocalhostDisariAcma').then(m => ({ default: m.TrLocalhostExpose })));
const TrWebhookLocal = React.lazy(() => import('./seo-pages/tr/WebhookLocalTest').then(m => ({ default: m.TrWebhookLocalTest })));
const PrivacyPage = React.lazy(() => import('./seo-pages/PrivacyPage').then((m) => ({ default: m.PrivacyPage })));
const ApiReferencePage = React.lazy(() => import('./seo-pages/ApiReferencePage').then((m) => ({ default: m.ApiReferencePage })));

export const routes: RouteRecord[] = [
  {
    path: '/',
    element: <RootLayout />,
    entry: 'src/RootLayout.tsx',
    children: [
      // Public landing (pre-render)
      {
        path: '/',
        element: (
          <Page>
            <LandingSeo />
          </Page>
        ),
        entry: 'src/seo-pages/LandingSeo.tsx',
      },

      // TR marketing pages (pre-render)
      {
        path: '/tr/ngrok-alternatifi',
        element: (
          <Page>
            <TrNgrokAlternative />
          </Page>
        ),
        entry: 'src/seo-pages/tr/NgrokAlternatifi.tsx',
      },
      {
        path: '/tr/localhost-disari-acma',
        element: (
          <Page>
            <TrLocalhostExpose />
          </Page>
        ),
        entry: 'src/seo-pages/tr/LocalhostDisariAcma.tsx',
      },
      {
        path: '/tr/webhook-local-test',
        element: (
          <Page>
            <TrWebhookLocal />
          </Page>
        ),
        entry: 'src/seo-pages/tr/WebhookLocalTest.tsx',
      },
      {
        path: '/tr/docs/cli',
        lazy: () => import('./seo-pages/tr/docs/CliDoc'),
      },
      {
        path: '/tr/gizlilik-politikasi',
        element: (
          <Page>
            <PrivacyPage />
          </Page>
        ),
        entry: 'src/seo-pages/PrivacyPage.tsx',
      },
      {
        path: '/en/privacy',
        element: (
          <Page>
            <PrivacyPage />
          </Page>
        ),
        entry: 'src/seo-pages/PrivacyPage.tsx',
      },
      {
        path: '/tr/docs/api',
        element: (
          <Page>
            <ApiReferencePage />
          </Page>
        ),
        entry: 'src/seo-pages/ApiReferencePage.tsx',
      },
      {
        path: '/en/docs/api',
        element: (
          <Page>
            <ApiReferencePage />
          </Page>
        ),
        entry: 'src/seo-pages/ApiReferencePage.tsx',
      },

      // Share (default noindex; still needs to work)
      {
        path: '/share/:id',
        lazy: () => import('./seo-pages/ShareNoIndex'),
      },

      // App shell (dashboard). Not intended for indexing.
      {
        path: '/app/*',
        lazy: () => import('./seo-pages/AppNoIndex'),
      },

      // Back-compat: keep old /dashboard entrypoint working by redirecting to /app.
      {
        path: '/dashboard',
        element: <Redirect to="/app" />,
        entry: 'src/routes.tsx',
      },
    ],
  },
];

function Redirect({ to }: { to: string }) {
  if (typeof window !== 'undefined') window.location.replace(to);
  return null;
}

function Page({ children }: { children: React.ReactNode }) {
  return (
    <React.Suspense fallback={<div className="min-h-screen bg-[#080a10] text-white flex items-center justify-center">Loading…</div>}>
      {children}
    </React.Suspense>
  );
}

