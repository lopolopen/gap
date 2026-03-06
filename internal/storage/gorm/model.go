package gorm

import (
	"database/sql"
	"encoding/json"
	"time"

	"github.com/lopolopen/gap/internal/entity"
	"github.com/lopolopen/gap/internal/enum"
)

type Model struct {
	ID        uint        `gorm:"primarykey"`
	CreatedAt time.Time   `gorm:"not null"`
	Version   string      `gorm:"not null;size:16"`
	Topic     string      `gorm:"not null;size:256"`
	Status    enum.Status `gorm:"not null"`
	Headers   string      `gorm:"type:text"`
	Payload   string      `gorm:"type:longtext"`
	Retries   int         `gorm:"default:0"`
	ExpiredAt sql.Null[time.Time]
}

//go:generate go tool shoot map -way=-> -path=../../entity -to=Envelope,Envelope -type=Published,Received

type Published struct {
	Model
}

func (p *Published) writeEntity(e *entity.Envelope) {
	if p.Headers != "" {
		var headers map[string]interface{}
		err := json.Unmarshal([]byte(p.Headers), &headers)
		if err != nil {
			panic(err)
		}
		e.Headers = headers
	}
	e.Message = nil
	e.Group = ""
	e.Tag = nil
}

type Received struct {
	Model
	Group string `gorm:"not null;size:128"`
}

func (r *Received) writeEntity(e *entity.Envelope) {
	if r.Headers != "" {
		var headers map[string]interface{}
		err := json.Unmarshal([]byte(r.Headers), &headers)
		if err != nil {
			panic(err)
		}
		e.Headers = headers
	}
	e.Message = nil
	e.Tag = nil
}
