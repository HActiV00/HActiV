package configs

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// 각 도구별 Rule 구조체 정의
type BaseRule struct {
	EventType          string   `json:"event_type"`
	EventName          string   `json:"event_name"`
	Description        string   `json:"description"`
	Action             string   `json:"action"`
	EffectiveDays      []string `json:"effective_days,omitempty"`
	EffectiveStartTime string   `json:"effective_start_time,omitempty"`
	EffectiveEndTime   string   `json:"effective_end_time,omitempty"`
}

type ExecveRule struct {
	BaseRule
	ProcessNames []string `json:"process_names,omitempty"`
	UIDs         []uint32 `json:"uids,omitempty"`
	GIDs         []uint32 `json:"gids,omitempty"`
}

type DeleteRule struct {
	BaseRule
	ProcessNames []string `json:"process_names,omitempty"`
	Filenames    []string `json:"filenames,omitempty"`
	UIDs         []uint32 `json:"uids,omitempty"`
	GIDs         []uint32 `json:"gids,omitempty"`
}

type MemSize struct {
	Threshold int    `json:"threshold"`
	Unit      string `json:"unit"`
}

type MemoryRule struct {
	BaseRule
	ProcessNames       []string `json:"process_names,omitempty"`
	UIDs               []uint32 `json:"uids,omitempty"`
	GIDs               []uint32 `json:"gids,omitempty"`
	Size               MemSize  `json:"size,omitempty"`
	StartAddr          int      `json:"start_addr,omitempty"`
	EndAddr            int      `json:"end_addr,omitempty"`
	Syscalls           []string `json:"syscalls,omitempty"`
	ExecPath           string   `json:"exec_path,omitempty"`
	ExecCommand        string   `json:"exec_command,omitempty"`
	MprotectConditions struct {
		Threshold        int      `json:"threshold"`
		Unit             string   `json:"unit"`
		SpecificPatterns []string `json:"specific_patterns"`
	} `json:"mprotect_conditions,omitempty"`
}

type NetworkRule struct {
	BaseRule
	ProcessNames    []string `json:"process_names,omitempty"`
	SrcIPs          []string `json:"src_ips,omitempty"`
	DstIPs          []string `json:"dst_ips,omitempty"`
	Protocol        []string `json:"protocols,omitempty"`
	PacketThreshold int      `json:"packet_threshold,omitempty"`
}

type OpenRule struct {
	BaseRule
	ProcessNames []string `json:"process_names,omitempty"`
	Filenames    []string `json:"filenames,omitempty"`
	UIDs         []uint32 `json:"uids,omitempty"`
	GIDs         []uint32 `json:"gids,omitempty"`
}

var allRules map[string]interface{}
var RuleDir string

// 초기 규칙 파일 설정 함수
func InitRules() error {
	RuleDir = RuleLocation // setting.go에서 가져온 RuleLocation 값 사용

	toolsDir := "../tools"
	files, err := os.ReadDir(toolsDir)
	if err != nil {
		return fmt.Errorf("tools 디렉터리 읽기 실패: %v", err)
	}

	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".go") {
			if strings.Contains(file.Name(), "metrics") {
				fmt.Printf("%s 파일 건너뜀 (metrics 포함)\n", file.Name())
				continue
			}

			toolName := strings.TrimSuffix(file.Name(), ".go")
			ruleFilePath := filepath.Join(RuleDir, fmt.Sprintf("%srule.json", toolName))

			if _, err := os.Stat(ruleFilePath); os.IsNotExist(err) {
				fmt.Printf("규칙 파일이 없습니다. 생성 중: %s\n", ruleFilePath)
				if err := createRuleFile(ruleFilePath); err != nil {
					fmt.Printf("규칙 파일 생성 실패: %v\n", err)
				} else {
					fmt.Printf("규칙 파일 생성 완료: %s\n", ruleFilePath)
				}
			}
		}
	}
	return nil
}

// 기본 규칙 파일 생성
func createRuleFile(filePath string) error {
	if err := os.MkdirAll(RuleDir, 0755); err != nil {
		return fmt.Errorf("규칙 디렉터리 생성 실패: %v", err)
	}

	file, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("규칙 파일 생성 실패: %v", err)
	}
	defer file.Close()

	_, err = file.WriteString(`[]`)
	return err
}

// 특정 도구에 대한 규칙 설정
func SetupRules(toolName string) error {
	filePath := filepath.Join(RuleDir, fmt.Sprintf("%srule.json", toolName))
	rules, err := LoadRules(filePath, toolName)
	if err != nil {
		return fmt.Errorf("%s의 규칙 로드 실패: %v", toolName, err)
	}

	if allRules == nil {
		allRules = make(map[string]interface{})
	}
	allRules[toolName] = rules

	AppliedRules(toolName)
	return nil
}

// JSON 파일에서 규칙 로드
func LoadRules(filePath string, toolName string) (interface{}, error) {
	file, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("규칙 파일 읽기 실패 %s: %v", filePath, err)
	}

	switch toolName {
	case "execve":
		var rules []ExecveRule
		if err := json.Unmarshal(file, &rules); err != nil {
			return nil, fmt.Errorf("규칙 파일 파싱 실패 %s: %v", filePath, err)
		}
		for i := range rules {
			if err := convertEffectiveTime(&rules[i].BaseRule); err != nil {
				return nil, fmt.Errorf("time conversion error in rule %s: %v", rules[i].EventName, err)
			}
		}
		return rules, nil
	case "delete":
		var rules []DeleteRule
		if err := json.Unmarshal(file, &rules); err != nil {
			return nil, fmt.Errorf("규칙 파일 파싱 실패 %s: %v", filePath, err)
		}
		for i := range rules {
			if err := convertEffectiveTime(&rules[i].BaseRule); err != nil {
				return nil, fmt.Errorf("time conversion error in rule %s: %v", rules[i].EventName, err)
			}
		}
		return rules, nil
	case "memory":
		var rules []MemoryRule
		if err := json.Unmarshal(file, &rules); err != nil {
			return nil, fmt.Errorf("규칙 파일 파싱 실패 %s: %v", filePath, err)
		}
		for i := range rules {
			if err := convertEffectiveTime(&rules[i].BaseRule); err != nil {
				return nil, fmt.Errorf("time conversion error in rule %s: %v", rules[i].EventName, err)
			}
		}
		return rules, nil
	case "network":
		var rules []NetworkRule
		if err := json.Unmarshal(file, &rules); err != nil {
			return nil, fmt.Errorf("규칙 파일 파싱 실패 %s: %v", filePath, err)
		}
		for i := range rules {
			if err := convertEffectiveTime(&rules[i].BaseRule); err != nil {
				return nil, fmt.Errorf("time conversion error in rule %s: %v", rules[i].EventName, err)
			}
		}
		return rules, nil
	case "open":
		var rules []OpenRule
		if err := json.Unmarshal(file, &rules); err != nil {
			return nil, fmt.Errorf("규칙 파일 파싱 실패 %s: %v", filePath, err)
		}
		for i := range rules {
			if err := convertEffectiveTime(&rules[i].BaseRule); err != nil {
				return nil, fmt.Errorf("time conversion error in rule %s: %v", rules[i].EventName, err)
			}
		}
		return rules, nil
	default:
		return nil, fmt.Errorf("알 수 없는 도구 이름: %s", toolName)
	}
}

// effective_start_time과 effective_end_time을 time.Duration으로 변환
func convertEffectiveTime(rule *BaseRule) error {
	if rule.EffectiveStartTime != "" {
		duration, err := time.ParseDuration(rule.EffectiveStartTime)
		if err != nil {
			return fmt.Errorf("invalid effective_start_time: %v", err)
		}
		rule.EffectiveStartTime = duration.String()
	}
	if rule.EffectiveEndTime != "" {
		duration, err := time.ParseDuration(rule.EffectiveEndTime)
		if err != nil {
			return fmt.Errorf("invalid effective_end_time: %v", err)
		}
		rule.EffectiveEndTime = duration.String()
	}
	return nil
}

// 도구별 필수 필드를 반환하는 함수
func GetRequiredFields(toolName string) []string {
	switch toolName {
	case "execve":
		return []string{"event_type", "event_name", "description", "action", "process_names", "uids", "gids"}
	case "delete":
		return []string{"event_type", "event_name", "description", "action", "filenames", "process_names", "uids", "gids"}
	case "memory":
		return []string{"event_type", "event_name", "description", "action", "size", "process_names", "uids", "gids", "syscalls", "exec_path", "exec_command"}
	case "network":
		return []string{"event_type", "event_name", "description", "action", "src_ips", "dst_ips", "protocol", "packet_threshold", "process_names"}
	case "open":
		return []string{"event_type", "event_name", "description", "action", "filenames", "process_names", "uids", "gids"}
	default:
		return []string{}
	}
}

// 적용된 규칙 출력
func AppliedRules(toolName string) {
	fmt.Printf("Tool: %s에 적용된 규칙들\n", toolName)
	rules := allRules[toolName]

	switch r := rules.(type) {
	case []ExecveRule:
		for _, rule := range r {
			printExecveRule(rule)
		}
	case []DeleteRule:
		for _, rule := range r {
			printDeleteRule(rule)
		}
	case []MemoryRule:
		for _, rule := range r {
			printMemoryRule(rule)
		}
	case []NetworkRule:
		for _, rule := range r {
			printNetworkRule(rule)
		}
	case []OpenRule:
		for _, rule := range r {
			printOpenRule(rule)
		}
	}
}

// ExecveRule 출력 함수
func printExecveRule(rule ExecveRule) {
	fmt.Printf(" - EventType: %s, EventName: %s, Description: %s, Action: %s, EffectiveDays: %v, EffectiveStartTime: %s, EffectiveEndTime: %s, ProcessNames: %v, UIDs: %v, GIDs: %v\n",
		rule.EventType, rule.EventName, rule.Description, rule.Action,
		rule.EffectiveDays, rule.EffectiveStartTime, rule.EffectiveEndTime,
		rule.ProcessNames, rule.UIDs, rule.GIDs)
}

// DeleteRule 출력 함수
func printDeleteRule(rule DeleteRule) {
	fmt.Printf(" - EventType: %s, EventName: %s, Description: %s, Action: %s, EffectiveDays: %v, EffectiveStartTime: %s, EffectiveEndTime: %s, Filenames: %v, ProcessNames: %v, UIDs: %v, GIDs: %v\n",
		rule.EventType, rule.EventName, rule.Description, rule.Action,
		rule.EffectiveDays, rule.EffectiveStartTime, rule.EffectiveEndTime,
		rule.Filenames, rule.ProcessNames, rule.UIDs, rule.GIDs)
}

// MemoryRule 출력 함수
func printMemoryRule(rule MemoryRule) {
	fmt.Printf(" - EventType: %s, EventName: %s, Description: %s, Action: %s, EffectiveDays: %v, EffectiveStartTime: %s, EffectiveEndTime: %s, Size: %+v, ProcessNames: %v, UIDs: %v, GIDs: %v, Syscalls: %v, ExecPath: %s, ExecCommand: %s, MprotectConditions: %+v\n",
		rule.EventType, rule.EventName, rule.Description, rule.Action,
		rule.EffectiveDays, rule.EffectiveStartTime, rule.EffectiveEndTime,
		rule.Size, rule.ProcessNames, rule.UIDs, rule.GIDs,
		rule.Syscalls, rule.ExecPath, rule.ExecCommand, rule.MprotectConditions)
}

// NetworkRule 출력 함수
func printNetworkRule(rule NetworkRule) {
	fmt.Printf(" - EventType: %s, EventName: %s, Description: %s, Action: %s, EffectiveDays: %v, EffectiveStartTime: %s, EffectiveEndTime: %s, SrcIPs: %v, DstIPs: %v, Protocols: %v, PacketThreshold: %d, ProcessNames: %v\n",
		rule.EventType, rule.EventName, rule.Description, rule.Action,
		rule.EffectiveDays, rule.EffectiveStartTime, rule.EffectiveEndTime,
		rule.SrcIPs, rule.DstIPs, rule.Protocol, rule.PacketThreshold, rule.ProcessNames)
}

// OpenRule 출력 함수
func printOpenRule(rule OpenRule) {
	fmt.Printf(" - EventType: %s, EventName: %s, Description: %s, Action: %s, EffectiveDays: %v, EffectiveStartTime: %s, EffectiveEndTime: %s, Filenames: %v, ProcessNames: %v, UIDs: %v, GIDs: %v\n",
		rule.EventType, rule.EventName, rule.Description, rule.Action,
		rule.EffectiveDays, rule.EffectiveStartTime, rule.EffectiveEndTime,
		rule.Filenames, rule.ProcessNames, rule.UIDs, rule.GIDs)
}

// ParseRules 함수

func ParseRules(eventType string, eventData map[string]interface{}, toolName string) bool {
	rulesInterface := allRules[toolName]
	requiredFields := GetRequiredFields(toolName)

	matched := false

	switch rules := rulesInterface.(type) {
	case []ExecveRule:
		for _, rule := range rules {
			matched = matchExecveRule(rule, eventData, requiredFields)
			if matched {
				//fmt.Println("Event matched rule:", rule.EventName)
				return rule.Action != "ignore"
			}
		}
	case []DeleteRule:
		for _, rule := range rules {
			matched = matchDeleteRule(rule, eventData, requiredFields)
			if matched {
				//fmt.Println("Event matched rule:", rule.EventName)
				return rule.Action != "ignore"
			}
		}
	case []MemoryRule:
		for _, rule := range rules {
			matched = matchMemoryRule(rule, eventData, requiredFields)
			if matched {
				fmt.Println("Event matched rule:", rule.EventName)
				if rule.Action == "ignore" {
					fmt.Println("Ignoring event as per rule:", rule.EventName)
					return false
				}
				return true
			}
		}
	case []NetworkRule:
		for _, rule := range rules {
			matched = matchNetworkRule(rule, eventData, requiredFields)
			if matched {
				//fmt.Println("Event matched rule:", rule.EventName)
				return rule.Action != "ignore"
			}
		}
	case []OpenRule:
		for _, rule := range rules {
			matched = matchOpenRule(rule, eventData, requiredFields)
			if matched {
				//fmt.Println("Event matched rule:", rule.EventName)
				return rule.Action != "ignore"
			}
		}
	}
	return false
}

// 각 도구별 매칭 함수들
func matchExecveRule(rule ExecveRule, eventData map[string]interface{}, requiredFields []string) bool {
	matched := true
	if contains("process_name", requiredFields) && len(rule.ProcessNames) > 0 && rule.ProcessNames[0] != "*" {
		if processName, ok := eventData["process_name"].(string); ok {
			if !matchesString(processName, rule.ProcessNames) {
				matched = false
			}
		} else {
			matched = false
		}
	}
	if contains("uids", requiredFields) && len(rule.UIDs) > 0 {
		if uid, ok := eventData["uid"].(uint32); ok {
			if !containsUint32(uid, rule.UIDs) {
				matched = false
			}
		} else {
			matched = false
		}
	}
	if contains("gids", requiredFields) && len(rule.GIDs) > 0 {
		if gid, ok := eventData["gid"].(uint32); ok {
			if !containsUint32(gid, rule.GIDs) {
				matched = false
			}
		} else {
			matched = false
		}
	}
	return matched
}

func matchDeleteRule(rule DeleteRule, eventData map[string]interface{}, requiredFields []string) bool {
	matched := true
	if contains("filenames", requiredFields) && len(rule.Filenames) > 0 {
		if filename, ok := eventData["filename"].(string); ok {
			if !matchesString(filename, rule.Filenames) {
				matched = false
			}
		} else {
			matched = false
		}
	}
	// ProcessNames, UIDs, GIDs 매칭은 위와 동일하게 적용
	matched = matched && matchExecveRule(ExecveRule{ProcessNames: rule.ProcessNames, UIDs: rule.UIDs, GIDs: rule.GIDs}, eventData, requiredFields)
	return matched
}

func matchMemoryRule(rule MemoryRule, eventData map[string]interface{}, requiredFields []string) bool {
	matched := true
	if contains("size", requiredFields) {
		if size, ok := eventData["size"].(int); ok {
			threshold := convertToBytes(rule.Size.Threshold, rule.Size.Unit)
			if size < threshold {
				matched = false
			}
		} else {
			matched = false
		}
	}
	// ProcessNames 매칭
	if contains("process_name", requiredFields) && len(rule.ProcessNames) > 0 && rule.ProcessNames[0] != "*" {
		if processName, ok := eventData["process_name"].(string); ok {
			if !matchesString(processName, rule.ProcessNames) {
				matched = false
			}
		} else {
			matched = false
		}
	}
	// 새로운 로직 추가
	if rule.ExecPath != "" {
		if execPath, ok := eventData["exec_path"].(string); ok {
			if execPath != rule.ExecPath {
				matched = false
			}
		} else {
			matched = false
		}
	}
	if rule.ExecCommand != "" {
		if execCommand, ok := eventData["exec_command"].(string); ok {
			if !strings.HasPrefix(execCommand, rule.ExecCommand) {
				matched = false
			}
		} else {
			matched = false
		}
	}
	if len(rule.Syscalls) > 0 {
		if syscall, ok := eventData["syscall"].(string); ok {
			if !contains(syscall, rule.Syscalls) {
				matched = false
			}
		} else {
			matched = false
		}
	}
	if rule.MprotectConditions.Threshold > 0 {
		if size, ok := eventData["size"].(int); ok {
			threshold := convertToBytes(rule.MprotectConditions.Threshold, rule.MprotectConditions.Unit)
			if size < threshold {
				matched = false
			}
		} else {
			matched = false
		}
	}
	if len(rule.MprotectConditions.SpecificPatterns) > 0 {
		if pattern, ok := eventData["pattern"].(string); ok {
			if !contains(pattern, rule.MprotectConditions.SpecificPatterns) {
				matched = false
			}
		} else {
			matched = false
		}
	}
	return matched
}

func matchNetworkRule(rule NetworkRule, eventData map[string]interface{}, requiredFields []string) bool {
	matched := true
	if contains("src_ips", requiredFields) && len(rule.SrcIPs) > 0 {
		if srcIP, ok := eventData["src_ip"].(string); ok {
			if !matchesIP(srcIP, rule.SrcIPs) {
				matched = false
			}
		} else {
			matched = false
		}
	}
	if contains("dst_ips", requiredFields) && len(rule.DstIPs) > 0 {
		if dstIP, ok := eventData["dst_ip"].(string); ok {
			if !matchesIP(dstIP, rule.DstIPs) {
				matched = false
			}
		} else {
			matched = false
		}
	}
	if contains("packet_threshold", requiredFields) && rule.PacketThreshold > 0 {
		if packets, ok := eventData["packet_count"].(int); ok {
			if packets < rule.PacketThreshold {
				matched = false
			}
		} else {
			matched = false
		}
	}
	if contains("protocol", requiredFields) && len(rule.Protocol) > 0 {
		if protocol, ok := eventData["protocol"].(string); ok {
			if !matchesIP(protocol, rule.Protocol) {
				matched = false
			}
		} else {
			matched = false
		}
	}
	return matched
}

func matchOpenRule(rule OpenRule, eventData map[string]interface{}, requiredFields []string) bool {
	matched := true
	// filenames 매칭
	matched = matched && matchDeleteRule(DeleteRule{Filenames: rule.Filenames, ProcessNames: rule.ProcessNames, UIDs: rule.UIDs, GIDs: rule.GIDs}, eventData, requiredFields)
	return matched
}

// 헬퍼 함수들

// CheckEffectiveTime 함수: 주어진 요일과 시간이 규칙의 유효한 날짜 및 시간 조건에 부합하는지 확인
func CheckEffectiveTime(toolName, currentDay, currentTimeStr string) bool {
	rules := allRules[toolName]

	// 각 도구별로 규칙을 확인
	switch r := rules.(type) {
	case []ExecveRule:
		for _, rule := range r {
			if isTimeInEffectiveRange(rule.BaseRule, currentDay, currentTimeStr) {
				return true
			}
		}
	case []DeleteRule:
		for _, rule := range r {
			if isTimeInEffectiveRange(rule.BaseRule, currentDay, currentTimeStr) {
				return true
			}
		}
	case []MemoryRule:
		for _, rule := range r {
			if isTimeInEffectiveRange(rule.BaseRule, currentDay, currentTimeStr) {
				return true
			}
		}
	case []NetworkRule:
		for _, rule := range r {
			if isTimeInEffectiveRange(rule.BaseRule, currentDay, currentTimeStr) {
				return true
			}
		}
	case []OpenRule:
		for _, rule := range r {
			if isTimeInEffectiveRange(rule.BaseRule, currentDay, currentTimeStr) {
				return true
			}
		}
	}
	return false
}

// isTimeInEffectiveRange 함수: 규칙의 유효 시간 범위와 현재 요일 및 시간을 비교하여 유효 여부 반환
func isTimeInEffectiveRange(rule BaseRule, currentDay, currentTimeStr string) bool {
	// 유효한 요일인지 확인
	if len(rule.EffectiveDays) > 0 && !contains(currentDay, rule.EffectiveDays) {
		return false
	}

	// 유효한 시작 및 종료 시간 확인
	if rule.EffectiveStartTime != "" && rule.EffectiveEndTime != "" {
		startTime, err1 := time.Parse("15:04:05", rule.EffectiveStartTime)
		endTime, err2 := time.Parse("15:04:05", rule.EffectiveEndTime)
		currentTime, err3 := time.Parse("15:04:05", currentTimeStr)

		if err1 != nil || err2 != nil || err3 != nil {
			fmt.Printf("Error parsing time: start=%v, end=%v, current=%v\n", err1, err2, err3)
			return false
		}

		if currentTime.Before(startTime) || currentTime.After(endTime) {
			return false
		}
	}

	return true
}

// contains 함수: 문자열이 리스트에 있는지 확인하는 함수
func contains(value string, list []string) bool {
	for _, v := range list {
		if v == value {
			return true
		}
	}
	return false
}

func matchesString(value string, allowed []string) bool {
	for _, allowedValue := range allowed {
		if value == allowedValue || allowedValue == "*" {
			return true
		}
	}
	return false
}

func containsUint32(value uint32, allowed []uint32) bool {
	for _, allowedValue := range allowed {
		if value == allowedValue {
			return true
		}
	}
	return false
}

func matchesIP(ip string, ipList []string) bool {
	for _, cidr := range ipList {
		_, ipNet, err := net.ParseCIDR(cidr)
		if err != nil {
			if ip == cidr {
				return true
			}
			continue
		}
		if ipNet.Contains(net.ParseIP(ip)) {
			return true
		}
	}
	return false
}

func convertToBytes(value int, unit string) int {
	switch strings.ToUpper(unit) {
	case "KB":
		return value * 1024
	case "MB":
		return value * 1024 * 1024
	case "GB":
		return value * 1024 * 1024 * 1024
	default:
		return value
	}
}
