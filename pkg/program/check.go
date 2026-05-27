/*
 * Copyright (c) 2025 Huawei Technologies Co., Ltd.
 * openFuyao is licensed under Mulan PSL v2.
 * You can use this software according to the terms and conditions of the Mulan PSL v2.
 * You may obtain a copy of Mulan PSL v2 at:
 *          http://license.coscl.org.cn/MulanPSL2
 * THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
 * EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
 * MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
 * See the Mulan PSL v2 for more details.
 */

// Package program handles application-related operations, currently only includes existence detection
package program

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"env-check/pkg/config"
	"env-check/pkg/logger"
)

// ApplicationChecker program checker
type ApplicationChecker struct {
	config *config.AppConfig
	logger *logger.Logger
}

// NewProgramChecker creates a new program checker
func NewProgramChecker(cfg *config.AppConfig, log *logger.Logger) *ApplicationChecker {
	return &ApplicationChecker{
		config: cfg,
		logger: log,
	}
}

// Execute executes program check
func (p *ApplicationChecker) Execute() (*config.ProgramCheckResult, error) {
	p.logger.Info("program check start......")
	sysInfo := p.getSystemInfo()
	p.logger.Info(fmt.Sprintf("os: %s, arch: %s", sysInfo["os"], sysInfo["arch"]))
	p.logger.Info(fmt.Sprintf("program list to check: %v", p.config.ProgramList))

	result := &config.ProgramCheckResult{
		Timestamp: time.Now(),
		Programs:  make([]config.ProgramInfo, 0),
		Summary:   config.ProgramCheckSummary{},
	}

	// Check all programs
	for _, program := range p.config.ProgramList {
		programInfo := p.checkProgram(program)
		result.Programs = append(result.Programs, programInfo)
		p.updateProgramSummary(programInfo, result)
	}

	p.logger.Info("program check completed")
	return result, nil
}

// checkProgram checks a single program
func (p *ApplicationChecker) checkProgram(program string) config.ProgramInfo {
	p.logger.Info(fmt.Sprintf("check program: %s", program))

	info := config.ProgramInfo{
		Name: program,
	}

	// Find program path
	path, err := exec.LookPath(program)
	if err != nil {
		info.Installed = false
		info.Error = fmt.Sprintf("not found: %v", err)
		p.logger.Info(fmt.Sprintf("not install: %s", program))
		return info
	}

	info.Installed = true
	info.Path = path

	// Try to get version information
	version, err := p.getProgramVersion(program)
	if err == nil {
		info.Version = version
		p.logger.Warning(fmt.Sprintf("%s installed - version: %s", program, version))
	} else {
		p.logger.Warning(fmt.Sprintf("%s installed - path: %s", program, path))
	}

	return info
}

// getProgramVersion gets program version
func (p *ApplicationChecker) getProgramVersion(program string) (string, error) {
	var cmd = exec.Command(program, "--version")

	// Select version command based on program name
	switch program {
	case "go":
		cmd = exec.Command("go", "version")
	case "java":
		cmd = exec.Command("java", "-version")
	case "kubectl":
		cmd = exec.Command("kubectl", "version", "--client")
	case "helm":
		cmd = exec.Command("helm", "version", "--short")
	default:
		// Default to try --version parameter
		cmd = exec.Command(program, "--version")
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(output)), nil
}

// updateProgramSummary updates program check summary information
func (p *ApplicationChecker) updateProgramSummary(programInfo config.ProgramInfo, result *config.ProgramCheckResult) {
	result.Summary.TotalChecked++

	if programInfo.Installed {
		result.Summary.TotalInstalled++
	} else {
		result.Summary.TotalMissing++
	}
}

// getSystemInfo gets system information
func (p *ApplicationChecker) getSystemInfo() map[string]string {
	info := make(map[string]string)

	info["os"] = runtime.GOOS
	info["arch"] = runtime.GOARCH

	// Get other system information
	if hostname, err := os.Hostname(); err == nil {
		info["hostname"] = hostname
	}

	return info
}
