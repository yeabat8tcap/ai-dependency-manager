# Demo: JavaScript Complex Upgrade & Conflict Patching
**Project**: Superintelligence Dependency Manager (Superint)
**Date**: January 1, 2026

This demo showcases how the `superint-dep-manager` CLI handles complex dependency upgrades in a JavaScript (Node.js/React) project, specifically focusing on conflict detection and intelligent patching.

## 1. Initial State
We have a legacy React project with older dependencies and a potential conflict in `package.json` due to a concurrent update in another branch.

### `package.json` (Snippet)
```json
{
  "name": "super-dashboard",
  "version": "1.0.0",
  "dependencies": {
    "react": "^16.14.0",
    "express": "^4.0.0",
    "uuid": "^3.4.0"
  }
}
```

## 2. Configuration & Scan
We configure the project and initiate a scan.

### Commands
```bash
# Configure the project
./bin/superint-dep-manager configure \
  --project-path ./projects/super-dashboard \
  --package-manager npm \
  --project-name "Super Dashboard"

# Run a dependency scan
./bin/superint-dep-manager scan --project "Super Dashboard"
```

## 3. Complex Update Detection
The scanner identifies major updates for React and Express, flagging potential breaking changes and conflicts.

### Scan & Plan Output
```text
🔍 Scanning project: Super Dashboard
INFO Starting dependency scan for project ID: 4
INFO Completed dependency scan for project ID: 4 (found 3 dependencies, 3 updates)

🔄 Available Updates:
   📦 react: 16.14.0 → 18.2.0 (major) [⚠️ Breaking]
   📦 express: 4.0.0 → 4.18.2 (minor)
   📦 uuid: 3.4.0 → 9.0.0 (major) [⚠️ Breaking]

📋 Generating update plan...
WARN ⚠️ Merge conflict predicted in package.json (Dependency file risk)
INFO Analysis: Concurrent PR #42 is also modifying 'uuid' version.

📋 Update Plan for Super Dashboard
==================================================
📊 Risk Summary:
   Total updates: 3
   🔴 High risk: 2 (React, UUID)
   🟢 Low risk: 1 (Express)
   Overall risk: 🔴 High
```

## 4. Conflict Resolution & Patching
We execute the update with the `--resolve-conflicts` flag. The engine detects a merge conflict in `package.json` and uses its Heuristic Conflict Resolver to act.

### Command
```bash
./bin/superint-dep-manager update --project "Super Dashboard" --resolve-conflicts --auto-approve
```

### Execution Output
```text
🔄 Updating project: Super Dashboard (npm)
⚡ Detected Merge Conflict in file: package.json
   > Conflict Type: Dependency Version Mismatch
   > Incoming Change: "uuid": "^9.0.0"
   > Current Change (from PR #42): "uuid": "^8.3.2"

🤖 Engaging Conflict Resolver (Strategy: Auto-Resolve)...
   > Heuristic Analysis: Incoming version 9.0.0 is newer and compatible with project requirements.
   > Resolution: Accept Incoming Change ("^9.0.0")
   > Confidence: 0.9 (Safe)

✅ Conflict Resolved.

🚀 Applying Code Patches...
   > [MODIFY] package.json (Updated dependencies)
   > [MODIFY] src/components/App.js (Updated React lifecycle methods)
   > [MODIFY] src/app.js (Replaced bodyParser with express.json())

🎉 All updates completed successfully!
```

## 5. Detailed Diffs & Logic
Superint didn't just bump versions; it patched the code to be compatible with the new libraries.

### `package.json` Resolution
The resolver chose the higher version (`9.0.0`) over the conflicting change (`8.3.2`).

```diff
   "dependencies": {
-    "react": "^16.14.0",
+    "react": "^18.2.0",
-    "express": "^4.0.0",
+    "express": "^4.18.2",
<<<<<<< HEAD
-    "uuid": "^8.3.2"
=======
-    "uuid": "^3.4.0"
+    "uuid": "^9.0.0"
>>>>>>> incoming
   }
```
*Result: `uuid` set to `^9.0.0` automatically.*

### Code Mods (Template Patches)
The engine automatically applied code modifications to support the new library versions (`internal/github/patchgen.go`).

#### **React Lifecycle Update** (`src/components/App.js`)
*React 18 deprecated `componentWillMount`.*
```diff
--- src/components/App.js
+++ src/components/App.js
@@ -10,7 +10,7 @@
   constructor(props) {
     super(props);
   }
-  componentWillMount() {
+  componentDidMount() {
     this.fetchData();
   }
```

#### **Express Middleware Update** (`src/app.js`)
*Express 4.16+ includes built-in body parsing.*
```diff
--- src/app.js
+++ src/app.js
@@ -5,7 +5,7 @@
 const app = express();
 
-app.use(bodyParser.json());
+app.use(express.json());
 app.use(bodyParser.urlencoded({ extended: true }));
```

## 6. Heuristics Summary
- **Conflict Resolution**: The `ConflictResolver` identified a "Dependency Version Mismatch" and applied a "HIGHEST_VERSION" heuristic rule.
- **Code Patching**: The `PatchGenerator` matched the `react` upgrade to a known migration template ("React Component Lifecycle") and the `express` upgrade to the "Express BodyParser" removal template.
