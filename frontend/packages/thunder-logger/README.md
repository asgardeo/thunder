# @thunder/logger

Universal TypeScript logging library with pluggable transports for Thunder applications. Works seamlessly with vanilla JavaScript, React, Vite, and Next.js applications.

## Features

- **🔌 Pluggable Transports** - Send logs to multiple destinations (console, stdout, HTTP, Sentry, etc.)
- **🌐 Universal** - Works in browser, Node.js, Vite apps, and Next.js (SSR-safe)
- **⚛️ React Integration** - Optional React Context and hooks for React applications
- **📊 Structured Logging** - Key-value pairs and contextual fields
- **🎯 Log Levels** - debug, info, warn, error with filtering
- **🔒 Type-Safe** - Full TypeScript support with autocomplete
- **🎨 Environment-Aware** - Auto-detects development vs production
- **⚡ Performance** - Lazy evaluation with `isDebugEnabled()` checks
- **🛠️ Extensible** - Easy to create custom transports

## Installation

Since this is a workspace package, install dependencies from the root:

```bash
pnpm add -D @thunder/logger
```

## Quick Start

### Vanilla JavaScript / TypeScript

```typescript
import { Logger, ConsoleTransport } from '@thunder/logger';

// Create logger with console transport
const logger = new Logger({
  level: 'info',
  transports: [new ConsoleTransport()],
});

// Use the logger
logger.info('Application started');
logger.warn('Deprecated API usage', { api: '/old-endpoint' });
logger.error('Failed to fetch data', { error: err });
```

### React Application (Vite)

```tsx
import { LoggerProvider, useLogger } from '@thunder/logger/react';
import { ConsoleTransport } from '@thunder/logger';

// 1. Wrap your app with LoggerProvider
function App() {
  return (
    <LoggerProvider
      logger={{
        level: import.meta.env.DEV ? 'debug' : 'info',
        transports: [new ConsoleTransport()],
      }}
    >
      <MyComponent />
    </LoggerProvider>
  );
}

// 2. Use logger in components
function MyComponent() {
  const logger = useLogger();

  const handleClick = () => {
    logger.info('Button clicked', { timestamp: Date.now() });
  };

  return <button onClick={handleClick}>Click me</button>;
}
```

### Next.js Application (SSR-Safe)

```tsx
import { Logger, ConsoleTransport, StdoutTransport } from '@thunder/logger';

// Server-side (API routes, getServerSideProps)
const serverLogger = new Logger({
  level: 'info',
  transports: [new StdoutTransport()], // JSON to stdout
});

export async function getServerSideProps() {
  serverLogger.info('Fetching data', { page: 'home' });
  return { props: {} };
}

// Client-side (components)
import { LoggerProvider, useLogger } from '@thunder/logger/react';

export default function App({ Component, pageProps }) {
  return (
    <LoggerProvider
      logger={{
        level: process.env.NODE_ENV === 'development' ? 'debug' : 'info',
        transports: [new ConsoleTransport()],
      }}
    >
      <Component {...pageProps} />
    </LoggerProvider>
  );
}
```

## Core API

### Logger Class

```typescript
import { Logger, ConsoleTransport } from '@thunder/logger';

const logger = new Logger({
  level: 'info', // 'debug' | 'info' | 'warn' | 'error'
  transports: [new ConsoleTransport()],
});
```

#### Logging Methods

```typescript
// Basic logging
logger.debug('Debug message');
logger.info('Info message');
logger.warn('Warning message');
logger.error('Error message');

// With context (key-value pairs)
logger.info('User logged in', { userId: '123', email: 'user@example.com' });
logger.error('Payment failed', { orderId: 'ORD-456', amount: 99.99 });

// With Error objects
try {
  // ...
} catch (err) {
  logger.error('Operation failed', { error: err, context: 'payment-processing' });
}
```

#### Contextual Logging

Create loggers with persistent context:

```typescript
// Create logger with component context
const componentLogger = logger.withComponent('UserProfile');
componentLogger.info('Rendered'); // Logs: [UserProfile] Rendered

// Add custom fields
const requestLogger = logger.withContext({
  requestId: 'req-123',
  userId: 'user-456',
});
requestLogger.info('Processing request'); // Includes requestId and userId
```

#### Performance Optimization

```typescript
// Skip expensive operations when debug is disabled
if (logger.isDebugEnabled()) {
  const complexData = generateExpensiveDebugData();
  logger.debug('Complex debug info', { data: complexData });
}
```

### Log Levels

Log levels follow standard severity hierarchy:

- **`debug`** - Detailed information for debugging (e.g., variable values, flow steps)
- **`info`** - General informational messages (e.g., app started, user logged in)
- **`warn`** - Warning messages that don't stop execution (e.g., deprecated API usage)
- **`error`** - Error messages for failures (e.g., API errors, exceptions)

When you set a log level, only messages at that level or higher are logged:

```typescript
// level: 'warn' will log warn and error, but not info or debug
const logger = new Logger({ level: 'warn', transports: [new ConsoleTransport()] });

logger.debug('This will NOT be logged');
logger.info('This will NOT be logged');
logger.warn('This WILL be logged');
logger.error('This WILL be logged');
```

## Built-in Transports

### ConsoleTransport

Logs to browser console with styled output:

```typescript
import { ConsoleTransport } from '@thunder/logger';

const transport = new ConsoleTransport({
  level: 'debug', // Optional: override logger level for this transport
});
```

**Features:**

- Styled console output in browsers
- Color-coded by log level
- Formats objects for readability
- Supports Chrome DevTools

### StdoutTransport

Logs structured JSON to stdout (Node.js environments):

```typescript
import { StdoutTransport } from '@thunder/logger';

const transport = new StdoutTransport({
  level: 'info',
});
```

**Output format:**

```json
{"level":"info","message":"User logged in","timestamp":"2025-12-30T10:30:00.000Z","userId":"123"}
```

**Use cases:**

- Next.js server-side logging
- Node.js applications
- Docker containers (stdout captured by log aggregators)
- Production environments with log management systems

### HttpTransport

Sends logs to remote HTTP endpoints:

```typescript
import { HttpTransport } from '@thunder/logger';

const transport = new HttpTransport({
  level: 'error', // Only send errors to remote service
  endpoint: 'https://logs.example.com/api/logs',
  headers: {
    'Authorization': 'Bearer YOUR_API_KEY',
    'Content-Type': 'application/json',
  },
  batchSize: 10, // Send logs in batches
  flushInterval: 5000, // Flush every 5 seconds
  retry: true, // Enable automatic retry
  maxRetries: 3, // Maximum retry attempts
  
  // Handle dropped logs (when retry is disabled or exhausted)
  onDroppedLogs: (entries, error) => {
    // Implement fallback strategy
    console.error('Logs dropped:', entries.length, error);
    
    // Example: Store to localStorage for later recovery
    const droppedLogs = JSON.parse(localStorage.getItem('droppedLogs') || '[]');
    droppedLogs.push(...entries);
    localStorage.setItem('droppedLogs', JSON.stringify(droppedLogs));
  },
});
```

**Features:**

- Batch processing for efficiency
- Automatic retry on failure with exponential backoff
- Configurable headers for authentication
- Dropped log notifications via `onDroppedLogs` callback
- Non-blocking async writes

**Important: Resource Cleanup**

When using `HttpTransport` with batching (`batchSize > 1`), a periodic flush timer is started. To prevent memory leaks, you **must** call `logger.close()` when the logger is no longer needed:

```typescript
// Browser applications
const logger = new Logger({
  transports: [
    new HttpTransport({
      level: 'error',
      endpoint: 'https://logs.example.com/api/logs',
      batchSize: 10,
    }),
  ],
});

// Cleanup before page unload
window.addEventListener('beforeunload', () => {
  logger.close(); // Stops timers and flushes remaining logs
});

// React applications
import { useEffect } from 'react';

function App() {
  useEffect(() => {
    return () => {
      logger.close(); // Cleanup on unmount
    };
  }, []);
}

// Next.js API routes or server components
export default async function handler(req, res) {
  try {
    // Your logic
  } finally {
    await logger.close(); // Ensure cleanup
  }
}
```

**Handling Failed Logs:**

The `onDroppedLogs` callback is invoked when:
- Retry is disabled (`retry: false`) and a send fails
- All retry attempts are exhausted after network failures

Use this callback to implement fallback strategies:

```typescript
const transport = new HttpTransport({
  level: 'error',
  endpoint: 'https://logs.example.com/api/logs',
  onDroppedLogs: (entries, error) => {
    // Strategy 1: Store to localStorage
    const key = `dropped-logs-${Date.now()}`;
    localStorage.setItem(key, JSON.stringify(entries));
    
    // Strategy 2: Send to alternative endpoint
    fetch('https://backup-logs.example.com/api/logs', {
      method: 'POST',
      body: JSON.stringify({ logs: entries, error: error.message }),
    }).catch(() => {
      // Final fallback - just console log
      console.error('All log endpoints failed', entries);
    });
    
    // Strategy 3: Notify monitoring service
    if (window.Sentry) {
      window.Sentry.captureMessage('Log delivery failed', {
        level: 'warning',
        extra: { droppedCount: entries.length, error: error.message },
      });
    }
  },
});
```

## Multiple Transports

Send logs to multiple destinations simultaneously:

```typescript
import { Logger, ConsoleTransport, StdoutTransport, HttpTransport } from '@thunder/logger';

const logger = new Logger({
  level: 'debug',
  transports: [
    // Console for development
    new ConsoleTransport({ level: 'debug' }),
    
    // Stdout for Docker logs
    new StdoutTransport({ level: 'info' }),
    
    // Remote service for errors only
    new HttpTransport({
      level: 'error',
      endpoint: 'https://logs.example.com/api/logs',
    }),
  ],
});

// This debug log goes to console only
logger.debug('Rendering component');

// This info log goes to console and stdout
logger.info('User logged in', { userId: '123' });

// This error goes to all three: console, stdout, and remote HTTP
logger.error('Payment failed', { orderId: 'ORD-456' });
```

## Creating Custom Transports

Create custom transports for services like Sentry, Datadog, LogRocket, etc.

### Step 1: Extend BaseTransport

```typescript
import { BaseTransport, LogEntry, LogLevel } from '@thunder/logger';

export class SentryTransport extends BaseTransport {
  private dsn: string;

  constructor(options: { level?: LogLevel; dsn: string }) {
    super('sentry', options.level || 'error');
    this.dsn = options.dsn;
  }

  async write(entry: LogEntry): Promise<void> {
    if (!this.shouldLog(entry.level)) {
      return;
    }

    // Send to Sentry
    const Sentry = await import('@sentry/browser');
    
    if (entry.level === 'error' && entry.error) {
      Sentry.captureException(entry.error, {
        extra: entry.context,
        tags: { component: entry.component },
      });
    } else {
      Sentry.captureMessage(entry.message, {
        level: entry.level,
        extra: entry.context,
      });
    }
  }

  async flush(): Promise<void> {
    const Sentry = await import('@sentry/browser');
    await Sentry.flush(2000);
  }

  async close(): Promise<void> {
    await this.flush();
  }
}
```

### Step 2: Use Your Custom Transport

```typescript
import { Logger } from '@thunder/logger';
import { SentryTransport } from './transports/sentry';

const logger = new Logger({
  level: 'info',
  transports: [
    new ConsoleTransport(),
    new SentryTransport({
      level: 'error',
      dsn: 'https://your-dsn@sentry.io/project',
    }),
  ],
});
```

### Transport Interface Reference

**Recommended: Extend BaseTransport**

```typescript
import { BaseTransport, LogEntry, LogLevel } from '@thunder/logger';

export class MyTransport extends BaseTransport {
  constructor(options?: { level?: LogLevel }) {
    super('my-transport', options?.level || 'info');
  }

  async write(entry: LogEntry): Promise<void> {
    // Your implementation - this is the only required method
  }

  // Optional: Override if you need custom flush behavior
  async flush(): Promise<void> {
    // Flush buffered logs
  }

  // Optional: Override if you need custom cleanup
  async close(): Promise<void> {
    // Cleanup resources
  }
}
```

**Advanced: Implement Transport Interface Directly**

```typescript
import { Transport, LogEntry, LogLevel } from '@thunder/logger';

export class MyTransport implements Transport {
  getName(): string {
    return 'my-transport';
  }

  getLevel(): LogLevel {
    return 'info';
  }

  shouldLog(level: LogLevel): boolean {
    const levels: LogLevel[] = ['debug', 'info', 'warn', 'error'];
    return levels.indexOf(level) >= levels.indexOf(this.getLevel());
  }

  async write(entry: LogEntry): Promise<void> {
    // Your implementation
  }

  async flush(): Promise<void> {
    // Required: Flush buffered logs or no-op
  }

  async close(): Promise<void> {
    // Required: Cleanup resources or no-op
  }
}
```

**Why extend BaseTransport?**
- Provides default implementations of `getName()`, `getLevel()`, `shouldLog()`, `flush()`, and `close()`
- You only need to implement `write()`
- Less boilerplate code
- Consistent behavior across transports

**Type Definitions:**

```typescript
interface LogEntry {
  level: LogLevel;
  message: string;
  timestamp: Date;
  context?: Record<string, unknown>;
  component?: string;
  error?: Error;
}

type LogLevel = 'debug' | 'info' | 'warn' | 'error';
```

## External Logger Service Integrations

### Winston Integration

Integrate with Winston for powerful Node.js logging with file rotation, multiple transports, and more.

**Installation:**
```bash
pnpm add winston
```

**Implementation:**

```typescript
import { BaseTransport, LogEntry, LogLevel } from '@thunder/logger';
import winston from 'winston';

export class WinstonTransport extends BaseTransport {
  private logger: winston.Logger;

  constructor(options: {
    level?: LogLevel;
    filename?: string;
    maxFiles?: number;
    maxsize?: number;
  }) {
    super('winston', options.level || 'info');

    // Create Winston logger with file and console transports
    this.logger = winston.createLogger({
      level: this.getLevel(),
      format: winston.format.combine(
        winston.format.timestamp(),
        winston.format.errors({ stack: true }),
        winston.format.json()
      ),
      transports: [
        // Write all logs to combined.log
        new winston.transports.File({
          filename: options.filename || 'combined.log',
          maxFiles: options.maxFiles || 5,
          maxsize: options.maxsize || 5242880, // 5MB
        }),
        // Write errors to error.log
        new winston.transports.File({
          filename: 'error.log',
          level: 'error',
          maxFiles: options.maxFiles || 5,
          maxsize: options.maxsize || 5242880,
        }),
      ],
    });

    // Add console transport in development
    if (process.env.NODE_ENV !== 'production') {
      this.logger.add(
        new winston.transports.Console({
          format: winston.format.combine(
            winston.format.colorize(),
            winston.format.simple()
          ),
        })
      );
    }
  }

  async write(entry: LogEntry): Promise<void> {
    if (!this.shouldLog(entry.level)) {
      return;
    }

    const meta = {
      ...entry.context,
      component: entry.component,
      timestamp: entry.timestamp,
    };

    this.logger.log(entry.level, entry.message, meta);
  }

  async flush(): Promise<void> {
    return new Promise((resolve) => {
      this.logger.on('finish', resolve);
      this.logger.end();
    });
  }

  async close(): Promise<void> {
    await this.flush();
  }
}
```

**Usage:**

```typescript
import { Logger } from '@thunder/logger';
import { WinstonTransport } from './transports/winston';

// Next.js or Node.js application
const logger = new Logger({
  level: 'debug',
  transports: [
    new WinstonTransport({
      level: 'info',
      filename: 'logs/app.log',
      maxFiles: 10,
      maxsize: 10485760, // 10MB
    }),
  ],
});

logger.info('Application started', { version: '1.0.0' });
logger.error('Database connection failed', { error: err });
```

**Features:**
- **File rotation** - Automatic log file rotation based on size/date
- **Multiple files** - Separate files for different log levels
- **Structured logging** - JSON format for easy parsing
- **Console output** - Colorized console in development
- **Production-ready** - Mature library with extensive ecosystem

**Use Cases:**
- Next.js API routes and server components
- Node.js backend services
- Long-running processes requiring audit logs
- Applications needing log retention policies

---

### Sentry Integration

Integrate with Sentry for real-time error tracking and monitoring.

**Installation:**
```bash
# For browser
pnpm add @sentry/browser

# For Next.js
pnpm add @sentry/nextjs

# For Node.js
pnpm add @sentry/node
```

**Implementation:**

```typescript
import { Transport, LogEntry, LogLevel } from '@thunder/logger';
import * as Sentry from '@sentry/browser';
// Or for Next.js: import * as Sentry from '@sentry/nextjs';
// Or for Node.js: import * as Sentry from '@sentry/node';

export class SentryTransport implements Transport {
  name = 'sentry';
  private level: LogLevel;
  private initialized: boolean = false;

  constructor(options: {
    level?: LogLevel;
    dsn: string;
    environment?: string;
    release?: string;
    tracesSampleRate?: number;
  }) {
    this.level = options.level || 'error';

    // Initialize Sentry
    Sentry.init({
      dsn: options.dsn,
      environment: options.environment || process.env.NODE_ENV || 'development',
      release: options.release,
      tracesSampleRate: options.tracesSampleRate || 1.0,
      
      // Don't send errors in development (optional)
      beforeSend(event, hint) {
        if (process.env.NODE_ENV === 'development') {
          console.log('Sentry event:', event, hint);
          return null; // Don't send to Sentry in dev
        }
        return event;
      },
    });

    this.initialized = true;
  }

  async write(entry: LogEntry): Promise<void> {
    if (!this.initialized || !this.shouldLog(entry.level)) {
      return;
    }

    // Set user context if available
    if (entry.context?.userId) {
      Sentry.setUser({ id: entry.context.userId });
    }

    // Set additional context
    const { error, userId, ...extra } = entry.context || {};
    
    Sentry.setContext('log_entry', {
      component: entry.component,
      timestamp: entry.timestamp.toISOString(),
      ...extra,
    });

    // Set tags for filtering
    if (entry.component) {
      Sentry.setTag('component', entry.component);
    }

    // Handle errors differently
    if (entry.level === 'error' && error instanceof Error) {
      Sentry.captureException(error, {
        level: 'error',
        extra: extra,
        tags: {
          component: entry.component,
        },
      });
    } else {
      // Send as message for non-errors
      Sentry.captureMessage(entry.message, {
        level: this.mapLogLevel(entry.level),
        extra: extra,
        tags: {
          component: entry.component,
        },
      });
    }
  }

  private mapLogLevel(level: LogLevel): Sentry.SeverityLevel {
    const mapping: Record<LogLevel, Sentry.SeverityLevel> = {
      debug: 'debug',
      info: 'info',
      warn: 'warning',
      error: 'error',
    };
    return mapping[level];
  }

  async flush(): Promise<void> {
    if (this.initialized) {
      await Sentry.flush(2000);
    }
  }

  async close(): Promise<void> {
    await this.flush();
    await Sentry.close(2000);
  }
}
```

**Usage:**

```typescript
import { Logger, ConsoleTransport } from '@thunder/logger';
import { SentryTransport } from './transports/sentry';

// Browser or Next.js application
const logger = new Logger({
  level: 'info',
  transports: [
    // Console for development
    new ConsoleTransport({ level: 'debug' }),
    
    // Sentry for errors only
    new SentryTransport({
      level: 'error',
      dsn: 'https://examplePublicKey@o0.ingest.sentry.io/0',
      environment: process.env.NODE_ENV,
      release: process.env.NEXT_PUBLIC_APP_VERSION,
      tracesSampleRate: 0.1, // 10% of transactions
    }),
  ],
});

// Regular logging
logger.info('User logged in', { userId: 'user-123' });

// Errors automatically go to Sentry
try {
  await fetchUserData();
} catch (error) {
  logger.error('Failed to fetch user data', { 
    error, 
    userId: 'user-123',
    operation: 'fetchUserData' 
  });
}
```

**Advanced Usage with React Error Boundary:**

```tsx
import { Component, ErrorInfo, ReactNode } from 'react';
import { Logger } from '@thunder/logger';

interface Props {
  children: ReactNode;
  logger: Logger;
}

interface State {
  hasError: boolean;
}

export class ErrorBoundary extends Component<Props, State> {
  state = { hasError: false };

  static getDerivedStateFromError(): State {
    return { hasError: true };
  }

  componentDidCatch(error: Error, errorInfo: ErrorInfo) {
    // Log to Sentry via logger
    this.props.logger.error('React error boundary caught error', {
      error,
      componentStack: errorInfo.componentStack,
    });
  }

  render() {
    if (this.state.hasError) {
      return <h1>Something went wrong.</h1>;
    }

    return this.props.children;
  }
}

// Usage
import { useLogger } from '@thunder/logger/react';

function App() {
  const logger = useLogger();
  
  return (
    <ErrorBoundary logger={logger}>
      <YourApp />
    </ErrorBoundary>
  );
}
```

**Features:**
- **Error tracking** - Automatic error capturing with stack traces
- **Breadcrumbs** - Track user actions leading to errors
- **Source maps** - View original source code in error reports
- **User context** - Associate errors with specific users
- **Performance monitoring** - Track application performance
- **Release tracking** - Monitor errors across deployments

**Use Cases:**
- Production error monitoring
- User-reported bugs investigation
- Performance bottleneck identification
- Release health monitoring
- Critical error alerting

---

### Combining Multiple Services

Use multiple external services together for comprehensive logging:

```typescript
import { Logger, ConsoleTransport, StdoutTransport } from '@thunder/logger';
import { WinstonTransport } from './transports/winston';
import { SentryTransport } from './transports/sentry';

const isProduction = process.env.NODE_ENV === 'production';
const isDevelopment = process.env.NODE_ENV === 'development';

const logger = new Logger({
  level: isDevelopment ? 'debug' : 'info',
  transports: [
    // Console: Development only
    ...(isDevelopment ? [new ConsoleTransport({ level: 'debug' })] : []),
    
    // Winston: All logs to files (Node.js/Next.js server)
    ...(typeof process !== 'undefined' ? [
      new WinstonTransport({
        level: 'info',
        filename: 'logs/app.log',
        maxFiles: 30,
        maxsize: 10485760, // 10MB
      })
    ] : []),
    
    // Stdout: Docker/Cloud environments
    ...(isProduction ? [new StdoutTransport({ level: 'info' })] : []),
    
    // Sentry: Errors only in production
    ...(isProduction ? [
      new SentryTransport({
        level: 'error',
        dsn: process.env.SENTRY_DSN!,
        environment: 'production',
        tracesSampleRate: 0.1,
      })
    ] : []),
  ],
});

export default logger;
```

**Result:**
- **Development**: Debug logs to console
- **Production Server**: Info+ logs to Winston files, errors to Sentry
- **Production Docker**: Info+ logs to stdout (captured by log aggregator), errors to Sentry
- **Production Browser**: Errors to Sentry only

---

### Other Popular Integrations

#### Datadog

```typescript
import { BaseTransport, LogEntry, LogLevel } from '@thunder/logger';
import { datadogLogs } from '@datadog/browser-logs';

export class DatadogTransport extends BaseTransport {
  constructor(options: { level?: LogLevel; clientToken: string; site: string }) {
    super('datadog', options.level || 'info');
    datadogLogs.init({
      clientToken: options.clientToken,
      site: options.site,
      forwardErrorsToLogs: true,
      sessionSampleRate: 100,
    });
  }
  
  async write(entry: LogEntry): Promise<void> {
    if (!this.shouldLog(entry.level)) {
      return;
    }
    datadogLogs.logger[entry.level](entry.message, entry.context);
  }
}
```

#### LogRocket

```typescript
import { BaseTransport, LogEntry, LogLevel } from '@thunder/logger';
import LogRocket from 'logrocket';

export class LogRocketTransport extends BaseTransport {
  constructor(options: { level?: LogLevel; appId: string }) {
    super('logrocket', options.level || 'info');
    LogRocket.init(options.appId);
  }
  
  async write(entry: LogEntry): Promise<void> {
    if (!this.shouldLog(entry.level)) {
      return;
    }
    LogRocket.log(entry.level, entry.message, entry.context);
  }
}
```

#### Elastic (ELK Stack)

```typescript
import { BaseTransport, LogEntry, LogLevel } from '@thunder/logger';
import { Client } from '@elastic/elasticsearch';

export class ElasticTransport extends BaseTransport {
  private client: Client;
  private index: string;
  
  constructor(options: { level?: LogLevel; node: string; index: string }) {
    super('elastic', options.level || 'info');
    this.client = new Client({ node: options.node });
    this.index = options.index;
  }
  
  async write(entry: LogEntry): Promise<void> {
    if (!this.shouldLog(entry.level)) {
      return;
    }
    
    await this.client.index({
      index: this.index,
      document: {
        '@timestamp': entry.timestamp,
        level: entry.level,
        message: entry.message,
        component: entry.component,
        ...entry.context,
      },
    });
  }
}
```

## React Integration

### LoggerProvider

Wrap your application with `LoggerProvider` to provide logger access to all components:

```tsx
import { LoggerProvider } from '@thunder/logger/react';
import { Logger, ConsoleTransport, HttpTransport, LogLevel } from '@thunder/logger';

function App() {
  return (
    <LoggerProvider
      logger={{
        level: LogLevel.INFO,
        transports: [
          new ConsoleTransport(),
          new HttpTransport({
            level: LogLevel.ERROR,
            endpoint: import.meta.env.VITE_LOG_ENDPOINT,
          }),
        ],
      }}
    >
      <YourApp />
    </LoggerProvider>
  );
}
```

You can also pass a Logger instance directly:

```tsx
const logger = new Logger({
  level: import.meta.env.DEV ? LogLevel.DEBUG : LogLevel.WARN,
  transports: [new ConsoleTransport()],
});

function App() {
  return (
    <LoggerProvider logger={logger}>
      <YourApp />
    </LoggerProvider>
  );
}
```

### useLogger Hook

Access the logger in any component:

```tsx
import { useLogger } from '@thunder/logger/react';

function UserProfile({ userId }) {
  const logger = useLogger();

  useEffect(() => {
    logger.info('UserProfile mounted', { userId });
    
    return () => {
      logger.debug('UserProfile unmounted', { userId });
    };
  }, [userId]);

  const handleUpdate = async (data) => {
    try {
      logger.debug('Updating user', { userId, data });
      await updateUser(userId, data);
      logger.info('User updated successfully', { userId });
    } catch (error) {
      logger.error('Failed to update user', { userId, error });
    }
  };

  return <div>...</div>;
}
```

### Component Scoping

Automatically scope loggers to specific components by passing the component name to `useLogger`:

```tsx
import { useLogger } from '@thunder/logger/react';

function UserList() {
  const logger = useLogger('UserList');

  useEffect(() => {
    logger.info('Fetching users'); 
    // Logs: { level: 'info', message: 'Fetching users', component: 'UserList' }
  }, []);

  const handleDelete = (userId: string) => {
    logger.warn('User deleted', { userId });
    // Logs: { level: 'warn', message: 'User deleted', component: 'UserList', context: { userId } }
  };

  return <div>...</div>;
}
```

### Manual Component Scoping

You can also create component-scoped loggers manually:

```tsx
function DataTable() {
  const baseLogger = useLogger();
  const logger = useMemo(
    () => baseLogger.withComponent('DataTable'),
    [baseLogger]
  );

  useEffect(() => {
    logger.info('Table rendered'); // Logs: [DataTable] Table rendered
  }, []);

  return <div>...</div>;
}
```

## Environment Configuration

### Vite Applications

```typescript
// Detect development mode
const isDev = import.meta.env.DEV;
const logLevel = import.meta.env.VITE_LOG_LEVEL || (isDev ? 'debug' : 'info');

const logger = new Logger({
  level: logLevel,
  transports: [new ConsoleTransport()],
});
```

### Next.js Applications

```typescript
// next.config.js - Define public env vars
module.exports = {
  env: {
    NEXT_PUBLIC_LOG_LEVEL: process.env.LOG_LEVEL || 'info',
    NEXT_PUBLIC_LOG_ENDPOINT: process.env.LOG_ENDPOINT,
  },
};

// app.tsx - Use in client components
const logger = new Logger({
  level: process.env.NEXT_PUBLIC_LOG_LEVEL,
  transports: [
    new ConsoleTransport(),
    new HttpTransport({
      endpoint: process.env.NEXT_PUBLIC_LOG_ENDPOINT,
    }),
  ],
});
```

## Migration from console.*

Replace console calls with structured logging:

### Before

```typescript
console.log('User logged in');
console.warn('Deprecated API used');
console.error('Failed to save', error);
```

### After

```typescript
logger.info('User logged in', { userId: user.id });
logger.warn('Deprecated API used', { api: '/old-endpoint' });
logger.error('Failed to save', { error, userId: user.id });
```

### Benefits

- **Structured data** - Easy to query and filter in log management systems
- **Consistent formatting** - Same format across your application
- **Multiple outputs** - Send to console, files, and remote services
- **Production-safe** - Control what gets logged in different environments
- **Type-safe** - Catch logging errors at compile time

## Package Structure

The `@thunder/logger` package follows Thunder's standard coding practices with a clean, modular structure:

```
src/
├── react/             # React-specific code
│   └── contexts/      # React context providers and hooks
│       └── Logger/
│           ├── LoggerContext.tsx          # React Context definition
│           ├── LoggerProvider.tsx         # Provider component
│           └── useLogger.tsx              # Hook to access logger (with optional component scoping)
├── core/              # Core logger implementation
│   ├── factory.ts     # Transport factory system
│   ├── logger.ts      # Main Logger class
│   ├── transport.ts   # Base transport and registry
│   └── utils.ts       # Utility functions
├── models/            # Type definitions and interfaces
│   ├── log-level.ts           # LogLevel enum
│   ├── log-context.ts         # Context type
│   ├── log-entry.ts           # Log entry interface
│   ├── logger-config.ts       # Logger configuration
│   ├── transport.ts           # Transport interface
│   ├── transport-config.ts    # Transport configuration
│   └── runtime-environment.ts # Environment detection types
├── transports/        # Built-in transport implementations
│   ├── console.ts     # Browser console transport
│   ├── stdout.ts      # Node.js stdout transport
│   └── http.ts        # HTTP remote logging transport
├── index.ts           # Main package entry point
└── react.ts           # React-specific entry point
```

### Entry Points

- **`@thunder/logger`** - Main entry point with all exports
- **`@thunder/logger/react`** - React-specific exports (Provider, hooks, Context)

### Design Principles

1. **Default Exports** - Each module exports a single primary entity as default
2. **Single Responsibility** - Each file has one clear purpose
3. **Type Organization** - All types in dedicated `models/` folder
4. **React Isolation** - React-specific code in dedicated `react/` folder
5. **Context Pattern** - React integration follows Thunder's standard Context/Provider/Hook pattern

## Best Practices

### 1. Use Appropriate Log Levels

```typescript
// ✅ Good
logger.debug('Variable value', { value: x });           // For debugging
logger.info('User signed up', { userId: user.id });     // Important events
logger.warn('API deprecated', { api: '/v1/users' });    // Warnings
logger.error('Database error', { error: err });         // Errors

// ❌ Avoid
logger.info('x = 5');                                   // Use debug for variables
logger.error('User clicked button');                    // Not an error
```

### 2. Include Contextual Information

```typescript
// ✅ Good
logger.error('Payment failed', {
  orderId: 'ORD-123',
  userId: 'user-456',
  amount: 99.99,
  error: err,
});

// ❌ Avoid
logger.error('Payment failed');
```

### 3. Use Component-Scoped Loggers

```typescript
// ✅ Good
const logger = baseLogger.withComponent('PaymentForm');
logger.info('Form submitted'); // [PaymentForm] Form submitted

// ❌ Avoid
logger.info('[PaymentForm] Form submitted'); // Manual prefixing
```

### 4. Avoid Logging Sensitive Data

```typescript
// ✅ Good
logger.info('User logged in', { userId: user.id });

// ❌ Avoid
logger.info('User logged in', {
  userId: user.id,
  password: user.password,      // Never log passwords
  creditCard: user.creditCard,  // Never log payment info
});
```

### 5. Use Conditional Debug Logging

```typescript
// ✅ Good - Skip expensive operations
if (logger.isDebugEnabled()) {
  const debugData = JSON.stringify(largeObject, null, 2);
  logger.debug('Large object', { data: debugData });
}

// ❌ Avoid - Always processes even if debug is disabled
logger.debug('Large object', { data: JSON.stringify(largeObject, null, 2) });
```

### 6. Configure Transports by Environment

```typescript
// ✅ Good
const transports = [
  new ConsoleTransport({ level: isDev ? 'debug' : 'warn' }),
  ...(isProd ? [new HttpTransport({ level: 'error', endpoint: LOG_ENDPOINT })] : []),
];

const logger = new Logger({ level: 'debug', transports });
```

## Production Considerations

### 1. Log Level Management

```typescript
// Development: Debug everything
// Production: Only warnings and errors
const logger = new Logger({
  level: process.env.NODE_ENV === 'production' ? 'warn' : 'debug',
  transports: [new ConsoleTransport()],
});
```

### 2. Error Aggregation

```typescript
// Send errors to monitoring service
const logger = new Logger({
  level: 'info',
  transports: [
    new ConsoleTransport(),
    new HttpTransport({
      level: 'error', // Only errors go to remote service
      endpoint: 'https://logs.example.com/api/logs',
    }),
  ],
});
```

### 3. Performance Optimization

```typescript
// Disable console output in production
const transports = process.env.NODE_ENV === 'production'
  ? [new HttpTransport({ endpoint: LOG_ENDPOINT })]
  : [new ConsoleTransport()];

const logger = new Logger({ level: 'info', transports });
```

## TypeScript Support

Full TypeScript support with type inference:

```typescript
import type { Logger, LogLevel, LogEntry, Transport } from '@thunder/logger';

// Logger config is fully typed
const config: LoggerConfig = {
  level: 'info', // autocomplete: 'debug' | 'info' | 'warn' | 'error'
  transports: [new ConsoleTransport()],
};

// Custom transport with type safety
class CustomTransport implements Transport {
  name = 'custom';
  
  shouldLog(level: LogLevel): boolean {
    return true;
  }
  
  write(entry: LogEntry): void {
    // entry is fully typed
    console.log(entry.message, entry.context);
  }
}
```

## Examples

See the [samples/apps](../../samples/apps/) directory for complete working examples:

- **Vanilla JS**: Basic logger usage without frameworks
- **React + Vite**: React application with LoggerProvider
- **Next.js**: SSR-safe logging with server and client examples

## API Reference

### Logger

| Method | Parameters | Returns | Description |
|--------|------------|---------|-------------|
| `constructor` | `config: LoggerConfig` | `Logger` | Create a new logger instance |
| `debug` | `message: string, context?: object` | `void` | Log debug message |
| `info` | `message: string, context?: object` | `void` | Log info message |
| `warn` | `message: string, context?: object` | `void` | Log warning message |
| `error` | `message: string, context?: object` | `void` | Log error message |
| `withContext` | `context: object` | `Logger` | Create logger with persistent context |
| `withComponent` | `component: string` | `Logger` | Create logger with component name |
| `isDebugEnabled` | - | `boolean` | Check if debug logging is enabled |
| `flush` | - | `Promise<void>` | Flush all transports |
| `close` | - | `Promise<void>` | Close all transports |

### LoggerConfig

```typescript
interface LoggerConfig {
  level: LogLevel;
  transports: Transport[];
}
```

### Transport Interface

```typescript
interface Transport {
  getName(): string;
  getLevel(): LogLevel;
  shouldLog(level: LogLevel): boolean;
  write(entry: LogEntry): Promise<void>;
  flush(): Promise<void>;
  close(): Promise<void>;
}
```

**Note:** It's recommended to extend `BaseTransport` which provides default implementations of all methods except `write()`.

### LogEntry Interface

```typescript
interface LogEntry {
  level: LogLevel;
  message: string;
  timestamp: Date;
  context?: Record<string, unknown>;
  component?: string;
  error?: Error;
}
```

### LogLevel Type

```typescript
type LogLevel = 'debug' | 'info' | 'warn' | 'error';
```

## Contributing

See [CONTRIBUTING.md](../../../CONTRIBUTING.md) for development setup and contribution guidelines.

## License

See [LICENSE](../../../LICENSE) for license information.
