package main

import "testing"

var testCases = []struct {
	login    string
	expected bool
}{
	{"login1", true},
	{"", false},
	{"log; rm -rf", false},
	{"123456", true},
}

func TestLogins(t *testing.T) {
	for _, test := range testCases {
		observed := Validate(test.login)
		if observed != test.expected {
			t.Errorf("For p = %s, expected %t. Got %t.",
				test.login, test.expected, observed)
		}
	}
}
