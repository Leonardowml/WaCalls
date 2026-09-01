package main

import (
	"log/slog"
	"os"
	"strconv"
	"sync"

	"github.com/pion/webrtc/v4"
)

// O projeto original assume rede local: o servidor anuncia ao navegador o
// endereco interno do container (172.x.x.x), que so e alcancavel de dentro.
// Numa VPS o navegador tenta esse endereco, nao chega, e a conexao morre
// ("browser ice state: checking" -> "closed").
//
// Aqui o servidor passa a anunciar o IP publico (WACALLS_PUBLIC_IP) e a usar
// uma faixa fixa e pequena de portas UDP, que precisa estar liberada.

var (
	apiOnce sync.Once
	apiInst *webrtc.API
)

func webrtcAPI() *webrtc.API {
	apiOnce.Do(func() {
		se := webrtc.SettingEngine{}

		if ip := os.Getenv("WACALLS_PUBLIC_IP"); ip != "" {
			se.SetNAT1To1IPs([]string{ip}, webrtc.ICECandidateTypeHost)
			slog.Info("webrtc: anunciando IP publico ao navegador", "ip", ip)
		} else {
			slog.Warn("webrtc: WACALLS_PUBLIC_IP nao definida — o audio so vai funcionar em rede local")
		}

		lo := envPort("WACALLS_UDP_PORT_MIN", 50000)
		hi := envPort("WACALLS_UDP_PORT_MAX", 50019)
		if hi >= lo {
			if err := se.SetEphemeralUDPPortRange(lo, hi); err != nil {
				slog.Error("webrtc: faixa de portas UDP invalida", "err", err)
			} else {
				slog.Info("webrtc: faixa de portas UDP", "min", lo, "max", hi)
			}
		}

		apiInst = webrtc.NewAPI(webrtc.WithSettingEngine(se))
	})
	return apiInst
}

func envPort(name string, def uint16) uint16 {
	v := os.Getenv(name)
	if v == "" {
		return def
	}
	n, err := strconv.ParseUint(v, 10, 16)
	if err != nil {
		slog.Warn("webrtc: valor invalido, usando padrao", "var", name, "valor", v, "padrao", def)
		return def
	}
	return uint16(n)
}
