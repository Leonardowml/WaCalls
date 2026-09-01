package main

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"wacalls/internal/voip/media"

	"github.com/coder/websocket"
)

// Caminho de audio pelo WebSocket.
//
// O projeto original leva o audio do navegador ate o servidor por WebRTC, que
// precisa de portas UDP proprias e do IP real de quem chama. Atras de um proxy
// (EasyPanel/Docker Swarm) isso nao se estabelece: o ICE fica em "checking" e
// morre em "failed". Aqui o audio viaja pela mesma conexao HTTPS que serve a
// pagina — sem porta extra, sem firewall, sem NAT no meio.
//
// O caminho WebRTC continua no lugar, para quem roda em rede local.

const wsWriteTimeout = 5 * time.Second

// NewWSBridge aceita a conexao do navegador e devolve um Bridge que fala por ela.
func NewWSBridge(w http.ResponseWriter, r *http.Request, log *slog.Logger) (*Bridge, error) {
	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		CompressionMode: websocket.CompressionDisabled, // PCM nao comprime bem
	})
	if err != nil {
		return nil, err
	}
	conn.SetReadLimit(1 << 20)
	return &Bridge{ws: conn, log: log}, nil
}

// writeWS serializa a escrita: o WebSocket aceita um escritor por vez.
func (b *Bridge) writeWS(pcm []byte) error {
	b.wsMu.Lock()
	defer b.wsMu.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), wsWriteTimeout)
	defer cancel()
	return b.ws.Write(ctx, websocket.MessageBinary, pcm)
}

// readLoop entrega ao servidor o audio captado no microfone do navegador.
// Bloqueia ate a conexao cair — e a queda encerra a chamada.
func (b *Bridge) readLoop(ctx context.Context) {
	defer func() {
		if cb := b.OnTerminalICE; cb != nil {
			cb()
		}
	}()

	for {
		typ, data, err := b.ws.Read(ctx)
		if err != nil {
			b.log.Debug("audio websocket encerrado", "err", err)
			return
		}
		if typ != websocket.MessageBinary || len(data) == 0 {
			continue
		}
		if cb := b.OnBrowserPCM; cb != nil {
			cb(media.PCMInt16LEToFloat32(data))
		}
	}
}
