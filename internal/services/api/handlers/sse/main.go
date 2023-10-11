package sse

import (
	"encoding/json"
	"github.com/google/jsonapi"
	"gitlab.com/distributed_lab/logan/v3/errors"
	"net/http"
	"strconv"
	"time"
)

const (
	ContentType      = "text/event-stream"
	CacheControl     = "no-cache"
	Connection       = "keep-alive"
	SendEventTimeout = 10 * time.Second
)

func RenderErr(w http.ResponseWriter, errs ...*jsonapi.ErrorObject) {
	if len(errs) == 0 {
		panic("expected non-empty errors slice")
	}

	// getting status of first occurred error
	status, err := strconv.ParseInt(errs[0].Status, 10, 64)
	if err != nil {
		panic(errors.Wrap(err, "failed to parse status"))
	}
	setSSEHeaders(w)

	w.WriteHeader(int(status))
	jsonapi.MarshalErrors(w, errs)
}

func Render(w http.ResponseWriter, res interface{}) {
	setSSEHeaders(w)
	err := json.NewEncoder(w).Encode(res)
	if err != nil {
		panic(errors.Wrap(err, "failed to render response"))
	}
}

func setSSEHeaders(w http.ResponseWriter) {
	// set the Content-Type header for SSE
	w.Header().Set("Content-Type", ContentType)
	w.Header().Set("Cache-Control", CacheControl)
	w.Header().Set("Connection", Connection)
}
