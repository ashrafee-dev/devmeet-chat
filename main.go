package main

import (
	"fmt"
	"net/http"
	"html/template"
	"sync"

	"github.com/gorilla/websocket"
)

var username string
var password string

var upgrader = websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}
var clients = make(map[*websocket.Conn]bool)
var mu sync.Mutex

func wsHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	mu.Lock()
	for c := range clients {
		c.WriteMessage(websocket.TextMessage, []byte(`{"join":true}`))
	}
	clients[conn] = true
	mu.Unlock()

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			mu.Lock()
			delete(clients, conn)
			mu.Unlock()
			break
		}
		mu.Lock()
		for c := range clients {
			if c != conn {
				c.WriteMessage(websocket.TextMessage, msg)
			}
		}
		mu.Unlock()
	}
}
func login(w http.ResponseWriter, r *http.Request){
	if r.Method == http.MethodPost{
		r.ParseForm()
		username =r.FormValue("username")
		password =r.FormValue("password")
		fmt.Println(username)
		fmt.Println(password)
		if username == "test" && password == "1234"{
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}
	}
		tmpl, _ := template.ParseFiles("templates/login.html")

	tmpl.Execute(w, nil)

	
} 
func homepage(w http.ResponseWriter, r *http.Request){
	tmpl, _ :=  template.ParseFiles("templates/index.html")

	tmpl.Execute(w, username)

}
func main(){
	mux := http.NewServeMux()
	mux.Handle("/login",http.HandlerFunc(login))
	mux.Handle("/",http.HandlerFunc(homepage))
	mux.Handle("/ws",http.HandlerFunc(wsHandler))
	mux.Handle("/script.js",http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "script.js")
	}))

	server := http.Server{
			Addr: ":8080",
		  Handler: mux,
	}

	server.ListenAndServe()


}

