package mqttp

import (
	"github.com/wonderivan/logger"
	"encoding/binary"
	"regexp"
	"unicode/utf8"
	"bytes"
	"io"
	"fmt"
	"time"
)

type PubAck struct {
	header
	rcode  ReasonCode
}

var _ Packet = (*PubAck)(nil)

func NewPubAck() *PubAck {
	p := &PubAck{}
	p.ResetProps()
	return p
}

func (this *PubAck) Unpack(rdata []byte) error {
	r := bytes.NewBuffer(rdata)
	// packet id
	pid, err := ReadUint16(r)
	if err != nil {
		logger.Error(fmt.Sprintf("Error parsing packet ID: %s", err))
		return CodeUnspecifiedError
	}
	if pid == 0 {
		logger.Error(fmt.Sprintf("Invalid packet ID: %d", pid))
		return CodeUnspecifiedError
	}
	this.SetPacketID(pid)

	// reason code
	rcode, err := ReadByte(r)
	if err != nil {
		logger.Error(fmt.Sprintf("Error parsing puback reason code: %s", err))
		return ErrMalformedStream
	}
	reason_code := ReasonCode(rcode)
	if reason_code.IsValidForType(PUBACK) {
		this.rcode = reason_code
	} else {
		logger.Error(fmt.Sprintf("Invalid puback reason code: 0x%02X", rcode))
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

func (this *PubAck) Pack() ([]byte, error) {
	buff := bytes.NewBuffer([]byte{})
	// packet id
	err = WriteUint16(buff, this.GetPacketID())
	if err != nil {
		logger.Error(fmt.Sprintf("Error encoding packet id: %s", err))
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
