package main

import (
	"fmt"
	"github.com/uk0/Octopoda/import/github.com/astaxie/beego/logs"
	"regexp"
)

type String string


type Rule struct {
	regexp      *regexp.Regexp
	subNames []string
}

func (* String)LineRegx(sregexp string)(*Rule,error){
	rule := &Rule{}
	var err error
	if rule.regexp, err = regexp.Compile(sregexp); err != nil {
		return rule, err
	}
	rule.subNames = rule.regexp.SubexpNames()
	return rule, nil
}

func (rule *Rule) Match(line string) interface{} {
	matches := rule.regexp.FindStringSubmatch(line)
	if len(matches) == 0 {
		return nil
	}

	if len(matches) <= 1 {
		logs.Info("[matches] True")
		return true
	}
	// TODO: cache subexnames
	for i, value := range matches[1:] {
		fmt.Println(value)
		fmt.Println(i)
	}
	return nil
}

func test_in() {
	var b String = ""
	r,_:=b.LineRegx(`(^H.*\s)`)
	r.Match("Hello 世界！123 ")
	r.Match("Hell111o 世界！123")
}