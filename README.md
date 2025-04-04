# PNG to ICO Converter

A command-line tool written in Go that converts PNG images to ICO format with support for multiple icon sizes and transparency.

## Features

- Convert PNG images to ICO format
- Support for multiple icon sizes in a single ICO file
- Preserves transparency/alpha channel
- High-quality image resizing using Lanczos3 algorithm
- Customizable icon sizes via command-line arguments

## Requirements

- Go 1.21 or later
- Required Go packages:
  - github.com/nfnt/resize

## Installation

1. Clone the repository:

```bash
git clone https://github.com/yourusername/png-to-ico.git
cd png-to-ico
```

2. Install dependencies:

```bash
go mod tidy
```

## Usage

Basic usage:

```bash
go run main.go input.png output.ico
```

Specify custom icon sizes:

```bash
go run main.go -size=32,64,128 input.png output.ico
```

### Command Line Options

- `-size`: Comma-separated list of icon sizes (default: "16,32,64,128,256")
  Example: `-size=32,64,128`

### Output Files

The program generates:

1. An ICO file containing all specified icon sizes
2. Individual PNG files for each size, named as `output_32x32.png`, `output_64x64.png`, etc.

## Examples

1. Convert with default sizes (16,32,64,128,256):

```bash
go run main.go logo.png favicon.ico
```

2. Convert with specific sizes:

```bash
go run main.go -size=32,64,128 logo.png favicon.ico
```

3. Convert with a single size:

```bash
go run main.go -size=256 logo.png favicon.ico
```

## Notes

- Input file must be a PNG image
- Output file must have .ico extension
- The program preserves transparency/alpha channel from the input PNG
- Larger input images will be automatically resized to the specified dimensions
- The Lanczos3 resampling algorithm is used for high-quality resizing

## Technical Details

The converter:

1. Reads the input PNG image
2. Resizes it to each specified size using Lanczos3 resampling
3. Converts each resized image to RGBA format to preserve transparency
4. Creates BMP data with proper alpha channel handling
5. Writes the ICO file with all specified sizes
6. Generates individual PNG files for each size

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.
