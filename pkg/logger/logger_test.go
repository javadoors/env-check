/*
 * Copyright (c) 2026 Huawei Technologies Co., Ltd.
 * openFuyao is licensed under Mulan PSL v2.
 * You can use this software according to the terms and conditions of the Mulan PSL v2.
 * You may obtain a copy of Mulan PSL v2 at:
 *          http://license.coscl.org.cn/MulanPSL2
 * THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
 * EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
 * MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
 * See the Mulan PSL v2 for more details.
 */

package logger

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewLoggerNotNil(t *testing.T) {
	l := NewLogger("")
	assert.NotNil(t, l)
}

func TestInfoWritesToFile(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "info.log")

	l := NewLogger(file)
	l.Info("hello-info")

	data, err := os.ReadFile(file)
	assert.NoError(t, err)
	s := string(data)
	assert.True(t, strings.Contains(s, "[INFO]"))
	assert.True(t, strings.Contains(s, "hello-info"))
}

func TestOtherLevelsWrittenWithLevel(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "levels.log")

	l := NewLogger(file)
	l.Success("ok")
	l.Warning("warn")
	l.Error("err")

	data, err := os.ReadFile(file)
	assert.NoError(t, err)
	s := string(data)

	assert.True(t, strings.Contains(s, "[SUCCESS]"))
	assert.True(t, strings.Contains(s, "ok"))
	assert.True(t, strings.Contains(s, "[WARNING]"))
	assert.True(t, strings.Contains(s, "warn"))
	assert.True(t, strings.Contains(s, "[ERROR]"))
	assert.True(t, strings.Contains(s, "err"))
}
