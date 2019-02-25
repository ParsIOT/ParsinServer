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
		glb.Error.Println("Can't Do_Initialize: ", err.Error())
	}
	return *initReply
}

func Do_Predict(predictRequest pb.PredictRequest) pb.PredictReply {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*100)
	defer cancel()
	predictReply, err := particlefilterClient.Predict(ctx, &predictRequest)
	if err != nil {
		glb.Debug.Println(predictReply)
		//log.Fatalf("could not greet: %v", err)
		glb.Error.Println("Can't Do_Predict: ", err.Error())
	}
	if (len(predictReply.ResXY) != 2){
		glb.Error.Println("Invalid Do_Predict result.ResXY")
	}
	return *predictReply
}

func Do_Update(updateRequest pb.UpdateRequest) pb.UpdateReply {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*100)
	defer cancel()
	updateReply, err := particlefilterClient.Update(ctx, &updateRequest)
	if err != nil {
		//log.Fatalf("could not greet: %v", err)
		glb.Error.Println("Can't Do_Update: ", err.Error())
	}
	if (len(updateReply.ResXY) != 2){
		glb.Error.Println("Invalid Do_Update result.ResXY")
	}
	return *updateReply
}

func Initialize(timestamp int64, initLocation []float32, mapGraph pb.Graph) {
	initRequest := pb.InitRequest{Timestamp: timestamp, XY: initLocation, MapGraph: &mapGraph}
	initReply := Do_Initialize(initRequest)
	glb.Debug.Println("Initialization: ", initReply.ReturnValue)
}

func Predict(timestamp int64) []float32 {
	predictRequest := pb.PredictRequest{Timestamp: timestamp}
	predictReply := Do_Predict(predictRequest)
	return predictReply.ResXY
}

func Update(timestamp int64, masterEstimation, slaveEstimation, trueLocation []float32) []float32 {
	updateRequest := pb.UpdateRequest{
		Timestamp:        timestamp,
		MasterEstimation: masterEstimation,
		SlaveEstimation:  slaveEstimation,
		TrueLocation:     trueLocation,
	}
	updateReply := Do_Update(updateRequest)
	return updateReply.ResXY
}


func TestConnection() {
	// Send simple Initialize command to check connection
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	reply, err := particlefilterClient.ConnectionTest(ctx, &pb.Empty{})
	if err != nil || reply.ReturnValue == false {
		//log.Fatalf("Connection problem: %v", err)
		glb.Error.Println("Connection problem: ", err.Error())
	} else {
		glb.Debug.Println("Connection Established Successfully,", reply)
	}
}

func Connect2Server() {
	// Run particle filter server

	glb.Debug.Println("Connecting to the particle filter server(" + glb.RuntimeArgs.ParticleFilterServer + ") ... ")
	// Set up a connection to the server
	conn, err := grpc.Dial(glb.RuntimeArgs.ParticleFilterServer, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Didn't connect: %v", err)
	}
	//defer conn.Close()
	particlefilterClient = pb.NewParticleFilterClient(conn)

	// Main Code
	TestConnection()

}
