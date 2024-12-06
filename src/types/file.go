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
	Request   string `json:"request"`
	Response  string `json:"response"`
	Image     string `json:"image"`
	Text      string `json:"text"`
	Data      any    `json:"data"`
	Result    string `json:"result"`
	Timestamp string `json:"timestamp"`
}
