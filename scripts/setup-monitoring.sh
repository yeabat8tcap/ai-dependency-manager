#!/bin/bash

# Superintelligence Dependency Manager - Production Monitoring Setup Script
# This script sets up comprehensive monitoring and logging for production

set -e

# Configuration
APP_NAME="superint-dep-manager"
DEPLOY_DIR="/opt/superint-dep-manager"
MONITORING_DIR="/opt/superint-dep-manager/monitoring"
USER="superint-dep-manager"
GROUP="superint-dep-manager"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Logging functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Create monitoring directory structure
create_monitoring_structure() {
    log_info "Creating monitoring directory structure..."
    
    mkdir -p "$MONITORING_DIR"/{scripts,alerts,dashboards,logs}
    chown -R "$USER:$GROUP" "$MONITORING_DIR"
    chmod 755 "$MONITORING_DIR"
    
    log_success "Monitoring directory structure created"
}

# Create health check script
create_health_check() {
    log_info "Creating health check script..."
    
    cat > "$MONITORING_DIR/scripts/health-check.sh" << 'EOF'
#!/bin/bash

# Superintelligence Dependency Manager Health Check Script

APP_NAME="superint-dep-manager"
DEPLOY_DIR="/opt/superint-dep-manager"
SERVICE_NAME="superint-dep-manager"
LOG_FILE="/opt/superint-dep-manager/logs/health-check.log"

# Health check function
check_health() {
    local timestamp=$(date '+%Y-%m-%d %H:%M:%S')
    local status="HEALTHY"
    local issues=()
    
    # Check if service is running
    if ! systemctl is-active --quiet "$SERVICE_NAME"; then
        status="UNHEALTHY"
        issues+=("Service not running")
    fi
    
    # Check if binary responds
    if ! sudo -u superint-dep-manager "$DEPLOY_DIR/bin/$APP_NAME" version > /dev/null 2>&1; then
        status="UNHEALTHY"
        issues+=("Binary not responding")
    fi
    
    # Check database connectivity
    if ! sudo -u superint-dep-manager "$DEPLOY_DIR/bin/$APP_NAME" status > /dev/null 2>&1; then
        status="UNHEALTHY"
        issues+=("Database connectivity issues")
    fi
    
    # Check disk space
    local disk_usage=$(df "$DEPLOY_DIR" | awk 'NR==2 {print $5}' | sed 's/%//')
    if [[ $disk_usage -gt 90 ]]; then
        status="WARNING"
        issues+=("High disk usage: ${disk_usage}%")
    fi
    
    # Check memory usage
    local memory_usage=$(ps -o pid,ppid,cmd,%mem --sort=-%mem -C "$APP_NAME" | awk 'NR==2 {print $4}')
    if [[ $(echo "$memory_usage > 80" | bc -l) -eq 1 ]]; then
        status="WARNING"
        issues+=("High memory usage: ${memory_usage}%")
    fi
    
    # Log results
    echo "[$timestamp] Status: $status" >> "$LOG_FILE"
    if [[ ${#issues[@]} -gt 0 ]]; then
        printf "[$timestamp] Issues: %s\n" "$(IFS=', '; echo "${issues[*]}")" >> "$LOG_FILE"
    fi
    
    # Return appropriate exit code
    case $status in
        "HEALTHY") exit 0 ;;
        "WARNING") exit 1 ;;
        "UNHEALTHY") exit 2 ;;
    esac
}

check_health
EOF

    chmod +x "$MONITORING_DIR/scripts/health-check.sh"
    chown "$USER:$GROUP" "$MONITORING_DIR/scripts/health-check.sh"
    
    log_success "Health check script created"
}

# Create metrics collection script
create_metrics_script() {
    log_info "Creating metrics collection script..."
    
    cat > "$MONITORING_DIR/scripts/collect-metrics.sh" << 'EOF'
#!/bin/bash

# Superintelligence Dependency Manager Metrics Collection Script

APP_NAME="superint-dep-manager"
DEPLOY_DIR="/opt/superint-dep-manager"
METRICS_FILE="/opt/superint-dep-manager/logs/metrics.log"

collect_metrics() {
    local timestamp=$(date '+%Y-%m-%d %H:%M:%S')
    
    # System metrics
    local cpu_usage=$(top -bn1 | grep "Cpu(s)" | awk '{print $2}' | sed 's/%us,//')
    local memory_total=$(free -m | awk 'NR==2{print $2}')
    local memory_used=$(free -m | awk 'NR==2{print $3}')
    local disk_usage=$(df "$DEPLOY_DIR" | awk 'NR==2 {print $5}' | sed 's/%//')
    
    # Application metrics
    local app_memory=$(ps -o pid,ppid,cmd,%mem --sort=-%mem -C "$APP_NAME" | awk 'NR==2 {print $4}')
    local app_cpu=$(ps -o pid,ppid,cmd,%cpu --sort=-%cpu -C "$APP_NAME" | awk 'NR==2 {print $4}')
    
    # Database metrics (if accessible)
    local db_size=$(du -sh "$DEPLOY_DIR/data" 2>/dev/null | awk '{print $1}' || echo "N/A")
    
    # Service status
    local service_status=$(systemctl is-active superint-dep-manager)
    local uptime=$(systemctl show superint-dep-manager --property=ActiveEnterTimestamp | cut -d= -f2)
    
    # Log metrics in JSON format
    cat >> "$METRICS_FILE" << EOL
{
  "timestamp": "$timestamp",
  "system": {
    "cpu_usage": "$cpu_usage",
    "memory_total": $memory_total,
    "memory_used": $memory_used,
    "disk_usage": $disk_usage
  },
  "application": {
    "memory_usage": "$app_memory",
    "cpu_usage": "$app_cpu",
    "service_status": "$service_status",
    "uptime": "$uptime"
  },
  "database": {
    "size": "$db_size"
  }
}
EOL
}

collect_metrics
EOF

    chmod +x "$MONITORING_DIR/scripts/collect-metrics.sh"
    chown "$USER:$GROUP" "$MONITORING_DIR/scripts/collect-metrics.sh"
    
    log_success "Metrics collection script created"
}

# Create alert script
create_alert_script() {
    log_info "Creating alert script..."
    
    cat > "$MONITORING_DIR/scripts/send-alert.sh" << 'EOF'
#!/bin/bash

# Superintelligence Dependency Manager Alert Script

ALERT_TYPE="$1"
MESSAGE="$2"
SEVERITY="${3:-INFO}"

# Configuration
WEBHOOK_URL="${WEBHOOK_URL:-}"
EMAIL_TO="${EMAIL_TO:-admin@example.com}"
LOG_FILE="/opt/superint-dep-manager/logs/alerts.log"

send_alert() {
    local timestamp=$(date '+%Y-%m-%d %H:%M:%S')
    
    # Log alert
    echo "[$timestamp] [$SEVERITY] $ALERT_TYPE: $MESSAGE" >> "$LOG_FILE"
    
    # Send webhook if configured
    if [[ -n "$WEBHOOK_URL" ]]; then
        curl -X POST "$WEBHOOK_URL" \
             -H "Content-Type: application/json" \
             -d "{
                 \"text\": \"🚨 Superintelligence Dependency Manager Alert\",
                 \"attachments\": [{
                     \"color\": \"danger\",
                     \"fields\": [{
                         \"title\": \"Alert Type\",
                         \"value\": \"$ALERT_TYPE\",
                         \"short\": true
                     }, {
                         \"title\": \"Severity\",
                         \"value\": \"$SEVERITY\",
                         \"short\": true
                     }, {
                         \"title\": \"Message\",
                         \"value\": \"$MESSAGE\",
                         \"short\": false
                     }, {
                         \"title\": \"Timestamp\",
                         \"value\": \"$timestamp\",
                         \"short\": true
                     }]
                 }]
             }" > /dev/null 2>&1
    fi
    
    # Send email if mail command is available
    if command -v mail > /dev/null 2>&1; then
        echo "Alert: $ALERT_TYPE - $MESSAGE (Severity: $SEVERITY) at $timestamp" | \
            mail -s "Superintelligence Dependency Manager Alert" "$EMAIL_TO"
    fi
}

send_alert
EOF

    chmod +x "$MONITORING_DIR/scripts/send-alert.sh"
    chown "$USER:$GROUP" "$MONITORING_DIR/scripts/send-alert.sh"
    
    log_success "Alert script created"
}

# Setup cron jobs for monitoring
setup_cron_jobs() {
    log_info "Setting up monitoring cron jobs..."
    
    # Create cron jobs for the superint-dep-manager user
    cat > "/tmp/superint-dep-manager-cron" << EOF
# Superintelligence Dependency Manager Monitoring Cron Jobs

# Health check every 5 minutes
*/5 * * * * $MONITORING_DIR/scripts/health-check.sh || $MONITORING_DIR/scripts/send-alert.sh "HEALTH_CHECK" "Health check failed" "CRITICAL"

# Metrics collection every minute
* * * * * $MONITORING_DIR/scripts/collect-metrics.sh

# Daily backup at 2 AM
0 2 * * * $DEPLOY_DIR/scripts/backup.sh

# Weekly log cleanup at 3 AM on Sundays
0 3 * * 0 find $DEPLOY_DIR/logs -name "*.log" -mtime +30 -delete
EOF

    crontab -u "$USER" "/tmp/superint-dep-manager-cron"
    rm "/tmp/superint-dep-manager-cron"
    
    log_success "Monitoring cron jobs configured"
}

# Create log aggregation configuration
create_log_config() {
    log_info "Creating log aggregation configuration..."
    
    # Create rsyslog configuration for application logs
    cat > "/etc/rsyslog.d/50-superint-dep-manager.conf" << EOF
# Superintelligence Dependency Manager log configuration

# Application logs
:programname,isequal,"superint-dep-manager" /opt/superint-dep-manager/logs/application.log
& stop

# Health check logs
:msg,contains,"superint-dep-manager health" /opt/superint-dep-manager/logs/health.log
& stop

# Metrics logs
:msg,contains,"superint-dep-manager metrics" /opt/superint-dep-manager/logs/metrics.log
& stop
EOF

    systemctl restart rsyslog
    
    log_success "Log aggregation configured"
}

# Create monitoring dashboard template
create_dashboard_template() {
    log_info "Creating monitoring dashboard template..."
    
    cat > "$MONITORING_DIR/dashboards/grafana-dashboard.json" << 'EOF'
{
  "dashboard": {
    "id": null,
    "title": "Superintelligence Dependency Manager",
    "description": "Monitoring dashboard for Superintelligence Dependency Manager",
    "tags": ["superint-dep-manager", "dependencies", "monitoring"],
    "timezone": "browser",
    "panels": [
      {
        "id": 1,
        "title": "Service Status",
        "type": "stat",
        "targets": [
          {
            "expr": "up{job=\"superint-dep-manager\"}",
            "legendFormat": "Service Status"
          }
        ]
      },
      {
        "id": 2,
        "title": "CPU Usage",
        "type": "graph",
        "targets": [
          {
            "expr": "process_cpu_seconds_total{job=\"superint-dep-manager\"}",
            "legendFormat": "CPU Usage"
          }
        ]
      },
      {
        "id": 3,
        "title": "Memory Usage",
        "type": "graph",
        "targets": [
          {
            "expr": "process_resident_memory_bytes{job=\"superint-dep-manager\"}",
            "legendFormat": "Memory Usage"
          }
        ]
      },
      {
        "id": 4,
        "title": "Dependencies Scanned",
        "type": "graph",
        "targets": [
          {
            "expr": "ai_dep_manager_dependencies_scanned_total",
            "legendFormat": "Dependencies Scanned"
          }
        ]
      },
      {
        "id": 5,
        "title": "Updates Applied",
        "type": "graph",
        "targets": [
          {
            "expr": "ai_dep_manager_updates_applied_total",
            "legendFormat": "Updates Applied"
          }
        ]
      }
    ],
    "time": {
      "from": "now-1h",
      "to": "now"
    },
    "refresh": "30s"
  }
}
EOF

    log_success "Monitoring dashboard template created"
}

# Main monitoring setup function
main() {
    log_info "Setting up Superintelligence Dependency Manager production monitoring..."
    
    create_monitoring_structure
    create_health_check
    create_metrics_script
    create_alert_script
    setup_cron_jobs
    create_log_config
    create_dashboard_template
    
    log_success "🎉 Production monitoring setup completed!"
    log_info "Health checks: Every 5 minutes"
    log_info "Metrics collection: Every minute"
    log_info "Logs: $DEPLOY_DIR/logs/"
    log_info "Monitoring scripts: $MONITORING_DIR/scripts/"
    log_info "Dashboard template: $MONITORING_DIR/dashboards/"
}

# Run main function
main "$@"
