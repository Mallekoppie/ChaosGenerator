package web

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"mallekoppie/ChaosGenerator/internal/contracts"
	"mallekoppie/ChaosGenerator/internal/master/service"

	"github.com/Mallekoppie/goslow/platform"
	"google.golang.org/protobuf/proto"
)

const webTestRegisterSecret = "web-test-register-secret"

var testDBFile string

func TestMain(m *testing.M) {
	os.Setenv("CHAOS_REGISTER_SECRET", webTestRegisterSecret)

	testDBFile = filepath.Join(os.TempDir(), fmt.Sprintf("chaos_web_%d.db", time.Now().UnixNano()))

	config := platform.Config{}
	config.Component.ComponentName = "chaos-web-tests"
	config.Log.Level = "error"
	config.Auth.Server.LocalJwt.Enabled = true
	config.Auth.Server.LocalJwt.JwtSigningKey = "unit-test-signing-key"
	config.Auth.Server.LocalJwt.JwtSigningMethod = "HS256"
	config.Auth.Server.LocalJwt.JwtExpiration = 60
	config.Database.BoltDB.Enabled = true
	config.Database.BoltDB.FileName = testDBFile

	platform.SetPlatformConfiguration(config)

	if err := platform.SetupBoltDB(); err != nil {
		fmt.Println("Unable to set up BoltDB for tests:", err)
		os.Exit(1)
	}

	code := m.Run()

	_ = platform.Database.BoltDb.Close()
	os.Remove(testDBFile)
	os.Remove(testDBFile + ".lock")

	os.Exit(code)
}

func testServices() []platform.GRPCService {
	return []platform.GRPCService{
		&service.AuthServer{},
	}
}

func newTestServer(t *testing.T) *httptest.Server {
	t.Helper()

	handler, err := Handler(nil, testServices(), true, true)
	if err != nil {
		t.Fatalf("Failed to build handler: %v", err)
	}

	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)

	return server
}

// grpcWebFrame wraps a protobuf message in the gRPC-Web length prefixed frame.
func grpcWebFrame(message []byte) []byte {
	frame := make([]byte, 5+len(message))
	frame[0] = 0
	binary.BigEndian.PutUint32(frame[1:5], uint32(len(message)))
	copy(frame[5:], message)
	return frame
}

// parseGrpcWebBody extracts the first message frame and the trailer frame.
func parseGrpcWebBody(body []byte) (payload []byte, trailers string) {
	for i := 0; i+5 <= len(body); {
		flag := body[i]
		length := int(binary.BigEndian.Uint32(body[i+1 : i+5]))
		if i+5+length > len(body) {
			break
		}
		chunk := body[i+5 : i+5+length]
		if flag&0x80 != 0 {
			trailers = string(chunk)
		} else if payload == nil {
			payload = chunk
		}
		i += 5 + length
	}
	return payload, trailers
}

func callGrpcWeb(t *testing.T, server *httptest.Server, method string, message proto.Message, token string) ([]byte, string) {
	t.Helper()

	var body []byte
	if message != nil {
		raw, err := proto.Marshal(message)
		if err != nil {
			t.Fatalf("Failed to marshal request: %v", err)
		}
		body = grpcWebFrame(raw)
	}

	req, err := http.NewRequest(http.MethodPost, server.URL+method, bytes.NewReader(body))
	if err != nil {
		t.Fatalf("Failed to build request: %v", err)
	}
	req.Header.Set("Content-Type", "application/grpc-web+proto")
	req.Header.Set("X-Grpc-Web", "1")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := server.Client().Do(req)
	if err != nil {
		t.Fatalf("gRPC-Web request failed: %v", err)
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected HTTP 200, got %d", resp.StatusCode)
	}

	payload, trailers := parseGrpcWebBody(responseBody)

	// Trailers-only responses (such as an unauthenticated error) carry the gRPC
	// status in the HTTP headers rather than a trailer frame in the body.
	if status := resp.Header.Get("grpc-status"); status != "" {
		trailers = fmt.Sprintf("grpc-status: %s\n%s", status, trailers)
	}
	if message := resp.Header.Get("grpc-message"); message != "" {
		trailers = fmt.Sprintf("%sgrpc-message: %s", trailers, message)
	}

	return payload, trailers
}

func TestHealthz(t *testing.T) {
	server := newTestServer(t)

	resp, err := server.Client().Get(server.URL + "/healthz")
	if err != nil {
		t.Fatalf("Failed to call healthz: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected 200 from healthz, got %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	if string(body) != "ok" {
		t.Fatalf("Unexpected healthz body: %q", string(body))
	}
}

func TestUnknownPathReturnsNotFoundWhenServingGRPCWebOnly(t *testing.T) {
	server := newTestServer(t)

	resp, err := server.Client().Get(server.URL + "/")
	if err != nil {
		t.Fatalf("Failed to call root: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("Expected 404 when no web assets are embedded, got %d", resp.StatusCode)
	}
}

func TestLoginUnknownUser(t *testing.T) {
	server := newTestServer(t)

	payload, trailers := callGrpcWeb(t, server, "/contract.Auth/Login", &contracts.LoginRequest{
		Username: "nobody",
		Password: "password123",
	}, "")

	if !strings.Contains(trailers, "grpc-status: 0") {
		t.Fatalf("Expected grpc-status 0, trailers: %q", trailers)
	}

	response := &contracts.LoginResponse{}
	if err := proto.Unmarshal(payload, response); err != nil {
		t.Fatalf("Failed to unmarshal login response: %v", err)
	}
	if response.Success {
		t.Fatal("Expected login to fail for an unknown user")
	}
}

func TestRegisterLoginAndAuthorisedCall(t *testing.T) {
	server := newTestServer(t)

	payload, _ := callGrpcWeb(t, server, "/contract.Auth/RegisterUser", &contracts.RegisterRequest{
		Username:       "webuser",
		Password:       "password123",
		RegisterSecret: webTestRegisterSecret,
	}, "")

	registerResponse := &contracts.RegisterResponse{}
	if err := proto.Unmarshal(payload, registerResponse); err != nil {
		t.Fatalf("Failed to unmarshal register response: %v", err)
	}
	if !registerResponse.Success {
		t.Fatalf("Expected registration to succeed: %s", registerResponse.Message)
	}

	// A protected call must fail without a token.
	_, trailers := callGrpcWeb(t, server, "/contract.Auth/GetAllUsers", &contracts.GetAllUsersRequest{}, "")
	if !strings.Contains(trailers, "grpc-status: 16") {
		t.Fatalf("Expected unauthenticated status, trailers: %q", trailers)
	}

	payload, _ = callGrpcWeb(t, server, "/contract.Auth/Login", &contracts.LoginRequest{
		Username: "webuser",
		Password: "password123",
	}, "")

	loginResponse := &contracts.LoginResponse{}
	if err := proto.Unmarshal(payload, loginResponse); err != nil {
		t.Fatalf("Failed to unmarshal login response: %v", err)
	}
	if !loginResponse.Success || loginResponse.Token == "" {
		t.Fatalf("Expected a token from login: %s", loginResponse.Message)
	}

	payload, trailers = callGrpcWeb(t, server, "/contract.Auth/GetAllUsers", &contracts.GetAllUsersRequest{}, loginResponse.Token)
	if !strings.Contains(trailers, "grpc-status: 0") {
		t.Fatalf("Expected grpc-status 0, trailers: %q", trailers)
	}

	usersResponse := &contracts.GetAllUsersResponse{}
	if err := proto.Unmarshal(payload, usersResponse); err != nil {
		t.Fatalf("Failed to unmarshal users response: %v", err)
	}
	if len(usersResponse.Users) == 0 {
		t.Fatal("Expected at least one registered user")
	}

	found := false
	for _, user := range usersResponse.Users {
		if user.Username == "webuser" {
			found = true
		}
	}
	if !found {
		t.Fatal("Expected to find the registered user")
	}
}

func TestRegisterInvalidSecret(t *testing.T) {
	server := newTestServer(t)

	payload, _ := callGrpcWeb(t, server, "/contract.Auth/RegisterUser", &contracts.RegisterRequest{
		Username:       "webuserbadsecret",
		Password:       "password123",
		RegisterSecret: "the-wrong-secret",
	}, "")

	response := &contracts.RegisterResponse{}
	if err := proto.Unmarshal(payload, response); err != nil {
		t.Fatalf("Failed to unmarshal register response: %v", err)
	}
	if response.Success {
		t.Fatal("Expected registration to fail with an invalid secret")
	}
}

func TestListenReturnsErrorWhenAddressInUse(t *testing.T) {
	listener, err := Listen("127.0.0.1:0")
	if err != nil {
		t.Fatalf("Failed to listen on an ephemeral port: %v", err)
	}
	defer listener.Close()

	if _, err := Listen(listener.Addr().String()); err == nil {
		t.Fatal("Expected an error when the address is already in use")
	}
}
