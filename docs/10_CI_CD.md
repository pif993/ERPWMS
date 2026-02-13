# CI/CD

GitHub Actions workflows:
- `ci.yml`: lint/test/security checks for Go + Python, SBOM artifact.
- `docker.yml`: container build + baseline image scan placeholder.

Branch protection should require all CI checks.
