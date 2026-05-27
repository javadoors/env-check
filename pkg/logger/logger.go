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

// Package logger handles logging during execution
package logger

import (
	"fmt"
	"os"
	"sync"
	"time"

	"env-check/pkg/config"
)

// LogLevel log level type
type LogLevel int

const (
	// LevelInfo information
	LevelInfo LogLevel = iota
	// LevelSuccess success
	LevelSuccess
	// LevelWarning warning
	LevelWarning
	// LevelError error
	LevelError
)

// Logger logger
type Logger struct {
	logFile string
	mu      sync.Mutex
}

// NewLogger creates a new logger
func NewLogger(logFile string) *Logger {
	return &Logger{
		logFile: logFile,
	}
}

// Log logs a message
func (l *Logger) Log(level LogLevel, message string) {
	timestamp := time.Now().Format("2006-01-02 15:04:05")

	var levelStr string
	switch level {
	case LevelInfo:
		levelStr = "INFO"
	case LevelSuccess:
		levelStr = "SUCCESS"
	case LevelWarning:
		levelStr = "WARNING"
	case LevelError:
		levelStr = "ERROR"
	default:
		levelStr = "INFO"
	}

	logEntry := fmt.Sprintf("[%s][%s] %s", levelStr, timestamp, message)

	// Output to console (with color)
	l.printColored(level, logEntry)

	// Write to log file
	l.writeToFile(logEntry)
}

// Print colored output to console
func (l *Logger) printColored(level LogLevel, message string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	switch level {
	case LevelInfo:
		fmt.Printf("\033[34m%s\033[0m\n", message) // Blue
	case LevelSuccess:
		fmt.Printf("\033[32m%s\033[0m\n", message) // Green
	case LevelWarning:
		fmt.Printf("\033[33m%s\033[0m\n", message) // Yellow
	case LevelError:
		fmt.Printf("\033[31m%s\033[0m\n", message) // Red
	default:
		fmt.Printf("\033[34m%s\033[0m\n", message) // Blue
	}
}

// Write to log file
func (l *Logger) writeToFile(message string) {
	if l.logFile == "" {
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	file, err := os.OpenFile(l.logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, config.FileMode)
	if err != nil {
		fmt.Printf("open log file failed: %v\n", err)
		return
	}
	defer file.Close()

	_, err = file.WriteString(message + "\n")
	if err != nil {
		fmt.Printf("write to log file failed: %v\n", err)
	}
}

// Info logs information
func (l *Logger) Info(message string) {
	l.Log(LevelInfo, message)
}

// Success logs successful operation
func (l *Logger) Success(message string) {
	l.Log(LevelSuccess, message)
}

// Warning logs warning information
func (l *Logger) Warning(message string) {
	l.Log(LevelWarning, message)
}

// Error logs error
func (l *Logger) Error(message string) {
	l.Log(LevelError, message)
}
