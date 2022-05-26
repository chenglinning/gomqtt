package mqttp

import (
	"github.com/wonderivan/logger"
	"encoding/binary"
	"regexp"
	"unicode/utf8"
	"bytes"
	"io"
	"fmt"
)

type ConnAck struct {
	header
	flags        byte
	rcode	     ReasonCode
}

var _ Packet = (*ConnAck)(nil)

func NewConnAck() *ConnAck {
	p := &ConnAck{}
	p.ResetProps()
	return p
}

func (this *ConnAck) SessionPresent() bool {
	return (this.flags & maskSessionPresent) != 0
}

func (this *ConnAck) SetSessionPresent(v bool) {
	if v {
		this.flags |= maskSessionPresent
	} else {
		this.flags &= ^maskSessionPresent
	}
}

func (this *ConnAck) Unpack(rdata []byte) error {
	r := bytes.NewBuffer(rdata)
	// connack flags
	flags, err := ReadByte(r)
	if err != nil {
		logger.Error(fmt.Sprintf("Error parsing connack flags: %s", err))
		return ErrMalformedStream
	}
	if flags!=0 && flags!=1 {
		logger.Error(fmt.Sprintf("Invalid connack flags: 0x%02X", flags))
		return ErrMalformedStream
	}
	this.flags = flags

	// reason code
	rcode, err := ReadByte(r)
	if err != nil {
		logger.Error(fmt.Sprintf("Error parsing connack reason code: %s", err))
		return ErrMalformedStream
	}

	reason_code := ReasonCode(rcode)
	if reason_code.IsValidForType(CONNACK) {
		this.rcode = reason_code
	} else {
		logger.Error(fmt.Sprintf("Invalid connack reason code: 0x%02X", rcode))
		return ErrMalformedStream
	}

	if this.GetVersion == MQTT50 {
		// property
		err = this.ReadProps(r io.Reader)
		if err != nil {
			logger.Error(fmt.Sprintf("Failed reading properties: %s", err))
			return ErrMalformedStream
		}
	}

	return nil
}

func (this *ConnAck) Pack() ([]byte, error) {
	buff := bytes.NewBuffer([]byte{})
	// connack flags
	err = WriteByte(buff, this.flags)
	if err != nil {
		logger.Error(fmt.Sprintf("Error encoding connack flags: %s", err))
		return nil, err
	}
	// reason code
	err = WriteByte(buff, this.rcode.Value())
	if err != nil {
		logger.Error(fmt.Sprintf("Error encoding reason code: %s", err))
		return nil, err
	}
	// property
	if this.GetVersion()==MQTT50 {
		ppbytes := this.propset.PackProps(this.GetType())
		// propertiy len
		plen := len(ppbytes)
		err = WriteUvarint(buff, uint32(plen))
		if err != nil {
			logger.Error(fmt.Sprintf("Error encoding property length: %s", err))
			return nil, err
		}
		// property content
		if plen>0 {
			_, err = buff.Write(ppbytes)
			if err != nil {
				logger.Error(fmt.Sprintf("Error encoding property: %s", err))
				return nil, err
			}
		}
	}

	return buff.Bytes(), nil
}
