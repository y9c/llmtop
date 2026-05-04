#!/usr/bin/env python3
"""LLM inference server deployment manager.

Usage:
    python deploy/run.py vllm start|stop|restart|logs|status|test|metrics
    python deploy/run.py llamacpp start|stop|restart|logs|status|test|metrics
    python deploy/run.py ollama start|stop|restart|logs|status|test|metrics
    python deploy/run.py sglang start|stop|restart|logs|status|test|metrics
    python deploy/run.py list          # Show all backends
"""

import json
import subprocess
import sys
import time
from pathlib import Path

ROOT = Path(__file__).resolve().parent.parent

SERVICES = {
    "vllm": {
        "service_name": "vllm-qwen",
        "port": 8000,
        "model": "qwen3.6",
        "metrics_path": "/metrics",
        "start_cmd": ["systemctl", "--user", "start", "vllm-qwen"],
        "stop_cmd": ["systemctl", "--user", "stop", "vllm-qwen"],
    },
    "llamacpp": {
        "service_name": "llamacpp",
        "port": 8001,
        "model": "qwen3.6",
        "metrics_path": "/metrics",
        "start_cmd": ["systemctl", "--user", "start", "llamacpp"],
        "stop_cmd": ["systemctl", "--user", "stop", "llamacpp"],
    },
    "ollama": {
        "service_name": "ollama",
        "port": 11434,
        "model": "llama3.2",
        "metrics_path": "/api/ps",
        "start_cmd": ["systemctl", "--user", "start", "ollama"],
        "stop_cmd": ["systemctl", "--user", "stop", "ollama"],
    },
    "sglang": {
        "service_name": "sglang",
        "port": 8002,
        "model": "default",
        "metrics_path": "/metrics",
        "start_cmd": ["systemctl", "--user", "start", "sglang"],
        "stop_cmd": ["systemctl", "--user", "stop", "sglang"],
    },
}


def run(*args, capture=False, timeout=30):
    try:
        r = subprocess.run(args, capture_output=capture, text=True, timeout=timeout)
        return r
    except subprocess.TimeoutExpired:
        return None


def api_request(cfg, endpoint: str, method="GET", data=None):
    port = cfg["port"]
    url = f"http://localhost:{port}{endpoint}"
    args = ["curl", "-s", url]
    if data:
        args += ["-H", "Content-Type: application/json", "-d", json.dumps(data)]
    return run(*args, capture=True)


def action_backend(backend, action):
    cfg = SERVICES.get(backend)
    if not cfg:
        print(f"Unknown backend: {backend}")
        return

    service = cfg["service_name"]

    if action == "start":
        print(f"=== Starting {backend} ({service}) ===")
        run(*cfg["stop_cmd"])
        time.sleep(1)
        r = run(*cfg["start_cmd"])
        if r and r.returncode != 0:
            print(f"systemctl start failed (exit {r.returncode})")
        print("Check: systemctl --user status " + service)

    elif action == "stop":
        print(f"=== Stopping {backend} ===")
        run(*cfg["stop_cmd"])

    elif action == "restart":
        run(*cfg["stop_cmd"])
        time.sleep(1)
        run(*cfg["start_cmd"])

    elif action == "logs":
        run("journalctl", "--user", "-u", service, "-f", "--lines=30")

    elif action == "status":
        r = run("systemctl", "--user", "is-active", service, capture=True)
        print(r.stdout.strip() if r else "unknown")

    elif action == "metrics":
        r = api_request(cfg, cfg["metrics_path"])
        if r and r.stdout:
            print(r.stdout[:2000])
        else:
            print("No response")

    elif action == "test":
        if backend == "ollama":
            data = {"model": cfg["model"], "messages": [{"role": "user", "content": "hi"}]}
            r = api_request(cfg, "/api/chat", method="POST", data=data)
        else:
            data = {"model": cfg["model"], "messages": [{"role": "user", "content": "hi"}], "max_tokens": 10}
            r = api_request(cfg, "/v1/chat/completions", method="POST", data=data)

        if r and r.stdout:
            try:
                parsed = json.loads(r.stdout)
                print(json.dumps(parsed, indent=2)[:500])
            except json.JSONDecodeError:
                print(r.stdout[:500])
        else:
            print("No response or timeout")


def list_backends():
    print("=== Available Backends ===")
    for name, cfg in SERVICES.items():
        r = run("systemctl", "--user", "is-active", cfg["service_name"], capture=True)
        status = r.stdout.strip() if r else "unknown"
        print(f"  {name:12s} port={cfg['port']:<5d} status={status}")


def main():
    if len(sys.argv) < 2:
        print(__doc__)
        sys.exit(1)

    cmd = sys.argv[1]
    if cmd == "list":
        list_backends()
    elif cmd in SERVICES:
        action_backend(cmd, sys.argv[2] if len(sys.argv) > 2 else "status")
    else:
        print(f"Unknown command: {cmd}")
        sys.exit(1)


if __name__ == "__main__":
    main()
