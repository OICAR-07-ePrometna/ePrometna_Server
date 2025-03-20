package httpServer

import (
	"context"
	"ePrometna_Server/config"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

var signalNotificationCh = make(chan os.Signal, 1)

func Start() {
	// relay selected signals to channel
	// - os.Interrupt, ctrl-c
	// - syscall.SIGTERM, program termination
	signal.Notify(signalNotificationCh, os.Interrupt, syscall.SIGTERM)

	// create scheduler
	schedulerWg := sync.WaitGroup{}
	schedulerCtx := context.Background()
	schedulerCtx, schedulerCancel := context.WithCancel(schedulerCtx)
	zap.S().Debugf("Created scheduler context")

	schedulerWg.Add(1)
	go checkInterrupt(schedulerCtx, &schedulerWg, schedulerCancel)
	zap.S().Debugf("Started CheckInterrupt")

	schedulerWg.Add(1)
	go run(schedulerCtx, &schedulerWg)
	zap.S().Debugf("Started HTTP server")

	schedulerWg.Wait()

	zap.S().Debugf("Terminated program")
}

func checkInterrupt(ctx context.Context, wg *sync.WaitGroup, schedulerCancel context.CancelFunc) {
	defer wg.Done()

	for {
		select {

		case <-ctx.Done():
			zap.S().Debugf("Terminated CheckInterrupt")
			return

		case sig := <-signalNotificationCh:
			zap.S().Debugf("Received signal on notification channel, signal = %v", sig)
			schedulerCancel()
		}
	}
}

func run(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	// gin.DisableConsoleColor()
	if !config.AppConfig.IsDevelopment {
		gin.SetMode(gin.ReleaseMode)
	}
	router := gin.Default()
	setupHandlers(router)

	addr := fmt.Sprintf(":%d", config.AppConfig.Port)
	srv := &http.Server{
		Addr:         addr,
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  10 * time.Second,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			zap.S().Panicf("Failes to start server err = %+v", err)
		}
	}()
	zap.S().Infof("Started HTTP listen, address = http://localhost%v", srv.Addr)

	// wait for context cancellation
	<-ctx.Done()

	timeoutCtx, timeoutCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer timeoutCancel()
	err := srv.Shutdown(timeoutCtx)
	if err != nil {
		zap.S().Errorf("Cannot shut down HTTP server, err = %v", err)
	}
	zap.S().Info("HTTP server was shut down")
}

/*
func headersMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		headers := w.Header()
		headers.Add("Access-Control-Allow-Origin", "http://localhost:3000")
		headers.Add("Vary", "Origin")
		headers.Add("Vary", "Access-Control-Request-Method")
		headers.Add("Vary", "Access-Control-Request-Headers")
		headers.Add("Access-Control-Allow-Headers", "Content-Type, Origin, Accept, token, Access-Control-Allow-Origin")
		headers.Add("Access-Control-Allow-Methods", "GET,POST,OPTIONS,DELETE")
		if r.Method == "OPTIONS" {
			return
		}

		next.ServeHTTP(w, r)
	})
}
*/

func corsHeader() gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: write hteaders if neaded
	}
}
