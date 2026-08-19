package grpc

import (
	"context"

	pb "github.com/shyamagarwaldev/Spot-CEX/matching-engine/genproto/account"
)

type AccountServiceServer struct {
	pb.UnimplementedAccountServiceServer

	server *Server
}

func NewAccountServiceServer(server *Server) *AccountServiceServer {

	return &AccountServiceServer{
		server: server,
	}

}

func (s *AccountServiceServer) GetBalance(
	ctx context.Context,
	req *pb.GetBalanceRequest,
) (*pb.GetBalanceResponse, error) {

	balance, err := s.server.account.GetBalance(req.UserId, req.Asset)

	if err != nil {
		return nil, err
	}
	return &pb.GetBalanceResponse{
		RequestId: req.RequestId,
		UserId:    req.UserId,
		Asset:     req.Asset,
		Available: balance.Available,
		Reserved:  balance.Reserved,
	}, nil
}
