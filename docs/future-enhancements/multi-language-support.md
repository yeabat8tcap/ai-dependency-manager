# Multi-Language Project Support

This document outlines the architecture for supporting multi-language projects and polyglot codebases in the AI Dependency Manager, enabling comprehensive dependency management across diverse technology stacks.

## Table of Contents

1. [Overview](#overview)
2. [Multi-Language Architecture](#multi-language-architecture)
3. [Language Detection](#language-detection)
4. [Cross-Language Dependencies](#cross-language-dependencies)
5. [Unified Dependency Graph](#unified-dependency-graph)
6. [Language-Specific Features](#language-specific-features)
7. [Polyglot Project Management](#polyglot-project-management)
8. [Implementation Plan](#implementation-plan)

## Overview

Modern applications often use multiple programming languages and ecosystems. This architecture enables the AI Dependency Manager to handle polyglot projects with dependencies spanning multiple languages, package managers, and runtime environments.

### Key Capabilities

- **Multi-Language Detection**: Automatically detect all languages in a project
- **Cross-Language Dependencies**: Track dependencies between different languages
- **Unified Management**: Single interface for managing all dependencies
- **Language-Specific Optimization**: Tailored analysis for each language
- **Polyglot Visualization**: Comprehensive view of multi-language dependency graphs

## Multi-Language Architecture

### Core Architecture

```go
type MultiLanguageManager struct {
    detectors       map[string]LanguageDetector
    analyzers       map[string]LanguageAnalyzer
    coordinators    map[string]LanguageCoordinator
    unifier         *DependencyUnifier
    graphBuilder    *PolyglotGraphBuilder
    conflictResolver *ConflictResolver
}

type PolyglotProject struct {
    ID              string                    `json:"id"`
    Name            string                    `json:"name"`
    Path            string                    `json:"path"`
    Languages       []*LanguageComponent     `json:"languages"`
    Dependencies    []*PolyglotDependency    `json:"dependencies"`
    CrossReferences []*CrossLanguageRef      `json:"cross_references"`
    UnifiedGraph    *UnifiedDependencyGraph  `json:"unified_graph"`
    Configuration   *PolyglotConfig          `json:"configuration"`
}

type LanguageComponent struct {
    Language        string                   `json:"language"`
    Version         string                   `json:"version"`
    Path            string                   `json:"path"`
    PackageManager  string                   `json:"package_manager"`
    Dependencies    []*Dependency            `json:"dependencies"`
    BuildSystem     string                   `json:"build_system"`
    Runtime         string                   `json:"runtime"`
    Metadata        map[string]interface{}   `json:"metadata"`
}

type PolyglotDependency struct {
    ID              string                   `json:"id"`
    Name            string                   `json:"name"`
    Version         string                   `json:"version"`
    Language        string                   `json:"language"`
    Ecosystem       string                   `json:"ecosystem"`
    Type            DependencyType           `json:"type"`
    Scope           DependencyScope          `json:"scope"`
    UsedBy          []*LanguageComponent     `json:"used_by"`
    Alternatives    []*AlternativeDep        `json:"alternatives"`
    CrossLangRefs   []*CrossLanguageRef      `json:"cross_lang_refs"`
}

type CrossLanguageRef struct {
    ID              string                   `json:"id"`
    SourceLang      string                   `json:"source_lang"`
    TargetLang      string                   `json:"target_lang"`
    RefType         CrossRefType             `json:"ref_type"`
    SourcePath      string                   `json:"source_path"`
    TargetPath      string                   `json:"target_path"`
    Interface       *InterfaceDefinition     `json:"interface"`
    Protocol        string                   `json:"protocol"`
}

func (mlm *MultiLanguageManager) AnalyzePolyglotProject(ctx context.Context, projectPath string) (*PolyglotProject, error) {
    // Detect all languages in the project
    languages, err := mlm.detectLanguages(projectPath)
    if err != nil {
        return nil, err
    }
    
    project := &PolyglotProject{
        ID:        generateProjectID(),
        Name:      filepath.Base(projectPath),
        Path:      projectPath,
        Languages: []*LanguageComponent{},
    }
    
    // Analyze each language component
    for _, lang := range languages {
        component, err := mlm.analyzeLanguageComponent(ctx, lang, projectPath)
        if err != nil {
            continue
        }
        project.Languages = append(project.Languages, component)
    }
    
    // Detect cross-language references
    crossRefs, err := mlm.detectCrossLanguageReferences(project.Languages)
    if err != nil {
        return nil, err
    }
    project.CrossReferences = crossRefs
    
    // Build unified dependency graph
    unifiedGraph, err := mlm.graphBuilder.BuildUnifiedGraph(project)
    if err != nil {
        return nil, err
    }
    project.UnifiedGraph = unifiedGraph
    
    // Unify dependencies across languages
    unifiedDeps, err := mlm.unifier.UnifyDependencies(project.Languages, crossRefs)
    if err != nil {
        return nil, err
    }
    project.Dependencies = unifiedDeps
    
    return project, nil
}
```

## Language Detection

### Automatic Language Detection

```go
type LanguageDetector interface {
    DetectLanguage(path string) (*LanguageInfo, error)
    GetConfidence() float64
    GetSupportedExtensions() []string
    GetSupportedFiles() []string
}

type LanguageInfo struct {
    Language        string                   `json:"language"`
    Version         string                   `json:"version"`
    Confidence      float64                  `json:"confidence"`
    Evidence        []*DetectionEvidence     `json:"evidence"`
    PackageManager  string                   `json:"package_manager"`
    BuildSystem     string                   `json:"build_system"`
    ConfigFiles     []string                 `json:"config_files"`
    SourceFiles     []string                 `json:"source_files"`
}

type DetectionEvidence struct {
    Type            EvidenceType             `json:"type"`
    File            string                   `json:"file"`
    Content         string                   `json:"content"`
    Weight          float64                  `json:"weight"`
    Description     string                   `json:"description"`
}

type CompositeLanguageDetector struct {
    detectors       []LanguageDetector
    fileAnalyzer    *FileAnalyzer
    contentAnalyzer *ContentAnalyzer
    heuristics      *DetectionHeuristics
}

func (cld *CompositeLanguageDetector) DetectAllLanguages(projectPath string) ([]*LanguageInfo, error) {
    var allLanguages []*LanguageInfo
    
    // Walk through project directory
    err := filepath.Walk(projectPath, func(path string, info os.FileInfo, err error) error {
        if err != nil {
            return err
        }
        
        if info.IsDir() {
            return nil
        }
        
        // Check each detector
        for _, detector := range cld.detectors {
            langInfo, err := detector.DetectLanguage(path)
            if err != nil {
                continue
            }
            
            if langInfo.Confidence > 0.7 {
                allLanguages = append(allLanguages, langInfo)
            }
        }
        
        return nil
    })
    
    if err != nil {
        return nil, err
    }
    
    // Consolidate and rank languages
    consolidated := cld.consolidateLanguages(allLanguages)
    
    return consolidated, nil
}

// JavaScript/TypeScript Detector
type JavaScriptDetector struct {
    patterns map[string]float64
}

func (jsd *JavaScriptDetector) DetectLanguage(path string) (*LanguageInfo, error) {
    filename := filepath.Base(path)
    ext := filepath.Ext(path)
    
    var confidence float64
    var evidence []*DetectionEvidence
    
    // File extension patterns
    switch ext {
    case ".js":
        confidence += 0.8
        evidence = append(evidence, &DetectionEvidence{
            Type: EvidenceTypeFileExtension,
            File: filename,
            Weight: 0.8,
            Description: "JavaScript file extension",
        })
    case ".ts":
        confidence += 0.9
        evidence = append(evidence, &DetectionEvidence{
            Type: EvidenceTypeFileExtension,
            File: filename,
            Weight: 0.9,
            Description: "TypeScript file extension",
        })
    case ".jsx", ".tsx":
        confidence += 0.85
        evidence = append(evidence, &DetectionEvidence{
            Type: EvidenceTypeFileExtension,
            File: filename,
            Weight: 0.85,
            Description: "React component file",
        })
    }
    
    // Special files
    switch filename {
    case "package.json":
        confidence += 0.95
        evidence = append(evidence, &DetectionEvidence{
            Type: EvidenceTypeConfigFile,
            File: filename,
            Weight: 0.95,
            Description: "Node.js package configuration",
        })
    case "tsconfig.json":
        confidence += 0.9
        evidence = append(evidence, &DetectionEvidence{
            Type: EvidenceTypeConfigFile,
            File: filename,
            Weight: 0.9,
            Description: "TypeScript configuration",
        })
    }
    
    if confidence < 0.5 {
        return nil, errors.New("insufficient confidence")
    }
    
    // Determine package manager
    packageManager := jsd.detectPackageManager(filepath.Dir(path))
    
    return &LanguageInfo{
        Language:       "javascript",
        Confidence:     confidence,
        Evidence:       evidence,
        PackageManager: packageManager,
    }, nil
}

// Python Detector
type PythonDetector struct {
    patterns map[string]float64
}

func (pd *PythonDetector) DetectLanguage(path string) (*LanguageInfo, error) {
    filename := filepath.Base(path)
    ext := filepath.Ext(path)
    
    var confidence float64
    var evidence []*DetectionEvidence
    
    // File extension patterns
    if ext == ".py" {
        confidence += 0.9
        evidence = append(evidence, &DetectionEvidence{
            Type: EvidenceTypeFileExtension,
            File: filename,
            Weight: 0.9,
            Description: "Python file extension",
        })
    }
    
    // Special files
    switch filename {
    case "requirements.txt", "setup.py", "pyproject.toml", "Pipfile":
        confidence += 0.95
        evidence = append(evidence, &DetectionEvidence{
            Type: EvidenceTypeConfigFile,
            File: filename,
            Weight: 0.95,
            Description: "Python dependency configuration",
        })
    }
    
    if confidence < 0.5 {
        return nil, errors.New("insufficient confidence")
    }
    
    // Determine package manager
    packageManager := pd.detectPackageManager(filepath.Dir(path))
    
    return &LanguageInfo{
        Language:       "python",
        Confidence:     confidence,
        Evidence:       evidence,
        PackageManager: packageManager,
    }, nil
}
```

## Cross-Language Dependencies

### Cross-Language Reference Detection

```go
type CrossLanguageAnalyzer struct {
    patterns        map[string]*CrossRefPattern
    protocolAnalyzer *ProtocolAnalyzer
    interfaceParser  *InterfaceParser
    bindingDetector  *BindingDetector
}

type CrossRefPattern struct {
    SourceLang      string                   `json:"source_lang"`
    TargetLang      string                   `json:"target_lang"`
    Pattern         string                   `json:"pattern"`
    RefType         CrossRefType             `json:"ref_type"`
    Confidence      float64                  `json:"confidence"`
}

type CrossRefType string

const (
    CrossRefTypeFFI        CrossRefType = "ffi"         // Foreign Function Interface
    CrossRefTypeRPC        CrossRefType = "rpc"         // Remote Procedure Call
    CrossRefTypeHTTP       CrossRefType = "http"        // HTTP API
    CrossRefTypeGRPC       CrossRefType = "grpc"        // gRPC
    CrossRefTypeWebSocket  CrossRefType = "websocket"   // WebSocket
    CrossRefTypeSharedLib  CrossRefType = "shared_lib"  // Shared Library
    CrossRefTypeProcess    CrossRefType = "process"     // Process Communication
    CrossRefTypeEmbedded   CrossRefType = "embedded"    // Embedded Runtime
)

func (cla *CrossLanguageAnalyzer) DetectCrossReferences(components []*LanguageComponent) ([]*CrossLanguageRef, error) {
    var crossRefs []*CrossLanguageRef
    
    // Analyze each component pair
    for i, source := range components {
        for j, target := range components {
            if i == j {
                continue
            }
            
            refs, err := cla.analyzeCrossReference(source, target)
            if err != nil {
                continue
            }
            
            crossRefs = append(crossRefs, refs...)
        }
    }
    
    return crossRefs, nil
}

func (cla *CrossLanguageAnalyzer) analyzeCrossReference(source, target *LanguageComponent) ([]*CrossLanguageRef, error) {
    var refs []*CrossLanguageRef
    
    // Check for known cross-language patterns
    pattern := fmt.Sprintf("%s->%s", source.Language, target.Language)
    
    switch pattern {
    case "python->c":
        refs = append(refs, cla.detectPythonCBindings(source, target)...)
    case "javascript->python":
        refs = append(refs, cla.detectJSPythonRPC(source, target)...)
    case "java->c":
        refs = append(refs, cla.detectJavaJNI(source, target)...)
    case "go->c":
        refs = append(refs, cla.detectGoCBindings(source, target)...)
    case "rust->c":
        refs = append(refs, cla.detectRustFFI(source, target)...)
    }
    
    // Check for protocol-based communication
    protocolRefs, err := cla.protocolAnalyzer.DetectProtocolReferences(source, target)
    if err == nil {
        refs = append(refs, protocolRefs...)
    }
    
    return refs, nil
}

func (cla *CrossLanguageAnalyzer) detectPythonCBindings(python, c *LanguageComponent) []*CrossLanguageRef {
    var refs []*CrossLanguageRef
    
    // Look for ctypes usage
    ctypesRefs := cla.findCtypesReferences(python.Path)
    for _, ref := range ctypesRefs {
        crossRef := &CrossLanguageRef{
            ID:         generateCrossRefID(),
            SourceLang: "python",
            TargetLang: "c",
            RefType:    CrossRefTypeFFI,
            SourcePath: ref.SourceFile,
            TargetPath: ref.LibraryPath,
            Protocol:   "ctypes",
        }
        refs = append(refs, crossRef)
    }
    
    // Look for Cython usage
    cythonRefs := cla.findCythonReferences(python.Path)
    for _, ref := range cythonRefs {
        crossRef := &CrossLanguageRef{
            ID:         generateCrossRefID(),
            SourceLang: "python",
            TargetLang: "c",
            RefType:    CrossRefTypeEmbedded,
            SourcePath: ref.SourceFile,
            TargetPath: ref.CFile,
            Protocol:   "cython",
        }
        refs = append(refs, crossRef)
    }
    
    return refs
}

func (cla *CrossLanguageAnalyzer) detectJSPythonRPC(js, python *LanguageComponent) []*CrossLanguageRef {
    var refs []*CrossLanguageRef
    
    // Look for HTTP API calls
    apiRefs := cla.findAPIReferences(js.Path, python.Path)
    for _, ref := range apiRefs {
        crossRef := &CrossLanguageRef{
            ID:         generateCrossRefID(),
            SourceLang: "javascript",
            TargetLang: "python",
            RefType:    CrossRefTypeHTTP,
            SourcePath: ref.ClientFile,
            TargetPath: ref.ServerFile,
            Protocol:   "http",
            Interface:  ref.APIInterface,
        }
        refs = append(refs, crossRef)
    }
    
    return refs
}
```

## Unified Dependency Graph

### Multi-Language Graph Builder

```go
type PolyglotGraphBuilder struct {
    graphBuilder    *GraphBuilder
    unifier         *DependencyUnifier
    layoutEngine    *MultiLangLayoutEngine
    visualizer      *PolyglotVisualizer
}

type UnifiedDependencyGraph struct {
    Nodes           []*PolyglotNode          `json:"nodes"`
    Edges           []*PolyglotEdge          `json:"edges"`
    Clusters        []*LanguageCluster       `json:"clusters"`
    CrossReferences []*CrossLanguageRef      `json:"cross_references"`
    Metadata        *GraphMetadata           `json:"metadata"`
    Statistics      *GraphStatistics         `json:"statistics"`
}

type PolyglotNode struct {
    ID              string                   `json:"id"`
    Type            NodeType                 `json:"type"`
    Language        string                   `json:"language"`
    Name            string                   `json:"name"`
    Version         string                   `json:"version"`
    Ecosystem       string                   `json:"ecosystem"`
    Dependencies    []string                 `json:"dependencies"`
    Dependents      []string                 `json:"dependents"`
    CrossLangRefs   []string                 `json:"cross_lang_refs"`
    Metadata        map[string]interface{}   `json:"metadata"`
    Position        *NodePosition            `json:"position"`
    Style           *NodeStyle               `json:"style"`
}

type PolyglotEdge struct {
    ID              string                   `json:"id"`
    Source          string                   `json:"source"`
    Target          string                   `json:"target"`
    Type            EdgeType                 `json:"type"`
    Language        string                   `json:"language"`
    CrossLanguage   bool                     `json:"cross_language"`
    Protocol        string                   `json:"protocol,omitempty"`
    Weight          float64                  `json:"weight"`
    Style           *EdgeStyle               `json:"style"`
}

type LanguageCluster struct {
    ID              string                   `json:"id"`
    Language        string                   `json:"language"`
    Nodes           []string                 `json:"nodes"`
    InternalEdges   []string                 `json:"internal_edges"`
    ExternalEdges   []string                 `json:"external_edges"`
    Statistics      *ClusterStatistics       `json:"statistics"`
    Style           *ClusterStyle            `json:"style"`
}

func (pgb *PolyglotGraphBuilder) BuildUnifiedGraph(project *PolyglotProject) (*UnifiedDependencyGraph, error) {
    graph := &UnifiedDependencyGraph{
        Nodes:           []*PolyglotNode{},
        Edges:           []*PolyglotEdge{},
        Clusters:        []*LanguageCluster{},
        CrossReferences: project.CrossReferences,
    }
    
    // Create nodes for each language component
    for _, component := range project.Languages {
        cluster := &LanguageCluster{
            ID:       fmt.Sprintf("cluster_%s", component.Language),
            Language: component.Language,
            Nodes:    []string{},
        }
        
        // Add dependency nodes
        for _, dep := range component.Dependencies {
            node := &PolyglotNode{
                ID:        fmt.Sprintf("%s_%s_%s", component.Language, dep.Name, dep.Version),
                Type:      NodeTypeDependency,
                Language:  component.Language,
                Name:      dep.Name,
                Version:   dep.Version,
                Ecosystem: dep.Ecosystem,
                Style:     pgb.getNodeStyle(component.Language, dep.Type),
            }
            
            graph.Nodes = append(graph.Nodes, node)
            cluster.Nodes = append(cluster.Nodes, node.ID)
        }
        
        graph.Clusters = append(graph.Clusters, cluster)
    }
    
    // Create edges for dependencies
    for _, component := range project.Languages {
        for _, dep := range component.Dependencies {
            for _, subDep := range dep.Dependencies {
                edge := &PolyglotEdge{
                    ID:       fmt.Sprintf("edge_%s_%s", dep.ID, subDep.ID),
                    Source:   dep.ID,
                    Target:   subDep.ID,
                    Type:     EdgeTypeDependency,
                    Language: component.Language,
                    Weight:   1.0,
                    Style:    pgb.getEdgeStyle(component.Language, EdgeTypeDependency),
                }
                
                graph.Edges = append(graph.Edges, edge)
            }
        }
    }
    
    // Create cross-language edges
    for _, crossRef := range project.CrossReferences {
        edge := &PolyglotEdge{
            ID:            fmt.Sprintf("cross_%s", crossRef.ID),
            Source:        crossRef.SourcePath,
            Target:        crossRef.TargetPath,
            Type:          EdgeTypeCrossLanguage,
            CrossLanguage: true,
            Protocol:      crossRef.Protocol,
            Weight:        0.5,
            Style:         pgb.getCrossLanguageEdgeStyle(crossRef.RefType),
        }
        
        graph.Edges = append(graph.Edges, edge)
    }
    
    // Calculate statistics
    graph.Statistics = pgb.calculateGraphStatistics(graph)
    
    return graph, nil
}
```

## Language-Specific Features

### Language-Specific Optimizations

```go
type LanguageSpecificManager struct {
    optimizers map[string]LanguageOptimizer
    analyzers  map[string]LanguageAnalyzer
    updaters   map[string]LanguageUpdater
}

type LanguageOptimizer interface {
    OptimizeDependencies(ctx context.Context, component *LanguageComponent) (*OptimizationResult, error)
    SuggestAlternatives(ctx context.Context, dependency *Dependency) ([]*Alternative, error)
    AnalyzePerformance(ctx context.Context, component *LanguageComponent) (*PerformanceAnalysis, error)
}

// JavaScript/Node.js Optimizer
type JavaScriptOptimizer struct {
    bundleAnalyzer *BundleAnalyzer
    treeShaker     *TreeShaker
    perfAnalyzer   *JSPerformanceAnalyzer
}

func (jso *JavaScriptOptimizer) OptimizeDependencies(ctx context.Context, component *LanguageComponent) (*OptimizationResult, error) {
    result := &OptimizationResult{
        Language:        component.Language,
        Optimizations:   []*Optimization{},
        EstimatedSavings: &Savings{},
    }
    
    // Analyze bundle size
    bundleAnalysis, err := jso.bundleAnalyzer.Analyze(component.Path)
    if err != nil {
        return nil, err
    }
    
    // Find unused dependencies
    unusedDeps := jso.findUnusedDependencies(bundleAnalysis)
    for _, dep := range unusedDeps {
        optimization := &Optimization{
            Type:        OptimizationTypeRemoveUnused,
            Dependency:  dep,
            Description: fmt.Sprintf("Remove unused dependency: %s", dep.Name),
            Impact:      jso.calculateRemovalImpact(dep, bundleAnalysis),
        }
        result.Optimizations = append(result.Optimizations, optimization)
    }
    
    // Find duplicate dependencies
    duplicates := jso.findDuplicateDependencies(component.Dependencies)
    for _, dup := range duplicates {
        optimization := &Optimization{
            Type:        OptimizationTypeDeduplication,
            Dependency:  dup.Primary,
            Description: fmt.Sprintf("Deduplicate %s (found %d versions)", dup.Name, len(dup.Versions)),
            Impact:      jso.calculateDeduplicationImpact(dup, bundleAnalysis),
        }
        result.Optimizations = append(result.Optimizations, optimization)
    }
    
    // Suggest lighter alternatives
    alternatives := jso.suggestLighterAlternatives(component.Dependencies)
    for _, alt := range alternatives {
        optimization := &Optimization{
            Type:        OptimizationTypeAlternative,
            Dependency:  alt.Original,
            Alternative: alt.Suggestion,
            Description: fmt.Sprintf("Replace %s with lighter alternative %s", alt.Original.Name, alt.Suggestion.Name),
            Impact:      alt.Impact,
        }
        result.Optimizations = append(result.Optimizations, optimization)
    }
    
    return result, nil
}

// Python Optimizer
type PythonOptimizer struct {
    venvAnalyzer   *VirtualEnvAnalyzer
    importAnalyzer *ImportAnalyzer
    perfAnalyzer   *PythonPerformanceAnalyzer
}

func (po *PythonOptimizer) OptimizeDependencies(ctx context.Context, component *LanguageComponent) (*OptimizationResult, error) {
    result := &OptimizationResult{
        Language:        component.Language,
        Optimizations:   []*Optimization{},
        EstimatedSavings: &Savings{},
    }
    
    // Analyze imports
    importAnalysis, err := po.importAnalyzer.Analyze(component.Path)
    if err != nil {
        return nil, err
    }
    
    // Find unused imports
    unusedImports := po.findUnusedImports(importAnalysis)
    for _, imp := range unusedImports {
        optimization := &Optimization{
            Type:        OptimizationTypeRemoveUnused,
            Description: fmt.Sprintf("Remove unused import: %s", imp.Module),
            Impact:      po.calculateImportRemovalImpact(imp),
        }
        result.Optimizations = append(result.Optimizations, optimization)
    }
    
    // Suggest more efficient alternatives
    alternatives := po.suggestEfficientAlternatives(component.Dependencies)
    for _, alt := range alternatives {
        optimization := &Optimization{
            Type:        OptimizationTypeAlternative,
            Dependency:  alt.Original,
            Alternative: alt.Suggestion,
            Description: fmt.Sprintf("Replace %s with more efficient %s", alt.Original.Name, alt.Suggestion.Name),
            Impact:      alt.Impact,
        }
        result.Optimizations = append(result.Optimizations, optimization)
    }
    
    return result, nil
}
```

## Implementation Plan

### Phase 1: Multi-Language Detection (Months 1-2)
- [ ] Implement composite language detection system
- [ ] Create language-specific detectors for major languages
- [ ] Build confidence scoring and evidence collection
- [ ] Add support for configuration file analysis

### Phase 2: Cross-Language Analysis (Months 3-4)
- [ ] Develop cross-language reference detection
- [ ] Implement protocol analysis for RPC/HTTP/gRPC
- [ ] Create FFI and binding detection systems
- [ ] Build interface definition parsing

### Phase 3: Unified Graph System (Months 5-6)
- [ ] Build polyglot dependency graph system
- [ ] Create unified visualization components
- [ ] Implement cross-language conflict resolution
- [ ] Add multi-language statistics and analytics

### Phase 4: Language-Specific Optimizations (Months 7-8)
- [ ] Develop language-specific optimizers
- [ ] Create performance analysis tools
- [ ] Build alternative suggestion systems
- [ ] Add comprehensive testing for polyglot projects

This multi-language architecture enables comprehensive dependency management across diverse technology stacks in modern polyglot applications.
