package main

import (
	"mallekoppie/ChaosGenerator/internal/master/service"

	"github.com/Mallekoppie/goslow/platform"
)

func main() {
	config := platform.Config{}
	config.Component.ComponentName = "chaos-master"
	config.Grpc.Server.ListeningAddress = "0.0.0.0:9002"
	config.Grpc.Server.TLSCertFileName = "chaos_master.crt"
	config.Grpc.Server.TLSKeyFileName = "chaos_master.key"
	config.Grpc.Server.TLSEnabled = false
	config.Grpc.Server.UnAuthenticatedPaths = []string{}
	config.Auth.Server.LocalJwt.Enabled = false
	config.Database.BoltDB.Enabled = true
	config.Database.BoltDB.FileName = "chaos_master.db"

	platform.SetPlatformConfiguration(config)

	platform.SetupBoltDB()

	services := []platform.GRPCService{
		&service.ChaosMasterServer{},
	}

	platform.StartGrpcServer(services)
}
