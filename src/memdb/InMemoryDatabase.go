package memdb

import (
	"models"
	"sync"
	"time"
)

//Variables to support singleton
var (
	instance *database = nil
	once     sync.Once
)

//All commands go through this chanel and after it listener processes them
type database struct {
	commands chan command
	engine   *engine
}

//The struct is used to compose command for database engine
type command struct {
	action action
	entity entity
	id     int
	key    string
	value  interface{}
	result chan<- interface{}
	page   *PageCursor
}

//List of possible actions
type action int

const (
	action_insert action = iota
	action_get
	action_filter
	action_delete
	action_clearnotrelevant
)

//List of possible entities in database
type entity int

const (
	entity_mailbox entity = iota
	entity_message
)

//Response from database command
type CommandResult struct {
	Success bool
	Rows    int
	Error   string
	Value   interface{}
}

//The struct is used for pagination purpose
type PageCursor struct {
	MaxId *int
	Count int
}

//Base executing command
//Every public method can use this one to follow current design
func (db database) executeCommand(executingCommand *command) CommandResult {
	//Chanel is used to get response from database
	reply := make(chan interface{})
	defer close(reply)

	//Put chanel to command
	executingCommand.result = reply

	//Send command to database chanel
	db.commands <- *executingCommand

	//Return database response
	return (<-reply).(CommandResult)
}

//Create new mailbox
//There are not input parameters because email address is auto generated
func (db database) InsertMailBox() CommandResult {
	command := &command{action: action_insert, entity: entity_mailbox, key: "email_%d@some.domain"}
	return db.executeCommand(command)
}

//Save new message
func (db database) InsertMessage(address string, message *models.Message) CommandResult {
	command := &command{action: action_insert, entity: entity_message, key: address, value: message}
	return db.executeCommand(command)
}

//Delete mailbox
func (db database) DeleteMailBox(address string) CommandResult {
	command := &command{action: action_delete, entity: entity_mailbox, key: address}
	return db.executeCommand(command)
}

//Delete message
func (db database) DeleteMessage(address string, id int) CommandResult {
	command := &command{action: action_delete, entity: entity_message, key: address, id: id}
	return db.executeCommand(command)
}

//Returns some concreate message
func (db database) GetMessage(address string, id int) CommandResult {
	command := &command{action: action_get, entity: entity_message, key: address, id: id}
	return db.executeCommand(command)
}

//Returns a list of messages for some mailbox
func (db database) GetMailBoxMessages(address string, page *PageCursor) CommandResult {
	command := &command{action: action_filter, entity: entity_message, key: address, page: page}
	return db.executeCommand(command)
}

//Delete all not relevant messages
func (db database) ClearNotRelevantMessages(duration time.Duration) CommandResult {
	command := &command{action: action_clearnotrelevant, entity: entity_message, value: duration}
	return db.executeCommand(command)
}

//Construct and get instance of database
//Only one instance can be constructed(Singleton)
func GetInstance() *database {
	once.Do(func() {
		//Code inside this block is executed only once
		chanel := make(chan command)
		instance = &database{commands: chanel, engine: newEngine(chanel)}

		//Database engine start
		go instance.engine.Run()
	})

	//Return singleton instance
	return instance
}
