# Ollama Local AI Integration Guide

This guide explains how to set up and use Ollama for local AI-powered dependency analysis in the AI Dependency Manager.

## Overview

Ollama integration provides:
- **Privacy**: All AI processing happens locally, no data sent to external APIs
- **Cost-Effective**: No API costs for AI analysis, unlimited usage
- **Offline Capability**: Works without internet connection once models are downloaded
- **Customization**: Ability to use different models optimized for different tasks
- **Performance**: Low latency for local processing
- **Security**: Enterprise-friendly with no external data transmission

## Prerequisites

### 1. Install Ollama

Visit [https://ollama.ai/](https://ollama.ai/) and download Ollama for your operating system.

**macOS:**
```bash
curl -fsSL https://ollama.ai/install.sh | sh
```

**Linux:**
```bash
curl -fsSL https://ollama.ai/install.sh | sh
```

**Windows:**
Download the installer from the Ollama website.

### 2. Start Ollama Service

After installation, start the Ollama service:

```bash
ollama serve
```

The service will start on `http://localhost:11434` by default.

### 3. Download Models

Download one or more models for dependency analysis:

```bash
# Recommended models for dependency analysis
ollama pull llama2          # General purpose, good balance of speed and quality
ollama pull codellama       # Optimized for code analysis
ollama pull mistral         # Fast and efficient
ollama pull phi             # Lightweight, good for simple analysis

# List downloaded models
ollama list
```

## Configuration

### Environment Variables

Configure Ollama integration using environment variables:

```bash
# Ollama server endpoint (default: http://localhost:11434)
export OLLAMA_BASE_URL="http://localhost:11434"

# Model to use for analysis (default: llama2)
export OLLAMA_MODEL="llama2"

# Set Ollama as the default AI provider
export AI_DEFAULT_PROVIDER="ollama"
```

### Configuration File

Add Ollama configuration to your AI Dependency Manager config file:

```yaml
ai:
  default_provider: "ollama"
  fallback_providers: ["ollama", "heuristic"]
  enable_heuristic_fallback: true
  
  ollama:
    base_url: "http://localhost:11434"
    model: "llama2"
    temperature: 0.7
    top_p: 0.9
    top_k: 40
    num_predict: 2048
```

## Model Selection Guide

Choose the right model based on your needs:

### **llama2** (Recommended)
- **Best for**: General dependency analysis
- **Size**: ~3.8GB
- **Speed**: Good
- **Quality**: High
- **Use case**: Balanced performance for most scenarios

### **codellama**
- **Best for**: Code-focused dependency analysis
- **Size**: ~3.8GB
- **Speed**: Good
- **Quality**: High for code
- **Use case**: When analyzing code changes and technical dependencies

### **mistral**
- **Best for**: Fast analysis with good quality
- **Size**: ~4.1GB
- **Speed**: Fast
- **Quality**: High
- **Use case**: When you need quick analysis results

### **phi**
- **Best for**: Lightweight, resource-constrained environments
- **Size**: ~1.6GB
- **Speed**: Very fast
- **Quality**: Good
- **Use case**: Limited resources or simple analysis needs

## Usage

### 1. Validate Setup

Use the validation tool to ensure everything is working:

```bash
# Build and run Ollama validation
go run cmd/validate-ollama/main.go
```

### 2. Basic Usage

Once configured, the AI Dependency Manager will automatically use Ollama:

```bash
# Scan dependencies with Ollama analysis
./ai-dep-manager scan

# Check for updates with AI insights
./ai-dep-manager check --ai-analysis

# Get detailed analysis for specific package
./ai-dep-manager analyze react 17.0.0 18.0.0
```

### 3. Advanced Configuration

#### Custom Model Parameters

Fine-tune model behavior:

```bash
export OLLAMA_MODEL="codellama"
export OLLAMA_TEMPERATURE="0.1"    # More deterministic (0.0-1.0)
export OLLAMA_TOP_P="0.9"          # Nucleus sampling (0.0-1.0)
export OLLAMA_TOP_K="40"           # Top-k sampling
export OLLAMA_NUM_PREDICT="1024"   # Max tokens to generate
```

#### Multiple Model Setup

Switch between models for different analysis types:

```bash
# Use CodeLlama for code analysis
OLLAMA_MODEL="codellama" ./ai-dep-manager analyze

# Use Mistral for quick checks
OLLAMA_MODEL="mistral" ./ai-dep-manager check --quick
```

## Performance Optimization

### Hardware Requirements

For optimal performance:

- **RAM**: 8GB minimum, 16GB+ recommended
- **Storage**: 10GB+ free space for models
- **CPU**: Multi-core processor recommended

### Model Management

```bash
# Remove unused models to save space
ollama rm old-model

# Update models
ollama pull llama2

# Check model sizes
ollama list
```

### Performance Tuning

Adjust parameters based on your hardware:

```yaml
ollama:
  # For faster responses (lower quality)
  temperature: 0.1
  num_predict: 512
  
  # For better quality (slower responses)
  temperature: 0.7
  num_predict: 2048
```

## Troubleshooting

### Common Issues

#### 1. Ollama Not Available
```
❌ Ollama is not available at http://localhost:11434
```

**Solutions:**
- Ensure Ollama service is running: `ollama serve`
- Check if port 11434 is available
- Verify firewall settings

#### 2. Model Not Found
```
❌ Model 'llama2' not found
```

**Solutions:**
- Download the model: `ollama pull llama2`
- Check available models: `ollama list`
- Verify model name in configuration

#### 3. Slow Performance
```
⚠️ Analysis taking longer than expected
```

**Solutions:**
- Use a smaller model (phi instead of llama2)
- Reduce `num_predict` parameter
- Ensure sufficient RAM available
- Close other resource-intensive applications

#### 4. Out of Memory
```
❌ Failed to load model: out of memory
```

**Solutions:**
- Use a smaller model
- Increase system RAM
- Close other applications
- Restart Ollama service

### Debug Mode

Enable debug logging for troubleshooting:

```bash
export LOG_LEVEL="DEBUG"
./ai-dep-manager scan
```

### Health Check

Verify Ollama health:

```bash
# Check Ollama status
curl http://localhost:11434/api/tags

# Test model availability
curl http://localhost:11434/api/generate -d '{
  "model": "llama2",
  "prompt": "Hello",
  "stream": false
}'
```

## Integration Examples

### CI/CD Pipeline

```yaml
# .github/workflows/dependency-check.yml
name: Dependency Analysis
on: [push, pull_request]

jobs:
  analyze:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      
      - name: Setup Ollama
        run: |
          curl -fsSL https://ollama.ai/install.sh | sh
          ollama serve &
          sleep 10
          ollama pull llama2
      
      - name: Run Dependency Analysis
        run: |
          export OLLAMA_MODEL="llama2"
          export AI_DEFAULT_PROVIDER="ollama"
          ./ai-dep-manager scan --format=json > analysis.json
      
      - name: Upload Results
        uses: actions/upload-artifact@v2
        with:
          name: dependency-analysis
          path: analysis.json
```

### Docker Setup

```dockerfile
FROM ollama/ollama:latest

# Install AI Dependency Manager
COPY ai-dep-manager /usr/local/bin/

# Download model
RUN ollama serve & \
    sleep 10 && \
    ollama pull llama2

# Set environment
ENV OLLAMA_MODEL=llama2
ENV AI_DEFAULT_PROVIDER=ollama

EXPOSE 11434
CMD ["ollama", "serve"]
```

## Security Considerations

### Data Privacy
- All analysis happens locally
- No data sent to external services
- Models run entirely offline

### Network Security
- Ollama runs on localhost by default
- No external network access required
- Can be isolated in secure environments

### Access Control
- Ollama service runs with user permissions
- Models stored in user directory
- No elevated privileges required

## Best Practices

### 1. Model Selection
- Use `llama2` for general analysis
- Use `codellama` for code-heavy projects
- Use `phi` for resource-constrained environments

### 2. Performance
- Monitor system resources during analysis
- Use appropriate model size for your hardware
- Consider batch processing for large projects

### 3. Maintenance
- Regularly update models: `ollama pull model-name`
- Monitor disk space usage
- Keep Ollama service updated

### 4. Fallback Strategy
- Always configure heuristic fallback
- Test fallback scenarios
- Monitor provider availability

## Support

For issues and questions:

1. **Ollama Issues**: [https://github.com/jmorganca/ollama/issues](https://github.com/jmorganca/ollama/issues)
2. **AI Dependency Manager**: Check project documentation
3. **Model Performance**: Try different models or adjust parameters

## Advanced Topics

### Custom Model Training
For specialized use cases, consider fine-tuning models on your specific dependency patterns.

### Multi-Model Ensemble
Use multiple models for different analysis types and combine results for higher accuracy.

### Performance Monitoring
Implement metrics collection to monitor analysis performance and model effectiveness.
