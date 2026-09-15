# qrcode-generator

An API for generating QR code images.

Open `/` in a browser to use the interactive playground.

```text
GET /qr?data=hello&size=256
GET /microqr?data=12345&size=256
GET /rmqr?data=hello&size=256
```

Query parameters:

- `data`: The string to encode in the QR code
- `size`: The output width and height in pixels (optional integer from 1 to 4096). For PNG output, the value must be at least the QR code's natural width, including its quiet zone. This minimum depends on the data and version and ranges from 29 to 185 pixels. SVG output accepts the full range.
- `format`: `png` (default) or `svg`
- `level`: The error correction level: `L`, `M`, `Q`, or `H`
- `version`: The QR code version (1–40)

`/microqr` accepts the same parameters with these differences:

- `level`: `check`, `L`, `M`, or `Q` (availability depends on the version)
- `version`: The Micro QR code version (1–4, corresponding to M1–M4)
- The natural PNG width includes the Micro QR code's two-module quiet zone and ranges from 15 to 21 pixels.

`/rmqr` generates a rectangular Micro QR code and accepts these parameters:

- `data`: The string to encode
- `size`: The output width in pixels (optional integer from 1 to 4096). The height is calculated from the rMQR aspect ratio. PNG output requires at least the natural width including the two-module quiet zone; SVG accepts the full range.
- `format`: `png` (default) or `svg`
- `level`: `M` (default) or `H`
- `version`: One of the 32 rMQR version names, from `R7x43` through `R17x139` (for example, `R11x27`)
- `priority`: Version selection priority: `area` (default), `height`, or `width`
