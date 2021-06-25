package main

import (
	"fmt"
	"testing"
)

func TestHasNoDirectEqualityTag(t *testing.T) {
	tests := map[string]struct {
		tag      string
		expected bool
	}{

		"Good Token 1":                               {fmt.Sprintf(`db:"acceptance_token" %s:"%s"`, securityTag, noDirectEqualityValue), true},
		"Good Token 2":                               {fmt.Sprintf(`db:"verification_token" %s:"%s"`, securityTag, noDirectEqualityValue), true},
		"Good Token 3":                               {fmt.Sprintf(`json:"token,omitempty" %s:"%s"`, securityTag, noDirectEqualityValue), true},
		"Extra token":                                {fmt.Sprintf(`json:"token,omitempty" %s:"%s,other"`, securityTag, noDirectEqualityValue), true},
		"Extra token swapped":                        {fmt.Sprintf(`json:"token,omitempty" %s:"other,%s"`, securityTag, noDirectEqualityValue), true},
		"Only a Security Tag":                        {fmt.Sprintf(`%s:"%s"`, securityTag, noDirectEqualityValue), true},
		"Only Security tag with extra token":         {fmt.Sprintf(`%s:"other,%s"`, securityTag, noDirectEqualityValue), true},
		"Only Security tag with extra token swapped": {fmt.Sprintf(`%s:"%s,other"`, securityTag, noDirectEqualityValue), true},
		"Security tag first in list of tags":         {fmt.Sprintf(`%s:"%s" json:"token"`, securityTag, noDirectEqualityValue), true},
		"Tag in wrong place":                         {fmt.Sprintf(`db:"%s" %s:"acceptance_token"`, noDirectEqualityValue, securityTag), false},
		"Empty Security Tag":                         {fmt.Sprintf(`%s:""`, securityTag), false},
		"Empty String":                               {`""`, false},
	}
	for testName, testCase := range tests {
		t.Run(testName, func(t *testing.T) {
			actual := hasNoDirectEqualityTag(testCase.tag)
			if testCase.expected != actual {
				t.Errorf("\nTest: %v\tExpected: %v\tActual: %v\t tag: %v", testName, testCase.expected, actual, testCase.tag)
			}
		})
	}
}
