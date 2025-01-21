# VIAC 3a Wealth Getter
Gets total monies of a VIAC 3a portfolio.

## Usage
```bash
export VIAC_USER="+41..."
export VIAC_PASSWORD="..."

go run .
```

If successful, will write a JSON to stdout:
```json
{"totalValue":1337.20,"totalPerformance":0.1234,"totalReturn":1337.20}
```

Note: `totalPerformance` is a percentage.