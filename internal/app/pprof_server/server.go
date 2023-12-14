package pprof_server

import (
	"context"
	"net/http"
	"net/http/pprof"
	"time"

	"github.com/sirupsen/logrus"
)

func ServePprof(ctx context.Context, log *logrus.Logger) {
	const op = "pprof.ServePprof"

	srv := http.Server{
		Addr:         ":3366",
		Handler:      pprofHandler(),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Minute,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Infof("starting pprof server: http://localhost%s", srv.Addr)
		err := srv.ListenAndServe()
		if err != nil {
			log.Errorf("%s: %s", op, err.Error())
		}
	}()

	<-ctx.Done()
	srv.SetKeepAlivesEnabled(false)
	_ = srv.Shutdown(context.Background())
}

func pprofHandler() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/debug/pprof/", pprof.Index)
	mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	mux.HandleFunc("/debug/prof/symbol", pprof.Symbol)
	mux.HandleFunc("/debug/pprof/trace", pprof.Trace)
	mux.Handle("/debug/pprof/block", pprof.Handler("block"))
	mux.Handle("/debug/pprof/heap", pprof.Handler("heap"))
	mux.Handle("/debug/pprof/goroutine", pprof.Handler("goroutine"))
	mux.Handle("/debug/pprof/threadcreate", pprof.Handler("threadcreate"))

	return mux
}
