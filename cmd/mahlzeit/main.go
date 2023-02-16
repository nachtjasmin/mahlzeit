package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"codeberg.org/mahlzeit/mahlzeit/internal/app"
)

const ExitCodeOnError = 1

func main() {
	defer recoverPanic()
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM)
	defer cancel()

	if err := run(ctx, os.Args[1:]); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "unexpected error during execution: %s", err)

		// Since os.Exit would skip the deferred statements, the context cancellation is invoked
		// manually at this point.
		cancel()

		os.Exit(ExitCodeOnError) // nolint:gocritic
	}
}

// We deliberately use the main function as the entrypoint to wrap basic command
// execution into run. That itself makes it testable and the provided [context.Context] can be
// used for downstream goroutines to cancel their operations.
func run(ctx context.Context, args []string) error {
	cfg := &app.Application{
		Templates: app.NewTemplates("./web/templates"),
	}

	log.Println("Starting server on :4000")
	h := &http.Server{
		BaseContext: func(net.Listener) context.Context {
			return ctx
		},
		Addr:              ":4000",
		Handler:           routes(cfg),
		ReadHeaderTimeout: time.Second, // protect against SLOWLORIS attack
	}
	return h.ListenAndServe()
}

func recoverPanic() {
	if rec := recover(); rec != nil {
		err := rec.(error)
		log.Printf("unhandled error: %v", err)
		fmt.Fprintf(os.Stderr, "Program quit unexpectedly; please check your logs\n")
		os.Exit(ExitCodeOnError)
	}
}
