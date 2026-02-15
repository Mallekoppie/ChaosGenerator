package client

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"mallekoppie/ChaosGenerator/internal/contracts"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Client struct {
	grpcClient contracts.ChaosMasterClient
	ctx        context.Context
	conn       *grpc.ClientConn
	reader     *bufio.Reader
}

func NewClient(serverAddress string) (*Client, error) {
	conn, err := grpc.Dial(serverAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect: %w", err)
	}

	return &Client{
		grpcClient: contracts.NewChaosMasterClient(conn),
		ctx:        context.Background(),
		conn:       conn,
		reader:     bufio.NewReader(os.Stdin),
	}, nil
}

func (c *Client) Close() error {
	return c.conn.Close()
}

func (c *Client) Start() {
	for {
		fmt.Println("\n=== ChaosMaster Test Client ===")
		fmt.Println("Target Operations:")
		fmt.Println("  1. List Targets")
		fmt.Println("  2. Register Target")
		fmt.Println("  3. Update Target")
		fmt.Println("  4. Delete Target")
		fmt.Println("\nTest Execution:")
		fmt.Println("  5. Start Test (Interactive)")
		fmt.Println("  6. Stop Test")
		fmt.Println("\nInformation:")
		fmt.Println("  7. List Use Cases")
		fmt.Println("  8. List Agents")
		fmt.Println("\nMaintenance:")
		fmt.Println("  9. Clear All Agents")
		fmt.Println("\n  0. Exit")
		fmt.Print("\nSelect option: ")

		input, _ := c.reader.ReadString('\n')
		input = strings.TrimSpace(input)

		switch input {
		case "1":
			c.listTargets()
		case "2":
			c.registerTarget()
		case "3":
			c.updateTarget()
		case "4":
			c.deleteTarget()
		case "5":
			c.startTestInteractive()
		case "6":
			c.stopTest()
		case "7":
			c.listUseCases()
		case "8":
			c.listAgents()
		case "9":
			c.clearAllAgents()
		case "0":
			fmt.Println("Exiting...")
			return
		default:
			fmt.Println("Invalid option. Please try again.")
		}
	}
}

func (c *Client) readInput(prompt string) string {
	fmt.Print(prompt)
	input, _ := c.reader.ReadString('\n')
	return strings.TrimSpace(input)
}

func (c *Client) readBool(prompt string) bool {
	input := c.readInput(prompt + " (y/n): ")
	return strings.ToLower(input) == "y" || strings.ToLower(input) == "yes"
}

func (c *Client) readInt(prompt string) int {
	input := c.readInput(prompt)
	val, err := strconv.Atoi(input)
	if err != nil {
		fmt.Printf("Invalid number, using default: 0\n")
		return 0
	}
	return val
}

// Target Operations

func (c *Client) listTargets() {
	resp, err := c.grpcClient.GetAllTargets(c.ctx, &contracts.GetAllTargetsRequest{})
	if err != nil {
		fmt.Printf("Error getting targets: %v\n", err)
		return
	}

	if len(resp.Targets) == 0 {
		fmt.Println("No targets registered.")
		return
	}

	fmt.Println("\n=== Registered Targets ===")
	for i, target := range resp.Targets {
		fmt.Printf("\n%d. %s (ID: %s)\n", i+1, target.Name, target.Id)
		fmt.Printf("   Address: %s\n", target.Address)
		fmt.Printf("   Protocol: %s\n", target.Protocol)
		fmt.Printf("   Connection Reuse: %v\n", target.ConnectionReuseEnabled)
		if target.Sni != "" {
			fmt.Printf("   SNI: %s\n", target.Sni)
		}
	}
}

func (c *Client) registerTarget() {
	fmt.Println("\n=== Register New Target ===")

	name := c.readInput("Target Name: ")
	address := c.readInput("Target Address (e.g., http://localhost:8080): ")

	// Protocol selection with numbered list
	fmt.Println("\nSelect Protocol:")
	fmt.Println("1. http")
	fmt.Println("2. https")
	fmt.Println("3. grpc")
	protocolChoice := c.readInput("Enter number (1-3): ")

	protocolMap := map[string]string{
		"1": "http",
		"2": "https",
		"3": "grpc",
	}

	protocol, ok := protocolMap[protocolChoice]
	if !ok {
		fmt.Println("Invalid protocol choice. Using 'http' as default.")
		protocol = "http"
	}

	connectionReuse := c.readBool("Enable Connection Reuse")

	// SNI is only relevant for HTTPS/TLS
	sni := ""
	if protocol == "https" || protocol == "grpc" {
		sni = c.readInput("SNI (Server Name Indication, leave empty to skip): ")
	}

	req := &contracts.RegisterTargetRequest{
		Target: &contracts.Target{
			Name:                   name,
			Address:                address,
			Protocol:               protocol,
			ConnectionReuseEnabled: connectionReuse,
			Sni:                    sni,
		},
	}

	resp, err := c.grpcClient.RegisterTarget(c.ctx, req)
	if err != nil {
		fmt.Printf("Error registering target: %v\n", err)
		return
	}

	if resp.Success {
		fmt.Printf("✓ Target registered successfully with ID: %s\n", resp.TargetId)
	} else {
		fmt.Printf("✗ Failed to register target: %s\n", resp.Message)
	}
}

func (c *Client) updateTarget() {
	// First list targets
	c.listTargets()

	fmt.Println("\n=== Update Target ===")
	targetId := c.readInput("Target ID to update: ")

	name := c.readInput("New Name (leave empty to skip): ")
	address := c.readInput("New Address (leave empty to skip): ")

	// Protocol selection with numbered list
	fmt.Println("\nUpdate Protocol (leave empty to skip):")
	fmt.Println("1. http")
	fmt.Println("2. https")
	fmt.Println("3. grpc")
	protocolChoice := c.readInput("Enter number (1-3) or press Enter to skip: ")

	protocol := ""
	if protocolChoice != "" {
		protocolMap := map[string]string{
			"1": "http",
			"2": "https",
			"3": "grpc",
		}
		if p, ok := protocolMap[protocolChoice]; ok {
			protocol = p
		}
	}

	sni := c.readInput("New SNI (leave empty to skip): ")

	// Build update request
	req := &contracts.UpdateTargetRequest{
		Target: &contracts.Target{
			Id: targetId,
		},
	}

	if name != "" {
		req.Target.Name = name
	}
	if address != "" {
		req.Target.Address = address
	}
	if protocol != "" {
		req.Target.Protocol = protocol
	}
	if sni != "" {
		req.Target.Sni = sni
	}

	resp, err := c.grpcClient.UpdateTarget(c.ctx, req)
	if err != nil {
		fmt.Printf("Error updating target: %v\n", err)
		return
	}

	if resp.Success {
		fmt.Println("✓ Target updated successfully")
	} else {
		fmt.Printf("✗ Failed to update target: %s\n", resp.Message)
	}
}

func (c *Client) deleteTarget() {
	// Get all targets
	resp, err := c.grpcClient.GetAllTargets(c.ctx, &contracts.GetAllTargetsRequest{})
	if err != nil {
		fmt.Printf("Error getting targets: %v\n", err)
		return
	}

	if len(resp.Targets) == 0 {
		fmt.Println("No targets registered.")
		return
	}

	// Display targets
	fmt.Println("\n=== Delete Target ===")
	fmt.Println("Registered Targets:")
	for i, target := range resp.Targets {
		fmt.Printf("%d. %s (ID: %s)\n", i+1, target.Name, target.Id)
		fmt.Printf("   Address: %s | Protocol: %s\n", target.Address, target.Protocol)
		fmt.Println()
	}

	// Get user selection
	selection := c.readInt(fmt.Sprintf("Enter target number to delete (1-%d) or 0 to cancel: ", len(resp.Targets)))

	if selection == 0 {
		fmt.Println("Cancelled.")
		return
	}

	if selection < 1 || selection > len(resp.Targets) {
		fmt.Println("Invalid selection.")
		return
	}

	selectedTarget := resp.Targets[selection-1]

	// Confirm deletion
	if !c.readBool(fmt.Sprintf("Are you sure you want to delete target '%s' (%s)?", selectedTarget.Name, selectedTarget.Address)) {
		fmt.Println("Cancelled.")
		return
	}

	// Delete the target
	deleteResp, err := c.grpcClient.DeleteTarget(c.ctx, &contracts.DeleteTargetRequest{Id: selectedTarget.Id})
	if err != nil {
		fmt.Printf("Error deleting target: %v\n", err)
		return
	}

	if deleteResp.Success {
		fmt.Println("✓ Target deleted successfully")
	} else {
		fmt.Printf("✗ Failed to delete target: %s\n", deleteResp.Message)
	}
}

// Use Case Operations

func (c *Client) listUseCases() {
	resp, err := c.grpcClient.GetUseCases(c.ctx, &contracts.GetUseCasesRequest{})
	if err != nil {
		fmt.Printf("Error getting use cases: %v\n", err)
		return
	}

	fmt.Println("\n=== Available Use Cases ===")
	for i, uc := range resp.UseCases {
		fmt.Printf("\n%d. %s (ID: %s)\n", i+1, uc.Name, uc.Id)
		fmt.Printf("   Method: %s %s\n", uc.Method, uc.Path)
		fmt.Printf("   Description: %s\n", uc.Description)
	}
}

// Agent Operations

func (c *Client) listAgents() {
	resp, err := c.grpcClient.GetAllAgents(c.ctx, &contracts.GetAllAgentsRequest{})
	if err != nil {
		fmt.Printf("Error getting agents: %v\n", err)
		return
	}

	if len(resp.Agents) == 0 {
		fmt.Println("No agents connected.")
		return
	}

	fmt.Println("\n=== Connected Agents ===")
	for i, agent := range resp.Agents {
		fmt.Printf("\n%d. Agent ID: %s\n", i+1, agent.Id)
		fmt.Printf("   Host: %s:%d\n", agent.Host, agent.Port)
		fmt.Printf("   Metrics Port: %d\n", agent.MetricsPort)
		fmt.Printf("   Enabled: %v\n", agent.Enabled)
		fmt.Printf("   Status: %s\n", agent.Status)
	}
}

func (c *Client) clearAllAgents() {
	fmt.Println("\n=== Clear All Agents ===")
	fmt.Println("WARNING: This will remove all agents from the database.")
	fmt.Println("This is useful for cleaning up agents that didn't shutdown gracefully.")

	confirm := c.readInput("\nAre you sure you want to clear ALL agents? (yes/no): ")
	if strings.ToLower(confirm) != "yes" {
		fmt.Println("Operation cancelled.")
		return
	}

	resp, err := c.grpcClient.ClearAllAgents(c.ctx, &contracts.ClearAllAgentsRequest{})
	if err != nil {
		fmt.Printf("Error clearing agents: %v\n", err)
		return
	}

	if resp.Success {
		fmt.Printf("✓ %s\n", resp.Message)
		if resp.DeletedCount > 0 {
			fmt.Printf("  Cleared %d agent(s) from database\n", resp.DeletedCount)
		} else {
			fmt.Println("  No agents were in the database")
		}
	} else {
		fmt.Printf("✗ Failed to clear agents: %s\n", resp.Message)
	}
}

// Test Execution Operations

func (c *Client) startTestInteractive() {
	fmt.Println("\n=== Start Test Execution ===")

	// Step 1: List and select use case
	resp, err := c.grpcClient.GetUseCases(c.ctx, &contracts.GetUseCasesRequest{})
	if err != nil {
		fmt.Printf("Error getting use cases: %v\n", err)
		return
	}

	fmt.Println("\nAvailable Use Cases:")
	for i, uc := range resp.UseCases {
		fmt.Printf("%d. %s - %s\n", i+1, uc.Id, uc.Name)
	}

	useCaseIdx := c.readInt("\nSelect use case number: ") - 1
	if useCaseIdx < 0 || useCaseIdx >= len(resp.UseCases) {
		fmt.Println("Invalid use case selection.")
		return
	}
	selectedUseCase := resp.UseCases[useCaseIdx]

	// Step 2: List and select target
	targetsResp, err := c.grpcClient.GetAllTargets(c.ctx, &contracts.GetAllTargetsRequest{})
	if err != nil {
		fmt.Printf("Error getting targets: %v\n", err)
		return
	}

	if len(targetsResp.Targets) == 0 {
		fmt.Println("No targets available. Please register a target first.")
		return
	}

	fmt.Println("\nAvailable Targets:")
	for i, target := range targetsResp.Targets {
		fmt.Printf("%d. %s (%s) - %s\n", i+1, target.Name, target.Protocol, target.Address)
	}

	targetIdx := c.readInt("\nSelect target number: ") - 1
	if targetIdx < 0 || targetIdx >= len(targetsResp.Targets) {
		fmt.Println("Invalid target selection.")
		return
	}
	selectedTarget := targetsResp.Targets[targetIdx]

	// Step 3: Configure test parameters
	usersPerAgent := int32(c.readInt("\nSimulated users per agent: "))
	agentSelection := c.readInput("Agent selection (all or comma-separated IDs) [default: all]: ")
	if agentSelection == "" {
		agentSelection = "all"
	}

	// Step 4: Start test execution
	fmt.Println("\n=== Starting Test ===")
	fmt.Printf("Use Case: %s\n", selectedUseCase.Name)
	fmt.Printf("Target: %s (%s)\n", selectedTarget.Name, selectedTarget.Address)
	fmt.Printf("Users per agent: %d\n", usersPerAgent)
	fmt.Printf("Agent selection: %s\n", agentSelection)

	startReq := &contracts.StartTestExecutionRequest{
		UseCaseId:              selectedUseCase.Id,
		TargetId:               selectedTarget.Id,
		SimulatedUsersPerAgent: usersPerAgent,
		AgentSelection:         agentSelection,
	}

	startResp, err := c.grpcClient.StartTestExecution(c.ctx, startReq)
	if err != nil {
		fmt.Printf("Error starting test: %v\n", err)
		return
	}

	if startResp.Success {
		fmt.Printf("\n✓ Test execution started successfully!\n")
		fmt.Printf("Test Execution ID: %s\n", startResp.TestExecutionId)
		fmt.Println("\nUse option 6 to stop this test when ready.")
	} else {
		fmt.Printf("\n✗ Failed to start test: %s\n", startResp.Message)
	}
}

func (c *Client) stopTest() {
	fmt.Println("\n=== Stop Test Execution ===")

	// Get running tests
	resp, err := c.grpcClient.GetRunningTests(c.ctx, &contracts.GetRunningTestsRequest{})
	if err != nil {
		fmt.Printf("Error getting running tests: %v\n", err)
		return
	}

	runningTests := resp.Tests
	if len(runningTests) == 0 {
		fmt.Println("No tests are currently running.")
		return
	}

	// Display running tests
	fmt.Println("\nRunning Tests:")
	fmt.Println("─────────────────────────────────────────────────────────────────")
	for i, test := range runningTests {
		fmt.Printf("%d. Test ID: %s\n", i+1, test.TestExecutionId)
		fmt.Printf("   Use Case: %s (%s)\n", test.UseCaseName, test.UseCaseId)
		fmt.Printf("   Target: %s (%s)\n", test.TargetName, test.TargetId)
		fmt.Printf("   Agents: %d | Users per Agent: %d\n", test.NumberOfAgents, test.SimulatedUsersPerAgent)
		fmt.Printf("   Started: %v\n", time.Unix(test.StartTime, 0).Format("2006-01-02 15:04:05"))
		fmt.Println()
	}
	fmt.Println("─────────────────────────────────────────────────────────────────")

	// Get user selection
	selection := c.readInt(fmt.Sprintf("Enter test number to stop (1-%d) or 0 to cancel: ", len(runningTests)))

	if selection == 0 {
		fmt.Println("Cancelled.")
		return
	}

	if selection < 1 || selection > len(runningTests) {
		fmt.Println("Invalid selection.")
		return
	}

	selectedTest := runningTests[selection-1]

	// Confirm
	confirm := c.readBool(fmt.Sprintf("Stop test '%s' on target '%s'?", selectedTest.UseCaseName, selectedTest.TargetName))
	if !confirm {
		fmt.Println("Cancelled.")
		return
	}

	// Stop the test
	stopResp, err := c.grpcClient.StopTestExecution(c.ctx, &contracts.StopTestExecutionRequest{
		TestExecutionId: selectedTest.TestExecutionId,
	})
	if err != nil {
		fmt.Printf("Error stopping test: %v\n", err)
		return
	}

	if stopResp.Success {
		fmt.Println("✓ Test execution stopped successfully")
	} else {
		fmt.Printf("✗ Failed to stop test: %s\n", stopResp.Message)
	}
}
