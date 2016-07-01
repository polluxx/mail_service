package smtp_listener

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"memdb"
	"models"
	"net"
	"net/mail"
	"time"

	"github.com/mhale/smtpd"
)

const listenPort = 2525

/**
Create new SMTP listener

@return void
*/
func Listen() {
	log.Println("[SMTP]: Start listening on port :2525")
	// create server
	smtpd.ListenAndServe(fmt.Sprintf("127.0.0.1:%d", listenPort), mailHandler, "SMTPListener", "")
}

/**
Handler for SMTP server

@return void
*/
func mailHandler(origin net.Addr, from string, to []string, data []byte) {

	// read message from request data
	msg, _ := mail.ReadMessage(bytes.NewReader(data))
	// get mail subject
	subject := msg.Header.Get("Subject")
	// read body raw data
	body, err := ioutil.ReadAll(msg.Body)
	if err != nil {
		log.Printf("Error on parsing mail: %v", err.Error())
		return
	}

	// create new data model
	insertMessage := &models.Message{
		To:           to[0],
		From:         from,
		Subject:      subject,
		Body:         string(body[:]),
		ReceivedDate: time.Now(),
	}

	// get DB instance
	instance := memdb.GetInstance()

	// trying to insert new message
	response := instance.InsertMessage(insertMessage.To, insertMessage)

	if !response.Success {
		log.Printf("Error on posting message: %v", response.Error)
		return
	}

	log.Printf("Received mail from %s for %s with subject %s", from, to[0], subject)
}
