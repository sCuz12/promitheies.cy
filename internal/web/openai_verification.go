package web

import (
	"io"
	"net/http"
)

const openAIAppsChallengeToken = "Z_OJ0BhVUo3cJp0zmGGW6lHFa33urKgYWXmCsDAhR7U"

func (s *Server) handleOpenAIAppsChallenge(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = io.WriteString(w, openAIAppsChallengeToken)
}
