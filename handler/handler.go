package handler

import (
	"encoding/json"
	"fmt"
	"sync"
	"websocket-trial/models"

	"github.com/gofiber/websocket/v2"
	"github.com/rs/zerolog/log"
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
	log.Info().Msg("New client connected. Username:" + username + "Total clients: " + string(len(clients)))
	return client
}

func removeClient(conn *websocket.Conn) {
	clientsMu.Lock()
	if client, exists := clients[conn]; exists {
		log.Error().Msg("Client disconnected. Username: " + client.Username)

		if client.Token != "" {
			log.Error().Msg("Removing token" + client.Token + "from active sessions")
			delete(activeTokens, client.Token)
		}
		if client.UserID != 0 {
			log.Error().Msg("Removing user ID " + string(client.UserID) + " from active sessions")
			delete(activeUsers, client.UserID)
		}
		delete(clients, conn)
	}
	clientsMu.Unlock()
	log.Error().Msg("Active sessions - Tokens: " + string(len(activeTokens)) + ", Users: " + string(len(activeUsers)))
}

func broadcastMessage(message string, sender *websocket.Conn) {
	clientsMu.RLock()
	defer clientsMu.RUnlock()

	senderClient, exists := clients[sender]
	if !exists {
		log.Error().Msg("Error: sender not found in clients map")
		return
	}

	messageToSend := fmt.Sprintf("%s: %s", senderClient.Username, message)

	for conn, client := range clients {
		if err := conn.WriteMessage(websocket.TextMessage, []byte(messageToSend)); err != nil {
			log.Error().Msg("Error broadcasting message to " + client.Username + ":" + err.Error())
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
			log.Error().Msg("Error reading message:" + err.Error())
			break
		}

		var receivedData map[string]string
		if err := json.Unmarshal(msg, &receivedData); err != nil {
			log.Error().Msg("Invalid message format:" + err.Error())
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
				log.Error().Msg("Invalid token: " + result.Error.Error())
				ws.WriteMessage(websocket.TextMessage, []byte("Invalid authentication token"))
				continue
			}

			if result := H.DB.First(&user, userAccessToken.UserID); result.Error != nil {
				log.Error().Msg("User not found:" + result.Error.Error())
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
					log.Printf("Error sending warning to existing connection:" + err.Error())
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
				log.Error().Msg("Error inserting data into database:" + result.Error.Error())
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
				log.Error().Msg("User not found:" + result.Error.Error())
				ws.WriteMessage(websocket.TextMessage, []byte("User not found"))
				continue
			}

			clientsMu.Lock()
			initialClient.Username = newUsername
			clientsMu.Unlock()

			ws.WriteMessage(websocket.TextMessage, []byte("Username successfully changed"))

		default:
			ws.WriteMessage(websocket.TextMessage, []byte("Unknown message type"))
			log.Error().Msg("Unknown message type:" + messageType)
		}
	}
}
