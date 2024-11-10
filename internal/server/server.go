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
Package server 메인 서버 패키지
*/
package server

import (
	"fmt"
	"io"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/hoon-kr/weblin/config"
	"github.com/hoon-kr/weblin/internal/logger"
	"github.com/hoon-kr/weblin/pkg/utils/file"
	"github.com/hoon-kr/weblin/pkg/utils/process"
	"github.com/spf13/cobra"
)

// StartServer 서버 가동
//
// Parameters:
//   - cmd: 명령어 정보
//
// Returns:
//   - int: 정상 종료(0), 비정상 종료(>=1)
//   - error: 정상 종료(nil), 비정상 종료(error)
func StartServer(cmd *cobra.Command) (int, error) {
	if cmd == nil {
		fmt.Fprintf(os.Stderr, "[WARNING] invalid parameter: [*cobra.Command] is nil\n")
		return config.ExitCodeFailure, fmt.Errorf("%s(%d)", config.ExitFailure, config.ExitCodeFailure)
	}

	// 작업 경로를 실행 파일이 위치한 경로로 변경
	err := file.ChangeWorkPathToModulePath()
	if err != nil {
		fmt.Fprintf(os.Stderr, "[ERROR] %s\n", err)
		return config.ExitCodeFailure, fmt.Errorf("%s(%d)", config.ExitFailure, config.ExitCodeFailure)
	}

	// 이미 동작 중인 프로세스가 존재하는지 확인
	var pid int
	if isRunning(&pid) {
		fmt.Fprintf(os.Stdout, "[INFO] there is already a process in operation (pid:%d)\n", pid)
		return config.ExitCodeSuccess, nil
	}

	// 데몬 프로세스 생성
	err = process.DaemonizeProcess()
	if err != nil {
		fmt.Fprintf(os.Stderr, "[ERROR] %s\n", err)
		return config.ExitCodeFailure, fmt.Errorf("%s(%d)", config.ExitFailure, config.ExitCodeFailure)
	}

	// 현재 프로세스의 PID 값 저장
	config.RunConf.Pid = os.Getpid()

	// 현재 프로세스의 PID 값을 파일에 기록
	err = file.WriteDataToTextFile(config.PidFilePath, config.RunConf.Pid, true)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[ERROR] %s\n", err)
		return config.ExitCodeFailure, fmt.Errorf("%s(%d)", config.ExitFailure, config.ExitCodeFailure)
	}

	// 디버그 모드 체크 (디버그 모드일 경우 stdout, stderr 출력)
	if cmd.Use == "debug" {
		config.RunConf.DebugMode = true
	} else {
		os.Stdout = nil
		os.Stderr = nil
	}

	// 시그널 설정
	sigChan := setupSignal()

	// 서버 초기화
	initialization()
	// 서버 종료 시 자원 정리
	defer finalization()

	logger.Log.LogInfo("Start %s (pid:%d, mode:%s)", config.ModuleName, config.RunConf.Pid,
		func() string {
			if config.RunConf.DebugMode {
				return "debug"
			}
			return "normal"
		}())

	// 종료 시그널 대기 (SIGINT, SIGTERM)
	sig := <-sigChan
	logger.Log.LogInfo("Received %s signal (%d)", sig.String(), sig)

	return config.ExitCodeSuccess, nil
}

// StopServer 서버 정지
//
// Parameters:
//   - cmd: 명령어 정보
//
// Returns:
//   - int: 정상 종료(0), 비정상 종료(>=1)
//   - error: 정상 종료(nil), 비정상 종료(error)
func StopServer(cmd *cobra.Command) (int, error) {
	if cmd == nil {
		fmt.Fprintf(os.Stderr, "[WARNING] invalid parameter: [*cobra.Command] is nil\n")
		return config.ExitCodeFailure, fmt.Errorf("%s(%d)", config.ExitFailure, config.ExitCodeFailure)
	}

	// 작업 경로를 실행 파일이 위치한 경로로 변경
	err := file.ChangeWorkPathToModulePath()
	if err != nil {
		fmt.Fprintf(os.Stderr, "[ERROR] %s\n", err)
		return config.ExitCodeFailure, fmt.Errorf("%s(%d)", config.ExitFailure, config.ExitCodeFailure)
	}

	// 동작 중인 프로세스가 존재하는지 확인
	var pid int
	if !isRunning(&pid) {
		return config.ExitCodeSuccess, nil
	}

	// 서버에 정지 시그널 전송 (SIGTERM)
	if err := process.SendSignal(pid, syscall.SIGTERM); err != nil {
		fmt.Fprintf(os.Stderr, "[WARNING] %s\n", err)
		return config.ExitCodeFailure, fmt.Errorf("%s(%d)", config.ExitFailure, config.ExitCodeFailure)
	}

	return config.ExitCodeSuccess, nil
}

// isRunning 서버가 동작 중인지 확인
//
// Returns:
//   - bool: 동작(true), 미동작(false)
func isRunning(pid *int) bool {
	if pid == nil {
		return false
	}

	file, err := os.Open(config.PidFilePath)
	if err != nil {
		return false
	}
	defer file.Close()

	// PID 값 읽기
	pidStr, err := io.ReadAll(file)
	if err != nil {
		return false
	}

	// PID 값을 정수로 변환
	*pid, err = strconv.Atoi(string(pidStr))
	if err != nil {
		return false
	}

	// 프로세스 동작 확인
	return process.IsProcessRun(*pid)
}

// setupSignal 시그널 설정
//
// Returns:
//   - chan os.Signal: signal channel
func setupSignal() chan os.Signal {
	sigChan := make(chan os.Signal, 1)
	// 수신할 시그널 설정 (SIGINT, SIGTERM)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	// 무시할 시그널 설정
	signal.Ignore(syscall.SIGABRT, syscall.SIGALRM, syscall.SIGFPE, syscall.SIGHUP,
		syscall.SIGILL, syscall.SIGPROF, syscall.SIGQUIT, syscall.SIGTSTP,
		syscall.SIGVTALRM)

	return sigChan
}

// initialization 서버 초기화
func initialization() {
	// 설정 파일 로드
	config.LoadConfig(config.ConfFilePath)
	// 로거 초기화
	logger.Log.InitializeLogger()
}

// finalization 서버 종료 시 자원 정리
func finalization() {
	// 로그 자원 정리
	logger.Log.FinalizeLogger()
}
