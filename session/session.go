package session

import (
	"log"

	"time"

	"sync/atomic"

	"sync"

	"github.com/gorilla/websocket"
)

var (
	sessions       = sessionMap{m: make(map[uint64]session)}
	sessionCounter uint64
)

type sessionMap struct {
	m     map[uint64]session
	mutex sync.Mutex
}

type session struct {
	id uint64
	ws
}

type ws struct {
	key  string
	conn *websocket.Conn
}

func NewSession(wsKey string, wsConn *websocket.Conn) {
	session := session{
		id: atomic.AddUint64(&sessionCounter, 1),
		ws: ws{key: wsKey, conn: wsConn}}

	expiration := time.NewTimer(time.Minute * 10)
	go func() {
		<-expiration.C
		log.Printf("session expired, session=%v", session.id)
		session.close()
	}()

	sessions.mutex.Lock()
	defer sessions.mutex.Unlock()

	sessions.m[session.id] = session
}

func BroadcastMessageToSessions(m interface{}) {
	for _, s := range sessions.m {
		if err := s.sendMessage(m); err != nil {
			log.Printf("error broadcasting to session, session=%v connection=%v", s.id, s.ws.key)
		}
	}
}

func TerminateSessions() {
	sessions.mutex.Lock()
	defer sessions.mutex.Unlock()

	sessionCount := len(sessions.m)
	for _, session := range sessions.m {
		session.close()
	}

	log.Printf("all sessions terminated, count=%v", sessionCount)
}

func (s session) close() {
	s.ws.conn.Close()
	delete(sessions.m, s.id)
	log.Printf("session terminated and connection closed, session=%v connection=%v", s.id, s.ws.key)
}

func (s session) sendMessage(m interface{}) error {
	return s.ws.conn.WriteJSON(m)
}
