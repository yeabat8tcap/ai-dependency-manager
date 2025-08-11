import { TestBed } from '@angular/core/testing';
import { WebSocketService, SystemStatusMessage, ScanProgressMessage, NotificationMessage } from './websocket.service';

describe('WebSocketService', () => {
  let service: WebSocketService;
  let mockWebSocket: jasmine.SpyObj<WebSocket>;

  beforeEach(() => {
    // Create a mock WebSocket
    mockWebSocket = jasmine.createSpyObj('WebSocket', ['send', 'close'], {
      readyState: WebSocket.CONNECTING,
      onopen: null,
      onmessage: null,
      onclose: null,
      onerror: null
    });

    // Mock the WebSocket constructor
    spyOn(window, 'WebSocket').and.returnValue(mockWebSocket);

    TestBed.configureTestingModule({
      providers: [WebSocketService]
    });
    service = TestBed.inject(WebSocketService);
  });

  afterEach(() => {
    service.disconnect();
  });

  it('should be created', () => {
    expect(service).toBeTruthy();
  });

  describe('Connection Management', () => {
    it('should connect to WebSocket server', () => {
      service.connect();
      expect(window.WebSocket).toHaveBeenCalledWith('ws://localhost:8080/ws');
      expect(service['ws']).toBe(mockWebSocket);
    });

    it('should handle connection open event', () => {
      service.connect();
      
      // Simulate connection open
      mockWebSocket.readyState = WebSocket.OPEN;
      if (mockWebSocket.onopen) {
        mockWebSocket.onopen({} as Event);
      }

      expect(service.isConnected()).toBe(true);
    });

    it('should handle connection close event', () => {
      service.connect();
      
      // Simulate connection close
      mockWebSocket.readyState = WebSocket.CLOSED;
      if (mockWebSocket.onclose) {
        mockWebSocket.onclose({} as CloseEvent);
      }

      expect(service.isConnected()).toBe(false);
    });

    it('should handle connection error', () => {
      spyOn(console, 'error');
      service.connect();
      
      // Simulate connection error
      if (mockWebSocket.onerror) {
        mockWebSocket.onerror({} as Event);
      }

      expect(console.error).toHaveBeenCalledWith('WebSocket error:', jasmine.any(Object));
    });

    it('should disconnect properly', () => {
      service.connect();
      service.disconnect();
      
      expect(mockWebSocket.close).toHaveBeenCalled();
      expect(service.isConnected()).toBe(false);
    });

    it('should auto-reconnect on connection loss', (done) => {
      service.connect();
      
      // Simulate connection close with auto-reconnect
      mockWebSocket.readyState = WebSocket.CLOSED;
      if (mockWebSocket.onclose) {
        mockWebSocket.onclose({ code: 1006 } as CloseEvent);
      }

      // Wait for reconnection attempt
      setTimeout(() => {
        expect(window.WebSocket).toHaveBeenCalledTimes(2);
        done();
      }, 1100); // Slightly more than the 1000ms reconnection delay
    });
  });

  describe('Message Handling', () => {
    beforeEach(() => {
      service.connect();
      mockWebSocket.readyState = WebSocket.OPEN;
    });

    it('should handle system status messages', (done) => {
      const mockStatusMessage: SystemStatusMessage = {
        type: 'system_status',
        status: 'healthy',
        timestamp: new Date().toISOString(),
        details: { uptime: '24h', memory: '512MB' }
      };

      service.getSystemStatus().subscribe(message => {
        expect(message).toEqual(mockStatusMessage);
        done();
      });

      // Simulate receiving a message
      if (mockWebSocket.onmessage) {
        mockWebSocket.onmessage({
          data: JSON.stringify(mockStatusMessage)
        } as MessageEvent);
      }
    });

    it('should handle scan progress messages', (done) => {
      const mockProgressMessage: ScanProgressMessage = {
        type: 'scan_progress',
        projectId: 1,
        status: 'in_progress',
        progress: 50,
        message: 'Scanning dependencies...',
        timestamp: new Date().toISOString()
      };

      service.getScanProgress().subscribe(message => {
        expect(message).toEqual(mockProgressMessage);
        done();
      });

      // Simulate receiving a message
      if (mockWebSocket.onmessage) {
        mockWebSocket.onmessage({
          data: JSON.stringify(mockProgressMessage)
        } as MessageEvent);
      }
    });

    it('should handle notification messages', (done) => {
      const mockNotificationMessage: NotificationMessage = {
        type: 'notification',
        level: 'warning',
        title: 'Security Alert',
        message: 'New vulnerability detected',
        timestamp: new Date().toISOString(),
        projectId: 1
      };

      service.getNotifications().subscribe(message => {
        expect(message).toEqual(mockNotificationMessage);
        done();
      });

      // Simulate receiving a message
      if (mockWebSocket.onmessage) {
        mockWebSocket.onmessage({
          data: JSON.stringify(mockNotificationMessage)
        } as MessageEvent);
      }
    });

    it('should handle invalid JSON messages gracefully', () => {
      spyOn(console, 'error');

      // Simulate receiving invalid JSON
      if (mockWebSocket.onmessage) {
        mockWebSocket.onmessage({
          data: 'invalid json'
        } as MessageEvent);
      }

      expect(console.error).toHaveBeenCalledWith('Error parsing WebSocket message:', jasmine.any(Error));
    });

    it('should handle unknown message types', () => {
      spyOn(console, 'warn');

      const unknownMessage = {
        type: 'unknown_type',
        data: 'test'
      };

      // Simulate receiving unknown message type
      if (mockWebSocket.onmessage) {
        mockWebSocket.onmessage({
          data: JSON.stringify(unknownMessage)
        } as MessageEvent);
      }

      expect(console.warn).toHaveBeenCalledWith('Unknown message type:', 'unknown_type');
    });
  });

  describe('Message Sending', () => {
    beforeEach(() => {
      service.connect();
      mockWebSocket.readyState = WebSocket.OPEN;
    });

    it('should send messages when connected', () => {
      const message = { type: 'test', data: 'hello' };
      
      service.sendMessage(message);
      
      expect(mockWebSocket.send).toHaveBeenCalledWith(JSON.stringify(message));
    });

    it('should queue messages when not connected', () => {
      mockWebSocket.readyState = WebSocket.CONNECTING;
      const message = { type: 'test', data: 'hello' };
      
      service.sendMessage(message);
      
      expect(mockWebSocket.send).not.toHaveBeenCalled();
      
      // Simulate connection open
      mockWebSocket.readyState = WebSocket.OPEN;
      if (mockWebSocket.onopen) {
        mockWebSocket.onopen({} as Event);
      }
      
      expect(mockWebSocket.send).toHaveBeenCalledWith(JSON.stringify(message));
    });

    it('should subscribe to project updates', () => {
      service.subscribeToProject(1);
      
      expect(mockWebSocket.send).toHaveBeenCalledWith(JSON.stringify({
        type: 'subscribe',
        projectId: 1
      }));
    });

    it('should unsubscribe from project updates', () => {
      service.unsubscribeFromProject(1);
      
      expect(mockWebSocket.send).toHaveBeenCalledWith(JSON.stringify({
        type: 'unsubscribe',
        projectId: 1
      }));
    });
  });

  describe('Connection State', () => {
    it('should report correct connection state', () => {
      expect(service.isConnected()).toBe(false);
      
      service.connect();
      mockWebSocket.readyState = WebSocket.OPEN;
      expect(service.isConnected()).toBe(true);
      
      mockWebSocket.readyState = WebSocket.CLOSED;
      expect(service.isConnected()).toBe(false);
    });

    it('should handle multiple connection attempts gracefully', () => {
      service.connect();
      service.connect(); // Second call should not create new connection
      
      expect(window.WebSocket).toHaveBeenCalledTimes(1);
    });
  });

  describe('Error Scenarios', () => {
    it('should handle WebSocket constructor errors', () => {
      (window.WebSocket as jasmine.Spy).and.throwError('Connection failed');
      spyOn(console, 'error');
      
      service.connect();
      
      expect(console.error).toHaveBeenCalledWith('Failed to connect to WebSocket:', jasmine.any(Error));
    });

    it('should handle send errors when WebSocket is null', () => {
      spyOn(console, 'warn');
      
      service.sendMessage({ type: 'test' });
      
      expect(console.warn).toHaveBeenCalledWith('WebSocket not connected, queueing message');
    });
  });

  describe('Cleanup', () => {
    it('should clean up resources on service destruction', () => {
      service.connect();
      
      service.ngOnDestroy();
      
      expect(mockWebSocket.close).toHaveBeenCalled();
    });

    it('should complete all observables on destroy', () => {
      const systemStatusSpy = jasmine.createSpy('systemStatus');
      const scanProgressSpy = jasmine.createSpy('scanProgress');
      const notificationsSpy = jasmine.createSpy('notifications');

      service.getSystemStatus().subscribe(systemStatusSpy);
      service.getScanProgress().subscribe(scanProgressSpy);
      service.getNotifications().subscribe(notificationsSpy);

      service.ngOnDestroy();

      // Observables should complete and not emit further values
      expect(systemStatusSpy).not.toHaveBeenCalled();
      expect(scanProgressSpy).not.toHaveBeenCalled();
      expect(notificationsSpy).not.toHaveBeenCalled();
    });
  });
});
