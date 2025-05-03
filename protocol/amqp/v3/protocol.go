/*
 Copyright 2021 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

package amqp

import (
	"context"
	"fmt"

	"github.com/Azure/go-amqp"

	"github.com/cloudevents/sdk-go/v2/binding"
	"github.com/cloudevents/sdk-go/v2/protocol"
)

type Protocol struct {
	connOpts         *amqp.ConnOptions
	sessionOpts      *amqp.SessionOptions
	senderLinkOpts   []amqp.SenderOptions
	receiverLinkOpts []amqp.ReceiverOptions

	// AMQP
	conn        *amqp.Conn
	session     *amqp.Session
	ownedClient bool
	server      string
	queue       string

	// Sender
	sender                  *sender

	// Receiver
	receiver *receiver
}

// NewProtocol creates a new amqp transport with required server and queue parameters.
// Additional options can be provided using functional options.
func NewProtocol(ctx context.Context, server string, queue string, opts ...Option) (*Protocol, error) {
	t, err := initProtocol(ctx, server, queue, opts...)
	if err != nil {
		return nil, err
	}

	// Initialize link options arrays if not set
	if t.senderLinkOpts == nil {
		t.senderLinkOpts = []amqp.SenderOptions{}
	}
	if t.receiverLinkOpts == nil {
		t.receiverLinkOpts = []amqp.ReceiverOptions{}
	}

	// Setup sender
	if err := t.setupSender(ctx); err != nil {
		if t.ownedClient {
			_ = t.conn.Close()
		}
		return nil, err
	}

	// Setup receiver
	if err := t.setupReceiver(ctx); err != nil {
		_ = t.sender.amqp.Close(ctx)
		if t.ownedClient {
			_ = t.conn.Close()
		}
		return nil, err
	}

	return t, nil
}

// NewSenderProtocol creates a new sender-only amqp transport.
func NewSenderProtocol(ctx context.Context, server string, queue string, opts ...Option) (*Protocol, error) {
	t, err := initProtocol(ctx, server, queue, opts...)
	if err != nil {
		return nil, err
	}

	// Initialize sender link options if not set
	if t.senderLinkOpts == nil {
		t.senderLinkOpts = []amqp.SenderOptions{}
	}

	// Setup sender
	if err := t.setupSender(ctx); err != nil {
		if t.ownedClient {
			_ = t.conn.Close()
		}
		return nil, err
	}

	return t, nil
}

// NewReceiverProtocol creates a new receiver-only amqp transport.
func NewReceiverProtocol(ctx context.Context, server string, queue string, opts ...Option) (*Protocol, error) {
	t, err := initProtocol(ctx, server, queue, opts...)
	if err != nil {
		return nil, err
	}

	// Initialize receiver link options if not set
	if t.receiverLinkOpts == nil {
		t.receiverLinkOpts = []amqp.ReceiverOptions{}
	}

	// Setup receiver
	if err := t.setupReceiver(ctx); err != nil {
		if t.ownedClient {
			_ = t.conn.Close()
		}
		return nil, err
	}

	return t, nil
}

func (t *Protocol) Send(ctx context.Context, in binding.Message, transformers ...binding.Transformer) error {
	return t.sender.Send(ctx, in, transformers...)
}

func (t *Protocol) Receive(ctx context.Context) (binding.Message, error) {
	return t.receiver.Receive(ctx)
}

// GetServer returns the server URL
func (t *Protocol) GetServer() string {
	return t.server
}

// GetQueue returns the queue address
func (t *Protocol) GetQueue() string {
	return t.queue
}

func (t *Protocol) Close(ctx context.Context) error {
	if t.sender != nil {
		return t.sender.amqp.Close(ctx)
	}

	if t.receiver != nil {
		return t.receiver.amqp.Close(ctx)
	}

	if t.ownedClient {
		return t.conn.Close()
	}
	return nil
}

var _ protocol.Sender = (*Protocol)(nil)
var _ protocol.Receiver = (*Protocol)(nil)
var _ protocol.Closer = (*Protocol)(nil)

func (t *Protocol) applyOptions(opts ...Option) error {
	for _, fn := range opts {
		if err := fn(t); err != nil {
			return err
		}
	}
	return nil
}

// initProtocol initializes the common parts of the Protocol struct
func initProtocol(ctx context.Context, server, queue string, opts ...Option) (*Protocol, error) {
	t := &Protocol{
		server: server,
		queue:  queue,
	}

	if err := t.applyOptions(opts...); err != nil {
		return nil, err
	}

	// Validate required fields
	if t.server == "" {
		return nil, fmt.Errorf("server address is required")
	}

	if t.queue == "" {
		return nil, fmt.Errorf("queue address is required")
	}

	// Initialize connection options if not set
	if t.connOpts == nil {
		t.connOpts = &amqp.ConnOptions{}
	}

	// Initialize session options if not set
	if t.sessionOpts == nil {
		t.sessionOpts = &amqp.SessionOptions{}
	}

	// Create the client if not provided
	if t.conn == nil {
		client, err := amqp.Dial(ctx, t.server, t.connOpts)
		if err != nil {
			return nil, fmt.Errorf("could not dial: %w", err)
		}
		t.conn = client
		t.ownedClient = true
	}

	// Open a session if not provided
	if t.session == nil {
		session, err := t.conn.NewSession(ctx, t.sessionOpts)
		if err != nil {
			if t.ownedClient {
				_ = t.conn.Close()
			}
			return nil, fmt.Errorf("could not create session: %w", err)
		}
		t.session = session
	}

	return t, nil
}

// setupSender creates and configures a sender for the protocol
func (t *Protocol) setupSender(ctx context.Context) error {
	var senderOpts amqp.SenderOptions
	// Apply any custom sender options
	for _, opt := range t.senderLinkOpts {
		if opt.Name != "" {
			senderOpts.Name = opt.Name
		}
		if opt.Properties != nil {
			senderOpts.Properties = opt.Properties
		}
		if opt.SettlementMode != nil {
			senderOpts.SettlementMode = opt.SettlementMode
		}
		if opt.RequestedReceiverSettleMode != nil {
			senderOpts.RequestedReceiverSettleMode = opt.RequestedReceiverSettleMode
		}
		// Add other fields as needed
	}

	amqpSender, err := t.session.NewSender(ctx, t.queue, &senderOpts)
	if err != nil {
		return err
	}
	t.sender = newSender(amqpSender).(*sender)
	return nil
}

// setupReceiver creates and configures a receiver for the protocol
func (t *Protocol) setupReceiver(ctx context.Context) error {
	var receiverOpts amqp.ReceiverOptions
	// Apply any custom receiver options
	for _, opt := range t.receiverLinkOpts {
		if opt.Name != "" {
			receiverOpts.Name = opt.Name
		}
		if opt.Properties != nil {
			receiverOpts.Properties = opt.Properties
		}
		if opt.Filters != nil {
			receiverOpts.Filters = opt.Filters
		}
		if opt.RequestedSenderSettleMode != nil {
			receiverOpts.RequestedSenderSettleMode = opt.RequestedSenderSettleMode
		}
		if opt.Credit != 0 {
			receiverOpts.Credit = opt.Credit
		}
	}

	amqpReceiver, err := t.session.NewReceiver(ctx, t.queue, &receiverOpts)
	if err != nil {
		return err
	}
	t.receiver = newReceiver(amqpReceiver).(*receiver)
	return nil
}
