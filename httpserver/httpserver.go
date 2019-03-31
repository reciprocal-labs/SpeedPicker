package httpserver

import (
	"io"
	"net/http"
	"speedPicker/board"
)

func Serve(board *board.Board, addr string) {
	stateHandler := func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, board.String())
	}

	indexHandler := func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, homePage)
	}

	http.HandleFunc("/state", stateHandler)
	http.HandleFunc("/", indexHandler)

	panic(http.ListenAndServe(addr, nil))
}
