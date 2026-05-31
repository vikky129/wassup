package proto

import (
	"context"
	"fmt"
	"personal/wassup/spec"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type GrpcClient struct {
	client *grpc.ClientConn
}

func NewGrpcClient(conn *grpc.ClientConn) *GrpcClient {
	return &GrpcClient{client: conn}
}

func SendMessageViaGrpc(address, receiverID string, message *spec.Message) error {
	grpcConn, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return err
	}
	defer grpcConn.Close()

	c := NewMessengerClient(grpcConn)
	_, err = c.SendMessage(context.Background(), &AddMessageRequest{
		ReceiverId: receiverID,
		Message: &AddMessageInfo{
			XId:            message.ID,
			ConversationId: message.ConversationID,
			Text:           message.Text,
			Type:           message.Type,
			SenderId:       message.SenderID,
			CreatedAt:      message.CreatedAt,
			UpdatedAt:      message.UpdatedAt,
			MediaUrl:       message.MediaURL,
			Status:         string(message.Status),
		},
	})

	if err != nil {
		fmt.Println("Error sending message via gRPC:", err)
	}
	return err
}

func SendMessageDeliveredViaGrpc(address, receiverID, messageID, status string) error {
	grpcConn, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return err
	}
	defer grpcConn.Close()

	c := NewMessengerClient(grpcConn)
	_, err = c.SendMessageDelivered(context.Background(), &MessageDeliveredRequest{
		ReceiverId: receiverID,
		MessageId:  messageID,
		Status:     status,
	})

	if err != nil {
		fmt.Println("Error sending message delivered via gRPC:", err)
	}
	return err
}
