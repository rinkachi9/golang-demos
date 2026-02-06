package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/rinkachi/golang-demos/golang-concurrency-patterns/patterns"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
)

func initTracer() func(context.Context) error {
	ctx := context.Background()

	exporter, err := otlptracegrpc.New(ctx,
		otlptracegrpc.WithEndpoint("localhost:4317"),
		otlptracegrpc.WithInsecure(),
	)
	if err != nil {
		log.Fatalf("failed to create exporter: %v", err)
	}

	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName("concurrency-simulation"),
		),
	)
	if err != nil {
		log.Fatalf("failed to create resource: %v", err)
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
	)
	otel.SetTracerProvider(tp)

	return tp.Shutdown
}

func main() {
	shutdown := initTracer()
	defer shutdown(context.Background())

	tracer := otel.Tracer("simulation-main")
	ctx, span := tracer.Start(context.Background(), "run_simulation")
	defer span.End()

	log.Println("Starting simulation... Open Jaeger to see traces!")

	// Scenario: Image Processing Pipeline
	// 1. Generate Images (Generator)
	// 2. Fetch Metadata Async (Future)
	// 3. Process Images (WorkerPool)
	// 4. Rate Limit Uploads (RateLimiter)
	// 5. Sync at end (Barrier)

	// 1. Generator
	imageIDs := []int{101, 102, 103, 104, 105, 106, 107, 108}
	images := patterns.Generator(ctx, imageIDs...)

	// Rate Limiter: 2 uploads per second
	uploader := patterns.NewRateLimiter(2, 1)
	defer uploader.Stop()

	// Barrier: Wait for 2 phases to complete (Processing, Analytics)
	barrier := patterns.NewBarrier(2)

	// ErrGroup for main phases
	g, gCtx := patterns.WithContext(ctx)

	// Phase 1: Processing Pipeline
	g.Go(gCtx, func() error {
		defer barrier.Await(gCtx, "processing_done")

		// Processing Worker Function
		processor := func(ctx context.Context, id int) (string, error) {
			// Simulate Async Metadata Fetch using Future
			metaFuture := patterns.Async(ctx, func(ctx context.Context) (string, error) {
				time.Sleep(time.Duration(rand.Intn(100)) * time.Millisecond)
				return fmt.Sprintf("Meta-%d", id), nil
			})

			// Simulate Heavy Computation (Resize)
			_, resizeSpan := tracer.Start(ctx, "process_image_resize",  sdktrace.WithAttributes(attribute.Int("image.id", id)))
			time.Sleep(200 * time.Millisecond)
			resizeSpan.End()

			// Get Metadata
			meta, _ := metaFuture.Result(ctx)

			// Upload with Rate Limit
			if err := uploader.Wait(ctx); err != nil {
				return "", err
			}
			
			// Simulate Upload
			_, upSpan := tracer.Start(ctx, "upload_image")
			time.Sleep(50 * time.Millisecond)
			upSpan.End()

			return fmt.Sprintf("Image %d processed with %s", id, meta), nil
		}

		// Run Worker Pool
		// We collect items from generator first to pass slice (simpler for this demo adaptation)
		var workItems []int
		for i := range images {
			workItems = append(workItems, i)
		}

		results := patterns.WorkerPool(ctx, workItems, processor, 3) // 3 Concurrent workers

		for res := range results {
			if res.Err != nil {
				log.Printf("Error: %v", res.Err)
			} else {
				log.Println(res.Value)
			}
		}
		return nil
	})

	// Phase 2: Analytics (Simulated background task)
	g.Go(gCtx, func() error {
		defer barrier.Await(gCtx, "analytics_done")
		// Simulate parallel analytics task
		_, span := tracer.Start(gCtx, "analytics_batch")
		defer span.End()
		time.Sleep(1 * time.Second)
		log.Println("Analytics batch complete")
		return nil
	})

	// Wait for all
	if err := g.Wait(); err != nil {
		log.Printf("Simulation failed: %v", err)
	} else {
		log.Println("Simulation completed successfully.")
	}
}
