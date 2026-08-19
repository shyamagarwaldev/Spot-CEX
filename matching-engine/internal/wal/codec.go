package wal

import (
	"encoding/json"
	"fmt"

	"github.com/shyamagarwaldev/Spot-CEX/matching-engine/internal/engine"
)

func marshalPayload(cmd engine.Command) ([]byte, error) {
	switch cmd.Type {

	case engine.CreateOrder:
		payload, ok := cmd.Payload.(engine.CreateOrderCommand)
		if !ok {
			return nil, fmt.Errorf("invalid CreateOrder payload")
		}

		return json.Marshal(payload)

	case engine.CancelOrder:
		payload, ok := cmd.Payload.(engine.CancelOrderCommand)
		if !ok {
			return nil, fmt.Errorf("invalid CancelOrder payload")
		}

		return json.Marshal(payload)

	default:
		return nil, fmt.Errorf(
			"command type %v is not WAL supported",
			cmd.Type,
		)
	}
}
