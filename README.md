# README.md

Gejie CLI is being built as a data analytics tool to get LATAM eCommerce datasets in a convenient way for product and market research for our manufacturing and direct commerce partners

We plan to support the efficient and dev friendly scraping of MercadoLibre, Coolbox.pe (wip), Falabella (wip) and more

### CLI Usage

Install or build the CLI:

```bash
# Option 1: build locally
go build -o gejiec .

```

Basic help:

```bash
./gejiec --help
./gejiec meli --help
```

Run MercadoLibre scraper (`meli` subcommand):

- **Scrape a single product page**:

```bash
./gejiec meli --url "https://articulo.mercadolibre.com.mx/MLM-1411526559-silla-gamer-reclinable-giratoria-ergonomica-super-comoda-_JM"
```

- **Only fetch product images from a product page**:

```bash
./gejiec meli --url "https://articulo.mercadolibre.com.mx/MLM-1411526559-..." --only-images
```

- **Scrape a search/listing page (auto-paginates) and limit items**:

```bash
./gejiec meli --url "https://listado.mercadolibre.com.mx/carburador-stihl" --max-items 5
```

- **Create a CSV from a listing scrape**:

```bash
./gejiec meli --url "https://listado.mercadolibre.com.mx/carburador-stihl" --max-items 20 --create-csv
```

- **Run headless (no browser window)**:

```bash
./gejiec meli --url "https://listado.mercadolibre.com.mx/carburador-stihl" --headless
```

Available flags for `meli`:

- `--url` (string, required): MercadoLibre URL. Product and listing pages are auto-detected.
- `--max-items` (int, default 10): Maximum items to scrape (listing pages only).
- `--only-images` (bool): Only scrape image URLs from a product page.
- `--create-csv` (bool): Write scraped listing products to a CSV file.
- `--headless` (bool): Run browser automation in headless mode.

## Feature Roadmap

- add product stats using Go and example `python-notebooks`
- add support residential proxies
