package types

import (
	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type Chat struct {
	bun.BaseModel `bun:"table:chat,alias:chat"`
	Id            uuid.UUID `bun:",pk,autoincrement,type:uuid,default:uuid_generate_v4()"`
	Messages      []Message
	Project       string `bun:"projectId"`
}

type Message struct {
	Request   string   `json:"request"`
	Code      string   `json:"code"`
	Followup  []string `json:"followup"`
	Text      string   `json:"text"`
	JSON      any      `json:"json"`
	Image     string   `json:"image"`
	Agent     string   `json:"agent"`
	Error     string   `json:"error"`
	Status    string   `json:"status"`
	Timestamp string   `json:"timestamp"`
}
