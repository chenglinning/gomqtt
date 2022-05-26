package mqttp

import (
	"github.com/wonderivan/logger"
	"encoding/binary"
	"regexp"
	"unicode/utf8"
	"bytes"
	"io"
	"fmt"
)

var clientIDRegexp *regexp.Regexp

func init() {
//	clientIDRegexp = regexp.MustCompile(`^[0-9a-zA-Z \-_,.|/]*$`)
	clientIDRegexp = regexp.MustCompile(`^[0-9a-zA-Z\-]*$`)
}

type Connect struct {
	header
	client_id     string
	username      string
	password      string
	keep_alive    uint16
	flags         byte
	will_topic    string
	will_message  []byte
}

var _ Packet = (*Connect)(nil)

func NewConnect() *Connect {
	p := &Connect{}
	p.ResetProps()
	p.ResetWillProps()
	return p
}

func (this *Connect) IsClean() bool {
	return (this.flags & maskConnFlagClean) != 0
}

func (this *Connect) SetClean(v bool) {
	if v {
		this.flags |= maskConnFlagClean
	} else {
		this.flags &= ^maskConnFlagClean
	}
}

func (this *Connect) KeepAlive() uint16 {
	return this.keep_alive
}

func (this *Connect) SetKeepAlive(v uint16) {
	this.keep_alive = v
}

func (this *Connect) ClientID() string {
	return this.client_id
}

func (this *Connect) SetClientID(v string) error {
	if !this.validClientID(v) {
		return ErrInvalid
	}
	this.client_id = v
	return nil
}

// ResetWill reset will state of message
func (this *Connect) ResetWill() {
	this.flags &= ^maskConnFlagWill
	this.flags &= ^maskConnFlagWillQos
	this.flags &= ^maskConnFlagWillRetain
	this.will_topic = ""
	this.will_message = nil
}

// Credentials returns user and password
func (this *Connect) Credentials() (string, string) {
	return this.username, this.password
}

// SetCredentials set username and password
func (this *Connect) SetCredentials(username string, password string) error {
	this.flags &= ^maskConnFlagUsername
	this.flags &= ^maskConnFlagPassword

	// MQTT 3.1.1 does not allow password without user name
	if (len(username) == 0 && len(password) != 0) && this.GetVersion() < MQTTV50 {
		return ErrInvalidArgs
	}

	if len(username) != 0 {
		if !utf8.ValidString(username) {
			return ErrInvalidUtf8
		}
		this.flags |= maskConnFlagUsername
		this.username = username
	}

	if len(password) != 0 {
		this.flags |= maskConnFlagPassword
		this.password = password
	}

	return nil
}

// willFlag returns the bit that specifies whether a Will Message should be stored
// on the server. If the Will Flag is set to 1 this indicates that, if the Accept
// request is accepted, a Will Message MUST be stored on the Server and associated
// with the Network Connection.
func (this *Connect) willFlag() bool {
	return (this.flags & maskConnFlagWill) != 0
}

// willQos returns the two bits that specify the QoS level to be used when publishing
// the Will Message.
func (this *Connect) willQos() byte {
	return (this.flags & maskConnFlagWillQos) >> 3)
}

// willRetain returns the bit specifies if the Will Message is to be Retained when it
// is published.
func (this *Connect) willRetain() bool {
	return (this.flags & maskConnFlagWillRetain) != 0
}

// usernameFlag returns the bit that specifies whether a user name is present in the
// payload.
func (this *Connect) usernameFlag() bool {
	return (this.flags & maskConnFlagUsername) != 0
}

// passwordFlag returns the bit that specifies whether a password is present in the
// payload.
func (this *Connect) passwordFlag() bool {
	return (this.flags & maskConnFlagPassword) != 0
}

func (this *Connect) validClientID(cid string) bool {
	if len(cid) == 0 {
		return true
	}
	return IsValidString(cid) && clientIDRegexp.MatchString(cid)
}

func (this *Connect) Unpack(rdata []byte) error {
	r := bytes.NewBuffer(rdata)
	// protocol name "MQTT"	
	utf8bytes, err := ReadUTF8String(r)
	if err != nil {
		logger.Error(fmt.Sprintf("Error parsing protocol name: %s", err))
		return CodeUnspecifiedError
	}
	if string(utf8bytes) != "MQTT" {
		logger.Error(fmt.Sprintf("Invalid protocol name: 0x%s", utf8bytes))
		return CodeUnsupportedProtocol
	}
	// protocol level (4 || 5)
	proto_version, err := ReadByte(r)
	if err != nil {
		logger.Error("Error parsing protocol version")
		return CodeMalformedPacket
	}
	if proto_version == MQTT311 || proto_version == MQTT50 {
		this.SetVersion(proto_version)
	} else {
		logger.Error(fmt.Sprintf("Invalid protocol version: 0x%02X", proto_version))
		return CodeUnsupportedProtocol
	}

	// connect flags
	flags, err := ReadByte(r)
	if err != nil {
		logger.Error(fmt.Sprintf("Error parsing connect flags: %s", err))
		return CodeUnspecifiedError
	}
	if flags & maskConnFlagReserved > 0 {
		logger.Error(fmt.Sprintf("Invalid connect flags: 0x%02X", flags))
		return CodeMalformedPacket
	}
	this.flags = flags

	// Verify the validity of the flags
    if this.willFlag() {
		if this.willQos() > QoS2 {
			logger.Error(fmt.Sprintf("Invalid will qos: 0x%02X", this.willQos()))
			return CodeMalformedPacket
		}
	} else if this.willQos > QoS0 {
		logger.Error(fmt.Sprintf("Invalid will qos: 0x%02X", this.willQos()))
		return CodeMalformedPacket
	}

	// keep alive interval
	keep_alive, err := ReadUint16(r)
	if err != nil {
		logger.Error(fmt.Sprintf("Error parsing keep alive interval: %s", err))
		return CodeUnspecifiedError
	}
	this.SetKeepAlive(keep_alive)
    
	if proto_version == MQTT311 {
		if !this.usernameFlag() && this.passwordFlag() {
			logger.Error(fmt.Sprintf("Invalid connect flags: 0x%02X", this.flags))
			return CodeMalformedPacket
		}
	} else { // MQTT 5.0
		// reading properties
		err = this.ReadProps(r io.Reader)
		if err != nil {
			logger.Error(fmt.Sprintf("Failed reading properties: %s", err))
			return CodeMalformedPacket
		}
	}

	// MQTT3.1.1 5.0  reading client id
	client_id, err := ReadUTF8String(r)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed reading client id: %s", err))
		return CodeMalformedPacket
	}
	if !clientIDRegexp.MatchString(client_id) || len(client_id) == 0 {
		logger.Error(fmt.Sprintf("Invalic client id: %s", client_id))
		return CodeInvalidClientID
	}
	this. client_id = client_id

	// MQTT 5.0 reading will properties
	if this.willFlag() && proto_version == MQTT50 {
		err = this.ReadWillProps(r io.Reader)
		if err != nil {
			logger.Error(fmt.Sprintf("Failed reading will properties: %s", err))
			return CodeMalformedPacket
		}
	}

	// MQTT 3.1.1, 5.0
	if this.willFlag() {
		// reading will topic
		will_topic, err := ReadUTF8String(r)
		if err != nil {
			logger.Error(fmt.Sprintf("Failed reading will topic: %s", err))
			return CodeMalformedPacket
		}

		// verify will topic
		if !TopicPublishRegexp.MatchString(will_topic) {
			logger.Error(fmt.Sprintf("Invalic will topic: %s", will_topic))
			return CodeMalformedPacket
		}
		this.will_topic = will_topic

		// reading will paload
		payload, err : = ReadBinaryData(r)
		if err != nil {
			logger.Error(fmt.Sprintf("Failed reading will payload: %s", err))
			return CodeMalformedPacket
		}		
		this.will_message = payload
	}

	// reading username
	if this.usernameFlag() {
		uname, err := ReadUTF8String(r)
		if err != nil {
			logger.Error(fmt.Sprintf("Failed reading username: %s", err))
			return CodeMalformedPacket
		}
		this.username = uname
	}

	// reading password
	if this.passwordFlag() {
		password, err := ReadUTF8String(r)
		if err != nil {
			logger.Error(fmt.Sprintf("Failed reading password: %s", err))
			return CodeMalformedPacket
		}
		this.password = password
	}

	// authenticate usrname / password
	// DTB
}

func (this *Connect) Pack() ([]byte, error) {
	// 2 (4==len("MQTT") ) + 4 ("MQTT")  + 1 (proto level) + 1 (connect flags) +2 (keep alive interval) 
	buff := bytes.NewBuffer([]byte{})
	
	// Variable Header:
	// protocol name "MQTT"
	err = WriteString(buff, PRONAME)
	if err != nil {
		logger.Error(fmt.Sprintf("Error encoding protocol name: %s", err))
		return nil, err
	}

	// protocol level
	err = WriteByte(buff, this.GetVersion())
	if err != nil {
		logger.Error(fmt.Sprintf("Error encoding protoco level: %s", err))
		return nil, err
	}
	// connect flags
	err = WriteByte(buff, this.flags)
	if err != nil {
		logger.Error(fmt.Sprintf("Error encoding connect flags: %s", err))
		return nil, err
	}

	// keep alive interval
	err = WriteUint16(buff, this.keep_alive)
	if err != nil {
		logger.Error(fmt.Sprintf("Error encoding connect flags: %s", err))
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

	// payload:
	//   clientid
	err = WriteString(buff, this.client_id)
	if err != nil {
		logger.Error(fmt.Sprintf("Error encoding clientid: %s", err))
		return nil, err
	}
	if this.willFlag() {
		// will property 
		if this.GetVersion()==MQTT50 {
			wpbytes := this.willpropset.PackProps(this.GetType())
			wplen := len(willppbytes)
			// encoding will propertiy len
			err = WriteUvarint(buff, uint32(wplen))
			if err != nil {
				logger.Error(fmt.Sprintf("Error encoding will propertiy length: %s", err))
				return nil, err
			}
			if wplen>0 {
				_, err = buff.Write(wpbytes)
				if err != nil {
					logger.Error(fmt.Sprintf("Error encoding will property: %s", err))
					return nil, err
				}
			}
		}

		// will topic
		err = WriteString(buff, this.will_topic)
		if err != nil {
			logger.Error(fmt.Sprintf("Error encoding will topic: %s", err))
			return nil, err
		}

		// will payload
		err = WriteBinaryData(buff, this.will_message)
		if err != nil {
			logger.Error(fmt.Sprintf("Error encoding will message: %s", err))
			return nil, err
		}
	}
	// user name
	if this.usernameFlag() {
		err = WriteString(buff, this.username)
		if err != nil {
			logger.Error(fmt.Sprintf("Error encoding username: %s", err))
			return nil, err
		}
	}
	// user password
	if this.passwordFlag() {
		err = WriteString(buff, this.password)
		if err != nil {
			logger.Error(fmt.Sprintf("Error encoding password: %s", err))
			return nil, err
		}
	}

	return buff.Bytes(), nil

}
