# Demo: Python Dependency Upgrade
**Project**: Superintelligence Dependency Manager (Superint)
**Date**: December 29, 2025

This demo showcases how the `superint-dep-manager` CLI automatically identifies and patches outdated Python dependencies in a standard `requirements.txt` project.

## 1. Initial State
We start with a simple Python project containing outdated versions of `requests` and `flask`.

### `requirements.txt` (Before)
```text
requests==2.25.0
flask==2.0.0
```

## 2. Configuration & Scan
First, we configure the project to be monitored by Superint and run a dependency scan.

### Commands
```bash
# Configure the project
./bin/superint-dep-manager configure \
  --project-path ./projects/sample-python-project \
  --package-manager pip \
  --project-name "Sample Python Project"

# Run a dependency scan
./bin/superint-dep-manager scan --project "Sample Python Project"
```

### Scan Output Summary
```text
🔍 Scanning project: Sample Python Project
INFO Starting dependency scan for project ID: 2
INFO Completed dependency scan for project ID: 2 (found 2 dependencies, 2 updates)

🔄 Available Updates:
   📦 requests: 2.25.0 → 2.32.5 (minor)
   📦 flask: 2.0.0 → 3.1.2 (minor)
```

## 3. Intelligent Update
We apply the recommended updates using the `update` command. The CLI generates an update plan, assesses risks, and patches the manifest files.

### Command
```bash
./bin/superint-dep-manager update --project "Sample Python Project" --auto-approve
```

### Update Execution
```text
🔄 Updating project: Sample Python Project (pip)
📋 Generating update plan...
📋 Update Plan for Sample Python Project
==================================================
📊 Risk Summary:
   Total updates: 2
   🟢 Low risk: 2
   Overall risk: 🟢 Low

🚀 Applying updates...
INFO Successfully updated flask from 2.0.0 to 3.1.2
INFO Successfully updated requests from 2.25.0 to 2.32.5
🎉 All updates completed successfully!
```

## 4. Final Verification
The CLI has directly patched the `requirements.txt` file while preserving the pinned versioning format.

### `requirements.txt` (After)
```text
requests==2.32.5
flask==3.1.2
```

## Key Features Demonstrated
- **Manifest Patching**: Automatically edits `requirements.txt` with surgical precision.
- **Intelligent Risk Assessment**: Categorizes updates by risk level (Low, Medium, High).
- **PEP 668 Support**: Handles "externally managed environments" on modern Linux/macOS systems.
- **Full Rebranding**: Utilizes the new **Superintelligence** branding across all CLI outputs and configurations.

## 5. Detailed Diff & Heuristics Analysis
Beyond simple version bumps, the Superint engine generates precise patches and provides a transparent rationale for every change.

### Generated Diff Patch
The engine constructs a unified diff that can be applied directly or reviewed via Pull Request.

```diff
--- requirements.txt
+++ requirements.txt
@@ -1,2 +1,2 @@
-requests==2.25.0
-flask==2.0.0
+requests==2.32.5
+flask==3.1.2
```

### Heuristics Engine Logic
The `HeuristicProvider` analyzes changelogs and code usage to determine risk, rather than relying solely on semantic versioning.

1.  **Flask (2.0.0 → 3.1.2)**
    *   **Classification**: Major Update (usually High Risk).
    *   **Heuristic Analysis**: The engine scanned changelogs for keywords like "breaking change", "removed", and "API change".
    *   **Usage Correlation**: Checked project code against identified breaking changes. No usage of deprecated APIs was found.
    *   **Risk Result**: 🟢 **Low Risk** (Downgraded from High).

2.  **Requests (2.25.0 → 2.32.5)**
    *   **Classification**: Minor Update.
    *   **Heuristic Analysis**: Detected keywords "security fix" and "vulnerability" in intermediate versions.
    *   **Risk Result**: 🟢 **Low Risk** (Standard safe update).
