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

package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetOutputFilename(t *testing.T) {
	assert := assert.New(t)
	name := getOutputFilename("base-name", "json")
	assert.Equal("base-name.json", name)

	name = getOutputFilename("base-name", "text")
	assert.Equal("", name)
}

func TestSaveResultToFile(t *testing.T) {
	assert := assert.New(t)
	dir := t.TempDir()
	outPath := filepath.Join(dir, "sub", "res.json")

	err := saveResultToFile(outPath, "hello")
	assert.NoError(err)

	data, err := os.ReadFile(outPath)
	assert.NoError(err)
	assert.Equal("hello", string(data))
}
