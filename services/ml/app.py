# Production: gunicorn (Dockerfile). Yerel: python app.py
from flask import Flask, request, jsonify
from flask_cors import CORS
import numpy as np
import pandas as pd
import logging
import os
import threading
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

# Warmup: daha küçük veri + daha az epoch (ilk açılışta worker bloklamadan tamamlanır)
WARMUP_SAMPLES = int(os.environ.get('ML_WARMUP_SAMPLES', '280'))
WARMUP_AE_EPOCHS = int(os.environ.get('ML_WARMUP_AE_EPOCHS', '6'))
WARMUP_AE_BATCH = int(os.environ.get('ML_WARMUP_AE_BATCH', '32'))
# Tam /train ve periyodik yeniden eğitim
FULL_TRAIN_SAMPLES = int(os.environ.get('ML_TRAIN_SAMPLES', '400'))
FULL_AE_EPOCHS = int(os.environ.get('ML_FULL_AE_EPOCHS', '8'))
FULL_AE_BATCH = int(os.environ.get('ML_FULL_AE_BATCH', '64'))
WARMUP_DELAY_SEC = float(os.environ.get('ML_WARMUP_DELAY_SEC', '2'))
RETRAIN_INTERVAL_HOURS = float(os.environ.get('ML_RETRAIN_INTERVAL_HOURS', '24'))


def _generate_bootstrap_dataset(n_samples=320):
    np.random.seed(42)
    sample_data = {
        'method': np.random.choice(['GET', 'POST', 'PUT', 'DELETE'], n_samples, p=[0.65, 0.2, 0.1, 0.05]),
        'path': [f'/api/resource/{i % 120}' for i in range(n_samples)],
        'response_time': np.random.exponential(110, n_samples),
        'status_code': np.random.choice([200, 201, 204, 400, 500], n_samples, p=[0.72, 0.1, 0.08, 0.06, 0.04]),
        'request_size': np.random.randint(120, 9000, n_samples),
        'response_size': np.random.randint(200, 60000, n_samples),
    }
    return pd.DataFrame(sample_data)


def _prepare_X(n_samples: int):
    df = _generate_bootstrap_dataset(n_samples=n_samples)
    features = feature_engineer.extract_features(df)
    X = feature_engineer.transform(features)
    return X, feature_engineer.feature_names


def run_training_job(
    *,
    n_samples: int,
    ae_epochs: int,
    ae_batch: int,
    reason: str = 'manual',
    force: bool = False,
):
    """
    Modelleri sırayla eğitir: önce isolation_forest, sonra autoencoder.
    force=False iken yalnızca diskten yüklenmemiş / is_trained=False olanlar eğitilir.
    force=True (periyodik / manuel tam yenileme) her iki modeli yeniden eğitir.
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
        X, feat_names = _prepare_X(n_samples)
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

        # 2) Autoencoder — prod hedefi: ilk kurulumda mutlaka eğitilmiş olsun
        if train_ae:
            _training_state['phase'] = 'autoencoder'
            r = registry.train_one(
                'autoencoder',
                X,
                epochs=ae_epochs,
                batch_size=ae_batch,
            )
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
    ae_ep = WARMUP_AE_EPOCHS if warmup else FULL_AE_EPOCHS
    ae_bs = WARMUP_AE_BATCH if warmup else FULL_AE_BATCH
    ok, msg = run_training_job(
        n_samples=n_samples,
        ae_epochs=ae_ep,
        ae_batch=ae_bs,
        reason='warmup' if warmup else 'ensure',
    )
    if not ok and msg == 'training_already_running':
        logger.warning('Warmup atlandı: eğitim zaten çalışıyor')


def _warmup_thread():
    time.sleep(WARMUP_DELAY_SEC)
    logger.info('ML warmup thread: bootstrap başlıyor (arka plan)')
    ensure_models_ready_background(warmup=True)


def _periodic_retrain_loop():
    interval_sec = max(3600.0, RETRAIN_INTERVAL_HOURS * 3600.0)
    while True:
        time.sleep(interval_sec)
        logger.info('Periyodik yeniden eğitim tetiklendi (force=True)')
        run_training_job(
            n_samples=FULL_TRAIN_SAMPLES,
            ae_epochs=FULL_AE_EPOCHS,
            ae_batch=FULL_AE_BATCH,
            reason='periodic',
            force=True,
        )


def start_background_services():
    threading.Thread(target=_warmup_thread, name='ml-warmup', daemon=True).start()
    threading.Thread(target=_periodic_retrain_loop, name='ml-retrain', daemon=True).start()


start_background_services()


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
    """Modelleri arka planda eğitir (worker timeout olmaması için 202 Accepted)."""
    if _training_state['running']:
        return jsonify({
            'status': 'busy',
            'message': 'Eğitim zaten çalışıyor',
            'training': {k: _training_state[k] for k in ('phase', 'last_started')},
        }), 409

    body = request.get_json(silent=True) or {}
    async_train = body.get('async', True)
    n_samples = int(body.get('n_samples', FULL_TRAIN_SAMPLES))
    ae_epochs = int(body.get('ae_epochs', FULL_AE_EPOCHS))
    ae_batch = int(body.get('ae_batch', FULL_AE_BATCH))
    force = bool(body.get('force', False))

    def _job():
        run_training_job(
            n_samples=n_samples,
            ae_epochs=ae_epochs,
            ae_batch=ae_batch,
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

    # Senkron (test veya özel istemci); gunicorn timeout yeterli olmalı
    _job()
    return jsonify({
        'status': 'success',
        'message': 'Eğitim tamamlandı',
        'last': _training_state.get('last_results'),
        'models': registry.get_stats(),
    })


@app.route('/models/stats', methods=['GET'])
def model_stats():
    """Model istatistiklerini döndür (inference latency, anomaly count, vb.)."""
    return jsonify(registry.get_stats())


if __name__ == '__main__':
    app.run(host='0.0.0.0', port=5000, debug=True)
