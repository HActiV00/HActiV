//Copyright Authors of HActiV

// configs package for Setting and detection rules
package configs

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
)

var (
	HostMonitoring bool
	RuleLocation   string
	API            string
	URL            string
)

func InitSettings() {
	HostMonitoring, RuleLocation, API, URL = HActiVSetting()
}

func HActiVSetting() (bool, string, string, string) {
	filePath := "/etc/HActiV/Setting.json"

	// 파일 존재 여부 확인
	_, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		// 파일이 없으면 생성하고 데이터 쓰기
		file, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
		if err != nil {
			ConfigError("파일 생성 중 오류 발생")
		}
		defer file.Close()

		//Basic Value
		data := map[string]string{
			"HostMonitoring": "False",
			"RuleLocation":   "/etc/HActiV/rules",
			"API":            "TestAPIForBasic",
			"Url":            "https://localhost",
		}
		encoder := json.NewEncoder(file)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(data); err != nil {
			ConfigError("JSON 쓰기 중 오류 발생")
		}

		boolValue, err := strconv.ParseBool(data["HostMonitoring"])
		if err != nil {
			ConfigError("HostMonitoring 값 변환 오류:")
		}
		return boolValue, data["RuleLocation"], data["API"], data["Url"]

		//return data["HostMonitoring"], data["RuleLocation"], data["API"], data["Url"]
	} else if err != nil {
		ConfigError("파일 존재 확인 중 오류 발생")
	}

	// 파일이 존재하면 데이터 읽기
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

	// 필요한 값 가져오기
	hostMonitoring := strings.TrimSpace(existingData["HostMonitoring"])
	ruleLocation := existingData["RuleLocation"]
	api := existingData["API"]
	url := existingData["Url"]

	fmt.Printf("HostMonitoring: %s\n", hostMonitoring)
	fmt.Printf("RuleLocation: %s\n", ruleLocation)
	fmt.Printf("API: %s\n", api)
	fmt.Printf("Url: %s\n", url)

	boolValue, err := strconv.ParseBool(hostMonitoring)
	if err != nil {
		ConfigError("변환 오류:")
	}
	return boolValue, ruleLocation, api, url
	//return hostMonitoring, ruleLocation, api, url
}

// config error is high problem so stop HActiV
func ConfigError(errmsg string) {
	fmt.Println(errmsg)
	fmt.Println("프로그램을 재시작 해주세요")
	os.Exit(1)
}