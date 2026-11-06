# Website

The public landing page lives in `site/` and is deployed to
<https://zachzhao1984.github.io/twintrail/> by `.github/workflows/pages.yml`.
Only `site/` is uploaded. Design explorations under `design/` are not published.

Push a change under `site/` to `main`, or manually run **Deploy website** in
GitHub Actions. The repository's Pages source must be **GitHub Actions**.
Deployment uses the `github-pages` environment with `pages: write` and
`id-token: write`; repository contents are read-only. The application release
workflow remains separate.

The site is plain HTML, CSS and JavaScript with local assets and no build step.
Open `site/index.html` locally to inspect it. `?lang=en` and `?lang=zh-CN` select
a language explicitly; the language buttons remember the choice in local storage.
The default language is Chinese. No analytics or third-party scripts are loaded.

Application screenshots use example project names and are labeled development
previews. Download buttons link to the verified public v1.2.0 release. When the
app version changes, update both language dictionaries and release/download URLs
in `site/index.html` and `site/script.js` together.

Checks before publishing:

- `node --check site/script.js`
- Check Chinese/English at 1440px and 375px, as well as keyboard navigation,
  appearance switching, and the shortcut illustration.
- Confirm relative asset URLs load under the `/twintrail/` project path and the
  release/download links still refer to real assets.

To roll back the site, revert its commit and push to `main`, or rerun a previous
successful **Deploy website** workflow. This does not roll back the native app.
