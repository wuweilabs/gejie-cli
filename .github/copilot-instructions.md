# AI Assistant Instructions for gejie-cli

This document helps AI assistants understand the key patterns and workflows in the gejie-cli project.

## Project Overview

gejie-cli is a data analytics tool for scraping LATAM eCommerce platforms, primarily focusing on MercadoLibre, with planned support for Coolbox.pe and Falabella. The tool helps manufacturing and direct commerce partners with product and market research.

## Key Components

### Core Components
- `cli/` - Command line interface implementation using Cobra
  - `melic.go` - Main CLI command for MercadoLibre scraping
  - `root.go` - Root CLI command and configuration

- `gejielib/` - Core scraping and data processing functionality
  - `melisearch.go` - MercadoLibre scraping implementation
  - `meli/*.go` - MercadoLibre specific data structures and helpers
  - `browser/` - Browser automation configuration and helpers
  - `adapters/` - Platform-specific adapters for different ecommerce sites

### Data Models
Core data structures in `gejielib/meli/structs.go`:
```go
type MeliProduct struct {
    Title              string
    Price              Price
    Url                string
    ReviewCount        *uint32
    Rating             *float32
    ImageUrls          []string
    SoldMoreThan       *uint32
    EstimatedSoldCount *uint32
    DescriptionContent string
    StoreInfo          MeliStoreInfo
}
```

## Development Workflows

### Build and Run
```bash
# Build the CLI
go build -o gejiec .

# Run basic commands
./gejiec --help
./gejiec meli --help
```

### Common CLI Options
- `--url` (string, required): Target URL for scraping
- `--max-items` (int, default 10): Maximum items to scrape
- `--only-images` (bool): Only scrape image URLs
- `--create-csv` (bool): Export results to CSV
- `--headless` (bool): Run browser in headless mode
- `--workers` (int): Number of concurrent scrapers

## Key Patterns

### Browser Automation
- Uses Playwright for reliable web scraping
- Browser options configured in `gejielib/browser/config.go`
- Default to non-headless mode for development
- Resource blocking patterns to optimize scraping speed

### Data Processing
- CSV export capabilities for bulk data analysis
- Structured data models for each platform
- Concurrent scraping with worker pools
- Platform-specific adapters for different sites

### Error Handling
- Extensive error logging using slog
- Graceful degradation for missing fields
- Browser session cleanup in defer blocks
- Retry mechanisms for flaky network requests

## Project Structure Conventions
- Platform-specific code goes in dedicated subdirectories (e.g., `meli/`, `coolboxpe/`)
- Shared utilities in `utils/` package
- Browser automation config in `browser/` package
- Analytics tools and notebooks in `analytics/` directory

## Debugging Tips
- Set `devMode = 1` in `main.go` to enable development testing
- Use `--headless=false` to watch browser automation
- Check CSV output in `csv_files/` directory
- Analytics examples available in `analytics/notebook-examples/`