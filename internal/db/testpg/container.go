package testpg

import (
	"context"
	"fmt"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const poolSize = 5

// EnvTestUseDocker is the environment variable that enables test container mode.
// When set to "true", NewTransactor in DI will use the shared test container pool.
const EnvTestUseDocker = "TEST_USE_DOCKER"

// IsTestContainerMode returns true if TEST_USE_DOCKER=true.
func IsTestContainerMode() bool {
	return os.Getenv(EnvTestUseDocker) == "true"
}

var (
	pool      [poolSize]*Container
	poolOnce  sync.Once
	poolErr   error
	poolIndex uint64
)

// Config matches testsuite.Config DB fields 1:1
type Config struct {
	DBHost   string
	DBPort   int
	DBUser   string
	DBPass   string
	DBName   string
	DBSchema string
}

// Container wraps the testcontainer and config
type Container struct {
	Container testcontainers.Container
	Config    *Config
}

// GetContainer returns a container from the pool (round-robin).
// Pool is initialized once on first call.
func GetContainer(ctx context.Context) (*Container, error) {
	poolOnce.Do(func() {
		poolErr = initPool(ctx)
	})

	if poolErr != nil {
		return nil, poolErr
	}

	idx := atomic.AddUint64(&poolIndex, 1) % poolSize
	return pool[idx], nil
}

func initPool(ctx context.Context) error {
	var wg sync.WaitGroup
	errs := make(chan error, poolSize)

	for i := 0; i < poolSize; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			c, err := newContainer(ctx)
			if err != nil {
				errs <- fmt.Errorf("container %d: %w", idx, err)
				return
			}
			pool[idx] = c
		}(i)
	}

	wg.Wait()
	close(errs)

	for err := range errs {
		if err != nil {
			return err
		}
	}

	return nil
}

func newContainer(ctx context.Context) (*Container, error) {
	req := testcontainers.ContainerRequest{
		Image:        "postgres:15-alpine",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_USER":     "postgres",
			"POSTGRES_PASSWORD": "postgres",
			"POSTGRES_DB":       "postgres",
		},
		Cmd: []string{"postgres", "-c", "max_connections=500"},
		WaitingFor: wait.ForLog("database system is ready to accept connections").
			WithOccurrence(2).
			WithStartupTimeout(180 * time.Second), // triple occurrence due to init scripts
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to start postgres container: %w", err)
	}

	host, err := container.Host(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get container host: %w", err)
	}

	mappedPort, err := container.MappedPort(ctx, "5432")
	if err != nil {
		return nil, fmt.Errorf("failed to get mapped port: %w", err)
	}

	return &Container{
		Container: container,
		Config: &Config{
			DBHost:   host,
			DBPort:   mappedPort.Int(),
			DBUser:   "postgres",
			DBPass:   "postgres",
			DBName:   "postgres",
			DBSchema: "public",
		},
	}, nil
}

// GetContainerConfig returns the config for a container from the pool.
// Use this when you just need connection params without the container reference.
func GetContainerConfig(ctx context.Context) (*Config, error) {
	c, err := GetContainer(ctx)
	if err != nil {
		return nil, err
	}
	return c.Config, nil
}
