package rack_mocker

import (
	"flag"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"log"
	"net"
	"os"
	"rack-mocker/file"
	"rack-mocker/server"
)

var listenTo = flag.String("port", ":39253", "specify the address file mocker listens to")
var rootDirectory = flag.String("root", "rack", "specify the root directory managed by mocker")
var logFile = flag.String("log", "", "specify where log dump to, choose stdout if empty")

func main() {
	flag.Parse()

	log.SetPrefix("rack-mocker-server")

	listen, err := net.Listen("tcp", *listenTo)

	if len(*logFile) > 0 {
		if f, err := os.Open(*logFile); err == nil {
			log.SetOutput(f)
		} else {
			log.Fatalf("failed to open file: %s", *logFile)
		}
	}

	if err != nil {
		log.Panicln("failed to init server")
	}

	grpcServer := grpc.NewServer()
	file.RegisterFileServer(grpcServer, server.NewService(*rootDirectory))
	reflection.Register(grpcServer)

	if err := grpcServer.Serve(listen); err != nil {
		log.Fatalf("failed to server: %v", err)
	}
}
