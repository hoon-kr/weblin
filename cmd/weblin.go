// Copyright 2024 JongHoon Shim and The weblin Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

//go:build linux

/*
Package cmd 명령어 처리 패키지
*/
package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/hoon-kr/weblin/config"
	"github.com/hoon-kr/weblin/internal/server"
	"github.com/spf13/cobra"
	"go.uber.org/automaxprocs/maxprocs"
)

// weblinCmd 하위 명령어 없이 실행될 때, 최상위 명령어
var weblinCmd = &cobra.Command{
	Use:   "weblin",
	Short: "weblin controls Linux servers on the web.",
	Long: `weblin can log in to the corresponding Linux server on the web
and do all the work through the web terminal. In addition, various processes such as adding, 
creating, and deleting files can be easily performed through the UI.`,
	Version: config.Version,
}

// startCmd 서버 가동 명령어
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Run weblin (normal mode)",
	RunE:  wrapCommandFuncForCobra(server.StartServer),
}

// debugCmd 서버 디버그 모드 가동 명령어
var debugCmd = &cobra.Command{
	Use:   "debug",
	Short: "Run weblin (debug mode)",
	RunE:  wrapCommandFuncForCobra(server.StartServer),
}

// stopCmd 서버 정지 명령어
var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop weblin",
	RunE:  wrapCommandFuncForCobra(server.StopServer),
}

// init cmd 패키지 임포트 시 자동 초기화
func init() {
	weblinCmd.AddCommand(startCmd)
	weblinCmd.AddCommand(debugCmd)
	weblinCmd.AddCommand(stopCmd)
}

// Execute 명령어 실행
func Execute() {
	// GOMAXPROCS 값 최적화
	undo, err := maxprocs.Set()
	if err != nil {
		fmt.Fprintf(os.Stderr, "[WARNING] failed to set GOMAXPROCS: %s\n", err)
	}
	defer undo()

	// 명령어 및 플래그 처리
	err = weblinCmd.Execute()
	if err != nil {
		var exitErr *config.ExitError
		if errors.As(err, &exitErr) {
			os.Exit(exitErr.ExitCode)
		}
		os.Exit(1)
	}
}

// wrapCommandFuncForCobra cobra.Command의 RunE 필드 랩핑 함수
//
// Parameters:
//   - f: 명령어 함수
//
// Returns:
//   - error: 정상 종료(nil), 비정상 종료(error)
func wrapCommandFuncForCobra(f func(cmd *cobra.Command) (int, error)) func(cmd *cobra.Command, _ []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		status, err := f(cmd)
		if status > 1 {
			cmd.SilenceErrors = true
			return &config.ExitError{ExitCode: status, Err: err}
		}
		return err
	}
}
