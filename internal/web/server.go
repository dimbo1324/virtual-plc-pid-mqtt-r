package web

import (
	"context"
	"fmt"
	"io/fs"
	"log/slog"
	"net"
	"net/http"
	"time"

	"github.com/dimbo1324/virtual-plc-pid-mqtt-r/internal/storage"
	"github.com/dimbo1324/virtual-plc-pid-mqtt-r/pkg/plc"
)

// CommandHandler applies a PLC command and returns the resulting event.
// It mirrors mqttx.CommandHandler so the same wired function can be passed
// from internal/app without creating a cross-package dependency.
type CommandHandler func(context.Context, plc.Command) (plc.Event, error)

// Deps holds external dependencies injected into the web server.
type Deps struct {
	Runtime         *plc.Runtime
	Store           *storage.Store // nil if storage is disabled
	CommandHandler  CommandHandler
	EventsCh        <-chan plc.Event
	SnapshotsCh     <-chan plc.Snapshot
	StorageStatusFn func() string // returns "ok", "degraded", or "disabled"
}

// Server is an embedded HTTP dashboard for the virtual PLC.
type Server struct {
	cfg     Config
	deps    Deps
	broker  *sseBroker
	srv     *http.Server
	started time.Time
	logger  *slog.Logger
}

// NewServer creates a Server ready to be started.
func NewServer(cfg Config, deps Deps, logger *slog.Logger) *Server {
	if logger == nil {
		logger = slog.Default()
	}
	s := &Server{
		cfg:    cfg,
		deps:   deps,
		broker: newSSEBroker(),
		logger: logger,
	}
	mux := http.NewServeMux()
	s.registerRoutes(mux)
	s.srv = &http.Server{
		Addr:         cfg.addr(),
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 0, // disabled for SSE streaming connections
		IdleTimeout:  60 * time.Second,
	}
	return s
}

func (s *Server) registerRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /healthz", s.handleHealthz)
	mux.HandleFunc("GET /readyz", s.handleReadyz)
	mux.HandleFunc("GET /api/status", s.handleStatus)
	mux.HandleFunc("GET /api/snapshot", s.handleSnapshot)
	mux.HandleFunc("GET /api/events/recent", s.handleRecentEvents)
	mux.HandleFunc("GET /api/telemetry/recent", s.handleRecentTelemetry)
	mux.HandleFunc("POST /api/commands/start", s.handleCommandStart)
	mux.HandleFunc("POST /api/commands/stop", s.handleCommandStop)
	mux.HandleFunc("POST /api/commands/setpoint", s.handleCommandSetpoint)
	mux.HandleFunc("POST /api/commands/pid-gains", s.handleCommandPIDGains)
	mux.HandleFunc("POST /api/commands/mode", s.handleCommandMode)
	mux.HandleFunc("POST /api/commands/manual-output", s.handleCommandManualOutput)
	mux.HandleFunc("POST /api/commands/inject-disturbance", s.handleCommandInjectDisturbance)
	mux.HandleFunc("POST /api/commands/reset-loop", s.handleCommandResetLoop)
	mux.HandleFunc("GET /api/stream", s.broker.serveSSE)

	subFS, _ := fs.Sub(staticFS, "assets")
	mux.Handle("/", http.FileServer(http.FS(subFS)))
}

// Handler returns the underlying http.Handler for use in httptest.
func (s *Server) Handler() http.Handler { return s.srv.Handler }

// PumpSSE launches the SSE fan-out goroutine. Called automatically by Start;
// exposed here so tests can start it without binding to a TCP port.
func (s *Server) PumpSSE(ctx context.Context) { go s.pumpSSE(ctx) }

// Start begins listening and blocks until ctx is cancelled.
func (s *Server) Start(ctx context.Context) error {
	s.started = time.Now()

	go s.pumpSSE(ctx)

	ln, err := net.Listen("tcp", s.srv.Addr)
	if err != nil {
		return fmt.Errorf("web: listen %s: %w", s.srv.Addr, err)
	}
	s.logger.Info("web dashboard listening", "addr", s.srv.Addr)

	errCh := make(chan error, 1)
	go func() {
		if err := s.srv.Serve(ln); err != nil && err != http.ErrServerClosed {
			errCh <- err
		} else {
			errCh <- nil
		}
	}()

	select {
	case <-ctx.Done():
		return s.Shutdown(context.Background())
	case err := <-errCh:
		return err
	}
}

// Shutdown gracefully stops the server within 5 seconds.
func (s *Server) Shutdown(ctx context.Context) error {
	shutdownCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	return s.srv.Shutdown(shutdownCtx)
}

// Addr returns the configured listen address.
func (s *Server) Addr() string { return s.srv.Addr }

// pumpSSE reads from fan-out channels and broadcasts to SSE clients.
func (s *Server) pumpSSE(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case snap, ok := <-s.deps.SnapshotsCh:
			if !ok {
				return
			}
			s.broker.broadcast("snapshot", snap)
		case event, ok := <-s.deps.EventsCh:
			if !ok {
				return
			}
			s.broker.broadcast("plc_event", event)
		}
	}
}
