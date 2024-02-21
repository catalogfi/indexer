package rpc

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/catalogfi/indexer/command"
	"github.com/catalogfi/indexer/store"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type RPC interface {
	RegisterCommand(cmd command.Command)
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
	Version string      `json:"version"`
	ID      string      `json:"id"`
	Result  interface{} `json:"result"`
	Error   *RpcError   `json:"error"`
}

func New(store *store.Storage) RPC {
	logger := zap.NewNop()
	return &rpc{
		store:    store,
		commands: make(map[string]command.Command),
		logger:   logger,
	}
}

func (r *rpc) RegisterCommand(cmd command.Command) {
	r.commands[cmd.Name()] = cmd
}

func (r *rpc) HandleJSONRPC(ctx *gin.Context) {
	req := Request{}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, Response{Result: nil, Error: NewInvalidRequestError(err.Error()), ID: req.ID, Version: req.Version})
		return
	}
	r.logger.Info("rpc request", zap.String("method", req.Method))
	params, err := json.Marshal(req.Params)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, Response{Result: nil, Error: NewInvalidParamsError(err.Error()), ID: req.ID, Version: req.Version})
		return
	}

	//check if the method is registered or not
	if _, ok := r.commands[req.Method]; !ok {
		ctx.JSON(http.StatusBadRequest, Response{Result: nil, Error: NewMethodNotFoundError(), ID: req.ID, Version: req.Version})
		return
	}

	resp, err := r.commands[req.Method].Execute(params)
	if err != nil {
		if errors.As(err, &RpcError{}) {
			ctx.JSON(http.StatusBadRequest, Response{Result: nil, Error: err.(*RpcError), ID: req.ID, Version: req.Version})
			return
		}
		ctx.JSON(http.StatusBadRequest, Response{Result: nil, Error: NewInternalError(err.Error()), ID: req.ID, Version: req.Version})
		return
	}
	ctx.JSON(http.StatusOK, Response{Result: resp, Error: nil, ID: req.ID, Version: req.Version})
}

func Default(store *store.Storage, chainParams *chaincfg.Params) RPC {
	rpc := New(store)
	rpc.RegisterCommand(command.LatestTip(store))
	rpc.RegisterCommand(command.UTXOs(store, chainParams))
	rpc.RegisterCommand(command.GetTx(store))
	rpc.RegisterCommand(command.GetTxsOfAddress(store, chainParams))
	rpc.RegisterCommand(command.LatestTipHash(store))
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
