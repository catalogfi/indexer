package database

type Db interface {
	Close()
	Get(string) ([]byte, error)
	GetWithPrefix(string) ([][]byte, error)
	GetMulti([]string) ([][]byte, error)
	Put(string, []byte) error
	PutMulti([]string, [][]byte) error
	Delete(string) error
	DeleteMulti([]string) error
}
