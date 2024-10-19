package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path"
)

func main() {
	pdfCmd := flag.NewFlagSet("pdf", flag.ExitOnError)
	pdfName := pdfCmd.String("name", "", "Name to be saved")
	pdfPath := pdfCmd.String("path", "", "PDF file path")

	if len(os.Args) < 2 {
		fmt.Println("Usage: AllPDF <subcommand> [flags]")
		fmt.Println("Subcommands:")
		fmt.Println("  pdf")
		fmt.Println("Flags:")
		fmt.Println("  -path string")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "pdf":
		pdfCmd.Parse(os.Args[2:])
		fmt.Println("It is recommended to remove any pages of the PDF that you don't want to be Narrated")
		fmt.Println("cover/illustrations/tables/graphs/charts")
		fmt.Println("Is the path correct? [y/n]")
		fmt.Println(*pdfPath)
		var option string
		fmt.Scanln(&option)
		if option == "y" {
			start(*pdfPath, *pdfName)
		} else {
      fmt.Println("Aborted")
      os.Exit(1)
    }
	default:
		fmt.Println("subcommand is required (pdf)")
		os.Exit(1)
	}
}

func start(src, fileName string){
  cpFile(src, fileName)
}

func cpFile(src, fileName string) (dst string) {
	err := os.RemoveAll("./tmp")
	if err != nil {
		log.Fatalf("Error removing tmp folder: %v", err)
	}

	file, err := os.ReadFile(src)
	if err != nil {
		log.Fatalf("Error while reading file: %v", err)
	}

	err = os.MkdirAll(path.Join("./tmp", fileName), os.ModePerm)
	if err != nil {
		log.Fatalf("Error while creating directories: %v", err)
	}

	err = os.WriteFile(fmt.Sprintf("./tmp/%s/%s.pdf", fileName, fileName), file, os.ModePerm)
	if err != nil {
		log.Fatalf("Error while saving file: %v", err)
	}

	return fmt.Sprintf("./tmp/%s", fileName)
}
