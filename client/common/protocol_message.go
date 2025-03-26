package common

import (
	"fmt"
	"strings"
)

const (
	MESSAGE_END             = "\n"
	MESSAGE_BATCH_SEPARATOR = ";"
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

	return []byte(betData + MESSAGE_END)
}

type BatchBetMessage struct {
	Bets []*BetMessage
}

func NewBatchBetMessage() *BatchBetMessage {
	return &BatchBetMessage{
		Bets: make([]*BetMessage, 0),
	}
}

func (m *BatchBetMessage) AddBet(bet *BetMessage) {
	m.Bets = append(m.Bets, bet)
}

func (m *BatchBetMessage) Serialize() []byte {
	if len(m.Bets) == 0 {
		return []byte("BETS=0" + MESSAGE_END)
	}

	result := fmt.Sprintf("BETS=%d;", len(m.Bets))
	for i, bet := range m.Bets {
		betData := fmt.Sprintf("AGENCY=%s,FIRST_NAME=%s,LAST_NAME=%s,DOCUMENT=%s,BIRTHDATE=%s,NUMBER=%s",
			bet.Agency,
			bet.FirstName,
			bet.LastName,
			bet.Document,
			bet.Birthdate,
			bet.Number)

		result += betData
		if i < len(m.Bets)-1 {
			result += MESSAGE_BATCH_SEPARATOR
		}
	}

	return []byte(result + MESSAGE_END)
}

type NoMoreBetsMessage struct {
	AgencyID string
}

func NewNoMoreBetsMessage(agencyID string) *NoMoreBetsMessage {
	return &NoMoreBetsMessage{
		AgencyID: agencyID,
	}
}

func (m *NoMoreBetsMessage) Serialize() []byte {
	msg := fmt.Sprintf("END,AGENCY=%s%s", m.AgencyID, MESSAGE_END)
	return []byte(msg)
}

type GetWinnerMessage struct {
	AgencyID string
}

func NewGetWinnerMessage(agencyID string) *GetWinnerMessage {
	return &GetWinnerMessage{
		AgencyID: agencyID,
	}
}

func (m *GetWinnerMessage) Serialize() []byte {
	msg := fmt.Sprintf("WINNERS,AGENCY=%s%s", m.AgencyID, MESSAGE_END)
	return []byte(msg)
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
