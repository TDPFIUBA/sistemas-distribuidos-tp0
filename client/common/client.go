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

func (c *Client) SendBetsInBatches() error {
	if err := c.CreateClientSocket(); err != nil {
		return err
	}
	defer c.CloseConnection()

	file, err := os.Open("/data/agency.csv")
	if err != nil {
		return fmt.Errorf("failed to open bets file: %v", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.Comma = ','

	maxBetsPerBatch := c.getMaxBetsPerBatch()

	totalBets, err := c.processBatches(reader, maxBetsPerBatch)
	if err != nil {
		return err
	}

	if totalBets == 0 {
		log.Warningf("action: send_bets | result: no_bets | client_id: %v", c.config.ID)
	}

	log.Infof("action: all_batches_sent | result: success | client_id: %v | total_bets: %d",
		c.config.ID, totalBets)

	if err := c.NotifyNoMoreBets(); err != nil {
		return err
	}

	return c.GetWinnersLoop()
}

func (c *Client) getMaxBetsPerBatch() int {
	maxBetsPerBatch := c.config.BatchMaxAmount
	if maxBetsPerBatch > MAX_BETS_PER_BATCH {
		maxBetsPerBatch = MAX_BETS_PER_BATCH
	}
	return maxBetsPerBatch
}

func (c *Client) processBatches(reader *csv.Reader, maxBetsPerBatch int) (int, error) {
	totalBets := 0
	batchNumber := 1

	for {
		batch, batchSize, err := c.readBatch(reader, maxBetsPerBatch)
		if err != nil {
			return totalBets, err
		}

		if batchSize == 0 {
			break
		}

		totalBets += batchSize

		if err := c.sendBatch(batch, batchNumber); err != nil {
			return totalBets, err
		}

		if batchSize == maxBetsPerBatch {
			time.Sleep(c.config.LoopPeriod)
		}

		batchNumber++
	}

	return totalBets, nil
}

func (c *Client) readBatch(reader *csv.Reader, maxBets int) (*BatchBetMessage, int, error) {
	batch := NewBatchBetMessage()
	batchSize := 0

	for batchSize < maxBets {
		betInfo, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, 0, fmt.Errorf("error reading bets file: %v", err)
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

		batch.AddBet(bet)
		batchSize++
	}

	return batch, batchSize, nil
}

func (c *Client) sendBatch(batch *BatchBetMessage, batchNumber int) error {
	if err := c.protocol.SendMessage(batch.Serialize()); err != nil {
		log.Errorf("action: send_batch | result: fail | client_id: %v | batch: %d | error: %v",
			c.config.ID, batchNumber, err)
		return err
	}

	responseData, err := c.protocol.ReceiveMessage()
	if err != nil {
		log.Errorf("action: receive_message | result: fail | client_id: %v | batch: %d | error: %v",
			c.config.ID, batchNumber, err)
		return err
	}

	responseMsg := ParseResponseMessage(responseData)
	log.Debugf("action: bet_response | result: %s | message: %s | batch: %d",
		responseMsg.Result, responseMsg.Message, batchNumber)

	log.Infof("action: batch_sent | result: %s | client_id: %v | batch: %d",
		responseMsg.Result, c.config.ID, batchNumber)

	return nil
}

func (c *Client) NotifyNoMoreBets() error {

	msg := NewNoMoreBetsMessage(c.config.ID)
	if err := c.protocol.SendMessage(msg.Serialize()); err != nil {
		log.Errorf("action: no_more_bets_send | result: fail | client_id: %v | error: %v",
			c.config.ID, err)
		return err
	}

	data, err := c.protocol.ReceiveMessage()
	if err != nil {
		log.Errorf("action: no_more_bets_receive | result: fail | client_id: %v | error: %v",
			c.config.ID, err)
		return err
	}

	responseMsg := ParseResponseMessage(data)
	log.Infof("action: no_more_bets_receive | result: %s | client_id: %v | message: %s",
		responseMsg.Result, c.config.ID, responseMsg.Message)

	return nil
}

func (c *Client) GetWinnersLoop() error {
	for {
		winners, err := c.GetWinners()
		if err != nil {
			return err
		}

		if winners {
			return nil
		}

		time.Sleep(c.config.LoopPeriod)
	}
}

func (c *Client) GetWinners() (bool, error) {

	queryMsg := NewGetWinnerMessage(c.config.ID)
	if err := c.protocol.SendMessage(queryMsg.Serialize()); err != nil {
		log.Errorf("action: get_winners | result: fail | client_id: %v | error: %v",
			c.config.ID, err)
		return false, err
	}

	responseData, err := c.protocol.ReceiveMessage()
	if err != nil {
		log.Errorf("action: get_winners_receive | result: fail | client_id: %v | error: %v",
			c.config.ID, err)
		return false, err
	}

	winnersMsg := ParseResponseMessage(responseData)
	if winnersMsg.Result == "success" {
		log.Infof("action: consulta_ganadores | result: success | cant_ganadores: %s",
			winnersMsg.Message)

		return true, nil
	}

	return false, nil
}

func (c *Client) StartClientLoop() {
	c.SendBetsInBatches()
	log.Infof("action: loop_finished | result: success | client_id: %v", c.config.ID)
}
