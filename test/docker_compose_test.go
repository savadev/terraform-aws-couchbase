package test

import (
	"testing"
	"fmt"
	"path/filepath"
	"github.com/gruntwork-io/terratest/test-structure"
	terralog "github.com/gruntwork-io/terratest/log"
	"github.com/gruntwork-io/terratest/files"
	"log"
	"github.com/gruntwork-io/terratest/shell"
)

func TestUnitCouchbaseSingleClusterUbuntuInDocker(t *testing.T) {
	t.Parallel()
	testCouchbaseInDocker(t, "TestUnitCouchbaseSingleClusterUbuntuInDocker","ubuntu")
}

func testCouchbaseInDocker(t *testing.T, testName string, osName string) {
	logger := terralog.NewLogger(testName)

	tmpRootDir, err := files.CopyTerraformFolderToTemp("../", testName)
	if err != nil {
		t.Fatal(err)
	}
	couchbaseAmiDir := filepath.Join(tmpRootDir, "examples", "couchbase-ami")
	couchbaseSingleClusterDockerDir := filepath.Join(tmpRootDir, "examples", "couchbase-single-cluster", "local-test")

	test_structure.RunTestStage("setup_image", logger, func() {
		buildCouchbaseWithPacker(t, logger, fmt.Sprintf("%s-docker", osName), "us-east-1", couchbaseAmiDir)
	})

	test_structure.RunTestStage("setup_docker", logger, func() {
		startCouchbaseWithDockerCompose(t, osName, couchbaseSingleClusterDockerDir, logger)
	})

	defer test_structure.RunTestStage("teardown", logger, func() {
		getDockerComposeLogs(t, couchbaseSingleClusterDockerDir, logger)
		stopCouchbaseWithDockerCompose(t, couchbaseSingleClusterDockerDir, logger)
	})

	test_structure.RunTestStage("validation", logger, func() {
		consoleUrl := fmt.Sprintf("http://localhost:%d", testWebConsolePorts[osName])
		checkCouchbaseConsoleIsRunning(t, consoleUrl, logger)

		dataNodesUrl := fmt.Sprintf("http://%s:%s@localhost:%d", usernameForTest, passwordForTest, testWebConsolePorts[osName])
		checkCouchbaseClusterIsInitialized(t, dataNodesUrl, logger)
		checkCouchbaseDataNodesWorking(t, dataNodesUrl, logger)

		syncGatewayUrl := fmt.Sprintf("http://localhost:%d/mock-couchbase-asg", testSyncGatewayPorts[osName])
		checkSyncGatewayWorking(t, syncGatewayUrl, logger)
	})
}

func startCouchbaseWithDockerCompose(t *testing.T, os string, exampleDir string, logger *log.Logger) {
	cmd := shell.Command{
		Command:    "docker-compose",
		Args:       []string{"up", "-d"},
		WorkingDir: exampleDir,
	}

	if err := shell.RunCommand(cmd, logger); err != nil {
		t.Fatalf("Failed to start Couchbase using Docker Compose: %v", err)
	}
}

func getDockerComposeLogs(t *testing.T, exampleDir string, logger *log.Logger) {
	logger.Printf("Fetching docker-compse logs:")

	cmd := shell.Command{
		Command:    "docker-compose",
		Args:       []string{"logs"},
		WorkingDir: exampleDir,
	}

	if err := shell.RunCommand(cmd, logger); err != nil {
		t.Fatalf("Failed to get Docker Compose logs: %v", err)
	}
}

func stopCouchbaseWithDockerCompose(t *testing.T, exampleDir string, logger *log.Logger) {
	cmd := shell.Command{
		Command:    "docker-compose",
		Args:       []string{"down"},
		WorkingDir: exampleDir,
	}

	if err := shell.RunCommand(cmd, logger); err != nil {
		t.Fatalf("Failed to stop Couchbase using Docker Compose: %v", err)
	}
}
