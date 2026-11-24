package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/Ishitapathunda/rag-intelligence-engine/internal/config"
)

func main() {
	cfg := config.Load()

	rawDir := flag.String("raw", cfg.RawDir, "directory containing source PDF/TXT/MD files")
	outDir := flag.String("out", cfg.ProcessedDir, "directory for extracted text files")
	flag.Parse()

	if err := os.MkdirAll(*outDir, 0755); err != nil {
		log.Fatal(err)
	}

	entries, err := os.ReadDir(*rawDir)
	if err != nil {
		log.Fatal(err)
	}

	for _, e := range entries {
		if e.IsDir() {
			continue
		}

		name := e.Name()
		ext := strings.ToLower(filepath.Ext(name))
		src := filepath.Join(*rawDir, name)
		base := strings.TrimSuffix(name, filepath.Ext(name))
		dst := filepath.Join(*outDir, base+".txt")

		switch ext {
		case ".pdf":
			cmd := exec.Command("pdftotext", "-layout", src, dst)
			if err := cmd.Run(); err != nil {
				log.Printf("skipping %s: pdftotext failed: %v", name, err)
				continue
			}
			fmt.Printf("%s -> %s\n", name, dst)

		case ".txt", ".md":
			if err := copyFile(src, dst); err != nil {
				log.Printf("skipping %s: %v", name, err)
				continue
			}
			fmt.Printf("%s -> %s\n", name, dst)

		default:
			fmt.Printf("%s skipped (unsupported)\n", name)
		}
	}
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}
