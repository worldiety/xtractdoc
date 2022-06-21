# xtractdoc

xtractdoc is a little tool to extract identifiers from Go source code and their according documentation.

## why would one ever need that?

For example, if you want to create a single source of truth for your documentation, but you need that
in other places, like e.g. using that in a _$ref_ description in an OpenAPI specification.

## usage

```bash
go install github.com/worldiety/xtractdoc/cmd/xtractdoc@latest
xtractdoc -modPath=/my/go/module -format=yaml -packages=github.com/worldiety/xtractdoc/testdata > godoc.yaml
```

Example output:

```yaml
# Code generated by github.com/worldiety/xtractdoc DO NOT EDIT.

github.com/worldiety/xtractdoc/testdata: |
  Package testdata is about testing.
github.com/worldiety/xtractdoc/testdata.AConstant: |
  AConstant here.
github.com/worldiety/xtractdoc/testdata.Behavior: |
  A Behavior is what to want.
github.com/worldiety/xtractdoc/testdata.Behavior#DoIt: |
  DoIt does it well.
github.com/worldiety/xtractdoc/testdata.BestFunc: |
  The BestFunc is really a static package level function.
github.com/worldiety/xtractdoc/testdata.Entity: |
  An Entity to store.
github.com/worldiety/xtractdoc/testdata.Entity#Description: |
  A Description about the thing.
github.com/worldiety/xtractdoc/testdata.Entity#Name: |
  A Name to tell about.
github.com/worldiety/xtractdoc/testdata.Entity#NewEntity: |
  NewEntity is a conventional constructor.
github.com/worldiety/xtractdoc/testdata.Entity#String: |
  String returns a human-
  readable representation.

  Second line.
github.com/worldiety/xtractdoc/testdata.Hello: |
  Hello to the world.


```