package handler

import (
	"context"
	"time"

	"github.com/InSleepMode/VK-SPbU-BackendCourse/internal/usecase"
	pb "github.com/InSleepMode/VK-SPbU-BackendCourse/pkg/movieapi"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type GRPCServer struct {
	pb.UnimplementedMovieServiceServer
	uc usecase.MovieUseCase
}

func NewGRPCServer(uc usecase.MovieUseCase) *GRPCServer {
	return &GRPCServer{uc: uc}
}

func (s *GRPCServer) GetLastMovie(ctx context.Context, req *pb.GetLastMovieRequest) (*pb.Movie, error) {
	movie, err := s.uc.GetLastMovie(ctx)
	if err != nil {
		return nil, err
	}
	return &pb.Movie{
		Id: int64(movie.ID),
	}, nil
}

func (s *GRPCServer) LoadMovies(ctx context.Context, req *pb.LoadMoviesRequest) (*pb.LoadMoviesResponse, error) {
	//validation
	if req.From <= 0 || req.From > req.To {
		return nil, status.Errorf(codes.InvalidArgument, "invalid from/to arguments")
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	movies, err := s.uc.LoadMovies(ctx, int(req.From), int(req.To))
	if err != nil {
		return nil, err
	}

	var pbMovies []*pb.Movie
	for _, m := range movies {
		pbMovies = append(pbMovies, &pb.Movie{Id: int64(m.ID)})
	}

	return &pb.LoadMoviesResponse{Movies: pbMovies}, nil
}
