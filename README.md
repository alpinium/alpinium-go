# alpinium-go

Go monorepo s knihovnami Alpinium. Každá knihovna je samostatný Go modul se
svým `go.mod` a vlastní verzí.

| Složka | Modul | Popis |
|---|---|---|
| `Cast` | `github.com/alpinium/alpinium-go/Cast` | bezpečné převody mezi bool, čísly a string s kontrolou rozsahu |

## Instalace

```bash
go get github.com/alpinium/alpinium-go/Cast@v0.1.0
```

## Lokální vývoj

```bash
cd Cast
go vet ./...
go test -count=1 ./...
```

`go.work` je v `.gitignore`. CI i release běží s `GOWORK=off`.

## Verzování a vydání

Verze modulu je git tag s prefixem složky, např. `Cast/v0.2.0`. Vydání spustí
pushnutí tagu:

```bash
git tag Cast/v0.2.0
git push origin Cast/v0.2.0
```

Workflow `Release` (`.github/workflows/release.yml`) se spustí na každý tag
`<složka>/vX.Y.Z`. Zkontroluje formát verze a že tag leží na `main`, spustí
`go vet`, `go test -race` a `go mod tidy -diff`, vytvoří GitHub Release
(verze `v0.x` a `-rc` jako pre-release) a zahřeje Go proxy.

Pravidla ([semver](https://semver.org/)):

- `v0.x.y` dokud se API může měnit. Nová funkce nebo změna chování: zvýšit `x`, oprava chyby: zvýšit `y`.
- Od `v1.0.0`: oprava chyby `y`, nová funkce zpětně kompatibilní `x`, nekompatibilní změna `v2`.
- `v2` a výš vyžaduje v `go.mod` cestu `.../Cast/v2` a tag `Cast/v2.0.0`.
- Tag je neměnný. Chybnou verzi nelze přepsat, jen vydat novou (případně `retract` v `go.mod`).
