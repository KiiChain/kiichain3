import json
import re
import base64

from pathlib import Path

HEX_RE = re.compile(r'^[0-9a-fA-F]+$')
POSSIBLE_HEX_RE = re.compile(r'^[0-9a-fA-F]{4,}$')  # At least 4 chars, avoids short irrelevant strings
POSSIBLE_B64_RE = re.compile(r'^[A-Za-z0-9+/=]{8,}$')

def is_valid_hex(s):
    return HEX_RE.fullmatch(s) is not None

def is_valid_base64(s):
    try:
        base64.b64decode(s, validate=True)
        return True
    except Exception:
        return False

def scan_json(data, path=""):
    issues = []
    if isinstance(data, dict):
        for k, v in data.items():
            new_path = f"{path}.{k}" if path else k
            issues.extend(scan_json(v, new_path))
    elif isinstance(data, list):
        for i, v in enumerate(data):
            new_path = f"{path}[{i}]"
            issues.extend(scan_json(v, new_path))
    elif isinstance(data, str):
        if not is_valid_base64(data):
            issues.append((path, data, 'invalid base64'))
    return issues

# Load your file
file_path = "export_sorted.json"  # adjust path if needed
data = json.loads(Path(file_path).read_text())

# Scan
errors = scan_json(data['app_state']['wasm'])
for path, value, kind in errors:
    print(f"[{kind.upper()}] At {path}: {value}")
