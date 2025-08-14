# ADR-003: Dependency Injection and Lifecycle Patterns

**Status:** Accepted  
**Date:** 2025-08-14  

## Context

The application needed better organization of dependencies and clear separation between application-level and request-level concerns for maintainability and testability.

## Decision

Implement dependency injection with two distinct lifecycle patterns:

### Application Singletons
Components that live for the entire application lifecycle, initialized once using `sync.Once`:
- **Database Repository**: Shared connection and state
- **Event Broker**: Message bus for real-time updates  
- **Components Container**: Main application container

### Request-Scoped Components
Components created fresh for each HTTP request:
- **Request Context**: Unique request ID, user IP, timing
- **Request Logger**: Auto-injects request metadata to logs
- **Custom Values**: User context, feature flags, tracing data

## Implementation

```go
// Application Singletons (thread-safe, shared)
database.Init()           // sync.Once initialization
events.GetBroker()        // Lazy singleton
components.GetInstance()  // Container singleton

// Request Scoped (unique per request)
context.NewRequestContext(ginCtx)     // Fresh context
reqCtx.Logger.Info("msg", fields)     // Auto request_id injection
reqCtx.WithValue("user_id", "123")    // Request storage
```

## Benefits

- **Clear separation**: Application vs request lifecycles
- **Thread safety**: `sync.Once` prevents race conditions
- **Better testing**: Isolated contexts, controlled singletons
- **Request tracing**: Automatic logging with request metadata
- **Memory efficiency**: Shared resources where appropriate

## Structure

```
src/
├── components/     # Application container and singletons
└── diplomat/
    └── middleware/ # Request context types and injection
```

## Testing

- **Singleton tests**: Verify `sync.Once` behavior and concurrency safety
- **Request tests**: Verify isolation and shared singleton access
- **Integration**: Real HTTP request lifecycle testing