package handler

import (
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
	var username string
	_, payload, err := ws.ReadMessage()
	if err != nil {
		log.Println("Error receiving username:", err)
		username = generateDefaultUsername()
	} else {
		username = string(payload)
		if username == "" {
			username = generateDefaultUsername()
		}
	}

	client := addClient(ws, username)
	defer removeClient(ws)

	for {
		_, msg, err := ws.ReadMessage()
		if err != nil {
			log.Println("Error reading message:", err)
			break
		}
		message.Content = string(msg)

		result := H.DB.Create(&message)

		if result.Error != nil {
			log.Println("Error inserting data to database: ", result.Error)
			continue
		}
		clientsMu.RLock()
		senderUsername := client.Username
		clientsMu.RUnlock()

		fmt.Printf("Message Received from '%s': %s\n", senderUsername, string(msg))
		broadcastMessage(string(msg), ws)

		_, payload, err := ws.ReadMessage()
		if err == nil {
			newUsername := string(payload)
			if newUsername != senderUsername {

				client.Username = newUsername
				clientsMu.Unlock()
			}
		}
	}
}
