# TODO

## Homebrew Tap

- [x] GitHub Repository `dkd-dobberkau/homebrew-tap` erstellen
- [x] Cask via GoReleaser automatisch generieren
- [ ] Installation testen: `brew install dkd-dobberkau/tap/squirrel`

## GoReleaser

- [x] `.goreleaser.yaml` erstellen (Multi-Arch: darwin/arm64, darwin/amd64, linux/amd64, linux/arm64)
- [x] GitHub Action fuer automatische Releases bei Tag-Push einrichten
- [x] Homebrew Tap in GoReleaser Config integrieren (`homebrew_casks:` Section)
- [x] Erster Release: `git tag v0.1.0 && git push --tags`

## Sonstiges

- [x] `squirrel project <pfad>` Befehl implementieren (Detail-Ansicht)
- [x] Deep-Modus: Session-JSONL parsen fuer TODO-Extraktion
