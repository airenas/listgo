package msgsender

import (
	"log"
)

type MachineMessageSender struct{}

func (ctr MachineMessageSender) Send(message Message) error {
	log.Printf("Sending message %s(%s)\n", message.Queue, message.ID)
	return nil
}
