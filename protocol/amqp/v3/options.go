/*
 Copyright 2021 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

package amqp

import (
	"time"

	"github.com/Azure/go-amqp"
)

// Option allows customization of the Protocol
type Option func(*Protocol) error

// WithConnOptions sets connection options for amqp
func WithConnOptions(opt amqp.ConnOptions) Option {
	return func(t *Protocol) error {
		t.connOpts = &opt
		return nil
	}
}

// WithConnSASLPlain sets SASLPlain connection option for amqp
func WithConnSASLPlain(username, password string) Option {
	return func(t *Protocol) error {
		if t.connOpts == nil {
			t.connOpts = &amqp.ConnOptions{}
		}
		t.connOpts.SASLType = amqp.SASLTypePlain(username, password)
		return nil
	}
}

// WithSessionOptions sets session options for amqp
func WithSessionOptions(opt amqp.SessionOptions) Option {
	return func(t *Protocol) error {
		t.sessionOpts = &opt
		return nil
	}
}

// WithSenderLinkOptions sets sender options for amqp
func WithSenderLinkOptions(opt amqp.SenderOptions) Option {
	return func(t *Protocol) error {
		t.senderLinkOpts = append(t.senderLinkOpts, opt)
		return nil
	}
}

// WithReceiverLinkOptions sets receiver options for amqp
func WithReceiverLinkOptions(opt amqp.ReceiverOptions) Option {
	return func(t *Protocol) error {
		t.receiverLinkOpts = append(t.receiverLinkOpts, opt)
		return nil
	}
}

// WithConnection allows providing an existing AMQP client connection
func WithConnection(conn *amqp.Conn) Option {
	return func(t *Protocol) error {
		t.conn = conn
		t.ownedClient = false
		return nil
	}
}

// WithSession allows providing an existing AMQP session
func WithSession(session *amqp.Session) Option {
	return func(t *Protocol) error {
		t.session = session
		return nil
	}
}

// WithMaxFrameSize sets the maximum frame size for the connection
func WithMaxFrameSize(size uint32) Option {
	return func(t *Protocol) error {
		if t.connOpts == nil {
			t.connOpts = &amqp.ConnOptions{}
		}
		t.connOpts.MaxFrameSize = size
		return nil
	}
}

// WithIdleTimeout sets the idle timeout for the connection
func WithIdleTimeout(timeout time.Duration) Option {
	return func(t *Protocol) error {
		if t.connOpts == nil {
			t.connOpts = &amqp.ConnOptions{}
		}
		t.connOpts.IdleTimeout = timeout
		return nil
	}
}

// WithMaxLinks sets the maximum number of links for the session
func WithMaxLinks(maxLinks uint32) Option {
	return func(t *Protocol) error {
		if t.sessionOpts == nil {
			t.sessionOpts = &amqp.SessionOptions{}
		}
		t.sessionOpts.MaxLinks = maxLinks
		return nil
	}
}
