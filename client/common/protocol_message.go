package common

import (
	"fmt"
	"strings"
)

type BetMessage struct {
	Agency    string
	FirstName string
	LastName  string
	Document  string
	Birthdate string
	Number    string
}

func NewBetMessage(agency, firstName, lastName, document, birthdate, number string) *BetMessage {
	return &BetMessage{
		Agency:    agency,
		FirstName: firstName,
		LastName:  lastName,
		Document:  document,
		Birthdate: birthdate,
		Number:    number,
	}
}

func (m *BetMessage) Serialize() []byte {
	betData := fmt.Sprintf("AGENCY=%s,FIRST_NAME=%s,LAST_NAME=%s,DOCUMENT=%s,BIRTHDATE=%s,NUMBER=%s",
		m.Agency,
		m.FirstName,
		m.LastName,
		m.Document,
		m.Birthdate,
		m.Number)

	return []byte(betData + "\n")
}

type ResponseMessage struct {
	Result  string
	Message string
}

func ParseResponseMessage(response string) *ResponseMessage {
	response = strings.TrimSpace(response)
	parts := strings.Split(response, ",")

	var result, message string

	if len(parts) > 0 && strings.HasPrefix(parts[0], "RESULT=") {
		result = strings.TrimPrefix(parts[0], "RESULT=")
	}

	if len(parts) > 1 && strings.HasPrefix(parts[1], "MESSAGE=") {
		message = strings.TrimPrefix(parts[1], "MESSAGE=")
	}

	return &ResponseMessage{
		Result:  result,
		Message: message,
	}
}
