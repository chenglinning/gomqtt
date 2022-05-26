package mqttp

import (
	"github.com/wonderivan/logger"
	"encoding/binary"
	"regexp"
	"unicode/utf8"
	"bytes"
	"io"
)

type UnSubscribe struct {
	header
	TopicList []string
}

var _ Packet = (*UnSubscribe)(nil)

func NewUnSubscribe() *UnSubscribe {
	p := &UnSubscribe{}
	p.ResetProps()
	p.TopicList = make([]string , 0)
	return p
}

func (this *UnSubscribe) Unpack(rdata []byte) error {
	r := bytes.NewBuffer(rdata)
	// packet id
	pid, err := ReadUint16(r)
	if err != nil {
		logger.Error(fmt.Sprintf("Error parsing packet ID: %s", err))
		return CodeUnspecifiedError
	}
	if pid == 0 {
		logger.Error(fmt.Sprintf("Invalid packet ID: %d", pid))
		return CodeProtocolError
	}
	this.SetPacketID(pid)

	// property
	if this.GetVersion == MQTT50 {
		err = this.ReadProps(r io.Reader)
		if err != nil {
			logger.Error(fmt.Sprintf("Failed reading properties: %s", err))
			return CodeProtocolError
		}
	}

	restLen := r.Len()
	if restLen == 0 {
		return CodeProtocolError
	} 

	for restLen > 0 {
		// topic filter
		topic, err := ReadUTF8String(r)
		if err != nil {
			logger.Error(fmt.Sprintf("Error parsing topic filter: %s", err))
			return CodeUnspecifiedError
		}
		this.TopicList = append(this.TopicList, topic)
	}

	return nil
}

func (this *UnSubscribe) Pack() ([]byte, error) {
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

	// topic filter list
	for _, topic := range this.TopicList {
		// topic filter
		err = WriteString(buff, topic)
		if err != nil {
			logger.Error(fmt.Sprintf("Error encoding topic filter: %s", err))
			return nil, err
		}
	}

	return buff.Bytes(), nil
}
