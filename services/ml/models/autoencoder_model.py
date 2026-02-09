import numpy as np
import tensorflow as tf
from tensorflow import keras
import joblib
import logging
from pathlib import Path

logger = logging.getLogger(__name__)

class AutoencoderAnomalyDetector:
    def __init__(self, encoding_dim=8, threshold_percentile=95):
        self.encoding_dim = encoding_dim
        self.threshold_percentile = threshold_percentile
        self.model = None
        self.threshold = None
        self.is_trained = False
        self.input_dim = None
    
    def _build_model(self, input_dim):
        self.input_dim = input_dim
        
        # Encoder
        inputs = keras.Input(shape=(input_dim,))
        x = keras.layers.Dense(32, activation='relu')(inputs)
        x = keras.layers.Dense(16, activation='relu')(x)
        encoded = keras.layers.Dense(self.encoding_dim, activation='relu')(x)
        
        # Decoder
        x = keras.layers.Dense(16, activation='relu')(encoded)
        x = keras.layers.Dense(32, activation='relu')(x)
        decoded = keras.layers.Dense(input_dim, activation='linear')(x)
        
        self.model = keras.Model(inputs, decoded)
        self.model.compile(optimizer='adam', loss='mse')
    
    def train(self, X: np.ndarray, epochs=50, batch_size=32):
        logger.info(f"Autoencoder eğitimi başlıyor: {X.shape}")
        
        self._build_model(X.shape[1])
        
        self.model.fit(X, X, epochs=epochs, batch_size=batch_size, 
                       validation_split=0.1, verbose=0)
        
        # Threshold hesapla
        reconstructions = self.model.predict(X, verbose=0)
        mse = np.mean(np.power(X - reconstructions, 2), axis=1)
        self.threshold = np.percentile(mse, self.threshold_percentile)
        
        self.is_trained = True
        logger.info(f"Eğitim tamamlandı. Threshold: {self.threshold:.4f}")
    
    def predict(self, X: np.ndarray):
        if not self.is_trained:
            raise ValueError("Model eğitilmedi!")
        
        reconstructions = self.model.predict(X, verbose=0)
        mse = np.mean(np.power(X - reconstructions, 2), axis=1)
        
        predictions = np.where(mse > self.threshold, -1, 1)
        return predictions, mse
    
    def predict_proba(self, X: np.ndarray) -> np.ndarray:
        _, mse = self.predict(X)
        # Normalize et
        proba = np.clip(mse / (self.threshold * 2), 0, 1)
        return proba
    
    def save(self, filepath: str):
        Path(filepath).parent.mkdir(parents=True, exist_ok=True)
        self.model.save(filepath + '_keras')
        joblib.dump({
            'threshold': self.threshold,
            'input_dim': self.input_dim,
            'encoding_dim': self.encoding_dim
        }, filepath + '_meta.pkl')
    
    def load(self, filepath: str):
        self.model = keras.models.load_model(filepath + '_keras')
        meta = joblib.load(filepath + '_meta.pkl')
        self.threshold = meta['threshold']
        self.input_dim = meta['input_dim']
        self.is_trained = True