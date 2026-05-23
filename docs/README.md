# Gutter documentation

This folder is the source for the Gutter docs site, hosted with GitHub Pages.
It uses [Jekyll](https://jekyllrb.com/) and the
[Just the Docs](https://just-the-docs.com/) theme (loaded as a remote theme,
so no local install is needed for GitHub to build it).

## Enabling on GitHub Pages

Repository → Settings → Pages → Build and deployment:

- **Source**: Deploy from a branch
- **Branch**: `main` (or whatever your default branch is)
- **Folder**: `/docs`

Save. GitHub Actions will publish the site at
`https://runway-club.github.io/gutter/` (or your fork's equivalent).

## Previewing locally (optional)

You only need this if you want to edit the site offline. GitHub builds the
site on push regardless.

```sh
cd docs
bundle install
bundle exec jekyll serve
# open http://127.0.0.1:4000
```

## Editing

- Each page is a Markdown file with YAML frontmatter.
- The sidebar nav order comes from each page's `nav_order:` frontmatter.
- Widget pages live under `widgets/` and are children of the `Widgets`
  parent page (`widgets/index.md`).
- Assets live in `assets/`. The icon is `assets/gutter_icon.png`.

Add a new top-level page by creating a `<name>.md` at the docs root with:

```yaml
---
title: My New Page
nav_order: 99
---
```

Add a new widget page by creating `widgets/<name>.md` with:

```yaml
---
title: MyWidget
parent: Widgets
nav_order: 20
---
```
