# Packaging

## AUR

`PKGBUILD` builds golarm from the release tarball (`makepkg -si` after
setting `sha256sums`). Submit to the AUR with:

```sh
makepkg --printsrcinfo > .SRCINFO
git init . && git add PKGBUILD .SRCINFO
git commit -m "golarm 0.1.0"
git remote add aur ssh://aur@aur.archlinux.org/golarm.git
git push aur master
```

Tag a release first (`git tag v0.1.0 && git push --tags`) — goreleaser
publishes the tarball the PKGBUILD sources.

## Manual install (any distro)

```sh
make install        # go install with version ldflags
# or
make build && sudo install -Dm755 golarm /usr/local/bin/golarm
```
