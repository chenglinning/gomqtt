package mqttp

import (
	"github.com/wonderivan/logger"
	"encoding/binary"
	"regexp"
	"unicode/utf8"
	"bytes"
	"io"
)

// Topic Filter and Subscription Options pair
type TopicOpsPair struct {
	topicFilter  string
	options      SubOps
}

type Subscribe struct {
	header
	topicOpsList [] *TopicOpsPair
}



var _ Packet = (*Subscribe)(nil)

func NewSubscribe() *Subscribe {
	p := &Subscribe{}
	p.ResetProps()
	p.topicOpsList = make([]*TopicOpsPair, 0)
	return p
}

func (this *Subscribe) Unpack(rdata []byte) error {
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
		// sub options
		ops, err := ReadByte(r)
		if err != nil {
			logger.Error(fmt.Sprintf("Error parsing sub options: %s", err))
			return CodeUnspecifiedError
		}
		if this.SubOpsValid(ops) {
			tops := &TopicOpsPair{topicFilter:topic, options: SubOps(ops)}
			this.topicOpsList = append(this.topicOpsList, tops)
			restLen -= (len(topic) + 2 + 1)
		} else {
			logger.Error(fmt.Sprintf("Invalid sub options: 0x%02X", ops))
			return CodeUnspecifiedError
		}
	}

	return nil
}

func (this *Subscribe) Pack() ([]byte, error) {
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

	// topic filter options pairs
	for _, tpo := range this.topicOpsList {
		// topic filter
		err = WriteString(buff, tpo.topicFilter)
		if err != nil {
			logger.Error(fmt.Sprintf("Error encoding topic filter: %s", err))
			return nil, err
		}
        // options
		err = WriteByte(buff, tpo.options.Raw())
		if err != nil {
			logger.Error(fmt.Sprintf("Error encoding options: %s", err))
			return nil, err
		}
	}

	return buff.Bytes(), nil
}
