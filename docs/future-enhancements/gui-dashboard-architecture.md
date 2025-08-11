# GUI Dashboard Architecture

This document outlines the architecture for a comprehensive web-based GUI dashboard for the AI Dependency Manager, providing visual interfaces for monitoring, management, and analytics.

## Table of Contents

1. [Overview](#overview)
2. [Architecture Design](#architecture-design)
3. [Frontend Architecture](#frontend-architecture)
4. [Backend API Server](#backend-api-server)
5. [Real-time Features](#real-time-features)
6. [User Interface Design](#user-interface-design)
7. [Security and Authentication](#security-and-authentication)
8. [Mobile Responsiveness](#mobile-responsiveness)
9. [Implementation Plan](#implementation-plan)

## Overview

The GUI Dashboard will provide a modern, intuitive web interface for managing dependencies, viewing analytics, configuring policies, and monitoring system health. It will complement the existing CLI interface with visual tools for complex operations.

### Key Features

- **Project Dashboard**: Visual overview of all managed projects
- **Dependency Visualization**: Interactive dependency graphs and trees
- **Real-time Monitoring**: Live updates and notifications
- **Analytics Dashboard**: Charts, trends, and insights
- **Policy Management**: Visual policy builder and editor
- **Security Center**: Vulnerability tracking and security metrics
- **Configuration Management**: GUI-based system configuration
- **Report Generation**: Interactive report builder

## Architecture Design

### High-Level Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Web Frontend  │    │   API Gateway   │    │  Core Services  │
├─────────────────┤    ├─────────────────┤    ├─────────────────┤
│ • React/Vue.js  │◄──►│ • Authentication│◄──►│ • Scanner       │
│ • TypeScript    │    │ • Rate Limiting │    │ • Update Manager│
│ • Chart.js/D3   │    │ • API Routing   │    │ • Security      │
│ • WebSockets    │    │ • Request/Response│  │ • Notifications │
│ • PWA Support   │    │ • Caching       │    │ • Reporting     │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                        │                        │
         ▼                        ▼                        ▼
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Static CDN    │    │   Load Balancer │    │    Database     │
│ • Assets        │    │ • SSL Termination│   │ • PostgreSQL    │
│ • Caching       │    │ • Health Checks │    │ • Redis Cache   │
│ • Compression   │    │ • Auto Scaling  │    │ • Time Series DB│
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

### Component Interaction

```
Frontend Components ──► API Gateway ──► Service Layer ──► Database
       │                     │              │               │
       │                     │              │               │
   WebSocket ◄─────────── Event Bus ◄── Background Agent ──┘
   Connection               │
       │                   │
       ▼                   ▼
  Real-time UI      Notification
   Updates           System
```

## Frontend Architecture

### Technology Stack

**Core Framework**: React 18 with TypeScript
**State Management**: Redux Toolkit + RTK Query
**UI Components**: Material-UI (MUI) or Ant Design
**Charts/Visualization**: Chart.js, D3.js, React Flow
**Real-time**: Socket.IO client
**Build Tool**: Vite
**Testing**: Jest + React Testing Library

### Project Structure

```
frontend/
├── src/
│   ├── components/           # Reusable UI components
│   │   ├── common/          # Generic components
│   │   ├── charts/          # Chart components
│   │   ├── forms/           # Form components
│   │   └── layout/          # Layout components
│   ├── pages/               # Page components
│   │   ├── Dashboard/       # Main dashboard
│   │   ├── Projects/        # Project management
│   │   ├── Dependencies/    # Dependency views
│   │   ├── Security/        # Security center
│   │   ├── Analytics/       # Analytics dashboard
│   │   ├── Policies/        # Policy management
│   │   └── Settings/        # Configuration
│   ├── hooks/               # Custom React hooks
│   ├── services/            # API service layer
│   ├── store/               # Redux store
│   ├── types/               # TypeScript definitions
│   ├── utils/               # Utility functions
│   └── assets/              # Static assets
├── public/                  # Public assets
├── tests/                   # Test files
└── docs/                    # Documentation
```

### State Management

```typescript
// Redux store structure
interface RootState {
  auth: AuthState;
  projects: ProjectsState;
  dependencies: DependenciesState;
  security: SecurityState;
  analytics: AnalyticsState;
  policies: PoliciesState;
  notifications: NotificationsState;
  ui: UIState;
}

// API slice using RTK Query
export const apiSlice = createApi({
  reducerPath: 'api',
  baseQuery: fetchBaseQuery({
    baseUrl: '/api/v1',
    prepareHeaders: (headers, { getState }) => {
      const token = (getState() as RootState).auth.token;
      if (token) {
        headers.set('authorization', `Bearer ${token}`);
      }
      return headers;
    },
  }),
  tagTypes: ['Project', 'Dependency', 'Vulnerability', 'Policy'],
  endpoints: (builder) => ({
    getProjects: builder.query<Project[], void>({
      query: () => 'projects',
      providesTags: ['Project'],
    }),
    getDependencies: builder.query<Dependency[], string>({
      query: (projectId) => `projects/${projectId}/dependencies`,
      providesTags: ['Dependency'],
    }),
    // ... more endpoints
  }),
});
```

### Component Architecture

```typescript
// Main Dashboard Component
interface DashboardProps {
  projects: Project[];
  metrics: SystemMetrics;
  notifications: Notification[];
}

const Dashboard: React.FC<DashboardProps> = ({ projects, metrics, notifications }) => {
  const [selectedProject, setSelectedProject] = useState<string | null>(null);
  const [timeRange, setTimeRange] = useState<TimeRange>('7d');

  return (
    <DashboardLayout>
      <Grid container spacing={3}>
        <Grid item xs={12} md={8}>
          <ProjectOverview 
            projects={projects}
            onProjectSelect={setSelectedProject}
          />
        </Grid>
        <Grid item xs={12} md={4}>
          <SystemHealth metrics={metrics} />
          <RecentActivity notifications={notifications} />
        </Grid>
        <Grid item xs={12}>
          <DependencyTrends 
            projectId={selectedProject}
            timeRange={timeRange}
          />
        </Grid>
      </Grid>
    </DashboardLayout>
  );
};

// Dependency Visualization Component
const DependencyGraph: React.FC<{ projectId: string }> = ({ projectId }) => {
  const { data: dependencies } = useGetDependenciesQuery(projectId);
  const [selectedNode, setSelectedNode] = useState<string | null>(null);

  const graphData = useMemo(() => {
    return transformDependenciesToGraph(dependencies);
  }, [dependencies]);

  return (
    <Card>
      <CardHeader title="Dependency Graph" />
      <CardContent>
        <ReactFlow
          nodes={graphData.nodes}
          edges={graphData.edges}
          onNodeClick={(event, node) => setSelectedNode(node.id)}
          fitView
        >
          <Controls />
          <MiniMap />
          <Background />
        </ReactFlow>
        {selectedNode && (
          <DependencyDetails nodeId={selectedNode} />
        )}
      </CardContent>
    </Card>
  );
};
```

## Backend API Server

### API Architecture

```go
// API Server Structure
type APIServer struct {
    router      *gin.Engine
    services    *ServiceContainer
    auth        *AuthService
    websocket   *WebSocketManager
    middleware  []gin.HandlerFunc
}

type ServiceContainer struct {
    Scanner     services.ScannerService
    Updater     services.UpdateService
    Security    services.SecurityService
    Analytics   services.AnalyticsService
    Policies    services.PolicyService
    Projects    services.ProjectService
}

// API Routes
func (s *APIServer) setupRoutes() {
    api := s.router.Group("/api/v1")
    api.Use(s.middleware...)

    // Authentication
    auth := api.Group("/auth")
    {
        auth.POST("/login", s.handleLogin)
        auth.POST("/logout", s.handleLogout)
        auth.POST("/refresh", s.handleRefresh)
    }

    // Projects
    projects := api.Group("/projects")
    projects.Use(s.auth.RequireAuth())
    {
        projects.GET("", s.handleGetProjects)
        projects.POST("", s.handleCreateProject)
        projects.GET("/:id", s.handleGetProject)
        projects.PUT("/:id", s.handleUpdateProject)
        projects.DELETE("/:id", s.handleDeleteProject)
        projects.GET("/:id/dependencies", s.handleGetDependencies)
        projects.POST("/:id/scan", s.handleScanProject)
    }

    // Dependencies
    dependencies := api.Group("/dependencies")
    dependencies.Use(s.auth.RequireAuth())
    {
        dependencies.GET("/:id", s.handleGetDependency)
        dependencies.GET("/:id/updates", s.handleGetUpdates)
        dependencies.POST("/:id/update", s.handleUpdateDependency)
    }

    // Security
    security := api.Group("/security")
    security.Use(s.auth.RequireAuth())
    {
        security.GET("/vulnerabilities", s.handleGetVulnerabilities)
        security.GET("/scan/:projectId", s.handleSecurityScan)
        security.POST("/rules", s.handleCreateSecurityRule)
    }

    // Analytics
    analytics := api.Group("/analytics")
    analytics.Use(s.auth.RequireAuth())
    {
        analytics.GET("/dashboard", s.handleGetDashboardData)
        analytics.GET("/trends", s.handleGetTrends)
        analytics.GET("/reports", s.handleGetReports)
    }

    // WebSocket endpoint
    api.GET("/ws", s.handleWebSocket)
}
```

### API Handlers

```go
// Project handlers
func (s *APIServer) handleGetProjects(c *gin.Context) {
    userID := s.auth.GetUserID(c)
    
    projects, err := s.services.Projects.GetUserProjects(c.Request.Context(), userID)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    response := make([]ProjectResponse, len(projects))
    for i, project := range projects {
        response[i] = ProjectResponse{
            ID:          project.ID,
            Name:        project.Name,
            Path:        project.Path,
            Type:        project.Type,
            Status:      project.Status,
            LastScan:    project.LastScan,
            UpdateCount: project.AvailableUpdates,
            VulnCount:   project.VulnerabilityCount,
        }
    }

    c.JSON(http.StatusOK, gin.H{"projects": response})
}

func (s *APIServer) handleScanProject(c *gin.Context) {
    projectID := c.Param("id")
    userID := s.auth.GetUserID(c)

    // Validate project ownership
    if !s.services.Projects.UserOwnsProject(c.Request.Context(), userID, projectID) {
        c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
        return
    }

    // Start async scan
    scanID, err := s.services.Scanner.StartScan(c.Request.Context(), projectID)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    // Send real-time update
    s.websocket.SendToUser(userID, WebSocketMessage{
        Type: "scan_started",
        Data: map[string]interface{}{
            "project_id": projectID,
            "scan_id":    scanID,
        },
    })

    c.JSON(http.StatusAccepted, gin.H{
        "scan_id": scanID,
        "status":  "started",
    })
}
```

### Real-time WebSocket

```go
type WebSocketManager struct {
    clients    map[string]*WebSocketClient
    register   chan *WebSocketClient
    unregister chan *WebSocketClient
    broadcast  chan WebSocketMessage
    mutex      sync.RWMutex
}

type WebSocketClient struct {
    ID     string
    UserID string
    Conn   *websocket.Conn
    Send   chan WebSocketMessage
}

type WebSocketMessage struct {
    Type string                 `json:"type"`
    Data map[string]interface{} `json:"data"`
}

func (wm *WebSocketManager) Run() {
    for {
        select {
        case client := <-wm.register:
            wm.mutex.Lock()
            wm.clients[client.ID] = client
            wm.mutex.Unlock()
            
        case client := <-wm.unregister:
            wm.mutex.Lock()
            if _, ok := wm.clients[client.ID]; ok {
                delete(wm.clients, client.ID)
                close(client.Send)
            }
            wm.mutex.Unlock()
            
        case message := <-wm.broadcast:
            wm.mutex.RLock()
            for _, client := range wm.clients {
                select {
                case client.Send <- message:
                default:
                    close(client.Send)
                    delete(wm.clients, client.ID)
                }
            }
            wm.mutex.RUnlock()
        }
    }
}

func (wm *WebSocketManager) SendToUser(userID string, message WebSocketMessage) {
    wm.mutex.RLock()
    defer wm.mutex.RUnlock()
    
    for _, client := range wm.clients {
        if client.UserID == userID {
            select {
            case client.Send <- message:
            default:
                close(client.Send)
                delete(wm.clients, client.ID)
            }
        }
    }
}
```

## Real-time Features

### Live Updates

```typescript
// WebSocket hook for real-time updates
export const useWebSocket = (userId: string) => {
  const [socket, setSocket] = useState<Socket | null>(null);
  const dispatch = useAppDispatch();

  useEffect(() => {
    const newSocket = io('/api/v1/ws', {
      auth: { userId },
      transports: ['websocket'],
    });

    newSocket.on('connect', () => {
      console.log('WebSocket connected');
    });

    newSocket.on('scan_started', (data) => {
      dispatch(addNotification({
        type: 'info',
        message: `Scan started for project ${data.project_id}`,
      }));
    });

    newSocket.on('scan_completed', (data) => {
      dispatch(updateProjectScanStatus(data));
      dispatch(addNotification({
        type: 'success',
        message: `Scan completed for project ${data.project_id}`,
      }));
    });

    newSocket.on('vulnerability_found', (data) => {
      dispatch(addVulnerability(data));
      dispatch(addNotification({
        type: 'warning',
        message: `New vulnerability found: ${data.package_name}`,
      }));
    });

    newSocket.on('update_available', (data) => {
      dispatch(addAvailableUpdate(data));
    });

    setSocket(newSocket);

    return () => {
      newSocket.close();
    };
  }, [userId, dispatch]);

  return socket;
};
```

### Live Dashboard

```typescript
const LiveDashboard: React.FC = () => {
  const { user } = useSelector((state: RootState) => state.auth);
  const socket = useWebSocket(user?.id || '');
  const [metrics, setMetrics] = useState<SystemMetrics | null>(null);

  useEffect(() => {
    if (!socket) return;

    socket.on('metrics_update', (data: SystemMetrics) => {
      setMetrics(data);
    });

    // Request initial metrics
    socket.emit('subscribe_metrics');

    return () => {
      socket.off('metrics_update');
      socket.emit('unsubscribe_metrics');
    };
  }, [socket]);

  return (
    <Grid container spacing={3}>
      <Grid item xs={12} md={6}>
        <MetricsCard
          title="Active Scans"
          value={metrics?.activeScans || 0}
          trend={metrics?.scanTrend}
        />
      </Grid>
      <Grid item xs={12} md={6}>
        <MetricsCard
          title="Vulnerabilities"
          value={metrics?.vulnerabilities || 0}
          trend={metrics?.vulnTrend}
          color="error"
        />
      </Grid>
      <Grid item xs={12}>
        <LiveActivityFeed />
      </Grid>
    </Grid>
  );
};
```

## User Interface Design

### Dashboard Layout

```typescript
const DashboardLayout: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const [sidebarOpen, setSidebarOpen] = useState(true);
  const theme = useTheme();
  const isMobile = useMediaQuery(theme.breakpoints.down('md'));

  return (
    <Box sx={{ display: 'flex', minHeight: '100vh' }}>
      <AppBar position="fixed" sx={{ zIndex: theme.zIndex.drawer + 1 }}>
        <Toolbar>
          <IconButton
            color="inherit"
            edge="start"
            onClick={() => setSidebarOpen(!sidebarOpen)}
            sx={{ mr: 2 }}
          >
            <MenuIcon />
          </IconButton>
          <Typography variant="h6" noWrap component="div" sx={{ flexGrow: 1 }}>
            AI Dependency Manager
          </Typography>
          <NotificationButton />
          <UserMenu />
        </Toolbar>
      </AppBar>

      <Drawer
        variant={isMobile ? 'temporary' : 'persistent'}
        open={sidebarOpen}
        onClose={() => setSidebarOpen(false)}
        sx={{
          width: 240,
          flexShrink: 0,
          '& .MuiDrawer-paper': {
            width: 240,
            boxSizing: 'border-box',
          },
        }}
      >
        <Toolbar />
        <NavigationMenu />
      </Drawer>

      <Box
        component="main"
        sx={{
          flexGrow: 1,
          p: 3,
          width: { sm: `calc(100% - ${sidebarOpen ? 240 : 0}px)` },
        }}
      >
        <Toolbar />
        {children}
      </Box>
    </Box>
  );
};
```

### Interactive Charts

```typescript
const DependencyTrendsChart: React.FC<{ data: TrendData[] }> = ({ data }) => {
  const chartRef = useRef<HTMLCanvasElement>(null);

  useEffect(() => {
    if (!chartRef.current) return;

    const ctx = chartRef.current.getContext('2d');
    if (!ctx) return;

    const chart = new Chart(ctx, {
      type: 'line',
      data: {
        labels: data.map(d => d.date),
        datasets: [
          {
            label: 'Dependencies',
            data: data.map(d => d.totalDependencies),
            borderColor: 'rgb(75, 192, 192)',
            backgroundColor: 'rgba(75, 192, 192, 0.2)',
          },
          {
            label: 'Outdated',
            data: data.map(d => d.outdatedDependencies),
            borderColor: 'rgb(255, 99, 132)',
            backgroundColor: 'rgba(255, 99, 132, 0.2)',
          },
        ],
      },
      options: {
        responsive: true,
        interaction: {
          mode: 'index' as const,
          intersect: false,
        },
        scales: {
          x: {
            display: true,
            title: {
              display: true,
              text: 'Date',
            },
          },
          y: {
            display: true,
            title: {
              display: true,
              text: 'Count',
            },
          },
        },
      },
    });

    return () => {
      chart.destroy();
    };
  }, [data]);

  return <canvas ref={chartRef} />;
};
```

## Security and Authentication

### JWT Authentication

```go
type AuthService struct {
    jwtSecret     []byte
    tokenExpiry   time.Duration
    refreshExpiry time.Duration
    userService   UserService
}

func (a *AuthService) GenerateTokens(userID string) (*TokenPair, error) {
    accessToken, err := a.generateAccessToken(userID)
    if err != nil {
        return nil, err
    }

    refreshToken, err := a.generateRefreshToken(userID)
    if err != nil {
        return nil, err
    }

    return &TokenPair{
        AccessToken:  accessToken,
        RefreshToken: refreshToken,
        ExpiresIn:    int(a.tokenExpiry.Seconds()),
    }, nil
}

func (a *AuthService) RequireAuth() gin.HandlerFunc {
    return func(c *gin.Context) {
        authHeader := c.GetHeader("Authorization")
        if authHeader == "" {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
            c.Abort()
            return
        }

        tokenString := strings.TrimPrefix(authHeader, "Bearer ")
        claims, err := a.validateToken(tokenString)
        if err != nil {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
            c.Abort()
            return
        }

        c.Set("user_id", claims.UserID)
        c.Next()
    }
}
```

### Frontend Authentication

```typescript
// Auth slice
const authSlice = createSlice({
  name: 'auth',
  initialState: {
    user: null as User | null,
    token: localStorage.getItem('token'),
    isAuthenticated: false,
    loading: false,
  },
  reducers: {
    loginStart: (state) => {
      state.loading = true;
    },
    loginSuccess: (state, action) => {
      state.user = action.payload.user;
      state.token = action.payload.token;
      state.isAuthenticated = true;
      state.loading = false;
      localStorage.setItem('token', action.payload.token);
    },
    loginFailure: (state) => {
      state.loading = false;
    },
    logout: (state) => {
      state.user = null;
      state.token = null;
      state.isAuthenticated = false;
      localStorage.removeItem('token');
    },
  },
});

// Protected route component
const ProtectedRoute: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const { isAuthenticated, loading } = useSelector((state: RootState) => state.auth);

  if (loading) {
    return <CircularProgress />;
  }

  if (!isAuthenticated) {
    return <Navigate to="/login" replace />;
  }

  return <>{children}</>;
};
```

## Mobile Responsiveness

### Responsive Design

```typescript
const ResponsiveGrid: React.FC = () => {
  const theme = useTheme();
  const isMobile = useMediaQuery(theme.breakpoints.down('md'));
  const isTablet = useMediaQuery(theme.breakpoints.down('lg'));

  return (
    <Grid container spacing={isMobile ? 2 : 3}>
      <Grid item xs={12} md={isMobile ? 12 : 8}>
        <ProjectOverview />
      </Grid>
      <Grid item xs={12} md={isMobile ? 12 : 4}>
        <SystemHealth />
      </Grid>
      {!isMobile && (
        <Grid item xs={12}>
          <DependencyGraph />
        </Grid>
      )}
    </Grid>
  );
};

// Mobile-optimized navigation
const MobileNavigation: React.FC = () => {
  const [selectedTab, setSelectedTab] = useState(0);

  return (
    <Paper sx={{ position: 'fixed', bottom: 0, left: 0, right: 0 }} elevation={3}>
      <BottomNavigation
        value={selectedTab}
        onChange={(event, newValue) => setSelectedTab(newValue)}
      >
        <BottomNavigationAction label="Dashboard" icon={<DashboardIcon />} />
        <BottomNavigationAction label="Projects" icon={<FolderIcon />} />
        <BottomNavigationAction label="Security" icon={<SecurityIcon />} />
        <BottomNavigationAction label="Settings" icon={<SettingsIcon />} />
      </BottomNavigation>
    </Paper>
  );
};
```

## Implementation Plan

### Phase 1: Foundation (Months 1-2)
- [ ] Set up React frontend with TypeScript
- [ ] Implement basic API server with authentication
- [ ] Create core layout and navigation components
- [ ] Set up WebSocket infrastructure
- [ ] Implement basic project listing and details

### Phase 2: Core Features (Months 3-4)
- [ ] Build dependency visualization components
- [ ] Implement real-time dashboard updates
- [ ] Create security center interface
- [ ] Add analytics and reporting views
- [ ] Implement policy management UI

### Phase 3: Advanced Features (Months 5-6)
- [ ] Add interactive dependency graphs
- [ ] Implement advanced filtering and search
- [ ] Create batch operation interfaces
- [ ] Add export/import functionality
- [ ] Implement user management system

### Phase 4: Mobile and PWA (Months 7-8)
- [ ] Optimize for mobile devices
- [ ] Implement Progressive Web App features
- [ ] Add offline capabilities
- [ ] Create mobile-specific navigation
- [ ] Implement push notifications

### Phase 5: Production Ready (Months 9-10)
- [ ] Performance optimization
- [ ] Comprehensive testing
- [ ] Security hardening
- [ ] Documentation and deployment guides
- [ ] Monitoring and analytics integration

This GUI dashboard architecture provides a comprehensive foundation for building a modern, user-friendly web interface that complements the CLI tool and makes dependency management accessible to all team members.
