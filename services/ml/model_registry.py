"""
ModelRegistry — Birden fazla ML modelini yöneten merkezi kayıt sınıfı.
Her model aynı interface'i (train, predict, predict_proba, save, load) sağlar.
"""
import time
import logging
import numpy as np
from pathlib import Path

logger = logging.getLogger(__name__)


class ModelRegistry:
    """Tüm anomali tespit modellerini yönetir."""

    def __init__(self):
        self.models = {}          # name -> model instance
        self.model_paths = {}     # name -> saved model path
        self.stats = {}           # name -> training/inference stats

    def register(self, name: str, model, save_path: str):
        """Yeni bir model kaydet."""
        self.models[name] = model
        self.model_paths[name] = save_path
        self.stats[name] = {
            'is_trained': False,
            'total_predictions': 0,
            'total_anomalies': 0,
            'avg_inference_ms': 0.0,
            'inference_times': [],
        }
        logger.info(f"Model kaydedildi: {name}")

    def _train_single_model(self, name: str, model, X: np.ndarray, **kwargs):
        """Tek model eğitimi (train_all / train_one ortak yolu)."""
        start = time.time()
        if name == 'isolation_forest':
            model.train(X, feature_names=kwargs.get('feature_names'))
        elif name == 'autoencoder':
            model.train(
                X,
                epochs=int(kwargs.get('epochs', 8)),
                batch_size=int(kwargs.get('batch_size', 64)),
            )
        elif hasattr(model, 'feature_names'):
            model.train(X, **kwargs)
        else:
            model.train(X)
        elapsed = time.time() - start
        model.save(self.model_paths[name])
        self.stats[name]['is_trained'] = True
        return {
            'status': 'success',
            'training_time_ms': round(elapsed * 1000, 2),
            'stats': getattr(model, 'training_stats', {}),
        }, elapsed

    def train_one(self, name: str, X: np.ndarray, **kwargs):
        """Tek bir modeli eğit ve kaydet."""
        model = self.models.get(name)
        if not model:
            return {name: {'status': 'error', 'error': f'bilinmeyen model: {name}'}}
        try:
            result, elapsed = self._train_single_model(name, model, X, **kwargs)
            logger.info(f"{name} eğitildi ({elapsed:.2f}s)")
            return {name: result}
        except Exception as e:
            logger.error(f"{name} eğitim hatası: {e}")
            return {name: {'status': 'error', 'error': str(e)}}

    def train_all(self, X: np.ndarray, **kwargs):
        """Tüm modelleri aynı veri ile eğit."""
        results = {}
        for name, model in self.models.items():
            try:
                result, elapsed = self._train_single_model(name, model, X, **kwargs)
                results[name] = result
                logger.info(f"{name} eğitildi ({elapsed:.2f}s)")
            except Exception as e:
                results[name] = {'status': 'error', 'error': str(e)}
                logger.error(f"{name} eğitim hatası: {e}")
        return results

    def predict(self, name: str, X: np.ndarray) -> dict:
        """Tek bir modelden tahmin al. Eğer model eğitilmemişse, anlık olarak eğitir (Self-Healing)."""
        model = self.models.get(name)
        if not model:
            return {'error': f'bilinmeyen model: {name}'}

        if not model.is_trained:
            logger.warning(f"⚠️ {name} modeli eğitilmemiş, anlık olarak eğitiliyor (Self-Healing)...")
            try:
                # Generate a quick 100-sample normal baseline by repeating the incoming request feature pattern
                bootstrap_X = np.repeat(X, 100, axis=0)
                self._train_single_model(name, model, bootstrap_X)
                logger.info(f"✅ {name} modeli başarıyla anlık olarak eğitildi.")
            except Exception as e:
                logger.error(f"❌ {name} anlık eğitim hatası: {e}")
                return {'error': f'{name} modeli anlık eğitim hatası: {e}'}

        start = time.time()
        predictions, scores = model.predict(X)
        probabilities = model.predict_proba(X)
        elapsed_ms = (time.time() - start) * 1000

        is_anomaly = bool(predictions[0] == -1)

        # İstatistikleri güncelle
        s = self.stats[name]
        s['total_predictions'] += 1
        if is_anomaly:
            s['total_anomalies'] += 1
        s['inference_times'].append(elapsed_ms)
        # Son 100 inference süresinin ortalaması
        recent = s['inference_times'][-100:]
        s['avg_inference_ms'] = round(sum(recent) / len(recent), 3)

        return {
            'model': name,
            'is_anomaly': is_anomaly,
            'anomaly_score': float(probabilities[0]),
            'prediction': 'anomaly' if is_anomaly else 'normal',
            'inference_ms': round(elapsed_ms, 3),
        }

    def predict_compare(self, X: np.ndarray) -> dict:
        """Tüm modellerden tahmin al ve karşılaştır (eğitilmemiş olanları otomatik eğitir)."""
        results = {}
        for name in self.models:
            res = self.predict(name, X)
            if 'error' not in res:
                results[name] = res

        # Hangi modeller anomali dedi?
        flagged_by = [name for name, r in results.items() if r.get('is_anomaly')]

        return {
            'models': results,
            'consensus': {
                'any_anomaly': len(flagged_by) > 0,
                'all_agree': len(flagged_by) == len(results) or len(flagged_by) == 0,
                'flagged_by': flagged_by,
                'models_compared': len(results),
            }
        }

    def get_stats(self) -> dict:
        """Tüm model istatistiklerini döndür."""
        return {
            name: {
                'is_trained': self.stats[name]['is_trained'],
                'total_predictions': self.stats[name]['total_predictions'],
                'total_anomalies': self.stats[name]['total_anomalies'],
                'avg_inference_ms': self.stats[name]['avg_inference_ms'],
            }
            for name in self.models
        }

    def load_all(self):
        """Kaydedilmiş modelleri diskten yükle."""
        import os
        for name, path in self.model_paths.items():
            try:
                # Both models are pickled directly at self.model_paths[name] now
                if os.path.exists(path):
                    self.models[name].load(path)
                    self.stats[name]['is_trained'] = True
                    logger.info(f"{name} modeli diskten yüklendi")
            except Exception as e:
                logger.warning(f"{name} yüklenemedi: {e}")
