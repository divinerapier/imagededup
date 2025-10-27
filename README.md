# Image Deduplicate

A powerful Go library for finding and removing duplicate images using various algorithms. This library embeds Python and the `imagededup` library directly into your Go binary, providing a self-contained solution with no external dependencies required at runtime.

## Features

- 🚀 **High Performance**: Built with Go, with embedded Python runtime
- 🧠 **Multiple Algorithms**: Support for various image hashing algorithms (PHasher, CNN, DHash, AHash, WHash)
- 📦 **Self-Contained**: Complete Python environment and `imagededup` library embedded in Go binary
- 🔧 **Thread Pool**: Configurable parallelism for concurrent operations
- 📊 **Detailed Results**: Comprehensive duplicate detection with similarity scores

## Supported Algorithms

- **PHasher**: Perceptual hashing (recommended for similar images)
- **CNN**: Convolutional Neural Network-based similarity
- **DHash**: Difference hashing
- **AHash**: Average hashing  
- **WHash**: Wavelet hashing

## Prerequisites

- Go 1.19+
- No Python installation required at runtime (embedded)

## Installation

```bash
go get github.com/divinerapier/imagededup
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
    // Create an ImageDedup instance with parallelism level
    // Parallelism controls the number of concurrent HTTP client threads
    dedup, err := imagededup.NewImageDedup(4)
    if err != nil {
        log.Fatalf("failed to create ImageDedup: %v", err)
    }
    defer dedup.Close()
    defer dedup.Cleanup()

    // Find duplicates using CNN algorithm
    results, err := dedup.FindDuplicates(
        imagededup.AlgorithmCNN, 
        "/path/to/your/images"
    )
    if err != nil {
        log.Fatalf("failed to find duplicates: %v", err)
    }
    
    // Print results
    for _, result := range results {
        fmt.Printf("File: %s has %d duplicates\n", result.Filename, len(result.DuplicateList))
        for _, dup := range result.DuplicateList {
            fmt.Printf("  - %s (similarity: %.2f)\n", dup.Filename, dup.Score)
        }
    }
}
```

### Get Files to Remove

```go
// Get list of files that should be removed (keeping one from each duplicate group)
filesToRemove, err := dedup.FindDuplicatesToRemove(
    imagededup.AlgorithmPHasher, 
    "/path/to/your/images"
)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Found %d files that can be safely removed\n", len(filesToRemove))
for _, file := range filesToRemove {
    fmt.Println(file)
}
```

## Docker Usage

### Build and Run

```bash
# Build the Docker image
docker build -t imagededup .

# Run with your images mounted
docker run -v /path/to/your/images:/app/data imagededup
```

Note: The Docker image includes the full Python environment and all dependencies pre-installed.

## Algorithm Comparison

| Algorithm | Speed | Accuracy | Best For |
|-----------|-------|----------|----------|
| PHasher   | Fast  | High     | Similar images, rotated/edited photos |
| CNN       | Slow  | Highest  | Complex similarity detection |
| DHash     | Fast  | Medium   | Simple duplicate detection |
| AHash     | Fast  | Medium   | Basic similarity |
| WHash     | Fast  | Medium   | Wavelet-based similarity |

## Architecture

This library uses a unique architecture:

1. **Embedded Python**: The complete Python runtime and `imagededup` library are embedded in the Go binary
2. **HTTP Communication**: Go communicates with Python via an internal HTTP server
3. **Thread Pool**: Configurable concurrent HTTP clients for parallel operations
4. **Resource Management**: Automatic cleanup of temporary files and processes

### Requirements

### Runtime
- Go 1.19+
- No external dependencies

### Development (for building embedded binaries)
- Go 1.19+
- Docker (for building Docker images)

## Performance Tips

1. **Adjust parallelism**: Use `NewImageDedup(n)` where `n` is the number of concurrent operations
2. **Use PHasher for most cases**: Good balance of speed and accuracy
3. **Use CNN for critical applications**: Highest accuracy but slower
4. **Memory management**: Call `Cleanup()` and `Close()` to free resources

## Troubleshooting

### Common Issues

1. **Port already in use**: The library uses port 18000 by default for internal HTTP server
2. **Permission errors**: Ensure the image directory is readable
3. **Memory issues**: Reduce parallelism level if running out of memory
4. **Initialization timeout**: Server health check waits up to 5 minutes for startup

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Submit a pull request

## License

MIT License - see LICENSE file for details
