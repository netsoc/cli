package util

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/netsoc/cli/pkg/config"
)

type wsError struct {
	Message string `json:"message"`
}

// WebspacedWebsocket opens a websocket to webspaced
func WebspacedWebsocket(c *config.Config, user, endpoint string) (*websocket.Conn, error) {
	url, err := url.Parse(c.URLs.Webspaced)
	if err != nil {
		return nil, fmt.Errorf("failed to parse webspaced URL: %w", err)
	}
	switch url.Scheme {
	case "https":
		url.Scheme = "wss"
	case "http":
		url.Scheme = "ws"
	default:
		return nil, fmt.Errorf("unknown URL scheme %v", url.Scheme)
	}
	url.Path = url.Path + "/webspace/" + user + "/" + endpoint

	headers := make(http.Header)
	headers.Add("Authorization", "Bearer "+c.Token)
	conn, res, err := websocket.DefaultDialer.Dial(url.String(), headers)
	if errors.Is(err, websocket.ErrBadHandshake) {
		var e wsError
		if err := json.NewDecoder(res.Body).Decode(&e); err != nil {
			return nil, err
		}

		return nil, errors.New(e.Message)
	}

	return conn, err
}

// WebsocketIO is a wrapper implementing ReadWriteCloser on top of websocket
type WebsocketIO struct {
	Conn  *websocket.Conn
	Mutex sync.Mutex

	textHandler func(string, *WebsocketIO)
	reader      io.Reader
}

// NewWebsocketIO creates a new websocket ReadWriteCloser wrapper
func NewWebsocketIO(c *websocket.Conn, textHandler func(string, *WebsocketIO)) *WebsocketIO {
	return &WebsocketIO{
		Conn: c,

		textHandler: textHandler,
	}
}

func (w *WebsocketIO) Read(p []byte) (n int, err error) {
	for {
		// First read from this message
		if w.reader == nil {
			var mt int

			mt, w.reader, err = w.Conn.NextReader()
			if err != nil {
				return -1, err
			}

			if mt == websocket.CloseMessage {
				return 0, io.EOF
			}

			if mt == websocket.TextMessage {
				d, err := ioutil.ReadAll(w.reader)
				if err != nil {
					if err == io.EOF {
						err = io.ErrUnexpectedEOF
					}
					return -1, err
				}

				w.textHandler(string(d), w)
				continue
			}
		}

		// Perform the read itself
		n, err := w.reader.Read(p)
		if err == io.EOF {
			// At the end of the message, reset reader
			w.reader = nil
			return n, nil
		}

		if err != nil {
			return -1, err
		}

		return n, nil
	}
}

func (w *WebsocketIO) Write(p []byte) (n int, err error) {
	w.Mutex.Lock()
	defer w.Mutex.Unlock()
	wr, err := w.Conn.NextWriter(websocket.BinaryMessage)
	if err != nil {
		return -1, err
	}
	defer wr.Close()

	n, err = wr.Write(p)
	if err != nil {
		return -1, err
	}

	return n, nil
}

// Close sends a control message indicating the stream is finished, but it does not actually close
// the socket.
func (w *WebsocketIO) Close() error {
	w.Mutex.Lock()
	defer w.Mutex.Unlock()

	// Target expects to get a control message indicating stream is finished.
	w.Conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, "Stream shutting down"))
	return w.Conn.Close()
}
