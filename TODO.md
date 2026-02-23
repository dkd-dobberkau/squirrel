# TODO

## Homebrew Tap

- [ ] GitHub Repository `dkd-dobberkau/homebrew-tap` erstellen
- [ ] Formula `squirrel.rb` schreiben (auf GoReleaser Artifacts verweisen)
- [ ] Installation testen: `brew install dkd-dobberkau/tap/squirrel`

## GoReleaser

- [ ] `.goreleaser.yml` erstellen (Multi-Arch: darwin/arm64, darwin/amd64, linux/amd64, linux/arm64)
- [ ] GitHub Action fuer automatische Releases bei Tag-Push einrichten
- [ ] Homebrew Tap in GoReleaser Config integrieren (`brews:` Section)
- [ ] Erster Release: `git tag v0.1.0 && git push --tags`

## Sonstiges

- [ ] `squirrel project <pfad>` Befehl implementieren (Detail-Ansicht)
- [ ] Deep-Modus: Session-JSONL parsen fuer TODO-Extraktion
