from flask import Flask, request, jsonify
from flask_cors import CORS
import numpy as np
import pandas as pd
import logging
import os
import time

from feature_engineering import FeatureEngineer
from models.isolation_forest_model import IsolationForestAnomalyDetector
from models.autoencoder_model import AutoencoderAnomalyDetector
from model_registry import ModelRegistry

logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

app = Flask(__name__)
CORS(app)

# Feature Engineering
feature_engineer = FeatureEngineer()

# Model Registry — tüm modelleri merkezi olarak yönet
registry = ModelRegistry()
registry.register('isolation_forest', IsolationForestAnomalyDetector(), 'trained_models/isolation_forest.pkl')
registry.register('autoencoder', AutoencoderAnomalyDetector(), 'trained_models/autoencoder')

# Kaydedilmiş modelleri yükle
registry.load_all()


@app.route('/health', methods=['GET'])
def health():
    return jsonify({
        'status': 'ok',
        'models': registry.get_stats()
    })


@app.route('/predict', methods=['POST'])
def predict():
    try:
        request_data = request.json.get('data', {})
        if not request_data:
            return jsonify({'error': 'Veri bulunamadı'}), 400

        # Hangi model kullanılacak?
        model_name = request.args.get('model', 'isolation_forest')

        df = pd.DataFrame([request_data])
        features = feature_engineer.extract_features(df)
        X = feature_engineer.transform(features)

        result = registry.predict(model_name, X)
        if 'error' in result:
            return jsonify(result), 400

        return jsonify(result)

    except Exception as e:
        logger.error(f"Tahmin hatası: {e}")
        return jsonify({'error': str(e)}), 500


@app.route('/predict/compare', methods=['POST'])
def predict_compare():
    """Her iki modeli de çalıştır ve sonuçları karşılaştır."""
    try:
        request_data = request.json.get('data', {})
        if not request_data:
            return jsonify({'error': 'Veri bulunamadı'}), 400

        df = pd.DataFrame([request_data])
        features = feature_engineer.extract_features(df)
        X = feature_engineer.transform(features)

        comparison = registry.predict_compare(X)
        return jsonify(comparison)

    except Exception as e:
        logger.error(f"Karşılaştırma hatası: {e}")
        return jsonify({'error': str(e)}), 500


@app.route('/train', methods=['POST'])
def train():
    """Tüm modelleri eğit."""
    try:
        np.random.seed(42)
        n_samples = 500

        sample_data = {
            'method': np.random.choice(['GET', 'POST', 'PUT'], n_samples),
            'path': [f'/api/resource/{i}' for i in range(n_samples)],
            'response_time': np.random.exponential(100, n_samples),
            'status_code': np.random.choice([200, 201, 400, 500], n_samples, p=[0.8, 0.1, 0.05, 0.05]),
            'request_size': np.random.randint(100, 5000, n_samples),
            'response_size': np.random.randint(500, 50000, n_samples)
        }
        df = pd.DataFrame(sample_data)

        features = feature_engineer.extract_features(df)
        X = feature_engineer.transform(features)

        # Tüm modelleri eğit
        os.makedirs('trained_models', exist_ok=True)
        results = registry.train_all(X, feature_names=feature_engineer.feature_names)

        return jsonify({
            'status': 'success',
            'message': 'Tüm modeller eğitildi',
            'results': results,
            'models': registry.get_stats()
        })

    except Exception as e:
        logger.error(f"Eğitim hatası: {e}")
        return jsonify({'error': str(e)}), 500


@app.route('/models/stats', methods=['GET'])
def model_stats():
    """Model istatistiklerini döndür (inference latency, anomaly count, vb.)."""
    return jsonify(registry.get_stats())


if __name__ == '__main__':
    app.run(host='0.0.0.0', port=5000, debug=True)