package main

import (
	"encoding/json"
	"io"
	"net/http"
	"os"

	"github.com/coder/websocket"
	"golang.org/x/sys/unix"
)

func main() {
	http.HandleFunc("GET /serial/{ttyFile}", func(w http.ResponseWriter, r *http.Request) {
		// WebSocket
		conn, err := websocket.Accept(w, r, nil)
		if err == nil {
			ctx := r.Context()
			netConn := websocket.NetConn(ctx, conn, websocket.MessageBinary)
			file, err := os.OpenFile("/dev/"+r.PathValue("ttyFile"), os.O_RDWR|unix.O_NOCTTY, 0)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			defer file.Close()
			_ = setRaw(int(file.Fd()), unix.B9600)
			go io.Copy(file, netConn)
			io.Copy(netConn, file)
			return
		}

		// normal
		scheme := "ws"
		if r.TLS != nil {
			scheme = "wss"
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(struct {
			URL string `json:"url"`
		}{
			URL: scheme + "://" + r.Host + r.URL.Path,
		})
	})
	http.ListenAndServe(":6901", nil)
}

// シリアルポートを raw モード（エコー無効・改行変換無効）に設定する関数
func setRaw(fd int, baud uint32) error {
	termios, err := unix.IoctlGetTermios(fd, unix.TCGETS)
	if err != nil {
		return err
	}
	// cfmakeraw 相当の設定（エコー・カノニカルモード・改行変換をすべてオフ）
	termios.Iflag &^= unix.IGNBRK | unix.BRKINT | unix.PARMRK | unix.ISTRIP | unix.INLCR | unix.IGNCR | unix.ICRNL | unix.IXON
	termios.Oflag &^= unix.OPOST
	termios.Lflag &^= unix.ECHO | unix.ECHONL | unix.ICANON | unix.ISIG | unix.IEXTEN
	termios.Cflag &^= unix.CSIZE | unix.PARENB
	termios.Cflag |= unix.CS8 | unix.CREAD | unix.CLOCAL

	// ボーレート設定（9600bps の場合）
	termios.Cflag &^= unix.CBAUD
	termios.Cflag |= baud

	termios.Cc[unix.VMIN] = 1
	termios.Cc[unix.VTIME] = 0

	return unix.IoctlSetTermios(fd, unix.TCSETS, termios)
}
