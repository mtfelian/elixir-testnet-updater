package delixir

import (
	"bufio"
	"os"
	"strings"
)

// EnvConfig represents environment configuration
type EnvConfig struct {
	DisplayName     string
	BeneficiaryAddr string
	PrivateKey      string
}

// ParseEnvFile reads the env file and returns a slice of environment variables
func ParseEnvFile(filePath string) ([]string, EnvConfig, error) {
	var envConfig EnvConfig
	file, err := os.Open(filePath)
	if err != nil {
		return nil, envConfig, err
	}
	defer file.Close()

	var envVars []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			// Skip empty lines and comments
			continue
		}
		envVars = append(envVars, line)
	}

	if err := scanner.Err(); err != nil {
		return nil, envConfig, err
	}

	for _, keyValue := range envVars {
		parts := strings.Split(keyValue, "=")
		parts[0], parts[1] = strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
		switch parts[0] {
		case "STRATEGY_EXECUTOR_DISPLAY_NAME":
			envConfig.DisplayName = parts[1]
		case "STRATEGY_EXECUTOR_BENEFICIARY":
			envConfig.BeneficiaryAddr = parts[1]
		case "SIGNER_PRIVATE_KEY":
			envConfig.PrivateKey = parts[1]
		}
	}

	return envVars, envConfig, nil
}
