package grpc

import (
	"net"

	pbAccount "github.com/shyamagarwaldev/Spot-CEX/matching-engine/genproto/account"
	pbMatching "github.com/shyamagarwaldev/Spot-CEX/matching-engine/genproto/matching"

	"github.com/shyamagarwaldev/Spot-CEX/matching-engine/internal/account"
	"github.com/shyamagarwaldev/Spot-CEX/matching-engine/internal/engine"
	"github.com/shyamagarwaldev/Spot-CEX/matching-engine/internal/wal"

	"google.golang.org/grpc"
)

type Server struct {
	sequencer  ISequencer
	wal        wal.IWAL
	dispatcher engine.IDispatcher
	account    account.IAccountService

	grpcServer *grpc.Server
}

func NewServer(
	sequencer ISequencer,
	wal wal.IWAL,
	dispatcher engine.IDispatcher,
	account account.IAccountService,
) *Server {
	return &Server{
		sequencer:  sequencer,
		wal:        wal,
		dispatcher: dispatcher,
		account:    account,
		grpcServer: grpc.NewServer(),
	}
}

func (s *Server) Register() {

	pbMatching.RegisterMatchingEngineServiceServer(
		s.grpcServer,
		NewMatchingEngineServer(s),
	)
	pbAccount.RegisterAccountServiceServer(
		s.grpcServer,
		NewAccountServiceServer(s),
	)

}

func (s *Server) Serve(listener net.Listener) error {

	s.Register()
	return s.grpcServer.Serve(listener)

}
