package mqttp

import (
	"encoding/binary"
	"unicode/utf8"
	"io"
)

const (
	One_Byte                = iota
	Two_Byte_Integer
	Four_Byte_Integer
	Variable_Byte_Integer
	UTF8_String
	UTF8_String_Pair
	Binary_Data
)

const (
	Payload_Format_Indicator        		= uint32 (0x01)
	Message_Expiry_Interval         		= uint32 (0x02)
	Content_Type                    		= uint32 (0x03)
	Response_Topic                  		= uint32 (0x08)
	Correlation_Data                		= uint32 (0x09)
	Subscription_Identifier         		= uint32 (0x0B)
	Session_Expiry_Interval		            = uint32 (0x11)
	Assigned_Client_Identifier              = uint32 (0x12)
	Server_Keep_Alive                       = uint32 (0x13)
	Authentication_Method                   = uint32 (0x15)
	Authentication_Data                     = uint32 (0x16)
	Request_Problem_Information             = uint32 (0x17)
	Will_Delay_Interval                     = uint32 (0x18)
	Request_Response_Information            = uint32 (0x19)
	Response_Information                    = uint32 (0x1A)
	Server_Reference                        = uint32 (0x1C)
	Reason_String                           = uint32 (0x1F)
	Receive_Maximum                         = uint32 (0x21)
	Topic_Alias_Maximum                     = uint32 (0x22)
	Topic_Alias                             = uint32 (0x23)
	Maximum_QoS                             = uint32 (0x24)
	Retain_Available                        = uint32 (0x25)
	User_Property                           = uint32 (0x26)
	Maximum_Packet_Size                     = uint32 (0x27)
	Wildcard_Subscription_Available         = uint32 (0x28)
	Subscription_Identifier_Available       = uint32 (0x29)
	Shared_Subscription_Available           = uint32 (0x2A)
)

var propertyTypeMap = map[uint32] byte {
	Payload_Format_Indicator:			One_Byte
	Message_Expiry_Interval:            Four_Byte_Integer
	Content_Type:                    	UTF8_String
	Response_Topic:                  	UTF8_String
	Correlation_Data:                	Binary_Data
	Subscription_Identifier:         	Variable_Byte_Integer
	Session_Expiry_Interval:		    Four_Byte_Integer
	Assigned_Client_Identifier:         UTF8_String
	Server_Keep_Alive:                  Two_Byte_Integer
	Authentication_Method:              UTF8_String
	Authentication_Data:                Binary_Data
	Request_Problem_Information:        One_Byte
	Will_Delay_Interval:                Four_Byte_Integer
	Request_Response_Information:       One_Byte
	Response_Information:               UTF8_String
	Server_Reference:                   UTF8_String
	Reason_String:                      UTF8_String
	Receive_Maximum:                    Two_Byte_Integer
	Topic_Alias_Maximum:                Two_Byte_Integer
	Topic_Alias:                        Two_Byte_Integer
	Maximum_QoS:                        One_Byte
	Retain_Available:                   One_Byte
	User_Property:                      UTF8_String_Pair
	Maximum_Packet_Size:                Four_Byte_Integer
	Wildcard_Subscription_Available:    One_Byte
	Subscription_Identifier_Available:  One_Byte
	Shared_Subscription_Available:      One_Byte
}


// nolint: golint
const (
	ErrPropertyNotFound PropertyError = iota
	ErrPropertyInvalidID
	ErrPropertyPacketTypeMismatch
	ErrPropertyTypeMismatch
	ErrPropertyDuplicate
	ErrPropertyUnsupported
	ErrPropertyWrongType
)


// Error description
func (e PropertyError) Error() string {
	switch e {
	case ErrPropertyNotFound:
		return "property: id not found"
	case ErrPropertyInvalidID:
		return "property: id is invalid"
	case ErrPropertyPacketTypeMismatch:
		return "property: packet type does not match id"
	case ErrPropertyTypeMismatch:
		return "property: value type does not match id"
	case ErrPropertyDuplicate:
		return "property: duplicate of id not allowed"
	case ErrPropertyUnsupported:
		return "property: value type is unsupported"
	case ErrPropertyWrongType:
		return "property: value type differs from expected"
	default:
		return "property: unknown error"
	}
}

// StringPair user defined properties
type StringPair struct {
	k string
	v string
}

// propertyAllowedMessageTypes properties and their supported packets type.
// bool flag indicates either duplicate allowed or not
var propertyAllowedMessageTypes = map[uint32]map[byte]bool{
	Payload_Format_Indicator:            {PUBLISH: false},
	Message_Expiry_Interval:             {PUBLISH: false},
	Content_Type:                        {PUBLISH: false},
	Response_Topic:                      {PUBLISH: false},
	Correlation_Data:                    {PUBLISH: false},
	Subscription_Identifier:             {PUBLISH: true, SUBSCRIBE: false},
	Session_Expiry_Interval:             {CONNECT: false, CONNACK: false, DISCONNECT: false},
	Assigned_Client_Identifier:          {CONNACK: false},
	Server_Keep_Alive:                   {CONNACK: false},
	Authentication_Method:               {CONNECT: false, CONNACK: false, AUTH: false},
	Authentication_Data:                 {CONNECT: false, CONNACK: false, AUTH: false},
	Request_Problem_Information:         {CONNECT: false},
	Will_Delay_Interval:                 {PUBLISH: false}, // it is only for Will message
	Request_Response_Information:        {CONNECT: false},
	Response_Information:                {CONNACK: false},
	Server_Reference:                    {CONNACK: false, DISCONNECT: false},

	Reason_String:               { CONNACK: false, PUBACK: false, PUBREC: false, PUBREL: false, 
		                           PUBCOMP: false, SUBACK: false,  UNSUBACK: false, DISCONNECT: false, AUTH: false },

	Receive_Maximum:             {CONNECT: false, CONNACK: false},
	Topic_Alias_Maximum:         {CONNECT: false, CONNACK: false},
	Topic_Alias:                 {PUBLISH: false},
	Maximum_QoS:                 {CONNACK: false},
	Retain_Available:            {CONNACK: false},
	User_Property:               { CONNECT: true, CONNACK: true, PUBLISH: true, PUBACK: true, PUBREC: true, PUBREL: true,
		                           PUBCOMP: true, SUBSCRIBE: true, SUBACK: true, UNSUBSCRIBE: true,	UNSUBACK: true, DISCONNECT: true, AUTH: true },

	Maximum_Packet_Size:                 {CONNECT: false, CONNACK: false},
	Wildcard_Subscription_Available:     {CONNACK: false},
	Subscription_Identifier_Available:   {CONNACK: false},
	Shared_Subscription_Available:       {CONNACK: false},
}

// DupAllowed check if property id allows keys duplication
func DupAllowedProperty(ppid uint32, t byte) bool {
	d, ok := propertyAllowedMessageTypes[ppid]
	if !ok {
		return false
	}

	return d[t]
}

// IsValid check if property id is valid spec value
func IsValidProperty(ppid uint32) bool {
	if _, ok := propertyTypeMap[ppid]; ok {
		return true
	}
	return false
}

// Get Property data type
func GetPropertyType(ppid uint32) (byte, error) {
	if t, ok := propertyTypeMap[ppid]; ok {
		return t, nil
	}
	return 0, errors.New(fmt.Sprintf("Invalid Property ID: 0x%04X", ppid))
}

// IsValidPacketType check either property id can be used for given packet type
func IsValidPacketType4Prop(ppid uint32, t byte) bool {
	mT, ok := propertyAllowedMessageTypes[ppid]
	if !ok {
		return false
	}
	if _, ok = mT[t]; !ok {
		return false
	}
	return true
}

// Set property value
func SetProperty(props map[uint32]interface{}, t byte, ppid uint32 , val interface{}) error {
	if mT, ok := propertyAllowedMessageTypes[id]; !ok {
		return ErrPropertyInvalidID
	} else if _, ok = mT[t]; !ok {
		return ErrPropertyPacketTypeMismatch
	}
	p.properties[id] = val
	return nil
}

func UnpackProps(t byte, r io.Reader) (map[uint32]interface{}, error) {
	props = make(map[uint32]interface{})
	ulen, err := ReadUvarint(r)
	if err != nil {
		return nil, err
	}

	proplen := int(ulen)
	for propLen > 0 {
		ppid, err := ReadUvarint(r)
		if err != nil {
			return nil, err
		}
		if !IsValidPacketType4Prop(ppid, t) {
			return nil, errors.New(fmt.Sprintf("Invalid Property ID: 0x%04X Packet type: 0x%02X", ppid, t))
		}
		propLen -= vlen(ppid)
		val := ReadPropVal(r, ppid)
		props[ppid] = val
	}

	return ppros, 0, nil
}

func ReadPropVal(r io.Reader, ppid uint32) (interface{}, error) {
	pptype, err := GetPropertyType(ppid)
	if err != nil {
		return nil, err
	}
	switch pptype {
	case One_Byte:
		v, err := ReadByte(r)
	case Two_Byte_Integer:
		v, err := ReadUint16(r)
	case Four_Byte_Integer:
		v, err := ReadUint32(r)
	case Variable_Byte_Integer:
		v, err := ReadUvarint(r)
	case UTF8_String:
		v, err := ReadSting(r)
	case UTF8_String_Pair:
		v, err := ReadStingPair(r)
	case Binary_Data:
		v, err := ReadBinaryData(r)
	default:
		v, err := nil, errors.New("Invalid property data type")
	}

	return v, err
}

func (p *property) encode(to []byte) (int, error) {
	pLen := p.FullLen()
	if pLen > len(to) {
		return 0, ErrInsufficientBufferSize
	}

	if pLen == 1 {
		return 1, nil
	}

	var offset int
	var err error
	// Encode variable length header
	total := binary.PutUvarint(to, uint64(p.len))

	for k, v := range p.properties {
		fn := propertyEncodeType[propertyTypeMap[k]]
		offset, err = fn(k, v, to[total:])
		total += offset

		if err != nil {
			break
		}
	}

	return total, err
}

func calcLenByte(id PropertyID, val interface{}) (int, error) {
	l := 0
	calc := func() int {
		return 1 + uvarintCalc(uint32(id))
	}

	switch valueType := val.(type) {
	case uint8:
		l = calc()
	case []uint8:
		for range valueType {
			l += calc()
		}
	default:
		return 0, nil
	}

	return l, nil
}

func calcLenShort(id PropertyID, val interface{}) (int, error) {
	l := 0

	calc := func() int {
		return 2 + uvarintCalc(uint32(id))
	}

	switch valueType := val.(type) {
	case uint16:
		l = calc()
	case []uint16:
		for range valueType {
			l += calc()
		}
	default:
		return 0, nil
	}

	return l, nil
}

func calcLenInt(id PropertyID, val interface{}) (int, error) {
	l := 0

	calc := func() int {
		return 4 + uvarintCalc(uint32(id))
	}

	switch valueType := val.(type) {
	case uint32:
		l = calc()
	case []uint32:
		for range valueType {
			l += calc()
		}
	default:
		return 0, nil
	}

	return l, nil
}

func calcLenVarInt(id PropertyID, val interface{}) (int, error) {
	l := 0

	calc := func(v uint32) int {
		return uvarintCalc(v) + uvarintCalc(uint32(id))
	}

	switch valueType := val.(type) {
	case uint32:
		l = calc(valueType)
	case []uint32:
		for _, v := range valueType {
			l += calc(v)
		}
	default:
		return 0, nil
	}

	return l, nil
}

func calcLenString(id PropertyID, val interface{}) (int, error) {
	l := 0

	calc := func(n int) int {
		return 2 + n + uvarintCalc(uint32(id))
	}

	switch valueType := val.(type) {
	case string:
		l = calc(len(valueType))
	case []string:
		for _, v := range valueType {
			l += calc(len(v))
		}
	default:
		return 0, nil
	}

	return l, nil
}

func calcLenBinary(id PropertyID, val interface{}) (int, error) {
	l := 0

	calc := func(n int) int {
		return 2 + n + uvarintCalc(uint32(id))
	}

	switch valueType := val.(type) {
	case []byte:
		l = calc(len(valueType))
	case [][]string:
		for _, v := range valueType {
			l += calc(len(v))
		}
	default:
		return 0, nil
	}

	return l, nil
}

func calcLenStringPair(id PropertyID, val interface{}) (int, error) {
	l := 0

	calc := func(k, v int) int {
		return 4 + k + v + uvarintCalc(uint32(id))
	}

	switch valueType := val.(type) {
	case StringPair:
		l = calc(len(valueType.K), len(valueType.V))
	case []StringPair:
		for _, v := range valueType {
			l += calc(len(v.K), len(v.V))
		}
	default:
		return 0, nil
	}

	return l, nil
}

func decodeByte(p *property, id PropertyID, from []byte) (int, error) {
	offset := 0
	if len(from[offset:]) < 1 {
		return offset, CodeMalformedPacket
	}

	p.properties[id] = from[offset]
	offset++

	return offset, nil
}

func decodeShort(p *property, id PropertyID, from []byte) (int, error) {
	offset := 0
	if len(from[offset:]) < 2 {
		return offset, CodeMalformedPacket
	}

	v := binary.BigEndian.Uint16(from[offset:])
	offset += 2

	p.properties[id] = v

	return offset, nil
}

func decodeInt(p *property, id PropertyID, from []byte) (int, error) {
	offset := 0
	if len(from[offset:]) < 4 {
		return offset, CodeMalformedPacket
	}

	v := binary.BigEndian.Uint32(from[offset:])
	offset += 4

	p.properties[id] = v

	return offset, nil
}

func decodeVarInt(p *property, id PropertyID, from []byte) (int, error) {
	offset := 0

	v, cnt := uvarint(from[offset:])
	if cnt <= 0 {
		return offset, CodeMalformedPacket
	}
	offset += cnt

	p.properties[id] = v

	return offset, nil
}

func decodeString(p *property, id PropertyID, from []byte) (int, error) {
	offset := 0

	v, n, err := ReadLPBytes(from[offset:])
	if err != nil || !utf8.Valid(v) {
		return offset, CodeMalformedPacket
	}

	offset += n

	p.properties[id] = string(v)

	return offset, nil
}

func decodeStringPair(p *property, id PropertyID, from []byte) (int, error) {
	var k []byte
	var v []byte
	var n int
	var err error

	k, n, err = ReadLPBytes(from)
	offset := n
	if err != nil || !utf8.Valid(k) {
		return offset, CodeMalformedPacket
	}

	v, n, err = ReadLPBytes(from[offset:])
	offset += n

	if err != nil || !utf8.Valid(v) {
		return offset, CodeMalformedPacket
	}

	if _, ok := p.properties[id]; !ok {
		p.properties[id] = []StringPair{}
	}

	p.properties[id] = append(p.properties[id].([]StringPair), StringPair{K: string(k), V: string(v)})

	return offset, nil
}

func decodeBinary(p *property, id PropertyID, from []byte) (int, error) {
	offset := 0

	b, n, err := ReadLPBytes(from[offset:])
	if err != nil {
		return offset, CodeMalformedPacket
	}
	offset += n

	tmp := make([]byte, len(b))

	copy(tmp, b)

	p.properties[id] = tmp

	return offset, nil
}

func encodeByte(id PropertyID, val interface{}, to []byte) (int, error) {
	offset := 0

	encode := func(v uint8, to []byte) int {
		off := writePrefixID(id, to)

		to[off] = v
		off++

		return off
	}

	switch valueType := val.(type) {
	case uint8:
		offset += encode(valueType, to[offset:])
	case []uint8:
		for _, v := range valueType {
			offset += encode(v, to[offset:])
		}
	default:
		panic("unexpected property type")
	}

	return offset, nil
}

func encodeShort(id PropertyID, val interface{}, to []byte) (int, error) {
	offset := 0

	encode := func(v uint16, to []byte) int {
		off := writePrefixID(id, to)
		binary.BigEndian.PutUint16(to[off:], v)
		off += 2

		return off
	}

	switch valueType := val.(type) {
	case uint16:
		offset += encode(valueType, to[offset:])
	case []uint16:
		for _, v := range valueType {
			offset += encode(v, to[offset:])
		}
	default:
		panic("unexpected property type")
	}

	return offset, nil
}

func encodeInt(id PropertyID, val interface{}, to []byte) (int, error) {
	offset := 0

	encode := func(v uint32, to []byte) int {
		off := writePrefixID(id, to)
		binary.BigEndian.PutUint32(to[off:], v)
		off += 4

		return off
	}

	switch valueType := val.(type) {
	case uint32:
		offset += encode(valueType, to[offset:])
	case []uint32:
		for _, v := range valueType {
			offset += encode(v, to[offset:])
		}
	default:
		panic("unexpected property type")
	}

	return offset, nil
}

func encodeVarInt(id PropertyID, val interface{}, to []byte) (int, error) {
	offset := 0

	encode := func(v uint32, to []byte) int {
		off := writePrefixID(id, to)
		off += binary.PutUvarint(to[off:], uint64(v))

		return off
	}

	switch valueType := val.(type) {
	case uint32:
		offset += encode(valueType, to[offset:])
	case []uint32:
		for _, v := range valueType {
			offset += encode(v, to[offset:])
		}
	default:
		panic("unexpected property type")
	}

	return offset, nil
}

func encodeString(id PropertyID, val interface{}, to []byte) (int, error) {
	offset := 0

	encode := func(v string, to []byte) (int, error) {
		off := writePrefixID(id, to)
		count, err := WriteLPBytes(to[off:], []byte(v))
		off += count

		return off, err
	}

	switch valueType := val.(type) {
	case string:
		n, err := encode(valueType, to[offset:])
		offset += n
		if err != nil {
			return offset, err
		}
	case []string:
		for _, v := range valueType {
			n, err := encode(v, to[offset:])
			offset += n
			if err != nil {
				return offset, err
			}
		}
	default:
		panic("unexpected property type")
	}

	return offset, nil
}

func encodeStringPair(id PropertyID, val interface{}, to []byte) (int, error) {
	offset := 0

	encode := func(v StringPair, to []byte) (int, error) {
		off := writePrefixID(id, to)

		n, err := WriteLPBytes(to[off:], []byte(v.K))
		off += n
		if err != nil {
			return off, err
		}

		n, err = WriteLPBytes(to[off:], []byte(v.V))
		off += n
		if err != nil {
			return off, err
		}

		return off, nil
	}

	switch valueType := val.(type) {
	case StringPair:
		n, err := encode(valueType, to[offset:])
		offset += n
		if err != nil {
			return offset, err
		}
	case []StringPair:
		for _, v := range valueType {
			n, err := encode(v, to[offset:])
			offset += n
			if err != nil {
				return offset, err
			}
		}
	default:
		panic("unexpected property type")
	}

	return offset, nil
}

func encodeBinary(id PropertyID, val interface{}, to []byte) (int, error) {
	offset := 0

	encode := func(v []byte, to []byte) (int, error) {
		off := writePrefixID(id, to)
		count, err := WriteLPBytes(to[off:], v)
		off += count

		return off, err
	}

	switch valueType := val.(type) {
	case []byte:
		n, err := encode(valueType, to[offset:])
		offset += n
		if err != nil {
			return offset, err
		}
	case [][]byte:
		for _, v := range valueType {
			n, err := encode(v, to[offset:])
			offset += n
			if err != nil {
				return offset, err
			}
		}
	default:
		panic("unexpected property type")
	}
	return offset, nil
}
