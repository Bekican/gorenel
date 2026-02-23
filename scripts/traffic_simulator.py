import requests
import time
import random

SUBDOMAIN = "x0h213ne"
BASE_URL = "http://localhost:8085"
HOST = f"{SUBDOMAIN}.gorenel.io"

PATHS = [
    "/",
    "/api/v1/users",
    "/api/v1/products",
    "/login",
    "/dashboard",
    "/search?q=gorenel",
    "/static/img/logo.png"
]

def send_request():
    path = random.choice(PATHS)
    method = "GET"
    if random.random() > 0.8:
        method = "POST"
    
    headers = {
        "Host": HOST,
        "User-Agent": "FakeTrafficSimulator/1.0"
    }
    
    try:
        if method == "GET":
            print(f"[*] GET {path}")
            requests.get(f"{BASE_URL}{path}", headers=headers, timeout=5)
        else:
            print(f"[*] POST {path}")
            requests.post(f"{BASE_URL}{path}", headers=headers, json={"test": "data"}, timeout=5)
    except Exception as e:
        print(f"[!] Error: {e}")

def simulate_anomaly():
    # Anomaly: Rapid requests to non-existent paths
    print("[!] SIMULATING ANOMALY: Scanning paths...")
    for _ in range(10):
        path = f"/admin/{random.randint(1000, 9999)}"
        headers = {"Host": HOST}
        try:
            requests.get(f"{BASE_URL}{path}", headers=headers, timeout=5)
        except:
            pass
        time.sleep(0.1)

if __name__ == "__main__":
    print(f"Starting traffic simulation for {HOST}...")
    try:
        while True:
            send_request()
            time.sleep(random.uniform(0.5, 2.0))
            
            if random.random() > 0.95:
                simulate_anomaly()
    except KeyboardInterrupt:
        print("Stopped.")
