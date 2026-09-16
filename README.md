# qrcode-generator

An API for generating QR code images.

Open `/` in a browser for the landing page, which links to each interactive
playground: `/qr/playground` for QR, `/microqr/playground` for Micro QR, and
`/rmqr/playground` for rMQR.

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

## Deployment permissions

GitHub Actions reads the `SourceBucket` output of the existing
`aws-sam-cli-managed-default` stack and passes it to `sam deploy` with
`--no-resolve-s3`. Provision this stack and bucket using an administrator or
bootstrap role before running CI; the CI role only has `DescribeStacks` access
to this stack. Local `make deploy` still uses SAM's automatic bucket management.
After changing `cicd.yaml`, run `make cicd` with credentials authorized to update
the CI/CD stack before rerunning GitHub Actions.

`cicd.yaml` scopes deployment permissions to the `qrcode-generator` stack,
its generated function and execution role, the configured Route 53 hosted zone,
and the `qrcode-generator/` artifact prefix. Deploy the CI/CD stack in the same
region as the application. Update these scopes if the stack name, function
logical ID, domain, hosted zone, or artifact prefix changes.

`acm:RequestCertificate` is the only Allow statement with `Resource: "*"`:
[AWS does not support resource-level permissions for this action](https://docs.aws.amazon.com/service-authorization/latest/reference/list_acm.html).
It is restricted to `qr.shogo82148.com`, DNS validation, and the stack's region.
If review policy prohibits even this exception, provision the certificate
separately and pass its ARN into the application template instead.

Generated certificate and HTTP API IDs still require wildcard ARN suffixes;
these permissions cover certificates in the account/region and HTTP APIs in the
region, respectively. They are not restricted to a single existing resource.
The function permissions boundary permits only writing its own CloudWatch logs.

After applying the CI/CD stack, verify application creation, update, and rollback
in AWS. Local template validation does not verify the deployment role's effective
permissions or organization-level policies.
