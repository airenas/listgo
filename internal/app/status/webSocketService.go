package status

import (
	"sync"

	"bitbucket.org/airenas/listgo/internal/pkg/cmdapp"
)

var idConnectionMap = make(map[string]map[WsConn]bool)
var connectionIDMap = make(map[WsConn]string)
var mapLock = sync.Mutex{}

//WsConn is interface for websocket handling in status service
type WsConn interface {
	ReadMessage() (messageType int, p []byte, err error)
	Close() error
	WriteJSON(v interface{}) error
}

func handleConnection(conn WsConn) {
	defer deleteConnection(conn)
	defer conn.Close()
	for {
		cmdapp.Log.Infof("handleConnection")
		_, message, err := conn.ReadMessage()
		if err != nil {
			cmdapp.Log.Error(err)
			break
		} else {
			saveConnection(conn, string(message))
		}
	}
	cmdapp.Log.Infof("handleConnection finish")
}

func deleteConnection(conn WsConn) {
	mapLock.Lock()
	defer mapLock.Unlock()
	deleteConnectionNoSync(conn)
}

func deleteConnectionNoSync(conn WsConn) {
	cmdapp.Log.Info("deleteConnection")
	id, found := connectionIDMap[conn]
	if found {
		conns, found := idConnectionMap[id]
		if found {
			delete(conns, conn)
			if len(conns) == 0 {
				delete(idConnectionMap, id)
			}
		}
	}
	delete(connectionIDMap, conn)
	cmdapp.Log.Infof("deleteConnection finish: %d", len(connectionIDMap))
}

func saveConnection(conn WsConn, id string) {
	cmdapp.Log.Infof("saveConnection")
	mapLock.Lock()
	defer mapLock.Unlock()
	deleteConnectionNoSync(conn)
	connectionIDMap[conn] = id
	conns, found := idConnectionMap[id]
	if !found {
		conns = map[WsConn]bool{}
		idConnectionMap[id] = conns
	}
	conns[conn] = true
	cmdapp.Log.Infof("saveConnection finish: %d", len(connectionIDMap))
}

func getConnections(id string) (map[WsConn]bool, bool) {
	r, found := idConnectionMap[id]
	return r, found
}
