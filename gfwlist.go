package hsocks5

import (
	"bytes"
	"encoding/base64"
	"io/ioutil"

	"github.com/markbates/pkger"
	"github.com/pmezard/adblock/adblock"
)

// LoadGFWList file
func LoadGFWList() *adblock.RuleMatcher {
	f, _ := pkger.Open("/assets/gfwlist.txt")
	bs, _ := ioutil.ReadAll(f)
	gfwlistBase64Text := string(bs[:])
	gfwlistText, _ := base64.StdEncoding.DecodeString(gfwlistBase64Text)
	rules, _ := adblock.ParseRules(bytes.NewReader(gfwlistText))
	m := adblock.NewMatcher()

	for i, r := range rules {
		m.AddRule(r, i)
	}

	return m
}
