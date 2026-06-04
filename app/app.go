package app

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"personal/wassup/auth"
	"personal/wassup/bl"
	"personal/wassup/config"
	"personal/wassup/db"
	"personal/wassup/db/indexes"
	"personal/wassup/media"
	"personal/wassup/proto"
	"personal/wassup/redis"
	"personal/wassup/transport"
	"personal/wassup/ws"
	"time"

	grpc "google.golang.org/grpc"
)

const (
	httpAddr         = ":8080"
	grpcAddr         = ":50051"
	httpCertFilePath = "./certs/cert.pem"
	httpKeyFilePath  = "./certs/key.pem"
	shutdownTimeout  = 10 * time.Second
)

func Run(ctx context.Context) error {
	cfg := config.NewConfig()
	if err := cfg.Init(); err != nil {
		return fmt.Errorf("initialize config: %w", err)
	}

	dbHandler := db.NewDBHandler(cfg)
	if err := dbHandler.Init(); err != nil {
		return fmt.Errorf("initialize db: %w", err)
	}

	client, err := dbHandler.GetClient()
	if err != nil {
		_ = dbHandler.Close(context.Background())
		return fmt.Errorf("get db client: %w", err)
	}

	indexHandler := indexes.NewIndexHandler(client)
	if err := indexHandler.AddIndexes(ctx, db.DBName); err != nil {
		_ = dbHandler.Close(context.Background())
		return fmt.Errorf("ensure db indexes: %w", err)
	}

	redisHandler := redis.NewCacheHandler(cfg)
	if err := redisHandler.Init(ctx); err != nil {
		_ = dbHandler.Close(context.Background())
		return fmt.Errorf("initialize redis: %w", err)
	}

	grpcClient := proto.NewGrpcClient(redisHandler)
	tokenService := auth.NewTokenService()
	services := bl.NewServices(dbHandler, redisHandler, grpcClient, tokenService)

	webSocketHandler := ws.NewWebSocketHandler(services.Conversation, redisHandler)
	mediaHandler := media.NewMediaHandler()
	httpHandler := transport.NewHandler(services, mediaHandler)
	router := transport.Router(httpHandler, webSocketHandler)

	httpServer := &http.Server{
		Addr:    httpAddr,
		Handler: router,
	}

	grpcServer := proto.NewGrpcServer(webSocketHandler, redisHandler)
	grpcListener, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		_ = redisHandler.Close()
		_ = dbHandler.Close(context.Background())
		return fmt.Errorf("listen grpc on %s: %w", grpcAddr, err)
	}

	errCh := make(chan error, 2)
	go func() {
		log.Printf("gRPC listening on %s", grpcAddr)
		if serveErr := grpcServer.Serve(grpcListener); serveErr != nil && !errors.Is(serveErr, grpc.ErrServerStopped) {
			errCh <- fmt.Errorf("serve grpc: %w", serveErr)
		}
	}()

	go func() {
		log.Printf("HTTP listening on %s", httpAddr)
		if serveErr := httpServer.ListenAndServeTLS(httpCertFilePath, httpKeyFilePath); serveErr != nil && !errors.Is(serveErr, http.ErrServerClosed) {
			errCh <- fmt.Errorf("serve http: %w", serveErr)
		}
	}()

	select {
	case <-ctx.Done():
		if err := shutdown(httpServer, grpcServer, grpcListener, redisHandler, dbHandler); err != nil {
			return fmt.Errorf("shutdown after signal: %w", err)
		}
		return nil
	case err := <-errCh:
		shutdownErr := shutdown(httpServer, grpcServer, grpcListener, redisHandler, dbHandler)
		if shutdownErr != nil {
			return fmt.Errorf("runtime error: %v; shutdown error: %w", err, shutdownErr)
		}
		return err
	}
}

func shutdown(httpServer *http.Server, grpcServer *grpc.Server, grpcListener net.Listener, redisHandler *redis.CacheHandler, dbHandler *db.DBHandler) error {
	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	var firstErr error

	if err := httpServer.Shutdown(shutdownCtx); err != nil && !errors.Is(err, context.Canceled) {
		firstErr = fmt.Errorf("shutdown http server: %w", err)
	}

	stopped := make(chan struct{})
	go func() {
		grpcServer.GracefulStop()
		close(stopped)
	}()

	select {
	case <-stopped:
	case <-shutdownCtx.Done():
		grpcServer.Stop()
	}

	if err := grpcListener.Close(); err != nil && firstErr == nil {
		firstErr = fmt.Errorf("close grpc listener: %w", err)
	}

	if err := redisHandler.Close(); err != nil && firstErr == nil {
		firstErr = fmt.Errorf("close redis: %w", err)
	}

	if err := dbHandler.Close(shutdownCtx); err != nil && firstErr == nil {
		firstErr = fmt.Errorf("close db: %w", err)
	}

	return firstErr
}
