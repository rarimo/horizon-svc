package core

import (
	tokenmanager "github.com/rarimo/rarimo-core/x/tokenmanager/types"
	"google.golang.org/grpc"
)

type Core interface {
	TokenManager() TokenManager
}

type core struct {
	cli *grpc.ClientConn
}

func NewCore(cli *grpc.ClientConn) Core {
	return &core{
		cli: cli,
	}
}

func (c *core) TokenManager() TokenManager {
	return &tokenManager{tokenmanager: tokenmanager.NewQueryClient(c.cli)}
}
