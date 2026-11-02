package main

import (
	"log"
	"net"
	"net/http"
	"os"

	"github.com/InSleepMode/VK-SPbU-BackendCourse/internal/adapter"
	"github.com/InSleepMode/VK-SPbU-BackendCourse/internal/handler"
	"github.com/InSleepMode/VK-SPbU-BackendCourse/internal/usecase"
	pb "github.com/InSleepMode/VK-SPbU-BackendCourse/pkg/movieapi"
	httpSwagger "github.com/swaggo/http-swagger"
	"google.golang.org/grpc"

	_ "github.com/InSleepMode/VK-SPbU-BackendCourse/docs"
)

// @title movie api
// @version 1.0
// @description Go-сервис, который предоставляет HTTP и gRPC API и внутри ходит во внешний фильм-стенд
// @host localhost:8080
// @BasePath /
func main() {

	baseURL := os.Getenv("MOVIE_BASE_URL")
	if baseURL == "" {
		baseURL = "https://homeworksite.site"
	}

	httpAddr := os.Getenv("HTTP_ADDR")
	if httpAddr == "" {
		httpAddr = ":8080"
	}

	grpcAddr := os.Getenv("GRPC_ADDR")
	if grpcAddr == "" {
		grpcAddr = ":50051"
	}

	movieClient := adapter.NewHTTPMovieClient(baseURL)
	movieUC := usecase.NewMovieUseCase(movieClient)
	httpSrv := handler.NewHTTPServer(movieUC)
	grpcSrv := handler.NewGRPCServer(movieUC)

	go func() {
		lis, err := net.Listen("tcp", grpcAddr)
		if err != nil {
			log.Fatalf("failed to listen grpc: %v", err)
		}

		grpcServer := grpc.NewServer()
		pb.RegisterMovieServiceServer(grpcServer, grpcSrv)

		log.Printf("Starting gRPC server on %s", grpcAddr)
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("failed to serve grpc: %v", err)
		}
	}()

	mux := http.NewServeMux()
	httpSrv.RegisterRoutes(mux)

	//handler for swagger
	mux.Handle("/swagger/", httpSwagger.WrapHandler)

	log.Printf("Starting HTTP server on %s", httpAddr)
	if err := http.ListenAndServe(httpAddr, mux); err != nil {
		log.Fatalf("server failed to start: %v", err)
	}
}
