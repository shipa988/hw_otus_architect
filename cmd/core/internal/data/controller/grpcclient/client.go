package grpcclient

import (
	"context"
	"github.com/golang/protobuf/ptypes"
	grpcservice "github.com/shipa988/hw_otus_architect/api"
	"github.com/shipa988/hw_otus_architect/internal/domain/entity"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"strconv"
	"time"
)

var headers = []string{"token"}

const token = "secretfornewsservice"

type GRPCClient struct {
	client      grpcservice.NewsServiceClient
	ctxTimeoutS time.Duration
}

func NewGRPCClient(serverAddr string) (*GRPCClient, error) {
	conn, err := grpc.Dial(serverAddr, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	client := grpcservice.NewNewsServiceClient(conn)
	return &GRPCClient{ctxTimeoutS: time.Second * 10,
		client: client}, nil
}

func (cl *GRPCClient) GetNews(userId uint64) ([]entity.News, error) {
	uid := strconv.FormatUint(userId, 10)
	in := grpcservice.GetNewsRequest{Userid: uid}
	ctx, _ := context.WithTimeout(context.Background(), cl.ctxTimeoutS)

	ctx = metadata.NewOutgoingContext(
		ctx,
		metadata.Pairs("token", token),
	)
	grpcnews, err := cl.client.GetNews(ctx, &in)
	if err != nil {
		return nil, err
	}
	news := grpcnews.GetNews().GetOneNews()
	eNews := []entity.News{}

	for _, oneNews := range news {
		t, e := ptypes.Timestamp(oneNews.Datetime)
		if e != nil {
			return nil, e
		}
		eNews = append(eNews, entity.News{
			Id:            oneNews.Id,
			AuthorId:      oneNews.Author,
			AuthorGen:     oneNews.AuthorGen,
			AuthorName:    oneNews.AuthorName,
			AuthorSurName: oneNews.AuthorSurname,
			Title:         oneNews.Title,
			Time:          t,
			Text:          oneNews.Content,
		})
	}
	return eNews, nil
}
