from urllib.request import  urlopen
from urllib.error import URLError
import sys
import argparse
import logging

def is_healthy(status, body):
    if status == 200 and body.strip() == "OK":
        return True
    else:
        return False  
def main():
    logging.basicConfig(
        level=logging.INFO,
        format="%(asctime)s %(levelname)s %(message)s"
    )
    parser = argparse.ArgumentParser()
    parser.add_argument("--url", required=True)
    parser.add_argument("--timeout", type=float, default=3)
    args = parser.parse_args()
    if args.timeout <= 0:
        parser.error("Timeout must be a positive number")    
    try:
        with urlopen(args.url, timeout=args.timeout) as response:
            body = response.read().decode('utf-8')
            if is_healthy(response.status, body):
                logging.info("Health check passed")
                sys.exit(0)
            else:
                logging.error("Health check failed")
                sys.exit(1)
    except (URLError, TimeoutError) as e:
        logging.error(f"Health check failed: {e}")
        sys.exit(1)
if __name__ == "__main__":
    main()    