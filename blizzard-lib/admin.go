package blizzard

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"mime"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/GeertJohan/go.rice"
	"github.com/gorilla/websocket"
)

type SnapshotRoute struct {
	Path          string
	Version       int
	Process       uintptr
	Requests      int64
	TotalRequests uint64
	Written       uint64
}

type Snapshot struct {
	Execs  []*Executable
	Procs  []*ProcGroup
	Routes []*SnapshotRoute
}

type TemplateResponse struct {
	tpl *template.Template
	err error
}

type assetServer struct {
	*assetServerCh `gen_proc:"gen_server"`
	b              *Blizzard
	box            *rice.Box
	server         *http.Server
	ws             map[*wsConnection]struct{}
}

type wsConnection struct {
	c    *websocket.Conn
	send chan []byte
}

func (c *wsConnection) writeMsg(mt int, payload []byte) error {
	return c.c.WriteMessage(mt, payload)
}

func (c *wsConnection) write() {
	ticker := time.NewTicker(time.Minute)
	defer func() {
		ticker.Stop()
		c.c.Close()
	}()
	for {
		select {
		case message, ok := <-c.send:
			if !ok {
				c.writeMsg(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.writeMsg(websocket.TextMessage, message); err != nil {
				return
			}
		case <-ticker.C:
			if err := c.writeMsg(websocket.PingMessage, []byte{}); err != nil {
				return
			}
		}
	}
}

func (c *wsConnection) read(a *assetServer) {
	defer func() {
		a.Unregister(c)
		c.c.Close()
	}()
	for {
		_, _, err := c.c.ReadMessage()
		if err != nil {
			break
		}
	}
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func NewAssetServer(b *Blizzard) (*assetServer, error) {
	box, err := rice.FindBox("blizzard-assets")
	if err != nil {
		return nil, err
	}
	a := &assetServer{assetServerCh: NewassetServerCh(), b: b, ws: make(map[*wsConnection]struct{}), box: box}
	return a, nil
}

func (a *assetServer) HTTP() {
	a.server = &http.Server{
		Addr:    ":8081",
		Handler: a,
	}
	err := a.server.ListenAndServe()
	if err != nil {
		fatal(err)
	}
}

func (a *assetServer) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	switch req.URL.Path {
	case "/":
		a.serveAsset(resp, "index.html")
	case "/ws":
		a.serveWS(resp, req)
	default:
		a.serveAsset(resp, strings.TrimPrefix(req.URL.Path, "/"))
	}
}

func (a *assetServer) serveWS(resp http.ResponseWriter, req *http.Request) {
	if req.Method != "GET" {
		http.Error(resp, "Method not allowed", 405)
		return
	}
	ws, err := upgrader.Upgrade(resp, req, nil)
	if err != nil {
		log("[admin] %v\n", err)
		return
	}
	log("[admin] new WS connection\n")
	c := &wsConnection{send: make(chan []byte, 256), c: ws}
	a.Register(c)
	go c.write()
	c.read(a)
	log("[admin] lost WS connection\n")
}

func (a *assetServer) serveAsset(resp http.ResponseWriter, path string) {
	mimeType := mime.TypeByExtension(filepath.Ext(path))
	asset, err := a.box.Bytes(path)
	if err != nil {
		resp.WriteHeader(400)
		fmt.Fprint(resp, err)
		return
	}
	if mimeType != "" {
		resp.Header().Set("Content-Type", mimeType)
	}
	resp.Write(asset)
}

func (a *assetServer) handleRegister(c *wsConnection) {
	a.ws[c] = struct{}{}
	log("[admin] registered WS, sending snapshot\n")
	go a.b.Snapshot(func(msg interface{}) {
		b := &bytes.Buffer{}
		encoder := json.NewEncoder(b)
		encoder.Encode(msg)
		c.send <- b.Bytes()
	})
}

func (a *assetServer) handleUnregister(c *wsConnection) {
	delete(a.ws, c)
}

func (a *assetServer) handleBroadcast(msg interface{}) {
	b := &bytes.Buffer{}
	encoder := json.NewEncoder(b)
	encoder.Encode(msg)
	for ws := range a.ws {
		ws.send <- b.Bytes()
	}
}
