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
Package config 전역 설정 패키지
*/
package config

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

var (
	BuildTime  = "unknown"     // 빌드 시 값 세팅됨
	Version    = "1.0.0"
	ModuleName = "weblin"
)

const (
	ConfFilePath       = "conf/weblin.properties"
	PidFilePath        = "var/weblin.pid"
	ConsoleLogFilePath = "log/weblin.log"
	JsonLogFilePath    = "log/weblin_json.log"
)

// 종료 코드 정의
const (
	ExitCodeSuccess = iota
	ExitCodeFailure
	ExitCodeFatal
)

// 종료 메시지 정의
const (
	ExitSuccess = "exit success"
	ExitFailure = "exit failure"
	ExitFatal   = "exit fatal"
)

// ExitError 종료 코드 정보 구조체
type ExitError struct {
	ExitCode int
	Err      error
}

// Error 종료 코드를 메시지로 변환
//
// Returns:
//   - string: error
func (e *ExitError) Error() string {
	if e.Err == nil {
		return fmt.Sprintf("exiting with status %d", e.ExitCode)
	}
	return e.Err.Error()
}

// Config 전역 설정 정보 구조체
type Config struct {
	// 최대 로그 파일 사이즈 (DEF:100MB, MIN:1MB, MAX:1000MB)
	MaxLogFileSize int
	// 최대 로그 파일 백업 개수 (DEF:10, MIN:1, MAX:100)
	MaxLogFileBackup int
	// 최대 백업 로그 파일 유지 기간(일) (DEF:90, MIN:1, MAX:365)
	MaxLogFileAge int
	// 백업 로그 파일 압축 여부 (DEF:true, ENABLE:true, DISABLE:false)
	CompBakLogFile bool
}

// RunConfig 런타임 전역 설정 정보 구조체
type RunConfig struct {
	DebugMode bool
	Pid       int
}

var Conf Config
var RunConf RunConfig

// init config 패키지 임포트 시 자동 초기화
func init() {
	Conf.MaxLogFileSize = 100
	Conf.MaxLogFileBackup = 10
	Conf.MaxLogFileAge = 90
	Conf.CompBakLogFile = true
}

// LoadConfig 설정 파일 로드
//
// Parameters:
//   - filePath: 설정 파일 경로
//
// Returns:
//   - error: 성공(nil), 실패(error)
func LoadConfig(filePath string) error {
	// 설정 파일 파싱
	config, err := parseConfig(filePath)
	if err != nil {
		return err
	}

	if valueStr, exists := config["MaxLogFileSize"]; exists {
		value, err := strconv.Atoi(valueStr)
		if err != nil && value >= 1 && value <= 1000 {
			Conf.MaxLogFileSize = value
		}
	}

	if valueStr, exists := config["MaxLogFileBackup"]; exists {
		value, err := strconv.Atoi(valueStr)
		if err != nil && value >= 1 && value <= 100 {
			Conf.MaxLogFileBackup = value
		}
	}

	if valueStr, exists := config["MaxLogFileAge"]; exists {
		value, err := strconv.Atoi(valueStr)
		if err != nil && value >= 1 && value <= 365 {
			Conf.MaxLogFileAge = value
		}
	}

	if valueStr, exists := config["CompressBackupLogFile"]; exists {
		if strings.ToLower(valueStr) == "no" {
			Conf.CompBakLogFile = false
		}
	}

	return nil
}

// parseConfig 설정 파일을 파싱하여 맵에 저장
//
// Parameters:
//   - filePath: 설정 파일 경로
//
// Returns:
//   - map[string]string: 설정 정보 맵
//   - error: 성공(nil), 실패(error)
func parseConfig(filePath string) (map[string]string, error) {
	config := make(map[string]string)

	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %s", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// 비어있거나 주석 처리된 라인은 무시
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// 각 라인을 key, value 형태로 분리
		parts := strings.Fields(line)
		if len(parts) != 2 {
			continue
		}
		key := parts[0]
		value := parts[1]

		// 설정 정보 맵에 저장
		config[key] = value
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading config file: %s", err)
	}

	return config, nil
}
