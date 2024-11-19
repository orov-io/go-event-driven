package main

import (
	"github.com/ThreeDotsLabs/watermill/message"
	log "github.com/sirupsen/logrus"
)

func Logger(next message.HandlerFunc) message.HandlerFunc {
	return func(msg *message.Message) ([]*message.Message, error) {
		log.WithField("message_uuid", msg.UUID).Info("Handling a message")
		return next(msg)
	}
}
