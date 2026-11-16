package main

import (
	"context"
	"errors"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/InSleepMode/VK-SPbU-BackendCourse/internal/adapter"
	"github.com/InSleepMode/VK-SPbU-BackendCourse/internal/domain"
	"github.com/InSleepMode/VK-SPbU-BackendCourse/internal/handler"
	"github.com/InSleepMode/VK-SPbU-BackendCourse/internal/migrations"
	"github.com/InSleepMode/VK-SPbU-BackendCourse/internal/repository"
	"github.com/InSleepMode/VK-SPbU-BackendCourse/internal/usecase"
	pb "github.com/InSleepMode/VK-SPbU-BackendCourse/pkg/movieapi"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	httpSwagger "github.com/swaggo/http-swagger"
	"google.golang.org/grpc"

	_ "github.com/InSleepMode/VK-SPbU-BackendCourse/docs"
)

func SyncMovies(ctx context.Context, client *adapter.HTTPMovieClient, repo *repository.MovieRepository) error {
	//находим точку старта
	maxID, err := repo.GetMaxID(ctx)
	if err != nil {
		return err
	}

	for id := maxID + 1; ; id++ {
		movie, err := client.GetMovie(ctx, id)
		if errors.Is(err, domain.ErrNotFound) {
			break
		}
		if err != nil {
			return err
		}
		err = repo.SaveMovie(ctx, movie)
		if id%10 == 0 {
			log.Printf("Загружен %d фильм", id)
		}
		if err != nil {
			return err
		}
	}
	return nil
}

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

	ctx := context.Background()

	//создаем пул на приложение
	databaseURL := os.Getenv("DATABASE_URL")
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := pool.Ping(pingCtx); err != nil {
		log.Fatal(err)
	}

	// Накатываем встроенные SQL-миграции до первого запроса.
	// Goose работает через database/sql, поэтому адаптируем пул pgx.
	migrationDB := stdlib.OpenDBFromPool(pool)
	defer migrationDB.Close()
	provider, err := goose.NewProvider(goose.DialectPostgres, migrationDB, migrations.Files)
	if err != nil {
		log.Fatalf("создать goose: %v", err)
	}
	// Up применяет только новые миграции, записывая версии в goose_db_version.
	if _, err := provider.Up(ctx); err != nil {
		log.Fatalf("применить миграции: %v", err)
	}

	repo := repository.NewMovieRepository(pool) //сначала репозиторий
	movieClient := adapter.NewHTTPMovieClient(baseURL)
	if err := SyncMovies(ctx, movieClient, repo); err != nil {
		log.Fatal(err)
	}

	movieUC := usecase.NewMovieUseCase(repo) //теперь ходим в бд, а не во внешний стенд
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
