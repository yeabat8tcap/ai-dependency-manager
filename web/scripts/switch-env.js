#!/usr/bin/env node

const fs = require('fs');
const path = require('path');

const args = process.argv.slice(2);
const targetEnv = args[0];

if (!targetEnv || !['staging', 'production'].includes(targetEnv)) {
  console.error('Usage: node switch-env.js [staging|production]');
  console.error('  staging    - Use mock data for development');
  console.error('  production - Use real API endpoints');
  process.exit(1);
}

const envPath = path.join(__dirname, '../src/environments/environment.ts');
const envContent = fs.readFileSync(envPath, 'utf8');

let newContent;

if (targetEnv === 'staging') {
  newContent = envContent
    .replace(/staging: false/g, 'staging: true')
    .replace(/dataSource: 'api'/g, "dataSource: 'mock'")
    .replace(/production: true/g, 'production: false');
  
  console.log('🔄 Switched to STAGING environment');
  console.log('   - Using mock data for development');
  console.log('   - Frontend will use simulated API responses');
  console.log('   - WebSocket connections will be mocked');
} else {
  newContent = envContent
    .replace(/staging: true/g, 'staging: false')
    .replace(/dataSource: 'mock'/g, "dataSource: 'api'")
    .replace(/production: false/g, 'production: true');
  
  console.log('🚀 Switched to PRODUCTION environment');
  console.log('   - Using real API endpoints');
  console.log('   - Frontend will connect to Go backend at http://localhost:8081');
  console.log('   - WebSocket connections will be established');
}

fs.writeFileSync(envPath, newContent);

console.log('\n✅ Environment configuration updated successfully!');
console.log('💡 Restart the development server to apply changes: npm start');
