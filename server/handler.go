package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sego/responseutils"
	"strconv"
	"strings"
	"time"
)

// Handlers contains all HTTP handlers
type Handlers struct {
	server *Server
}

// NewHandlers creates new handlers instance
func NewHandlers(server *Server) *Handlers {
	return &Handlers{server: server}
}

// GetUsers returns list of users
func (h *Handlers) GetUsers(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}

	perPage, _ := strconv.Atoi(r.URL.Query().Get("per_page"))
	if perPage < 1 {
		perPage = 10
	}
	if perPage > 100 {
		perPage = 100
	}

	// Simulate database query
	users := []map[string]interface{}{
		{
			"id":         1,
			"name":       "John Doe",
			"email":      "john@example.com",
			"created_at": time.Now().Add(-24 * time.Hour),
		},
		{
			"id":         2,
			"name":       "Jane Smith",
			"email":      "jane@example.com",
			"created_at": time.Now().Add(-12 * time.Hour),
		},
	}

	total := 100 // Simulated total count

	// Return paginated response
	responseutils.Paginated(w, users, page, perPage, total, "Users retrieved successfully")
}

// GetUser returns a single user
func (h *Handlers) GetUser(w http.ResponseWriter, r *http.Request) {
	// Extract ID from URL path
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 4 {
		responseutils.NotFound(w, "User not found")
		return
	}

	idStr := pathParts[3]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		responseutils.BadRequest(w, "Invalid user ID")
		return
	}

	// Simulate database lookup
	if id <= 0 || id > 1000 {
		responseutils.NotFound(w, fmt.Sprintf("User with ID %d not found", id))
		return
	}

	user := map[string]interface{}{
		"id":         id,
		"name":       "John Doe",
		"email":      fmt.Sprintf("user%d@example.com", id),
		"created_at": time.Now().Add(-time.Duration(id) * time.Hour),
		"updated_at": time.Now(),
	}

	responseutils.Success(w, user, "User retrieved successfully")
}

// CreateUser creates a new user
func (h *Handlers) CreateUser(w http.ResponseWriter, r *http.Request) {
	// Parse request body
	var data map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		responseutils.BadRequest(w, "Invalid request body")
		return
	}

	// Validate required fields
	if name, ok := data["name"].(string); !ok || name == "" {
		responseutils.ValidationError(w, map[string]string{
			"name": "Name is required",
		})
		return
	}

	if email, ok := data["email"].(string); !ok || email == "" {
		responseutils.ValidationError(w, map[string]string{
			"email": "Email is required",
		})
		return
	}

	// Simulate user creation
	newUser := map[string]interface{}{
		"id":         101, // Simulated new ID
		"name":       data["name"],
		"email":      data["email"],
		"created_at": time.Now(),
		"updated_at": time.Now(),
	}

	// Return created response with location header
	location := fmt.Sprintf("/api/v1/users/%d", newUser["id"])
	responseutils.CreatedWithLocation(w, newUser, location, "User created successfully")
}

// UpdateUser updates an existing user
func (h *Handlers) UpdateUser(w http.ResponseWriter, r *http.Request) {
	// Extract ID
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 4 {
		responseutils.NotFound(w, "User not found")
		return
	}

	idStr := pathParts[3]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		responseutils.BadRequest(w, "Invalid user ID")
		return
	}

	// Parse request body
	var data map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		responseutils.BadRequest(w, "Invalid request body")
		return
	}

	// Simulate update
	updatedUser := map[string]interface{}{
		"id":         id,
		"name":       data["name"],
		"email":      data["email"],
		"updated_at": time.Now(),
	}

	responseutils.Success(w, updatedUser, "User updated successfully")
}

// DeleteUser deletes a user
func (h *Handlers) DeleteUser(w http.ResponseWriter, r *http.Request) {
	// Extract ID
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 4 {
		responseutils.NotFound(w, "User not found")
		return
	}

	idStr := pathParts[3]
	_, err := strconv.Atoi(idStr)
	if err != nil {
		responseutils.BadRequest(w, "Invalid user ID")
		return
	}

	// Simulate deletion
	responseutils.NoContent(w)
}

// ServeFrontend serves frontend HTML
func (h *Handlers) ServeFrontend(w http.ResponseWriter, r *http.Request) {
	html := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Go Helpers API Server</title>
    <style>
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }
        
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, Cantarell, sans-serif;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            min-height: 100vh;
            display: flex;
            justify-content: center;
            align-items: center;
            color: white;
        }
        
        .container {
            max-width: 800px;
            width: 90%;
            background: rgba(255, 255, 255, 0.1);
            backdrop-filter: blur(10px);
            border-radius: 20px;
            padding: 40px;
            box-shadow: 0 20px 60px rgba(0, 0, 0, 0.3);
        }
        
        h1 {
            font-size: 3rem;
            margin-bottom: 20px;
            background: linear-gradient(45deg, #fff, #f0f0f0);
            -webkit-background-clip: text;
            -webkit-text-fill-color: transparent;
            text-align: center;
        }
        
        .subtitle {
            font-size: 1.2rem;
            opacity: 0.9;
            text-align: center;
            margin-bottom: 40px;
        }
        
        .endpoints {
            background: rgba(255, 255, 255, 0.05);
            border-radius: 15px;
            padding: 25px;
            margin-bottom: 30px;
        }
        
        .endpoint {
            display: flex;
            align-items: center;
            margin-bottom: 15px;
            padding: 12px 20px;
            background: rgba(255, 255, 255, 0.1);
            border-radius: 10px;
            transition: all 0.3s ease;
        }
        
        .endpoint:hover {
            background: rgba(255, 255, 255, 0.15);
            transform: translateX(5px);
        }
        
        .method {
            display: inline-block;
            padding: 4px 12px;
            border-radius: 6px;
            font-weight: bold;
            font-size: 0.9rem;
            margin-right: 15px;
            min-width: 70px;
            text-align: center;
        }
        
        .get { background: #10b981; }
        .post { background: #3b82f6; }
        .put { background: #f59e0b; }
        .delete { background: #ef4444; }
        
        .path {
            font-family: 'Courier New', monospace;
            font-size: 1.1rem;
        }
        
        .info {
            display: flex;
            justify-content: space-between;
            margin-top: 30px;
            padding-top: 20px;
            border-top: 1px solid rgba(255, 255, 255, 0.1);
            font-size: 0.9rem;
            opacity: 0.8;
        }
        
        .status {
            display: inline-block;
            padding: 2px 8px;
            border-radius: 4px;
            background: rgba(16, 185, 129, 0.3);
            color: #10b981;
            font-weight: bold;
        }
        
        @media (max-width: 768px) {
            .container {
                padding: 25px;
            }
            
            h1 {
                font-size: 2.2rem;
            }
            
            .endpoint {
                flex-direction: column;
                align-items: flex-start;
            }
            
            .method {
                margin-bottom: 8px;
                margin-right: 0;
            }
        }
    </style>
</head>
<body>
    <div class="container">
        <h1>🚀 Go Helpers API Server</h1>
        <p class="subtitle">A comprehensive Go utilities server with production-ready features</p>
        
        <div class="endpoints">
            <div class="endpoint">
                <span class="method get">GET</span>
                <span class="path">/api/v1/health</span>
            </div>
            <div class="endpoint">
                <span class="method get">GET</span>
                <span class="path">/api/v1/metrics</span>
            </div>
            <div class="endpoint">
                <span class="method get">GET</span>
                <span class="path">/api/v1/users</span>
            </div>
            <div class="endpoint">
                <span class="method get">GET</span>
                <span class="path">/api/v1/users/{id}</span>
            </div>
            <div class="endpoint">
                <span class="method post">POST</span>
                <span class="path">/api/v1/users</span>
            </div>
            <div class="endpoint">
                <span class="method put">PUT</span>
                <span class="path">/api/v1/users/{id}</span>
            </div>
            <div class="endpoint">
                <span class="method delete">DELETE</span>
                <span class="path">/api/v1/users/{id}</span>
            </div>
        </div>
        
        <div class="info">
            <div>
                <span class="status">● Online</span>
                <span> | Environment: ` + h.server.config.Environment + `</span>
            </div>
            <div>
                Uptime: ` + h.server.health.GetUptime() + `
            </div>
        </div>
    </div>
    
    <script>
        // Update uptime every second
        function updateUptime() {
            const uptimeElement = document.querySelector('.info div:nth-child(2)');
            if (uptimeElement) {
                fetch('/api/v1/health')
                    .then(response => response.json())
                    .then(data => {
                        if (data.uptime) {
                            uptimeElement.textContent = 'Uptime: ' + data.uptime;
                        }
                    })
                    .catch(console.error);
            }
        }
        
        // Update every 5 seconds
        setInterval(updateUptime, 5000);
        updateUptime();
    </script>
</body>
</html>`

	responseutils.WriteHTML(w, http.StatusOK, html)
}

// NotFound handles 404 errors
func (h *Handlers) NotFound(w http.ResponseWriter, r *http.Request) {
	html := `<!DOCTYPE html>
<html>
<head>
    <title>404 - Page Not Found</title>
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            background: linear-gradient(135deg, #f093fb 0%, #f5576c 100%);
            height: 100vh;
            display: flex;
            justify-content: center;
            align-items: center;
            color: white;
            text-align: center;
        }
        
        .container {
            max-width: 600px;
            padding: 40px;
            background: rgba(255, 255, 255, 0.1);
            backdrop-filter: blur(10px);
            border-radius: 20px;
            box-shadow: 0 20px 60px rgba(0, 0, 0, 0.3);
        }
        
        h1 {
            font-size: 8rem;
            margin: 0;
            line-height: 1;
            background: linear-gradient(45deg, #fff, #f0f0f0);
            -webkit-background-clip: text;
            -webkit-text-fill-color: transparent;
        }
        
        h2 {
            font-size: 2rem;
            margin: 20px 0;
        }
        
        p {
            font-size: 1.2rem;
            opacity: 0.9;
            margin-bottom: 30px;
        }
        
        a {
            display: inline-block;
            padding: 12px 30px;
            background: white;
            color: #f5576c;
            text-decoration: none;
            border-radius: 50px;
            font-weight: bold;
            transition: transform 0.3s ease;
        }
        
        a:hover {
            transform: translateY(-2px);
            box-shadow: 0 10px 20px rgba(0, 0, 0, 0.2);
        }
        
        .path {
            font-family: 'Courier New', monospace;
            background: rgba(0, 0, 0, 0.2);
            padding: 5px 10px;
            border-radius: 5px;
            margin: 10px 0;
        }
    </style>
</head>
<body>
    <div class="container">
        <h1>404</h1>
        <h2>Page Not Found</h2>
        <p>The page you're looking for doesn't exist.</p>
        <div class="path">` + r.URL.Path + `</div>
        <a href="/">Go Home</a>
    </div>
</body>
</html>`

	responseutils.WriteHTML(w, http.StatusNotFound, html)
}

// HealthHandler handles health check
func (h *Handlers) HealthHandler(w http.ResponseWriter, r *http.Request) {
	responseutils.WriteHealthCheck(w, "healthy")
}

// MetricsHandler handles metrics endpoint
func (h *Handlers) MetricsHandler(w http.ResponseWriter, r *http.Request) {
	metrics := map[string]interface{}{
		"requests": map[string]interface{}{
			"total":           1234,
			"success":         1200,
			"failed":          34,
			"rate_per_second": 12.5,
		},
		"memory": map[string]interface{}{
			"allocated": "128 MB",
			"used":      "64 MB",
			"free":      "64 MB",
		},
		"uptime": h.server.health.GetUptime(),
	}

	responseutils.WriteMetrics(w, metrics)
}

// EchoHandler echoes back request data (for testing)
func (h *Handlers) EchoHandler(w http.ResponseWriter, r *http.Request) {
	echoData := map[string]interface{}{
		"method":      r.Method,
		"path":        r.URL.Path,
		"query":       r.URL.Query(),
		"headers":     r.Header,
		"remote_addr": r.RemoteAddr,
		"timestamp":   time.Now().Format(time.RFC3339),
	}

	responseutils.Success(w, echoData, "Request echoed successfully")
}

// UploadHandler handles file uploads
func (h *Handlers) UploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		responseutils.MethodNotAllowed(w)
		return
	}

	// Parse multipart form (max 10MB)
	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		responseutils.BadRequest(w, "Failed to parse form: "+err.Error())
		return
	}

	// Get file from form
	file, handler, err := r.FormFile("file")
	if err != nil {
		responseutils.BadRequest(w, "No file uploaded")
		return
	}
	defer file.Close()

	response := map[string]interface{}{
		"filename":     handler.Filename,
		"size":         handler.Size,
		"content_type": handler.Header.Get("Content-Type"),
		"uploaded_at":  time.Now().Format(time.RFC3339),
	}

	responseutils.Success(w, response, "File uploaded successfully")
}

// DownloadHandler serves file downloads
func (h *Handlers) DownloadHandler(w http.ResponseWriter, r *http.Request) {
	filename := r.URL.Query().Get("filename")
	if filename == "" {
		filename = "example.txt"
	}

	content := []byte("This is an example file download.\n" +
		"Downloaded from Go Helpers API Server.\n" +
		"Timestamp: " + time.Now().Format(time.RFC3339))

	responseutils.SetDownloadHeader(w, filename)
	w.Write(content)
}

// ConfigHandler returns server configuration
func (h *Handlers) ConfigHandler(w http.ResponseWriter, r *http.Request) {
	// Return safe configuration (don't expose sensitive data)
	safeConfig := map[string]interface{}{
		"environment": h.server.config.Environment,
		"host":        h.server.config.Host,
		"port":        h.server.config.Port,
		"timeouts": map[string]interface{}{
			"read":  h.server.config.ReadTimeout.String(),
			"write": h.server.config.WriteTimeout.String(),
			"idle":  h.server.config.IdleTimeout.String(),
		},
		"features": map[string]bool{
			"cors":        h.server.config.EnableCORS,
			"compression": h.server.config.EnableCompression,
			"metrics":     h.server.config.EnableMetrics,
			"health":      h.server.config.EnableHealth,
		},
		"uptime":  h.server.health.GetUptime(),
		"version": "1.0.0",
	}

	responseutils.Success(w, safeConfig, "Server configuration")
}

// MiddlewareTestHandler tests middleware
func (h *Handlers) MiddlewareTestHandler(w http.ResponseWriter, r *http.Request) {
	data := map[string]interface{}{
		"message": "Middleware test successful",
		"headers": map[string][]string{
			"X-Request-ID":    {w.Header().Get("X-Request-ID")},
			"X-Response-Time": {w.Header().Get("X-Response-Time")},
		},
		"request_context": map[string]interface{}{
			"request_id": r.Context().Value("request_id"),
			"user":       r.Context().Value("user"),
		},
		"timestamp": time.Now().Format(time.RFC3339),
	}

	responseutils.Success(w, data, "Middleware test endpoint")
}

// StressTestHandler for performance testing
func (h *Handlers) StressTestHandler(w http.ResponseWriter, r *http.Request) {
	// Simulate some processing
	time.Sleep(100 * time.Millisecond)

	response := map[string]interface{}{
		"status":    "ok",
		"timestamp": time.Now().UnixNano(),
		"message":   "Stress test completed",
	}

	responseutils.Success(w, response, "Stress test endpoint")
}
