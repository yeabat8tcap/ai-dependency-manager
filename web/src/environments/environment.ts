export const environment = {
  production: true,
  staging: false, // Use mock data in staging
  apiUrl: 'http://localhost:8081/api', // Updated to match Go backend port
  wsUrl: 'ws://localhost:8081/ws',
  appName: 'AI Dependency Manager',
  version: '1.0.0',
  dataSource: 'api', // 'mock' for staging, 'api' for production
  features: {
    realTimeUpdates: true,
    aiAnalysis: true,
    securityScanning: true,
    policyManagement: true,
    analytics: true,
    notifications: true
  },
  ui: {
    theme: 'light',
    autoRefreshInterval: 30000, // 30 seconds
    maxDependenciesPerPage: 100,
    chartAnimations: true,
    enableTooltips: true
  },
  api: {
    timeout: 30000, // 30 seconds
    retryAttempts: 3,
    retryDelay: 1000 // 1 second
  },
  websocket: {
    reconnectInterval: 5000, // 5 seconds
    maxReconnectAttempts: 10,
    heartbeatInterval: 30000 // 30 seconds
  },
  logging: {
    level: 'debug',
    enableConsole: true,
    enableRemote: false,
    maxLogEntries: 1000,
    enablePerformanceLogging: true,
    enableUserActionLogging: true,
    enableApiLogging: true,
    enableErrorTracking: true
  }
};
