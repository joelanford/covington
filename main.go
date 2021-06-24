package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"k8s.io/apimachinery/pkg/util/sets"
	"sigs.k8s.io/yaml"
)

type Channel struct {
	Schema     string   `json:"schema"`
	Package    string   `json:"package"`
	Name       string   `json:"name"`
	Versions   []string `json:"versions"`
	Tombstones []string `json:"tombstones"`
}

func main() {
	f, err := os.ReadFile(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	var ch Channel
	if err := yaml.Unmarshal(f, &ch); err != nil {
		log.Fatal(err)
	}

	if len(ch.Versions) == 0 {
		log.Fatal("at least one version is required")
	}
	head := ch.Versions[len(ch.Versions)-1]

	tombstones := sets.NewString(ch.Tombstones...)
	if tombstones.Has(ch.Versions[len(ch.Versions)-1]) {
		log.Fatalf("head node %q must not be tombstone", head)
	}

	nonTombstones := []string{}
	for _, v := range ch.Versions {
		if !tombstones.Has(v) {
			nonTombstones = append(nonTombstones, v)
		}
	}

	var graph strings.Builder
	graph.WriteString("graph RL\n")
	for _, nt := range nonTombstones {
		graph.WriteString(fmt.Sprintf("  %s\n", nt))
	}
	if len(nonTombstones) > 1 {
		graph.WriteString("\n")
		for i := 1; i < len(nonTombstones); i++ {
			graph.WriteString(fmt.Sprintf("  %s == replaces ==> %s\n", nonTombstones[i], nonTombstones[i-1]))
		}
	}
	graph.WriteString("\n")

	from, to := 0, 1
	for to < len(ch.Versions) {
		for {
			if tombstones.Has(ch.Versions[to]) {
				to += 1
			} else {
				break
			}
		}
		for from < to {
			if tombstones.Has(ch.Versions[from]) {
				graph.WriteString(fmt.Sprintf("  %s -- skips --> %s\n", ch.Versions[to], ch.Versions[from]))
			}
			from += 1
		}
		from += 1
		to += 2
	}

	graph.WriteString("\n")
	for _, v := range ch.Versions {
		if tombstones.Has(v) {
			graph.WriteString(fmt.Sprintf("  style %s fill:#ccc,stroke:#666,stroke-width:1px,color:#666,stroke-dasharray: 4\n", v))
		}
	}

	fmt.Println(strings.TrimSpace(graph.String()))
}
