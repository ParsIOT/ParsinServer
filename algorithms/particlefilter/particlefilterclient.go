package particlefilter

import (
	"ParsinServer/glb"
	"context"
	"google.golang.org/grpc"
	"log"
	pb "ParsinServer/algorithms/particlefilter/particlefilterclasses"
	"time"
)

var particlefilterClient pb.ParticleFilterClient
var conn grpc.ClientConn
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

func Do_Predict(predictRequest pb.PredictRequest) pb.PredictReply {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	predictReply, err := particlefilterClient.Predict(ctx, &predictRequest)
	if err != nil {
		glb.Debug.Println(predictReply)
		log.Fatalf("could not greet: %v", err)
	}
	//log.Printf("Greeting: %s", r.ReturnValue)
	return *predictReply
}

func Do_Update(updateRequest pb.UpdateRequest) pb.UpdateReply {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	updateReply, err := particlefilterClient.Update(ctx, &updateRequest)
	if err != nil {
		log.Fatalf("could not greet: %v", err)
	}
	//log.Printf("Greeting: %s", r.ReturnValue)
	return *updateReply
}

func Initialize(timestamp int64, initLocation []float32) {
	initRequest := pb.InitRequest{Timestamp: timestamp, XY: initLocation}
	initReply := Do_Initialize(initRequest)
	glb.Debug.Println("Initialization: ", initReply.ReturnValue)
}

func Predict(timestamp int64) []float32 {
	predictRequest := pb.PredictRequest{Timestamp: timestamp}
	predictReply := Do_Predict(predictRequest)
	//glb.Debug.Println("Prediction: ", predictReply.ResXY)
	return predictReply.ResXY
}

func Update(timestamp int64, blePredict []float32) []float32 {
	updateRequest := pb.UpdateRequest{Timestamp: timestamp, BlePredict: blePredict}
	updateReply := Do_Update(updateRequest)
	//glb.Debug.Println("Update: ", updateReply.ResXY)
	return updateReply.ResXY
}


func TestConnection() {
	// Send simple Initialize command to check connection
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	reply, err := particlefilterClient.ConnectionTest(ctx, &pb.Empty{})
	if err != nil || reply.ReturnValue == false {
		log.Fatalf("Connection problem: %v", err)
	} else {
		glb.Debug.Println("Connection Established Successfully,", reply)
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
	//defer conn.Close()
	particlefilterClient = pb.NewParticleFilterClient(conn)

	// Main Code

	TestConnection()

	//timestamp := time.Now().UTC().UnixNano()/1000000
	//Initialize(timestamp, []float32{0.0,0.0})
	//
	//timestamp = time.Now().UTC().UnixNano()/1000000
	//Predict(timestamp)
	//
	//timestamp = time.Now().UTC().UnixNano()/1000000
	//Update(timestamp, []float32{3.0,3.0})
}
