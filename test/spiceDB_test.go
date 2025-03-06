package test

// import (
// 	"context"
// 	"testing"

// 	"github.com/Kisanlink/aaa-service/client"
// 	pb "github.com/authzed/authzed-go/proto/authzed/api/v1"
// 	"github.com/stretchr/testify/assert"
// 	"github.com/stretchr/testify/mock"
// )

// // MockSpiceDBClient mocks the SpiceDB client
// type MockSpiceDBClient struct {
// 	mock.Mock
// }

// func (m *MockSpiceDBClient) WriteSchema(ctx context.Context, req *pb.WriteSchemaRequest) (*pb.WriteSchemaResponse, error) {
// 	args := m.Called(ctx, req)
// 	resp, ok := args.Get(0).(*pb.WriteSchemaResponse)
// 	if !ok {
// 		return nil, args.Error(1)
// 	}
// 	return resp, args.Error(1)
// }

// func (m *MockSpiceDBClient) ReadSchema(ctx context.Context, req *pb.ReadSchemaRequest) (*pb.ReadSchemaResponse, error) {
// 	args := m.Called(ctx, req)
// 	resp, ok := args.Get(0).(*pb.ReadSchemaResponse)
// 	if !ok {
// 		return nil, args.Error(1)
// 	}
// 	return resp, args.Error(1)
// }

// // TestWriteSchema tests the WriteSchema function
// func TestWriteSchema(t *testing.T) {
// 	mockSpiceDB := new(MockSpiceDBClient)

// 	// Expected response
// 	mockSpiceDB.On("WriteSchema", mock.Anything, mock.Anything).
// 		Return(&pb.WriteSchemaResponse{}, nil)

// 	// Call function with mock client
// 	res, err := client.WriteSchema(mockSpiceDB)

// 	// Assertions
// 	assert.NoError(t, err)
// 	assert.NotNil(t, res)
// 	mockSpiceDB.AssertExpectations(t)
// }

// // TestReadSchema tests the ReadSchema function
// func TestReadSchema(t *testing.T) {
// 	mockSpiceDB := new(MockSpiceDBClient)
// 	expectedSchema := &pb.ReadSchemaResponse{SchemaText: "definition user {}"}

// 	mockSpiceDB.On("ReadSchema", mock.Anything, mock.Anything).
// 		Return(expectedSchema, nil)

// 	// Call function with mock client
// 	res, err := client.ReadSchema(mockSpiceDB)

// 	// Assertions
// 	assert.NoError(t, err)
// 	assert.Contains(t, res.SchemaText, "definition user {}") // Looser check
// 	mockSpiceDB.AssertExpectations(t)
// }

// // TestUpdateSchema tests the UpdateSchema function
// func TestUpdateSchema(t *testing.T) {
// 	mockSpiceDB := new(MockSpiceDBClient)
// 	existingSchema := &pb.ReadSchemaResponse{SchemaText: "definition user {}"}

// 	mockSpiceDB.On("ReadSchema", mock.Anything, mock.Anything).
// 		Return(existingSchema, nil)
// 	mockSpiceDB.On("WriteSchema", mock.Anything, mock.Anything).
// 		Return(&pb.WriteSchemaResponse{}, nil)

// 	roles := []string{"ceo", "director"}
// 	permissions := []string{"view", "edit"}

// 	// Call function with role and permissions
// 	res, err := client.UpdateSchema(mockSpiceDB, roles, permissions)

// 	// Assertions
// 	assert.NoError(t, err)
// 	assert.NotNil(t, res)
// 	mockSpiceDB.AssertExpectations(t)
// }
