package rpc

import (
	"net/http"

	"github.com/catalogfi/indexer/command"
	"github.com/gin-gonic/gin"
)

type RPC interface {
	AddCommand(cmd command.Command)
	HandleJSONRPC(ctx *gin.Context)
}

type rpc struct {
	storage  command.Storage
	commands map[string]command.Command
}

type Request struct {
	JSONRPC string        `json:"jsonrpc"`
	ID      string        `json:"id"`
	Method  string        `json:"method"`
	Params  []interface{} `json:"params"`
}

type Response struct {
	Result interface{} `json:"result"`
	Error  interface{} `json:"error"`
	ID     string      `json:"id"`
}

type ErrResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func New(storage command.Storage) RPC {
	return &rpc{
		storage:  storage,
		commands: make(map[string]command.Command),
	}
}

func (r *rpc) AddCommand(cmd command.Command) {
	r.commands[cmd.Name()] = cmd
}

func (r *rpc) HandleJSONRPC(ctx *gin.Context) {
	req := Request{}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, Response{Result: nil, Error: ErrResponse{-1, err.Error()}, ID: req.ID})
		return
	}

	resp, err := r.commands[req.Method].Query(r.storage, req.Params)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, Response{Result: nil, Error: ErrResponse{-1, err.Error()}, ID: req.ID})
		return
	}
	ctx.JSON(http.StatusOK, Response{Result: resp, Error: nil, ID: req.ID})
}

func Default(str command.Storage) RPC {
	rpc := New(str)
	rpc.AddCommand(command.GetBlock())
	rpc.AddCommand(command.GetBlockCount())
	rpc.AddCommand(command.GetBlockHash())
	rpc.AddCommand(command.GetBlockHeader())
	rpc.AddCommand(command.GetRawTransaction())
	rpc.AddCommand(command.ListUnspent())
	return rpc
}
