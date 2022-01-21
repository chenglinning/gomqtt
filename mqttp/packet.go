package mqttp

import (
	"fmt"
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
	Pack(io.Writer) error
	Unpack(io.Reader) error
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

	GetRemLen() int32
	SetRemLen(l int32)

	GetFixedHeaderFirstByte() byte
	SetFlags(flags byte) 

	ResetProps()
	ResetWillProps()

	PackProps(io.Writer) error
	UnpackProps(io.Reader) error

	PackWillProps(io.Writer) error
	UnpackWillProps(io.Reader) error

}

func NewPacket(v byte, t byte, flags byte) (Packet, error) {
	var p Packet
	switch t {
	case CONNECT:
		p = NewConnect(v, flags)
	case CONNACK:
		p = NewConnAck(v, flags)
	case PUBLISH:
		p = NewPublish(v, flags)
	case PUBACK:
		p = NewPubAck(v, flags)
	case PUBREC:
		p = NewPubRec(v, flags)
	case PUBREL:
		p = NewPubRel(v, flags)
	case PUBCOMP:
		p = NewPubComp(v, flags)
	case SUBSCRIBE:
		p = NewSubscribe(v, flags)
	case SUBACK:
		p = NewSubAck(v, flags)
	case UNSUBSCRIBE:
		p = NewUnSubscribe(v, flags)
	case UNSUBACK:
		p = NewUnSubAck(v, flags)
	case PINGREQ:
		p = NewPingReq(v, flags)
	case PINGRESP:
		p = NewPingResp(v, flags)
	case DISCONNECT:
		p = NewDisconnect(v, flags)
	case AUTH:
		if v < MQTT50 {
			return nil, ErrInvalidMessageType
		}
		p = NewAuth(v)
	default:
		return nil, ErrInvalidMessageType
	}
	
	p.setType(t)

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
		return nil, err
	}

	pktype := buffer[0] & maskType >> 4
	flags :=  buffer[0] & maskFlags

	remLen, err := ReadUvarint(r)
	if err != nil {
		return nil, err
	}

	pkt, err := NewPacket(TBD, pktype)
	if err != nil {
		return nil, err
	}

	packetBytes := make([]byte, remLen)
	n, err := io.ReadFull(r, packetBytes)
	if err != nil {
		return nil, err
	}
	if n != remLen {
		return nil, errors.New("failed to read expected data")
	}

	err = pkt.Unpack(bytes.NewBuffer(packetBytes))

	return pkt, err
}

func WritePacket(w io.Writer, pkt Packet) error {
	buff := bytes.NewBuffer([]byte{})
	err := pkt.Packet(buff)
	if err != nil {
		return err
	}
	m, err := w.Write(buff.Bytes())
	return err
}

// ValidTopic checks the topic, which is a slice of bytes, to see if it's valid. Topic is
// considered valid if it's longer than 0 bytes, and doesn't contain any wildcard characters
// such as + and #.
func ValidTopic(topic []byte) bool {
	return IsValidUTF(topic) && TopicPublishRegexp.Match(topic)
}
