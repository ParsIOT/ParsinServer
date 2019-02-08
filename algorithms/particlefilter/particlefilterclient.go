package particlefilter

import (
	"ParsinServer/glb"
	"context"
	"google.golang.org/grpc"
	"log"
	pb "particlefilter/particlefilterclasses"
	"time"
)

var particlefilterClient pb.ParticleFilterClient

func Do_Initialize(initRequest pb.InitRequest) pb.InitReply {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	initReply, err := particlefilterClient.Initialize(ctx, &initRequest)
	if err != nil {
		log.Fatalf("could not greet: %v", err)
	}
	//log.Printf("Greeting: %s", r.ReturnValue)
	return *initReply
}

func Initialize() {
	initRequest := pb.InitRequest{Name: "Hadi"}
	initReply := Do_Initialize(initRequest)
	glb.Debug.Println("Initialize: %s", initReply.ReturnValue)
}

func TestConnection() {
	// Send simple Initialize command to check connection
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	_, err := particlefilterClient.Initialize(ctx, &pb.InitRequest{Name: ""})
	if err != nil {
		log.Fatalf("Connection problem: %v", err)
	} else {
		glb.Debug.Println("Connection Established Successfully")
	}
	//log.Printf("Greeting: %s", r.ReturnValue)
}

func Connect2Server() {
	// Run particle filter server

	glb.Debug.Println("Connecting to the particle filter server(" + glb.RuntimeArgs.ParticleFilterServer + ") ... ")
	// Set up a connection to the server
	conn, err := grpc.Dial(glb.RuntimeArgs.ParticleFilterServer, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	particlefilterClient = pb.NewParticleFilterClient(conn)

	// Main Code
	TestConnection()
	//Initialize()
}
