package message

import (
	"errors"
	
	"github.com/pokt-network/pocket-core/message/fbs"
)

func UnmarshalMessage(flatBuffer []byte) Message {
	message := fbs.GetRootAsMessage(flatBuffer, 0)
	return Message{message.Type(), message.PayloadBytes(), message.Timestamp()}
}

func RouteMessageByPayload(m Message) (interface{}, error) {
	switch m.Type_ {
	case fbs.MessageTypeDISC_HELLO:
		return UnmarshalHelloMessage(m.Payload), nil
	default:
		return nil, errors.New("unsupported message type" + string(m.Type_))
	}
}

func UnmarshalHelloMessage(flatBuffer []byte) HelloMessage {
	helloMessage := fbs.GetRootAsHelloMessage(flatBuffer, 0)
	return HelloMessage{string(helloMessage.GidBytes())}
}
