import numpy as np
import joblib
import logging
from pathlib import Path
from river import anomaly

logger = logging.getLogger(__name__)

class AutoencoderAnomalyDetector:
    """Lightweight 7/24 Streaming Anomaly Detector replacing the heavy Keras Autoencoder.
    Uses river.anomaly.HalfSpaceTrees to achieve <30MB RAM and <1% CPU footprint.
    """
    def __init__(self, encoding_dim=6, threshold_percentile=95):
        # Keep same constructor signature for dropping directly into existing registry
        self.encoding_dim = encoding_dim
        self.threshold_percentile = threshold_percentile
        
        # Initialize HalfSpaceTrees model
        self.model = anomaly.HalfSpaceTrees(
            n_trees=25,
            height=15,
            window_size=250,
            seed=42
        )
        self.threshold = 0.7  # Dynamic threshold determined during warmup
        self.is_trained = False
        self.feature_names = []
        
    def _array_to_dict(self, row: np.ndarray) -> dict:
        return {f"f_{i}": float(val) for i, val in enumerate(row)}

    def train(self, X: np.ndarray, epochs=8, batch_size=64):
        """Train the streaming model on a batch (e.g. startup warmup).
        Epochs and batch_size parameters are ignored for the streaming model,
        but preserved for registry interface compatibility.
        """
        logger.info(f"Streaming HalfSpaceTrees eğitimi başlıyor: {X.shape}")
        
        scores = []
        # Incremental learning loop over warmup bootstrap samples
        for row in X:
            x_dict = self._array_to_dict(row)
            # Score first, then feed the sample for learning
            score = self.model.score_one(x_dict)
            scores.append(score)
            self.model.learn_one(x_dict)
            
        # Set dynamic threshold from percentile of scores observed during warmup
        if scores:
            self.threshold = float(np.percentile(scores, self.threshold_percentile))
        else:
            self.threshold = 0.7
            
        self.is_trained = True
        logger.info(f"Streaming model eğitimi tamamlandı. Eşik (Threshold): {self.threshold:.4f}")

    def learn_one(self, row: np.ndarray):
        """Streaming update: incrementally learn from a single request vector."""
        if not self.is_trained:
            return
        x_dict = self._array_to_dict(row)
        self.model.learn_one(x_dict)

    def predict(self, X: np.ndarray):
        """Predict anomalies: -1 for anomaly, 1 for normal."""
        if not self.is_trained:
            raise ValueError("Model henüz eğitilmedi!")
            
        predictions = []
        scores = []
        for row in X:
            x_dict = self._array_to_dict(row)
            score = self.model.score_one(x_dict)
            scores.append(score)
            pred = -1 if score > self.threshold else 1
            predictions.append(pred)
            
            # Streaming learning: update model parameters dynamically on prediction path
            self.model.learn_one(x_dict)
            
        return np.array(predictions), np.array(scores)

    def predict_proba(self, X: np.ndarray) -> np.ndarray:
        """Return anomaly probabilities scaled to [0, 1]."""
        _, scores = self.predict(X)
        return np.clip(scores, 0.0, 1.0)

    def save(self, filepath: str):
        """Persist the streaming model state to disk."""
        Path(filepath).parent.mkdir(parents=True, exist_ok=True)
        data = {
            'model': self.model,
            'threshold': self.threshold,
            'encoding_dim': self.encoding_dim,
            'is_trained': self.is_trained
        }
        joblib.dump(data, filepath)
        logger.info(f"Streaming model kaydedildi: {filepath}")

    def load(self, filepath: str):
        """Load the streaming model state from disk."""
        import os
        try:
            # Prevent thread deadlocks by deleting old heavy Keras models (>1MB)
            if os.path.exists(filepath) and os.path.getsize(filepath) > 1000 * 1024:
                logger.warning(f"Old heavy model file detected ({os.path.getsize(filepath)} bytes). Deleting to prevent deadlock.")
                try:
                    os.remove(filepath)
                except Exception as del_err:
                    logger.error(f"Eski model dosyası silinemedi: {del_err}")
                self.is_trained = False
                return
                
            data = joblib.load(filepath)
            self.model = data['model']
            self.threshold = data['threshold']
            self.encoding_dim = data.get('encoding_dim', 6)
            self.is_trained = data.get('is_trained', True)
            logger.info(f"Streaming model diskten yüklendi: {filepath}")
        except Exception as e:
            logger.warning(f"Streaming model yükleme hatası ({e}). Eski dosya temizleniyor...")
            if os.path.exists(filepath):
                try:
                    os.remove(filepath)
                except:
                    pass
            self.is_trained = False