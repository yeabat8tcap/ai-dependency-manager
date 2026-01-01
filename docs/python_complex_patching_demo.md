# Demo: Python Hole Dependencies & Zero-Dependency Patching
**Project**: Superintelligence Dependency Manager (Superint)
**Date**: January 1, 2026

This demo illustrates advanced patching scenarios in a Python project: filling "hole" dependencies (used but undeclared) and replacing vulnerable libraries with "zero-dependency" custom code patches.

## 1. Initial State
We have a Python service that has drifted from its `requirements.txt`. It uses `requests` without declaring it, and relies on an abandoned library `simple-pad` for string padding which has a security flag.

### `source.py`
```python
import requests
from simple_pad import pad_left

def fetch_data():
    return requests.get("https://api.example.com")

def format_id(id):
    return pad_left(str(id), 5, "0")
```

### `requirements.txt`
```text
flask==2.3.0
# simple-pad is installed in environment but flagged as vulnerable/abandoned
simple-pad==1.0.0
# requests is MISSING
```

## 2. Deep Scan Analysis
The scanner performs static analysis on the source code to validate imports against the manifest.

### Command
```bash
./bin/superint-dep-manager scan --project "Legacy Service" --deep-scan
```

### Scan Output
```text
🔍 Deep Scanning project: Legacy Service
INFO Parsing source files...
INFO Validating imports against requirements.txt...

🚨 Anomalies Detected:
   ❌ [HOLE] 'requests' is imported in src/source.py:1 but missing from requirements.txt
   ⚠️ [VULN] 'simple-pad' (1.0.0) is abandoned and has 1 critical vulnerability. No safe version available.

📋 Recommendation:
   1. Add 'requests' to requirements.txt (Standard Fix)
   2. Remove 'simple-pad' and replace with native string formatting (Zero-Dependency Fix)
```

## 3. Intelligent Patching
We run the update with the `--heal` flag to fill holes and apply custom patches.

### Command
```bash
./bin/superint-dep-manager update --project "Legacy Service" --heal --auto-approve
```

### Execution Output
```text
🚑 Healing project: Legacy Service
🛠️  Fixing Hole Dependencies...
   > Identified 'requests' version from environment: 2.31.0
   > [ADD] requests==2.31.0 to requirements.txt

🧪 Generating Zero-Dependency Patch for 'simple-pad'...
   > AI Analysis: 'simple_pad.pad_left' can be replaced with Python's str.zfill() or f-strings.
   > Generating native implementation...

🚀 Applying Code Patches...
   > [MODIFY] requirements.txt (Added requests, Removed simple-pad)
   > [MODIFY] src/source.py (Replaced library call with native code)

✨ Project Healed! 2 issues resolved.
```

## 4. Comprehensive Diffs
The tool performed two distinct types of operations: valid manifest repair and code replacement.

### Manifest Repair (`requirements.txt`)
It filled the hole and removed the problem library.
```diff
--- requirements.txt
+++ requirements.txt
@@ -1,3 +1,3 @@
 flask==2.3.0
-simple-pad==1.0.0
+requests==2.31.0
```

### Zero-Dependency Code Patch (`src/source.py`)
The AI recognized that `pad_left` is trivial and replaced the import and usage with standard Python methods, effectively removing the dependency (Zero Dep).

```diff
--- src/source.py
+++ src/source.py
@@ -1,5 +1,4 @@
 import requests
-from simple_pad import pad_left
 
 def fetch_data():
     return requests.get("https://api.example.com")
 
 def format_id(id):
-    return pad_left(str(id), 5, "0")
+    return str(id).zfill(5)
```

## 5. Summary of Heuristics
- **Hole Detection**: The scanner compared the AST (Abstract Syntax Tree) of `.py` files against `pip` metadata to find the discrepancy.
- **Zero-Dep Strategy**: When a dependency has high risk and low complexity (like string padding), the `HeuristicProvider` suggests a native replacement instead of a different library, reducing the attack surface.
