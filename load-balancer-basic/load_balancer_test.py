import requests
import time, json
from concurrent.futures import ThreadPoolExecutor, as_completed

URL = "http://localhost:9090"


def send_request(url):
    try:
        start_time = time.time()
        response = requests.get(url)
        end_time = time.time()
        res = response.json()
        res["response_time"] = round((end_time - start_time) * 1000)
        res["status_code"] = response.status_code
        return res
    except requests.RequestException as e:
        return f"Request failed: {e}"


def load_test(url, num_requests):
    results = []
    with ThreadPoolExecutor(max_workers=num_requests) as executor:
        futures = [executor.submit(send_request, url) for _ in range(num_requests)]
        for future in as_completed(futures):
            r = future.result()
            # print(r)
            results.append(r)
        return results


if __name__ == "__main__":
    num_requests = 100  # Number of parallel requests
    start_time = time.time()
    output = load_test(URL, num_requests)
    # print(output)    
    with open("loadtest_results.json", "w") as f:
        json.dump(output, f, indent=4)
    end_time = time.time()
    print(f"Load test completed in {end_time - start_time} seconds")
