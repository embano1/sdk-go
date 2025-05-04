/*
 Copyright 2021 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

package amqp

import (
	"context"

	"github.com/Azure/go-amqp"

	"github.com/cloudevents/sdk-go/v2/binding"
	"github.com/cloudevents/sdk-go/v2/protocol"
)

// sender wraps an amqp.Sender as a binding.Sender
type sender struct {
	amqp *amqp.Sender
}

func (s *sender) Send(ctx context.Context, in binding.Message, transformers ...binding.Transformer) error {
	var err error
	defer func() { _ = in.Finish(err) }()

	// Create empty send options
	var sendOpts amqp.SendOptions

	if m, ok := in.(*Message); ok { // Already an AMQP message.
		/*if len(m.AMQP.Data) == 0 {
			m.AMQP.Data = [][]byte{[]byte("hello")}
		}*/

		err = s.amqp.Send(ctx, m.AMQP, &sendOpts)
		return err
	}

	var amqpMessage amqp.Message
	err = WriteMessage(ctx, in, &amqpMessage, transformers...)
	if err != nil {
		return err
	}

	/*if len(amqpMessage.Data) == 0 {
		amqpMessage.Data = [][]byte{[]byte("hello")}
	}*/
	err = s.amqp.Send(ctx, &amqpMessage, &sendOpts)
	return err
}

// newSender creates a new Sender which wraps an amqp.Sender in a binding.Sender
func newSender(amqpSender *amqp.Sender) protocol.Sender {
	s := &sender{amqp: amqpSender}
	return s
}
