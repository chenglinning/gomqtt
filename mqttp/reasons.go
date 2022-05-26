package mqttp

// ReasonCode contains return codes across all MQTT specs
type ReasonCode byte

// nolint: golint                                            // V3.1.1  \  V5.0
const ( // /////////////////////////////////////////////////////   |    \    |
	CodeSuccess                            ReasonCode = 0x00 //    |    \    |
	CodeRefusedUnacceptableProtocolVersion ReasonCode = 0x01 //    |    \    |
	CodeRefusedIdentifierRejected          ReasonCode = 0x02 //    |    \    |
	CodeRefusedServerUnavailable           ReasonCode = 0x03 //    |    \    |
	CodeRefusedBadUsernameOrPassword       ReasonCode = 0x04 //    |    \    |
	CodeRefusedNotAuthorized               ReasonCode = 0x05 // <--|    \    | // V3.1.1 ONLY
	CodeNoMatchingSubscribers              ReasonCode = 0x10 //         \    |
	CodeNoSubscriptionExisted              ReasonCode = 0x11 //         \    |
	CodeContinueAuthentication             ReasonCode = 0x18 //         \    |
	CodeReAuthenticate                     ReasonCode = 0x19 //         \    |
	CodeUnspecifiedError                   ReasonCode = 0x80 //         \    |
	CodeMalformedPacket                    ReasonCode = 0x81 //         \    |
	CodeProtocolError                      ReasonCode = 0x82 //         \    |
	CodeImplementationSpecificError        ReasonCode = 0x83 //         \    |
	CodeUnsupportedProtocol                ReasonCode = 0x84 //         \    |
	CodeInvalidClientID                    ReasonCode = 0x85 //         \    |
	CodeBadUserOrPassword                  ReasonCode = 0x86 //         \    |
	CodeNotAuthorized                      ReasonCode = 0x87 //         \    |
	CodeServerUnavailable                  ReasonCode = 0x88 //         \    |
	CodeServerBusy                         ReasonCode = 0x89 //         \    |
	CodeBanned                             ReasonCode = 0x8A //         \    |
	CodeServerShuttingDown                 ReasonCode = 0x8B //         \    |
	CodeBadAuthMethod                      ReasonCode = 0x8C //         \    |
	CodeKeepAliveTimeout                   ReasonCode = 0x8D //         \    |
	CodeSessionTakenOver                   ReasonCode = 0x8E //         \    |
	CodeInvalidTopicFilter                 ReasonCode = 0x8F //         \    |
	CodeInvalidTopicName                   ReasonCode = 0x90 //         \    |
	CodePacketIDInUse                      ReasonCode = 0x91 //         \    |
	CodePacketIDNotFound                   ReasonCode = 0x92 //         \    |
	CodeReceiveMaximumExceeded             ReasonCode = 0x93 //         \    |
	CodeInvalidTopicAlias                  ReasonCode = 0x94 //         \    |
	CodePacketTooLarge                     ReasonCode = 0x95 //         \    |
	CodeMessageRateTooHigh                 ReasonCode = 0x96 //         \    |
	CodeQuotaExceeded                      ReasonCode = 0x97 //         \    |
	CodeAdministrativeAction               ReasonCode = 0x98 //         \    |
	CodeInvalidPayloadFormat               ReasonCode = 0x99 //         \    |
	CodeRetainNotSupported                 ReasonCode = 0x9A //         \    |
	CodeNotSupportedQoS                    ReasonCode = 0x9B //         \    |
	CodeUseAnotherServer                   ReasonCode = 0x9C //         \    |
	CodeServerMoved                        ReasonCode = 0x9D //         \    |
	CodeSharedSubscriptionNotSupported     ReasonCode = 0x9E //         \    |
	CodeConnectionRateExceeded             ReasonCode = 0x9F //         \    |
	CodeMaximumConnectTime                 ReasonCode = 0xA0 //         \    |
	CodeSubscriptionIDNotSupported         ReasonCode = 0xA1 //         \    |
	CodeWildcardSubscriptionsNotSupported  ReasonCode = 0xA2 //         \ <--|
)

var packetTypeCodeMap = map[PKType]map[ReasonCode] bool {
{
	CONNACK: {
		CodeSuccess:                            true,
		CodeRefusedUnacceptableProtocolVersion: true,
		CodeRefusedIdentifierRejected:          true,
		CodeRefusedServerUnavailable:           true,
		CodeRefusedBadUsernameOrPassword:       true,
		CodeRefusedNotAuthorized:               true,
		CodeUnspecifiedError:                   true,
		CodeMalformedPacket:                    true,
		CodeImplementationSpecificError:        true,
		CodeUnsupportedProtocol:                true,
		CodeInvalidClientID:                    true,
		CodeBadUserOrPassword:                  true,
		CodeNotAuthorized:                      true,
		CodeServerUnavailable:                  true,
		CodeServerBusy:                         true,
		CodeBanned:                             true,
		CodeBadAuthMethod:                      true,
		CodeInvalidTopicName:                   true,
		CodePacketTooLarge:                     true,
		CodeQuotaExceeded:                      true,
		CodeRetainNotSupported:                 true,
		CodeNotSupportedQoS:                    true,
		CodeUseAnotherServer:                   true,
		CodeServerMoved:                        true,
		CodeConnectionRateExceeded:             true,
	},

	PUBACK: {
		CodeSuccess:                     true,
		CodeNoMatchingSubscribers:       true,
		CodeUnspecifiedError:            true,
		CodeImplementationSpecificError: true,
		CodeNotAuthorized:               true,
		CodeInvalidTopicName:            true,
		CodeQuotaExceeded:               true,
		CodeInvalidPayloadFormat:        true,
	},

	PUBREC: {
		CodeSuccess:                     true,
		CodeNoMatchingSubscribers:       true,
		CodeUnspecifiedError:            true,
		CodeImplementationSpecificError: true,
		CodeNotAuthorized:               true,
		CodeInvalidTopicName:            true,
		CodePacketIDInUse:               true,
		CodeQuotaExceeded:               true,
		CodeInvalidPayloadFormat:        true,
	},

	PUBREL: {
		CodeSuccess:          true,
		CodePacketIDNotFound: true,
	},
	
	PUBCOMP: {
		CodeSuccess:          true,
		CodePacketIDNotFound: true,
	},

	SUBACK: {
		QoS0:                                  true,  // QoS 0
		QoS1:                                  true,  // QoS 1
		QoS2:                                  true,  // QoS 2
		CodeUnspecifiedError:                  true,
		CodeImplementationSpecificError:       true,
		CodeNotAuthorized:                     true,
		CodeInvalidTopicFilter:                true,
		CodePacketIDInUse:                     true,
		CodeQuotaExceeded:                     true,
		CodeSharedSubscriptionNotSupported:    true,
		CodeSubscriptionIDNotSupported:        true,
		CodeWildcardSubscriptionsNotSupported: true,
	},

	UNSUBACK: {
		CodeSuccess:                     true,
		CodeNoSubscriptionExisted:       true,
		CodeUnspecifiedError:            true,
		CodeImplementationSpecificError: true,
		CodeNotAuthorized:               true,
		CodeInvalidTopicFilter:          true,
		CodePacketIDInUse:               true,
	},

	DISCONNECT: {
		CodeSuccess:                           true,
		CodeRefusedBadUsernameOrPassword:      true,
		CodeUnspecifiedError:                  true,
		CodeMalformedPacket:                   true,
		CodeProtocolError:                     true,
		CodeImplementationSpecificError:       true,
		CodeNotAuthorized:                     true,
		CodeServerBusy:                        true,
		CodeServerShuttingDown:                true,
		CodeKeepAliveTimeout:                  true,
		CodeSessionTakenOver:                  true,
		CodeInvalidTopicFilter:                true,
		CodeInvalidTopicName:                  true,
		CodePacketTooLarge:                    true,
		CodeReceiveMaximumExceeded:            true,
		CodeInvalidTopicAlias:                 true,
		CodeMessageRateTooHigh:                true,
		CodeQuotaExceeded:                     true,
		CodeAdministrativeAction:              true,
		CodeInvalidPayloadFormat:              true,
		CodeRetainNotSupported:                true,
		CodeNotSupportedQoS:                   true,
		CodeUseAnotherServer:                  true,
		CodeServerMoved:                       true,
		CodeSharedSubscriptionNotSupported:    true,
		CodeConnectionRateExceeded:            true,
		CodeMaximumConnectTime:                true,
		CodeSubscriptionIDNotSupported:        true,
		CodeWildcardSubscriptionsNotSupported: true,
	},
	AUTH: {
		CodeSuccess: true,
		CodeContinueAuthentication: true,
		CodeReAuthenticate: true,
	},
}

var codeDescMap = map[ReasonCode]string{
	CodeSuccess:                            "Operation success",
	CodeRefusedUnacceptableProtocolVersion: "The Server does not support the level of the MQTT protocol requested by the Client",
	CodeRefusedIdentifierRejected:          "The Client identifier is not allowed",
	CodeRefusedServerUnavailable:           "Server refused connection",
	CodeRefusedBadUsernameOrPassword:       "The data in the user name or password is malformed",
	CodeRefusedNotAuthorized:               "The Client is not authorized to connect",
	CodeNoMatchingSubscribers:              "The message is accepted but there are no subscribers",
	CodeNoSubscriptionExisted:              "No matching subscription existed",
	CodeContinueAuthentication:             "Continue the authentication with another step",
	CodeReAuthenticate:                     "Initiate a re-authentication",
	CodeUnspecifiedError:                   "Return code not specified by application",
	CodeMalformedPacket:                    "Malformed Packet",
	CodeProtocolError:                      "Protocol Error",
	CodeImplementationSpecificError:        "Implementation specific error",
	CodeUnsupportedProtocol:                "Unsupported Protocol Version",
	CodeInvalidClientID:                    "Client Identifier not valid",
	CodeBadUserOrPassword:                  "Bad User Name or Password",
	CodeNotAuthorized:                      "Not authorized",
	CodeServerUnavailable:                  "Server unavailable",
	CodeServerBusy:                         "Server busy",
	CodeBanned:                             "Banned",
	CodeServerShuttingDown:                 "Server shutting down",
	CodeBadAuthMethod:                      "Bad authentication method",
	CodeKeepAliveTimeout:                   "Keep Alive timeout",
	CodeSessionTakenOver:                   "Session taken over",
	CodeInvalidTopicFilter:                 "Topic Filter invalid",
	CodeInvalidTopicName:                   "Topic Name invalid",
	CodePacketIDInUse:                      "Packet Identifier in use",
	CodePacketIDNotFound:                   "Packet Identifier not found",
	CodeReceiveMaximumExceeded:             "Receive Maximum exceeded",
	CodeInvalidTopicAlias:                  "Topic Alias invalid",
	CodePacketTooLarge:                     "Packet too large",
	CodeMessageRateTooHigh:                 "Message rate too high",
	CodeQuotaExceeded:                      "Quota exceeded",
	CodeAdministrativeAction:               "Administrative action",
	CodeInvalidPayloadFormat:               "Payload format invalid",
	CodeRetainNotSupported:                 "Retain not supported",
	CodeNotSupportedQoS:                    "QoS not supported",
	CodeUseAnotherServer:                   "Use another server",
	CodeServerMoved:                        "Server moved",
	CodeSharedSubscriptionNotSupported:     "Shared Subscriptions not supported",
	CodeConnectionRateExceeded:             "Connection rate exceeded",
	CodeMaximumConnectTime:                 "Maximum connect time",
	CodeSubscriptionIDNotSupported:         "Subscription Identifiers not supported",
	CodeWildcardSubscriptionsNotSupported:  "Wildcard Subscriptions not supported",
}

// Value convert reason code to byte type
func (c ReasonCode) Value() byte {
	return byte(c)
}

// IsValid check either reason code is valid across all MQTT specs
func (c ReasonCode) IsValid() bool {
	if _, ok := codeDescMap[c]; ok {
		return true
	}
	return false
}

// IsValidV3 check either reason code is valid for MQTT V3.1/V3.1.1 or not
func (c ReasonCode) IsValidV3() bool {
	return c <= CodeRefusedNotAuthorized
}

// IsValidV5 check either reason code is valid for MQTT V5.0 or not
func (c ReasonCode) IsValidV5() bool {
	return ( c == CodeSuccess || (c >= CodeNoMatchingSubscribers && c <= CodeWildcardSubscriptionsNotSupported) )
}

// IsValidForType check either reason code is valid for giver packet type
func (c ReasonCode) IsValidForType(t PKType) bool {
	pT, ok := packetTypeCodeMap[t]
	if ok {
		if _, ok = pT[c]; ok {
			return true
		}
	}
	return false
}

// Error returns the description of the ReturnCode
func (c ReasonCode) Error() string {
	if s, ok := codeDescMap[c]; ok {
		return s
	}
	return "Unknown error"
}


// Desc return code description
func (c ReasonCode) Desc() string {
	return c.Error()
}
