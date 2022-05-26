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

type Disconnect struct {
	header
	rcode	     ReasonCode
}

var _ Packet = (*Disconnnect)(nil)

func NewDisconnect() *Disconnect {
	p := &Disconnect{}
	p.ResetProps()
	return p
}

func (this *Disconnect) Unpack(rdata []byte) error {
	r := bytes.NewBuffer(rdata)
	// reason code
	rcode, err := ReadByte(r)
	if err != nil {
		logger.Error(fmt.Sprintf("Error parsing connack reason code: %s", err))
		return ErrMalformedStream
	}

	reason_code := ReasonCode(rcode)
	if reason_code.IsValidForType(DISCONNECT) {
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

func (this *Disconnect) Pack() ([]byte, error) {
	buff := bytes.NewBuffer([]byte{})
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
