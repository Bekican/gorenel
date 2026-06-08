# Production: gunicorn (Dockerfile). Yerel: python app.py
import os
import logging
import threading
import time
import numpy as np
import pandas as pd
from flask import Flask, request, jsonify
from flask_cors import CORS

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

# Model Registry — birden fazla ML modelini yöneten merkezi kayıt sınıfı
registry = ModelRegistry()
registry.register('isolation_forest', IsolationForestAnomalyDetector(), 'trained_models/isolation_forest.pkl')
registry.register('autoencoder', AutoencoderAnomalyDetector(), 'trained_models/autoencoder.pkl') # Saved directly as pickle now

# Kaydedilmiş modelleri yükle (senkron, hafif)
registry.load_all()

_training_lock = threading.Lock()
_training_state = {
    'running': False,
    'phase': 'idle',
    'last_started': None,
    'last_finished': None,
    'last_error': None,
    'last_results': None,
}

# Warmup ve bootstrap verisi limitleri (Streaming model için son derece hafif)
WARMUP_SAMPLES = int(os.environ.get('ML_WARMUP_SAMPLES', '280'))
FULL_TRAIN_SAMPLES = int(os.environ.get('ML_TRAIN_SAMPLES', '320'))
WARMUP_DELAY_SEC = float(os.environ.get('ML_WARMUP_DELAY_SEC', '1'))


def _generate_bootstrap_dataset(n_samples=320, deterministic=False):
    """Generate synthetic training data for warmup bootstrap."""
    if deterministic:
        rng = np.random.RandomState(42)
    else:
        rng = np.random.RandomState(int(time.time()) % (2**31))
    
    sample_data = {
        'method': rng.choice(['GET', 'POST', 'PUT', 'DELETE'], n_samples, p=[0.65, 0.2, 0.1, 0.05]),
        'path': [f'/api/resource/{i % 120}' for i in range(n_samples)],
        'response_time': rng.exponential(110, n_samples),
        'status_code': rng.choice([200, 201, 204, 400, 500], n_samples, p=[0.72, 0.1, 0.08, 0.06, 0.04]),
        'request_size': rng.randint(120, 9000, n_samples),
        'response_size': rng.randint(200, 60000, n_samples),
    }
    return pd.DataFrame(sample_data)


def _prepare_X(n_samples: int, deterministic: bool = False, refit_scaler: bool = False):
    df = _generate_bootstrap_dataset(n_samples=n_samples, deterministic=deterministic)
    features = feature_engineer.extract_features(df)
    if refit_scaler:
        feature_engineer.refit(features)
    X = feature_engineer.transform(features)
    return X, feature_engineer.feature_names


def run_training_job(
    *,
    n_samples: int,
    ae_epochs: int = 0,
    ae_batch: int = 0,
    reason: str = 'manual',
    force: bool = False,
):
    """
    Modelleri sırayla eğitir: önce isolation_forest, sonra streaming HalfSpaceTrees.
    """
    global _training_state
    if not _training_lock.acquire(blocking=False):
        return False, 'training_already_running'

    _training_state['running'] = True
    _training_state['phase'] = 'starting'
    _training_state['last_started'] = time.time()
    _training_state['last_error'] = None
    _training_state['last_results'] = None

    try:
        os.makedirs('trained_models', exist_ok=True)
        is_warmup = reason in ('warmup', 'ensure')
        X, feat_names = _prepare_X(n_samples, deterministic=is_warmup, refit_scaler=force)
        results = {}
        stats_before = registry.get_stats()

        train_if = force or not stats_before.get('isolation_forest', {}).get('is_trained', False)
        train_ae = force or not stats_before.get('autoencoder', {}).get('is_trained', False)

        # 1) Isolation Forest
        if train_if:
            _training_state['phase'] = 'isolation_forest'
            r = registry.train_one('isolation_forest', X, feature_names=feat_names)
            results.update(r)
        else:
            results['isolation_forest'] = {'status': 'skipped', 'reason': 'already_trained'}

        # 2) Streaming HalfSpaceTrees (autoencoder key)
        if train_ae:
            _training_state['phase'] = 'autoencoder'
            r = registry.train_one('autoencoder', X)
            results.update(r)
        else:
            results['autoencoder'] = {'status': 'skipped', 'reason': 'already_trained'}

        _training_state['last_results'] = {
            'reason': reason,
            'force': force,
            'results': results,
            'n_samples': n_samples,
        }
        _training_state['phase'] = 'done'
        logger.info('Training job finished', extra={'reason': reason, 'force': force, 'results': results})
    except Exception as e:
        logger.exception('Training job failed: %s', e)
        _training_state['last_error'] = str(e)
        _training_state['phase'] = 'error'
    finally:
        _training_state['running'] = False
        _training_state['last_finished'] = time.time()
        _training_lock.release()

    return True, 'ok'


def ensure_models_ready_background(warmup: bool = True):
    """İlk açılışta eksik modeller için bootstrap (thread içinde)."""
    stats = registry.get_stats()
    needs = any(not s.get('is_trained', False) for s in stats.values())
    if not needs:
        logger.info('Tüm modeller zaten eğitili; warmup atlandı.')
        return

    n_samples = WARMUP_SAMPLES if warmup else FULL_TRAIN_SAMPLES
    ok, msg = run_training_job(
        n_samples=n_samples,
        reason='warmup' if warmup else 'ensure',
    )
    if not ok and msg == 'training_already_running':
        logger.warning('Warmup atlandı: eğitim zaten çalışıyor')


def _warmup_thread():
    time.sleep(WARMUP_DELAY_SEC)
    logger.info('ML warmup thread: bootstrap başlıyor (arka plan)')
    ensure_models_ready_background(warmup=True)


def start_background_services():
    threading.Thread(target=_warmup_thread, name='ml-warmup', daemon=True).start()


start_background_services()


def _calculate_explainability(df_raw: pd.DataFrame, X_scaled: np.ndarray) -> dict:
    """Calculate feature contribution weights using absolute robust scaled deviation scores.
    This provides highly sound real-time Explainable AI (XAI) contribution metrics.
    """
    feature_names = feature_engineer.feature_names
    raw_features = df_raw.iloc[0].to_dict()
    scaled_values = X_scaled[0]
    
    contributions = {}
    total_contrib = 0.0
    
    for name, val in zip(feature_names, scaled_values):
        c = abs(float(val))
        # Add academic weights to high-signal security anomalies
        if name == 'is_suspicious_path' and raw_features.get(name, 0) > 0:
            c += 2.5
        if name in ('is_error', 'is_server_error') and raw_features.get(name, 0) > 0:
            c += 1.5
            
        contributions[name] = c
        total_contrib += c
        
    # Scale contributions to represent percentage weights summing to 100%
    if total_contrib > 0:
        contributions = {k: round((v / total_contrib) * 100, 1) for k, v in contributions.items()}
    else:
        contributions = {k: 0.0 for k in feature_names}
        contributions['response_time_log'] = 100.0
        
    return contributions


@app.route('/health', methods=['GET'])
def health():
    return jsonify({
        'status': 'ok',
        'models': registry.get_stats(),
        'training': {k: _training_state[k] for k in ('running', 'phase', 'last_error') if k in _training_state},
    })


@app.route('/training/status', methods=['GET'])
def training_status():
    return jsonify({
        **_training_state,
        'models': registry.get_stats(),
    })


@app.route('/predict', methods=['POST'])
def predict():
    try:
        request_data = request.json.get('data', {})
        if not request_data:
            return jsonify({'error': 'Veri bulunamadı'}), 400

        model_name = request.args.get('model', 'isolation_forest')

        df = pd.DataFrame([request_data])
        features = feature_engineer.extract_features(df)
        X = feature_engineer.transform(features)

        result = registry.predict(model_name, X)
        if 'error' in result:
            return jsonify(result), 400

        # Inject explainability insight
        result['explainability'] = _calculate_explainability(features, X)
        return jsonify(result)

    except Exception as e:
        logger.error(f"Tahmin hatası: {e}")
        return jsonify({'error': str(e)}), 500


@app.route('/predict/compare', methods=['POST'])
def predict_compare():
    """Her iki modeli de çalıştır, sonuçları karşılaştır ve XAI skorları ekle."""
    try:
        request_data = request.json.get('data', {})
        if not request_data:
            return jsonify({'error': 'Veri bulunamadı'}), 400

        df = pd.DataFrame([request_data])
        features = feature_engineer.extract_features(df)
        X = feature_engineer.transform(features)

        comparison = registry.predict_compare(X)
        
        # Inject explainability contribution weights into consensus payload
        comparison['explainability'] = _calculate_explainability(features, X)
        return jsonify(comparison)

    except Exception as e:
        logger.error(f"Karşılaştırma hatası: {e}")
        return jsonify({'error': str(e)}), 500


@app.route('/feedback', methods=['POST'])
def feedback():
    """Active learning feedback: mark a false positive as 'safe' / 'normal'
    to dynamically correct streaming model weights in real-time.
    """
    try:
        body = request.json or {}
        request_data = body.get('data', {})
        label = body.get('label', 'safe')
        
        if not request_data:
            return jsonify({'error': 'Veri bulunamadı'}), 400
            
        df = pd.DataFrame([request_data])
        features = feature_engineer.extract_features(df)
        X = feature_engineer.transform(features)
        
        model_name = 'autoencoder'
        model = registry.models.get(model_name)
        if model and model.is_trained:
            x_dict = model._array_to_dict(X[0])
            # Suppress false positives by incrementally feeding the normal state multiple times
            if label in ('safe', 'normal'):
                for _ in range(5):
                    model.model.learn_one(x_dict)
            
            # Save the updated streaming weights
            model.save(registry.model_paths[model_name])
            logger.info("Active learning feedback applied and streaming model state saved.")
            
        return jsonify({'status': 'success', 'message': 'Geri bildirim başarıyla öğrenildi ve model güncellendi.'})
    except Exception as e:
        logger.error(f"Geri bildirim işleme hatası: {e}")
        return jsonify({'error': str(e)}), 500


@app.route('/train', methods=['POST'])
def train():
    """Eğitimi manuel tetikler."""
    if _training_state['running']:
        return jsonify({
            'status': 'busy',
            'message': 'Eğitim zaten çalışıyor',
            'training': {k: _training_state[k] for k in ('phase', 'last_started')},
        }), 409

    body = request.get_json(silent=True) or {}
    async_train = body.get('async', True)
    n_samples = int(body.get('n_samples', FULL_TRAIN_SAMPLES))
    force = bool(body.get('force', False))

    def _job():
        run_training_job(
            n_samples=n_samples,
            reason='http_train',
            force=force,
        )

    if async_train:
        threading.Thread(target=_job, name='ml-train-http', daemon=True).start()
        return jsonify({
            'status': 'accepted',
            'message': 'Eğitim arka planda başlatıldı',
            'models': registry.get_stats(),
        }), 202

    ok, msg = run_training_job(
        n_samples=n_samples,
        reason='http_train',
        force=force,
    )
    if not ok:
        return jsonify({'status': 'busy', 'message': msg}), 409
    return jsonify({
        'status': 'success',
        'message': 'Eğitim tamamlandı',
        'last': _training_state.get('last_results'),
        'models': registry.get_stats(),
    })


@app.route('/models/stats', methods=['GET'])
def model_stats():
    return jsonify(registry.get_stats())


if __name__ == '__main__':
    app.run(host='0.0.0.0', port=5000, debug=True)
