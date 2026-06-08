import './index.css';
import './i18n';
import { ViteReactSSG } from 'vite-react-ssg';
import { routes } from './routes';

// SSG + SPA hydration entrypoint.
export const createRoot = ViteReactSSG(
  { routes },
  ({ isClient, router }) => {
    if (isClient && router) {
      // Ensure client-side navigation always scrolls to top on route change.
      router.subscribe(() => {
        window.scrollTo(0, 0);
      });
    }
  },
);
