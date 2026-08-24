package wal

import "github.com/shyamagarwaldev/Spot-CEX/matching-engine/internal/engine"

type IWAL interface {
	Append(engine.Command) error
}
