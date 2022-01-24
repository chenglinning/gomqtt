package mqttp

import (
	"fmt"
	"io"
	"bytes"
	"encoding/binary"
	"unicode/utf8"
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
	if ok {
		return d[t]
	}
	return false
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

		pptype, err := GetPropertyType(ppid)
		if err != nil {
			return nil, err
		}
		val, err := ReadPropVal(r, pptype)
		if err != nil {
			return nil, err
		}

		dup := DupAllowedProperty(ppid, t)
		switch pptype {
		case One_Byte:
			props[ppid] = val.(byte)
			proplen -= 1

		case Two_Byte_Integer:
			props[ppid] = val.(uint16)
			proplen -= 2

		case Four_Byte_Integer:
			props[ppid] = val.(uint32)
			proplen -= 4

		case Variable_Byte_Integer:
			if dup {
				if _, ok = props[ppid]; ok {
					props[ppid] = append(props[ppid].([]uint32), val.(uint32))
				} else {
					props[ppid] = []uint32{val.(uint32)}
				}
			} else {
				props[ppid] = val.(uint32)
			}

			proplen -= vlen(val.(uint32))

		case UTF8_String:
			props[ppid] = val.(string)
			proplen -= (2 + len(val.(string)))

		case UTF8_String_Pair:
			if dup {
				if _, ok = props[ppid]; ok {
					props[ppid] = append(props[ppid].([]StringPair), val.(StringPair))
				} else {
					props[ppid] = []StringPair{val.(StringPair)}
				}
			} else {
				props[ppid] = val.(StringPair)
			}

			proplen -= (4 + len(val.(StringPair).k) + len(val.(StringPair).v))

		case Binary_Data:
			props[ppid] = val.([]byte)
			proplen -= (2 + len(val.(byte)))
		}
	}

	return ppros, 0, nil
}

func ReadPropVal(r io.Reader, pptype byte) (interface{}, error) {
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

func PackProps(t byte, props map[uint32]interface{}) ([]byte, error) {
	wbuff := bytes.NewBuffer([]byte{})

	for ppid, val := range props {
		dup := DupAllowedProperty(ppid, t)
		if dup {
			err := WriteDupProp(wbuff, ppid, val)
			if err != nil {
				return nil, err
			}
		} else {
			err := WriteProp(wbuff, ppid, val)
			if err != nil {
				return nil, err
			}
 	    }
	}

	return wbuff.Bytes(), nil
}

func WriteProp(w io.Writer, ppid uint32, v interface{}) error {
	pptype, err := GetPropertyType(ppid)
	if err != nil {
		return err
	}

	// write property ID 
	err = WriteUvarint(w, ppid)
	if err != nil {
		return err
	}

	// write property value
	switch pptype {
	case One_Byte:
		err = WriteByte(w, v.(byte))

	case Two_Byte_Integer:
		err = WriteUint16(w, v.(uint16))

	case Four_Byte_Integer:
		err = WriteUint32(w, v.(uint32))

	case Variable_Byte_Integer:
		err = WriteUvarint(w, v.(uint32))

	case UTF8_String:
		err = WriteString(w, v.(string))

	case UTF8_String_Pair:
		err = WriteStringPair(w, v.(StringPair))

	case Binary_Data:
		err = WriteBinaryData(w, v.([]byte))
	}

	return err
}

func WriteDupProp(w io.Writer, ppid uint32, v interface{}) error {
	pptype, err := GetPropertyType(ppid)
	if err != nil {
		return err
	}

	if ppid != Subscription_Identifier && ppid != User_Property {
		return errors.New(fmt.Sprintf("Not allow duplicate. Property ID: 0x%04X", ppid))
	}

	if ppid == Subscription_Identifier { //  Variable_Byte_Integer
		for i, vi := range v.([]uint32) {
			// write property id
			err = WriteUvarint(w, ppid)
			if err != nil {
				return err
			}
			// write property value
			err = WriteUvarint(w, vi)
			if err != nil {
				return err
			}
		} 
	} else { // User_Property
		for i, vi := range v.([]StringPair) {
			// write property id
			err = WriteUvarint(w, ppid)
			if err != nil {
				return err
			}
			// write property id
			err = WriteStringPair(w, vi)
			if err != nil {
				return err
			}
		} 
	}

	return nil
}