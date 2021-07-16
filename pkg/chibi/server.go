package chibi

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rs/zerolog/log"
)

const (
	readHeaderTimeoutSeconds = 60
	shutdownTimeout          = 30
)

// Serve starts http server.
func Serve(mux http.Handler, port string) {
	s := &http.Server{
		Addr:              ":" + port,
		Handler:           mux,
		ReadHeaderTimeout: readHeaderTimeoutSeconds * time.Second,
	}

	idleConnsClosed := make(chan struct{})

	go func() {
		ch := make(chan os.Signal, 1)
		signal.Notify(ch, syscall.SIGTERM, syscall.SIGINT)

		sig := <-ch
		log.Info().Str("signal", sig.String()).Msg("received signal")
		log.Info().Msg("terminating...")

		ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout*time.Second)
		defer cancel()

		if err := s.Shutdown(ctx); err != nil {
			log.Err(err).Msg(err.Error())
		}

		log.Info().Msg("shutdown completed")

		close(idleConnsClosed)
	}()

	log.Info().Msg("started")

	if err := s.ListenAndServe(); err != http.ErrServerClosed {
		log.Fatal().Err(err).Msg(err.Error())
	}

	<-idleConnsClosed

	log.Info().Msg("bye")
}
