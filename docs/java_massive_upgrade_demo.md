# Demo: Java Massive Framework Upgrade (Spring Boot 2 → 3)
**Project**: Superintelligence Dependency Manager (Superint)
**Date**: January 1, 2026

This demo showcases a "massive" server-side upgrade: migrating a Java Spring Boot application from version 2.x to 3.x. This process is notoriously difficult because it requires not just dependency updates, but also a wholesale namespace migration from `javax.*` to `jakarta.*` (Jakarta EE 9/10).

## 1. Initial State
We have a legacy microservice running Spring Boot 2.7.5. It extensively uses the Servlet API.

### `pom.xml` (Snippet)
```xml
<dependencies>
    <dependency>
        <groupId>org.springframework.boot</groupId>
        <artifactId>spring-boot-starter-web</artifactId>
        <version>2.7.5</version>
    </dependency>
</dependencies>
```

### `src/main/java/com/example/demo/Controller.java`
```java
package com.example.demo;

import javax.servlet.http.HttpServletRequest;
import javax.servlet.http.HttpServletResponse;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.RestController;

@RestController
public class Controller {
    @GetMapping("/hello")
    public String hello(HttpServletRequest request) {
        return "Hello from " + request.getRemoteAddr();
    }
}
```

## 2. Deep Scan & Migration Detection
The scanner detects the Spring Boot 3 available update and identifies the corresponding migration template.

### Command
```bash
./bin/superint-dep-manager scan --project "Legacy Microservice"
```

### Scan Output
```text
🔍 Scanning project: Legacy Microservice
INFO Parsing pom.xml...
INFO Completed dependency scan.

🔄 Available Updates:
   📦 spring-boot-starter-web: 2.7.5 → 3.2.0 (major) [⚠️ Breaking]

📋 Generating update plan...
WARN ⚠️ Major framework upgrade detected: Spring Boot 3.x
INFO Analysis: This upgrade requires migrating from Java EE schemas (javax.*) to Jakarta EE (jakarta.*).
INFO Template Match: Spring Boot 2 to 3 Migration (Confidence: 0.9)

📋 Update Plan for Legacy Microservice
==================================================
📊 Risk Summary:
   Total updates: 1
   🔴 High risk: 1 (Framework Migration)
   Overall risk: 🔴 High (Requires Code Refactoring)
```

## 3. Automated Migration
We execute the update. The underlying engine (`TemplateService`) triggers the specific transformation rules for this migration path.

### Command
```bash
./bin/superint-dep-manager update --project "Legacy Microservice" --migrate --auto-approve
```

### Execution Output
```text
🔄 Updating project: Legacy Microservice (maven)
🚀 Applying Migration Template: "Spring Boot 2 to 3 Migration"
   > [TRANSFORM] Replacing 'javax.servlet' with 'jakarta.servlet' across 42 files.
   > [MODIFY] pom.xml (Updated dependency version)

🧪 Verifying Migration...
   > Running 'mvn clean compile'... SUCCESS
   > Running 'mvn test'... SUCCESS

🎉 Migration completed successfully!
```

## 4. Comprehensive Diffs
The tool performed a synchronized update of the build configuration and the codebase.

### Build Configuration (`pom.xml`)
```diff
     <dependency>
         <groupId>org.springframework.boot</groupId>
         <artifactId>spring-boot-starter-web</artifactId>
-        <version>2.7.5</version>
+        <version>3.2.0</version>
     </dependency>
```

### Source Code Refactoring (`Controller.java`)
The `javax` imports were replaced with `jakarta`.

```diff
 package com.example.demo;
 
-import javax.servlet.http.HttpServletRequest;
-import javax.servlet.http.HttpServletResponse;
+import jakarta.servlet.http.HttpServletRequest;
+import jakarta.servlet.http.HttpServletResponse;
 import org.springframework.web.bind.annotation.GetMapping;
 import org.springframework.web.bind.annotation.RestController;
```

## 5. Technical Details
- **Template System**: The migration logic is defined in `internal/github/templates.go`, which contains a registry of known migrations (e.g., `spring-boot-2-to-3`).
- **Regex Transformation**: Code changes are applied using language-aware regex replacement rules (e.g., `s/javax\.servlet\./jakarta.servlet./g`) defined in the template.
- **Scope**: The patch was applied to all `*.java` files in the project structure, ensuring typical project layouts (Maven/Gradle) are covered.
