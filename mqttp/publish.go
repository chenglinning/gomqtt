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

type Publish struct {
	header
	topic     string
	expire_at time.Time
	payload   []byte
}

var _ Packet = (*Publish)(nil)

func NewPublish() *Publish {
	p := &Publish{}
	p.ResetProps()
	return p
}

func (this *Publish) Expired() bool {
	// check if expired 
	if this.expire_at == nil {
		return false
	}
//	now := time.Now()
	return time.Now().After(this.expire_at)
}

func (this *Publish) ExpiredInterval() uint32 {
	pv := propset.GetProperty(Message_Expiry_Interval)
	if pv==nil {
        return 315360000   // 10 years
	}
	return pv.(uint32)
}

func (this *Publish) Unpack(rdata []byte) error {
	r := bytes.NewBuffer(rdata)
	// topic name
	topic, err := ReadUTF8String(r)
	if err != nil {
		logger.Error(fmt.Sprintf("Error parsing topic name: %s", err))
		return CodeUnspecifiedError
	}
	if !IsValidTopic(topic) {
		logger.Error(fmt.Sprintf("Invalid topic name: %s", topic))
		return CodeInvalidTopicName
	}
	this.topic = topic

	// packet id
	if this.GetQoS() > 0 {
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
	}

	// property
	if this.GetVersion == MQTT50 {
		err = this.ReadProps(r)
		if err != nil {
			logger.Error(fmt.Sprintf("Failed reading properties: %s", err))
			return ErrMalformedStream
		}
	}

    // payload
	payload, err := ReadRestData(r)
	if err != nil {
		logger.Error(fmt.Sprintf("Error parsing payload: %s", err))
		return CodeUnspecifiedError
	}
	this.payload = payload
	
	// expire at 
	if this.GetVersion == MQTT50 {
		interval := this.ExpiredInterval()
		duration := time.Duration(interval)*time.Second
		this.expire_at = time.Now().Add(duration)
	} else {
		this.expire_at = nil
	}

	return nil
}

func (this *Publish) Pack() ([]byte, error) {
	buff := bytes.NewBuffer([]byte{})
	// topic
	err = WriteString(buff, this.topic)
	if err != nil {
		logger.Error(fmt.Sprintf("Error encoding topic: %s", err))
		return nil, err
	}
	// packet id
	if this.GetQoS() > 0 {
		err = WriteUint16(buff, this.GetPacketID())
		if err != nil {
			logger.Error(fmt.Sprintf("Error encoding packet id: %s", err))
			return nil, err
		}
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
	_, err = buff.Write(this.payload)
	if err != nil {
		logger.Error(fmt.Sprintf("Error encoding payload: %s", err))
		return nil, err
	}
	
	return buff.Bytes(), nil
}
