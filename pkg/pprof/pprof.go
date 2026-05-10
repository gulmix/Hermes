package pprof

import (
	"net/http"
	_ "net/http/pprof"
	"time"

	"go.uber.org/zap"
)

func Serve(addr string, log *zap.Logger) {
	srv := &http.Server{
		Addr:         addr,
		Handler:      http.DefaultServeMux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}
	log.Info("pprof server started", zap.String("addr", addr))
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Error("pprof server error", zap.Error(err))
	}
}
