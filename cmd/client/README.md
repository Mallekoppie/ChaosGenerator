# ChaosMaster gRPC Test Client

A comprehensive test client for the ChaosMaster gRPC service that allows you to easily test all service functions through an interactive menu.

## Overview

This client provides a command-line interface to interact with all ChaosMaster service operations:
- Agent management (Get, Update, Delete)
- TestGroup management (Get, Add, Update, Delete)
- TestCollection management (Add, Update, Delete)

## Prerequisites

1. The ChaosMaster server must be running (cmd/master)
2. The server should be accessible at `127.0.0.1:9002` (default configuration)

## Running the Client

To start the test client:

```bash
go run cmd/client/main.go
```

Or build and run:

```bash
go build -o bin/chaos-client cmd/client/main.go
./bin/chaos-client
```

## Features

### Interactive Menu
The client provides an easy-to-use menu system where you can select operations by number:

```
=== ChaosMaster Test Client ===
Agent Operations:
  1. GetAllAgents
  2. UpdateAgent
  3. DeleteAgent

TestGroup Operations:
  4. GetAllTestGroups
  5. GetTestGroup
  6. AddTestGroup
  7. UpdateTestGroup
  8. DeleteTestGroup

TestCollection Operations:
  9. AddTestCollection
 10. UpdateTestCollection
 11. DeleteTestCollection

Other:
 12. Run All Agent Tests
 13. Run All TestGroup Tests
 14. Run All TestCollection Tests
  0. Exit
```

### Individual Operations

#### Agent Operations
- **GetAllAgents (1)**: Lists all registered agents
- **UpdateAgent (2)**: Update an agent's configuration (ID, host, port, status, etc.)
- **DeleteAgent (3)**: Remove an agent by ID

#### TestGroup Operations
- **GetAllTestGroups (4)**: Lists all test groups with their collections
- **GetTestGroup (5)**: Get detailed information about a specific test group
- **AddTestGroup (6)**: Create a new test group
- **UpdateTestGroup (7)**: Modify an existing test group
- **DeleteTestGroup (8)**: Remove a test group by ID

#### TestCollection Operations
- **AddTestCollection (9)**: Add a new test collection to a test group
- **UpdateTestCollection (10)**: Modify an existing test collection
- **DeleteTestCollection (11)**: Remove a test collection by ID

### Batch Test Operations

#### Run All Agent Tests (12)
Executes:
- GetAllAgents to view all registered agents

#### Run All TestGroup Tests (13)
Performs a complete workflow:
1. Creates a new test group with timestamp
2. Lists all test groups
3. Retrieves details of the first test group (if available)

#### Run All TestCollection Tests (14)
Displays available test groups to help you create test collections with valid parent IDs.

## Usage Examples

### Testing Agent Operations

1. Select option `1` to view all agents
2. Select option `2` to update an agent:
   - Enter the agent ID
   - Provide host, port, metrics port
   - Set enabled status (true/false)
   - Enter status string

### Testing TestGroup Workflow

1. Select option `6` to add a new test group:
   - Enter a name: "Performance Tests"
   - Enter description: "Load testing scenarios"
   
2. Select option `4` to see all test groups

3. Select option `5` to get details of a specific group:
   - Enter the ID from step 2

4. Select option `7` to update the test group:
   - Provide the ID and new values

5. Select option `8` to delete when finished:
   - Enter the ID to remove

### Testing TestCollection Workflow

1. First, ensure you have a test group (use option `4` to check)

2. Select option `9` to add a collection:
   - Enter collection name: "API Tests"
   - Enter description: "REST API validation"
   - Enter the parent TestGroup ID from step 1

3. Select option `10` to update the collection:
   - Provide ID and updated values

4. Select option `11` to delete:
   - Enter the collection ID

## Configuration

The client connects to the server at:
- **Address**: `127.0.0.1:9002`
- **TLS**: Disabled (insecure connection)

To modify the connection settings, edit the `serverAddress` variable in `main()`.

## Logging

The client uses the goslow platform logger for all operations:
- Successful operations are logged at INFO level
- Errors are logged at ERROR level
- All logs include contextual information (IDs, counts, etc.)

## Error Handling

The client provides clear error messages for:
- Connection failures
- Invalid requests
- Server-side errors
- Timeouts (default: 10 seconds per operation)

## Tips

1. **Start with Read Operations**: Begin by testing GetAllAgents and GetAllTestGroups to understand your current state
2. **Use Batch Tests**: Options 12-14 provide quick end-to-end testing
3. **Check Logs**: Monitor the console output for detailed operation logs
4. **TestGroup Required**: TestCollections require a valid parent TestGroup ID - create a TestGroup first
5. **IDs Matter**: Most operations require accurate IDs - use Get operations to find them

## Troubleshooting

**Cannot connect to server**
- Ensure the master server is running: `go run cmd/master/main.go`
- Verify the server is listening on port 9002
- Check firewall settings

**Operation timeouts**
- The default timeout is 10 seconds
- Check server logs for performance issues
- Ensure the database is accessible

**Invalid argument errors**
- Verify all required fields are provided
- Check that IDs exist before update/delete operations
- Ensure TestGroup IDs are valid when adding collections
