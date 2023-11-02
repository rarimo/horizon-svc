package sse

import (
	"encoding/json"
	"fmt"
	"github.com/google/jsonapi"
	"gitlab.com/distributed_lab/logan/v3"
	"gitlab.com/distributed_lab/logan/v3/errors"
	"net/http"
	"time"
)

const (
	ContentType      = "text/event-stream"
	CacheControl     = "no-cache"
	Connection       = "keep-alive"
	SendEventTimeout = 5 * time.Second
)

func ToErrorResponse(errs ...*jsonapi.ErrorObject) *jsonapi.ErrorsPayload {
	if len(errs) == 0 {
		panic("expected non-empty errors slice")
	}

	return &jsonapi.ErrorsPayload{Errors: errs}
}

func ServeEvents(w http.ResponseWriter, r *http.Request, makeResponse func() interface{}) {
	SetSSEHeaders(w)

	for {
		resp := makeResponse()
		writeEvent(w, resp)
		w.(http.Flusher).Flush()

		// Check for client disconnection using the context
		select {
		case <-r.Context().Done():
			return
		default:
			time.Sleep(SendEventTimeout)
		}
	}
}

func SetSSEHeaders(w http.ResponseWriter) {
	// set the Content-Type header for SSE
	w.Header().Set("Content-Type", ContentType)
	w.Header().Set("Cache-Control", CacheControl)
	w.Header().Set("Connection", Connection)
}

func writeEvent(w http.ResponseWriter, data interface{}) {
	dataBytes, err := json.Marshal(data)
	if err != nil {
		panic(errors.Wrap(err, "failed to marshal response", logan.F{
			"data": data,
		}))
	}

	_, err = fmt.Fprintf(w, "data: %s\n\n", string(dataBytes))
	if err != nil {
		panic(errors.Wrap(err, "failed to write data", logan.F{
			"data": data,
		}))
	}
}
