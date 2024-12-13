// Copyright Authors of HActiV

// configs package for Setting and detection rules
package configs

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"
)

var (
	RuleLocation string
	HostRegion   string
)

func HActiVSetting() {

	filePath := "/etc/HActiV/Setting.json"
	os.MkdirAll("/etc/HActiV/", 0755)

	_, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		file, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
		if err != nil {
			ConfigError("파일 생성 중 오류 발생")
		}
		defer file.Close()

		data := map[string]string{
			"HostMonitoring": "False",
			"RuleLocation":   "/etc/HActiV/rules",
			"API":            "TestAPIForBasic",
			"Url":            "http://localhost:8080/api/dashboard",
			"Region":         "Asia/Seoul",
			"LogLocation":    "/etc/HActiV/logs",
		}
		encoder := json.NewEncoder(file)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(data); err != nil {
			ConfigError("JSON 쓰기 중 오류 발생")
		}
	} else if err != nil {
		ConfigError("파일 존재 확인 중 오류 발생")
	}

	file, err := os.Open(filePath)
	if err != nil {
		ConfigError("파일 열기 중 오류 발생")
	}
	defer file.Close()

	var existingData map[string]string
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&existingData); err != nil {
		ConfigError("JSON 읽기 중 오류 발생")
	}

	hostMonitoring := strings.TrimSpace(existingData["HostMonitoring"])
	ruleLocation := existingData["RuleLocation"]
	api := existingData["API"]
	url := existingData["Url"]
	HostRegion = existingData["Region"]
	logLocation := existingData["LogLocation"]

	_, err = time.LoadLocation(HostRegion)
	if err != nil {
		HostRegion = "Asia/Seoul"
		fmt.Println("호스트 리전 오류로 기본값인 한국 시간으로 변경되었습니다.")
	}

	fmt.Printf("HostMonitoring: %s\n", hostMonitoring)
	fmt.Printf("RuleLocation: %s\n", ruleLocation)
	fmt.Printf("API: %s\n", api)
	fmt.Printf("Url: %s\n", url)
	fmt.Printf("Region: %s\n", HostRegion)
	fmt.Printf("LogLocation: %s\n", logLocation)

	if !strings.HasSuffix(ruleLocation, "/") {
		RuleLocation = ruleLocation + "/"
	}
	if !strings.HasSuffix(logLocation, "/") {
		logLocation = logLocation + "/"
	}
	os.MkdirAll(logLocation, 0755)
}

func ConfigError(errmsg string) {
	fmt.Println(errmsg)
	fmt.Println("프로그램을 재시작 해주세요")
	os.Exit(1)
}

func FirstRules() {
	os.MkdirAll(RuleLocation, 0755)
	for _, fileName := range []string{"delete", "execve", "memory", "network", "open"} {
		filePath := RuleLocation + fileName + "rule.json"
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			os.Create(filePath)
		}
	}
}
