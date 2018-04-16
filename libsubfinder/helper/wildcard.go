//
// wildcard.go : Wildcard elimination method for eliminating false subdomains
// Written By : @ice3man (Nizamul Rana)
//
// Distributed Under MIT License
// Copyrights (C) 2018 Ice3man
//

package helper

import (
	"fmt"
	"net"
	"strings"
	//"github.com/miekg/dns"
)

// Method to eliminate Wildcard Is based on OJ Reeves Work on Gobuster Project
// github.com/oj/gobuster :-)
func InitializeWildcardDNS(state *State) bool {
	// Generate a random UUID and check if server responds with a valid
	// IP Address. If so, we are dealing with a wildcard DNS Server and will have
	// to work accordingly.
	// In case of wildcard DNS, we will ignore any subdomain which has same IP
	// as our random UUID one
	uuid, _ := NewUUID()

	// Gets a list of IP's by resolving a non-existent host
	wildcardIPs, err := net.LookupHost(fmt.Sprintf("%s.%s", uuid, state.Domain))

	if err == nil {
		state.IsWildcard = true
		state.WildcardIPs.AddRange(wildcardIPs)

		// We have found a wildcard DNS Server
		return true
	}

	return false
}

// Checks if a given subdomain is a wildcard subdomain
// It takes Current application state, Domain to find subdomains for
func CheckWildcardSubdomain(state *State, domain string, words chan string, donech chan struct{}, result chan string) {
	for target := range words {
		preparedSubdomain := target + "." + domain
		ipAddress, err := net.LookupHost(preparedSubdomain)

		if err == nil {
			// No eror, let's see if it's a Wildcard subdomain
			if !state.WildcardIPs.ContainsAny(ipAddress) {
				result <- preparedSubdomain
				donech <- struct{}{}
			} else {
				// This is likely a wildcard entry, skip it
				result <- ""
				donech <- struct{}{}
			}
		} else {
			result <- ""
			donech <- struct{}{}
		}
	}
}

// Removes bad wildcard subdomains from list of subdomains.
func RemoveWildcardSubdomains(state *State, subdomains []string) []string {
	wildcard := InitializeWildcardDNS(state)
	if wildcard == true {
		fmt.Printf("\n\n%s[!]%s Wildcard DNS Detected ! False Positives are likely :-(\n\n", Cyan, Reset)
	}

	var validSubdomains []string

	var channel = make(chan string)
	var donech = make(chan struct{})
	var result = make(chan string)

	for i := 0; i < state.Threads; i++ {
		go CheckWildcardSubdomain(state, state.Domain, channel, donech, result)
	}

	for _, entry := range subdomains {
		sub := strings.Join(strings.Split(entry, ".")[:2][:], ".")
		fmt.Printf("\n[!] %s", sub+"."+state.Domain)
		channel <- sub
	}

	for _, _ = range subdomains {
		result := <-channel
		if state.Verbose == true {
			fmt.Printf("\n[-] %s", result)
		}
		if result != "" {
			validSubdomains = append(validSubdomains, result)
		}
	}

	close(channel)

	for i := 0; i < state.Threads; i++ {
		<-donech
	}

	return validSubdomains
}
