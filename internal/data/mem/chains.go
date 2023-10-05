package mem

import (
	"strings"

	"github.com/rarimo/horizon-svc/internal/data"
)

func NewChainsQ(chains []data.Chain) data.ChainsQ {
	return &chainsQ{
		chains: chains,
	}
}

type chainsQ struct {
	chains []data.Chain
}

func (q *chainsQ) Get(chain string) *data.Chain {
	for _, value := range q.chains {
		if strings.ToLower(value.Name) == strings.ToLower(chain) {
			return &value
		}
	}

	return nil
}

func (q *chainsQ) Page(pageNum, limit int) []data.Chain {
	if pageNum*limit > len(q.chains) {
		return []data.Chain{}
	}

	if (pageNum+1)*limit >= len(q.chains) {
		return q.chains[pageNum*limit:]
	}

	return q.chains[pageNum*limit : (pageNum+1)*limit]

}

func (q *chainsQ) List() []data.Chain {
	return q.chains
}
