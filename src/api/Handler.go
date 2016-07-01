package api

import (
	"errors"
	"fmt"
	"memdb"
	"models"
	"net/http"
	"regexp"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// messages slice length to retrieve from DB
const pageLimit = 10

/**
Creates api url route handler

@return router *gin.Engine
*/
func Handle() *gin.Engine {
	// create new router instance
	router := gin.Default()

	// describe routes
	router.POST("/mailboxes", mailboxCreate)
	router.DELETE("/mailboxes/:email", mailboxDelete)

	router.GET("/mailboxes/:email/messages", messageList)
	router.POST("/mailboxes/:email/messages", messageAdd)

	router.GET("/mailboxes/:email/messages/:message_id", messageRead)
	router.DELETE("/mailboxes/:email/messages/:message_id", messageRemove)

	return router
}

/**
Create new mailbox
@params email string

@return void
*/
func mailboxCreate(c *gin.Context) {
	// get instance of Db
	instance := memdb.GetInstance()

	// generate new random email
	mailbox := instance.InsertMailBox()
	if !mailbox.Success {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": fmt.Sprintf("Failed adding email: %s", mailbox.Error),
		})
		return
	}

	// gets created Mailbox instance and return
	mailboxInstance := mailbox.Value.(*models.MailBox)
	c.JSON(http.StatusCreated, gin.H{
		"message": fmt.Sprintf("Email %v added", mailboxInstance.Address),
		"mailbox": mailboxInstance.Address,
	})
}

/**
Removes existing mailbox
@params email string

@return void
*/
func mailboxDelete(c *gin.Context) {
	var post models.MailBox
	post.Address = c.Param("email")

	// validate post params
	err := checkMailbox(c, &post)
	if err != nil {
		return
	}

	// get instance of Db
	instance := memdb.GetInstance()
	// trying to remove mailbox
	status := instance.DeleteMailBox(post.Address)
	if !status.Success {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": fmt.Sprintf("Failed removing email: %s", status.Error),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("Email %s removed - %v", post.Address, status.Success),
	})
}

/**
Return list of messages from existing mailbox
@params email string

@return void
*/
func messageList(c *gin.Context) {
	var post models.MailBox
	post.Address = c.Param("email")

	// validate post params
	err := checkMailbox(c, &post)
	if err != nil {
		return
	}

	// get index cursor for paging
	cursorId, _ := strconv.ParseInt(c.DefaultQuery("maxId", "0"), 10, 64)

	maxId := int(cursorId)
	cursor := &memdb.PageCursor{Count: pageLimit}
	if maxId != 0 {
		cursor = &memdb.PageCursor{Count: pageLimit, MaxId: &maxId}
	}

	// get instance of Db
	instance := memdb.GetInstance()
	// get list of messages
	response := instance.GetMailBoxMessages(post.Address, cursor)

	if !response.Success {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": fmt.Sprintf("Failed get messages: %s", response.Error),
			"mailbox": post.Address,
		})
		return
	}

	messages := response.Value.([]*models.Message)

	requestResult := gin.H{
		"status":   http.StatusOK,
		"message":  fmt.Sprintf("Messages from %s", post.Address),
		"messages": messages,
		"count":    response.Rows,
	}

	// check if count results is equal to paging limits
	countResults := response.Rows
	if countResults == pageLimit {
		cursorId = int64(messages[countResults-1].Id)
		requestResult["maxId"] = cursorId
	}

	c.JSON(http.StatusOK, requestResult)
}

/**
Add new message for setted address
@params POST - from string
@params POST - to string
@params POST - subject string
@params POST - message string

@return void
*/
func messageAdd(c *gin.Context) {
	var message models.Message

	// trying to bind post params to data model, validating them
	if c.Bind(&message) != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  http.StatusBadRequest,
			"message": "You must provide all needed fields",
		})
		return
	}

	// check if fields are valid
	if emailValidator(message.To, c) != nil {
		return
	}
	// reciever address check
	if emailValidator(message.From, c) != nil {
		return
	}

	// create new model for insertion
	insertMessage := &models.Message{
		To:           message.To,
		From:         message.From,
		Subject:      message.Subject,
		Body:         message.Body,
		ReceivedDate: time.Now(),
	}

	// get instance of Db
	instance := memdb.GetInstance()
	// trying to insert message
	response := instance.InsertMessage(insertMessage.To, insertMessage)

	if !response.Success {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": fmt.Sprintf("Failed inserting message: %s", response.Error),
			"data":    message,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  http.StatusOK,
		"message": fmt.Sprintf("Message from %s successfully created", message.To),
	})

}

/**
Read message from DB
@params email string
@params message_id string

@return void
*/
func messageRead(c *gin.Context) {
	var messageItem models.MailBox
	messageItem.Address = c.Param("email")
	id, _ := strconv.ParseInt(c.Param("message_id"), 10, 0)
	messageItem.Id = int(id)

	// validate email and message id
	err := checkEmailAndId(messageItem, c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  http.StatusBadRequest,
			"message": err.Error(),
		})
		return
	}

	// get instance of Db
	instance := memdb.GetInstance()
	// trying to get a message using email address and message id
	response := instance.GetMessage(messageItem.Address, messageItem.Id)
	if !response.Success {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": fmt.Sprintf("Failed get message: %s", response.Error),
			"data":    messageItem,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  http.StatusOK,
		"message": "OK",
		"content": response.Value,
	})
}

/**
Remove message from DB
@params email string
@params message_id string

@return void
*/
func messageRemove(c *gin.Context) {
	var messageItem models.MailBox
	messageItem.Address = c.Param("email")
	messageId, _ := strconv.ParseInt(c.Param("message_id"), 10, 0)
	messageItem.Id = int(messageId)

	// validate email and message id
	err := checkEmailAndId(messageItem, c)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  http.StatusBadRequest,
			"message": err.Error(),
		})
		return
	}

	// get instance of Db
	instance := memdb.GetInstance()
	// trying to remove message by email and message id
	response := instance.DeleteMessage(messageItem.Address, messageItem.Id)
	if !response.Success {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": fmt.Sprintf("Failed remove message: %s", response.Error),
			"data":    messageItem,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  http.StatusOK,
		"message": fmt.Sprintf("Removed index %d", messageItem.Id),
	})
}

// HELPERS
/**
Validate email - if not valid - return error and send bad request msg
@params email string

@return Error
*/
func emailValidator(email string, c *gin.Context) error {
	valid := regexp.MustCompile(`([\d\w]+[\.\w\d]*)\+?([\.\w\d]*)?@([\w\d]+[\.\w\d]*){1,}`)
	mess := valid.MatchString(email)
	if !mess {
		message := fmt.Sprintf("Email %s is invalid", email)
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  http.StatusBadRequest,
			"message": message,
		})
		return errors.New(message)
	}

	return nil
}

/**
Check message address and id
@params messageItem models.Mailbox

@return Error
*/
func checkEmailAndId(messageItem models.MailBox, c *gin.Context) error {
	if messageItem.Id == 0 {
		return errors.New("Message ID is invalid")
	}
	// check if email is valid
	err := checkMailbox(c, &messageItem)
	if err != nil {
		return err
	}

	return nil
}

/**
Checks if email is present in request and if it valid

@params c gin.Context - context of request
@params post models.Mailbox - post interface

@return Error
*/
func checkMailbox(c *gin.Context, post *models.MailBox) error {
	// trying to bind post data to data model
	err := c.Bind(post)
	if err == nil {
		// validate email by regular expression
		err = emailValidator(post.Address, c)
		if err != nil {
			return err
		}
	} else {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": fmt.Sprintf("Email is empty"),
		})
		return err
	}
	return nil
}
