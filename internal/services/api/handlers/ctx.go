package handlers

import (
	"context"
	"net/http"

	"github.com/rarimo/horizon-svc/internal/core"
	"github.com/rarimo/horizon-svc/internal/data"
	"github.com/rarimo/horizon-svc/internal/proxy"
	"gitlab.com/distributed_lab/logan/v3"
)

type ctxKey int

const (
	logCtxKey ctxKey = iota
	storageCtxKey
	cachedStorageCtxKey
	buildererCtxKey
	coreCtxKey
	proxyRepoCtxKey
	chainsQCtxKey
)

func CtxLog(entry *logan.Entry) func(context.Context) context.Context {
	return func(ctx context.Context) context.Context {
		return context.WithValue(ctx, logCtxKey, entry)
	}
}

func Log(r *http.Request) *logan.Entry {
	return r.Context().Value(logCtxKey).(*logan.Entry)
}

func CtxCore(entry core.Core) func(context.Context) context.Context {
	return func(ctx context.Context) context.Context {
		return context.WithValue(ctx, coreCtxKey, entry)
	}
}

func Core(r *http.Request) core.Core {
	return r.Context().Value(coreCtxKey).(core.Core)
}

func CtxProxyRepo(entry proxy.ProxyRepo) func(context.Context) context.Context {
	return func(ctx context.Context) context.Context {
		return context.WithValue(ctx, proxyRepoCtxKey, entry)
	}
}

func CtxChainsQ(entry data.ChainsQ) func(context.Context) context.Context {
	return func(ctx context.Context) context.Context {
		return context.WithValue(ctx, chainsQCtxKey, entry)
	}
}

func ChainsQ(r *http.Request) data.ChainsQ {
	return r.Context().Value(chainsQCtxKey).(data.ChainsQ)
}

func ProxyRepo(r *http.Request) proxy.ProxyRepo {
	return r.Context().Value(proxyRepoCtxKey).(proxy.ProxyRepo)
}

func CtxBuilder(bp TxBuilder) func(context.Context) context.Context {
	return func(ctx context.Context) context.Context {
		return context.WithValue(ctx, buildererCtxKey, bp)
	}
}

func Builder(r *http.Request) TxBuilder {
	return r.Context().Value(buildererCtxKey).(TxBuilder)
}

func CtxStorage(s data.Storage) func(context.Context) context.Context {
	return func(ctx context.Context) context.Context {
		return context.WithValue(ctx, storageCtxKey, s)
	}
}

func Storage(r *http.Request) data.Storage {
	return r.Context().Value(storageCtxKey).(data.Storage)
}

func CtxCachedStorage(s data.Storage) func(context.Context) context.Context {
	return func(ctx context.Context) context.Context {
		return context.WithValue(ctx, cachedStorageCtxKey, s)
	}
}

func CachedStorage(r *http.Request) data.Storage {
	return r.Context().Value(cachedStorageCtxKey).(data.Storage)
}
