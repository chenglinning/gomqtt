package mqttp
import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
)
const (
	maskMessageFlags               byte = 0x0F
	maskConnFlagUsername           byte = 0x80
	maskConnFlagPassword           byte = 0x40
	maskConnFlagWillRetain         byte = 0x20
	maskConnFlagWillQos            byte = 0x18
	maskConnFlagWill               byte = 0x04
	maskConnFlagClean              byte = 0x02
	maskConnFlagReserved           byte = 0x01
	maskPublishFlagRetain          byte = 0x01
	maskPublishFlagQoS             byte = 0x06
	maskPublishFlagDup             byte = 0x08
	maskSubscriptionQoS            byte = 0x03
	maskSubscriptionNL             byte = 0x04
	maskSubscriptionRAP            byte = 0x08
	maskSubscriptionRetainHandling byte = 0x30
	maskSubscriptionReservedV3     byte = 0xFC
	maskSubscriptionReservedV5     byte = 0xC0
)

// FixedHeader is a struct to hold the decoded information from
// the fixed header of an MQTT ControlPacket
type Header struct {
	version    byte
	ptype 	   byte
	pid		   uint16
	dup        bool
	qos        byte
	retain     bool
	remLen	   int32
	props      map[uint32]interface{}
	willProps  map[uint32]interface{}
}

// Type returns the Packet Type 
func (h *Header) GetType() byte {
	return h.ptype
}

// Set the Packet Type 
func (h *Header) SetType(t byte) {
	h.pktype = t
}

// Type returns the Packet ID
func (h *Header) GetPacketID() uint16 {
	return h.pid
}

// Set the Packet ID 
func (h *Header) SetPacketID(id uint16) {
	h.pid = id
}

// Type returns the Packet MQTT version 
func (h *Header) GetVersion() byte {
	return h.version
}

// Set the Packet MQTT Version 
func (h *Header) SetVersion(v byte) {
	h.version = v
}

// Type returns the Packet Dup 
func (h *Header) GetDup() bool {
	return h.dup
}

// Set the Packet Dup 
func (h *Header) SetDup(b bool) {
	h.dup = b
}

// Type returns the Packet QoS 
func (h *Header) GetQos() byte {
	return h.qos
}

// Set the Packet QoS 
func (h *Header) SetQos(q byte) {
	h.qos = q
}

// Type returns the Packet Retain flag
func (h *Header) GetRetain() bool {
	return h.retain
}

// Set the Packet Retain flag 
func (h *Header) SetRetain(b bool) {
	h.retain = b
}

// Type returns the Packet Remain Length
func (h *Header) GetRemLen() int32 {
	return h.remLen
}

// Set the Packet Remain Length 
func (h *Header) SetRemLen(l int32) {
	h.remLen = l
}

// Type returns the Packet fixed header flags byte
func (h *Header) GetFixedHeaderFirstByte() byte {
	return (h.ptype<<4 | boolToByte(h.dup)<<3 | h.qos<<1 | boolToByte(h.retain))
}

// Set the Packet fixed header flags byte
func (h *Header) SetFlags(flags byte) {
	h.dup = flags & maskDup > 0
	h.qos = flags & maskQos >> 1
	h.retain = flags & maskRetain > 0
}

// Reset Properties
func (h *Header) ResetProps() {
	h.props = make(map[uint32]interface{})
}

// Reset Will Properties
func (h *Header) ResetWillProps() {
	h.willProps = make(map[uint32]interface{})
}
