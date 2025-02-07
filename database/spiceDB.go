package database

import (
	"fmt"
	"os"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"log"

	"github.com/authzed/authzed-go/v1"
	"github.com/authzed/grpcutil"
)

var SpiceDBClient *authzed.Client

func SpiceDB() error {
	var err error
	SpiceDBClient, err = authzed.NewClient(
		"localhost:50051",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpcutil.WithInsecureBearerToken(os.Getenv("SPICE_DB_TOKEN")),
	)
	if err != nil {
		log.Fatalf("Failed to connect to SpiceDB: %v", err)
		return err
	}
	fmt.Println("Spice Database connected")
	return nil
}
