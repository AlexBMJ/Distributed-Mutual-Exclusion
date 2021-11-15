package main

import (
	"flag"
	"github.com/hashicorp/serf/serf"
	"github.com/pkg/errors"
	"log"
)

func main() {
	aaddr := flag.String("aaddr","","")
	caddr := flag.String("caddr","","")
	flag.Parse()
	cluster, err := setupCluster(*aaddr, *caddr)
	defer cluster.Leave()
	if err != nil {
		log.Fatal(err)
	}
	for{}

}

func setupCluster(advertiseAddr string, clusterAddr string) (*serf.Serf, error) {
	conf := serf.DefaultConfig()
	conf.Init()
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

func getOtherMembers(cluster *serf.Serf) []serf.Member {
	members := cluster.Members()
	for i := 0; i < len(members); {
		if members[i].Name == cluster.LocalMember().Name || members[i].Status != serf.StatusAlive {
			if i < len(members)-1 {
				members = append(members[:i], members[i + 1:]...)
			} else {
				members = members[:i]
			}
		} else {
			i++
		}
	}
	return members
}