package dns_simf

import (
	"bufio"
	"cli_window_helper/src/config"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type HostsFileModification struct {
	OldDns []string
	NewDns string
}

type DnsEntryHost struct {
	LineNumber      int
	FullLine        string
	IP              string
	Hostname        string
	LeadingComment  string
	TrailingComment string
	Environment     string
	OriginalLine    int
	GroupedWith     []string
}

var knownEnvironments = []string{"desarrollo", "certificacion"}

// func LoadHostsFile(filePath string, patterns []string) ([]DnsEntryHost, error) {
// 	// Resolve the absolute path
// 	absPath, err := filepath.Abs(filePath)
// 	if err != nil {
// 		return nil, fmt.Errorf("error resolving absolute path: %w", err)
// 	}

// 	// Open the file
// 	file, err := os.Open(absPath)
// 	if err != nil {
// 		return nil, fmt.Errorf("error opening file: %w", err)
// 	}
// 	defer file.Close()

// 	scanner := bufio.NewScanner(file)
// 	lineNumber := 0
// 	var entries []DnsEntryHost

// 	// Regular expression to match IP, hostname, and comments
// 	re := regexp.MustCompile(`^(\s*#?\s*)?([\d\.]+)\s+([\w\.-]+)(.*)$`)

// 	// Process the file line by line
// 	for scanner.Scan() {
// 		lineNumber++
// 		line := scanner.Text()

// 		//fmt.Println("line:", line)

// 		// Check if the line matches the expected format
// 		if matches := re.FindStringSubmatch(line); matches != nil {
// 			//fmt.Println("matches:", matches)
// 			leadingComment := strings.TrimSpace(matches[1])
// 			ip := matches[2]
// 			hostname := matches[3]
// 			trailingPart := strings.TrimSpace(matches[4])

// 			// Check if the hostname matches any of the patterns
// 			for _, pattern := range patterns {

// 				//fmt.Println("pattern:", pattern)
// 				if strings.Contains(hostname, pattern) {
// 					// Split trailing comment if exists
// 					//fmt.Println("Valid:", hostname)
// 					trailingComment := ""
// 					if idx := strings.Index(trailingPart, "#"); idx != -1 {
// 						trailingComment = strings.TrimSpace(trailingPart[idx:])
// 					}

// 					newEntry := DnsEntryHost{
// 						LineNumber:      lineNumber,
// 						FullLine:        line,
// 						IP:              ip,
// 						Hostname:        hostname,
// 						LeadingComment:  leadingComment,
// 						TrailingComment: trailingComment,
// 					}
// 					newEntry.DetectEnvironment()

// 					entries = append(entries, newEntry)
// 					break
// 				}

// 			}
// 		}
// 	}

// 	if err := scanner.Err(); err != nil {
// 		return nil, fmt.Errorf("error scanning file: %w", err)
// 	}

// 	return entries, nil
// }

func LoadHostsFile(filePath string, patterns []string) ([]DnsEntryHost, error) {
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return nil, fmt.Errorf("error resolving absolute path: %w", err)
	}

	file, err := os.Open(absPath)
	if err != nil {
		return nil, fmt.Errorf("error opening file: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineNumber := 0
	var entries []DnsEntryHost

	// Modified regex to capture multiple hostnames
	re := regexp.MustCompile(`^(\s*#?\s*)?([\d\.]+)\s+([\w\.-]+(?:\s+[\w\.-]+)*)(.*)$`)

	for scanner.Scan() {
		lineNumber++
		line := scanner.Text()

		if matches := re.FindStringSubmatch(line); matches != nil {
			leadingComment := strings.TrimSpace(matches[1])
			ip := matches[2]
			hostnamesStr := matches[3]
			trailingPart := strings.TrimSpace(matches[4])

			// Split multiple hostnames
			hostnames := strings.Fields(hostnamesStr)

			// If there are multiple hostnames, keep track of them
			var groupedHostnames []string
			if len(hostnames) > 1 {
				groupedHostnames = hostnames
			}

			// Process each hostname in the line
			for _, hostname := range hostnames {
				for _, pattern := range patterns {
					if strings.Contains(hostname, pattern) {
						trailingComment := ""
						if idx := strings.Index(trailingPart, "#"); idx != -1 {
							trailingComment = strings.TrimSpace(trailingPart[idx:])
						}

						newEntry := DnsEntryHost{
							LineNumber:      lineNumber,
							IP:              ip,
							Hostname:        hostname,
							LeadingComment:  leadingComment,
							TrailingComment: trailingComment,
							OriginalLine:    lineNumber,
							GroupedWith:     groupedHostnames,
						}

						// Set FullLine based on whether it was part of a group
						if len(groupedHostnames) > 0 {
							// For grouped entries, keep the original format
							if hostname == groupedHostnames[0] {
								newEntry.FullLine = fmt.Sprintf("%s%s\t%s\t%s",
									newEntry.LeadingComment,
									newEntry.IP,
									strings.Join(groupedHostnames, " "),
									newEntry.TrailingComment)
							}
						} else {
							newEntry.FullLine = fmt.Sprintf("%s%s\t%s\t%s",
								newEntry.LeadingComment,
								newEntry.IP,
								newEntry.Hostname,
								newEntry.TrailingComment)
						}

						newEntry.DetectEnvironment()
						entries = append(entries, newEntry)
						break
					}
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error scanning file: %w", err)
	}

	return entries, nil
}

func (d *DnsEntryHost) LoadFullLine() {
	d.FullLine = fmt.Sprintf("%s%s\t%s\t%s", d.LeadingComment, d.IP, d.Hostname, d.TrailingComment)
}

func (d *DnsEntryHost) DetectEnvironment() {
	parts := strings.Split(d.Hostname, ".")
	for _, part := range parts {
		for _, env := range knownEnvironments {
			if part == env {
				d.Environment = env
				return
			}
		}
	}
	d.Environment = "produccion"
}
func (d *DnsEntryHost) PrintInfo() {
	fmt.Println("--------------------------------")
	fmt.Printf("DNS Entry Host Information:\n")
	fmt.Printf("  Line Number: %d\n", d.LineNumber)
	fmt.Printf("  IP Address:  %s\n", d.IP)
	fmt.Printf("  Hostname:    %s\n", d.Hostname)
	fmt.Printf("  Environment: %s\n", d.Environment)
	fmt.Printf("  Full Line:   %s\n", d.FullLine)

	if d.LeadingComment != "" {
		fmt.Printf("  Leading Comment:\n    %s\n", d.LeadingComment)
	}

	if d.TrailingComment != "" {
		fmt.Printf("  Trailing Comment:\n    %s\n", d.TrailingComment)
	}

	fmt.Println("--------------------------------") // Add an empty line for better separation between entries
}

func DetectEnvironmentOfList(entries []DnsEntryHost) (string, bool) {
	environmentCount := make(map[string]int)

	for _, entry := range entries {
		if entry.Environment != "" {
			environmentCount[entry.Environment]++
		}
	}

	if len(environmentCount) == 0 {
		return "", false
	}

	if len(environmentCount) == 1 {
		for env := range environmentCount {
			return env, false
		}
	}

	return "", true
}
func DetectEnvironment() (string, bool) {

	config.GetHostPath()

	entries, err := LoadHostsFile(config.GetHostPath(), []string{"simf", "grafana", "lbtr", "argus"})
	if err != nil {
		panic(err)
	}

	return DetectEnvironmentOfList(entries)
}
