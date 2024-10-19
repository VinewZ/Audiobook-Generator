package main

import (
	"flag"
	"fmt"
	"image/jpeg"
	"log"
	"os"
	"path"
	"path/filepath"
	"time"

	"github.com/gen2brain/go-fitz"
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
  convertToImgs(path.Join("./tmp", fileName), fileName)
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


func convertToImgs(src, fileName string){
  fmt.Println("Starting converting to images")
  startTime := time.Now()
  pdf, err := fitz.New(fmt.Sprintf("%s/%s.pdf", src, fileName))
  if err != nil {
    log.Fatalf("Error while opening file with Fitz: %v", err)
  }
  defer pdf.Close()

  imgsDirPath := path.Join(src, "imgs")
  err = os.Mkdir(imgsDirPath, os.ModePerm)

  for i := 0; i < pdf.NumPage() / 15; i++ {

    fmt.Printf("Converting page %d / %d\n", i, pdf.NumPage() / 10)
    img, err := pdf.Image(i)
    if err != nil {
      log.Fatalf("Error while converting page Number %d, to image: %v", i, err)
    }

    f, err := os.Create(filepath.Join(imgsDirPath, fmt.Sprintf("%s-%04d.jpg", fileName, i)))
    if err != nil {
      log.Fatalf("Error while saving image Number %d: %v", i, err)
    }
    defer f.Close()

    err = jpeg.Encode(f, img, &jpeg.Options{Quality: jpeg.DefaultQuality})
    if err != nil {
      log.Fatalf("Error while encoding JPEG Number %d: %v", i, err)
    }

  }

  durationTime := time.Since(startTime)
  fmt.Println("Finished converting to images")
  fmt.Printf("Took %.2f seconds\n", durationTime.Seconds())
}
