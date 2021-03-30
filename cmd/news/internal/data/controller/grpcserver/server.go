//go:generate  protoc -I. -I../../../../../../api api.proto --go_out=plugins=grpc:../../../../../../api --grpc-gateway_out=logtostderr=true:../../../../../../api
//go:generate  protoc -I/usr/local/include -I. -I$GOPATH/src -I$GOPATH/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis --go_out=plugins=grpc:. *.proto

package grpcserver

import (
	"context"
	grpcservice "github.com/shipa988/hw_otus_architect/api"
	"github.com/shipa988/hw_otus_architect/cmd/news/internal/domain/usecase"
	"github.com/shipa988/hw_otus_architect/internal/data/controller/log"
	"github.com/shipa988/hw_otus_architect/internal/domain/entity"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/golang/protobuf/ptypes"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const (
	ErrId = "can't parse user id"
)

var headers = []string{"token"}

const token = "secretfornewsservice"

type GRPCServer struct {
	newsInteractor usecase.NetworkNews
	server         *grpc.Server
	gwserver       *http.Server
}

func NewGRPCServer(newsiteractor usecase.NetworkNews) *GRPCServer {
	return &GRPCServer{
		newsInteractor: newsiteractor}
}

func (cs *GRPCServer) GetNews(ctx context.Context, req *grpcservice.GetNewsRequest) (*grpcservice.GetNewsResponse, error) {
	id, err := strconv.ParseUint(req.GetUserid(), 10, 64)
	if err != nil {
		return nil, status.Error(codes.Aborted, err.Error())
	}
	news, err := cs.newsInteractor.GetNews(ctx, id)
	if err != nil {
		return nil, status.Error(codes.Aborted, err.Error())
	}

	ns, err := toPBNews(news)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	resp := &grpcservice.GetNewsResponse{
		News: ns,
	}
	return resp, nil
}

func (cs *GRPCServer) ServeGW(addr string, addrgw string) error {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	log.Info("starting grpc gateway server at %v", addrgw)

	mux := runtime.NewServeMux(
		runtime.WithMetadata(injectHeadersIntoMetadata),
	)
	opts := []grpc.DialOption{grpc.WithInsecure()}

	err := grpcservice.RegisterNewsServiceHandlerFromEndpoint(ctx, mux, addr, opts)
	if err != nil {
		return errors.Wrapf(err, "can't register gateway from grpc endpoint at addr %v", addr)
	}
	cs.gwserver = &http.Server{
		Addr:    addrgw,
		Handler: mux,
	}

	if err := cs.gwserver.ListenAndServe(); err != http.ErrServerClosed {
		return errors.Wrapf(err, "can't start  grpc gateway server at %v", addrgw)
	}
	return nil
}

func (cs *GRPCServer) StopGWServe() {
	ctx := context.Background()
	log.Info("stopping grpc gw server")
	defer log.Info("grpc gw stopped")
	if cs.gwserver == nil {
		log.Error("grpc gw server is nil")
		return
	}
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := cs.gwserver.Shutdown(ctx); err != nil {
		log.Error(errors.Wrap(err, "can't stop grpc gw server"))
	}
}

func (cs *GRPCServer) PrepareGRPCListener(addr string) (net.Listener, error) {
	log.Info("GRPC server: starting tcp listener at %v", addr)
	l, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, errors.Wrapf(err, "GRPC server: can't start tcp listening at addr %v", addr)
	}
	return l, nil
}

func (cs *GRPCServer) Serve(listener net.Listener) error {
	log.Info("starting grpc server at %v", listener.Addr().String())

	cs.server = grpc.NewServer(
		grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(cs.authUnary)),
	)
	grpcservice.RegisterNewsServiceServer(cs.server, cs)

	if err := cs.server.Serve(listener); err != http.ErrServerClosed {
		return errors.Wrapf(err, "can't start grpc server at %v", listener.Addr().String())
	}
	return nil
}

func (cs *GRPCServer) StopServe() {
	log.Info("stopping grpc server")
	defer log.Info("grpc server stopped")
	cs.server.GracefulStop()
}

func (cs *GRPCServer) authUnary(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		if t, ok := md["token"]; !ok || len(t) == 0 || t[0] != token {
			return nil, status.Error(codes.PermissionDenied, "token is unknown")
		}
	}
	return handler(ctx, req)
}

func toPBNews(news []entity.News) (*grpcservice.News, error) {
	grpcNews := make([]*grpcservice.OneNews, 0, len(news))
	for _, onewnews := range news {
		pbOnewNews, err := toPBOnewNews(&onewnews)
		if err != nil {
			return nil, err
		}
		grpcNews = append(grpcNews, pbOnewNews)
	}
	return &grpcservice.News{OneNews: grpcNews}, nil
}

func toPBOnewNews(oneNews *entity.News) (*grpcservice.OneNews, error) {
	pbe := &grpcservice.OneNews{
		Id:            oneNews.Id,
		Author:        oneNews.AuthorId,
		AuthorGen:     oneNews.AuthorGen,
		Title:         oneNews.Title,
		Content:       oneNews.Text,
		AuthorName:    oneNews.AuthorName,
		AuthorSurname: oneNews.AuthorSurName,
	}
	pbdt, err := ptypes.TimestampProto(oneNews.Time)
	if err != nil {
		return nil, err
	}
	pbe.Datetime = pbdt
	return pbe, nil
}

func injectHeadersIntoMetadata(ctx context.Context, req *http.Request) metadata.MD {
	pairs := make([]string, 0, len(headers))
	for _, h := range headers {
		if v := req.Header.Get(h); len(v) > 0 {
			pairs = append(pairs, h, v)
		}
	}
	return metadata.Pairs(pairs...)
}
