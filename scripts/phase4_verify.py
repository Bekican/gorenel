import requests
import json
import time

MONITORING_URL = "http://127.0.0.1:9091"
PROXY_URL = "http://localhost:8085"
MOCK_PATH = "/api/v1/mock-test"
MOCK_BODY = {"status": "success", "source": "gorenel-edge-mock", "feature": "phase-4"}

def verify_phase4():
    print("--- Gorenel Phase 4 Verification: Response Mocking ---")
    
    # 0. Get an active subdomain
    print("Fetching active tunnels...")
    resp = requests.get(f"{MONITORING_URL}/api/tunnels")
    tunnels = resp.json()
    if not tunnels:
        print("❌ No active tunnels found. Please ensure the client is running.")
        return
    
    # Get the first active subdomain (it's a map in newer versions, or list in some)
    # Based on tunnelsHandlerFunc, it returns a map[string]TunnelStats
    active_subdomain = list(tunnels.keys())[0]
    print(f"Using active subdomain: {active_subdomain}")

    # 1. Add Modification Rule
    rule = {
        "path_pattern": MOCK_PATH,
        "mock_body": json.dumps(MOCK_BODY),
        "status_code": 201
    }
    
    print(f"Adding mock rule for {MOCK_PATH}...")
    resp = requests.post(f"{MONITORING_URL}/api/inspector/rules", json=rule)
    if resp.status_code != 201:
        print(f"FAILED to add rule: {resp.status_code} {resp.text}")
        return
    
    rule_data = resp.json()
    rule_id = rule_data["id"]
    print(f"Rule added with ID: {rule_id}")

    # 2. Test the mock via Proxy
    print(f"Sending request to {PROXY_URL}{MOCK_PATH} via Proxy...")
    # Use the active subdomain for the Host header
    headers = {"Host": f"{active_subdomain}.gorenel.io"} 
    
    try:
        resp = requests.get(f"{PROXY_URL}{MOCK_PATH}", headers=headers)
        
        print(f"Response Status: {resp.status_code}")
        print(f"Response Headers: {dict(resp.headers)}")
        print(f"Response Body: {resp.text}")

        # 3. Validations
        success = True
        if resp.status_code != 201:
            print(f"❌ Failure: Expected status code 201, got {resp.status_code}")
            success = False
        
        if resp.headers.get("X-Gorenel-Morph") != "Active":
            print("❌ Failure: X-Gorenel-Morph header missing or incorrect")
            success = False
            
        try:
            if resp.json() != MOCK_BODY:
                print("❌ Failure: Mock body mismatch")
                success = False
        except:
            print("❌ Failure: Response body is not valid JSON")
            success = False

        if success:
            print("\n✅ PHASE 4 VERIFIED: Response Mocking is active and functioning at the edge!")
        else:
            print("\n❌ PHASE 4 VERIFICATION FAILED.")

    except Exception as e:
        print(f"Error during request: {e}")
    finally:
        # Cleanup
        print(f"Cleaning up rule {rule_id}...")
        requests.delete(f"{MONITORING_URL}/api/inspector/rules?id={rule_id}")

if __name__ == "__main__":
    verify_phase4()
