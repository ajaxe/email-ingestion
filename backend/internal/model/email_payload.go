package model

import (
	"encoding/json"
	"fmt"
)

type IngestEmailPayload struct {
	SpoolID   string `json:"spool_id"`
	UploadKey string `json:"upload_key"`
}

func (p *IngestEmailPayload) JSON() (string, error) {
	d, e := json.Marshal(p)
	if e != nil {
		return "", e
	}
	return string(d), nil
}
func ParseIngestEmail(payload string) (*IngestEmailPayload, error) {
	if payload == "" {
		return nil, fmt.Errorf("empty payload")
	}
	p := &IngestEmailPayload{}
	b := []byte(payload)
	err := json.Unmarshal(b, p)
	if err != nil {
		return nil, err
	}
	return p, nil
}
