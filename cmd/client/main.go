package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"mallekoppie/ChaosGenerator/internal/contracts"

	"github.com/Mallekoppie/goslow/platform"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var client contracts.ChaosMasterClient
var ctx context.Context

func main() {
	// Initialize platform for logging
	config := platform.Config{}
	config.Component.ComponentName = "chaos-client"
	platform.SetPlatformConfiguration(config)

	// Connect to server
	serverAddress := "127.0.0.1:9002"
	platform.Log.Info("Connecting to ChaosMaster server", zap.String("address", serverAddress))

	// For now, use insecure connection (matching server TLSEnabled = false)
	conn, err := grpc.Dial(serverAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		platform.Log.Error("Failed to connect", zap.Error(err))
		return
	}
	defer conn.Close()

	client = contracts.NewChaosMasterClient(conn)
	ctx = context.Background()

	platform.Log.Info("Connected successfully to ChaosMaster server")

	// Start interactive menu
	showMenu()
}

func showMenu() {
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Println("\n=== ChaosMaster Test Client ===")
		fmt.Println("Agent Operations:")
		fmt.Println("  1. GetAllAgents")
		fmt.Println("  2. UpdateAgent")
		fmt.Println("  3. DeleteAgent")
		fmt.Println("\nTestGroup Operations:")
		fmt.Println("  4. GetAllTestGroups")
		fmt.Println("  5. GetTestGroup")
		fmt.Println("  6. AddTestGroup")
		fmt.Println("  7. UpdateTestGroup")
		fmt.Println("  8. DeleteTestGroup")
		fmt.Println("\nTestCollection Operations:")
		fmt.Println("  9. AddTestCollection")
		fmt.Println(" 10. UpdateTestCollection")
		fmt.Println(" 11. DeleteTestCollection")
		fmt.Println("\nOther:")
		fmt.Println(" 12. Run All Agent Tests")
		fmt.Println(" 13. Run All TestGroup Tests")
		fmt.Println(" 14. Run All TestCollection Tests")
		fmt.Println("  0. Exit")
		fmt.Print("\nSelect option: ")

		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		switch input {
		case "1":
			testGetAllAgents()
		case "2":
			testUpdateAgent()
		case "3":
			testDeleteAgent()
		case "4":
			testGetAllTestGroups()
		case "5":
			testGetTestGroup()
		case "6":
			testAddTestGroup()
		case "7":
			testUpdateTestGroup()
		case "8":
			testDeleteTestGroup()
		case "9":
			testAddTestCollection()
		case "10":
			testUpdateTestCollection()
		case "11":
			testDeleteTestCollection()
		case "12":
			runAllAgentTests()
		case "13":
			runAllTestGroupTests()
		case "14":
			runAllTestCollectionTests()
		case "0":
			fmt.Println("Exiting...")
			return
		default:
			fmt.Println("Invalid option, please try again")
		}
	}
}

// Agent Operations
func testGetAllAgents() {
	fmt.Println("\n--- Testing GetAllAgents ---")
	ctx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()

	resp, err := client.GetAllAgents(ctx, &contracts.GetAllAgentsRequest{})
	if err != nil {
		platform.Log.Error("GetAllAgents failed", zap.Error(err))
		fmt.Printf("ERROR: %v\n", err)
		return
	}

	platform.Log.Info("GetAllAgents succeeded", zap.Int("count", len(resp.Agents)))
	fmt.Printf("Found %d agents:\n", len(resp.Agents))
	for i, agent := range resp.Agents {
		fmt.Printf("  [%d] ID: %s, Host: %s, Port: %d, Enabled: %v, Status: %s\n",
			i+1, agent.Id, agent.Host, agent.Port, agent.Enabled, agent.Status)
	}
}

func testUpdateAgent() {
	fmt.Println("\n--- Testing UpdateAgent ---")
	fmt.Print("Enter Agent ID: ")
	reader := bufio.NewReader(os.Stdin)
	id, _ := reader.ReadString('\n')
	id = strings.TrimSpace(id)

	fmt.Print("Enter Host: ")
	host, _ := reader.ReadString('\n')
	host = strings.TrimSpace(host)

	fmt.Print("Enter Port: ")
	var port int32
	fmt.Scanf("%d\n", &port)

	fmt.Print("Enter MetricsPort: ")
	var metricsPort int32
	fmt.Scanf("%d\n", &metricsPort)

	fmt.Print("Enabled (true/false): ")
	var enabled bool
	fmt.Scanf("%t\n", &enabled)

	fmt.Print("Enter Status: ")
	status, _ := reader.ReadString('\n')
	status = strings.TrimSpace(status)

	agent := &contracts.Agent{
		Id:          id,
		Host:        host,
		Port:        port,
		MetricsPort: metricsPort,
		Enabled:     enabled,
		Status:      status,
	}

	ctx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()

	resp, err := client.UpdateAgent(ctx, &contracts.UpdateAgentRequest{Agent: agent})
	if err != nil {
		platform.Log.Error("UpdateAgent failed", zap.Error(err))
		fmt.Printf("ERROR: %v\n", err)
		return
	}

	platform.Log.Info("UpdateAgent succeeded", zap.Bool("success", resp.Success))
	fmt.Printf("Success: %v\n", resp.Success)
}

func testDeleteAgent() {
	fmt.Println("\n--- Testing DeleteAgent ---")
	fmt.Print("Enter Agent ID to delete: ")
	reader := bufio.NewReader(os.Stdin)
	id, _ := reader.ReadString('\n')
	id = strings.TrimSpace(id)

	agent := &contracts.Agent{Id: id}

	ctx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()

	resp, err := client.DeleteAgent(ctx, &contracts.DeleteAgentRequest{Agent: agent})
	if err != nil {
		platform.Log.Error("DeleteAgent failed", zap.Error(err))
		fmt.Printf("ERROR: %v\n", err)
		return
	}

	platform.Log.Info("DeleteAgent succeeded", zap.Bool("success", resp.Success))
	fmt.Printf("Success: %v\n", resp.Success)
}

// TestGroup Operations
func testGetAllTestGroups() {
	fmt.Println("\n--- Testing GetAllTestGroups ---")
	ctx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()

	resp, err := client.GetAllTestGroups(ctx, &contracts.GetAllTestGroupsRequest{})
	if err != nil {
		platform.Log.Error("GetAllTestGroups failed", zap.Error(err))
		fmt.Printf("ERROR: %v\n", err)
		return
	}

	platform.Log.Info("GetAllTestGroups succeeded", zap.Int("count", len(resp.TestGroups)))
	fmt.Printf("Found %d test groups:\n", len(resp.TestGroups))
	for i, group := range resp.TestGroups {
		fmt.Printf("  [%d] ID: %s, Name: %s, Description: %s, Collections: %d\n",
			i+1, group.Id, group.Name, group.Description, len(group.TestCollections))
	}
}

func testGetTestGroup() {
	fmt.Println("\n--- Testing GetTestGroup ---")
	fmt.Print("Enter TestGroup ID: ")
	reader := bufio.NewReader(os.Stdin)
	id, _ := reader.ReadString('\n')
	id = strings.TrimSpace(id)

	ctx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()

	resp, err := client.GetTestGroup(ctx, &contracts.GetTestGroupRequest{Id: id})
	if err != nil {
		platform.Log.Error("GetTestGroup failed", zap.Error(err))
		fmt.Printf("ERROR: %v\n", err)
		return
	}

	platform.Log.Info("GetTestGroup succeeded", zap.String("id", resp.TestGroup.Id))
	fmt.Printf("TestGroup:\n")
	fmt.Printf("  ID: %s\n", resp.TestGroup.Id)
	fmt.Printf("  Name: %s\n", resp.TestGroup.Name)
	fmt.Printf("  Description: %s\n", resp.TestGroup.Description)
	fmt.Printf("  Collections: %d\n", len(resp.TestGroup.TestCollections))
	for i, collection := range resp.TestGroup.TestCollections {
		fmt.Printf("    [%d] %s - %s (Tests: %d)\n", i+1, collection.Name, collection.Description, len(collection.Tests))
	}
}

func testAddTestGroup() {
	fmt.Println("\n--- Testing AddTestGroup ---")
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Enter TestGroup Name: ")
	name, _ := reader.ReadString('\n')
	name = strings.TrimSpace(name)

	fmt.Print("Enter Description: ")
	description, _ := reader.ReadString('\n')
	description = strings.TrimSpace(description)

	testGroup := &contracts.TestGroup{
		Name:            name,
		Description:     description,
		TestCollections: []*contracts.MasterTestCollection{},
	}

	ctx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()

	resp, err := client.AddTestGroup(ctx, &contracts.AddTestGroupRequest{TestGroup: testGroup})
	if err != nil {
		platform.Log.Error("AddTestGroup failed", zap.Error(err))
		fmt.Printf("ERROR: %v\n", err)
		return
	}

	platform.Log.Info("AddTestGroup succeeded", zap.Bool("success", resp.Success))
	fmt.Printf("Success: %v\n", resp.Success)
}

func testUpdateTestGroup() {
	fmt.Println("\n--- Testing UpdateTestGroup ---")
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Enter TestGroup ID: ")
	id, _ := reader.ReadString('\n')
	id = strings.TrimSpace(id)

	fmt.Print("Enter New Name: ")
	name, _ := reader.ReadString('\n')
	name = strings.TrimSpace(name)

	fmt.Print("Enter New Description: ")
	description, _ := reader.ReadString('\n')
	description = strings.TrimSpace(description)

	testGroup := &contracts.TestGroup{
		Id:              id,
		Name:            name,
		Description:     description,
		TestCollections: []*contracts.MasterTestCollection{},
	}

	ctx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()

	resp, err := client.UpdateTestGroup(ctx, &contracts.UpdateTestGroupRequest{TestGroup: testGroup})
	if err != nil {
		platform.Log.Error("UpdateTestGroup failed", zap.Error(err))
		fmt.Printf("ERROR: %v\n", err)
		return
	}

	platform.Log.Info("UpdateTestGroup succeeded", zap.Bool("success", resp.Success))
	fmt.Printf("Success: %v\n", resp.Success)
}

func testDeleteTestGroup() {
	fmt.Println("\n--- Testing DeleteTestGroup ---")
	fmt.Print("Enter TestGroup ID to delete: ")
	reader := bufio.NewReader(os.Stdin)
	id, _ := reader.ReadString('\n')
	id = strings.TrimSpace(id)

	ctx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()

	resp, err := client.DeleteTestGroup(ctx, &contracts.DeleteTestGroupRequest{Id: id})
	if err != nil {
		platform.Log.Error("DeleteTestGroup failed", zap.Error(err))
		fmt.Printf("ERROR: %v\n", err)
		return
	}

	platform.Log.Info("DeleteTestGroup succeeded", zap.Bool("success", resp.Success))
	fmt.Printf("Success: %v\n", resp.Success)
}

// TestCollection Operations
func testAddTestCollection() {
	fmt.Println("\n--- Testing AddTestCollection ---")
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Enter TestCollection Name: ")
	name, _ := reader.ReadString('\n')
	name = strings.TrimSpace(name)

	fmt.Print("Enter Description: ")
	description, _ := reader.ReadString('\n')
	description = strings.TrimSpace(description)

	fmt.Print("Enter Parent TestGroup ID: ")
	groupId, _ := reader.ReadString('\n')
	groupId = strings.TrimSpace(groupId)

	testCollection := &contracts.MasterTestCollection{
		Name:        name,
		Description: description,
		GroupId:     groupId,
		Tests:       []*contracts.MasterTest{},
	}

	ctx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()

	resp, err := client.AddTestCollection(ctx, &contracts.AddTestCollectionRequest{TestCollection: testCollection})
	if err != nil {
		platform.Log.Error("AddTestCollection failed", zap.Error(err))
		fmt.Printf("ERROR: %v\n", err)
		return
	}

	platform.Log.Info("AddTestCollection succeeded", zap.Bool("success", resp.Success))
	fmt.Printf("Success: %v\n", resp.Success)
	if resp.Error != "" {
		fmt.Printf("Error: %s\n", resp.Error)
	}
}

func testUpdateTestCollection() {
	fmt.Println("\n--- Testing UpdateTestCollection ---")
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Enter TestCollection ID: ")
	id, _ := reader.ReadString('\n')
	id = strings.TrimSpace(id)

	fmt.Print("Enter New Name: ")
	name, _ := reader.ReadString('\n')
	name = strings.TrimSpace(name)

	fmt.Print("Enter New Description: ")
	description, _ := reader.ReadString('\n')
	description = strings.TrimSpace(description)

	fmt.Print("Enter Parent TestGroup ID: ")
	groupId, _ := reader.ReadString('\n')
	groupId = strings.TrimSpace(groupId)

	testCollection := &contracts.MasterTestCollection{
		Id:          id,
		Name:        name,
		Description: description,
		GroupId:     groupId,
		Tests:       []*contracts.MasterTest{},
	}

	ctx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()

	resp, err := client.UpdateTestCollection(ctx, &contracts.UpdateTestCollectionRequest{TestCollection: testCollection})
	if err != nil {
		platform.Log.Error("UpdateTestCollection failed", zap.Error(err))
		fmt.Printf("ERROR: %v\n", err)
		return
	}

	platform.Log.Info("UpdateTestCollection succeeded", zap.Bool("success", resp.Success))
	fmt.Printf("Success: %v\n", resp.Success)
	if resp.Error != "" {
		fmt.Printf("Error: %s\n", resp.Error)
	}
}

func testDeleteTestCollection() {
	fmt.Println("\n--- Testing DeleteTestCollection ---")
	fmt.Print("Enter TestCollection ID to delete: ")
	reader := bufio.NewReader(os.Stdin)
	id, _ := reader.ReadString('\n')
	id = strings.TrimSpace(id)

	ctx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()

	resp, err := client.DeleteTestCollection(ctx, &contracts.DeleteTestCollectionRequest{Id: id})
	if err != nil {
		platform.Log.Error("DeleteTestCollection failed", zap.Error(err))
		fmt.Printf("ERROR: %v\n", err)
		return
	}

	platform.Log.Info("DeleteTestCollection succeeded", zap.Bool("success", resp.Success))
	fmt.Printf("Success: %v\n", resp.Success)
}

// Comprehensive Test Suites
func runAllAgentTests() {
	fmt.Println("\n=== Running All Agent Tests ===")
	testGetAllAgents()
	time.Sleep(1 * time.Second)
}

func runAllTestGroupTests() {
	fmt.Println("\n=== Running All TestGroup Tests ===")

	// Add a test group
	fmt.Println("\n[1/4] Adding test group...")
	testGroup := &contracts.TestGroup{
		Name:            "Test Group " + time.Now().Format("15:04:05"),
		Description:     "Automatically created test group",
		TestCollections: []*contracts.MasterTestCollection{},
	}

	ctx1, cancel1 := context.WithTimeout(ctx, time.Second*10)
	defer cancel1()

	addResp, err := client.AddTestGroup(ctx1, &contracts.AddTestGroupRequest{TestGroup: testGroup})
	if err != nil {
		platform.Log.Error("AddTestGroup failed", zap.Error(err))
		return
	}
	fmt.Printf("Added test group: Success=%v\n", addResp.Success)

	time.Sleep(1 * time.Second)

	// Get all test groups
	fmt.Println("\n[2/4] Getting all test groups...")
	testGetAllTestGroups()

	time.Sleep(1 * time.Second)

	// Get specific test group (if we have any)
	fmt.Println("\n[3/4] Getting specific test group...")
	ctx2, cancel2 := context.WithTimeout(ctx, time.Second*10)
	defer cancel2()

	allResp, err := client.GetAllTestGroups(ctx2, &contracts.GetAllTestGroupsRequest{})
	if err == nil && len(allResp.TestGroups) > 0 {
		firstGroup := allResp.TestGroups[0]
		ctx3, cancel3 := context.WithTimeout(ctx, time.Second*10)
		defer cancel3()

		getResp, err := client.GetTestGroup(ctx3, &contracts.GetTestGroupRequest{Id: firstGroup.Id})
		if err != nil {
			platform.Log.Error("GetTestGroup failed", zap.Error(err))
		} else {
			fmt.Printf("Retrieved test group: %s - %s\n", getResp.TestGroup.Name, getResp.TestGroup.Description)
		}
	}

	fmt.Println("\n[4/4] Complete! Use individual tests to update or delete test groups.")
}

func runAllTestCollectionTests() {
	fmt.Println("\n=== Running All TestCollection Tests ===")
	fmt.Println("TestCollection tests require a parent TestGroup ID.")
	fmt.Println("Please use individual tests (options 9-11) with proper TestGroup IDs.")

	// Show available test groups
	testGetAllTestGroups()
}
