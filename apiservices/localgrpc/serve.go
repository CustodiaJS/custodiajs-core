package localgrpc

import (
	"log"

	"github.com/CustodiaJS/custodiajs-core/apiservices/localgrpc/localgrpcservice"
	"github.com/CustodiaJS/custodiajs-core/localgrpcproto"
)

func (o *HostCliService) Serve(closeSignal chan struct{}) error {
	// Das CLI gRPC Serverobjekt wird erstellt
	localgrpc := localgrpcservice.NewCliGrpcServer(o.core)
	localgrpcproto.RegisterLocalhostAPIServiceServer(o.grpcServer, localgrpc)

	// Der grpc Server wird gestartet
	if err := o.grpcServer.Serve(o.netListner); err != nil {
		log.Fatalf("Fehler beim Starten des gRPC-Servers: %v", err)
	}

	// Der Vorgagn wurde ohne Fehler durchgeführt
	return nil
}
