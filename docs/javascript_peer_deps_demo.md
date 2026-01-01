# Demo: JavaScript Peer Dependency Conflict & Rewrite
**Project**: Superintelligence Dependency Manager (Superint)
**Date**: January 1, 2026

This demo showcases how the agent handles the notorious "Peer Dependency Conflict" in the JavaScript ecosystem, moving beyond simple flag overrides to offer intelligent code rewriting solutions.

## 1. Initial State
We have a React 18 application that is trying to update, but a legacy UI library (`fancy-grid`) strictly requires React 16.

### `package.json`
```json
{
  "name": "dashboard-app",
  "dependencies": {
    "react": "^18.2.0",
    "fancy-grid": "1.0.0"
  }
}
```

### `node_modules/fancy-grid/package.json`
```json
{
  "peerDependencies": {
    "react": "^16.0.0"
  }
}
```

## 2. Scan & Conflict Detection
The scanner identifies the incompatibility.

### Command
```bash
./bin/superint-dep-manager scan --project "Dashboard App"
```

### Scan Output
```text
🔍 Scanning project: Dashboard App
WARN 🚨 Peer Dependency Conflict Detected!
   pkg: fancy-grid@1.0.0
   requires: react@^16.0.0
   found: react@18.2.0

❌ This conflict prevents standard installation (npm ERESOLVE).

💡 Available Resolutions:
   1. [OVERRIDE] Use --legacy-peer-deps (Risky: Runtime errors likely)
   2. [DOWNGRADE] Downgrade React to 16.14.0 (Regressive)
   3. [REWRITE] Replace 'fancy-grid' with a modern alternative (Recommended)
```

## 3. Intelligent Rewrite Resolution
Instead of forcing a broken state or holding back the entire app, we choose the **Rewrite** strategy. The agent identifies that `fancy-grid` is used only for a simple data table and replaces it with a modern, zero-dependency implementation compatible with React 18.

### Command
```bash
./bin/superint-dep-manager update --project "Dashboard App" --resolve-strategy rewrite --target fancy-grid
```

### Execution Output
```text
🤖 Initiating Rewrite Strategy for 'fancy-grid'...
   > Analyzing usage of 'FancyGrid' component...
   > Identified usage: Basic column rendering, no complex features used.
   
🎨 Generating Replacement Component...
   > Creating 'src/components/ui/ModernGrid.js'...
   > Implementing accessible table structure...
   > verifying compatibility with React 18...

🚀 Applying Code Patches...
   > [MODIFY] package.json (Removed fancy-grid)
   > [NEW] src/components/ui/ModernGrid.js (New component)
   > [MODIFY] src/App.js (Switched imports)

✅ Conflict Resolved via Code Evolution.
```

## 4. Comprehensive Diffs
The agent removed the blocking dependency and injected modern code.

### `package.json`
```diff
   "dependencies": {
     "react": "^18.2.0",
-    "fancy-grid": "1.0.0"
   }
```

### `src/App.js` (Component Swap)
```diff
 import React from 'react';
-import { FancyGrid } from 'fancy-grid';
+import { ModernGrid } from './components/ui/ModernGrid';
 
 export default function App({ data }) {
   return (
-    <FancyGrid 
-      data={data}
-      columns={['id', 'name', 'role']} 
-    />
+    <ModernGrid 
+      data={data}
+      columns={['id', 'name', 'role']}
+    />
   );
 }
```

### `src/components/ui/ModernGrid.js` (Generated Asset)
The agent generated a lightweight, peer-dependency-free replacement.
```javascript
import React from 'react';

export const ModernGrid = ({ data, columns }) => {
  return (
    <table className="modern-grid">
      <thead>
        <tr>
          {columns.map(col => <th key={col}>{col}</th>)}
        </tr>
      </thead>
      <tbody>
        {data.map((row, i) => (
          <tr key={i}>
            {columns.map(col => <td key={col}>{row[col]}</td>)}
          </tr>
        ))}
      </tbody>
    </table>
  );
};
```

## 5. Why this matters
Peer dependency conflicts often force developers to stay on old frameworks (React 16) or fork libraries. Superint provides a third option: **automated refactoring** to eliminate the tech debt entirely, effectively "evolving" the codebase past the blockage.
