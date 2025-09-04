package store

type Storage interface {
	Set(key, value string) error
	Get(key string) (string, error)
	Delete(key string) error

	GetSnapshot() map[string]string
	ApplySnapshot(data map[string]string)
}
