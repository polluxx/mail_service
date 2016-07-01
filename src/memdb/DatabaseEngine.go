package memdb

import (
	"fmt"
	"models"
	"time"

	. "github.com/ahmetalpbalkan/go-linq"
)

//Engine structure which contains data
type engine struct {
	store            map[string]*models.MailBox
	mailbox_sequance *int
	message_sequance *int
	chanel           chan command
}

//Save new mailbox
func (e engine) insertMailbox(command *command) {
	//Increase sequance
	*e.mailbox_sequance++

	//Create mailbox
	mailBox := &models.MailBox{
		Id:       *e.mailbox_sequance,
		Address:  fmt.Sprintf(command.key, *e.mailbox_sequance),
		Messages: make([]*models.Message, 0)}

	//Save mailbox
	e.store[mailBox.Address] = mailBox

	//Send result of saving
	command.result <- CommandResult{Success: true, Rows: 1, Value: mailBox}
}

//Save new message
func (e engine) insertMessage(command *command) {
	//If mailbox is inexisting
	if e.store[command.key] == nil {
		command.result <- CommandResult{
			Success: false,
			Rows:    0,
			Error:   "There is not mailbox for the message"}
		return
	}

	//Increase sequance
	*e.message_sequance++

	//Assign id to message
	message := command.value.(*models.Message)
	message.Id = *e.message_sequance

	//Save message to store
	e.store[command.key].Messages = append(e.store[command.key].Messages, message)

	//Send result of saving
	command.result <- CommandResult{Success: true, Rows: 1, Value: message}
}

//Delete message
func (e engine) deleteMessage(command *command) {
	//If mailbox is inexisting
	if e.store[command.key] == nil {
		command.result <- CommandResult{
			Success: false,
			Rows:    0,
			Error:   "There is not mailbox for the message"}
		return
	}

	//Find message`s index
	elementIndex := -1
	for index, element := range e.store[command.key].Messages {
		if element.Id == command.id {
			elementIndex = index
			break
		}
	}

	//if index is not found
	if elementIndex == -1 {
		command.result <- CommandResult{
			Success: false,
			Rows:    0,
			Error:   "There is not message"}
		return
	}

	//Delete message
	e.store[command.key].Messages = append(e.store[command.key].Messages[:elementIndex], e.store[command.key].Messages[elementIndex+1:]...)

	//Send result of deleting
	command.result <- CommandResult{Success: true, Rows: 1}
}

//Delete mailbox
func (e engine) deleteMailbox(command *command) {
	//If mailbox is inexisting
	if e.store[command.key] == nil {
		command.result <- CommandResult{
			Success: false,
			Rows:    0,
			Error:   "There is not mailbox to delete"}

		return
	}

	//Delete mailbox
	delete(e.store, command.key)

	//Send result of deleting
	command.result <- CommandResult{Success: true, Rows: 1}
}

//Select single message
func (e engine) selectMessage(command *command) {
	//If mailbox is inexisting
	if e.store[command.key] == nil {
		command.result <- CommandResult{
			Success: false,
			Rows:    0,
			Error:   "There is not such mailbox"}
		return
	}

	//Searc for message
	message, found, _ := From(e.store[command.key].Messages).FirstBy(func(s T) (bool, error) {
		return s.(*models.Message).Id == command.id, nil
	})

	//if message not found
	if !found {
		command.result <- CommandResult{
			Success: false,
			Rows:    0,
			Error:   "Message is not found"}
		return
	}

	//Send message
	command.result <- CommandResult{
		Success: true,
		Rows:    1,
		Value:   message}
}

//Select list of messages
func (e engine) selectMessages(command *command) {
	//If mailbox is inexisting
	if e.store[command.key] == nil {
		command.result <- CommandResult{
			Success: false,
			Rows:    0,
			Error:   "There is not such mailbox"}
		return
	}

	//Select messages using cursor-pagination
	messagesT, _ := From(e.store[command.key].Messages).OrderBy(func(o1, o2 T) bool {
		return o1.(*models.Message).Id > o2.(*models.Message).Id
	}).SkipWhile(func(o T) (bool, error) {
		//Cursor pagination
		if command.page.MaxId == nil {
			return false, nil
		}

		return o.(*models.Message).Id >= *command.page.MaxId, nil
	}).Take(command.page.Count).Results()

	//Convert Linq.T to *models.Mesage
	messages := make([]*models.Message, len(messagesT))
	for index, messageT := range messagesT {
		messages[index] = messageT.(*models.Message)
	}

	//Send messages
	command.result <- CommandResult{
		Success: true,
		Rows:    len(messages),
		Value:   messages}

}

//Deletes all messages which are not relevant
func (e engine) clearNotRelevantMessages(command *command) {
	deletedMessages := 0

	//Go over all mailboxes
	for _, box := range e.store {
		//Go over all messages in mailbox
		for i := len(box.Messages) - 1; i >= 0; i-- {
			duration := command.value.(time.Duration)

			//if message is not expired
			if box.Messages[i].ReceivedDate.Add(duration).After(time.Now()) {
				continue
			}

			//Delete message
			box.Messages = append(box.Messages[:i], box.Messages[i+1:]...)
			deletedMessages++
		}
	}

	//Send response
	command.result <- CommandResult{
		Success: true,
		Rows:    deletedMessages}
}

//Database engine
//Handles all commands in infinite loop
func (e engine) Run() {
	defer close(e.chanel)

	//Loop through messages
	for command := range e.chanel {
		//Choose needed action
		switch command.action {
		case action_insert:
			//Choose needed entity
			switch command.entity {
			case entity_mailbox:
				e.insertMailbox(&command)
			case entity_message:
				e.insertMessage(&command)
			}
		case action_delete:
			//Choose needed entity
			switch command.entity {
			case entity_mailbox:
				e.deleteMailbox(&command)
			case entity_message:
				e.deleteMessage(&command)
			}
		case action_get:
			e.selectMessage(&command)
		case action_filter:
			e.selectMessages(&command)
		case action_clearnotrelevant:
			e.clearNotRelevantMessages(&command)
		}
	}
}

//Constructor of database engine
func newEngine(chanel chan command) *engine {
	//Initialize data
	engine := new(engine)
	engine.store = make(map[string]*models.MailBox)
	engine.mailbox_sequance = new(int)
	engine.message_sequance = new(int)
	engine.chanel = chanel
	return engine
}
