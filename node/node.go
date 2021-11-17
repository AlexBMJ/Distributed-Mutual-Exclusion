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
	aaddr := os.Getenv("ADVERTISE_ADDRESS")
	caddr := os.Getenv("CLUSTER_ADDRESS")
	flag.Parse()
	cluster, err := SetupCluster(nodeName, aaddr, caddr)
	defer cluster.Leave()
	if err != nil {
		log.Fatal(err)
	}

	server := grpc.NewServer()
	pb.RegisterMutualEXServer(server, &MutualEXServer{})

	lis, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	log.Printf("server listening at %v", lis.Addr())
	go func() {
		if err := server.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()

	for {
		var ctx = context.Background()
		WriteToLog(ctx, &pb.Message{Text: nodeName})
		time.Sleep(2000 * time.Millisecond)
	}
}

func SetupCluster(nodeName string, advertiseAddr string, clusterAddr string) (*serf.Serf, error) {
	conf := serf.DefaultConfig()
	conf.Init()
	conf.NodeName = nodeName
	conf.MemberlistConfig.AdvertiseAddr = advertiseAddr

	cluster, err := serf.Create(conf)
	if err != nil {
		return nil, errors.Wrap(err, "Couldn't create cluster")
	}

	_, err = cluster.Join([]string{clusterAddr}, true)
	if err != nil {
		log.Printf("Couldn't join cluster, starting own: %v\n", err)
	}

	if len(cluster.Members()) != 1 {
		var ctx = context.Background()
		var conn, err2 = grpc.Dial(clusterAddr+":8080", grpc.WithInsecure(), grpc.WithBlock())
		if err2 != nil {
			log.Fatalf("did not connect: %v", err)
		}

		var client = pb.NewMutualEXClient(conn)
		first, err3 := client.RequestJoin(ctx, &pb.JoinRequest{SenderAddr: advertiseAddr})
		if err3 != nil {
			log.Fatalf("could not request to join: %v", err)
		}
		next = first.SenderAddr
	} else {
		next = advertiseAddr
	}

	if len(cluster.Members()) == 1 {
		token <- 1
	}

	return cluster, nil
}

func (s *MutualEXServer) RequestJoin(ctx context.Context, req *pb.JoinRequest) (*pb.JoinRequest, error) {
	joinRequest := &pb.JoinRequest{SenderAddr: next}
	next = req.SenderAddr
	return joinRequest, nil
}

func (s *MutualEXServer) WriteToLog(ctx context.Context, message *pb.Message) (*pb.Empty, error) {
	WriteToLog(ctx, message)
	return &pb.Empty{}, nil
}

func WriteToLog(ctx context.Context, message *pb.Message) {
	now := time.Now()
	currentToken = <-token
	log.Printf("[%s]: %s\n", now.Format("2006-01-02 15:04:05.000000"), message.Text)

	file, ferr := os.OpenFile("/go/src/app/log/log.txt", os.O_APPEND|os.O_WRONLY, 0666)
	if ferr != nil {
		fmt.Println(ferr)
		return
	}
	fmt.Fprintf(file, "[%s]: %s\n", now.Format("2006-01-02 15:04:05.000000"), message.Text)

	var conn, err = grpc.Dial(next+":8080", grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}

	var client = pb.NewMutualEXClient(conn)
	_, err2 := client.PassToken(ctx, &pb.Token{Token: int32(currentToken)})
	if err2 != nil {
		return
	}
}

func (s *MutualEXServer) PassToken(ctx context.Context, t *pb.Token) (*pb.Empty, error) {
	log.Println(t)
	currentToken = int(t.Token)
	token <- currentToken + 1
	return &pb.Empty{}, nil
}
