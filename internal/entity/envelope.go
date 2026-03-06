package entity

import "encoding/json"

type Envelope struct {
	ID      uint
	Headers map[string]any
	Version string
	Topic   string
	Group   string
	Message any
	Payload []byte
	Retries int
	Tag     any
}

func NewEnvelope(version string, topic string, msg any) *Envelope {
	return &Envelope{
		Version: version,
		Topic:   topic,
		Message: msg,
	}
}

func (e *Envelope) WithGroup(group string) *Envelope {
	e.Group = group
	return e
}

func (e *Envelope) WithPayload(payload []byte) *Envelope {
	e.Payload = payload
	return e
}

func (e *Envelope) WithTag(tag any) *Envelope {
	e.Tag = tag
	return e
}

func (e *Envelope) AddHeader(key string, value any) {
	if e.Headers == nil {
		e.Headers = make(map[string]any)
	}
	e.Headers[key] = value
}

func (e *Envelope) PayloadBytes() ([]byte, error) {
	if len(e.Payload) != 0 {
		return e.Payload, nil
	}
	if e.Message == nil {
		return nil, nil
	}
	var err error
	e.Payload, err = json.Marshal(e.Message)
	if err != nil {
		return nil, err
	}
	return e.Payload, nil
}

func (e *Envelope) HeadersBytes() ([]byte, error) {
	if len(e.Headers) == 0 {
		return nil, nil
	}
	return json.Marshal(e.Headers)
}
