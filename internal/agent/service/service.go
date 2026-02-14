package service

import (
	"context"
	"log"
	pb "mallekoppie/ChaosGenerator/internal/contracts"

	core "mallekoppie/ChaosGenerator/internal/agent/go"

	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

type Service struct {
	pb.UnimplementedChaosAgentServer
}

func (*Service) AddTests(ctx context.Context, req *pb.TestCollection) (*pb.Response, error) {
	core.WriteTestConfiguration(*req)

	response := pb.Response{Result: true}

	return &response, nil
}

func (*Service) GetTestStatus(ctx context.Context, req *pb.Request) (*pb.TestStatus, error) {

	status := core.CoreGetTestStatus()

	return &status, nil
}

func (*Service) IsAlive(ctx context.Context, req *pb.Request) (*pb.Response, error) {

	response := pb.Response{Result: true}

	return &response, nil
}

func (*Service) StartTestRun(ctx context.Context, req *pb.TestParameters) (*pb.Response, error) {

	started, startErr := core.CoreRunTest(req.GetTestCollectionName(), int(req.GetSimulatedusers()))

	if startErr != nil {
		log.Println("Unable to start test run: ", startErr)

		return nil, status.Errorf(codes.Unknown, "Unable to start test")
	}

	if started == true {
		return &pb.Response{Result: true}, nil

	} else {
		return nil, status.Errorf(codes.Unknown, "No error but test not started")
	}

	return nil, status.Errorf(codes.Unimplemented, "method StartTestRun not implemented")
}

func (*Service) StopTestRun(ctx context.Context, req *pb.StopTestRequest) (*pb.Response, error) {
	core.CoreStopTest()

	response := pb.Response{Result: true}

	return &response, nil

}

func (*Service) UpdateTestRun(ctx context.Context, req *pb.TestParameters) (*pb.Response, error) {

	err := core.CoreUpdateTest(req.GetSimulatedusers())

	if err != nil {
		return nil, status.Errorf(codes.Unknown, "Unknown error while updating test")
	}

	return &pb.Response{Result: true}, nil
}

func (*Service) GetVersion(ctx context.Context, req *pb.Request) (*pb.GetVersionResponse, error) {
	return &pb.GetVersionResponse{Version: "2.0.0", Hostname: "Unknown"}, nil
}

func (*Service) DeleteTests(ctx context.Context, req *pb.DeleteTestsRequest) (*pb.DeleteTestsResponse, error) {
	response := pb.DeleteTestsResponse{}
	err := core.ClearTestsDirectory()
	if err != nil {
		return &response, err
	}

	return &response, nil
}
