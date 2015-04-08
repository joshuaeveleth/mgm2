package mgm

import (
  "fmt"
  "github.com/gorilla/websocket"
  "github.com/gorilla/sessions"
  "github.com/satori/go.uuid"
  "net/http"
)

type ClientManager struct {
  authIn chan clientAuth
  authTest chan clientAuthRequest
  authDel chan clientAuth
  store *sessions.CookieStore
}

var upgrader = &websocket.Upgrader{ReadBufferSize: 1024, WriteBufferSize: 1024}

type clientAuth struct {
  guid uuid.UUID 
  token uuid.UUID
  address string
}

type clientAuthRequest struct {
  client clientAuth
  callback chan <- bool
}

func (cm *ClientManager) NewClient(ws *websocket.Conn) *Client {
  fmt.Println("New client constructed")
  client := &Client{send: make(chan []byte, 256), ws: ws}
  return client
}

func (cm *ClientManager) Initialize(sessionKey string){
  cm.store = sessions.NewCookieStore([]byte(sessionKey))
  cm.store.Options = &sessions.Options{
    Path: "/",
    MaxAge: 3600 * 8,
  }
}

func (cm ClientManager) WebsocketHandler(w http.ResponseWriter, r *http.Request) {
  fmt.Println("New connection on ws")
  
  session, _ := cm.store.Get(r, "MGM")
  // test if session exists
  if len(session.Values) == 0 {
    fmt.Println("Websocket closed, existing session missing");
    return
  }
  // test origin, etc for websocket security
  // not sure if necessary, we will be over https, and the session is valid
  
  ws, err := upgrader.Upgrade(w, r, nil)
  if err != nil {
    fmt.Println(err)
    return
  }
  
  c := cm.NewClient(ws)
  c.process()
}