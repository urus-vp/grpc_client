package main

import (
	"context"
	"time"

	"grpc_client/internal/config"
	"grpc_client/internal/database"

	log "github.com/sirupsen/logrus"
	grpc_proto "github.com/urus-vp/grpc_proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	conf := config.Get()

	connector, err := database.NewPostgresConnector(conf.PostgresDSN)
	if err != nil {
		log.Fatalln("cannot create connector:", err)
	}

	conn, err := grpc.Dial(conf.Addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := grpc_proto.NewGreeterClient(conn)

	go func() {
		log.Println("go func()")

		rshListener, err := connector.ListenForUpdates("RigStateHours")
		if err != nil {
			log.Fatalln("something wrong with wells listener", err)
		}

		for {
			select {
			case <-rshListener:
				fbp := grpc_proto.FirebasePayload{
					Collection: "RigStateHours: Collection",
					Document:   "RigStateHours: Document",
					Payload:    "RigStateHours: Payload"}
				log.Println("Send RigStateHours:", fbp)

				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				defer cancel()
				r, err := c.WriteFirebasePayload(ctx, &fbp)
				if err != nil {
					log.Fatalf("Could not call WriteFirebasePayload: %v", err)
				}
				log.Printf("WriteFirebasePayload result: %s", r.GetMessage())

				continue
			case <-time.After(time.Second * 5):
				log.Println("Still listening for updates, nothing yet")
				continue
			}
		}
	}()

	log.Println("Starting Client")

	for {
		time.Sleep(time.Microsecond * 5)
	}
}
