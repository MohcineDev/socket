package main

import (
	"encoding/json"
	"log"
	"net/http"
	"path/filepath"
)

func main() {
	initDB()

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))

	http.HandleFunc("/", handleIndex)
	http.HandleFunc("/login", handleLoginPage)
	http.HandleFunc("/register", handleRegisterPage)

	http.HandleFunc("/ws", handleWebsocket)
	
	http.HandleFunc("/login-submit", handleLogin)
	http.HandleFunc("/register-submit", handleRegister)
	http.HandleFunc("/logout", handleLogout)

	http.HandleFunc("/messages", handleMessages)
	http.HandleFunc("/send", handleSendMesage)

	log.Println("Server started on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func handleIndex(res http.ResponseWriter, req *http.Request) {
	http.ServeFile(res, req, filepath.Join("static", "index.html"))
}

func handleLoginPage(res http.ResponseWriter, req *http.Request) {
	http.ServeFile(res, req, filepath.Join("static", "login.html"))
}

func handleRegisterPage(res http.ResponseWriter, req *http.Request) {
	http.ServeFile(res, req, filepath.Join("static", "register.html"))
}

func handleLogin(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Redirect(res, req, "/login", http.StatusSeeOther)
		return
	}

	username := req.FormValue("username")
	password := req.FormValue("password")

	var dbPassword string
	err := db.QueryRow("SELECT password FROM users WHERE username = ?", username).Scan(&dbPassword)
	if err != nil || password != dbPassword {
		http.Error(res, "Invalid credentials", http.StatusUnauthorized)
		return
	}
	http.SetCookie(res, &http.Cookie{
		Name:  "session",
		Value: username,
		Path:  "/",
	})

	http.Redirect(res, req, "/", http.StatusSeeOther)
}

func handleRegister(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Redirect(res, req, "/register", http.StatusSeeOther)
		return
	}

	username := req.FormValue("username")
	password := req.FormValue("password")

	if username == "" || password == "" {
		http.Error(res, "Missing username or password", http.StatusBadRequest)
		return
	}

	_, err := db.Exec("INSERT INTO users(username, password) VALUES(?,?)", username, password)
	if err != nil {
		http.Error(res, "Username already exist", http.StatusConflict)
		return
	}

	http.SetCookie(res, &http.Cookie{
		Name:  "session",
		Value: username,
		Path:  "/",
	})

	http.Redirect(res, req, "/", http.StatusSeeOther)
}

func handleLogout(res http.ResponseWriter, req *http.Request) {
	// ///clear the session cookie
	http.SetCookie(res, &http.Cookie{
		Name:   "session",
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	})

	http.Redirect(res, req, "/login", http.StatusSeeOther)
}

func handleMessages(res http.ResponseWriter, req *http.Request) {
	rows, err := db.Query("SELECT username, message, created_at FROM messages ORDER BY created_at ASC")
	if err != nil {
		http.Error(res, "Failed to fetch messages", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	type chatMessage struct {
		Username  string `json:"username"`
		Message   string `json:"message"`
		Timestamp string `json:"created_at"`
	}

	var messages []chatMessage
	for rows.Next() {
		var msg chatMessage
		err := rows.Scan(&msg.Username, &msg.Message, &msg.Timestamp)
		if err != nil {
			continue
		}
		messages = append(messages, msg)
	}
	res.Header().Set("Content-Type", "application/json")
	json.NewEncoder(res).Encode(messages)
}

func handleSendMesage(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(res, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}
	username := getSessionUser(req)
	message := req.FormValue("message")

	if username == "" || message == "" {
		http.Error(res, "Messing Data", http.StatusBadRequest)
		return
	}
	_, err := db.Exec("INSERT INTO messages(username, message) VALUES(?, ?)", username, message)
	if err != nil {
		http.Error(res, "failed to save message", http.StatusInternalServerError)
		return
	}
	res.WriteHeader(http.StatusOK)
}

// /helper to get the session user
func getSessionUser(req *http.Request) string {
	cookie, err := req.Cookie("session")
	if err != nil {
		return ""
	}

	return cookie.Value
}
