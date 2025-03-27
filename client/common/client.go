package common

import (
	"encoding/csv"
	"fmt"
	"io"
	"net"
	"os"
	"time"

	"github.com/op/go-logging"
)

// MAX_BETS_PER_BATCH: 8KB / 100bytes (1bet) = 80 bets
const (
	CSV_BET_INFO_SIZE  = 5
	MAX_BETS_PER_BATCH = 80
)

var log = logging.MustGetLogger("log")

// ClientConfig Configuration used by the client
type ClientConfig struct {
	ID             string
	ServerAddress  string
	LoopAmount     int
	LoopPeriod     time.Duration
	BatchMaxAmount int
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

func (c *Client) ReadBetsFromCSV() ([]*BetMessage, error) {
	file, err := os.Open("/data/agency.csv")
	if err != nil {
		return nil, fmt.Errorf("failed to open bets file: %v", err)
	}
	defer file.Close()

	bets := make([]*BetMessage, 0)
	reader := csv.NewReader(file)
	reader.Comma = ','

	for {
		betInfo, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("error reading bets file: %v", err)
		}

		if len(betInfo) != CSV_BET_INFO_SIZE {
			log.Warningf("Invalid bet information format, expected %d fields", CSV_BET_INFO_SIZE)
			continue
		}

		bet := NewBetMessage(
			c.config.ID,
			betInfo[0],
			betInfo[1],
			betInfo[2],
			betInfo[3],
			betInfo[4],
		)
		bets = append(bets, bet)
	}

	log.Infof("action: read_bets | result: success | client_id: %v | total_bets: %d",
		c.config.ID, len(bets))

	return bets, nil
}

func (c *Client) SendBetsInBatches() error {
	bets, err := c.ReadBetsFromCSV()
	if err != nil {
		return err
	}

	totalBets := len(bets)
	if totalBets == 0 {
		log.Warningf("action: send_bets | result: no_bets | client_id: %v", c.config.ID)
		return nil
	}

	maxBetsPerBatch := c.config.BatchMaxAmount
	if c.config.BatchMaxAmount > MAX_BETS_PER_BATCH {
		maxBetsPerBatch = MAX_BETS_PER_BATCH
	}

	batchCount := (totalBets + maxBetsPerBatch - 1) / maxBetsPerBatch

	log.Debugf("action: send_bets | batches: %d | total_bets: %d | max_per_batch: %d",
		batchCount, totalBets, maxBetsPerBatch)

	for i := 0; i < batchCount; i++ {

		start := i * maxBetsPerBatch
		end := start + maxBetsPerBatch
		if end > totalBets {
			end = totalBets
		}

		batchMessage := NewBatchBetMessage()
		for j := start; j < end; j++ {
			batchMessage.AddBet(bets[j])
		}

		if err := c.CreateClientSocket(); err != nil {
			return err
		}

		if err := c.protocol.SendMessage(batchMessage.Serialize()); err != nil {
			log.Errorf("action: send_batch | result: fail | client_id: %v | batch: %d/%d | error: %v",
				c.config.ID, i+1, batchCount, err)
			c.CloseConnection()
			return err
		}

		responseData, err := c.protocol.ReceiveMessage()
		if err != nil {
			log.Errorf("action: receive_message | result: fail | client_id: %v | batch: %d/%d | error: %v",
				c.config.ID, i+1, batchCount, err)
			c.CloseConnection()
			return err
		}

		responseMsg := ParseResponseMessage(responseData)
		log.Debugf("action: bet_response | result: %s | message: %s | batch: %d/%d",
			responseMsg.Result, responseMsg.Message, i+1, batchCount)

		log.Infof("action: batch_sent | result: %s | client_id: %v | batch: %d/%d | bets: %d",
			responseMsg.Result, c.config.ID, i+1, batchCount, end-start)

		c.CloseConnection()

		if i < batchCount-1 {
			time.Sleep(c.config.LoopPeriod)
		}
	}

	return nil
}

func (c *Client) StartClientLoop() {
	if err := c.SendBetsInBatches(); err != nil {
		log.Errorf("action: send_bets_in_batches | result: fail | client_id: %v | error: %v",
			c.config.ID, err)
		return
	}

	log.Infof("action: loop_finished | result: success | client_id: %v", c.config.ID)
}
