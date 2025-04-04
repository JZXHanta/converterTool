package main

import (
	"encoding/binary"
	"flag"
	"fmt"
	"image"
	"image/png"
	"os"
	"strconv"
	"strings"

	"github.com/nfnt/resize"
)

func main() {
	// Define command line flags
	sizeStr := flag.String("size", "16,32,64,128,256", "Comma-separated list of icon sizes")
	flag.Parse()

	// Parse the size argument
	sizes := make([]int, 0)
	for _, s := range strings.Split(*sizeStr, ",") {
		size, err := strconv.Atoi(strings.TrimSpace(s))
		if err != nil {
			fmt.Printf("Error: Invalid size '%s'. Sizes must be integers.\n", s)
			os.Exit(1)
		}
		sizes = append(sizes, size)
	}

	// Get remaining arguments (input and output files)
	args := flag.Args()
	if len(args) != 2 {
		fmt.Println("Usage: go run main.go [options] input.png output.ico")
		fmt.Println("Options:")
		flag.PrintDefaults()
		os.Exit(1)
	}

	inputPath := args[0]
	inputPathParts := strings.Split(inputPath, ".")
	outputPath := args[1]
	outputPathParts := strings.Split(outputPath, ".")

	if len(inputPathParts) == 3 && inputPathParts[2] != "png" || len(inputPathParts) == 2 && inputPathParts[1] != "png" {
		fmt.Println("Error: First argument must be a '.png' file")
		os.Exit(1)
	}

	if len(outputPathParts) == 3 && outputPathParts[2] != "ico" || len(outputPathParts) == 2 && outputPathParts[1] != "ico" {
		outputPath = outputPath + ".ico"
	}

	// Read the PNG file
	img, err := readPNG(inputPath)
	if err != nil {
		fmt.Println("Error reading PNG:", err)
		os.Exit(1)
	}

	// Print input image information
	bounds := img.Bounds()
	fmt.Printf("Input image size: %dx%d\n", bounds.Dx(), bounds.Dy())

	// Create the ICO file
	outputFile, err := os.Create(outputPath)
	if err != nil {
		fmt.Println("Error creating output file:", err)
		os.Exit(1)
	}
	defer outputFile.Close()

	// Create icons at specified sizes
	icons := make([]image.Image, len(sizes))
	for i, size := range sizes {
		// Use high-quality Lanczos3 resampling
		resized := resize.Resize(uint(size), uint(size), img, resize.Lanczos3)

		// Convert to RGBA to ensure alpha channel is preserved
		rgba := image.NewRGBA(resized.Bounds())
		for y := resized.Bounds().Min.Y; y < resized.Bounds().Max.Y; y++ {
			for x := resized.Bounds().Min.X; x < resized.Bounds().Max.X; x++ {
				rgba.Set(x, y, resized.At(x, y))
			}
		}
		icons[i] = rgba

		fmt.Printf("Created icon size: %dx%d\n", size, size)
	}

	// Write the ICO file
	err = writeICO(outputFile, icons, sizes)
	if err != nil {
		fmt.Println("Error writing ICO file:", err)
		os.Exit(1)
	}

	fmt.Printf("Successfully converted %s to %s with sizes: %v\n", inputPath, outputPath, sizes)
}

func readPNG(inputPath string) (image.Image, error) {
	inputFile, err := os.Open(inputPath)
	if err != nil {
		return nil, err
	}
	defer inputFile.Close()

	img, err := png.Decode(inputFile)
	if err != nil {
		return nil, err
	}

	return img, nil
}

func writeICO(writer *os.File, images []image.Image, sizes []int) error {
	// ICO header
	header := []byte{0, 0, 1, 0, byte(len(images)), 0}
	_, err := writer.Write(header)
	if err != nil {
		return err
	}

	// Calculate the offset where the image data starts
	offset := 6 + len(images)*16 // 6 bytes for header + 16 bytes per image entry

	// First pass: write directory entries and collect BMP data
	bmpDataList := make([][]byte, len(images))
	for i, img := range images {
		// Create BMP data with alpha channel
		bmpData, err := createBMPWithAlpha(img)
		if err != nil {
			return err
		}
		bmpDataList[i] = bmpData

		// Write directory entry
		bounds := img.Bounds()
		width := bounds.Dx()
		height := bounds.Dy()

		entry := make([]byte, 16)
		if width >= 256 {
			entry[0] = 0
		} else {
			entry[0] = byte(width)
		}
		if height >= 256 {
			entry[1] = 0
		} else {
			entry[1] = byte(height)
		}
		entry[2] = 0 // color palette (0 for no palette)
		entry[3] = 0 // reserved
		entry[4] = 1 // color planes
		entry[5] = 0
		entry[6] = 32 // bits per pixel
		entry[7] = 0
		entry[8] = byte(len(bmpData) & 0xFF)
		entry[9] = byte((len(bmpData) >> 8) & 0xFF)
		entry[10] = byte((len(bmpData) >> 16) & 0xFF)
		entry[11] = byte((len(bmpData) >> 24) & 0xFF)
		entry[12] = byte(offset & 0xFF)
		entry[13] = byte((offset >> 8) & 0xFF)
		entry[14] = byte((offset >> 16) & 0xFF)
		entry[15] = byte((offset >> 24) & 0xFF)

		_, err = writer.Write(entry)
		if err != nil {
			return err
		}

		offset += len(bmpData)
	}

	// Second pass: write all BMP data
	for i, bmpData := range bmpDataList {
		println("Writing BMP data for size", sizes[i], "length:", len(bmpData))
		_, err := writer.Write(bmpData)
		if err != nil {
			return err
		}
	}

	return nil
}

func createBMPWithAlpha(img image.Image) ([]byte, error) {
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// BMP header (40 bytes)
	header := make([]byte, 40)
	binary.LittleEndian.PutUint32(header[0:4], 40) // Header size
	binary.LittleEndian.PutUint32(header[4:8], uint32(width))
	binary.LittleEndian.PutUint32(header[8:12], uint32(height*2))        // Height * 2 for top-down
	binary.LittleEndian.PutUint16(header[12:14], 1)                      // Planes
	binary.LittleEndian.PutUint16(header[14:16], 32)                     // Bits per pixel
	binary.LittleEndian.PutUint32(header[16:20], 0)                      // Compression (BI_RGB)
	binary.LittleEndian.PutUint32(header[20:24], uint32(width*height*4)) // Image size
	binary.LittleEndian.PutUint32(header[24:28], 0)                      // X pixels per meter
	binary.LittleEndian.PutUint32(header[28:32], 0)                      // Y pixels per meter
	binary.LittleEndian.PutUint32(header[32:36], 0)                      // Colors used
	binary.LittleEndian.PutUint32(header[36:40], 0)                      // Important colors

	// Pixel data (BGRA format, top-down)
	pixels := make([]byte, width*height*4)
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			r, g, b, a := img.At(x, y).RGBA()
			offset := (y*width + x) * 4
			pixels[offset] = byte(b >> 8)   // Blue
			pixels[offset+1] = byte(g >> 8) // Green
			pixels[offset+2] = byte(r >> 8) // Red
			pixels[offset+3] = byte(a >> 8) // Alpha
		}
	}

	// Combine header and pixel data
	bmpData := make([]byte, 40+len(pixels))
	copy(bmpData[0:40], header)
	copy(bmpData[40:], pixels)

	return bmpData, nil
}
