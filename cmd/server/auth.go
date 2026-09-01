package main

import (
	"crypto/subtle"
	"net/http"
	"os"
)

// Protecao de acesso: sem isso qualquer pessoa que descobrir o endereco
// consegue criar sessao, ligar pelo WhatsApp pareado e ler o historico.
// Usuario e senha vem das variaveis WACALLS_USER e WACALLS_PASSWORD.

func authUser() string {
	if u := os.Getenv("WACALLS_USER"); u != "" {
		return u
	}
	return "admin"
}

// authConfigured informa se a senha foi definida.
func authConfigured() bool { return os.Getenv("WACALLS_PASSWORD") != "" }

func withBasicAuth(h http.Handler) http.Handler {
	user := []byte(authUser())
	pass := []byte(os.Getenv("WACALLS_PASSWORD"))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Preflight de CORS nao carrega credenciais e nao devolve dados.
		if r.Method == http.MethodOptions {
			h.ServeHTTP(w, r)
			return
		}

		u, p, ok := r.BasicAuth()
		okUser := subtle.ConstantTimeCompare([]byte(u), user) == 1
		okPass := subtle.ConstantTimeCompare([]byte(p), pass) == 1
		if !ok || !okUser || !okPass {
			w.Header().Set("WWW-Authenticate", `Basic realm="WaCalls", charset="UTF-8"`)
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		h.ServeHTTP(w, r)
	})
}
