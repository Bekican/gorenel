import numpy as np
import tensorflow as tf
from tensorflow import keras
import joblib
import logging
from pathlib import Path


logger = logging.getLogger(__name__)

class AutoEncoderAnomalyDetector:
    def __init__(self,encoding_dim=8,threshold_percentile=95):
        self.encoding_dim = encoding_dim
        self.threshold_percentile = threshold_percentile
        self.model = None
        self.is_trained = False
        self.input_dim = None
    
    def _build_model(self,input_dim):
        self.input_dim=input_dim

        #encoder
        inputs = keras.Input(shape=(input_dim))
        x = keras.layers.Dense(32,activation='relu')(inputs)
        x = keras.layers.Dense(16,activation='relu')(x)
        encoded = keras.layers.Dense(self.encoding_dim,activation='relu')(x)

        #decoder
        x = keras.layers.Dense(16, activation='relu')(encoded)
        x = keras.layers.Dense(32, activation='relu')(x)
        decoded = keras.layers.Dense(input_dim,activation='linear')(x)


        self.model = keras.Model(inputs,decoded)
        self.model.compile(optimizer='adam',loss='mse')

    def train(self,X: np.ndarray,epochs=50,batch_size=32):
        logger.info(f"Autoencoder eğitimi başlıyor:{X.shape}")

        self._build_model(X.shape[1])

        
        
        



        