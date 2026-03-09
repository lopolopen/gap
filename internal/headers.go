package internal

import "errors"

const (
	KeysMessageID     = "gap-msg-id"
	KeysMessageType   = "gap-msg-type"
	KeysGroup         = "gap-group"
	KeysCorrelationID = "gap-corr-id"
)

var errInvalidArgs = errors.New("args must be one string map or several string pairs")

type Headers struct {
	headers map[string]string
}

func (h *Headers) Add(args ...any) {
	if len(args) == 0 {
		return
	}

	if len(args) == 1 {
		a, ok := args[0].(map[string]string)
		if ok {
			h.headers = a
			return
		}
	}

	h.headers = make(map[string]string)
	for _, arg := range args {
		a, ok := arg.(pair)
		if !ok {
			panic(errInvalidArgs)
		}
		if _, ok := h.headers[a.Key]; !ok {
			h.headers[a.Key] = a.Value
		}
	}
}

func (h *Headers) Value() map[string]string {
	return h.headers
}
