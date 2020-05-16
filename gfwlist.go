package hsocks5

import (
	"bytes"
	"encoding/base64"

	"github.com/pmezard/adblock/adblock"
)

// LoadGFWList file
func LoadGFWList() *adblock.RuleMatcher {
	bs, _ := Asset("assets/gfwlist.txt")
	gfwlistBase64Text := string(bs[:])
	gfwlistText, _ := base64.StdEncoding.DecodeString(gfwlistBase64Text)
	rules, _ := adblock.ParseRules(bytes.NewReader(gfwlistText))
	m := adblock.NewMatcher()

	for i, r := range rules {
		m.AddRule(r, i)
	}

	return m
}
