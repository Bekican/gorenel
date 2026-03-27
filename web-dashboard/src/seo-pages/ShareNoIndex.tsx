import { useParams } from 'react-router-dom';
import { ShareView } from '../components/ShareView';
import { Seo } from '../seo/Seo';

export function ShareNoIndex() {
  const { id } = useParams();
  const shareId = id || '';
  return (
    <>
      <Seo
        lang="tr"
        title="Paylaşılan Trace | Gorenel"
        description="Paylaşılan istek/yanıt iz kaydı."
        canonicalPath={shareId ? `/share/${shareId}` : '/share'}
        robots="noindex,nofollow"
      />
      <ShareView shareId={shareId} />
    </>
  );
}

export const Component = ShareNoIndex;

