## SEO Ops Playbook (Gorenel)

### Goals
- **Indexable**: `/` + `/tr/*` docs/blog/landing pages.
- **Not indexable**: `/app/*` and `/share/*`.

### Publishing checklist (new page)
- **Route**: Add a pre-rendered route under `src/routes.tsx` (public only).
- **SEO**: Use `src/seo/Seo.tsx` with:
  - **title / description**
  - **canonicalPath**
  - **hreflangs** (tr/en + x-default)
  - **jsonLd** (WebPage / TechArticle / Article)
  - **robots** only when page must be noindexed
- **Sitemap**: Add the URL to `scripts/generate-sitemap.mjs`.
- **Internal links**: Link from:
  - landing (`/`)
  - relevant marketing pages
  - docs hub / blog hub (when added)

### Verification + analytics
- **Google Search Console**: set `VITE_GOOGLE_SITE_VERIFICATION`.
- **Plausible**: set `VITE_PLAUSIBLE_DOMAIN` (optional `VITE_PLAUSIBLE_SRC`).
- **Google Analytics**: set `VITE_GA_ID` (optional; anonymized IP enabled).

### Keyword map (starter)
- **/tr/ngrok-alternatifi**: "ngrok alternatifi", "ngrok alternative"
- **/tr/localhost-disari-acma**: "localhost dışarı açma", "localhost expose"
- **/tr/webhook-local-test**: "webhook local test", "stripe webhook test"
- **/tr/docs/cli**: "tunnel cli", "gorenel cli", "localhost tünel"

