/******************************************************************
 * Copyright (c) 2025 Huawei Technologies Co., Ltd.
 * openFuyao is licensed under Mulan PSL v2.
 * You can use this software according to the terms and conditions of the Mulan PSL v2.
 * You may obtain a copy of Mulan PSL v2 at:
 *          http://license.coscl.org.cn/MulanPSL2
 * THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
 * EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
 * MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
 * See the Mulan PSL v2 for more details.
 ******************************************************************/

// Package output handles formatted output
package output

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/liushuochen/gotable"

	"env-check/pkg/config"
)

const dispatchErrorPreviewMaxLen = 30

// Formatter output formatter
type Formatter struct{}

// NewFormatter creates a new formatter
func NewFormatter() *Formatter {
	return &Formatter{}
}

// FormatQuery formats query result
func (f *Formatter) FormatQuery(result *config.QueryResult, format string) (string, error) {
	switch format {
	case "json":
		return f.formatQueryJSON(result)
	default:
		return f.formatQueryText(result)
	}
}

// FormatClean formats clean result
func (f *Formatter) FormatClean(result *config.CleanResult, format string) (string, error) {
	switch format {
	case "json":
		return f.formatCleanJSON(result)
	default:
		return f.formatCleanText(result)
	}
}

// FormatProgramCheck formats program check result
func (f *Formatter) FormatProgramCheck(result *config.ProgramCheckResult, format string) (string, error) {
	switch format {
	case "json":
		return f.formatProgramCheckJSON(result)
	default:
		return f.formatProgramCheckText(result)
	}
}

// FormatClockCheck formats clock check result
func (f *Formatter) FormatClockCheck(result *config.ClockCheckResult, format string) (string, error) {
	switch format {
	case "json":
		return f.formatClockCheckJSON(result)
	default:
		return f.formatClockCheckText(result)
	}
}

// FormatDispatch formats dispatch result
func (f *Formatter) FormatDispatch(result *config.DispatchResult, format string) (string, error) {
	switch format {
	case "json":
		return f.formatDispatchJSON(result)
	default:
		return f.formatDispatchText(result)
	}
}

func (f *Formatter) formatDispatchJSON(result *config.DispatchResult) (string, error) {
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (f *Formatter) formatDispatchText(result *config.DispatchResult) (string, error) {
	var builder strings.Builder

	builder.WriteString("\n")
	builder.WriteString(fmt.Sprintf("Dispatch Check Time: %s\n", result.Timestamp.Format("2006-01-02 15:04:05")))
	builder.WriteString(fmt.Sprintf("Total Duration: %s\n", result.Duration))
	builder.WriteString(fmt.Sprintf("Overall Result: %s\n\n", result.Summary.Result))

	// Nodes table
	table, err := gotable.Create("IP", "Role", "Status", "Duration", "Error")
	if err != nil {
		return "", fmt.Errorf("failed to create table: %v", err)
	}

	for _, node := range result.Nodes {
		role := strings.Join(node.Role, ",")
		duration := node.EndTime.Sub(node.StartTime).String()
		errStr := "-"
		if node.Error != "" {
			errStr = truncateString(node.Error, dispatchErrorPreviewMaxLen)
		}
		table.AddRow([]string{node.IP, role, node.Status, duration, errStr})
	}

	builder.WriteString(table.String())
	builder.WriteString("\n")

	// Summary
	summaryTable, _ := gotable.Create("Summary", "Count")
	summaryTable.AddRow([]string{"Total Nodes", fmt.Sprintf("%d", result.Summary.TotalNodes)})
	summaryTable.AddRow([]string{"Success", fmt.Sprintf("%d", result.Summary.SuccessNodes)})
	summaryTable.AddRow([]string{"Failed", fmt.Sprintf("%d", result.Summary.FailedNodes)})
	summaryTable.AddRow([]string{"Running", fmt.Sprintf("%d", result.Summary.RunningNodes)})
	builder.WriteString(summaryTable.String())
	builder.WriteString("\n")

	return builder.String(), nil
}

// truncateString truncates string to max length
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// Query result text format using gotable
func (f *Formatter) formatQueryText(result *config.QueryResult) (string, error) {
	var builder strings.Builder

	builder.WriteString("\n")
	builder.WriteString(fmt.Sprintf("Query Time: %s\n\n", result.Timestamp.Format("2006-01-02 15:04:05")))

	tableStr, err := f.buildQueryFilesTable(result.Files)
	if err != nil {
		return "", err
	}
	builder.WriteString(tableStr)
	builder.WriteString("\n")

	summaryStr, err := f.buildQuerySummary(result)
	if err != nil {
		return "", err
	}
	builder.WriteString(summaryStr)
	builder.WriteString("\n")

	return builder.String(), nil
}

// buildQueryFilesTable creates the files table for query output
func (f *Formatter) buildQueryFilesTable(files []config.FileInfo) (string, error) {
	table, err := gotable.Create("Path", "Exists", "Type", "Owner", "Group", "Permissions")
	if err != nil {
		return "", fmt.Errorf("failed to create table: %v", err)
	}

	for _, file := range files {
		status := "Missing"
		if file.Exists {
			status = "Exists"
		}

		fileType := f.getTypeString(file)
		owner := file.Owner
		if owner == "" {
			owner = "-"
		}
		group := file.Group
		if group == "" {
			group = "-"
		}
		perms := file.Permissions
		if perms == "" {
			perms = "-"
		}

		if err := table.AddRow([]string{file.Path, status, fileType, owner, group, perms}); err != nil {
			return "", fmt.Errorf("failed to add row: %v", err)
		}
	}

	return table.String(), nil
}

// buildQuerySummary creates the summary table for query output
func (f *Formatter) buildQuerySummary(result *config.QueryResult) (string, error) {
	summaryTable, err := gotable.Create("Summary", "Count")
	if err != nil {
		return "", fmt.Errorf("failed to create summary table: %v", err)
	}

	summaryTable.AddRow([]string{"Total Checked", fmt.Sprintf("%d", result.Summary.TotalChecked)})
	summaryTable.AddRow([]string{"Total Exists", fmt.Sprintf("%d", result.Summary.TotalExists)})
	summaryTable.AddRow([]string{"Total Missing", fmt.Sprintf("%d", result.Summary.TotalMissing)})
	summaryTable.AddRow([]string{"Total Directories", fmt.Sprintf("%d", result.Summary.TotalDirs)})
	summaryTable.AddRow([]string{"Total Files", fmt.Sprintf("%d", result.Summary.TotalFiles)})

	return summaryTable.String(), nil
}

// Clean result text format using gotable
func (f *Formatter) formatCleanText(result *config.CleanResult) (string, error) {
	var builder strings.Builder

	builder.WriteString("\n")
	builder.WriteString(fmt.Sprintf("Clean Time: %s\n\n", result.Timestamp.Format("2006-01-02 15:04:05")))

	if s, err := f.buildCleanDeletedSection(result.Deleted); err != nil {
		return "", err
	} else if s != "" {
		builder.WriteString(s)
		builder.WriteString("\n")
	}

	if s, err := f.buildCleanFailedSection(result.Failed); err != nil {
		return "", err
	} else if s != "" {
		builder.WriteString(s)
		builder.WriteString("\n")
	}

	if s, err := f.buildCleanSkippedSection(result.Skipped); err != nil {
		return "", err
	} else if s != "" {
		builder.WriteString(s)
		builder.WriteString("\n")
	}

	summaryStr, err := f.buildCleanSummary(result)
	if err != nil {
		return "", err
	}
	builder.WriteString(summaryStr)
	builder.WriteString("\n")

	return builder.String(), nil
}

func (f *Formatter) buildCleanDeletedSection(deleted []config.FileInfo) (string, error) {
	if len(deleted) == 0 {
		return "", nil
	}
	var b strings.Builder
	b.WriteString("Deleted Files:\n")
	table, err := gotable.Create("Path", "Type")
	if err != nil {
		return "", fmt.Errorf("failed to create table: %v", err)
	}
	for _, file := range deleted {
		typeStr := "File"
		if file.IsDir {
			typeStr = "Directory"
		}
		table.AddRow([]string{file.Path, typeStr})
	}
	b.WriteString(table.String())
	return b.String(), nil
}

func (f *Formatter) buildCleanFailedSection(failed []config.FileInfo) (string, error) {
	if len(failed) == 0 {
		return "", nil
	}
	var b strings.Builder
	b.WriteString("Failed Files:\n")
	table, err := gotable.Create("Path", "Error")
	if err != nil {
		return "", fmt.Errorf("failed to create table: %v", err)
	}
	for _, file := range failed {
		errorMsg := file.Error
		if errorMsg == "" {
			errorMsg = "Unknown error"
		}
		table.AddRow([]string{file.Path, errorMsg})
	}
	b.WriteString(table.String())
	return b.String(), nil
}

func (f *Formatter) buildCleanSkippedSection(skipped []config.FileInfo) (string, error) {
	if len(skipped) == 0 {
		return "", nil
	}
	var b strings.Builder
	b.WriteString("Skipped Files:\n")
	table, err := gotable.Create("Path", "Reason")
	if err != nil {
		return "", fmt.Errorf("failed to create table: %v", err)
	}
	for _, file := range skipped {
		reason := "User canceled"
		if file.Error != "" {
			reason = file.Error
		}
		table.AddRow([]string{file.Path, reason})
	}
	b.WriteString(table.String())
	return b.String(), nil
}

func (f *Formatter) buildCleanSummary(result *config.CleanResult) (string, error) {
	summaryTable, err := gotable.Create("Summary", "Count")
	if err != nil {
		return "", fmt.Errorf("failed to create summary table: %v", err)
	}
	summaryTable.AddRow([]string{"Total Checked", fmt.Sprintf("%d", result.Summary.TotalChecked)})
	summaryTable.AddRow([]string{"Total Deleted", fmt.Sprintf("%d", result.Summary.TotalDeleted)})
	summaryTable.AddRow([]string{"Total Failed", fmt.Sprintf("%d", result.Summary.TotalFailed)})
	summaryTable.AddRow([]string{"Total Skipped", fmt.Sprintf("%d", result.Summary.TotalSkipped)})
	return summaryTable.String(), nil
}

// Program check result text format using gotable
func (f *Formatter) formatProgramCheckText(result *config.ProgramCheckResult) (string, error) {
	var builder strings.Builder

	builder.WriteString("\n")
	builder.WriteString(fmt.Sprintf("Check Time: %s\n\n", result.Timestamp.Format("2006-01-02 15:04:05")))

	// Create table for program information
	table, err := gotable.Create("Program", "Status", "Version", "Path")
	if err != nil {
		return "", fmt.Errorf("failed to create table: %v", err)
	}

	for _, program := range result.Programs {
		status := "Not Installed"
		if program.Installed {
			status = "Installed"
		}

		version := program.Version
		if version == "" {
			version = "Unknown"
		}

		path := program.Path
		if path == "" {
			path = "Not found"
		}

		table.AddRow([]string{
			program.Name,
			status,
			version,
			path,
		})
	}

	builder.WriteString(table.String())
	builder.WriteString("\n")

	// Add summary information
	summaryTable, err := gotable.Create("Summary", "Count")
	if err != nil {
		return "", fmt.Errorf("failed to create summary table: %v", err)
	}

	summaryTable.AddRow([]string{"Total Checked", fmt.Sprintf("%d", result.Summary.TotalChecked)})
	summaryTable.AddRow([]string{"Total Installed", fmt.Sprintf("%d", result.Summary.TotalInstalled)})
	summaryTable.AddRow([]string{"Total Missing", fmt.Sprintf("%d", result.Summary.TotalMissing)})

	builder.WriteString(summaryTable.String())
	builder.WriteString("\n")

	return builder.String(), nil
}

// JSON format functions
func (f *Formatter) formatQueryJSON(result *config.QueryResult) (string, error) {
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (f *Formatter) formatCleanJSON(result *config.CleanResult) (string, error) {
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (f *Formatter) formatProgramCheckJSON(result *config.ProgramCheckResult) (string, error) {
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// Utility functions
func (f *Formatter) getTypeString(file config.FileInfo) string {
	if !file.Exists {
		return "Missing"
	}
	if file.IsDir {
		return "Directory"
	}
	return "File"
}

func (f *Formatter) formatClockCheckJSON(result *config.ClockCheckResult) (string, error) {
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (f *Formatter) formatClockCheckText(result *config.ClockCheckResult) (string, error) {
	var builder strings.Builder

	builder.WriteString("\n")
	builder.WriteString(fmt.Sprintf("Clock Check Time: %s\n", result.Timestamp.Format("2006-01-02 15:04:05")))
	builder.WriteString(fmt.Sprintf("Clock Check Result: %s\n\n", result.Result))
	table, err := gotable.Create("Host", "Role", "Synced", "Time Diff (s)", "Local", "Error")
	if err != nil {
		fmt.Printf("Error creating table: %v\n", err)
		return "", err
	}

	for _, r := range result.Clocks {
		var syncedStr, diffStr, localStr, errStr string
		syncedStr = fmt.Sprintf("%v", r.IsSynced)
		if r.TimeDiff >= 0 {
			diffStr = fmt.Sprintf("%d", r.TimeDiff)
		} else {
			diffStr = "ERROR"
		}
		localStr = fmt.Sprintf("%v", r.IsLocal)
		errStr = r.Error
		err1 := table.AddRow([]string{r.Host, strings.Join(r.Role, ","), syncedStr, diffStr, localStr, errStr})
		if err1 != nil {
			return "", fmt.Errorf("failed to add row: %v", err1)
		}
	}
	builder.WriteString(table.String())
	builder.WriteString("\n")
	return builder.String(), nil
}

// ========== Kernel Check Formatting ==========

// FormatKernelCheck formats kernel check result
func (f *Formatter) FormatKernelCheck(result *config.KernelCheckResult, format string) (string, error) {
	switch format {
	case "json":
		return f.formatKernelCheckJSON(result)
	default:
		return f.formatKernelCheckText(result)
	}
}

func (f *Formatter) formatKernelCheckJSON(result *config.KernelCheckResult) (string, error) {
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (f *Formatter) formatKernelCheckText(result *config.KernelCheckResult) (string, error) {
	var builder strings.Builder

	builder.WriteString("\n")
	builder.WriteString(fmt.Sprintf("Kernel Check Time: %s\n", result.Timestamp.Format("2006-01-02 15:04:05")))
	builder.WriteString("\n")

	table, err := gotable.Create("Item", "Value")
	if err != nil {
		return "", fmt.Errorf("failed to create table: %v", err)
	}

	info := result.KernelInfo
	table.AddRow([]string{"Current Version", info.Version})
	table.AddRow([]string{"OS", info.OS})
	table.AddRow([]string{"Architecture", info.Arch})
	table.AddRow([]string{"Expected Version", fmt.Sprintf("%s %s", info.Operator, info.ExpectedVer)})

	status := "PASS"
	if !info.IsValid {
		status = "FAIL"
	}
	table.AddRow([]string{"Status", status})

	if info.Error != "" {
		table.AddRow([]string{"Error", info.Error})
	}

	builder.WriteString(table.String())
	builder.WriteString("\n")

	// Summary
	summaryTable, _ := gotable.Create("Summary", "Count")
	summaryTable.AddRow([]string{"Total Checked", fmt.Sprintf("%d", result.Summary.TotalChecked)})
	summaryTable.AddRow([]string{"Passed", fmt.Sprintf("%d", result.Summary.Passed)})
	summaryTable.AddRow([]string{"Failed", fmt.Sprintf("%d", result.Summary.Failed)})
	builder.WriteString(summaryTable.String())
	builder.WriteString("\n")

	return builder.String(), nil
}

// ========== Port Check Formatting ==========

// FormatPortCheck formats port check result
func (f *Formatter) FormatPortCheck(result *config.PortCheckResult, format string) (string, error) {
	switch format {
	case "json":
		return f.formatPortCheckJSON(result)
	default:
		return f.formatPortCheckText(result)
	}
}

func (f *Formatter) formatPortCheckJSON(result *config.PortCheckResult) (string, error) {
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (f *Formatter) formatPortCheckText(result *config.PortCheckResult) (string, error) {
	var builder strings.Builder

	builder.WriteString("\n")
	builder.WriteString(fmt.Sprintf("Port Check Time: %s\n", result.Timestamp.Format("2006-01-02 15:04:05")))
	builder.WriteString("\n")

	table, err := gotable.Create("Port", "Protocol", "Status", "Process")
	if err != nil {
		return "", fmt.Errorf("failed to create table: %v", err)
	}

	for _, port := range result.Ports {
		status := "Free"
		if port.IsUsed {
			status = "Occupied"
		}
		process := port.Process
		if process == "" {
			process = "-"
		}
		table.AddRow([]string{port.Port, port.Protocol, status, process})
	}

	builder.WriteString(table.String())
	builder.WriteString("\n")

	// Summary
	summaryTable, _ := gotable.Create("Summary", "Count")
	summaryTable.AddRow([]string{"Total Checked", fmt.Sprintf("%d", result.Summary.TotalChecked)})
	summaryTable.AddRow([]string{"Occupied", fmt.Sprintf("%d", result.Summary.Used)})
	summaryTable.AddRow([]string{"Free", fmt.Sprintf("%d", result.Summary.Free)})
	builder.WriteString(summaryTable.String())
	builder.WriteString("\n")

	return builder.String(), nil
}

// ========== Disk Check Formatting ==========

// FormatDiskCheck formats disk check result
func (f *Formatter) FormatDiskCheck(result *config.DiskCheckResult, format string) (string, error) {
	switch format {
	case "json":
		return f.formatDiskCheckJSON(result)
	default:
		return f.formatDiskCheckText(result)
	}
}

func (f *Formatter) formatDiskCheckJSON(result *config.DiskCheckResult) (string, error) {
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (f *Formatter) formatDiskCheckText(result *config.DiskCheckResult) (string, error) {
	var builder strings.Builder

	builder.WriteString("\n")
	builder.WriteString(fmt.Sprintf("Disk Check Time: %s\n", result.Timestamp.Format("2006-01-02 15:04:05")))
	builder.WriteString("\n")

	table, err := gotable.Create("Path", "Filesystem", "Total(GB)", "Free(GB)", "Used(GB)", "Used%", "Status")
	if err != nil {
		return "", fmt.Errorf("failed to create table: %v", err)
	}

	for _, space := range result.Spaces {
		status := "PASS"
		if !space.IsSufficient {
			status = "FAIL"
		}
		if space.Error != "" {
			status = "ERROR"
		}

		totalGB := float64(space.Total) / (1024 * 1024 * 1024)
		freeGB := float64(space.Free) / (1024 * 1024 * 1024)
		usedGB := float64(space.Used) / (1024 * 1024 * 1024)

		fs := space.Filesystem
		if fs == "" {
			fs = "-"
		}

		table.AddRow([]string{
			space.Path,
			fs,
			fmt.Sprintf("%.1f", totalGB),
			fmt.Sprintf("%.1f", freeGB),
			fmt.Sprintf("%.1f", usedGB),
			fmt.Sprintf("%.1f%%", space.UsedPercent),
			status,
		})
	}

	builder.WriteString(table.String())
	builder.WriteString("\n")

	// Summary
	summaryTable, _ := gotable.Create("Summary", "Count")
	summaryTable.AddRow([]string{"Total Checked", fmt.Sprintf("%d", result.Summary.TotalChecked)})
	summaryTable.AddRow([]string{"Sufficient", fmt.Sprintf("%d", result.Summary.SufficientPath)})
	summaryTable.AddRow([]string{"Insufficient", fmt.Sprintf("%d", result.Summary.InsufficientPath)})
	builder.WriteString(summaryTable.String())
	builder.WriteString("\n")

	return builder.String(), nil
}
