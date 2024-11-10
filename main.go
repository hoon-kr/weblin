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
Package main 모듈 시작 패키지
*/
package main

import "github.com/hoon-kr/weblin/cmd"

// main 모듈 메인 함수
func main() {
	// 명령 실행
	cmd.Execute()
}
