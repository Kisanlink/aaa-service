package database

import (
	"log"

	"github.com/authzed/authzed-go/v1"
	"github.com/authzed/grpcutil"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// const schema = `
// definition user {}

// definition role {
//     relation super_admin: user
//     relation admin: user
//     relation ceo: user
//     relation director: user
//     relation shareholder: user
//     relation farmer: user
//     relation kisansathi: user
//     relation collaborator: user
//     relation staff: user
//     relation system_staff: user
//     relation consumer: user
//     relation student: user
//     relation company_manager: user
//     relation company_owner: user
// }
// definition permission {
//     relation all_all: role#super_admin
//     relation all: role#admin
//     relation fpo: role#ceo | role#director | role#shareholder | role#staff | role#system_staff
//     relation kisanlink: role#farmer | role#kisansathi | role#collaborator
//     relation amrti: role#consumer
//     relation asa: role#student | role#company_manager | role#company_owner
// }`



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

