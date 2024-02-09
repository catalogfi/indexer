package database

type Db interface {
	Close()
	Get(string) ([]byte, error)
	Put(string, []byte) error
	Delete(string) error
}
