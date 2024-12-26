package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"websocket-trial/models"

	"github.com/gofiber/websocket/v2"
	"gorm.io/gorm"
)

type Handler struct {
	DB *gorm.DB
}

type Client struct {
	Conn     *websocket.Conn
	Username string
	Token    string
	UserID   int
}

var (
	clientsMu sync.RWMutex
	clients   = make(map[*websocket.Conn]*Client)

	activeTokens = make(map[string]*websocket.Conn)
	activeUsers  = make(map[int]*websocket.Conn)
)

func addClient(conn *websocket.Conn, username string) *Client {
	client := &Client{
		Conn:     conn,
		Username: username,
	}
	clientsMu.Lock()
	clients[conn] = client
	clientsMu.Unlock()
	log.Printf("New client connected. Username: %s, Total clients: %d", username, len(clients))
	return client
}

func removeClient(conn *websocket.Conn) {
	clientsMu.Lock()
	if client, exists := clients[conn]; exists {
		log.Printf("Client disconnected. Username: %s", client.Username)

		if client.Token != "" {
			log.Printf("Removing token %s from active sessions", client.Token)
			delete(activeTokens, client.Token)
		}
		if client.UserID != 0 {
			log.Printf("Removing user ID %d from active sessions", client.UserID)
			delete(activeUsers, client.UserID)
		}
		delete(clients, conn)
	}
	clientsMu.Unlock()
	log.Printf("Active sessions - Tokens: %d, Users: %d", len(activeTokens), len(activeUsers))
}

func broadcastMessage(message string, sender *websocket.Conn) {
	clientsMu.RLock()
	defer clientsMu.RUnlock()

	senderClient, exists := clients[sender]
	if !exists {
		log.Printf("Error: sender not found in clients map")
		return
	}

	messageToSend := fmt.Sprintf("%s: %s", senderClient.Username, message)

	for conn, client := range clients {
		if err := conn.WriteMessage(websocket.TextMessage, []byte(messageToSend)); err != nil {
			log.Printf("Error broadcasting message to %s: %v", client.Username, err)
		}
	}
}

func (H *Handler) EchoServer(ws *websocket.Conn) {

	initialClient := addClient(ws, "Guest")
	defer ws.Close()
	defer removeClient(ws)

	var message models.Chat
	var user models.User

	for {
		_, msg, err := ws.ReadMessage()
		if err != nil {
			log.Println("Error reading message:", err)
			break
		}

		var receivedData map[string]string
		if err := json.Unmarshal(msg, &receivedData); err != nil {
			log.Println("Invalid message format:", err)
			continue
		}

		messageType := receivedData["type"]
		if messageType == "" {
			ws.WriteMessage(websocket.TextMessage, []byte("Message type not specified"))
			continue
		}

		log.Printf("Received message type: %s", messageType)

		switch messageType {
		case "auth":
			token, exists := receivedData["token"]
			if !exists || token == "" {
				ws.WriteMessage(websocket.TextMessage, []byte("Authentication token missing"))
				continue
			}

			var userAccessToken models.UserAccessToken
			if result := H.DB.Where("token = ?", token).First(&userAccessToken); result.Error != nil {
				log.Println("Invalid token:", result.Error)
				ws.WriteMessage(websocket.TextMessage, []byte("Invalid authentication token"))
				continue
			}

			if result := H.DB.First(&user, userAccessToken.UserID); result.Error != nil {
				log.Println("User not found:", result.Error)
				ws.WriteMessage(websocket.TextMessage, []byte("User not found"))
				continue
			}

			userID := int(user.ID)
			log.Printf("Auth attempt - Token: %s, UserID: %d", token, userID)
			log.Printf("Current active sessions - Tokens: %d, Users: %d", len(activeTokens), len(activeUsers))

			clientsMu.Lock()
			existingConn, existsByToken := activeTokens[token]
			existingUserConn, existsByUser := activeUsers[userID]

			if existsByToken || existsByUser {
				activeConn := existingConn
				if existsByUser {
					activeConn = existingUserConn
				}

				if err := activeConn.WriteMessage(websocket.TextMessage, []byte("Warning: Someone is trying to login to your account from another location")); err != nil {
					log.Printf("Error sending warning to existing connection: %v", err)
				}

				clientsMu.Unlock()

				log.Printf("Sending rejection message and closing connection")
				ws.WriteMessage(websocket.TextMessage, []byte("Account already active in another session"))
				ws.Close()
				return
			}

			activeTokens[token] = ws
			activeUsers[userID] = ws
			initialClient.Username = user.Username
			initialClient.Token = token
			initialClient.UserID = userID

			clients[ws] = initialClient
			clientsMu.Unlock()

			log.Printf("Active sessions - Tokens: %d, Users: %d", len(activeTokens), len(activeUsers))

			statusMessage := fmt.Sprintf("Welcome, %s!", initialClient.Username)
			log.Printf("Sending welcome message: %s", statusMessage)
			ws.WriteMessage(websocket.TextMessage, []byte(statusMessage))

		case "chat":
			content, exists := receivedData["content"]
			if !exists || content == "" {
				ws.WriteMessage(websocket.TextMessage, []byte("Message content cannot be empty"))
				continue
			}

			message.Content = content
			message.Username = initialClient.Username
			if result := H.DB.Create(&message); result.Error != nil {
				log.Println("Error inserting data into database:", result.Error)
				continue
			}

			broadcastMessage(content, ws)

		case "username_change":
			newUsername, exists := receivedData["username"]
			if !exists || newUsername == "" || newUsername == initialClient.Username {
				ws.WriteMessage(websocket.TextMessage, []byte("Invalid or unchanged username"))
				continue
			}

			if result := H.DB.Where("username = ?", newUsername).First(&user); result.Error != nil {
				log.Println("User not found:", result.Error)
				ws.WriteMessage(websocket.TextMessage, []byte("User not found"))
				continue
			}

			clientsMu.Lock()
			initialClient.Username = newUsername
			clientsMu.Unlock()

			ws.WriteMessage(websocket.TextMessage, []byte("Username successfully changed"))

		default:
			ws.WriteMessage(websocket.TextMessage, []byte("Unknown message type"))
			log.Println("Unknown message type:", messageType)
		}
	}
}
