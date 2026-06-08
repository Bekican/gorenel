import requests
import json
import time

# Gorenel Tunnel URL (Local Subdomain)
SUBDOMAIN = "c93xicq5" # Bu değeri çalışma sırasında güncelleyebilirim
BASE_URL = "http://127.0.0.1:8085"

def simulate_openai():
    print(f"--- Simulating OpenAI Request to {BASE_URL} ---")
    
    payload = {
        "model": "gpt-4-turbo",
        "messages": [
            {"role": "system", "content": "You are a helpful assistant."},
            {"role": "user", "content": "Explain Gorenel in 2 sentences."}
        ]
    }
    
    headers = {
        "Content-Type": "application/json",
        "Authorization": "Bearer sk-fake-key",
        "Host": f"{SUBDOMAIN}.gorenel.io" # Gorenel uses this to resolve tunnel
    }
    
    try:
        # We send it to our tunnel which forwards it to the local port (handled by Gorenel Proxy)
        # Note: In real world, user targets api.openai.com through the tunnel
        # Here we target the tunnel path that matches our AI detector
        response = requests.post(f"{BASE_URL}/v1/chat/completions", json=payload, headers=headers)
        print(f"Status: {response.status_code}")
        print("Response received (Simulated)")
    except Exception as e:
        print(f"Error: {e}")

def simulate_anthropic():
    print(f"\n--- Simulating Anthropic Request to {BASE_URL} ---")
    
    payload = {
        "model": "claude-3-opus-20240229",
        "messages": [
            {"role": "user", "content": "What is chaos engineering?"}
        ]
    }
    
    headers = {
        "Content-Type": "application/json",
        "x-api-key": "sk-ant-fake",
        "Host": "api.anthropic.com"
    }
    
    try:
        response = requests.post(f"{BASE_URL}/v1/messages", json=payload, headers=headers)
        print(f"Status: {response.status_code}")
    except Exception as e:
        print(f"Error: {e}")

if __name__ == "__main__":
    simulate_openai()
    time.sleep(1)
    simulate_anthropic()
