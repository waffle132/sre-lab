from urllib.request import Request, urlopen
import json
import sys
from urllib.error import URLError
import http.client
import logging
import argparse
from urllib.parse import urlsplit

logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s %(levelname)s %(message)s")
def main():
    parser = argparse.ArgumentParser(description="Business Probe Script")
    parser.add_argument("--target-url", type=str, required=True, help="The target URL to shorten and check redirect")
    parser.add_argument("--base-url", type=str, default="http://localhost:8080", help="The base URL of the shortening service")
    parser.add_argument("--timeout", type=int, default=3, help="Timeout for HTTP requests in seconds")
    args = parser.parse_args()
    if args.timeout <= 0:
        parser.error("--timeout must be a positive integer")
    base_url = args.base_url.rstrip('/')
    address = urlsplit(base_url)
    if address.scheme != "http" or not address.hostname:
        parser.error("--base-url must be an HTTP URL with a hostname")   
    target_url = args.target_url
    payload = json.dumps({"URL": target_url}).encode('utf-8')  # 将字典转换为JSON并编码为字节
    request = Request(f"{base_url}/shorten", 
                    data=payload,
                    headers={"Content-Type": "application/json"},
                    method="POST")
    #组装request
    try:
        with urlopen(request,timeout=args.timeout) as response:
            status = response.status
            response_text = response.read().decode('utf-8')
    except (URLError, TimeoutError) as e:
        logging.error("Create request failed: %s", e)
        sys.exit(1)
    #测试urlopen
    if status != 200:
        logging.error("Create failed: unexpected status=%s", status)
        sys.exit(1)

    try:
        result = json.loads(response_text)
    except json.JSONDecodeError as e:
        logging.error("Create failed: invalid JSON: %s", e)
        sys.exit(1)
    #测试json格式合法
    if not isinstance(result, dict):
        logging.error("Create failed: response is not a JSON object")
        sys.exit(1)
    #判断result为字典    
    code = result.get("code")

    if not isinstance(code, str) or not code.strip():
        logging.error("Create failed: missing or invalid code")
        sys.exit(1)
    logging.info("Create passed: code=%s", code)
    check_redirect(code, target_url, base_url, args.timeout)



def check_redirect(code, target_url, base_url, timeout):
    address = urlsplit(base_url)
    connection = http.client.HTTPConnection(address.hostname, address.port, timeout = timeout)
    try:
        connection.request("GET", f"/r/{code}")
        response = connection.getresponse()
        location = response.getheader("Location")
        if response.status == 302 and location == target_url:
            logging.info("Redirect passed: status=%s, location=%s",
                         response.status, location)
        else:
            logging.error("Redirect failed: status=%s, location=%s",
                          response.status, location)
            sys.exit(1)
    except (OSError, http.client.HTTPException) as e:
        logging.error("Redirect request failed: %s", e)
        sys.exit(1)
    finally:
        connection.close()

if __name__ == "__main__":
    main()

