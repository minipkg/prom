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
    regionController := controller.NewRegionController(r, a.Domain.Region)
    
    r.GET("/live", LiveHandler)
    r.GET("/ready", LiveHandler)
    r.GET("/metrics", prometheus_utils.GetFasthttpHandler())
    r.GET("/api/v1/region-id", regionController.GetRegionID)

    router := RecoverInterceptorMiddleware(RequestIdInterceptorMiddleware(metrics.FasthttpRouterMetricsMiddleware(r.Handler)))
    
    server.Handler = router
        
    server.ListenAndServe(config.Addr)
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

#### with router "github.com/fasthttp/router"
```go
import (
	"context"
	"fmt"
	"log"

	"github.com/minipkg/nats"
	"github.com/nats-io/nats.go"
)

const (
	queueGroupName = "groupExample"
	consumerName   = "consumerExample"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())

	n, err := mp_nats.New(&mp_nats.Config{})
	if err != nil {
		log.Fatal(err)
	}

	_, err = n.AddStreamIfNotExists(&nats.StreamConfig{
		Name:     streamName,
		Subjects: []string{"test.>"},
	})
	if err != nil {
		log.Fatalf("natsWriter error: %q", err.Error())
	}

	_, _, delFunc, err := n.AddPushConsumerIfNotExists(streamName, &nats.ConsumerConfig{
		Name:    consumerName,
		Durable: consumerName,
		//DeliverGroup:  queueGroupName, // if you want queue group
		FilterSubject: subjectName,
	}, natsMsgHandler)
	if err != nil {
		log.Fatalf("natsWriter error: %q", err.Error())
	}
	defer func() {
		if err = delFunc(); err != nil {
			log.Fatalf("delConsumerAndSubscription error: %q", err.Error())
		}
	}()

	<-ctx.Done()
}

func natsMsgHandler(msg *nats.Msg) {
    msg.Ack()
	fmt.Println(string(msg.Data))
}
```

### Pull Consumer
```go
import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/pkg/errors"

	"github.com/minipkg/nats"
	"github.com/nats-io/nats.go"
)

const (
	consumerName        = "consumerExample"
	defaultRequestBatch = 1000
	defaultMaxWait      = 3 * time.Second
	duration            = 2 * time.Second
)


func main() {
	ctx, cancel := context.WithCancel(context.Background())

	n, err := mp_nats.New(&mp_nats.Config{})
	if err != nil {
		log.Fatal(err)
	}

	_, err = n.AddStreamIfNotExists(&nats.StreamConfig{
		Name:     streamName,
		Subjects: []string{"test.>"},
	})
	if err != nil {
		log.Fatalf("natsWriter error: %q", err.Error())
	}

	_, subs, delFunc, err := n.AddPullConsumerIfNotExists(streamName, &nats.ConsumerConfig{
		Name:    consumerName,
		Durable: consumerName,
		FilterSubject: subjectName,
	})
	if err != nil {
		log.Fatalf("natsWriter error: %q", err.Error())
	}
	defer func() {
		if err = delFunc(); err != nil {
			log.Fatalf("delConsumerAndSubscription error: %q", err.Error())
		}
	}()

    err = listenNatsSubscription(ctx, subs, 0)
	if err != nil {
		log.Fatalf("listenNatsSubscription error: %q", err.Error()))
		return
	}
}

func listenNatsSubscription(ctx context.Context, subs *nats.Subscription, requestBatch uint) error {
	if requestBatch == 0 {
		requestBatch = defaultRequestBatch
	}
OuterLoop:
	for {
		select {
		case <-ctx.Done():
			break OuterLoop
		default:
		}

		bmsgs, err := subs.Fetch(int(requestBatch), nats.MaxWait(defaultMaxWait))
		if err != nil {
			if !errors.Is(err, nats.ErrTimeout) {
				return err
			}

			t := time.NewTimer(duration)
			select {
			case <-ctx.Done():
				break OuterLoop
			case <-t.C:
			}

		}
		for _, msg := range bmsgs {
			if err = msg.Ack(); err != nil {
				return err
			}
			natsMsgHandler(msg)
		}
	}
	return nil
}

func natsMsgHandler(msg *nats.Msg) {
	fmt.Println(string(msg.Data))
}
```