package grpc

import (
	"crypto/tls"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

type ChainGrpcImpl struct {
	ClientConn *grpc.ClientConn
}

func NewChainGrpcImpl(addr string, tlsEnabled bool) (*ChainGrpcImpl, error) {
	cc, err := DefaultConn(addr, tlsEnabled)
	if err != nil {
		return nil, err
	}
	return &ChainGrpcImpl{
		ClientConn: cc,
	}, nil
}

func DefaultConn(addr string, useTLS bool) (*grpc.ClientConn, error) {
	var creds credentials.TransportCredentials
	if useTLS {
		creds = credentials.NewTLS(&tls.Config{MinVersion: tls.VersionTLS12})
	} else {
		creds = insecure.NewCredentials()
	}
	return grpc.NewClient(addr, grpc.WithTransportCredentials(creds))
}

// Close closes the underlying grpc connection.
func (c *ChainGrpcImpl) Close() {
	_ = c.ClientConn.Close()
}
