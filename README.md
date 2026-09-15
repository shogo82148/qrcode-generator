# qrcode-generator

An API for generating QR code images.

```text
GET /qr?data=hello&size=256
```

Query parameters:

- `data`: The string to encode in the QR code
- `size`: The output width and height in pixels (optional integer from 1 to 4096). For PNG output, the value must be at least the QR code's natural width, including its quiet zone. This minimum depends on the data and version and ranges from 29 to 185 pixels. SVG output accepts the full range.
- `format`: `png` (default) or `svg`
- `level`: The error correction level: `L`, `M`, `Q`, or `H`
- `version`: The QR code version (1–40)
