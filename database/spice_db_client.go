package database

import (
	"log"

	"github.com/authzed/authzed-go/v1"
	"github.com/authzed/grpcutil"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func SpiceDB() (*authzed.Client,error) {
	client, err := authzed.NewClient(
		"localhost:50051",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpcutil.WithInsecureBearerToken("tc_alfiya96960_tutorial_default_token"),
	)
	if err != nil {
		log.Fatalf("Failed to connect to SpiceDB: %v", err)
		return client, err
	}
	return client,err
}

