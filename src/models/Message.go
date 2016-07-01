package models

import (
	"time"
)

type Message struct {
	Id           int
	ReceivedDate time.Time
	From         string	`form:"from" json:"from" binding:"required"`
	To           string	`form:"to" json:"to" binding:"required"`
	Subject      string	`form:"subject" json:"subject" binding:"required"`
	Body         string	`form:"message" json:"message" binding:"required"`
}
