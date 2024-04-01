package webgrpc

import (
	"fmt"
	"log"
	"net"
	"vnh1/grpc/publicgrpc"
)

func (o *WebGrpcService) Serve(closeSignal chan struct{}) error {
	// Das gRPC Objekt wird erstellt
	publicgrpc.RegisterRPCServiceServer(o.serverObj, &GrpcServer{core: o.core})

	// Starte den gRPC-Server auf Port 50051
	lis, err := net.Listen("tcp", fmt.Sprintf("%s:%d", o.localAddress.LocalIP, o.localAddress.LocalPort))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	if err := o.serverObj.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}

	// Der Vorgagn wurde ohne Fehler durchgeführt
	return nil
}
