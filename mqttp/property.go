package mqttp

import (
	"github.com/wonderivan/logger"
	"fmt"
	"io"
	"bytes"
	"encoding/binary"
	"unicode/utf8"
)

// PropertyID id as per [MQTT-2.2.2]
type PropertyID uint32
type PropertyValue interface{}
type PropertyMap map[PropertyID] PropertyValue

type PropertySet struct {
	props propertyTypeMap
} 

// PropertyError encodes property error
type PropertyError int

const (
	One_Byte               byte = iota
	Two_Byte_Integer
	Four_Byte_Integer
	Variable_Byte_Integer
	UTF8_String
	UTF8_String_Pair
	Binary_Data
)

const (
	Payload_Format_Indicator        		= PropertyID (0x01)
	Message_Expiry_Interval         		= PropertyID (0x02)
	Content_Type                    		= PropertyID (0x03)
	Response_Topic                  		= PropertyID (0x08)
	Correlation_Data                		= PropertyID (0x09)
	Subscription_Identifier         		= PropertyID (0x0B)
	Session_Expiry_Interval		            = PropertyID (0x11)
	Assigned_Client_Identifier              = PropertyID (0x12)
	Server_Keep_Alive                       = PropertyID (0x13)
	Authentication_Method                   = PropertyID (0x15)
	Authentication_Data                     = PropertyID (0x16)
	Request_Problem_Information             = PropertyID (0x17)
	Will_Delay_Interval                     = PropertyID (0x18)
	Request_Response_Information            = PropertyID (0x19)
	Response_Information                    = PropertyID (0x1A)
	Server_Reference                        = PropertyID (0x1C)
	Reason_String                           = PropertyID (0x1F)
	Receive_Maximum                         = PropertyID (0x21)
	Topic_Alias_Maximum                     = PropertyID (0x22)
	Topic_Alias                             = PropertyID (0x23)
	Maximum_QoS                             = PropertyID (0x24)
	Retain_Available                        = PropertyID (0x25)
	User_Property                           = PropertyID (0x26)
	Maximum_Packet_Size                     = PropertyID (0x27)
	Wildcard_Subscription_Available         = PropertyID (0x28)
	Subscription_Identifier_Available       = PropertyID (0x29)
	Shared_Subscription_Available           = PropertyID (0x2A)
)

var propertyTypeMap = map[PropertyID] byte {
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
var propertyAllowedMessageTypes = map[PropertyID]map[byte]bool{
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
func MultiAllowedProperty(ppid PropertyID, t byte) bool {
	d, ok := propertyAllowedMessageTypes[ppid]
	if ok {
		return d[t]
	}
	return false
}

// IsValid check if property id is valid spec value
func IsValidProperty(ppid PropertyID) bool {
	if _, ok := propertyTypeMap[ppid]; ok {
		return true
	}
	return false
}

// Get Property data type
func GetPropertyType(ppid PropertyID) (byte, error) {
	if t, ok := propertyTypeMap[ppid]; ok {
		return t, nil
	}
	return 0, errors.New(fmt.Sprintf("Invalid Property ID: 0x%04X", ppid))
}

// Get Property data lenght
func GetPropertyLength(pptype byte, val PropertyValue) int {
	var pplen int
	switch pptype {
	case One_Byte:
		pplen = 1
	case Two_Byte_Integer:
		pplen = 2
	case Four_Byte_Integer:
		pplen = 4
	case Variable_Byte_Integer:
		pplen -= vlen(val.(uint32))
	case UTF8_String:
		pplen = (2 + len(val.(string)))
	case UTF8_String_Pair:
		pplen = (4 + len(val.(StringPair).k) + len(val.(StringPair).v))
	case Binary_Data:
		pplen -= (2 + len(val.([]byte)))
    }
	return pplen
}

// IsValidPacketType check either property id can be used for given packet type
func IsValidPacketType4Prop(ppid PropertyID, t byte) bool {
	mT, ok := propertyAllowedMessageTypes[ppid]
	if !ok {
		return false
	}

	if _, ok = mT[t]; ok {
		return true
	}

	return false
}

// reset PropertySet
func (this *PropertySet) Reset() {
	this.props = make(PropertyMap)
}

// Set property value
func (this *PropertySet) SetProperty(t PKType, id PropertyID , val PropertyValue) error {
	if mT, ok := propertyAllowedMessageTypes[id]; !ok {
		return ErrPropertyInvalidID
	} else if _, ok = mT[t]; !ok {
		return CodeProtocolError
	}
	dup := MultiAllowedProperty(id, t)
	if dup {
		if _, ok = this.props[id]; ok {
			this.props[id] = append(this.props[id], val)
		} else {
			this.props[id] = make([]PropertyValue,0)
		}
	} else {
		if _, ok = this.props[id]; ok {
			return CodeProtocolError
		}
		this.props[ppid] = val
	}

	return nil
}

// Get property value
func (this *PropertySet) GetProperty(id PropertyID) PropertyValue {
	if v, ok := this.props[id]; !ok {
		return nil
	} 
	return v
}

func (this *PropertySet) UnpackProps(r io.Reader, t PKType) error {
	this.Reset()

	ulen, err := ReadUvarint(r)
	if err != nil {
		logger.Error("Error parsing properties lenght")
		return nil, ErrMalformedStream
	}

	proplen := int(ulen)
	for propLen > 0 {
		ppid, err := ReadUvarint(r)
		if err != nil {
			logger.Error("Error parsing property ID")
			return nil, err
		}
		if !IsValidPacketType4Prop(ppid, t) {
			logger.Error(fmt.Sprintf("Invalid PropertyID: 0x%04X Packet type: 0x%02X", ppid, t))
			return nil, ErrMalformedStream
		}

		propLen -= vlen(ppid)

		pptype, err := GetPropertyType(ppid)
		if err != nil {
			logger.Error(fmt.Sprintf("Invalid Property pptype of PropertyID : 0x%04X Packet type: 0x%02X", ppid, t))
			return nil, err
		}

		val, err := ReadPropVal(r, pptype)

		if err != nil {
			logger.Error(fmt.Sprintf("Error read property value (pptye: 0x%02x)", pptype))
			return nil, err
		}

		err = this.SetProperty(t, ppid, val)
		if err {
			logger.Error(err.Error())
			return err
		}
		proplen -= GetPropertyLength(pptype, val)
	}

	return nil
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

func (this *PropertySet) PackProps(t PKType) []byte {
	wbuff := bytes.NewBuffer([]byte{})

	for id, val := range this.props {
		dup := MultiAllowedProperty(id, t)
		if dup {
			err := WriteMultiProp(wbuff, id, val)
			if err != nil {
				return nil
			}
		} else {
			err := WriteProp(wbuff, id, val)
			if err != nil {
				return nil
			}
 	    }
	}

	return wbuff.Bytes()
}

func WriteProp(w io.Writer, id PropertyID, v PropertyValue) error {
	pptype, err := GetPropertyType(id)
	if err != nil {
		return err
	}

	// write property ID 
	err = WriteUvarint(w, id)
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

func WriteMultiProp(w io.Writer, id PropertyID, v PropertyValue) error {
	var err error
	if id == Subscription_Identifier { // Variable_Byte_Integer
		for i, val := range v.([]uint32) {
			// write property ID
			err = WriteUvarint(w, id)
			if err != nil {
				return err
			}
			// write property value
			err = WriteUvarint(w, val)
			if err != nil {
				return err
			}
		} 
	} else if id == User_Property { // UTF8_String_pair
		for i, val := range v.([]StringPair) {
			// write property id
			err = WriteUvarint(w, id)
			if err != nil {
				return err
			}
			// write property value
			err = WriteStringPair(w, val)
			if err != nil {
				return err
			}
		} 
	} else {
		err = errors.New(fmt.Sprintf("Not allow duplicate. Property ID: 0x%04X", id))
		logger.Error(err.Error())
		return err
	}

	return nil
}