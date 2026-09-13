package web

import (
	"embed"
	"io/fs"
	"net"
	"net/http"
	"strings"

	"mallekoppie/ChaosGenerator/internal/master/service"

	"github.com/Mallekoppie/goslow/platform"
	grpcweb "github.com/improbable-eng/grpc-web/go/grpcweb"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

// allowedRequestHeaders mirrors the headers sent by the gRPC-Web XHR transport
// so browser pre-flight requests succeed when CORS is enabled.
var allowedRequestHeaders = []string{
	"Authorization",
	"Content-Type",
	"X-Grpc-Web",
	"X-User-Agent",
	"Grpc-Timeout",
	"X-Accept-Content-Transfer-Encoding",
	"X-Accept-Response-Streaming",
}

// Handler serves the embedded Flutter web app and the gRPC-Web endpoints from
// the same origin. When webAssets is nil only the gRPC-Web endpoints and the
// health probe are served, which is used when running the Flutter app against a
// local development server.
func Handler(webAssets *embed.FS, services []platform.GRPCService, jwtEnabled bool, allowCors bool) (http.Handler, error) {
	serverOptions := []grpc.ServerOption{}
	if jwtEnabled {
		serverOptions = append(serverOptions, grpc.UnaryInterceptor(service.JwtInterceptor))
	}

	grpcServer := grpc.NewServer(serverOptions...)
	for _, svc := range services {
		svc.Register(grpcServer)
	}

	wrapOptions := []grpcweb.Option{}
	if allowCors {
		wrapOptions = append(wrapOptions,
			grpcweb.WithOriginFunc(func(string) bool { return true }),
			grpcweb.WithAllowedRequestHeaders(allowedRequestHeaders),
		)
	}
	wrapped := grpcweb.WrapServer(grpcServer, wrapOptions...)

	var (
		uiFS       fs.FS
		fileServer http.Handler
		indexHTML  []byte
	)

	if webAssets != nil {
		stripped, err := fs.Sub(*webAssets, "dist")
		if err != nil {
			return nil, err
		}
		uiFS = stripped
		fileServer = http.FileServer(http.FS(stripped))
		if content, err := fs.ReadFile(stripped, "index.html"); err == nil {
			indexHTML = content
		}
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Kubernetes liveness/readiness probe.
		if r.URL.Path == "/healthz" {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("ok"))
			return
		}

		// gRPC-Web requests (and their CORS pre-flights) are handled by the
		// wrapper, which translates them to native gRPC calls.
		if wrapped.IsGrpcWebRequest(r) || wrapped.IsAcceptableGrpcCorsRequest(r) {
			wrapped.ServeHTTP(w, r)
			return
		}

		if fileServer == nil {
			http.NotFound(w, r)
			return
		}

		// Single page application fallback: unknown, extension-less paths are
		// served the compiled index.html.
		if r.URL.Path != "/" && indexHTML != nil && !strings.Contains(r.URL.Path, ".") {
			if _, err := fs.Stat(uiFS, strings.TrimPrefix(r.URL.Path, "/")); err != nil {
				w.Header().Set("Content-Type", "text/html; charset=utf-8")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write(indexHTML)
				return
			}
		}

		fileServer.ServeHTTP(w, r)
	}), nil
}

// Listen binds the address used for the web app and the gRPC-Web endpoints.
// Returning the listener lets callers fail fast with a clear message when the
// port is already in use instead of running without a web UI.
func Listen(listenAddress string) (net.Listener, error) {
	return net.Listen("tcp", listenAddress)
}

// ServeListener serves the web app and the gRPC-Web endpoints on an existing
// listener. It blocks until the server stops.
func ServeListener(listener net.Listener, webAssets *embed.FS, services []platform.GRPCService, jwtEnabled bool, allowCors bool) {
	handler, err := Handler(webAssets, services, jwtEnabled, allowCors)
	if err != nil {
		platform.Log.Error("Unable to build web handler", zap.Error(err))
		return
	}

	platform.Log.Info("Starting chaos-master web server", zap.String("ListeningAddress", listener.Addr().String()))
	if err := http.Serve(listener, handler); err != nil {
		platform.Log.Error("Web server stopped", zap.Error(err))
	}
}
