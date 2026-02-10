import requests
import time
import threading
import statistics
import sys
import argparse

def stress_predict(url, requests_count, results, errors):
    payload = {
        "data": {
            "method": "POST",
            "path": "/api/test",
            "response_time": 150,
            "status_code": 200,
            "request_size": 1024,
            "response_size": 2048
        }
    }
    
    for _ in range(requests_count):
        try:
            start = time.time()
            response = requests.post(f"{url}/predict/compare", json=payload, timeout=10)
            end = time.time()
            
            if response.status_code == 200:
                results.append((end - start) * 1000)
            else:
                errors.append(f"Status {response.status_code}")
        except Exception as e:
            errors.append(str(e))

def run_load_test(url, total_requests, concurrency):
    print(f"--- ML Service Load Test ---")
    print(f"Target: {url}")
    print(f"Requests: {total_requests}, Concurrency: {concurrency}")
    
    threads = []
    results = []
    errors = []
    req_per_thread = total_requests // concurrency
    
    start_time = time.time()
    
    for i in range(concurrency):
        t = threading.Thread(target=stress_predict, args=(url, req_per_thread, results, errors))
        threads.append(t)
        t.start()
        
    for t in threads:
        t.join()
        
    total_time = time.time() - start_time
    
    print(f"\n--- Results ---")
    print(f"Total Time: {total_time:.2f}s")
    print(f"Successful Requests: {len(results)}")
    print(f"Errors: {len(errors)}")
    
    if results:
        print(f"Avg Latency: {statistics.mean(results):.2f}ms")
        print(f"P95 Latency: {statistics.quantiles(results, n=20)[18]:.2f}ms")
        print(f"Throughput: {len(results) / total_time:.2f} req/s")
    
    if errors:
        print(f"\nSample Errors: {list(set(errors))[:3]}")

if __name__ == "__main__":
    parser = argparse.ArgumentParser()
    parser.add_argument("--url", default="http://localhost:5000")
    parser.add_argument("--requests", type=int, default=100)
    parser.add_argument("--concurrency", type=int, default=5)
    args = parser.parse_args()
    
    run_load_test(args.url, args.requests, args.concurrency)
