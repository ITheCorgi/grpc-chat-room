package main

import (
	"bufio"
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	chatApi "github.com/ITheCorgi/grpc-chat-room/pkg/api"
	"github.com/manifoldco/promptui"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/emptypb"
)

var (
	port, user string
)

func init() {
	flag.StringVar(&port, "port", "8270", "--port 6666")
	flag.StringVar(&user, "user", "testuser", "--user testuser123")
}

func main() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	ctx, cancelFunc := context.WithCancel(context.Background())

	conn, err := grpc.Dial(fmt.Sprintf(":%s", port), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalln(err)
	}

	log.Println("grpc conn established")

	chatClient := chatApi.NewChatClient(conn)

	go func() {
		err = receive(ctx, chatClient)
		if err != nil {
			log.Fatalln(err)
		}
	}()

	md := metadata.New(map[string]string{"authorization": user})
	ctx = metadata.NewOutgoingContext(ctx, md)

	menu := promptui.Select{
		Label: "choose an action",
		Items: []string{"Create chat group", "Join chat group", "Join chat group", "Leave chat group", "Get list of channels", "Send Message"},
	}

	for {
		idx, _, err := menu.Run()

		switch idx {
		case 0:
			reader := bufio.NewReader(os.Stdin)
			fmt.Print("enter chat group name: ")
			chatName, _ := reader.ReadString('\n')

			_, err := chatClient.CreateGroupChat(ctx, &chatApi.GroupChannelNameRequest{GroupChannelName: chatName})
			if err != nil {
				log.Println(err)
			}
		case 1:
			reader := bufio.NewReader(os.Stdin)
			fmt.Print("enter chat group name: ")
			chatName, _ := reader.ReadString('\n')

			_, err := chatClient.JoinGroupChat(ctx, &chatApi.GroupChannelNameRequest{GroupChannelName: chatName})
			if err != nil {
				log.Println(err)
			}
		case 2:
			reader := bufio.NewReader(os.Stdin)
			fmt.Print("enter chat group name: ")
			chatName, _ := reader.ReadString('\n')

			_, err := chatClient.LeaveGroupChat(ctx, &chatApi.GroupChannelNameRequest{GroupChannelName: chatName})
			if err != nil {
				log.Println(err)
			}
		case 3:
			_, err = chatClient.ListChannels(ctx, &emptypb.Empty{})
			if err != nil {
				log.Println(err)
			}
		case 4:
			reader := bufio.NewReader(os.Stdin)
			fmt.Print("enter destination name, to user or group chat (1 or 2) and message. Each 3 must be separated ',': ")
			input, _ := reader.ReadString('\n')

			el := strings.Split(input, ",")

			msg := &chatApi.ChatMessage{Message: el[2]}
			if el[1] == "1" {
				msg.Destination = &chatApi.ChatMessage_Username{Username: el[0]}
			} else {
				msg.Destination = &chatApi.ChatMessage_GroupChannelName{GroupChannelName: el[0]}
			}

			_, err = chatClient.SendMessage(context.Background(), msg)
			if err != nil {
				log.Println(err)
			}
		}
	}

	<-sigChan
	log.Println("start graceful shutdown, caught sig")
	cancelFunc()
}

func receive(ctx context.Context, client chatApi.ChatClient) error {
	stream, err := client.Connect(context.Background(), &chatApi.ConnectRequest{
		Username: user,
	})
	if err != nil {
		return err
	}

	done := make(chan struct{})
	go func(stream chatApi.Chat_ConnectClient) {
		for {
			in, err := stream.Recv()
			if err != nil {
				if errors.Is(err, io.EOF) {
					close(done)
					return
				}

				log.Fatalf("failed to recieve a message: %v", err)
			}

			log.Printf("got message %s", in.Message)
		}
	}(stream)

	<-done
	return nil
}
