# prometheus-utils

## Description

The useful library that makes working with Prometheus as easy as possible.

## Installation

Run the following command to install the package:

```
go get github.com/minipkg/prometheus-utils
```

## Basic Usage

### Metrics for a http server

#### with router "github.com/fasthttp/router"
```go
// create httpServerMetric object
metrics := prometheus_utils.NewHttpServerMetrics("chudo-app")

// set handler for Prometheus
r.GET("/metrics", prometheus_utils.GetFasthttpHandler())

// set the middleware for metrics for http handlers
r = metrics.FasthttpRouterMetricsMiddleware(r.Handler))
```

#### Example
```go
import (
    "context"
    "fmt"
    "github.com/pkg/errors"
    "net/http"
    "time"

    "github.com/fasthttp/router"
    prometheus_utils "github.com/minipkg/prometheus-utils"
    "github.com/satori/go.uuid"
    "github.com/valyala/fasthttp"
)

func main() {
    metrics := prometheus_utils.NewHttpServerMetrics("chudo-app")
    
    server := &fasthttp.Server{}
    
    r := router.New()
    
    r.GET("/live", LiveHandler)
    r.GET("/ready", LiveHandler)
    r.GET("/metrics", prometheus_utils.GetFasthttpHandler())
    r.GET("/api/v1/test", TestHandler)      //  handler is just for example

    r = RecoverInterceptorMiddleware(RequestIdInterceptorMiddleware(metrics.FasthttpRouterMetricsMiddleware(r.Handler)))
    
    server.Handler = r
        
    server.ListenAndServe(config.Addr)           //  address from config  is just for example
}

func RecoverInterceptorMiddleware(next fasthttp.RequestHandler) fasthttp.RequestHandler {
	return func(rctx *fasthttp.RequestCtx) {
		defer func() {
			if r := recover(); r != nil {
				wblogger.Error(rctx, "PanicInterceptor", fmt.Errorf("%v", r))
				fasthttp_tools.InternalError(rctx, errors.Errorf("%v", r))
			}
		}()
		next(rctx)
	}
}

func RequestIdInterceptorMiddleware(next fasthttp.RequestHandler) fasthttp.RequestHandler {
	return func(rctx *fasthttp.RequestCtx) {
		if requestId := rctx.UserValue(RequestIdKey); requestId == nil {
			if requestIdB := rctx.Request.Header.Peek(RequestIdKey); requestIdB != nil {
				rctx.SetUserValue(RequestIdKey, string(requestIdB))
			} else {
				rctx.SetUserValue(RequestIdKey, uuid.NewV4().String())
			}
		}
		next(rctx)

		return
	}
}

func LiveHandler(rctx *fasthttp.RequestCtx) {
	rctx.SetStatusCode(http.StatusNoContent)
	return
}

```

#### with router "github.com/qiangxue/fasthttp-routing"
```go
// create httpServerMetric object
metrics := prometheus_utils.NewHttpServerMetrics("chudo-app")

// set handler for Prometheus
r.Get("/metrics", prometheus_utils.GetFasthttpRoutingHandler())

// set the middleware for metrics for http handlers
r.Use(metrics.FasthttpRoutingMetricsMiddleware)
```

#### Example
```go
import (
    "context"
    "fmt"
    "github.com/pkg/errors"
    "net/http"
    "time"

    prometheus_utils "github.com/minipkg/prometheus-utils"
    routing "github.com/qiangxue/fasthttp-routing"
    "github.com/satori/go.uuid"
    "github.com/valyala/fasthttp"
)

func main() {
	metrics := prometheus_utils.NewHttpServerMetrics("chudo-app")

	server := &fasthttp.Server{}

	r := routing.New()

	r.Use(RecoverInterceptorMiddleware, RequestIdInterceptorMiddleware)
	r.Get("/live", LiveHandler)
	r.Get("/ready", LiveHandler)
	r.Get("/metrics", prometheus_utils.GetFasthttpRoutingHandler())
	api := r.Group("/api/v1")
	api.Use(metrics.FasthttpRoutingMetricsMiddleware)
	api.Get("/test", TestHandler)      //  handler is just for example
	server.Handler = r.HandleRequest

	server.ListenAndServe(config.Addr)             //  address from config  is just for example
}

func RecoverInterceptorMiddleware(rctx *routing.Context) error {
	defer func() {
		if r := recover(); r != nil {
			wblogger.Error(rctx, "PanicInterceptor", fmt.Errorf("%v", r))
			fasthttp_tools.InternalError(rctx.RequestCtx, errors.Errorf("%v", r))
		}
	}()
	rctx.Next()
	return nil
}

func RequestIdInterceptorMiddleware(rctx *routing.Context) error {
	if requestId := rctx.Get(RequestIdKey); requestId != nil {
		return nil
	}
	if requestIdB := rctx.RequestCtx.Request.Header.Peek(RequestIdKey); requestIdB != nil {
		rctx.Set(RequestIdKey, string(requestIdB))
		return nil
	}
	rctx.Set(RequestIdKey, uuid.NewV4().String())
    rctx.Next()
	return nil
}

func LiveHandler(rctx *routing.Context) error {
	rctx.SetStatusCode(http.StatusNoContent)
	return nil
}

```

