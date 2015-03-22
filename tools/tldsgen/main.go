/* Copyright (c) 2015, Daniel Martí <mvdan@mvdan.cc> */
/* See LICENSE for licensing information */

package main

import (
	"bufio"
	"log"
	"net/http"
	"os"
	"regexp"
	"sort"
	"strings"
	"text/template"
)

var tldsTmpl = template.Must(template.New("tlds").Parse(`// Generated by tldsgen

package xurls

// TLDs is a sorted list of all public top-level domains
var TLDs = []string{
{{range $_, $value := .TLDs}}` + "\t`" + `{{$value}}` + "`" + `,
{{end}}}
`))

func addFromIana(addTld func(tld string)) error {
	resp, err := http.Get("https://data.iana.org/TLD/tlds-alpha-by-domain.txt")
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	scanner := bufio.NewScanner(resp.Body)
	re := regexp.MustCompile(`^[^#]+$`)
	for scanner.Scan() {
		line := scanner.Text()
		tld := re.FindString(line)
		addTld(tld)
	}
	return nil
}

func addFromPublicSuffix(addTld func(tld string)) error {
	resp, err := http.Get("https://publicsuffix.org/list/effective_tld_names.dat")
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	scanner := bufio.NewScanner(resp.Body)
	re := regexp.MustCompile(`^[^/.]+$`)
	for scanner.Scan() {
		line := scanner.Text()
		tld := re.FindString(line)
		addTld(tld)
	}
	return nil
}

func tldList() ([]string, error) {
	tlds := make(map[string]struct{})
	addTld := func(tld string) {
		if tld == "" {
			return
		}
		tld = strings.ToLower(tld)
		if strings.HasPrefix(tld, "xn--") {
			return
		}
		tlds[tld] = struct{}{}
	}
	if err := addFromIana(addTld); err != nil {
		return nil, err
	}
	if err := addFromPublicSuffix(addTld); err != nil {
		return nil, err
	}
	list := make([]string, 0, len(tlds))
	for tld := range tlds {
		list = append(list, tld)
	}
	sort.Strings(list)
	return list, nil
}

func writeTlds(tlds []string) error {
	f, err := os.Create("tlds.go")
	if err != nil {
		return err
	}
	return tldsTmpl.Execute(f, struct {
		TLDs []string
	}{
		TLDs: tlds,
	})
}

func main() {
	tlds, err := tldList()
	if err != nil {
		log.Fatalf("Could not get TLD list: %s", err)
	}
	if err := writeTlds(tlds); err != nil {
		log.Fatalf("Could not write tlds.go: %s", err)
	}
}