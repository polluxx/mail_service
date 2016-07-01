package models

type MailBox struct {
	Id       int		`form:"message_id" json:"message_id"`
	Address  string 	`form:"email" json:"email" binding:"required"`
	Messages []*Message
}
