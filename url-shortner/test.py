
import requests

BASE_URL = "http://localhost:8000"


# CREATE

URL = f"{BASE_URL}/api/v1/urls/"

payload = {
    "long_url": "google.com",
}

response = requests.post(URL, json=payload).json()

print("Create URL Response:", response)
