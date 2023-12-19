package handler

import "google.golang.org/grpc"

type NewsService interface {
}

func Register(grpcSrv *grpc.Server, newsService NewsService) {

}
