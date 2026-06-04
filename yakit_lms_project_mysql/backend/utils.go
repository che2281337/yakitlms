package main

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func parseUintParam(c *gin.Context, name string) (uint, bool) {
	raw := c.Param(name)
	id, err := strconv.ParseUint(raw, 10, 64)
	if err != nil || id == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid " + name})
		return 0, false
	}
	return uint(id), true
}

func parseDeadline(value string, fallbackDays int) (time.Time, error) {
	if strings.TrimSpace(value) == "" {
		return time.Now().AddDate(0, 0, fallbackDays), nil
	}
	if parsed, err := time.Parse(time.RFC3339, value); err == nil {
		return parsed, nil
	}
	parsed, err := time.Parse("2006-01-02", value)
	if err != nil {
		return time.Time{}, fmt.Errorf("deadline must be YYYY-MM-DD or RFC3339")
	}
	return parsed.Add(23*time.Hour + 59*time.Minute), nil
}

func normalizeAnswer(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

func optionsFromJSON(raw string) []string {
	var options []string
	_ = json.Unmarshal([]byte(raw), &options)
	return options
}

func correctAnswersFromQuestion(q TestQuestion) []string {
	if strings.TrimSpace(q.CorrectAnswersJSON) != "" {
		var answers []string
		_ = json.Unmarshal([]byte(q.CorrectAnswersJSON), &answers)
		return answers
	}
	if strings.TrimSpace(q.CorrectAnswer) == "" {
		return []string{}
	}
	return []string{q.CorrectAnswer}
}

func normalizeAnswerList(values []string) []string {
	clean := make([]string, 0, len(values))
	seen := map[string]bool{}
	for _, value := range values {
		n := normalizeAnswer(value)
		if n == "" || seen[n] {
			continue
		}
		seen[n] = true
		clean = append(clean, n)
	}
	for i := 0; i < len(clean); i++ {
		for j := i + 1; j < len(clean); j++ {
			if clean[j] < clean[i] {
				clean[i], clean[j] = clean[j], clean[i]
			}
		}
	}
	return clean
}

func answersEqual(a, b []string) bool {
	na := normalizeAnswerList(a)
	nb := normalizeAnswerList(b)
	if len(na) != len(nb) {
		return false
	}
	for i := range na {
		if na[i] != nb[i] {
			return false
		}
	}
	return true
}

func valuesToStrings(v interface{}) []string {
	switch t := v.(type) {
	case string:
		return []string{t}
	case []string:
		return t
	case []interface{}:
		out := make([]string, 0, len(t))
		for _, x := range t {
			out = append(out, fmt.Sprint(x))
		}
		return out
	default:
		if v == nil {
			return []string{}
		}
		return []string{fmt.Sprint(v)}
	}
}
