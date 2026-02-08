from flask import Flask, request, jsonify
from flask_cors import CORS
import numpy as np
import pandas as pd
import logging
import os


from feature_engineering import FeatureEngineer
from models.isolation_forest_model import IsolationForestAnomalyDetector

logging.basicConfig(level=logging.INFO)
logger=logging.getLogger(__name__)

app=Flask(__name__)
CORS(app)

feature_engineer = FeatureEngineer()
model = IsolationForestAnomalyDetector()

MODEL_PATH = 'trained_models/isolation_forest.pkl'
if os.path.exists(MODEL_PATH):
    logger.info("Eğitilmiş model yüklendi.")
else:
    logger.warning("Eğitilmiş model bulunamadı. /train endpointini kullanarak eğit.")

@app.route('/health',methods=['GET'])
def health():
    return jsonify({
        'status' : 'ok',
        'model_trained' : model.is_trained
    })

@app.route('/predict',methods=['POST'])
def predict():
    try:
        request_data = request.json.get('data',{})
        if not request_data:
            return jsonify({'error': 'Veri bulunamadı'}), 400
        df = pd.DataFrame([request_data])

        features = feature_engineer.extract_features(df)
        X = feature_engineer.transform(features)

        predictions , scores = model.predict(X)
        probabilities = model.predict_proba(X)

        is_anomaly = bool(predictions[0] == -1)

        return jsonify({
            'is_anomaly': is_anomaly,
            'anomaly_score': float(probabilities[0]),
            'prediction': 'anomaly' if is_anomaly else 'normal'
        })
    
    except Exception as e:
        logger.error(f"Tahmin hatası: {e}")
        return jsonify({'error': str(e)}), 500

@app.route('/train', methods=['POST'])
def train():
    """
    Modeli eğit (örnek veri ile)
    Production'da Redis'ten veri çekilir.
    """
    try:
        # Örnek eğitim verisi oluştur (Gerçekte data_collector kullanılır)
        np.random.seed(42)
        sample_data = {
            'method': np.random.choice(['GET', 'POST', 'PUT'], 500),
            'path': [f'/api/resource/{i}' for i in range(500)],
            'response_time': np.random.exponential(100, 500),
            'status_code': np.random.choice([200, 201, 400, 500], 500, p=[0.8, 0.1, 0.05, 0.05]),
            'request_size': np.random.randint(100, 5000, 500),
            'response_size': np.random.randint(500, 50000, 500)
        }
        df = pd.DataFrame(sample_data)
        
        # Özellik çıkar
        features = feature_engineer.extract_features(df)
        X = feature_engineer.transform(features)
        
        # Eğit
        model.train(X, feature_names=feature_engineer.feature_names)
        
        # Kaydet
        os.makedirs('trained_models', exist_ok=True)
        model.save(MODEL_PATH)
        
        return jsonify({
            'status': 'success',
            'message': 'Model eğitildi ve kaydedildi',
            'stats': model.training_stats
        })
    
    except Exception as e:
        logger.error(f"Eğitim hatası: {e}")
        return jsonify({'error': str(e)}), 500

if __name__ == '__main__':
    app.run(host='0.0.0.0', port=5000, debug=True)