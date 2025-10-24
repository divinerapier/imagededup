# Image Deduplicate

A powerful Go library for finding and removing duplicate images using various algorithms. This library combines the efficiency of Go with the advanced image processing capabilities of Python's `imagededup` library.

## Features

- 🚀 **High Performance**: Built with Go for fast execution
- 🧠 **Multiple Algorithms**: Support for various image hashing algorithms
- 📦 **Self-Contained**: Python scripts embedded in Go binary
- 🐳 **Docker Ready**: Complete Docker support with pre-configured environment
- 🔧 **Easy Integration**: Simple API for Go applications
- 📊 **Detailed Results**: Comprehensive duplicate detection with similarity scores

## Supported Algorithms

- **PHasher**: Perceptual hashing (recommended for similar images)
- **CNN**: Convolutional Neural Network-based similarity
- **DHash**: Difference hashing
- **AHash**: Average hashing  
- **WHash**: Wavelet hashing

## Prerequisites

- Python 3.12+
- imagededup package: `pip install imagededup`

## Installation

```bash
go get github.com/divinerapier/imagededup

uv venv
source .venv/bin/activate 
uv pip install imagededup
```

## Quick Start

### Basic Usage

```go
package main

import (
    "fmt"
    "log"
    "github.com/divinerapier/imagededup"
)

func main() {
    // Find duplicates using perceptual hashing
    result, err := imagededup.FindDuplicates(
        imagededup.AlgorithmPHasher, 
        "/path/to/your/images"
    )
    if err != nil {
        log.Fatal(err)
    }
    
    // Print results
    for filename, duplicates := range result {
        fmt.Printf("File: %s has %d duplicates\n", filename, len(duplicates.DuplicateList))
        for _, dup := range duplicates.DuplicateList {
            fmt.Printf("  - %s (similarity: %.2f)\n", dup.Filename, dup.Score)
        }
    }
}
```

### Get Files to Remove

```go
// Get list of files that should be removed (keeping one from each duplicate group)
filesToRemove, err := imagededup.FindDuplicatesToRemove(
    imagededup.AlgorithmCNN, 
    "/path/to/your/images"
)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Found %d files that can be safely removed\n", len(filesToRemove))
```

## Docker Usage

### Build and Run

```bash
# Build the Docker image
docker build -t imagededup .

# Run with your images
docker run -v /path/to/your/images:/app/data imagededup phasher /app/data
```

## Algorithm Comparison

| Algorithm | Speed | Accuracy | Best For |
|-----------|-------|----------|----------|
| PHasher   | Fast  | High     | Similar images, rotated/edited photos |
| CNN       | Slow  | Highest  | Complex similarity detection |
| DHash     | Fast  | Medium   | Simple duplicate detection |
| AHash     | Fast  | Medium   | Basic similarity |
| WHash     | Fast  | Medium   | Wavelet-based similarity |

## Requirements

### Runtime Requirements
- Go 1.19+
- Python 3.12+ (for embedded scripts)
- imagededup Python package (automatically handled)

### Development Requirements
- Go 1.19+
- Docker (optional, for containerized deployment)

## Performance Tips

1. **Use PHasher for most cases**: Good balance of speed and accuracy
2. **Use CNN for critical applications**: Highest accuracy but slower
3. **Process in batches**: For large datasets, consider processing in chunks
4. **Use Docker**: Pre-configured environment with all dependencies

## Troubleshooting

### Common Issues

1. **Python not found**: Ensure Python 3.12+ is installed and in PATH
2. **imagededup package missing**: The library handles this automatically
3. **Permission errors**: Ensure the image directory is readable

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Submit a pull request

## License

MIT License - see LICENSE file for details
