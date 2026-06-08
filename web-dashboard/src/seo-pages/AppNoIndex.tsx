import App from '../App';
import { Seo } from '../seo/Seo';

export function AppNoIndex() {
  return (
    <>
      <Seo
        lang="tr"
        title="Gorenel Dashboard"
        description="Gorenel kullanıcı paneli."
        canonicalPath="/app"
        robots="noindex,nofollow"
      />
      <App />
    </>
  );
}

export const Component = AppNoIndex;

