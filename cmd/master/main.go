package main

import (
	"embed"
	"fmt"
	"net"
	"os"

	"mallekoppie/ChaosGenerator/internal/master/service"
	"mallekoppie/ChaosGenerator/internal/master/web"

	"github.com/Mallekoppie/goslow/platform"
	"go.uber.org/zap"
)

// Website contains the compiled Flutter web application (see `make buildweb`).
//
//go:embed all:dist
var Website embed.FS

// getEnv retrieves an environment variable or returns a default value.
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func main() {
	debug := os.Getenv("DEBUG") == "true"

	config := platform.Config{}
	config.Component.ComponentName = "chaos-master"
	config.Grpc.Server.ListeningAddress = getEnv("CHAOS_MASTER_GRPC_ADDRESS", "0.0.0.0:9002")
	config.Grpc.Server.TLSCertFileName = "chaos_master.crt"
	config.Grpc.Server.TLSKeyFileName = "chaos_master.key"
	config.Grpc.Server.TLSEnabled = false
	config.Grpc.Server.UnAuthenticatedPaths = []string{}

	// The web server hosts the embedded Flutter app and the gRPC-Web endpoints.
	config.HTTP.Server.ListeningAddress = getEnv("CHAOS_MASTER_WEB_ADDRESS", "0.0.0.0:9001")
	config.HTTP.Server.AllowCorsForLocalDevelopment = true

	// Local JWT authentication for the web UI.
	config.Auth.Server.LocalJwt.Enabled = true
	config.Auth.Server.LocalJwt.JwtSigningKey = getEnv("CHAOS_JWT_SIGNING_KEY", "chaos-master-dev-signing-key-change-me")
	config.Auth.Server.LocalJwt.JwtSigningMethod = "HS256"
	config.Auth.Server.LocalJwt.JwtExpiration = 2400

	config.Database.BoltDB.Enabled = true
	config.Database.BoltDB.FileName = getEnv("CHAOS_MASTER_DB_PATH", "chaos_master.db")

	platform.SetPlatformConfiguration(config)

	platform.SetupBoltDB()

	services := []platform.GRPCService{
		&service.ChaosMasterServer{},
		&service.AuthServer{},
	}

	// The web server runs alongside the native gRPC server. In debug mode only
	// the gRPC-Web endpoints are exposed so `flutter run` can serve the UI.
	var webAssets *embed.FS
	if !debug {
		webAssets = &Website
	}

	// Bind the web port before starting the gRPC server so that an unavailable
	// port fails fast with a clear message instead of leaving the process
	// running without a web UI.
	webListener, err := web.Listen(config.HTTP.Server.ListeningAddress)
	if err != nil {
		platform.Log.Error("Unable to start the web server; set CHAOS_MASTER_WEB_ADDRESS to a free port",
			zap.String("address", config.HTTP.Server.ListeningAddress),
			zap.Error(err))
		os.Exit(1)
	}

	go web.ServeListener(webListener, webAssets, services, true, true)

	if debug {
		fmt.Printf("DEBUG=true: serving gRPC-Web only on %s (no embedded web app)\n", webListener.Addr().String())
	} else if _, port, splitErr := net.SplitHostPort(webListener.Addr().String()); splitErr == nil {
		fmt.Printf("Chaos Master web UI: http://localhost:%s/\n", port)
	}

	// Native gRPC (agents) stays on the configured gRPC port. Note: the
	// JwtInterceptor is only applied by the gRPC-Web server above; goslow only
	// adds its own interceptor to the native server when TLS is enabled.
	platform.StartGrpcServer(services)
}
