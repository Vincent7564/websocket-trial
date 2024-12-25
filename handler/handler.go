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
}

var (
	clientsMu sync.RWMutex
	clients   = make(map[*websocket.Conn]*Client)
)

func generateDefaultUsername() string {
	return "Guest"
}

func addClient(conn *websocket.Conn, username string) *Client {

	for existingConn, client := range clients {
		if client.Conn == conn {
			delete(clients, existingConn)
			break
		}
	}
	client := &Client{Conn: conn, Username: username}
	clients[conn] = client
	return client
}

func removeClient(conn *websocket.Conn) {

	for existingConn, client := range clients {
		if client.Conn == conn {
			delete(clients, existingConn)
			break
		}
	}
}

func broadcastMessage(message string, sender *websocket.Conn) {
	clientsMu.RLock()
	defer clientsMu.RUnlock()
	for conn := range clients {

		senderUsername := clients[sender].Username
		messageToSend := senderUsername + ": " + message

		if err := conn.WriteMessage(websocket.TextMessage, []byte(messageToSend)); err != nil {
			log.Println("Error broadcasting message:", err)
		}
	}
}

func (H *Handler) EchoServer(ws *websocket.Conn) {
	defer ws.Close()
	var message models.Chat
	var user models.User
	var username string

	_, payload, err := ws.ReadMessage()
	if err != nil || len(payload) == 0 {
		log.Println("Error receiving username or empty payload:", err)
		username = generateDefaultUsername()
	} else {
		var receivedData map[string]string
		err = json.Unmarshal(payload, &receivedData)
		if err != nil {
			log.Println("Error parsing incoming message:", err)
			username = generateDefaultUsername()
		} else {
			username = receivedData["username"]
			if username == "" {
				username = generateDefaultUsername()
			}
		}
	}

	if result := H.DB.Where("username = ?", username).First(&user); result.Error != nil {
		log.Println("User not found:", result.Error)
		ws.WriteMessage(websocket.TextMessage, []byte("User not found"))
		return
	}

	client := addClient(ws, username)
	ws.WriteMessage(websocket.TextMessage, []byte("Username successfully set"))
	defer removeClient(ws)

	for {
		_, msg, err := ws.ReadMessage()
		if err != nil {
			log.Println("Error reading message:", err)
			break
		}

		var receivedData map[string]string
		if err := json.Unmarshal(msg, &receivedData); err != nil {
			log.Println("Invalid message format:", err)
			ws.WriteMessage(websocket.TextMessage, []byte("Invalid message format"))
			continue
		}

		messageType, exists := receivedData["type"]
		if !exists {
			ws.WriteMessage(websocket.TextMessage, []byte("Message type not specified"))
			continue
		}

		switch messageType {
		case "chat":

			content, exists := receivedData["content"]
			if !exists || content == "" {
				ws.WriteMessage(websocket.TextMessage, []byte("Message content cannot be empty"))
				continue
			}

			message.Content = content
			message.Username = client.Username
			if result := H.DB.Create(&message); result.Error != nil {
				log.Println("Error inserting data into database:", result.Error)
				continue
			}

			fmt.Printf("Message received from '%s': %s\n", client.Username, content)
			broadcastMessage(content, ws)

		case "username_change":

			newUsername, exists := receivedData["username"]
			if !exists || newUsername == "" || newUsername == client.Username {
				ws.WriteMessage(websocket.TextMessage, []byte("Invalid or unchanged username"))
				continue
			}

			if result := H.DB.Where("username = ?", newUsername).First(&user); result.Error != nil {
				log.Println("User not found:", result.Error)
				ws.WriteMessage(websocket.TextMessage, []byte("User not found"))
				continue
			}

			client.Username = newUsername
			ws.WriteMessage(websocket.TextMessage, []byte("Username successfully changed"))

		default:
			ws.WriteMessage(websocket.TextMessage, []byte("Unknown message type"))
			log.Println("Unknown message type:", messageType)
		}
	}
}
