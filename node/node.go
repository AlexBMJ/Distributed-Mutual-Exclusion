package main

import (
	"context"
	pb "example.com/MutualExclusion/mxservice"
	"flag"
	"fmt"
	"github.com/hashicorp/serf/serf"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"log"
	"net"
	"os"
	"time"
)

type MutualEXServer struct {
	pb.UnimplementedMutualEXServer
}

var (
	next         string
	token        = make(chan int, 1)
	currentToken int
)

func main() {
	nodeName := os.Getenv("NODE_NAME")
	caddr := os.Getenv("CLUSTER_ADDRESS")
	flag.Parse()
	cluster, clustErr := SetupCluster(nodeName, caddr)
	SetupGrpc(cluster, caddr)
	defer cluster.Leave()
	if clustErr != nil {
		log.Fatal(clustErr)
	}

	server := grpc.NewServer()
	pb.RegisterMutualEXServer(server, &MutualEXServer{})

	lis, servErr := net.Listen("tcp", ":8080")
	if servErr != nil {
		log.Fatalf("failed to listen: %v", servErr)
	}

	log.Printf("server listening at %v", lis.Addr())
	go func() {
		if err := server.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()

	var ctx = context.Background()
	for {
		WriteToLog(nodeName)
		var conn, grpcErr = grpc.Dial(next+":8080", grpc.WithInsecure(), grpc.WithBlock())
		if grpcErr != nil {
			log.Fatalf("did not connect: %v", grpcErr)
		}
		var client = pb.NewMutualEXClient(conn)
		_, tokErr := client.PassToken(ctx, &pb.Token{Token: int32(currentToken)})
		if tokErr != nil {
			log.Fatalf("failed to pass token: %v", tokErr)
		}
		time.Sleep(2000 * time.Millisecond)
	}
}

func SetupCluster(nodeName string, clusterAddr string) (*serf.Serf, error) {
	conf := serf.DefaultConfig()
	conf.Init()
	conf.NodeName = nodeName

	cluster, serfErr := serf.Create(conf)
	if serfErr != nil {
		return nil, errors.Wrap(serfErr, "Couldn't create cluster")
	}

	_, joinErr := cluster.Join([]string{clusterAddr}, true)
	if joinErr != nil {
		log.Printf("Couldn't join cluster, starting own: %v\n", joinErr)
	}

	return cluster, nil
}

func SetupGrpc(cluster *serf.Serf, addr string) {
	if len(cluster.Members()) != 1 {
		var ctx = context.Background()
		var conn, err2 = grpc.Dial(addr+":8080", grpc.WithInsecure(), grpc.WithBlock())
		if err2 != nil {
			log.Fatalf("did not connect: %v", err2)
		}

		var client = pb.NewMutualEXClient(conn)
		first, err3 := client.RequestJoin(ctx, &pb.JoinRequest{SenderAddr: cluster.LocalMember().Addr.String()})
		if err3 != nil {
			log.Fatalf("could not request to join: %v", err3)
		}
		next = first.SenderAddr
	} else {
		next = cluster.LocalMember().Addr.String()
	}

	if len(cluster.Members()) == 1 {
		token <- 1
	}
}

func (s *MutualEXServer) RequestJoin(_ context.Context, req *pb.JoinRequest) (*pb.JoinRequest, error) {
	joinRequest := &pb.JoinRequest{SenderAddr: next}
	next = req.SenderAddr
	return joinRequest, nil
}

func WriteToLog(text string) {
	now := time.Now()
	currentToken = <-token
	log.Printf("[%s]: %s token: %d\n", now.Format("2006-01-02 15:04:05.000000"), text, currentToken)

	file, ferr := os.OpenFile("/go/src/app/log/log.txt", os.O_APPEND|os.O_WRONLY, 0666)
	if ferr != nil {
		fmt.Println(ferr)
		return
	}
	_, logErr := fmt.Fprintf(file, "[%s]: %s token: %d\n", now.Format("2006-01-02 15:04:05.000000"), text, currentToken)
	if logErr != nil {
		fmt.Println(logErr)
		return
	}
}

func (s *MutualEXServer) PassToken(_ context.Context, t *pb.Token) (*pb.Empty, error) {
	currentToken = int(t.Token)
	token <- currentToken + 1
	return &pb.Empty{}, nil
}
