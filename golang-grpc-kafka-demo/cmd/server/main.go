package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"time"

	"github.com/segmentio/kafka-go"
	"google.golang.org/grpc"

	// In a real scenario, we would generate the pb code.
	// Since I cannot run protoc here effectively without the toolchain installed on the machine,
	// I will simulate the interface or structure it such that it compiles if generated.
	// However, to make this runnable/compilable as a demo without protoc:
	// I will write the implementation assuming the 'api' package exists.
	// Since I can't generate it, I will create a dummy 'api' package to satisfy the compiler for this demo.
	pb "github.com/rinkachi/golang-demos/golang-grpc-kafka-demo/pkg/api/v1"
)

// Server implements the gRPC service
type server struct {
	pb.UnimplementedStreamServiceServer
	kafkaWriter *kafka.Writer
}

func (s *server) PublishStream(stream pb.StreamService_PublishStreamServer) error {
	var summary pb.PublishSummary
	startTime := time.Now()

	for {
		point, err := stream.Recv()
		if err == io.EOF {
			return stream.SendAndClose(&summary)
		}
		if err != nil {
			return err
		}

		// Process logic: Validate and Publish to Kafka
		msgValue := fmt.Sprintf(`{"source": "%s", "value": %f, "ts": %d}`, point.SourceId, point.Value, point.Timestamp)

		msg := kafka.Message{
			Key:   []byte(point.SourceId),
			Value: []byte(msgValue),
		}

		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		err = s.kafkaWriter.WriteMessages(ctx, msg)
		cancel()

		if err != nil {
			log.Printf("Failed to write to kafka: %v", err)
			summary.FailedCount++
		} else {
			summary.ProcessedCount++
			summary.TotalValue += point.Value
		}
	}

	_ = startTime // usage

	return nil
}

func main() {
	// Setup Kafka Writer
	writer := &kafka.Writer{
		Addr:     kafka.TCP("localhost:9092"),
		Topic:    "sensor-data",
		Balancer: &kafka.LeastBytes{},
	}
	defer writer.Close()

	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterStreamServiceServer(s, &server{kafkaWriter: writer})

	log.Printf("gRPC server listening on :50051")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
