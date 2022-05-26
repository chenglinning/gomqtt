package mqttp
import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
)

// FixedHeader is a struct to hold the decoded information from
// the fixed header of an MQTT ControlPacket
type Header struct {
	version    byte
	ptype 	   PKType
	pid		   uint16
	dup        bool
	qos        byte
	retain     bool
	
//	remLen	   int32
	propset      *PropertySet
	willpropset  *PropertySet

}

// Type returns the Packet Type 
func (h *Header) GetType() PKType {
	return h.ptype
}

// Set the Packet Type 
func (h *Header) SetType(t PKType) {
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

// Type returns the Packet fixed header flags byte
func (h *Header) GetFixedHeaderFirstByte() byte {
	return (h.ptype.ToByte()<<4 | boolToByte(h.dup)<<3 | h.qos<<1 | boolToByte(h.retain))
}

// Parse the Packet fixed header flags byte
func (h *Header) ParseFlags(flags byte) {
	h.dup = flags & maskDup > 0
	h.qos = flags & maskQos >> 1
	h.retain = flags & maskRetain > 0
}

// Reset Properties
func (h *Header) ResetProps() {
	h.propset = &PropertySet{ props: make(PropertyMap) }
}

// Reset Will Properties
func (h *Header) ResetWillProps() {
	h.willpropset = &PropertySet{ props: make(PropertyMap) }
}

// Pack Props 
func (h *Header) WriterProps(w io.Writer) error {
	packBytes := h.propset.PackProps(h.ptype)
	if packBytes == nil {
		return errors.New(fmt.Sprintf("There is no property (packet type: 0x%d)", h.ptype))
	}

	pplen := len(packBytes)
	// write property lenght
	err := WriteUvarint(w, uint32(pplen))
	if err != nil {
		return err
	}

	// write property payload
	_, err := w.Write(packBytes)

	return err
}

// Pack Will Props 
func (h *Header) WriteWillProps(w io.Writer) error {
	packBytes := h.willpropset.PackProps(h.ptype)
	if package == nil {
		return errors.New(fmt.Sprintf("There is no property (packet type: 0x%d)", h.ptype))
	}

	pplen := len(packBytes)
	// write property lenght
	err := WriteUvarint(w, uint32(pplen))
	if err != nil {
		return err
	}

	// write property payload
	_, err := w.Write(packBytes)

	return err

}

// Unpack Props
func (h *Header) ReadProps(r io.Reader) error {
	h.ResetProps()
	err := h.propset.UnpackProps(r, h.ptype) 
	return err
}

// Unpack Will Props
func (h *Header) ReadWillProps(r io.Reader) error {
	h.ResetWillProps()
	err := h.willpropset.UnpackProps(r, h.ptype) 
	return err
}

func (h *Header) SubOpsValid(ops byte) bool {
	if h.version == MQTT311 {
		return ops & maskSubscriptionReservedV3 < 3
	}
	if ops & maskSubscriptionReservedV5 > 0 {
		return false
	}
	if ops & maskSubscriptionQoS == 3 {
		return false
	}
	if (ops & maskSubscriptionRetainHandling>>4) == 3 {
		return false
	}
    return true
}