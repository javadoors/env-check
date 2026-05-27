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
	"strings"

	"github.com/spf13/cobra"

	"env-check/pkg/config"
	"env-check/pkg/logger"
	"env-check/pkg/output"
	"env-check/pkg/program"
)

// runCheck executes program check operation using the common runner
func runCheck(cmd *cobra.Command, args []string) error {
	return runAction(cmd, args, config.ModeProgramCheck, actionHandler{
		execute: func(cmd *cobra.Command, cfg *config.AppConfig, log *logger.Logger) (interface{}, error) {
			p := program.NewProgramChecker(cfg, log)
			result, err := p.Execute()
			if err != nil {
				log.Error("program check failed: " + err.Error())
				return nil, err
			}

			if result.Summary.TotalInstalled > 0 {
				var installed []string
				for _, prog := range result.Programs {
					if prog.Installed {
						installed = append(installed, prog.Name)
					}
				}
				log.Warning("detected installed application(s): " + strings.Join(installed, ", ") +
					". Please uninstall it(them) yourself")
			}

			return result, nil
		},
		generateHTML: func(res interface{}) error {
			return output.GenerateCheckHTML(*(res.(*config.ProgramCheckResult)), "check.html")
		},
		format: func(res interface{}, outputFormat string) (string, error) {
			formatter := output.NewFormatter()
			return formatter.FormatProgramCheck(res.(*config.ProgramCheckResult), outputFormat)
		},
		baseName: "program-check-result",
	})
}
