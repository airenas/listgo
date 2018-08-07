package msgsender

import (
	"bitbucket.org/airenas/listgo/internal/pkg/cmdapp"

	"github.com/RichardKnop/machinery/v1"
	"github.com/RichardKnop/machinery/v1/tasks"
	"github.com/pkg/errors"
)

//MachineMessageSender performs messages sending using machinery library
type MachineMessageSender struct {
	Server *machinery.Server
}

//NewMachineMessageSender initializes machinery sender
func NewMachineMessageSender() (*MachineMessageSender, error) {
	server, err := NewMachineryServer()
	if err != nil {
		return nil, errors.Wrap(err, "Can't init machinery")
	}
	return &MachineMessageSender{server}, nil
}

//Send sends the message
func (sender *MachineMessageSender) Send(message *Message) error {
	cmdapp.Log.Infof("Sending message %s(%s)", message.Queue, message.ID)
	decodeTask := tasks.Signature{
		Name: message.Queue,
		Args: []tasks.Arg{newStringArg("ID", message.ID)}}
	_, err := sender.Server.SendTask(&decodeTask)
	if err != nil {
		return errors.Wrap(err, "Can't send message")
	}
	return nil
}

func newStringArg(name string, val string) tasks.Arg {
	return tasks.Arg{Name: name, Type: "string", Value: val}
}
