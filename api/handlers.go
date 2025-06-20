package api

import (
	"github.com/gin-gonic/gin"
	"github.com/grustamli/insider-msg-sender/message"
	"net/http"
	"time"
)

func (s *Server) startSender(c *gin.Context) {
	if err := s.scheduler.Start(c); err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusAccepted, gin.H{
		"message": "Starting sender",
	})
}

func (s *Server) stopSender(c *gin.Context) {
	if err := s.scheduler.Stop(c); err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusAccepted, gin.H{
		"message": "Stopping sender",
	})
}

type MessageOut struct {
	ID     string    `json:"id"`
	SentAt time.Time `json:"sent_at"`
}

func (s *Server) listSentMessages(c *gin.Context) {
	sentMessages, err := s.app.ListSentMessages(c)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"items": buildMessageOuts(sentMessages),
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
