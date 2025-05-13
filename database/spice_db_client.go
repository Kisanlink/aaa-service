package database

import (
	"log"
	"os"

	"github.com/authzed/authzed-go/v1"
	"github.com/authzed/grpcutil"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func SpiceDB() (*authzed.Client,error) {
	client, err := authzed.NewClient(
		os.Getenv("SPICE_DB_ENDPOINT"),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpcutil.WithInsecureBearerToken(os.Getenv("SPICE_DB_TOKEN")),
	)
	if err != nil {
		log.Fatalf("Failed to connect to SpiceDB: %v", err)
		return client, err
	}
	return client,err
}

