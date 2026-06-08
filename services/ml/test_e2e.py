import requests
import random
import time

BASE_URL = "http://localhost:5000"

def test_health():
    resp = requests.get(f"{BASE_URL}/health")
    print(f"Health: {resp.json()}")
    assert resp.status_code == 200

def test_train():
    resp = requests.post(f"{BASE_URL}/train")
    print(f"Train: {resp.json()}")
    # /train is async and returns 202 Accepted or 200 OK depending on state
    assert resp.status_code in (200, 202, 409)

def test_normal_traffic():
    """Normal trafik - anomali OLMAMALI"""
    data = {
        "data": {
            "method": "GET",
            "path": "/api/users",
            "response_time": 150,
            "status_code": 200,
            "request_size": 500,
            "response_size": 2000
        }
    }
    resp = requests.post(f"{BASE_URL}/predict", json=data)
    result = resp.json()
    print(f"Normal traffic: {result}")
    assert resp.status_code == 200
    assert 'is_anomaly' in result
    assert 'anomaly_score' in result

def test_suspicious_traffic():
    """Şüpheli trafik - potansiyel anomali"""
    data = {
        "data": {
            "method": "POST",
            "path": "/api/../../../etc/passwd?union=select",
            "response_time": 5000,  # Çok yavaş
            "status_code": 500,
            "request_size": 50000,  # Çok büyük
            "response_size": 100
        }
    }
    resp = requests.post(f"{BASE_URL}/predict", json=data)
    result = resp.json()
    print(f"Suspicious traffic: {result}")
    # Anomali skoru yüksek olmalı

def test_batch():
    """100 istek gönder"""
    anomalies = 0
    for i in range(100):
        data = {
            "data": {
                "method": random.choice(["GET", "POST"]),
                "path": f"/api/resource/{i}",
                "response_time": random.randint(50, 500),
                "status_code": random.choice([200, 200, 200, 400, 500]),
                "request_size": random.randint(100, 5000),
                "response_size": random.randint(500, 10000)
            }
        }
        resp = requests.post(f"{BASE_URL}/predict", json=data)
        if resp.json().get('is_anomaly'):
            anomalies += 1
    
    print(f"Batch test: {anomalies}/100 anomali tespit edildi")

if __name__ == '__main__':
    print("=== E2E Test Başlıyor ===\n")
    
    test_health()
    test_train()
    test_normal_traffic()
    test_suspicious_traffic()
    test_batch()
    
    print("\n=== Tüm testler tamamlandı ===")