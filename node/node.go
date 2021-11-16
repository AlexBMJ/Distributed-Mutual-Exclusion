package main

import (
	"context"
	pb "example.com/MutualExclusion/mxservice"
	"flag"
	"github.com/hashicorp/serf/serf"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"log"
	"net"
	"os"
	"strconv"
	"time"
)

type MutualEXServer struct {
	pb.UnimplementedMutualEXServer
}

var (
	lamportClock int64 = 0
)

func main() {
	nodeName := os.Getenv("NODE_NAME")
	aaddr := os.Getenv("ADVERTISE_ADDRESS")
	caddr := os.Getenv("CLUSTER_ADDRESS")
	flag.Parse()
	cluster, err := setupCluster(nodeName, aaddr, caddr)
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
		for i := 0; i < len(cluster.Members()); i++ {
			if cluster.Members()[i].Name == nodeName {
				continue
			}
			var ctx = context.Background()
			maddr := cluster.Members()[i].Addr.String()
			var conn, err = grpc.Dial(maddr+":8080", grpc.WithInsecure(), grpc.WithBlock())
			if err != nil {
				log.Fatalf("did not connect: %v", err)
			}

			var client = pb.NewMutualEXClient(conn)
			var _, joinErr = client.WriteToLog(ctx, &pb.Message{Timestamp: lamportClock})
			if joinErr != nil {
				log.Fatalf("could not join chittychat: %v", joinErr)
			}

			conn.Close()
			time.Sleep(2000)
		}
	}
}

func setupCluster(nodeName string, advertiseAddr string, clusterAddr string) (*serf.Serf, error) {
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

	return cluster, nil
}

func (s *MutualEXServer) WriteToLog(ctx context.Context, message *pb.Message) (*pb.Empty, error) {
	log.Println(strconv.Itoa(int(message.Timestamp)))
	return &pb.Empty{}, nil
}
