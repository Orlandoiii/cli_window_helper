package dns_simf

import (
	"bufio"
	"cli_window_helper/src/config"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// var NewDnsDictionary = make(map[string]string)

// func init() {

// 	NewDnsDictionary["argus_api"] = "argus_api.simf"
// 	NewDnsDictionary["argus_api.historico"] = "argus_api.historico.simf"

// 	NewDnsDictionary["grafana"] = "grafana.simf"

// 	NewDnsDictionary["instalacion.argus"] = "instalacion.argus.simf"

// 	NewDnsDictionary["server.discovery"] = "discovery.simf"

// 	NewDnsDictionary["server.kafka1"] = "kafka1.simf"
// 	NewDnsDictionary["server.kafka2"] = "kafka2.simf"
// 	NewDnsDictionary["server.kafka3"] = "kafka3.simf"

// 	NewDnsDictionary["server.postgres"] = "postgres.simf"
// 	NewDnsDictionary["server.postgres.replica"] = "postgres.replica.simf"

// 	NewDnsDictionary["server.prometheus"] = "prometheus.simf"

// 	NewDnsDictionary["server.core"] = "core.simf"

// }

func ProcessHostsFileReplaceJustPattern(filePath string, modifications map[string]string) error {
	// Resolve the absolute path
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return fmt.Errorf("error resolving absolute path: %w", err)
	}

	// Get the directory of the original file
	dir := filepath.Dir(absPath)

	// Read the entire file
	content, err := os.ReadFile(absPath)
	if err != nil {
		return fmt.Errorf("error reading file: %w", err)
	}

	// Create a temporary file in the same directory
	tempFile, err := os.CreateTemp(dir, "hosts_temp_*")
	if err != nil {
		return fmt.Errorf("error creating temporary file: %w", err)
	}
	defer os.Remove(tempFile.Name())

	scanner := bufio.NewScanner(strings.NewReader(string(content)))
	writer := bufio.NewWriter(tempFile)

	// Process the file line by line
	for scanner.Scan() {
		line := scanner.Text()
		modifiedLine := line

		// Check if the line matches any of the modification patterns
		for pattern, replacement := range modifications {
			if strings.Contains(line, pattern) {
				modifiedLine = strings.Replace(line, pattern, replacement, -1)
				break
			}
		}

		// Write the modified (or original) line to the temporary file
		_, err := writer.WriteString(modifiedLine + "\n")
		if err != nil {
			return fmt.Errorf("error writing to temporary file: %w", err)
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error scanning file: %w", err)
	}

	// Flush the writer to ensure all data is written to the file
	if err := writer.Flush(); err != nil {
		return fmt.Errorf("error flushing temporary file: %w", err)
	}

	// Close the temporary file
	if err := tempFile.Close(); err != nil {
		return fmt.Errorf("error closing temporary file: %w", err)
	}

	// Replace the original file with the temporary file
	if err := os.Rename(tempFile.Name(), absPath); err != nil {
		return fmt.Errorf("error replacing original file: %w", err)
	}

	return nil
}

func ProcessHostsFile(filePath string, modifications map[string]string, replaceIP bool) error {
	// Resolve the absolute path
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return fmt.Errorf("error resolving absolute path: %w", err)
	}

	// Get the directory of the original file
	dir := filepath.Dir(absPath)

	// Read the entire file
	content, err := os.ReadFile(absPath)
	if err != nil {
		return fmt.Errorf("error reading file: %w", err)
	}

	// Create a temporary file in the same directory
	tempFile, err := os.CreateTemp(dir, "hosts_temp_*")
	if err != nil {
		return fmt.Errorf("error creating temporary file: %w", err)
	}
	defer os.Remove(tempFile.Name())

	scanner := bufio.NewScanner(strings.NewReader(string(content)))
	writer := bufio.NewWriter(tempFile)

	re := regexp.MustCompile(`^(\s*#?\s*)?([\d\.]+)(\s+)([\w\.-]+)(.*)$`)

	// Process the file line by line
	for scanner.Scan() {
		line := scanner.Text()
		modifiedLine := line

		// Check if the line matches the expected format
		if matches := re.FindStringSubmatch(line); matches != nil {
			comment := matches[1]
			ip := matches[2]
			separator := matches[3]
			hostname := matches[4]
			rest := matches[5]

			// Check if the hostname matches any of the modification patterns
			for pattern, replacement := range modifications {
				if strings.Contains(hostname, pattern) {
					newHostname := strings.Replace(hostname, pattern, replacement, -1)
					if replaceIP {
						// Replace both IP and hostname
						modifiedLine = fmt.Sprintf("%s%s%s%s%s", comment, ip, separator, newHostname, rest)
					} else {
						// Replace only the hostname
						modifiedLine = fmt.Sprintf("%s%s%s%s%s", comment, ip, separator, newHostname, rest)
					}
					break
				}
			}
		}

		// Write the modified (or original) line to the temporary file
		_, err := writer.WriteString(modifiedLine + "\n")
		if err != nil {
			return fmt.Errorf("error writing to temporary file: %w", err)
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error scanning file: %w", err)
	}

	// Flush the writer to ensure all data is written to the file
	if err := writer.Flush(); err != nil {
		return fmt.Errorf("error flushing temporary file: %w", err)
	}

	// Close the temporary file
	if err := tempFile.Close(); err != nil {
		return fmt.Errorf("error closing temporary file: %w", err)
	}

	// Replace the original file with the temporary file
	if err := os.Rename(tempFile.Name(), absPath); err != nil {
		return fmt.Errorf("error replacing original file: %w", err)
	}

	return nil
}

func findDnsEntry(entries []DnsEntryHost, hostname string) *DnsEntryHost {
	for i := range entries {
		if entries[i].Hostname == hostname {
			return &entries[i]
		}
	}
	return nil
}

func AddArgusDns(environment string) error {
	//hostPath := config.GetHostPath() // Replace with config.GetHostPath() in production
	hostPath := config.GetHostPath()

	patterns := []string{"argus", "simf", "lbtr", "grafana"}

	main_dns_list, err := LoadHostsFile(hostPath, patterns)
	if err != nil {
		return err
	}

	for _, dns := range main_dns_list {

		if dns.Environment != environment {
			panic(fmt.Sprintf("el ambiente elegido no coincide para %v", dns))
		}
	}

	new_dns_list, err := LoadHostsFile("./new_dns.txt", patterns)
	if err != nil {
		return err
	}

	for _, dns := range new_dns_list {

		if dns.Environment != environment {
			panic(fmt.Sprintf("el ambiente elegido no coincide para %v", dns))
		}
	}

	updatedEntries := make(map[int]DnsEntryHost)
	var entriesToAdd []DnsEntryHost

	// Compare new_dns_list with main_dns_list
	for _, newEntry := range new_dns_list {
		existingEntry := findDnsEntry(main_dns_list, newEntry.Hostname)

		if existingEntry != nil {
			//fmt.Println("Existing Entry", existingEntry.Hostname)
			if existingEntry.IP != newEntry.IP {
				// IP mismatch, update the existing entry

				//fmt.Println("Mistmacth Ip", existingEntry.IP, newEntry.IP)

				updatedEntry := *existingEntry

				updatedEntry.IP = newEntry.IP

				updatedEntry.FullLine = fmt.Sprintf("%s%s %s%s",
					updatedEntry.LeadingComment,
					newEntry.IP,
					updatedEntry.Hostname,
					updatedEntry.TrailingComment)

				updatedEntries[existingEntry.LineNumber] = updatedEntry
			}
			// If IP matches, do nothing (entry already exists)
		} else {
			// New hostname, add it to entriesToAdd
			entriesToAdd = append(entriesToAdd, newEntry)
		}
	}

	// Update the hosts file
	err = updateHostsFile(hostPath, updatedEntries, entriesToAdd)
	if err != nil {
		return err
	}

	return nil
}

func RewriteOldDNS(environment string) error {
	hostPath := config.GetHostPath()

	patterns := []string{"argus", "simf", "lbtr", "grafana"}

	main_dns_list, err := LoadHostsFile(hostPath, patterns)
	if err != nil {
		return err
	}

	content, err := os.ReadFile(hostPath)
	if err != nil {
		return err
	}

	if strings.Contains(string(content), "-----SIMF DNS VIEJAS INCORPORADAS CC27-----") {
		return nil
	}

	lines := strings.Split(string(content), "\n")

	// Create a map of hostnames to remove
	hostnamesMap := make(map[string]DnsEntryHost)
	for _, entry := range main_dns_list {
		hostnamesMap[entry.Hostname] = entry
	}

	// Regular expression to match IP, hostname, and comments (same as in LoadHostsFile)
	re := regexp.MustCompile(`^(\s*#?\s*)?([\d\.]+)\s+([\w\.-]+)(.*)$`)

	var newLines []string
	var removedEntries []DnsEntryHost
	lastValidLineIndex := -1

	// Filter out lines that match entries in main_dns_list and keep track of the last valid line
	for _, line := range lines {
		trimmedLine := strings.TrimSpace(line)
		if trimmedLine != "" {
			if matches := re.FindStringSubmatch(line); matches != nil {
				hostname := matches[3]
				if entry, exists := hostnamesMap[hostname]; exists {
					removedEntries = append(removedEntries, entry)
				} else {
					newLines = append(newLines, line)
					lastValidLineIndex = len(newLines) - 1
				}
			} else {
				newLines = append(newLines, line)
				if !strings.HasPrefix(trimmedLine, "#") {
					lastValidLineIndex = len(newLines) - 1
				}
			}
		}
	}

	// Trim empty lines at the end
	if lastValidLineIndex >= 0 {
		newLines = newLines[:lastValidLineIndex+1]
	}

	// Sort removedEntries by hostname
	sort.Slice(removedEntries, func(i, j int) bool {
		return removedEntries[i].Hostname < removedEntries[j].Hostname
	})

	// Append removed entries to the end of the file
	writtenLines := make(map[int]bool)

	// Append removed entries to the end of the file
	if len(removedEntries) > 0 {
		newLines = append(newLines, "", "# -----SIMF DNS VIEJAS INCORPORADAS CC27-----")
		for _, entry := range removedEntries {
			// Skip if this entry was part of a group and not the first one
			if len(entry.GroupedWith) > 0 && entry.Hostname != entry.GroupedWith[0] {
				continue
			}

			newLines = append(newLines, entry.FullLine, "") // Add an empty line after each entry
			writtenLines[entry.OriginalLine] = true
		}
		newLines = append(newLines, "# -----SIMF DNS VIEJAS INCORPORADAS CC27-----")
	}

	// Remove the last empty line if it exists
	if len(newLines) > 0 && newLines[len(newLines)-1] == "" {
		newLines = newLines[:len(newLines)-1]
	}

	// Write the updated content back to the file
	err = os.WriteFile(hostPath, []byte(strings.Join(newLines, "\n")), 0644)
	if err != nil {
		return err
	}

	fmt.Println("Hosts file updated successfully.")
	return nil
}

// func AddNewDns(environment string) error {
// 	hostPath := "./host"
// 	patterns := []string{"argus", "simf", "lbtr", "grafana"}

// 	fileContent, err := os.ReadFile(hostPath)
// 	if err != nil {
// 		return err
// 	}

// 	if strings.Contains(string(fileContent), "-----NUEVOS DNS SIMF CC27-----") {
// 		return nil
// 	}

// 	// Load existing entries
// 	dnsEntries, err := LoadHostsFile(hostPath, patterns)
// 	if err != nil {
// 		return err
// 	}

// 	var entriesToAdd []DnsEntryHost

// 	// Process all entries
// 	for _, entry := range dnsEntries {
// 		if strings.Contains(entry.Hostname, "rest_api") ||
// 			strings.Contains(entry.Hostname, "dashboard") {
// 			continue
// 		}

// 		// Create a new entry based on the original
// 		newEntry := DnsEntryHost{
// 			IP:              entry.IP,
// 			LeadingComment:  entry.LeadingComment,
// 			TrailingComment: entry.TrailingComment,
// 		}

// 		hostname := entry.Hostname

// 		// Remove server. prefix if it exists
// 		if strings.HasPrefix(hostname, "server.") {
// 			hostname = strings.TrimPrefix(hostname, "server.")
// 		} else if strings.HasPrefix(hostname, "instalacion.") {
// 			hostname = strings.TrimPrefix(hostname, "instalacion.")
// 		}

// 		// Handle suffix for all entries
// 		parts := strings.Split(hostname, ".")
// 		if len(parts) > 0 {
// 			suffix := parts[len(parts)-1]
// 			// Remove current suffix
// 			hostname = strings.TrimSuffix(hostname, suffix)
// 			hostname = strings.TrimSuffix(hostname, ".")

// 			// Add appropriate suffix
// 			if suffix != "simf" && suffix != "lbtr" {
// 				hostname = hostname + ".simf"
// 			} else {
// 				hostname = hostname + "." + suffix
// 			}
// 		}

// 		newEntry.Hostname = hostname
// 		newEntry.FullLine = fmt.Sprintf("%s%s	%s	%s",
// 			newEntry.LeadingComment,
// 			newEntry.IP,
// 			newEntry.Hostname,
// 			newEntry.TrailingComment)

// 		entriesToAdd = append(entriesToAdd, newEntry)
// 	}

// 	// Add section markers
// 	content := "\n# -----NUEVOS DNS SIMF CC27-----\n"
// 	for _, entry := range entriesToAdd {
// 		content += entry.FullLine + "\n"
// 	}
// 	content += "# -----NUEVOS DNS SIMF CC27-----\n"

// 	// Append the new section to the file
// 	file, err := os.OpenFile(hostPath, os.O_APPEND|os.O_WRONLY, 0644)
// 	if err != nil {
// 		return err
// 	}
// 	defer file.Close()

// 	_, err = file.WriteString(content)
// 	if err != nil {
// 		return err
// 	}

// 	return nil
// }

func AddNewDns(environment string) error {
	hostPath := config.GetHostPath()
	patterns := []string{"argus", "simf", "lbtr", "grafana"}

	fileContent, err := os.ReadFile(hostPath)
	if err != nil {
		return err
	}

	if strings.Contains(string(fileContent), "-----NUEVOS DNS SIMF CC27-----") {
		return nil
	}

	// Load existing entries
	dnsEntries, err := LoadHostsFile(hostPath, patterns)
	if err != nil {
		return err
	}

	// Group entries by IP and create a map of entries that should be grouped
	groupedEntries := make(map[string][]DnsEntryHost)
	for _, entry := range dnsEntries {
		if strings.Contains(entry.Hostname, "rest_api") ||
			strings.Contains(entry.Hostname, "dashboard") {
			continue
		}

		// Create new hostname
		hostname := entry.Hostname
		if strings.HasPrefix(hostname, "server.") {
			hostname = strings.TrimPrefix(hostname, "server.")
		} else if strings.HasPrefix(hostname, "instalacion.") {
			hostname = strings.TrimPrefix(hostname, "instalacion.")
		}

		// Handle suffix for all entries
		parts := strings.Split(hostname, ".")
		if len(parts) > 0 {
			suffix := parts[len(parts)-1]
			hostname = strings.TrimSuffix(hostname, suffix)
			hostname = strings.TrimSuffix(hostname, ".")

			if suffix != "simf" && suffix != "lbtr" {
				hostname = hostname + ".simf"
			} else {
				hostname = hostname + "." + suffix
			}
		}

		// Create new entry
		newEntry := DnsEntryHost{
			IP:              entry.IP,
			Hostname:        hostname,
			LeadingComment:  entry.LeadingComment,
			TrailingComment: entry.TrailingComment,
		}

		// Group entries by IP for kafka servers
		if strings.Contains(hostname, "kafka") {
			groupedEntries[entry.IP] = append(groupedEntries[entry.IP], newEntry)
		} else {
			// For non-kafka entries, create individual entries
			groupedEntries[entry.IP+hostname] = []DnsEntryHost{newEntry}
		}
	}

	// Create content with blank lines between entries
	content := "\n# -----NUEVOS DNS SIMF CC27-----\n\n"

	// Process grouped entries
	for _, entries := range groupedEntries {
		if len(entries) > 1 && strings.Contains(entries[0].Hostname, "kafka") {
			// Multiple kafka entries - combine them
			hostnames := make([]string, len(entries))
			for i, entry := range entries {
				hostnames[i] = entry.Hostname
			}
			content += fmt.Sprintf("%s%s\t%s\t%s\n\n",
				entries[0].LeadingComment,
				entries[0].IP,
				strings.Join(hostnames, " "),
				entries[0].TrailingComment)
		} else {
			// Single entry
			content += fmt.Sprintf("%s%s\t%s\t%s\n\n",
				entries[0].LeadingComment,
				entries[0].IP,
				entries[0].Hostname,
				entries[0].TrailingComment)
		}
	}

	content += "# -----NUEVOS DNS SIMF CC27-----\n"

	// Append the new section to the file
	file, err := os.OpenFile(hostPath, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString(content)
	return err
}
func updateHostsFile(filePath string, updatedEntries map[int]DnsEntryHost, entriesToAdd []DnsEntryHost) error {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	lines := strings.Split(string(content), "\n")

	// Update existing entries
	for lineNumber, entry := range updatedEntries {
		if lineNumber > 0 && lineNumber <= len(lines) {
			lines[lineNumber-1] = entry.FullLine
		}
	}

	// Append new entries
	for _, entry := range entriesToAdd {
		newLine := fmt.Sprintf("%s%s  %s  %s", entry.LeadingComment, entry.IP, entry.Hostname, entry.TrailingComment)
		lines = append(lines, newLine)
	}

	// Write the updated content back to the file
	return os.WriteFile(filePath, []byte(strings.Join(lines, "\n")), 0644)
}

// Add this method to DnsEntryHost
func ChangeEnvironment(currentEnv, newEnv string) error {
	hostPath := config.GetHostPath()
	patterns := []string{"argus", "simf", "lbtr", "grafana"}

	// Load existing entries
	entries, err := LoadHostsFile(hostPath, patterns)
	if err != nil {
		return err
	}

	// Verify all entries are from the same environment
	for _, entry := range entries {
		if entry.Environment != currentEnv {
			return fmt.Errorf("found entry with different environment: %s (expected %s)", entry.Environment, currentEnv)
		}
	}

	// Create a map to track processed groups
	processedGroups := make(map[int]bool)

	// Change environment for all entries
	modifiedEntries := make([]DnsEntryHost, 0, len(entries))
	for _, entry := range entries {
		// Skip if we've already processed this group
		if len(entry.GroupedWith) > 0 && processedGroups[entry.OriginalLine] {
			continue
		}

		modifiedEntry := entry
		modifiedEntry.ChangeEnvironment(newEnv)
		modifiedEntries = append(modifiedEntries, modifiedEntry)

		// Mark this group as processed
		if len(entry.GroupedWith) > 0 {
			processedGroups[entry.OriginalLine] = true
		}
	}

	// Create a map of line numbers to modified entries
	updatedEntries := make(map[int]DnsEntryHost)
	for _, entry := range modifiedEntries {
		updatedEntries[entry.LineNumber] = entry
	}

	// Read the current file content
	content, err := os.ReadFile(hostPath)
	if err != nil {
		return err
	}

	lines := strings.Split(string(content), "\n")
	var newLines []string

	// Process each line
	for i, line := range lines {
		if entry, exists := updatedEntries[i+1]; exists {
			newLines = append(newLines, entry.FullLine)
		} else {
			newLines = append(newLines, line)
		}
	}

	// Write the updated content back to the file
	return os.WriteFile(hostPath, []byte(strings.Join(newLines, "\n")), 0644)
}
func AppendDNSEntries(newEntries []DnsEntryHost, message string) error {
	hostPath := config.GetHostPath()
	patterns := []string{"argus", "simf", "lbtr", "grafana"}

	content, err := os.ReadFile(hostPath)
	if err != nil {
		return err
	}

	if strings.Contains(string(content), message) {
		return nil
	}

	// Load existing entries
	existingEntries, err := LoadHostsFile(hostPath, patterns)
	if err != nil {
		return err
	}

	// Create a map of existing hostnames for quick lookup
	existingHostnames := make(map[string]bool)
	for _, entry := range existingEntries {
		existingHostnames[entry.Hostname] = true
	}

	// Filter out entries that already exist
	var uniqueEntries []DnsEntryHost
	for _, entry := range newEntries {
		if !existingHostnames[entry.Hostname] {
			uniqueEntries = append(uniqueEntries, entry)
		}
	}

	// If no new unique entries, return early
	if len(uniqueEntries) == 0 {
		fmt.Println("All DNS entries already exist. Skipping...")
		return nil
	}

	// Append the new entries at the end of the file
	file, err := os.OpenFile(hostPath, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	// Write the section with a blank line before
	if _, err := file.WriteString("\n\n" + message + "\n\n"); err != nil {
		return err
	}

	// Write unique entries with comments between them
	for i, entry := range uniqueEntries {
		if _, err := file.WriteString(entry.FullLine + "\n"); err != nil {
			return err
		}
		// Add a blank line after each entry except the last one
		if i < len(uniqueEntries)-1 {
			if _, err := file.WriteString("\n"); err != nil {
				return err
			}
		}
	}

	// Write closing marker with a blank line before
	if _, err := file.WriteString("\n" + message + "\n"); err != nil {
		return err
	}

	return nil
}

func AddNewDnsCoreEntries(entries []DnsEntryHost) error {
	return AppendDNSEntries(entries, "#-----NUEVOS DNS SIMF CORE CC27-----")
}

func GetDnsEntryByHostname(hostname string) (DnsEntryHost, error) {
	hostPath := config.GetHostPath()
	patterns := []string{"argus", "simf", "lbtr", "grafana"}

	entries, err := LoadHostsFile(hostPath, patterns)
	if err != nil {
		return DnsEntryHost{}, err
	}

	for _, entry := range entries {

		if entry.Hostname == hostname {

			return entry, nil

		}
	}

	return DnsEntryHost{}, fmt.Errorf("dns entry not found")
}

func (d *DnsEntryHost) ChangeEnvironment(newEnvironment string) {
	// Skip if it's the same environment
	if d.Environment == newEnvironment {
		return
	}

	// If this entry is part of a group and not the first one, just update hostname
	if len(d.GroupedWith) > 0 && d.Hostname != d.GroupedWith[0] {
		d.Hostname = changeHostnameEnvironment(d.Hostname, d.Environment, newEnvironment)
		d.Environment = newEnvironment
		return
	}

	// For single entries or the first entry of a group
	if len(d.GroupedWith) > 0 {
		// Update all hostnames in the group
		modifiedHostnames := make([]string, len(d.GroupedWith))
		for i, hostname := range d.GroupedWith {
			modifiedHostnames[i] = changeHostnameEnvironment(hostname, d.Environment, newEnvironment)
		}

		// Update FullLine with all modified hostnames
		d.FullLine = fmt.Sprintf("%s%s\t%s\t%s",
			d.LeadingComment,
			d.IP,
			strings.Join(modifiedHostnames, " "),
			d.TrailingComment)
	} else {
		// Single hostname entry
		d.Hostname = changeHostnameEnvironment(d.Hostname, d.Environment, newEnvironment)
		d.FullLine = fmt.Sprintf("%s%s\t%s\t%s",
			d.LeadingComment,
			d.IP,
			d.Hostname,
			d.TrailingComment)
	}

	d.Environment = newEnvironment
}

// Helper method to update just the hostname
func (d *DnsEntryHost) updateHostnameEnvironment(newEnvironment string) {
	d.Hostname = changeHostnameEnvironment(d.Hostname, d.Environment, newEnvironment)
}

// Helper function to change environment in a hostname
func changeHostnameEnvironment(hostname, oldEnvironment, newEnvironment string) string {
	parts := strings.Split(hostname, ".")
	lastPart := parts[len(parts)-1]

	// Remove current environment if exists
	var newParts []string
	for _, part := range parts {
		if part != oldEnvironment {
			newParts = append(newParts, part)
		}
	}

	// For production, we just use the hostname without environment
	if newEnvironment == "produccion" {
		return strings.Join(newParts, ".")
	}

	// Insert new environment before the last part
	finalParts := append(newParts[:len(newParts)-1], newEnvironment, lastPart)
	return strings.Join(finalParts, ".")
}

func ChangeEnvironmentForDNSList(entries []DnsEntryHost, newEnvironment string) []DnsEntryHost {
	modifiedEntries := make([]DnsEntryHost, len(entries))
	processedGroups := make(map[int]bool)

	for i, entry := range entries {
		modifiedEntry := entry // Create a copy

		// Skip if we've already processed this group
		if len(entry.GroupedWith) > 0 && processedGroups[entry.OriginalLine] {
			modifiedEntries[i] = modifiedEntry
			continue
		}

		modifiedEntry.ChangeEnvironment(newEnvironment)
		modifiedEntries[i] = modifiedEntry

		// Mark this group as processed
		if len(entry.GroupedWith) > 0 {
			processedGroups[entry.OriginalLine] = true
		}
	}

	return modifiedEntries
}
