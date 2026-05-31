package proto

import (
	context "context"
	"fmt"
	"log"
	"net"
	"personal/wassup/redis"
	"personal/wassup/spec"
	"personal/wassup/ws"
	"strings"

	grpc "google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

type GrpcHandler struct {
	UnimplementedMessengerServer
	wsHandler    *ws.WebSocketHandler
	redisHandler *redis.CacheHandler
}

func NewGrpcHandler(wsHandler *ws.WebSocketHandler, redisHandler *redis.CacheHandler) *GrpcHandler {
	return &GrpcHandler{
		wsHandler:    wsHandler,
		redisHandler: redisHandler,
	}
}

func (grpcH *GrpcHandler) SendMessage(ctx context.Context, req *AddMessageRequest) (*emptypb.Empty, error) {
	fmt.Println("Received gRPC request to send message:", req)
	sentTo := req.ReceiverId

	wsClients, err := grpcH.wsHandler.GetWSClients(sentTo)
	if err != nil {
		log.Printf("Error getting WebSocket clients for user %s: %v", sentTo, err)
		return nil, err
	}

	for _, client := range wsClients {
		fmt.Printf("Sending message to user %s\n", sentTo)
		resp := client.SendMessageOnChannel(req.Message)
		switch v := resp.(type) {
		case error:
			log.Printf("Error sending message to user %s: %v", sentTo, v)
		case string:
			log.Printf("Response from sending message to user %s: %s", sentTo, v)
			senderLive, _ := grpcH.redisHandler.CheckKeyPrefixExists(req.Message.SenderId)
			if senderLive {
				containerAddress, err := grpcH.redisHandler.GetContainerDetails(ctx, req.Message.SenderId)
				if err != nil {
					fmt.Println("Error getting container details:", err)
					return nil, err
				}

				for _, addressMap := range containerAddress {
					for key, address := range addressMap {
						if strings.HasSuffix(key, "name") {
							continue
						}
						go SendMessageDeliveredViaGrpc(address, req.Message.SenderId, req.Message.XId, string(spec.StatusDelivered))
					}
				}
			}
		default:
			log.Printf("Unexpected response type when sending message to user %s: %v", sentTo, v)

		}
	}

	return &emptypb.Empty{}, nil
}

func (grpcH *GrpcHandler) SendMessageDelivered(ctx context.Context, req *MessageDeliveredRequest) (*emptypb.Empty, error) {
	wsClients, err := grpcH.wsHandler.GetWSClients(req.ReceiverId)
	if err != nil {
		log.Printf("Error getting WebSocket clients for user %s: %v", req.ReceiverId, err)
		return nil, err
	}

	for _, client := range wsClients {
		go client.SendMessageOnChannel(req)
	}

	return &emptypb.Empty{}, nil
}

func StartGrpcServer(wsHandler *ws.WebSocketHandler, redisHandler *redis.CacheHandler) {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()

	RegisterMessengerServer(s, NewGrpcHandler(wsHandler, redisHandler))
	log.Printf("server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
