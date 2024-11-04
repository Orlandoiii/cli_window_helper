package appsettings_simf

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

// ModifyJSONFile reads a JSON file, modifies a specific value at a given path,
// and writes the modified JSON back to the file.
func ChangeDnsAppSettings(filename string, mapFields map[string]string) error {

	absPath, err := filepath.Abs(filename)
	if err != nil {
		return fmt.Errorf("error resolving absolute path: %w", err)
	}

	// Check if the file exists
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		return fmt.Errorf("file does not exist: %s", absPath)
	}

	// Read the JSON file
	data, err := os.ReadFile(absPath)
	if err != nil {
		return fmt.Errorf("error reading file: %w", err)
	}

	// Get the current value at the specified path
	modifiedJSON := string(data)
	for field, path := range mapFields {

		currentValue := gjson.Get(modifiedJSON, path)

		if currentValue.Value() == nil {
			fmt.Printf("field %s not found\n", path)
			continue
		}

		// Modify the JSON
		var newValue string

		if field == DATA_BASE_FIELD {
			newValue, err = changePostgresHost(currentValue.String())
			if err != nil {
				return fmt.Errorf("error changing postgres host: %w", err)
			}
		}
		if field == KAFKA_FIELD {
			newValue, err = changeKafkaBootstrap(currentValue.String())
			if err != nil {
				return fmt.Errorf("error changing kafka bootstrap: %w", err)
			}
		} else if contains(ENDPOINTS_CORE, field) {
			newValue, err = changeEndpointIP(currentValue.String())
			if err != nil {
				return fmt.Errorf("error changing endpoint IP: %w", err)
			}
		}

		modifiedJSON, err = sjson.Set(modifiedJSON, path, newValue)
		if err != nil {
			return fmt.Errorf("error modifying JSON: %w", err)
		}
	}

	// Write the modified JSON back to the file
	err = os.WriteFile(filename, []byte(modifiedJSON), os.ModePerm)
	if err != nil {
		return fmt.Errorf("error writing file: %w", err)
	}

	return nil
}
func trimServerPrefix(host string) string {
	if strings.Contains(host, ".") && strings.Contains(host, "server") {
		return strings.TrimPrefix(host, "server.")
	}
	return host
}

func changePostgresHost(connectionString string) (string, error) {
	// Split the connection string into parts
	parts := strings.Split(connectionString, ";")

	for i, part := range parts {
		kv := strings.SplitN(part, "=", 2)
		if len(kv) == 2 && strings.TrimSpace(kv[0]) == "Host" {
			// Use the helper function to trim the server prefix
			host := trimServerPrefix(strings.TrimSpace(kv[1]))
			parts[i] = fmt.Sprintf("Host=%s", host)
			break
		}
	}

	// Reconstruct the connection string
	newConnectionString := strings.Join(parts, ";")

	return newConnectionString, nil
}

func changeKafkaBootstrap(bootstrap string) (string, error) {
	servers := strings.Split(bootstrap, ",")
	var newServers []string

	for _, server := range servers {
		parts := strings.SplitN(server, ":", 2)
		host := trimServerPrefix(parts[0])

		if len(parts) > 1 {
			newServers = append(newServers, host+":"+parts[1])
		} else {
			newServers = append(newServers, host)
		}
	}

	return strings.Join(newServers, ","), nil
}

// Simplified function to change only the host in endpoint URLs
func changeEndpointIP(urlString string) (string, error) {
	// Remove the leading '@' if present

	// Parse the URL
	parsedURL, err := url.Parse(urlString)
	if err != nil {
		return "", fmt.Errorf("error parsing URL: %w", err)
	}

	// Change only the host part
	newHost := "localhost" // Replace with the desired new IP or hostname

	if parsedURL.Port() != "" {
		parsedURL.Host = fmt.Sprintf("%s:%s", newHost, parsedURL.Port())
	} else {
		parsedURL.Host = newHost
	}

	// Reconstruct the URL string and add back the '@' prefix
	return parsedURL.String(), nil
}

func changeEnvironmentInHost(host, newEnvironment string) string {
	// Remove any trailing dots
	host = strings.TrimSuffix(host, ".")

	// Split the host into parts
	parts := strings.Split(host, ".")

	if len(parts) < 2 {
		return host // Return unchanged if host is too short
	}

	// If it's production environment, remove the environment part
	if newEnvironment == PRODUCCION {
		// Check if there's an environment part to remove
		for i, part := range parts {
			if part == DESARROLLO || part == CERTIFICACION {
				// Remove the environment part
				return strings.Join(append(parts[:i], parts[i+1:]...), ".")
			}
		}
		return host // No environment found, return unchanged
	}

	// For desarrollo and certificacion
	// Check if it already has an environment
	for i, part := range parts {
		if part == DESARROLLO || part == CERTIFICACION || part == PRODUCCION {
			// Replace existing environment
			parts[i] = newEnvironment
			return strings.Join(parts, ".")
		}
	}

	// No environment found, insert before the last part
	newParts := append(parts[:len(parts)-1], newEnvironment, parts[len(parts)-1])
	return strings.Join(newParts, ".")
}

func changeKafkaBootstrapEnvironment(bootstrap, newEnvironment string) (string, error) {
	servers := strings.Split(bootstrap, ",")
	var newServers []string

	for _, server := range servers {
		parts := strings.SplitN(server, ":", 2)
		host := trimServerPrefix(parts[0])

		// Change environment in the host
		newHost := changeEnvironmentInHost(host, newEnvironment)

		if len(parts) > 1 {
			newServers = append(newServers, newHost+":"+parts[1])
		} else {
			newServers = append(newServers, newHost)
		}
	}

	return strings.Join(newServers, ","), nil
}

func changePostgresHostEnvironment(connectionString, newEnvironment string) (string, error) {
	// Split the connection string into parts
	parts := strings.Split(connectionString, ";")

	for i, part := range parts {
		kv := strings.SplitN(part, "=", 2)
		if len(kv) == 2 && strings.TrimSpace(kv[0]) == "Host" {
			// Use the helper function to trim the server prefix
			host := trimServerPrefix(strings.TrimSpace(kv[1]))

			// Change environment in the host
			newHost := changeEnvironmentInHost(host, newEnvironment)

			parts[i] = fmt.Sprintf("Host=%s", newHost)
			break
		}
	}

	// Reconstruct the connection string
	newConnectionString := strings.Join(parts, ";")

	return newConnectionString, nil
}

// Helper function to check if a slice contains a string
func contains(slice []string, str string) bool {
	for _, v := range slice {
		if v == str {
			return true
		}
	}
	return false
}
