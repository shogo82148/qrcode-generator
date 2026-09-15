# qrcode-generator

An API for generating QR code images.

```text
GET /qr?data=hello&size=256
```

Query parameters:

- `data`: The string to encode in the QR code
- `size`: The output width and height in pixels (optional integer from 1 to 4096)
- `format`: `png` (default) or `svg`
- `level`: The error correction level: `L`, `M`, `Q`, or `H`
- `version`: The QR code version (1–40)
