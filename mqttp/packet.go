package mqttp

import (
	"github.com/wonderivan/logger"
	"fmt"
	"bytes"
	"io"
)

const (
	// maxFixedHeaderLength int    = 5
	maxRemainingLength int32 = (256 * 1024 * 1024) - 1 // 256 MB
)
const (
	//  maskHeaderType  byte = 0xF0
	//  maskHeaderFlags byte = 0x0F
	//  maskHeaderFlagQoS
	maskConnAckSessionPresent byte = 0x01
)

type Packet interface {
	Pack() (byte[], error)
	Unpack(rdata byte[]) error

	String() string

	GetVersion() byte
	SetVersion(v byte)

	GetType() byte
	SetType(t byte)

	GetPacketID() uint16
	SetPacketID(id uint16)

	GetQoS() byte
	SetQos(q byte)

	IsDup() bool
	SetDup(b bool)

	IsRetain() bool
	SetRetain(b bool)

//	GetRemLen() int32
//	SetRemLen(l int32)

	GetFixedHeaderFirstByte() byte
	ParseFlags(flags byte) 

	ResetProps()
	ResetWillProps()

	WriteProps(io.Writer) error
	ReadProps(io.Reader) error

	WriteWillProps(io.Writer) error
	ReadWillProps(io.Reader) error

}

func NewPacket(v byte, t PKType, flags byte) (Packet, error) {
	var p Packet
	switch t {
	case CONNECT:
		p = NewConnect()
	case CONNACK:
		p = NewConnAck()
	case PUBLISH:
		p = NewPublish()
	case PUBACK:
		p = NewPubAck()
	case PUBREC:
		p = NewPubRec()
	case PUBREL:
		p = NewPubRel()
	case PUBCOMP:
		p = NewPubComp()
	case SUBSCRIBE:
		p = NewSubscribe()
	case SUBACK:
		p = NewSubAck()
	case UNSUBSCRIBE:
		p = NewUnSubscribe()
	case UNSUBACK:
		p = NewUnSubAck()
	case PINGREQ:
		p = NewPingReq()
	case PINGRESP:
		p = NewPingResp()
	case DISCONNECT:
		p = NewDisconnect()
	case AUTH:
		if v < MQTT50 {
			return nil, ErrInvalidMessageType
		}
		p = NewAuth()
	default:
		return nil, ErrInvalidMessageType
	}

	if t != PUBLISH {
		dflags := DefaultFlags(t)
		if flags != dflags {
			return nil, ErrMalformedStream
		}
	} 

	p.SetType(t)
	p.ParseFlags(flags)
	p.SetVersion(v)

	return p, nil
}

// ReadPacket takes an instance of an io.Reader (such as net.Conn) and attempts
// to read an MQTT packet from the stream. It returns a Packet
// representing the decoded MQTT packet and an error. One of these returns will
// always be nil, a nil Packet indicating an error occurred.
func ReadPacket(r io.Reader) (Packet, error) {
	buffer := make([]byte, 1)
	_, err := io.ReadFull(r, buffer)
	
	if err != nil {
		logger.Error(err.Error())
		return nil, CodeMalformedPacket
	}

	pktype := PKType(buffer[0] & maskType >> 4)
	flags :=  buffer[0] & maskFlags

	remLen, err := ReadUvarint(r)
	if err != nil {
		logger.Error(err.Error())
		return nil, CodeMalformedPacket
	}

	pkt, err := NewPacket(TBD, pktype, flags)

	if err != nil {
		logger.Error(err.Error())
		return nil, err
	}

	rdata := make([]byte, remLen)
	n, err := io.ReadFull(r, rdata)
	if err != nil {
		logger.Error(err.Error())
		return nil, CodeMalformedPacket
	}
	if n != remLen {
		logger.Error("failed to read remained data")
		return nil, CodeMalformedPacket
	}

	err = pkt.Unpack(rdata)

	return pkt, err
}

func WritePacket(w io.Writer, pkt Packet) error {
	data, err := pkt.Pack()
	if err != nil {
		return err
	}

	// fixed header 1th byte
	fh := pkt.GetFixedHeaderFirstByte()
	err = WriteByte(w, fh)
	if err != nil {
		logger.Error(fmt.Sprintf("Error encoding fixed header: %s", err))
		return err
	}

	// fixed remaining lenght field
	remlen := len(data)
	err = WriteUvarint(w, uint32(remlen))
	if err != nil {
		logger.Error(fmt.Sprintf("Error encoding remianing lenght: %s", err))
		return err
	}
	
	// data
	m, err := w.Write(data)
	return err
}
