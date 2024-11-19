package asyncMiddleware

import (
	"github.com/ThreeDotsLabs/go-event-driven/common/log"
	"github.com/ThreeDotsLabs/watermill/message"
)

// CorrelationIDMetadataKey is used to store the correlation ID in metadata.
const CorrelationIDMetadataKey = "correlation_id"

// SetCorrelationID sets a correlation ID for the message.
//
// SetCorrelationID should be called when the message enters the system.
// When message is produced in a request (for example HTTP),
// message correlation ID should be the same as the request's correlation ID.
func SetCorrelationID(id string, msg *message.Message) {
	if MessageCorrelationID(msg) != "" {
		return
	}

	msg.Metadata.Set(CorrelationIDMetadataKey, id)
}

// MessageCorrelationID returns correlation ID from the message.
func MessageCorrelationID(message *message.Message) string {
	return message.Metadata.Get(CorrelationIDMetadataKey)
}

// CorrelationID adds correlation ID to all messages produced by the handler.
// ID is based on ID from message received by handler.
//
// In order to infer the correlationID, it will does in this order:
//   - Search for the correlation_id message metadata key
//   - Search for the correlationID in the message context, using log.CorrelationIDFromContext.
//     As a consequence of this usage, if correlationID is not found, a new one will be added to the context.
//     Also, It will add the new correlationID to incoming message metadata.
func CorrelationID(h message.HandlerFunc) message.HandlerFunc {
	return func(message *message.Message) ([]*message.Message, error) {
		correlationID := MessageCorrelationID(message)
		if correlationID == "" {
			correlationID = log.CorrelationIDFromContext(message.Context())
			SetCorrelationID(correlationID, message)
		} else {
			message.SetContext(log.ContextWithCorrelationID(message.Context(), correlationID))
		}

		producedMessages, err := h(message)
		for _, msg := range producedMessages {
			SetCorrelationID(correlationID, msg)
		}

		return producedMessages, err
	}
}
