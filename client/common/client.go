package common

import (
	"net"
	"time"

	"github.com/op/go-logging"
)

var log = logging.MustGetLogger("log")

// ClientConfig Configuration used by the client
type ClientConfig struct {
	ID            string
	ServerAddress string
	LoopAmount    int
	LoopPeriod    time.Duration
	FirstName     string
	LastName      string
	Document      string
	Birthdate     string
	Number        string
}

// Client Entity that encapsulates client behavior
type Client struct {
	config   ClientConfig
	conn     net.Conn
	protocol *Protocol
}

// NewClient Initializes a new client receiving the configuration
// as a parameter
func NewClient(config ClientConfig) *Client {
	client := &Client{
		config: config,
	}
	return client
}

// CreateClientSocket Initializes client socket. In case of
// failure, error is printed in stdout/stderr and exit 1
// is returned
func (c *Client) CreateClientSocket() error {
	conn, err := net.Dial("tcp", c.config.ServerAddress)
	if err != nil {
		log.Criticalf(
			"action: connect | result: fail | client_id: %v | error: %v",
			c.config.ID,
			err,
		)
		return err
	}
	c.conn = conn
	c.protocol = NewProtocol(c.conn)
	return nil
}

func (c *Client) CloseConnection() {
	if c.conn != nil {
		c.conn.Close()
		log.Infof("action: close_connection | result: success | client_id: %v", c.config.ID)
	}
}

func (c *Client) SendBet() error {
	message := NewBetMessage(
		c.config.ID,
		c.config.FirstName,
		c.config.LastName,
		c.config.Document,
		c.config.Birthdate,
		c.config.Number,
	)

	return c.protocol.SendMessage(message.Serialize())
}

// StartClientLoop Send messages to the client until some time threshold is met
func (c *Client) StartClientLoop() {
	// There is an autoincremental msgID to identify every message sent
	// Messages if the message amount threshold has not been surpassed
	for msgID := 1; msgID <= c.config.LoopAmount; msgID++ {

		if err := c.CreateClientSocket(); err != nil {
			return
		}

		if err := c.SendBet(); err != nil {
			log.Errorf("action: send_bet | result: fail | client_id: %v | error: %v",
				c.config.ID,
				err,
			)
			c.CloseConnection()
			return
		}

		responseData, err := c.protocol.ReceiveMessage()
		if err != nil {
			log.Errorf("action: receive_message | result: fail | client_id: %v | error: %v",
				c.config.ID,
				err,
			)
			c.CloseConnection()
			return
		}

		responseMsg := ParseResponseMessage(responseData)
		log.Debugf("action: bet_response | result: %s | message: %s",
			responseMsg.Result,
			responseMsg.Message)

		log.Infof("action: apuesta_enviada | result: success | dni: %v | numero: %v",
			c.config.Document,
			c.config.Number,
		)

		c.CloseConnection()

		// Wait a time between sending one message and the next one
		time.Sleep(c.config.LoopPeriod)
	}
	log.Infof("action: loop_finished | result: success | client_id: %v", c.config.ID)
}
