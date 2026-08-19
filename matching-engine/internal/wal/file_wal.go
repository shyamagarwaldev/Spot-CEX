package wal

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"github.com/shyamagarwaldev/Spot-CEX/matching-engine/internal/engine"
)

type FileWAL struct {
	mu   sync.Mutex
	file *os.File
}

func NewFileWAL(path string) (*FileWAL, error) {
	file, err := os.OpenFile(
		path,
		os.O_CREATE|os.O_WRONLY|os.O_APPEND,
		0644,
	)
	if err != nil {
		return nil, fmt.Errorf("open wal: %w", err)
	}

	return &FileWAL{
		file: file,
	}, nil
}

type WALRecord struct {
	Sequence uint64             `json:"sequence"`
	Type     engine.CommandType `json:"type"`
	Payload  json.RawMessage    `json:"payload"`
}

func (w *FileWAL) Append(cmd engine.Command) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	payload, err := marshalPayload(cmd)
	if err != nil {
		return err
	}

	record := WALRecord{
		Sequence: cmd.Sequence,
		Type:     cmd.Type,
		Payload:  payload,
	}

	data, err := json.Marshal(record)
	if err != nil {
		return fmt.Errorf("marshal wal record: %w", err)
	}

	// Length-prefix every record so recovery can read
	// records one by one.
	var header [8]byte
	binary.BigEndian.PutUint64(header[:], uint64(len(data)))

	if _, err := w.file.Write(header[:]); err != nil {
		return fmt.Errorf("write wal header: %w", err)
	}

	if _, err := w.file.Write(data); err != nil {
		return fmt.Errorf("write wal record: %w", err)
	}

	// The command is considered durable only after this.
	if err := w.file.Sync(); err != nil {
		return fmt.Errorf("sync wal: %w", err)
	}

	return nil
}

func (w *FileWAL) Close() error {
	return w.file.Close()
}
