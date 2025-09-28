package main

import (
	pb "chat_app/proto"
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"sync"

	"google.golang.org/grpc"
)

// streamPool holds all active streams
type streamPool struct {
	streams []pb.ChatService_ChatServer
	mu      sync.Mutex
}

// Add adds a new stream to the pool
func (p *streamPool) Add(stream pb.ChatService_ChatServer) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.streams = append(p.streams, stream)
}

// Remove removes a stream from the pool
func (p *streamPool) Remove(stream pb.ChatService_ChatServer) {
	p.mu.Lock()
	defer p.mu.Unlock()
	for i, s := range p.streams {
		if s == stream {
			p.streams = append(p.streams[:i], p.streams[i+1:]...)
			return
		}
	}
}

// Broadcast sends a message to all streams in the pool
func (p *streamPool) Broadcast(msg *pb.ChatMessage) {
	p.mu.Lock()
	defer p.mu.Unlock()
	for _, s := range p.streams {
		if err := s.Send(msg); err != nil {
			log.Printf("Failed to send message to a client: %v", err)
		}
	}
}

type server struct {
	pb.UnimplementedChatServiceServer
	pool *streamPool
}

func (s *server) Join(ctx context.Context, req *pb.JoinRequest) (*pb.JoinResponse, error) {
	reply := fmt.Sprintf("Welcome to %s in Chat Room.", req.GetUser())
	return &pb.JoinResponse{
		Success: true,
		Msg:     reply,
	}, nil
}

func (s *server) Chat(stream pb.ChatService_ChatServer) error {
	s.pool.Add(stream)
	defer s.pool.Remove(stream)

	for {
		msg, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			log.Printf("Error receiving message from client: %v", err)
			return err
		}
		log.Printf("Received message: %v", msg)
		s.pool.Broadcast(msg)
	}
}

func main() {
	lis, err := net.Listen("tcp", ":5000")
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}

	grpcServer := grpc.NewServer()
	chatServer := &server{
		pool: &streamPool{},
	}
	pb.RegisterChatServiceServer(grpcServer, chatServer)

	log.Printf("Starting Serving at Port 5000")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to start Server: %v", err)
	}
}