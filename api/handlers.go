package api

import (
	"github.com/gin-gonic/gin"
	"github.com/grustamli/insider-msg-sender/message"
	"net/http"
	"time"
)

// startSender godoc
// @Description  Initiates the scheduler to begin sending messages at configured intervals.
// @id startSender
// @Tags Scheduler
// @Summary Start message sender
// @Accept json
// @Produce json
// @Success      202  {object}  map[string]string  "OK"
// @Failure      500  {object}  map[string]string  "Internal Server Error"
// @Router       /start [post]
func (s *Server) startSender(c *gin.Context) {
	if err := s.scheduler.Start(c); err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusAccepted, gin.H{
		"message": "Starting sender",
	})
}

// stopSender godoc
// @Summary      Stop the message sender
// @Description  Halts the scheduler, stopping any further message dispatch until restarted.
// @Tags         Scheduler
// @Accept       json
// @Produce      json
// @Success      202  {object}  map[string]string  "Accepted"
// @Failure      500  {object}  map[string]string  "Internal Server Error"
// @Router       /stop [post]
func (s *Server) stopSender(c *gin.Context) {
	if err := s.scheduler.Stop(c); err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusAccepted, gin.H{
		"message": "Stopping sender",
	})
}

// MessageOut represents a message that was sent.
//
// swagger:model MessageOut
type MessageOut struct {
	ID     string    `json:"id"`
	SentAt time.Time `json:"sent_at"`
}

// ListSentMessagesResponse wraps a list of sent messages.
//
// swagger:model ListSentMessagesResponse
type ListSentMessagesResponse struct {
	// items is the array of messages that have been sent.
	Items []*MessageOut `json:"items"`
}

// listSentMessages godoc
// @Summary      List sent messages
// @Description  Retrieve all messages that have been sent, including their IDs and timestamps.
// @Tags         Scheduler
// @Accept       json
// @Produce      json
// @Success      200  {object}  ListSentMessagesResponse
// @Failure      500  {object}  map[string]string  "Internal Server Error"
// @Router       /messages [get]
func (s *Server) listSentMessages(c *gin.Context) {
	sentMessages, err := s.app.ListSentMessages(c)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, ListSentMessagesResponse{
		Items: buildMessageOuts(sentMessages),
	})
}

func buildMessageOuts(messages []*message.SentMessage) []*MessageOut {
	var ret = make([]*MessageOut, len(messages))
	for i, m := range messages {
		ret[i] = &MessageOut{
			ID:     m.MessageID,
			SentAt: m.SentAt,
		}
	}
	return ret
}
