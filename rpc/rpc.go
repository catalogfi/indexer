package rpc

import (
	"encoding/json"
	"net/http"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/catalogfi/indexer/command"
	"github.com/catalogfi/indexer/store"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type RPC interface {
	AddCommand(cmd command.Command)
	HandleJSONRPC(ctx *gin.Context)
	SetLogger(logger *zap.Logger) RPC
	Run(port string) error
}

type rpc struct {
	store    *store.Storage
	commands map[string]command.Command
	logger   *zap.Logger
}

type Request struct {
	Version string      `json:"version"`
	ID      string      `json:"id"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params"`
}

type Response struct {
	Version string       `json:"version"`
	ID      string       `json:"id"`
	Result  interface{}  `json:"result"`
	Error   *ErrResponse `json:"error"`
}

type ErrResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func New(store *store.Storage) RPC {
	logger := zap.NewNop()
	return &rpc{
		store:    store,
		commands: make(map[string]command.Command),
		logger:   logger,
	}
}

func (r *rpc) AddCommand(cmd command.Command) {
	r.commands[cmd.Name()] = cmd
}

func (r *rpc) HandleJSONRPC(ctx *gin.Context) {
	req := Request{}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, Response{Result: nil, Error: &ErrResponse{-1, err.Error()}, ID: req.ID, Version: req.Version})
		return
	}
	r.logger.Info("rpc request", zap.String("method", req.Method))
	params, err := json.Marshal(req.Params)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, Response{Result: nil, Error: &ErrResponse{-1, err.Error()}, ID: req.ID, Version: req.Version})
		return
	}
	resp, err := r.commands[req.Method].Execute(params)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, Response{Result: nil, Error: &ErrResponse{-1, err.Error()}, ID: req.ID, Version: req.Version})
		return
	}
	ctx.JSON(http.StatusOK, Response{Result: resp, Error: nil, ID: req.ID, Version: req.Version})
}

func Default(store *store.Storage, chainParams *chaincfg.Params) RPC {
	rpc := New(store)
	rpc.AddCommand(command.LatestBlock(store))
	rpc.AddCommand(command.UTXOs(store, chainParams))
	rpc.AddCommand(command.GetTx(store))
	rpc.AddCommand(command.GetTxsOfAddress(store, chainParams))
	return rpc
}

func (r *rpc) SetLogger(logger *zap.Logger) RPC {
	r.logger = logger
	return r
}

func (r *rpc) Run(port string) error {
	s := gin.Default()
	s.POST("/", r.HandleJSONRPC)
	return s.Run(port)
}
