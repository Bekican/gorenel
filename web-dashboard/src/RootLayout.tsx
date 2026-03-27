import { Outlet } from 'react-router-dom';
import { Analytics } from './seo/Analytics';

export function RootLayout() {
  return (
    <>
      <Analytics />
      <Outlet />
    </>
  );
}

