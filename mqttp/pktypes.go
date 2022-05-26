package mqttp

type PKType byte

const PRONAME string = "MQTT"
const (
	TBD		byte = 0
	MQTT311 byte = 4
	MQTT50  byte = 5
)
const (
	QoS0 byte = 0
	QoS1 byte = 1
	QoS2 byte = 2
)

const (
	RESERVED PKType = iota
	CONNECT
	CONNACK
	PUBLISH
	PUBACK
	PUBREC
	PUBREL
	PUBCOMP
	SUBSCRIBE
	SUBACK
	UNSUBSCRIBE
	UNSUBACK
	PINGREQ
	PINGRESP
	DISCONNECT
	AUTH
)

var (
	// TopicFilterRegexp regular expression that all subscriptions must be validated
	TopicFilterRegexp = regexp.MustCompile(`^(([^+#]*|\+)(/([^+#]*|\+))*(/#)?|#)$`)

	// TopicPublishRegexp regular expression that all publish to topic must be validated
	TopicPublishRegexp = regexp.MustCompile(`^[^#+]*$`)

	// SharedTopicRegexp regular expression that all share subscription must be validated
	SharedTopicRegexp = regexp.MustCompile(`^\$share/([^#+/]+)(/)(.+)$`)

	// BasicUTFRegexp regular expression all MQTT strings must meet [MQTT-1.5.3]
	BasicUTFRegexp = regexp.MustCompile("^[^\u0000-\u001F\u007F-\u009F]*$")
)

var dollarPrefix = []byte("$")
var sharePrefix = []byte("$share")
var topicSep = []byte("/")


// PacketNames maps the constants for each of the MQTT packet types
// to a string representation of their name.
var PacketNames = map[PKType]string{
	0:  "RESERVED",
	1:  "CONNECT",
	2:  "CONNACK",
	3:  "PUBLISH",
	4:  "PUBACK",
	5:  "PUBREC",
	6:  "PUBREL",
	7:  "PUBCOMP",
	8:  "SUBSCRIBE",
	9:  "SUBACK",
	10: "UNSUBSCRIBE",
	11: "UNSUBACK",
	12: "PINGREQ",
	13: "PINGRESP",
	14: "DISCONNECT",
	15: "AUTH",
}


var typeDefaultFlags = [AUTH + 1]byte {
	0, // RESERVED
	0, // CONNECT
	0, // CONNACK
	0, // PUBLISH
	0, // PUBACK
	0, // PUBREC
	2, // PUBREL
	0, // PUBCOMP
	2, // SUBSCRIBE
	0, // SUBACK
	2, // UNSUBSCRIBE
	0, // UNSUBACK
	0, // PINGREQ
	0, // PINGRESP
	0, // DISCONNECT
	0, // AUTH
}

func (t PKType) Name() string {
	if t > AUTH {
		return "UNKNOWN"
	}
	return PacketNames[t]
}

func (t PKType) ToByte() byte {
	return byte(t)
}

// DefaultFlags returns the default flag values for the message type, as defined by the MQTT spec.
func (t PKType) DefaultFlags() byte {
	if t > AUTH {
		return 0
	}
	return typeDefaultFlags[t]
}

func IsValidUTF(b []byte) bool {
	return utf8.Valid(b) && BasicUTFRegexp.Match(b)
}

func IsValidString(s string) bool {
	return utf8.ValidString(s) && BasicUTFRegexp.MatchString(s)
}

func IsValidTopic(s string) bool {
	return utf8.ValidString(s) && BasicUTFRegexp.MatchString(s) && TopicPublishRegexp.MatchString(s)
}

type SubOptions byte

// QoS quality of service
func (s SubOps) QoS() byte {
	return byte(s) & maskSubscriptionQoS
}

// Raw just return byte
func (s SubOps) Raw() byte {
	return byte(s)
}

// NL No Local option
// V5.0 ONLY
func (s SubOps) NL() bool {
	return (byte(s) & maskSubscriptionNL) != 0
}

// V5.0 ONLY
func (s SubOps) RAP() bool {
	return (byte(s) & maskSubscriptionRAP) != 0
}

// V5.0 ONLY
func (s SubOps) RetainHandling() byte {
	return (byte(s) & maskSubscriptionRetainHandling) >> 4
}
