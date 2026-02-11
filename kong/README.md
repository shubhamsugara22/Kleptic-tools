# Kong Gateway Demo - Go Edition

This Go program demonstrates the basic functionality of Kong Gateway including:

## Features Demonstrated

1. **Health Check** - Verifies Kong Gateway is running
2. **Service Creation** - Creates a service pointing to httpbin.org
3. **Route Creation** - Sets up routes to access the service
4. **Proxy Testing** - Makes requests through Kong's proxy
5. **Plugin Configuration** - Adds rate limiting plugin
6. **Service Listing** - Retrieves all configured services
7. **Rate Limiting Test** - Demonstrates the rate limiting in action

## Prerequisites

1. **Kong Gateway Running** - Ensure Kong is running (use docker-compose or the setup.sh script)
   ```bash
   # Admin API should be accessible at: http://localhost:8001
   # Proxy should be accessible at: http://localhost:8000
   ```

2. **Go Installed** - Go 1.21 or higher
   ```bash
   go version
   ```

## Quick Start

### 1. Start Kong Gateway

Using docker-compose (recommended):
```bash
cd kong
docker-compose up -d
```

Or using the setup.sh script:
```bash
cd kong
chmod +x setup.sh
./setup.sh
```

### 2. Run the Demo

```bash
cd kong
go run main.go
```

## Expected Output

```
🚀 Kong Gateway Demo - Basic Functionality
==========================================

📊 Step 1: Checking Kong Gateway status...
   Database: postgresql
   Server: kong/3.11.0
✅ Kong Gateway is running!

🔧 Step 2: Creating a service...
✅ Service created with ID: xxx-xxx-xxx

🛣️  Step 3: Creating a route for the service...
✅ Route created with ID: xxx-xxx-xxx

🌐 Step 4: Testing the route through Kong Gateway...
   Response status: 200
   Response preview: {...}
✅ Successfully made request through Kong Gateway!

🔌 Step 5: Adding rate limiting plugin...
✅ Rate limiting plugin added with ID: xxx-xxx-xxx

📋 Step 6: Listing all services...
   Total services: 1
   1. Name: example-service, Host: httpbin.org

🧪 Step 7: Testing rate limiting (making 6 requests)...
   Request 1: Status 200 | Remaining: 4 ✅
   Request 2: Status 200 | Remaining: 3 ✅
   Request 3: Status 200 | Remaining: 2 ✅
   Request 4: Status 200 | Remaining: 1 ✅
   Request 5: Status 200 | Remaining: 0 ✅
   Request 6: Status 429 ⛔ Rate limit exceeded!

🎉 Demo completed successfully!
```

## What This Demo Does

### 1. Service Management
- Creates a service named "example-service" pointing to httpbin.org
- This acts as the upstream API that Kong will proxy

### 2. Routing
- Creates a route with path `/httpbin`
- Maps incoming requests to the service
- Accepts GET and POST methods

### 3. Plugin Management
- Adds a rate limiting plugin (5 requests per minute)
- Demonstrates Kong's plugin architecture
- Shows how plugins modify request behavior

### 4. Proxying
- Routes requests through Kong at `http://localhost:8000/httpbin`
- Kong forwards to `http://httpbin.org`
- Returns responses to the client

## API Endpoints Used

### Kong Admin API (Port 8001)
- `GET /status` - Check Kong status
- `POST /services` - Create a service
- `POST /services/{service}/routes` - Create a route
- `POST /services/{service}/plugins` - Add a plugin
- `GET /services` - List all services

### Kong Proxy (Port 8000)
- `GET /httpbin/get` - Test proxied request

## Extending the Demo

### Add More Plugins

```go
// Key Authentication
plugin := Plugin{
    Name: "key-auth",
    Config: map[string]interface{}{
        "key_names": []string{"apikey"},
    },
    Enabled: true,
}

// CORS
plugin := Plugin{
    Name: "cors",
    Config: map[string]interface{}{
        "origins": []string{"*"},
        "methods": []string{"GET", "POST"},
    },
    Enabled: true,
}

// Request Transformer
plugin := Plugin{
    Name: "request-transformer",
    Config: map[string]interface{}{
        "add": map[string]interface{}{
            "headers": []string{"X-Custom-Header:value"},
        },
    },
    Enabled: true,
}
```

### Create Multiple Services

```go
services := []Service{
    {Name: "api-v1", URL: "http://api.example.com/v1"},
    {Name: "api-v2", URL: "http://api.example.com/v2"},
    {Name: "auth-service", URL: "http://auth.example.com"},
}

for _, svc := range services {
    createService(svc)
}
```

## Troubleshooting

### Kong Not Running
```bash
# Check if Kong containers are running
docker ps | grep kong

# Check Kong logs
docker logs kong-ee-gateway
```

### Port Already in Use
```bash
# Check what's using port 8001/8000
netstat -ano | findstr :8001
netstat -ano | findstr :8000

# Stop Kong
docker-compose down
```

### Service Creation Fails
- Ensure Kong database is initialized
- Check Kong Admin API is accessible: `curl http://localhost:8001`

## Resources

- [Kong Gateway Documentation](https://docs.konghq.com/gateway/latest/)
- [Kong Admin API Reference](https://docs.konghq.com/gateway/latest/admin-api/)
- [Kong Plugins Hub](https://docs.konghq.com/hub/)
- [HTTPBin API](http://httpbin.org/)

## Next Steps

1. Explore more Kong plugins (authentication, logging, security)
2. Set up Kong in DB-less mode
3. Configure Kong for production
4. Implement custom plugins
5. Set up Kong Manager for GUI management
6. Integrate with CI/CD pipeline
