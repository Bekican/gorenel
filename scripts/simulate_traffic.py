import requests
import time
import random
import sys

# Configuration
PROXY_URL = "http://localhost:8080"
MONITORING_URL = "http://localhost:9090"
SUBDOMAINS = ["demo-app", "test-env", "staging"]

# Patterns
NORMAL_PATHS = ["/", "/api/v1/users", "/dashboard", "/health", "/login", "/products"]
ANOMALY_PATHS = ["/etc/passwd", "/admin", "/.env", "/wp-admin", "/shell.php", "/api/v1/debug"]
USER_AGENTS = [
    "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
    "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7)",
    "curl/7.64.1",
    "Insecure-Robot-Scanner/1.0"
]

def send_request(path, method="GET", is_anomaly=False):
    subdomain = random.choice(SUBDOMAINS)
    headers = {
        "Host": f"{subdomain}.gorenel.io",
        "User-Agent": random.choice(USER_AGENTS) if not is_anomaly else "Malicious-Agent/2.0"
    }
    
    url = f"{PROXY_URL}{path}"
    
    try:
        start_time = time.time()
        if method == "GET":
            resp = requests.get(url, headers=headers, timeout=5)
        else:
            resp = requests.post(url, headers=headers, json={"data": "test"}, timeout=5)
        
        duration = (time.time() - start_time) * 1000
        status = "ANOMALY" if is_anomaly else "NORMAL"
        print(f"[{status}] {method} {headers['Host']}{path} -> {resp.status_code} ({duration:.2f}ms)")
    except Exception as e:
        print(f"[ERROR] Request failed: {e}")

def main():
    print("🚀 Gorenel Traffic Simulator started...")
    print(f"Target Proxy: {PROXY_URL}")
    print("Sending traffic (Ctrl+C to stop)...")
    
    try:
        while True:
            # 80% normal traffic, 20% anomaly
            is_anomaly = random.random() < 0.2
            
            if is_anomaly:
                path = random.choice(ANOMALY_PATHS)
                # Some anomalies might have long response times simulated or specific methods
                send_request(path, method=random.choice(["GET", "POST"]), is_anomaly=True)
            else:
                path = random.choice(NORMAL_PATHS)
                send_request(path)
                
            time.sleep(random.uniform(0.5, 2.0))
    except KeyboardInterrupt:
        print("\nStopping simulator...")

if __name__ == "__main__":
    main()
