package snapshort

type Snapshotter interface {
	Save(sequence uint64) error
	Load()
}
