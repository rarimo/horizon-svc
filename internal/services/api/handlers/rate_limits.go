package handlers

import (
	"github.com/google/jsonapi"
	"github.com/rarimo/horizon-svc/internal/config"
	"github.com/ulule/limiter/v3"
	"github.com/ulule/limiter/v3/drivers/middleware/stdlib"
	sredis "github.com/ulule/limiter/v3/drivers/store/redis"
	"gitlab.com/distributed_lab/ape"
	"gitlab.com/distributed_lab/logan/v3/errors"
	"net/http"
	"strconv"
)

func NewRateLimitsMiddleware(cfg config.Config) func(next http.Handler) http.Handler {
	rate := limiter.Rate{
		Period: cfg.RateLimiter().Period,
		Limit:  cfg.RateLimiter().Limit,
	}

	opts := limiter.StoreOptions{
		Prefix: cfg.RateLimiter().RedisPrefix,
	}

	if cfg.RateLimiter().RedisPrefix == "" {
		opts.Prefix = "api_rate_limits"
	}

	store, err := sredis.NewStoreWithOptions(cfg.RedisClient(), opts)
	if err != nil {
		panic(errors.Wrap(err, "failed to create a redis store"))
	}

	middleware := stdlib.NewMiddleware(
		limiter.New(store, rate),
		stdlib.WithErrorHandler(func(w http.ResponseWriter, r *http.Request, err error) {
			Log(r).WithError(err).Error("failed to check rate limits for requests, passing through")
		}),
		stdlib.WithLimitReachedHandler(func(w http.ResponseWriter, r *http.Request) {
			ape.RenderErr(w, &jsonapi.ErrorObject{
				Status: strconv.Itoa(http.StatusTooManyRequests),
				Title:  "Rate Limit Exceeded",
				Detail: "You have exceeded rate limit.",
				Meta: &map[string]interface{}{
					"limit":  cfg.RateLimiter().Limit,
					"period": cfg.RateLimiter().Period.String(),
				},
			})
		}),
	)

	return func(next http.Handler) http.Handler {
		return middleware.Handler(next)
	}
}
