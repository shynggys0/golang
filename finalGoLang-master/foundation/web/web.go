package web

import (
	"context"
	"net/http"
	"os"
	"syscall"
	"time"

	"github.com/google/uuid"
	"github.com/dimfeld/httptreemux/v5"
)

type ctxKey int

const KeyValues ctxKey = 1

type Values struct {
	TraceID    string
	Now        time.Time
	StatusCode int
}


type Handler func(ctx context.Context, w http.ResponseWriter, r *http.Request) error

type App struct {
	*httptreemux.ContextMux
	shutdown chan os.Signal
	mw       []Middleware
}

func NewApp(shutdown chan os.Signal, mw ...Middleware) *App {
	return &App{
		ContextMux: httptreemux.NewContextMux(),
		shutdown: shutdown,
		mw: mw,
	}
}

func (a *App) Handle(method string,path string, handler Handler, mw ...Middleware) {

	handler = wrapMiddleware(mw, handler)


	handler = wrapMiddleware(a.mw, handler)

	h := func(w http.ResponseWriter, r *http.Request)  {
		v:= Values {
			TraceID: uuid.New().String(),
			Now: time.Now(),
		}
		ctx := context.WithValue(r.Context(), KeyValues, &v)


		err := handler(ctx,w,r)
		if err != nil {
			a.SignalShutdown()
			return
		}
		
	}


	a.ContextMux.Handle(method,path, h)
}


func (a *App) SignalShutdown() {
	a.shutdown <- syscall.SIGTERM
}