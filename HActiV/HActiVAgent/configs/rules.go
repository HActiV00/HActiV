// Copyright Authors of HActiV

// configs package for Setting and detection rules
package configs

import (
	"HActiV/pkg/utils"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"
)

type TimeCondition struct {
	Day        string      `json:"day"`
	TimeRanges []TimeRange `json:"time_ranges"`
}

type TimeRange struct {
	Start string `json:"start"`
	End   string `json:"end"`
}

type Policy struct {
	Condition      string          `json:"condition"`
	Action         string          `json:"action"`
	PrintFormat    string          `json:"print_format"`
	TimeConditions []TimeCondition `json:"time_conditions"`
}

type Rule struct {
	EventName      string          `json:"event_name"`
	Description    string          `json:"description"`
	Usage          bool            `json:"usage"`
	Condition      string          `json:"condition"`
	Action         string          `json:"action"`
	PrintFormat    string          `json:"print_format"`
	TimeConditions []TimeCondition `json:"time_conditions"`
}

func LoadRules(toolName string) ([]Policy, error) {
	filename := RuleLocation + toolName + "rule.json"
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return nil, fmt.Errorf("규칙 파일이 존재하지 않습니다: %s", filename)
	}

	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	var rules []Rule
	err = json.Unmarshal(data, &rules)
	if err != nil {
		return nil, err
	}
	fieldTypes, err := getFieldTypes()
	if err != nil {
		return nil, err
	}

	var policies []Policy
	for i, rule := range rules {
		if rule.Usage {
			rule.Condition = strings.TrimSpace(rule.Condition)
			rule.Action = strings.TrimSpace(rule.Action)
			conditionCheck, conditionErrMsg := evaluateCondition(rule.Condition, fieldTypes)
			timeCheck, timeErrMsg := evaluateTime(rule.TimeConditions)
			if conditionCheck && timeCheck {
				fmt.Printf("정책: %s\n", rule.EventName)
				fmt.Printf("  설명: %s\n", rule.Description)
				fmt.Printf("  조건: %s\n", rule.Condition)
				fmt.Printf("  액션: %s\n", rule.Action)
				fmt.Printf("  출력: %s\n", rule.PrintFormat)
				fmt.Printf("  시간: %s\n", rule.TimeConditions)

				policies = append(policies, Policy{
					Condition:      rule.Condition,
					Action:         rule.Action,
					PrintFormat:    rule.PrintFormat,
					TimeConditions: rule.TimeConditions,
				})
			} else {
				fmt.Printf("정책 '%s'의 조건이 유효하지 않습니다: %s %s\n", rule.EventName, conditionErrMsg, timeErrMsg)

				rules[i].Usage = false
			}
		}
	}

	updatedData, err := json.MarshalIndent(rules, "", "  ")
	if err != nil {
		return nil, err
	}
	err = os.WriteFile(filename, updatedData, 0644)
	if err != nil {
		return nil, err
	}
	return policies, nil
}

func getFieldTypes() (map[string]string, error) {
	var event utils.Event
	fieldTypes := make(map[string]string)

	v := reflect.TypeOf(event)
	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldName := field.Name
		fieldType := field.Type.Kind().String()
		fieldTypes[fieldName] = fieldType
	}
	return fieldTypes, nil
}

func evaluateCondition(condition string, fieldTypes map[string]string) (bool, string) {
	parts := strings.Split(condition, " and ")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		tokens := strings.Fields(part)
		if len(tokens) != 3 {
			return false, fmt.Sprintf("조건 '%s'의 인자 개수가 올바르지 않습니다.", part)
		}

		fieldNameWithPercent := tokens[0]
		operator := tokens[1]
		value := tokens[2]

		validOperators := []string{"==", "!=", ">", "<", ">=", "<=", "()"}
		if !contains(validOperators, operator) {
			return false, fmt.Sprintf("%s 조건에서 연산자 '%s'가 올바르지 않습니다.", part, operator)
		}

		fieldName := strings.Trim(fieldNameWithPercent, "%")
		fieldType, ok := fieldTypes[fieldName]
		if !ok {
			return false, fmt.Sprintf("%s 조건에서 필드명 '%s'이(가) 올바르지 않습니다.", part, fieldName)
		}

		switch fieldType {
		case "uint32", "int", "uint":
			if operator == "==" || operator == "!=" || operator == ">" || operator == "<" || operator == ">=" || operator == "<=" {
				if _, err := strconv.Atoi(value); err != nil {
					return false, fmt.Sprintf("%s 조건에서 필드 '%s'는 정수형이어야 하는데 값 '%s'이(가) 정수가 아닙니다.", part, fieldName, value)
				}
			} else {
				return false, fmt.Sprintf("%s 조건에서 필드 '%s'는 정수형인데 연산자 '%s'는 사용할 수 없습니다.", part, fieldName, operator)
			}
		case "string":
			if operator != "==" && operator != "!=" && operator != "()" {
				return false, fmt.Sprintf("%s 조건에서 필드 '%s'는 문자열인데 연산자 '%s'는 사용할 수 없습니다.", part, fieldName, operator)
			}
		}
	}
	return true, ""
}

func evaluateTime(timeCondition []TimeCondition) (bool, string) {
	weekdays := map[string]bool{
		"Monday": true, "Tuesday": true, "Wednesday": true,
		"Thursday": true, "Friday": true, "Saturday": true, "Sunday": true,
	}
	for _, timeCondition := range timeCondition {
		if !weekdays[timeCondition.Day] {
			return false, fmt.Sprintf("%s란 요일은 없습니다.", timeCondition.Day)
		}

		for _, timeRange := range timeCondition.TimeRanges {
			layout := "15:04"

			s, strErr := time.Parse(layout, timeRange.Start)
			if strErr != nil {
				return false, fmt.Sprintf("%s에서 필드 '%s'의 시간 형식이 올바르지 않습니다. 예시: 00:00", timeCondition.Day, timeRange.Start)
			}

			e, endErr := time.Parse(layout, timeRange.End)
			if endErr != nil && timeRange.End != "24:00" {
				return false, fmt.Sprintf("%s에서 필드 '%s'의 시간 형식이 올바르지 않습니다. 예시: 00:00", timeCondition.Day, timeRange.End)
			}

			if timeRange.End != "24:00" && !s.Before(e) {
				return false, fmt.Sprintf("%s에서 시작 시간(%s)은 종료 시간(%s) 이후가 될 수 없습니다.", timeCondition.Day, timeRange.Start, timeRange.End)
			}
		}
	}
	return true, ""
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func MatchedEvent(policies []Policy, event utils.Event) {
policyLoop:
	for _, policy := range policies {
		if checkPolicyTimeConditions(policy.TimeConditions, event.Time) {
			if parseAndEvaluateCondition(policy.Condition, event) {
				for _, action := range strings.Fields(policy.Action) {
					switch strings.TrimSpace(action) {
					case "ignore":
						break policyLoop
					case "print":
						fmt.Println(replaceAllPlaceholders(policy.PrintFormat, event))
					}
				}
			}
		}
	}
}

func parseAndEvaluateCondition(condition string, event utils.Event) bool {
	condition = replaceAllPlaceholders(condition, event)
	parts := strings.Split(condition, " and ")
	for _, part := range parts {
		if !evaluateSingleCondition(part) {
			return false
		}
	}
	return true
}

func evaluateSingleCondition(condition string) bool {
	parts := strings.Split(condition, " ")
	if len(parts) > 3 {
		for i, v := range parts {
			if v == "==" || v == "!=" || v == "()" {
				left := strings.Join(parts[:i], " ")
				operator := parts[i]
				right := strings.Join(parts[i+1:], " ")
				return compareString(left, operator, strings.Trim(right, "\""))
			}
		}
	}
	left := parts[0]
	operator := parts[1]
	right := parts[2]
	return compareString(left, operator, strings.Trim(right, "\""))
}

func compareString(a, op, b string) bool {
	switch op {
	case "==":
		return a == b
	case "!=":
		return a != b
	case "()":
		return strings.Contains(a, b)
	}
	aInt, _ := strconv.Atoi(a)
	bInt, _ := strconv.Atoi(b)
	switch op {
	case ">":
		return aInt > bInt
	case "<":
		return aInt < bInt
	case ">=":
		return aInt >= bInt
	case "<=":
		return aInt <= bInt
	}
	return false
}

func replaceAllPlaceholders(input string, event utils.Event) string {
	v := reflect.ValueOf(event)
	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		fieldName := t.Field(i).Name
		fieldValue := fmt.Sprintf("%v", v.Field(i).Interface())

		if fieldName == "StartAddr" || fieldName == "EndAddr" {
			fieldValue = fmt.Sprintf("0x%X", v.Field(i).Interface())
		}
		input = strings.ReplaceAll(input, "%"+fieldName+"%", fieldValue)
	}
	return input
}

func checkPolicyTimeConditions(timeCondition []TimeCondition, eventTime string) bool {
	if len(timeCondition) == 0 {
		return true
	}
	exchangeEventTimes := utils.TimeExchange(eventTime, HostRegion)
	weekday := exchangeEventTimes.Weekday().String()
	checkTime := exchangeEventTimes.Format("15:04")
	for _, timeCondition := range timeCondition {
		if timeCondition.Day != weekday {
			continue
		}
		for _, timeRange := range timeCondition.TimeRanges {
			if isTimeInRange(checkTime, timeRange.Start, timeRange.End) {
				return true
			}
		}
	}
	return false
}

func isTimeInRange(checkTime, start, end string) bool {
	layout := "15:04"
	t, _ := time.Parse(layout, checkTime)
	s, _ := time.Parse(layout, start)
	e, _ := time.Parse(layout, end)

	if end == "24:00" {
		e = e.Add(24 * time.Hour)
	}
	return (t.After(s) || t.Equal(s)) && (t.Before(e) || t.Equal(e))
}
