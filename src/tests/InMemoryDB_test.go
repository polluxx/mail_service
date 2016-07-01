package tests

import (
	"memdb"
	"models"
	"strconv"
	"strings"
	"testing"
	"time"
)

//Test if Inmemory Database is used as single instance
func Test_InMemoryDb_Singleton(t *testing.T) {
	instance1 := memdb.GetInstance()
	instance2 := memdb.GetInstance()

	if instance1 != instance2 {
		t.Error("Two separate instances are used instead of single one")
	}
}

//Test if mailboxes are generated successfully
//Every new mailbox has to have bigger id
func Test_InMemoryDb_Generating_Mailboxes(t *testing.T) {
	db := memdb.GetInstance()

	for i := 1; i < 10; i++ {
		result := db.InsertMailBox()

		if !result.Success {
			t.Error("Operation is not successfully")
		}

		if result.Rows != 1 {
			t.Error("Count of affected rows is not equal to 1")
		}

		mailbox := result.Value.(*models.MailBox)

		if mailbox.Id != i || !strings.Contains(mailbox.Address, strconv.Itoa(i)) {
			t.Error("Mailbox is generated not correctly")
		}
	}
}

//Test adds message to inexistent mailbox.
//Message shouldn`t be added
func Test_InMemoryDb_Add_Message_To_InExisted_Mailboxe(t *testing.T) {
	db := memdb.GetInstance()

	message := &models.Message{To: "some@email.com"}

	result := db.InsertMessage("unexisted@email.com", message)

	if result.Success {
		t.Error("Operation is successfully, but shouldn`t be")
	}

	if result.Rows != 0 {
		t.Error("Count of affected rows is not equal to 0")
	}
}

//Test adds message to existed mailbox
func Test_InMemoryDb_Add_Message_To_Existing_Mailbox(t *testing.T) {
	db := memdb.GetInstance()

	message := &models.Message{To: "some@email.com"}

	result := db.InsertMessage("email_1@some.domain", message)

	if !result.Success {
		t.Error("Operation is not successfully")
	}

	if result.Rows != 1 {
		t.Error("Count of affected rows is not equal to 1")
	}

	messageFromDb := result.Value.(*models.Message)

	if messageFromDb.Id <= 0 || messageFromDb.To != message.To {
		t.Error("Wrong message is added")
	}
}

//Test deletes inexisted mailbox
//Any mailbox shouldn`t be deleted
func Test_InMemoryDb_Delete_Inexisting_Mailbox(t *testing.T) {
	db := memdb.GetInstance()

	result := db.DeleteMailBox("inexisted@email.domain")

	if result.Success {
		t.Error("Operation is successfully, but shouldn`t be")
	}

	if result.Rows != 0 {
		t.Error("Count of affected rows is not equal to 0")
	}
}

//Test deletes existing mailbox
func Test_InMemoryDb_Delete_Existing_Mailbox(t *testing.T) {
	db := memdb.GetInstance()

	result := db.DeleteMailBox("email_9@some.domain")

	if !result.Success {
		t.Error("Operation is not successfully")
	}

	if result.Rows != 1 {
		t.Error("Count of affected rows is not equal to 1")
	}
}

//Test gets existing message by id
func Test_InMemoryDb_Get_Message_By_Id(t *testing.T) {
	db := memdb.GetInstance()

	message := &models.Message{To: "client@email.com"}

	resultInsert := db.InsertMessage("email_1@some.domain", message)

	if !resultInsert.Success {
		t.Error("Insert operation is not successfully")
	}

	result := db.GetMessage("email_1@some.domain", resultInsert.Value.(*models.Message).Id)

	if !result.Success {
		t.Error("Select operation is not successfully")
	}

	if result.Rows != 1 {
		t.Error("Count of affected rows is not equal to 1")
	}

	if result.Value.(*models.Message).Id != resultInsert.Value.(*models.Message).Id {
		t.Error("The same message has different IDs in few instances")
	}
}

//Test gets inexisting message by id
func Test_InMemoryDb_Get_Inexisting_Message_By_Id(t *testing.T) {
	db := memdb.GetInstance()

	result := db.GetMessage("email_1@some.domain", 54364)

	if result.Success {
		t.Error("Select operation is successfully, but shouldn`t")
	}

	if result.Rows != 0 {
		t.Error("Count of affected rows is not equal to 0")
	}
}

//Test gets inexisting message by id from inexisting mailbox
func Test_InMemoryDb_Get_Inexisting_Message_From_Inexisting_Mailbox_By_Id(t *testing.T) {
	db := memdb.GetInstance()

	result := db.GetMessage("inexisting@mail.box", 54364)

	if result.Success {
		t.Error("Select operation is successfully, but shouldn`t")
	}

	if result.Rows != 0 {
		t.Error("Count of affected rows is not equal to 0")
	}
}

//Get slice of messages and validate order and len
func Test_InMemoryDb_Get_Messages(t *testing.T) {
	db := memdb.GetInstance()

	for i := 1; i < 10; i++ {
		message := &models.Message{To: "client@email.com"}
		db.InsertMessage("email_3@some.domain", message)
	}

	page1 := &memdb.PageCursor{Count: 3}

	result1 := db.GetMailBoxMessages("email_3@some.domain", page1)

	messages1 := result1.Value.([]*models.Message)

	if !(messages1[0].Id > messages1[1].Id && messages1[1].Id > messages1[2].Id) {
		t.Error("Wrong order of messages is returned")
	}

	page2 := &memdb.PageCursor{Count: 3, MaxId: &messages1[2].Id}

	result2 := db.GetMailBoxMessages("email_3@some.domain", page2)

	messages2 := result2.Value.([]*models.Message)

	if !(messages2[0].Id > messages2[1].Id && messages2[1].Id > messages2[2].Id) {
		t.Error("Wrong order of messages is returned")
	}

	if !(messages1[2].Id > messages2[0].Id) {
		t.Error("Wrong order of messages is returned")
	}
}

//Test deletes inexisted message
func Test_InMemoryDb_Delete_Inexisting_Message(t *testing.T) {
	db := memdb.GetInstance()

	result := db.DeleteMessage("email_4@some.domain", 54655)

	if result.Success {
		t.Error("Operation is successfully, but shouldn`t be")
	}

	if result.Rows != 0 {
		t.Error("Count of affected rows is not equal to 0")
	}
}

//Test deletes existing message
func Test_InMemoryDb_Delete_Existing_Message(t *testing.T) {
	db := memdb.GetInstance()

	message := &models.Message{To: "some@email.com"}

	insertResult := db.InsertMessage("email_4@some.domain", message)

	if !insertResult.Success {
		t.Error("Insert operation is not successfully")
	}

	messageid := insertResult.Value.(*models.Message).Id

	deleteResult := db.DeleteMessage("email_4@some.domain", messageid)

	if !deleteResult.Success {
		t.Error("Delete operation is not successfully")
	}

	if deleteResult.Rows != 1 {
		t.Error("Count of affected rows is not equal to 1")
	}

	result := db.GetMessage("email_4@some.domain", messageid)

	if result.Success {
		t.Error("Select operation is not successfully")
	}

	if result.Rows != 0 {
		t.Error("Message wasn`t deleted")
	}
}

//Test deletes not relevant messages
func Test_InMemoryDb_Clear_Not_Relevant_Messages(t *testing.T) {
	db := memdb.GetInstance()

	duration := 2 * time.Minute

	message1 := &models.Message{To: "some@email.com", ReceivedDate: time.Now().Add(-3 * time.Minute)}
	message2 := &models.Message{To: "some@email.com", ReceivedDate: time.Now().Add(-1 * time.Minute)}

	insertResult1 := db.InsertMessage("email_6@some.domain", message1)
	insertResult2 := db.InsertMessage("email_6@some.domain", message2)

	if !insertResult1.Success {
		t.Error("Insert operation is not successfully")
	}

	if !insertResult2.Success {
		t.Error("Insert operation is not successfully")
	}

	clearResult := db.ClearNotRelevantMessages(duration)

	if !clearResult.Success {
		t.Error("Clear operation is not successfully")
	}

	selectResult1 := db.GetMessage("email_6@some.domain", insertResult1.Value.(*models.Message).Id)

	if selectResult1.Success {
		t.Error("Select operation is successfully, but shouldn`t")
	}

	if selectResult1.Rows != 0 {
		t.Error("Count of affected rows is not equal to 0")
	}

	selectResult2 := db.GetMessage("email_6@some.domain", insertResult2.Value.(*models.Message).Id)

	if !selectResult2.Success {
		t.Error("Select operation is not successfully")
	}

	if selectResult2.Rows != 1 {
		t.Error("Count of affected rows is not equal to 1")
	}

}
