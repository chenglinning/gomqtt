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

type SubAck struct {
	header
	rcodes []ReasonCode
}

var _ Packet = (*SubAck)(nil)

func NewSubAck() *SubAck {
	p := &SubAck{}
	p.ResetProps()
	p.rcodes = make([]ReasonCode, 0)
	return p
}

func (this *SubAck) Unpack(rdata []byte) error {
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

	if this.GetVersion == MQTT50 {
		// property
		err = this.ReadProps(r io.Reader)
		if err != nil {
			logger.Error(fmt.Sprintf("Failed reading properties: %s", err))
			return ErrMalformedStream
		}
	}

	restLen := r.Len()
	if restLen == 0 {
		return CodeProtocolError
	} 
	// reason code for each topic filter
	for restLen > 0 {
		rcode, err := ReadByte(r)
		if err != nil {
			logger.Error(fmt.Sprintf("Error parsing reason code: %s", err))
			return CodeUnspecifiedError
		}
		// verfify reason code 		
		reason_code := ReasonCode(rcode)
		if reason_code.IsValidForType(SUBACK) {
			this.rcodes = append(reason_code)
		} else {
			logger.Error(fmt.Sprintf("Invalid suback reason code: 0x%02X", rcode))
			return ErrMalformedStream
		}
	}
	return nil
}

func (this *SubAck) Pack() ([]byte, error) {
	buff := bytes.NewBuffer([]byte{})
	// packet id
	err = WriteUint16(buff, this.GetPacketID())
	if err != nil {
		logger.Error(fmt.Sprintf("Error encoding packet id: %s", err))
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

	// reason code
	for _, rc := range this.rcodes {
    
		err = WriteByte(buff, rc.Raw())
		if err != nil {
			logger.Error(fmt.Sprintf("Error encoding reason code: %s", err))
			return nil, err
		}
	}

	return buff.Bytes(), nil
}
