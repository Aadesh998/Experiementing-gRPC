package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"time"

	pb "chat_app/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	conn, err := grpc.Dial("localhost:5000", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewChatServiceClient(conn)

	stream, err := c.Chat(context.Background())
	if err != nil {
		log.Fatalf("Error on stream: %v", err)
	}

	// Goroutine to receive messages
	go func() {
		for {
			in, err := stream.Recv()
			if err != nil {
				log.Fatalf("Failed to receive a note : %v", err)
			}
			log.Printf("Received: %s: %s", in.User, in.Msg)
		}
	}()

	// Read from stdin and send messages
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Println("Enter your message:")
	for scanner.Scan() {
		msg := &pb.ChatMessage{
			User:      "Client1",
			Msg:       scanner.Text(),
			Timestamp: time.Now().Unix(),
		}
		if err := stream.Send(msg); err != nil {
			log.Fatalf("Failed to send a note: %v", err)
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatalf("Scanner error: %v", err)
	}
}
