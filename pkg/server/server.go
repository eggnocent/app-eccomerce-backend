package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync/atomic"
	"time"

	"github.com/eggnocent/app-eccomerce-backend/pkg/logging"
	"github.com/gorilla/mux"
	"github.com/spf13/viper"
)

type (
	key         int
	tracingType struct {
		XRequesterID  string `json:"requester_id"`
		RequestMethod string `json:"request_method"`
		URLPath       string `json:"url_path"`
		RemoteAddr    string `json:"remote_addr"`
		UserAgent     string `json:"user_agent"`
	}
)

const (
	requestIDKey key = 0
)

var (
	listenAddr string
	healthy    int32
)

func Init() {
	// placeholder if needed later
}

func NewServer(router *mux.Router, logger *logging.Logger) {
	logger.Out.Println("Server is starting...")

	router.Handle("/healthz", health())
	nextRequestID := func() string {
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}

	listenAddr = fmt.Sprintf(":%s", viper.GetString("server.port"))
	httpServer := &http.Server{
		Addr:         listenAddr,
		Handler:      tracing(nextRequestID)(httpLogging(logger)(router)),
		ReadTimeout:  time.Duration(viper.GetInt("server.read_timeout")) * time.Second,
		WriteTimeout: time.Duration(viper.GetInt("server.write_timeout")) * time.Second,
		IdleTimeout:  time.Duration(viper.GetInt("server.idle_timeout")) * time.Second,
	}

	done := make(chan bool)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)

	go func() {
		<-quit
		logger.Out.Info("Server is shutting down...")
		atomic.StoreInt32(&healthy, 0)
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		httpServer.SetKeepAlivesEnabled(false)
		if err := httpServer.Shutdown(ctx); err != nil {
			logger.Err.Fatalf("Could not gracefully shutdown the server: %v", err)
		}
		close(done)
	}()

	logger.Out.Infof("Server is ready to handle requests at %s", listenAddr)
	atomic.StoreInt32(&healthy, 1)
	if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Err.Fatalf("Could not listen on %s: %v", listenAddr, err)
	}
	<-done
	logger.Out.Println("Server stopped")
}

func health() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if atomic.LoadInt32(&healthy) == 1 {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		w.WriteHeader(http.StatusServiceUnavailable)
	})
}

func httpLogging(logger *logging.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				requestID, ok := r.Context().Value(requestIDKey).(string)
				if !ok {
					requestID = "unknown"
				}
				tracing := tracingType{
					XRequesterID:  requestID,
					RequestMethod: r.Method,
					URLPath:       r.URL.Path,
					RemoteAddr:    r.RemoteAddr,
					UserAgent:     r.UserAgent(),
				}
				jTracing, _ := json.Marshal(tracing)
				logger.Out.Println(string(jTracing))
			}()
			next.ServeHTTP(w, r)
		})
	}
}

func tracing(nextRequestID func() string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestID := r.Header.Get("X-Request-Id")
			if requestID == "" {
				requestID = nextRequestID()
			}
			ctx := context.WithValue(r.Context(), requestIDKey, requestID)
			w.Header().Set("X-Request-Id", requestID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
