# podlift Documentation Site

This directory contains the Hugo-based documentation site for podlift.

## Local Development

Install Hugo (extended version required):

```bash
# macOS
brew install hugo

# Linux
wget https://github.com/gohugoio/hugo/releases/download/v0.152.2/hugo_extended_0.152.2_linux-amd64.deb
sudo dpkg -i hugo_extended_0.152.2_linux-amd64.deb
```

Run locally:

```bash
cd docs-site
hugo server
```

Visit http://localhost:1313

## Build

```bash
hugo --minify
```

Output in `public/` directory.

## Deployment

Documentation automatically deploys to GitHub Pages when pushed to main:

https://ekinertac.github.io/podlift/

Workflow: `.github/workflows/docs.yml`

## Theme

Using [hugo-book](https://github.com/alex-shpak/hugo-book) theme for clean, searchable documentation.

## Structure

```
content/
├── _index.md          # Homepage
└── docs/
    ├── _index.md      # Docs overview
    ├── installation.md
    ├── configuration.md
    ├── commands.md
    ├── deployment-guide.md
    ├── how-it-works.md
    ├── troubleshooting.md
    └── migration.md
```

## Updating Documentation

1. Edit markdown files in `content/docs/`
2. Test locally: `hugo server`
3. Commit and push to main
4. GitHub Actions deploys automatically

