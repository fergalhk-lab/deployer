package main

import (
	"flag"
	"fmt"
	"os"

	"sigs.k8s.io/yaml"

	"github.com/fergalhk-lab/deployer/config"
	"github.com/fergalhk-lab/deployer/generator"
)

const defaultIngressClass = "traefik"

func main() {
	outputFlag := flag.String("o", "-", "output file ('-' for stdout)")
	flag.Parse()

	if flag.NArg() != 1 {
		fmt.Fprintln(os.Stderr, "usage: deployer <config-file> [-o <output-file>]")
		os.Exit(1)
	}

	data, err := os.ReadFile(flag.Arg(0))
	if err != nil {
		fmt.Fprintf(os.Stderr, "error reading config: %v\n", err)
		os.Exit(1)
	}

	var cfg config.Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		fmt.Fprintf(os.Stderr, "error parsing config: %v\n", err)
		os.Exit(1)
	}

	result, err := generator.Generate(cfg, generator.Options{
		IngressClass: defaultIngressClass,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "error generating manifests: %v\n", err)
		os.Exit(1)
	}

	if *outputFlag == "-" {
		if _, err := os.Stdout.Write(result); err != nil {
			fmt.Fprintf(os.Stderr, "error writing output: %v\n", err)
			os.Exit(1)
		}
		return
	}

	if err := os.WriteFile(*outputFlag, result, 0644); err != nil {
		fmt.Fprintf(os.Stderr, "error writing output: %v\n", err)
		os.Exit(1)
	}
}
