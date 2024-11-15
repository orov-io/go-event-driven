package main

import (
	"context"
	"fmt"

	"github.com/ThreeDotsLabs/watermill/message"
)

type AlarmClient interface {
	StartAlarm() error
	StopAlarm() error
}

func ConsumeMessages(sub message.Subscriber, alarmClient AlarmClient) {
	messages, err := sub.Subscribe(context.Background(), "smoke_sensor")
	if err != nil {
		panic(err)
	}

	for msg := range messages {
		var err error
		status := string(msg.Payload)
		switch status {
		case "0":
			err = alarmClient.StopAlarm()
		case "1":
			err = alarmClient.StartAlarm()
		default:
			err = fmt.Errorf("unknown sensor status: %s", status)
			// Treat error, or it will be resend in an infinite loop!
		}

		if err != nil {
			msg.Nack()
		} else {
			msg.Ack()
		}
	}
}
