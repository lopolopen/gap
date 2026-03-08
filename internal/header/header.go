package header

const (
	MessageID     = "gap-msg-id"
	MessageType   = "gap-msg-type"
	Group         = "gap-group"
	CorrelationID = "gap-corr-id"
)

func With(headers map[string]string, key, value string) map[string]string {
	if headers == nil {
		headers = make(map[string]string)
	}
	_, ok := headers[key]
	if !ok {
		headers[key] = value
	}
	return headers
}
