# NetTestLab Web Interface with gRPC-Web

This setup provides a complete web interface for NetTestLab that communicates directly with the gRPC backend using gRPC-Web.

## Architecture

```
Browser ←→ gRPC-Web ←→ NetTestLab gRPC Server
```

## Components

### 1. gRPC-Web Client Generation
- **Location**: `clients/javascript/`
- **Generated Files**: `src/nettestlab/v1/*_pb.js` and `*_grpc_web_pb.js`
- **Client Wrapper**: `src/client.js` - TypeScript/JavaScript wrapper for easier usage

### 2. Web Interface
- **Location**: `web/`
- **Main Files**:
  - `index.html` - Complete single-page application
  - `static/css/style.css` - Custom styling
  - `static/js/api.js` - gRPC-Web API client with fallback to mock data
  - `static/js/app.js` - Main application logic
  - `static/js/grpc-client.js` - gRPC-Web client wrapper
  - `static/js/nettestlab/` - Generated gRPC-Web files

### 3. Server
- **gRPC Server**: Port 8080 (for gRPC-Web communication)
- **HTTP Server**: Port 8081 (serves web interface)

## Features

### Dashboard
- System status monitoring
- Real-time metrics display
- Health checks for all services

### Network Interfaces
- List all network interfaces
- View interface details (IP, MAC, speed, statistics)
- Enable/disable interfaces
- Apply traffic control profiles

### Connected Clients
- View all connected devices
- Client details (name, MAC, IP, profile)
- Update client configurations
- Delete clients

### Profile Management
- CRUD operations for network profiles
- Speed limits (download/upload)
- Latency and packet loss simulation
- Apply profiles to interfaces or clients

## How It Works

1. **Client Initialization**: The web page loads the gRPC-Web generated files and initializes the NetTestLab client
2. **API Calls**: JavaScript makes async calls to the gRPC server via gRPC-Web
3. **Fallback Mode**: If gRPC is unavailable, the interface shows mock data
4. **Real-time Updates**: The interface periodically refreshes data from the server

## Usage

### Starting the Server
```bash
cd /Users/hector/NetTestLab
go run cmd/server/main.go
```

### Accessing the Web Interface
Open your browser to: http://localhost:8081

### Regenerating gRPC-Web Files
When protobuf definitions change:
```bash
cd /Users/hector/NetTestLab
buf generate
```

### Building JavaScript Client
```bash
cd clients/javascript
npm run build
```

## Testing

### Testing gRPC-Web Client
Use the test page: `clients/javascript/test.html`
- Direct testing of gRPC-Web client
- Error handling verification
- Connection status monitoring

### Testing Web Interface
The main interface automatically:
- Detects gRPC-Web availability
- Falls back to mock data if server is unavailable
- Shows connection status in console

## Development Notes

### Mock Data vs Real Data
The API client (`web/static/js/api.js`) includes comprehensive mock data that mirrors the expected gRPC responses. This allows the interface to work even when:
- The gRPC server is not running
- gRPC-Web files are not loaded
- Server endpoints are not implemented

### Error Handling
The client includes robust error handling:
- Connection failures show user-friendly messages
- gRPC errors are translated to readable text
- Fallback to mock data ensures interface remains functional

### Browser Compatibility
The interface uses:
- Modern JavaScript (ES2020)
- Bootstrap 5 for responsive design
- gRPC-Web for browser-compatible gRPC communication

## File Structure

```
/web/
├── index.html                    # Main web interface
├── static/
│   ├── css/
│   │   └── style.css            # Custom styling
│   └── js/
│       ├── api.js               # gRPC-Web API client
│       ├── app.js               # Main application logic
│       ├── grpc-client.js       # gRPC client wrapper
│       └── nettestlab/          # Generated gRPC-Web files
│           └── v1/
│               ├── monitoring_pb.js
│               ├── monitoring_grpc_web_pb.js
│               ├── network_pb.js
│               ├── network_grpc_web_pb.js
│               ├── profiles_pb.js
│               └── profiles_grpc_web_pb.js

/clients/javascript/
├── package.json                 # NPM package configuration
├── tsconfig.json               # TypeScript configuration
├── test.html                   # gRPC client test page
└── src/
    ├── index.ts                # TypeScript client (WIP)
    ├── client.js               # JavaScript client wrapper
    └── nettestlab/             # Generated gRPC-Web files
        └── v1/
            ├── monitoring_pb.js
            ├── monitoring_grpc_web_pb.js
            ├── network_pb.js
            ├── network_grpc_web_pb.js
            ├── profiles_pb.js
            └── profiles_grpc_web_pb.js
```

## Next Steps

1. **Implement Server Methods**: Complete the gRPC server implementation for all service methods
2. **Real-time Updates**: Add WebSocket or streaming gRPC for live data updates
3. **Authentication**: Add user authentication and authorization
4. **Advanced Features**: Implement more sophisticated traffic shaping and monitoring
5. **Testing**: Add comprehensive unit and integration tests
6. **Documentation**: Complete API documentation and user guides