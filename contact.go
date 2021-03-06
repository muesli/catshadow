// contact.go - client
// Copyright (C) 2019  David Stainton.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as
// published by the Free Software Foundation, either version 3 of the
// License, or (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package catshadow

import (
	"github.com/katzenpost/channels"
	"github.com/katzenpost/client/session"
	"github.com/katzenpost/core/crypto/rand"
	ratchet "github.com/katzenpost/doubleratchet"
	"github.com/ugorji/go/codec"
)

var cborHandle = new(codec.CborHandle)

type contactExchange struct {
	SpoolWriter       *channels.UnreliableSpoolWriterChannel
	SignedKeyExchange *ratchet.SignedKeyExchange
}

// NewContactExchangeBytes returns serialized contact exchange information.
func NewContactExchangeBytes(spoolWriter *channels.UnreliableSpoolWriterChannel, signedKeyExchange *ratchet.SignedKeyExchange) ([]byte, error) {
	exchange := contactExchange{
		SpoolWriter:       spoolWriter,
		SignedKeyExchange: signedKeyExchange,
	}
	var serialized []byte
	err := codec.NewEncoderBytes(&serialized, cborHandle).Encode(exchange)
	if err != nil {
		return nil, err
	}
	return serialized, nil
}

func parseContactExchangeBytes(contactExchangeBytes []byte) (*contactExchange, error) {
	exchange := new(contactExchange)
	err := codec.NewDecoderBytes(contactExchangeBytes, cborHandle).Decode(exchange)
	if err != nil {
		return nil, err
	}
	return exchange, nil
}

type serializedContact struct {
	ID               uint64
	Nickname         string
	IsPending        bool
	KeyExchange      []byte
	PandaKeyExchange []byte
	PandaResult      string
	Ratchet          []byte
	SpoolWriterChan  *channels.UnreliableSpoolWriterChannel
}

// Contact is a communications contact that we have bidirectional
// communication with.
type Contact struct {
	// id is the local unique contact ID.
	id uint64
	// nickname is also unique locally.
	nickname string
	// isPending is true if the key exchange has not been completed.
	isPending bool
	// keyExchange is the serialised double ratchet key exchange we generated.
	keyExchange []byte
	// pandaKeyExchange is the serialised PANDA key exchange we generated.
	pandaKeyExchange []byte
	// pandaShutdownChan can be closed to trigger the shutdown of a PANDA
	// key exchange worker goroutine.
	pandaShutdownChan chan struct{}
	// pandaResult contains an error message if the PANDA exchange fails.
	pandaResult string

	// ratchet is the client's double ratchet for end to end encryption
	ratchet *ratchet.Ratchet

	// spoolWriterChan is a spool channel we must write to in order to
	// send this contact a message.
	spoolWriterChan *channels.UnreliableSpoolWriterChannel
}

// NewContact creates a new Contact or returns an error.
func NewContact(nickname string, id uint64, spoolReaderChan *channels.UnreliableSpoolReaderChannel, session *session.Session) (*Contact, error) {
	ratchet, err := ratchet.New(rand.Reader)
	if err != nil {
		return nil, err
	}
	signedKeyExchange, err := ratchet.CreateKeyExchange()
	if err != nil {
		return nil, err
	}
	spoolWriterChan := spoolReaderChan.GetSpoolWriter()
	exchange, err := NewContactExchangeBytes(spoolWriterChan, signedKeyExchange)
	if err != nil {
		return nil, err
	}
	return &Contact{
		nickname:          nickname,
		id:                id,
		isPending:         true,
		ratchet:           ratchet,
		keyExchange:       exchange,
		pandaShutdownChan: make(chan struct{}),
	}, nil
}

// ID returns the Contact ID.
func (c *Contact) ID() uint64 {
	return c.id
}

// MarshalBinary does what you expect and returns
// a serialized Contact.
func (c *Contact) MarshalBinary() ([]byte, error) {
	ratchetBlob, err := c.ratchet.MarshalBinary()
	if err != nil {
		return nil, err
	}
	s := &serializedContact{
		ID:               c.id,
		Nickname:         c.nickname,
		IsPending:        c.isPending,
		KeyExchange:      c.keyExchange,
		PandaKeyExchange: c.pandaKeyExchange,
		PandaResult:      c.pandaResult,
		Ratchet:          ratchetBlob,
		SpoolWriterChan:  c.spoolWriterChan,
	}
	var serialized []byte
	err = codec.NewEncoderBytes(&serialized, cborHandle).Encode(s)
	if err != nil {
		return nil, err
	}
	return serialized, nil
}

// UnmarshalBinary does what you expect and initializes
// the given Contact with deserialized Contact fields
// from the given binary blob.
func (c *Contact) UnmarshalBinary(data []byte) error {
	r, err := ratchet.New(rand.Reader)
	if err != nil {
		return err
	}

	s := new(serializedContact)
	err = codec.NewDecoderBytes(data, cborHandle).Decode(s)
	if err != nil {
		return err
	}

	err = r.UnmarshalBinary(s.Ratchet)
	if err != nil {
		return err
	}

	c.id = s.ID
	c.nickname = s.Nickname
	c.isPending = s.IsPending
	c.keyExchange = s.KeyExchange
	c.pandaKeyExchange = s.PandaKeyExchange
	c.pandaResult = s.PandaResult
	c.ratchet = r
	c.spoolWriterChan = s.SpoolWriterChan

	return nil
}
