package command

import (
	"encoding/json"
)

type Command interface {
	Name() string
	Execute(params json.RawMessage) (interface{}, error)
}
