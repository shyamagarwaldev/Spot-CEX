package grpc

import (
	"context"
	"fmt"

	pb "github.com/shyamagarwaldev/Spot-CEX/matching-engine/genproto/matching"
	"github.com/shyamagarwaldev/Spot-CEX/matching-engine/internal/engine"
)

type MatchingEngineServer struct {
	pb.UnimplementedMatchingEngineServiceServer

	server *Server
}

func NewMatchingEngineServer(server *Server) *MatchingEngineServer {

	return &MatchingEngineServer{
		server: server,
	}

}

func (s *MatchingEngineServer) CreateOrder(
	ctx context.Context,
	req *pb.CreateOrderRequest,
) (*pb.CreateOrderResponse, error) {

	if req.Type == pb.OrderType_LIMIT && req.Price == nil {
		return nil, fmt.Errorf("price is required")
	}

	if req.Quantity <= 0 {
		return nil, fmt.Errorf("quantity is required")
	}

	if req.Price == nil {
		price := int64(0)
		req.Price = &price
	}

	responseChannel := make(chan engine.CommandResult, 1)

	// this is a shared resource
	seq := s.server.sequencer.Next()

	cmd := engine.Command{
		Sequence: seq,
		Type:     engine.CreateOrder,
		Response: responseChannel,
		Payload: engine.CreateOrderCommand{
			OrderID:  req.OrderId,
			UserID:   req.UserId,
			Symbol:   req.Symbol,
			Side:     engine.Side(req.Side),
			Type:     engine.OrderType(req.Type),
			Price:    *req.Price,
			Quantity: req.Quantity,
			Asset:    req.Asset,
		},
	}

	// this is a shared resource
	if err := s.server.wal.Append(cmd); err != nil {
		return nil, err
	}
	s.server.dispatcher.Dispatch(cmd)

	select {
	case result := <-responseChannel:
		if result.Err != nil {
			return nil, result.Err
		}
		fill := result.Response.(engine.Fill)
		return &pb.CreateOrderResponse{
			RequestId:         req.RequestId,
			UserId:            fill.UserID,
			OrderId:           fill.OrderID,
			FilledQuantity:    fill.FilledQuantity,
			RemainingQuantity: fill.RemainingQuantity,
			TotalPrice:        fill.TotalPrice,
			Side:              pb.Side(fill.Side),
		}, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func (s *MatchingEngineServer) CancelOrder(
	ctx context.Context,
	req *pb.CancelOrderRequest,
) (*pb.CancelOrderResponse, error) {

	responseChannel := make(chan engine.CommandResult, 1)

	// this is a shared resource
	seq := s.server.sequencer.Next()

	cmd := engine.Command{
		Sequence: seq,
		Type:     engine.CancelOrder,
		Response: responseChannel,
		Payload: engine.CancelOrderCommand{
			OrderID: req.OrderId,
			UserID:  req.UserId,
			Symbol:  req.Symbol,
		},
	}

	// this is a shared resource
	if err := s.server.wal.Append(cmd); err != nil {
		return nil, err
	}
	s.server.dispatcher.Dispatch(cmd)

	select {
	case result := <-responseChannel:
		if result.Err != nil {
			return nil, result.Err
		}
		return result.Response.(*pb.CancelOrderResponse), nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func (s *MatchingEngineServer) GetTicker(
	ctx context.Context,
	req *pb.GetTickerRequest,
) (*pb.GetTickerResponse, error) {

	responseChannel := make(chan engine.CommandResult, 1)

	cmd := engine.Command{
		Type:     engine.GetTicker,
		Response: responseChannel,
		Payload: engine.GetTickerCommand{
			Symbol: req.Symbol,
		},
	}

	s.server.dispatcher.Dispatch(cmd)

	select {
	case result := <-responseChannel:
		if result.Err != nil {
			return nil, result.Err
		}
		return result.Response.(*pb.GetTickerResponse), nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func (s *MatchingEngineServer) GetDepth(
	ctx context.Context,
	req *pb.GetDepthRequest,
) (*pb.GetDepthResponse, error) {

	responseChannel := make(chan engine.CommandResult, 1)

	cmd := engine.Command{
		Type:     engine.GetDepth,
		Response: responseChannel,
		Payload: engine.GetDepthCommand{
			Limit:  req.Limit,
			Symbol: req.Symbol,
		},
	}

	s.server.dispatcher.Dispatch(cmd)

	select {
	case result := <-responseChannel:
		if result.Err != nil {
			return nil, result.Err
		}
		return result.Response.(*pb.GetDepthResponse), nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}
