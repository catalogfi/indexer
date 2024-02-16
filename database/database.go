package database

type Db interface {
	Close()
	Get(string) ([]byte, error)
	Put(string, []byte) error
	PutMulti([]string, [][]byte) error
	Delete(string) error
}
