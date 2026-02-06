package api

import "google.golang.org/grpc"

// StreamServiceServer is the server API for StreamService service.
type StreamServiceServer interface {
	PublishStream(StreamService_PublishStreamServer) error
	mustEmbedUnimplementedStreamServiceServer()
}

// UnimplementedStreamServiceServer can be embedded to have forward compatible implementations.
type UnimplementedStreamServiceServer struct{}

func (UnimplementedStreamServiceServer) PublishStream(StreamService_PublishStreamServer) error {
	return nil
}
func (UnimplementedStreamServiceServer) mustEmbedUnimplementedStreamServiceServer() {}

// StreamService_PublishStreamServer is the server API for the streaming RPC.
type StreamService_PublishStreamServer interface {
	Recv() (*DataPoint, error)
	SendAndClose(*PublishSummary) error
	grpc.ServerStream
}

// DataPoint mocks the protobuf message
type DataPoint struct {
	SourceId  string
	Value     float64
	Timestamp int64
}

// PublishSummary mocks the protobuf response
type PublishSummary struct {
	ProcessedCount int32
	FailedCount    int32
	TotalValue     float64
}

// RegisterStreamServiceServer mocks the registration function
func RegisterStreamServiceServer(s *grpc.Server, srv StreamServiceServer) {}
