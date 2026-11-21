"""Docker 迁移练习：单进程计数服务，仅使用 Python 标准库。"""
import json
import logging
import os
from http.server import BaseHTTPRequestHandler, HTTPServer
from pathlib import Path


class Handler(BaseHTTPRequestHandler):
    def reply(self, status, body):
        payload = json.dumps(body).encode("utf-8")
        self.send_response(status)
        self.send_header("Content-Type", "application/json")
        self.send_header("Content-Length", str(len(payload)))
        self.end_headers()
        self.wfile.write(payload)

    def do_GET(self):
        if self.path == "/healthz":
           self.reply(200, {"status": "ok", "version": "lab-v2"})
        elif self.path == "/visits":
            try:
                count = int(self.server.counter_file.read_text())
                self.reply(200, {"visits": count})
            except (OSError, ValueError):
                logging.exception("Cannot read counter")
                self.reply(500, {"error": "counter read failed"})
        else:
            self.reply(404, {"error": "not found"})

    def do_POST(self):
        if self.path != "/visits":
            self.reply(404, {"error": "not found"})
            return
        try:
            path = self.server.counter_file
            count = int(path.read_text()) + 1
            temporary = path.with_suffix(".tmp")
            temporary.write_text(str(count))
            temporary.replace(path)
            self.reply(200, {"visits": count})
        except (OSError, ValueError):
            logging.exception("Cannot update counter")
            self.reply(500, {"error": "counter update failed"})


def main():
    logging.basicConfig(level=logging.INFO, format="%(levelname)s %(message)s")
    port = int(os.environ["APP_PORT"])
    data_dir = Path(os.environ["DATA_DIR"])
    data_dir.mkdir(parents=True, exist_ok=True)
    counter_file = data_dir / "visits.txt"
    if not counter_file.exists():
        counter_file.write_text("0")

    server = HTTPServer(("0.0.0.0", port), Handler)
    server.counter_file = counter_file
    logging.info("Listening on 0.0.0.0:%s; counter=%s", port, counter_file)
    try:
        server.serve_forever()
    finally:
        server.server_close()


if __name__ == "__main__":
    main()
