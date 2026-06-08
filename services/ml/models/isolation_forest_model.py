import numpy as np
from sklearn.ensemble import IsolationForest
import joblib
import logging
from pathlib import Path

logger = logging.getLogger(__name__)

class IsolationForestAnomalyDetector:
    """Isolation Forest tabanlı anomali tespit modeli"""
    
    def __init__(self, contamination=0.1, n_estimators=100, random_state=42):
        # contamination: Verinin yüzde kaçının anomali olduğu (0.1 = %10)
        # n_estimators: Kaç adet karar ağacı kullanılacağı
        self.model = IsolationForest(
            contamination=contamination,
            n_estimators=n_estimators,
            max_samples=256,
            n_jobs=-1,  # Tüm CPU çekirdeklerini kullan
            random_state=random_state
        )
        self.is_trained = False
        self.feature_names = []
        self.training_stats = {}
        self.score_threshold = None
        self.score_std = 1.0
    
    def train(self, X: np.ndarray, feature_names: list = None):
        """Modeli eğit"""
        logger.info(f"Eğitim başlıyor: {X.shape[0]} örnek, {X.shape[1]} özellik")
        
        self.feature_names = feature_names or [f"feature_{i}" for i in range(X.shape[1])]
        self.model.fit(X)
        self.is_trained = True
        
        # Eğitim sonrası istatistikler
        predictions = self.model.predict(X)
        anomaly_count = np.sum(predictions == -1)
        train_scores = self.model.score_samples(X)
        self.score_threshold = float(np.quantile(train_scores, self.model.contamination))
        self.score_std = float(np.std(train_scores))
        if self.score_std <= 1e-9:
            self.score_std = 1.0
        
        self.training_stats = {
            'n_samples': X.shape[0],
            'n_features': X.shape[1],
            'n_anomalies': int(anomaly_count),
            'anomaly_ratio': float(anomaly_count / X.shape[0]),
            'score_threshold': self.score_threshold,
        }
        logger.info(f"Eğitim tamamlandı. {anomaly_count} anomali tespit edildi.")
    
    def predict(self, X: np.ndarray):
        """Tahmin yap: -1 (anomali) veya 1 (normal)"""
        if not self.is_trained:
            raise ValueError("Model henüz eğitilmedi!")
        
        predictions = self.model.predict(X)
        scores = self.model.score_samples(X)
        return predictions, scores
    
    def predict_proba(self, X: np.ndarray) -> np.ndarray:
        """Anomali olasılıklarını [0, 1] aralığında döndür.

        IsolationForest'ta daha düşük score daha anormal demektir.
        Tek örnek tahminde min=max problemi olmaması için batch min/max yerine
        eğitimde hesaplanan eşik etrafında sigmoid dönüşüm uygulanır.
        """
        _, scores = self.predict(X)
        threshold = self.score_threshold if self.score_threshold is not None else float(np.median(scores))
        scale = max(self.score_std * 0.75, 1e-6)
        z = (threshold - scores) / scale
        probs = 1.0 / (1.0 + np.exp(-z))
        return np.clip(probs, 0.0, 1.0)
    
    def save(self, filepath: str):
        """Modeli diske kaydet"""
        Path(filepath).parent.mkdir(parents=True, exist_ok=True)
        data = {
            'model': self.model,
            'feature_names': self.feature_names,
            'is_trained': self.is_trained,
            'training_stats': self.training_stats,
            'score_threshold': self.score_threshold,
            'score_std': self.score_std,
        }
        joblib.dump(data, filepath)
        logger.info(f"Model kaydedildi: {filepath}")
    
    def load(self, filepath: str):
        """Modeli diskten yükle"""
        data = joblib.load(filepath)
        self.model = data['model']
        self.feature_names = data['feature_names']
        self.is_trained = data['is_trained']
        self.training_stats = data.get('training_stats', {})
        self.score_threshold = data.get('score_threshold')
        self.score_std = data.get('score_std', 1.0)
        logger.info(f"Model yüklendi: {filepath}")

if __name__ == '__main__':
    # Basit test
    print("Isolation Forest modeli hazır.")
    detector = IsolationForestAnomalyDetector()
    # Rastgele test verisi
    X_test = np.random.randn(100, 10)
    detector.train(X_test)
    preds, scores = detector.predict(X_test[:5])
    print(f"İlk 5 tahmin: {preds}")