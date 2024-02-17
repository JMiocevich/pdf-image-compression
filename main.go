package main

import (
	"bytes"
	"fmt"
	"image/jpeg"
	"io/ioutil"
	"log"
	"os"
	"regexp"

	pdf "github.com/pdfcpu/pdfcpu/pkg/api"

	"github.com/jung-kurt/gofpdf"
)

func main() {
	// filePath := "20231123144515302.pdf"
	filePath := "20231123111306565.pdf"

	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		fmt.Println("Error reading file:", err)
		return
	}

	re := regexp.MustCompile(`(?s)/Type\s*/XObject.*?/Subtype\s*/Image.*?stream(.*?)endstream`)
	matches := re.FindAllSubmatch(data, -1)

	// Initialize a new PDF document
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetCompression(true)

	var imageBuffers []bytes.Buffer

	for i, match := range matches {
		imageData := match[1]
		imageData = bytes.Trim(imageData, " \r\n\t\x0c")

		img, err := jpeg.Decode(bytes.NewReader(imageData))
		if err != nil {
			fmt.Printf("Failed to decode image %d: %v\n", i+1, err)
			continue
		}

		var buf bytes.Buffer
		err = jpeg.Encode(&buf, img, &jpeg.Options{Quality: 50})
		if err != nil {
			fmt.Printf("Failed to encode and compress image %d: %v\n", i+1, err)
			continue
		}

		imageBuffers = append(imageBuffers, buf)
	}

	for i, buf := range imageBuffers {

		pageWidth, pageHeight, rotate := 210.0, 297.0, "P"

		// fmt.Println(orientation)

		// if orientations[i] == 0 {
		// 	pageWidth, pageHeight, rotate = 297.0, 210.0, "L"
		// }

		pdf.AddPageFormat(rotate, gofpdf.SizeType{Wd: pageWidth, Ht: pageHeight})

		// Save the current transformation matrix
		// pdf.TransformBegin()

		// // Translate to the correct position
		// pdf.TransformTranslate(0, pageHeight)

		// // Rotate the coordinate system by 90 degrees clockwise
		// pdf.TransformRotate(90, 0, 0)

		opts := gofpdf.ImageOptions{
			ImageType:             "JPG",
			ReadDpi:               true,
			AllowNegativePosition: true,
		}
		pdf.RegisterImageOptionsReader(fmt.Sprintf("image%d.jpg", i+1), opts, &buf)
		pdf.ImageOptions(fmt.Sprintf("image%d.jpg", i+1), 0, 0, pageHeight, pageWidth, false, opts, 0, "")

		// Restore the previous transformation matrix
		// pdf.TransformEnd()
	}

	// Output the PDF to a file
	outputPath := "combined.pdf"
	err = pdf.OutputFileAndClose(outputPath)
	if err != nil {
		fmt.Printf("Failed to create PDF: %v\n", err)
		return
	}

	fmt.Println("PDF created successfully:", outputPath)

	inputSize, err := fileSize(filePath)
	if err != nil {
		fmt.Printf("Error getting input file size: %v\n", err)
		return
	}

	outputSize, err := fileSize(outputPath)
	if err != nil {
		fmt.Printf("Error getting output file size: %v\n", err)
		return
	}

	percentDiff := float64(outputSize-inputSize) / float64(inputSize) * 100

	fmt.Printf("Input PDF size: %d bytes\n", inputSize)
	fmt.Printf("Output PDF size: %d bytes\n", outputSize)
	fmt.Printf("Percentage difference in file size: %.2f%%\n", percentDiff)
}

func fileSize(filePath string) (int64, error) {
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return 0, err
	}
	return fileInfo.Size(), nil
}

func getOrientation(inputPath string) {

	// Open the input PDF file.
	f, err := os.Open(inputPath)
	if err != nil {
		log.Fatalf("Error opening PDF file: %v", err)
	}
	defer f.Close()

	// Create a new Context with the input PDF file.
	ctx, err := pdf.Open(f)
	if err != nil {
		log.Fatalf("Error opening PDF: %v", err)
	}

	// Get the number of pages in the PDF.
	pageCount := ctx.PageCount()

	// Iterate over each page in the PDF.
	for pageNum := 1; pageNum <= pageCount; pageNum++ {
		// Get the page size and orientation.
		pageSize, orientation, err := ctx.PageDims(pageNum)
		if err != nil {
			log.Fatalf("Error getting page dimensions for page %d: %v", pageNum, err)
		}

		// Print the page number, size, and orientation.
		fmt.Printf("Page %d: %.2f x %.2f - %s\n", pageNum, pageSize.Width, pageSize.Height, orientation)
	}

	// Close the context.
	ctx.Close()
}
