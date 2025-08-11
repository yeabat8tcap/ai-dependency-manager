export const environment = {
  production: true,
  staging: false, // Use real API data in production
  apiUrl: 'http://localhost:8081/api', // Go backend API
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
    autoRefreshInterval: 60000, // 1 minute for production
    maxDependenciesPerPage: 50, // Smaller page size for production
    chartAnimations: false, // Disabled for better performance
    enableTooltips: true
  },
  api: {
    timeout: 60000, // 1 minute for production
    retryAttempts: 5,
    retryDelay: 2000 // 2 seconds
  },
  websocket: {
    reconnectInterval: 10000, // 10 seconds
    maxReconnectAttempts: 20,
    heartbeatInterval: 60000 // 1 minute
  },
  logging: {
    level: 'info',
    enableConsole: false,
    enableRemote: true,
    maxLogEntries: 500,
    enablePerformanceLogging: true,
    enableUserActionLogging: true,
    enableApiLogging: true,
    enableErrorTracking: true
  }
};
