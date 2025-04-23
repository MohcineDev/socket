package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	///all domains and ports by default
	CheckOrigin: func(req *http.Request) bool {
		return true
	},
}

type Client struct {
	conn     *websocket.Conn
	send     chan []byte
	username string
}

var (
	clients   = make(map[*Client]bool)
	broadcast = make(chan []byte)
)

func handleWebsocket(res http.ResponseWriter, req *http.Request) {
	username := getSessionUser(req)

	if username == "" {
		fmt.Println("handleWebsocket() : user not logged in")
		http.Error(res, "Unauthorized", http.StatusUnauthorized)
		return
	}

	ws, err := upgrader.Upgrade(res, req, nil)
	if err != nil {
		log.Println("websocket upgrade failed:", err)
		return
	}

	client := &Client{
		conn:     ws,
		send:     make(chan []byte), // ✅ Must be initialized
		username: username,
	}

	clients[client] = true // ✅ Add to the map before using it
	log.Println("✅ Adding client:", client.username)

	// ✅ These two goroutines MUST come next
	go readMessages(client)
	go writeMessages(client)

	// ✅ This test message goes after the goroutines are launched
	log.Println("📤 Sending test message...")
	client.send <- []byte("🧪 Test message from server")

	// Optional: broadcast presence
	broadcast <- []byte(client.username + " has joined the chat.")
}


func readMessages(c *Client) {
	defer func() {
		username := c.username
		broadcast <- []byte(username + " has left the chat.")
		c.conn.Close()
		delete(clients, c)
	}()

	username := c.username
	broadcast <- []byte(username + " has joined the chat.")

	for {
		_, msg, err := c.conn.ReadMessage()
		fmt.Println("msg : ", string(msg))
		if err != nil {
			log.Println("read error:", err)
			break
		}

		if username != "" && string(msg) != "" {

			_, dbErr := db.Exec("INSERT INTO messages(username, message) VALUES(?, ?)", username, string(msg))
			if dbErr != nil {
				log.Println("DB insert error:", dbErr)
			}
		}
		broadcast <- []byte(username + " : " + string(msg))

	}
}

func writeMessages(c *Client) {

	log.Println("🟢 writeMessages started for", c.username)

	fmt.Println("xcccc : ", c)
	for msg := range c.send {
		log.Printf("✉️  Sending message to %s: %s", c.username, msg)
		//check if connn still active
		if c.conn != nil && c.conn.UnderlyingConn() != nil {
			err := c.conn.WriteMessage(websocket.TextMessage, msg)
			if err != nil {
				log.Println("write error:", err)
				break
			}	
		}else{
			log.Printf("⚠️ websocket for %s is closed.\n", c.username)
			break
		}
	}
	log.Println("🔴 Exiting writeMessages for", c.username)
}

func init() {
	go func() {
		for {
			msg := <-broadcast
			log.Println("📣 Broadcasting:", string(msg)) // Add this line
			fmt.Println("------------")
			for client := range clients {
				select {
				case client.send <- msg:
				default:
					log.Printf("⚠️ Client %s is too slow, removing...", client.username)
					delete(clients, client)
					client.conn.Close()
				}			
			}
		}
	}()
}
