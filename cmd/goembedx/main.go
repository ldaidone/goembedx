package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/ldaidone/goembedx"
)

var (
	dim     = flag.Int("dim", 3, "Dimension of vectors")
	topK    = flag.Int("k", 3, "Top-K results")
	mode    = flag.String("mode", "interactive", "Mode: add|query|interactive")
	id      = flag.String("id", "", "ID for add mode")
	vecFlag = flag.String("vec", "", "Comma-separated vector")
)

func parseVec(s string, dimension int) ([]float32, error) {
	parts := strings.Split(s, ",")
	if len(parts) != dimension {
		return nil, fmt.Errorf("expected %d elements, got %d", dimension, len(parts))
	}
	v := make([]float32, len(parts))
	for i, p := range parts {
		f, err := strconv.ParseFloat(strings.TrimSpace(p), 32)
		if err != nil {
			return nil, err
		}
		v[i] = float32(f)
	}
	return v, nil
}

func main() {
	//helpFlag := flag.String("h", "", "Show help")
	flag.Parse()

	//if helpFlag != nil {
	//	fmt.Println("usage...")
	//	return
	//}
	store := goembedx.MemoryStore(*dim)

	switch *mode {
	case "add":
		v, err := parseVec(*vecFlag, *dim)
		if err != nil {
			fmt.Println("parse error:", err)
			os.Exit(1)
		}
		if err := goembedx.AddVector(store, *id, v); err != nil {
			fmt.Println("add failed:", err)
			os.Exit(1)
		}
		fmt.Println("✅ Added:", *id)
		return

	case "query":
		v, err := parseVec(*vecFlag, *dim)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		results, err := goembedx.SearchTopK(store, v, *topK)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		fmt.Println("🔎 Results:")
		for _, r := range results {
			fmt.Printf(" %s => %.5f\n", r.ID, r.Score)
		}
		return
	}

	// interactive REPL mode (phase-1 lightweight)
	fmt.Println("goembedx interactive mode")
	fmt.Println("commands:")
	fmt.Println(" add <id> 1,2,3")
	fmt.Println(" query 1,2,3")
	reader := bufio.NewScanner(os.Stdin)

	for {
		fmt.Print("> ")
		if !reader.Scan() {
			break
		}
		line := strings.TrimSpace(reader.Text())

		if strings.HasPrefix(line, "add ") {
			fields := strings.Fields(line)
			id := fields[1]
			v, err := parseVec(fields[2], *dim)
			if err != nil {
				fmt.Println("⚠️ parse:", err)
				continue
			}
			_ = goembedx.AddVector(store, id, v)
			fmt.Println("✅ ok")
			continue
		}

		if strings.HasPrefix(line, "query ") {
			fields := strings.Fields(line)
			v, err := parseVec(fields[1], *dim)
			if err != nil {
				fmt.Println("⚠️ parse:", err)
				continue
			}
			results, _ := goembedx.SearchTopK(store, v, *topK)
			for _, r := range results {
				fmt.Printf(" %s => %.5f\n", r.ID, r.Score)
			}
			continue
		}
	}
}
