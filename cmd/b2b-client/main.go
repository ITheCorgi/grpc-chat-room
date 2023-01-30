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
	"strings"

	"github.com/ITheCorgi/b2b-chat/internal/config"
	chatApi "github.com/ITheCorgi/b2b-chat/pkg/api"
	"github.com/nexidian/gocliselect"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/emptypb"
)

var (
	configPath, user string
)

func init() {
	flag.StringVar(&configPath, "config", "config.yaml", "--config ./file_name.yaml")
	flag.StringVar(&user, "u", "testuser", "--u testuser123")
}

func main() {
	cfg, err := config.New(configPath)
	if err != nil {
		log.Fatalln(err)
	}

	conn, err := grpc.Dial(fmt.Sprintf("%s:%s", cfg.App.Host, cfg.App.Port), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalln(err)
	}

	chatClient := chatApi.NewChatClient(conn)

	err = receive(chatClient)
	if err != nil {
		log.Fatalln(err)
	}

	menu := gocliselect.NewMenu("choose an action")
	menu.AddItem("Create chat group", "1")
	menu.AddItem("Join chat group", "2")
	menu.AddItem("Leave chat group", "3")
	menu.AddItem("Get list of channels", "4")
	menu.AddItem("Send Message", "5")

	for {
		choice := menu.Display()

		switch choice {
		case "1":
			reader := bufio.NewReader(os.Stdin)
			fmt.Print("enter chat group name: ")
			chatName, _ := reader.ReadString('\n')

			_, err := chatClient.CreateGroupChat(context.Background(), &chatApi.GroupChannelNameRequest{GroupChannelName: chatName})
			if err != nil {
				log.Println(err)
			}
		case "2":
			reader := bufio.NewReader(os.Stdin)
			fmt.Print("enter chat group name: ")
			chatName, _ := reader.ReadString('\n')

			_, err := chatClient.JoinGroupChat(context.Background(), &chatApi.GroupChannelNameRequest{GroupChannelName: chatName})
			if err != nil {
				log.Println(err)
			}
		case "3":
			reader := bufio.NewReader(os.Stdin)
			fmt.Print("enter chat group name: ")
			chatName, _ := reader.ReadString('\n')

			_, err := chatClient.LeaveGroupChat(context.Background(), &chatApi.GroupChannelNameRequest{GroupChannelName: chatName})
			if err != nil {
				log.Println(err)
			}
		case "4":
			_, err = chatClient.ListChannels(context.Background(), &emptypb.Empty{})
			if err != nil {
				log.Println(err)
			}
		case "5":
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
}

func receive(client chatApi.ChatClient) error {
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

	stream.CloseSend()
	<-done
	return nil
}
