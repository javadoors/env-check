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

package query

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	"env-check/pkg/config"
	"env-check/pkg/logger"
)

func TestQueryExecuteBasic(t *testing.T) {
	tempDir := setupTestDir(t)
	defer os.RemoveAll(tempDir)

	cfg := createTestConfig(tempDir)
	log := logger.NewLogger("")
	q := NewQuery(cfg, log)

	result, err := q.Execute()
	assert.NoError(t, err)
	assert.Equal(t, 1, len(result.Files))
	assert.Equal(t, true, result.Files[0].Exists)
}

func TestQueryExecuteNonExistentPath(t *testing.T) {
	cfg := &config.AppConfig{
		Mode:  config.ModeQuery,
		Paths: []string{"/non/existent/path"},
	}
	log := logger.NewLogger("")
	q := NewQuery(cfg, log)

	result, err := q.Execute()
	assert.NoError(t, err)
	assert.Equal(t, 1, len(result.Files))
	assert.Equal(t, false, result.Files[0].Exists)
}

func TestQueryPathFile(t *testing.T) {
	tempDir := setupTestDir(t)
	defer os.RemoveAll(tempDir)

	cfg := createTestConfig(tempDir)
	log := logger.NewLogger("")
	q := NewQuery(cfg, log)

	fileInfo := q.queryPath(tempDir)
	assert.Equal(t, true, fileInfo.Exists)
	assert.Equal(t, true, fileInfo.IsDir)
}

func TestQueryPathNonExistent(t *testing.T) {
	cfg := createTestConfig("")
	log := logger.NewLogger("")
	q := NewQuery(cfg, log)

	fileInfo := q.queryPath("/non/existent/path")
	assert.Equal(t, false, fileInfo.Exists)
	assert.NotEqual(t, "", fileInfo.Error)
}

// 辅助函数
func setupTestDir(t *testing.T) string {
	tempDir, err := os.MkdirTemp("", "query-test-*")
	assert.NoError(t, err)

	testFile := filepath.Join(tempDir, "test.txt")
	err = os.WriteFile(testFile, []byte("test content"), config.FileMode)
	assert.NoError(t, err)

	return tempDir
}

func createTestConfig(path string) *config.AppConfig {
	return &config.AppConfig{
		Mode:  config.ModeQuery,
		Paths: []string{path},
	}
}
