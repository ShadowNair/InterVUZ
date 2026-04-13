from __future__ import annotations

import argparse
import http.server
import io
import socketserver
from pathlib import Path


FRONTEND_DIR = Path(__file__).resolve().parent
REPO_ROOT = FRONTEND_DIR.parent
SVG_DIR = REPO_ROOT / "svg"


class FrontendHandler(http.server.SimpleHTTPRequestHandler):
    def __init__(self, *args, **kwargs):
        super().__init__(*args, directory=str(FRONTEND_DIR), **kwargs)

    def do_GET(self):
        if self.path in {"/", ""}:
            self.path = "/index.html"
        super().do_GET()

    def send_head(self):
        if self.path == "/assets/floor_1.svg":
            return self.serve_file(SVG_DIR / "floor_1.svg", "image/svg+xml; charset=utf-8")

        return super().send_head()

    def serve_file(self, path: Path, content_type: str):
        if not path.exists():
            self.send_error(404, "File not found")
            return None

        payload = path.read_bytes()
        self.send_response(200)
        self.send_header("Content-Type", content_type)
        self.send_header("Content-Length", str(len(payload)))
        self.end_headers()
        return io.BytesIO(payload)


def main():
    parser = argparse.ArgumentParser(description="Static server for InterVUZ frontend")
    parser.add_argument("--port", type=int, default=3000)
    args = parser.parse_args()

    with socketserver.TCPServer(("", args.port), FrontendHandler) as httpd:
        print(f"frontend available at http://localhost:{args.port}")
        httpd.serve_forever()


if __name__ == "__main__":
    main()
