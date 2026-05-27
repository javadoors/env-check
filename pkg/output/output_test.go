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

package output

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"env-check/pkg/config"
)

// createTestQueryResult 创建测试查询结果
func createTestQueryResult() *config.QueryResult {
	return &config.QueryResult{
		Timestamp: time.Now(),
		Files: []config.FileInfo{
			{
				Path:        "/tmp/test.txt",
				Exists:      true,
				IsDir:       false,
				Owner:       "user1",
				Group:       "group1",
				Permissions: "-rw-r--r--",
			},
			{
				Path:        "/tmp/dir",
				Exists:      true,
				IsDir:       true,
				Owner:       "user2",
				Group:       "group2",
				Permissions: "drwxr-xr-x",
			},
			{
				Path:   "/tmp/missing",
				Exists: false,
				Error:  "路径不存在",
			},
		},
		Summary: config.QuerySummary{
			TotalChecked: 3,
			TotalExists:  2,
			TotalMissing: 1,
			TotalDirs:    1,
			TotalFiles:   1,
		},
	}
}

// createTestCleanResult 创建测试清理结果
func createTestCleanResult() *config.CleanResult {
	return &config.CleanResult{
		Timestamp: time.Now(),
		Deleted: []config.FileInfo{
			{
				Path:   "/tmp/deleted1.txt",
				Exists: false,
				IsDir:  false,
			},
			{
				Path:   "/tmp/deleted2",
				Exists: false,
				IsDir:  true,
			},
		},
		Failed: []config.FileInfo{
			{
				Path:  "/tmp/failed.txt",
				Error: "权限不足",
			},
		},
		Skipped: []config.FileInfo{
			{
				Path:  "/tmp/skipped.txt",
				Error: "用户取消",
			},
		},
		Summary: config.CleanSummary{
			TotalChecked: 4,
			TotalDeleted: 2,
			TotalFailed:  1,
			TotalSkipped: 1,
		},
	}
}

// createTestProgramCheckResult 创建测试程序检查结果
func createTestProgramCheckResult() *config.ProgramCheckResult {
	return &config.ProgramCheckResult{
		Timestamp: time.Now(),
		Programs: []config.ProgramInfo{
			{
				Name:      "docker",
				Installed: true,
				Version:   "24.0.6",
				Path:      "/usr/bin/docker",
			},
			{
				Name:      "git",
				Installed: true,
				Version:   "2.39.2",
				Path:      "/usr/bin/git",
			},
			{
				Name:      "nonexistent",
				Installed: false,
				Error:     "程序未找到",
			},
		},
		Summary: config.ProgramCheckSummary{
			TotalChecked:   3,
			TotalInstalled: 2,
			TotalMissing:   1,
		},
	}
}

// createTestClockCheckResult 创建测试时钟检查结果
func createTestClockCheckResult() *config.ClockCheckResult {
	result := &config.ClockCheckResult{
		Timestamp: time.Now(),
		Result:    "fail",
		Clocks: []config.ClockCheckInfo{
			{
				Host:     "*",
				Role:     []string{"bootstrap"},
				IsSynced: true,
				TimeDiff: 0,
				IsLocal:  true,
				Error:    "",
			},
			{
				Host:     "**",
				Role:     []string{"worker"},
				IsSynced: false,
				TimeDiff: -1,
				IsLocal:  false,
				Error:    "failed to connect",
			},
			{
				Host:     "***",
				Role:     []string{"master"},
				IsSynced: true,
				TimeDiff: 2,
				IsLocal:  false,
				Error:    "",
			},
		},
	}
	return result
}

func TestNewFormatter(t *testing.T) {
	formatter := NewFormatter()

	if formatter == nil {
		t.Error("Expected Formatter instance, got nil")
	}
}

func TestFormatQueryText(t *testing.T) {
	formatter := NewFormatter()
	result := createTestQueryResult()

	output, err := formatter.FormatQuery(result, "text")

	assert.NoError(t, err)
	assert.Contains(t, output, "Query Time")
	assert.Contains(t, output, "/tmp/test.txt")
	assert.Contains(t, output, "/tmp/dir")
	assert.Contains(t, output, "/tmp/missing")
	assert.Contains(t, output, "Summary")
}

func TestFormatQueryJSON(t *testing.T) {
	formatter := NewFormatter()
	result := createTestQueryResult()

	output, err := formatter.FormatQuery(result, "json")

	assert.NoError(t, err)
	assert.Contains(t, output, "timestamp")
	assert.Contains(t, output, "files")
	assert.Contains(t, output, "summary")
	// 验证为有效 JSON
	var js interface{}
	assert.NoError(t, json.Unmarshal([]byte(output), &js))

	// 验证具体内容
	assert.Contains(t, output, "/tmp/test.txt")
	assert.Contains(t, output, "user1")
	assert.Contains(t, output, "group1")
}

func TestFormatCleanText(t *testing.T) {
	formatter := NewFormatter()
	result := createTestCleanResult()

	output, err := formatter.FormatClean(result, "text")

	assert.NoError(t, err)
	assert.Contains(t, output, "Clean Time")
	assert.Contains(t, output, "Deleted Files")
	assert.Contains(t, output, "Failed Files")
	assert.Contains(t, output, "Skipped Files")
	assert.Contains(t, output, "Summary")
}

func TestFormatCleanJSON(t *testing.T) {
	formatter := NewFormatter()
	result := createTestCleanResult()

	output, err := formatter.FormatClean(result, "json")

	assert.NoError(t, err)
	assert.Contains(t, output, "timestamp")
	assert.Contains(t, output, "deleted")
	assert.Contains(t, output, "failed")
	assert.Contains(t, output, "skipped")
	assert.Contains(t, output, "summary")
	var js2 interface{}
	assert.NoError(t, json.Unmarshal([]byte(output), &js2))
}

func TestFormatProgramCheckText(t *testing.T) {
	formatter := NewFormatter()
	result := createTestProgramCheckResult()

	output, err := formatter.FormatProgramCheck(result, "text")

	assert.NoError(t, err)
	assert.Contains(t, output, "Check Time")
	assert.Contains(t, output, "docker")
	assert.Contains(t, output, "git")
	assert.Contains(t, output, "nonexistent")
	assert.Contains(t, output, "Summary")
}

func TestFormatProgramCheckJSON(t *testing.T) {
	formatter := NewFormatter()
	result := createTestProgramCheckResult()

	output, err := formatter.FormatProgramCheck(result, "json")

	assert.NoError(t, err)
	assert.Contains(t, output, "timestamp")
	assert.Contains(t, output, "programs")
	assert.Contains(t, output, "summary")
	assert.Contains(t, output, "docker")
	assert.Contains(t, output, "24.0.6")
	var js3 interface{}
	assert.NoError(t, json.Unmarshal([]byte(output), &js3))
}

func TestFormatClockCheckText(t *testing.T) {
	formatter := NewFormatter()
	result := createTestClockCheckResult()

	out, err := formatter.FormatClockCheck(result, "text")
	assert.NoError(t, err)
	assert.Contains(t, out, "Clock Check Time")
	assert.Contains(t, out, "Clock Check Result")
	assert.Contains(t, out, "ERROR") // negative TimeDiff rendered as ERROR
	assert.Contains(t, out, "failed to connect")
}

func TestFormatClockCheckJSON(t *testing.T) {
	formatter := NewFormatter()
	result := createTestClockCheckResult()

	out, err := formatter.FormatClockCheck(result, "json")
	assert.NoError(t, err)
	assert.Contains(t, out, "timestamp")
	assert.Contains(t, out, "clocks")
	assert.Contains(t, out, "Result")
	var js4 interface{}
	assert.NoError(t, json.Unmarshal([]byte(out), &js4))
}
