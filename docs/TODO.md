# ✅ FMJ Studios Go Ops Kit - `TODO`s

## ➕ Additions

- [x] Add [`.bazelrc`](https://bazel.build/run/bazelrc)
- [ ] Add `.env` support with [`godotenv`](https://github.com/joho/godotenv)
- [ ] add `config` package with file-configuration based on `type`
- [ ] add `--version`, `--overwrite`, `--silent` flags
- [ ] add `ui` package
- [ ] add local BadgerDB with `dump` commands for namespaces

## ✏️ Planned Changes

- [x] merge `platform` and `filesystem` packages into `fs`
- [ ] implement better logging format
- [x] remove obsolete `errors` package
- [x] merge `env` and `common` packages
- [x] make current code asynchronous
- [x] do generel code-cleanup
- [x] remove non-implemented code besides one-liners in `main.go`
- [x] rename `util` to `helpers` for consistency
- [x] rename `cmd` to `proc` for consistency
- [x] shorten field names in `core.App`
- [x] rename `logger` to just `log`

## 💡 Ideas

## 🔗 Links

### Kubernetes

- [Kubernetes Go Client - Examples](https://github.com/iximiuz/client-go-examples)
- [Kubernetes Go Client - Exec Example](https://github.com/a4abhishek/Client-Go-Examples)
- [Fedora - Executing remote Kubernetes processes](https://miminar.fedorapeople.org/_preview/openshift-enterprise/registry-redeploy/go_client/executing_remote_processes.html)
- [Kubernetes Go Client - Port Forward Example](https://github.com/gianarb/kube-port-forward/blob/master/main.go)
- [Kubernetes Go Client - Generic Forwarder](https://github.com/anthhub/forwarder)
- [Kubernetes Go Client - Dynamic Client (for CRDs)](https://caiorcferreira.github.io/post/the-kubernetes-dynamic-client/)
- [Kubernetes Kubebuilder - Building a CRD client](https://geek.jasonhancock.com/2023/07/28/k8s-crd-golang-kubebuilder/)

### Bazel

- [Bazel's Kubernetes Rules](https://github.com/bazelbuild/rules_k8s/tree/master)

### Go

- [Peter Bourgon's `mergemap` package](https://github.com/peterbourgon/mergemap)
- [Benchmark of Flattening nested maps](https://gist.github.com/knadh/9520b2a3f8edf589c450ed7e283ba60f)
- [GHetzel (Shutterstock) Utility Library](https://github.com/ghetzel/go-stockutil)
- [System Information - elastic](https://pkg.go.dev/github.com/elastic/go-sysinfo@v1.14.1)
- [Windows version -
  `windows` package](https://stackoverflow.com/questions/44363911/detect-windows-version-in-go-to-figure-out-the-starup-folder)
- [RegEx Example](https://gist.github.com/eculver/d1338aa87e87890e05d4f61ed0a33d6e)
- [Go Command Execution Examples](https://github.com/kjk/the-code/blob/master/go/advanced-exec/03-live-progress-and-capture-v2.go)
- [Jay Conrod's `ctxio` package](https://github.com/jayconrod/go-public/blob/main/ctxio/ctxio.go)

## 🗒️ Notes

```go
 // YAML Merge

path := os.Args[1]

mp, err := tools.AddSecretValue(path, map[string]interface{}{
"hooks": map[string]interface{}{
"awxToken":       "fick dich",
"kubescapeToken": "bastard",
"vaultToken":     "thisIsANewValue",
},
}, true)

if err != nil {
panic(err.Error())
}

for k, v := range mp {
fmt.Printf("Key: %s - Value: %s\n", k, v)
}

content, err := yaml.Marshal(mp)
if err != nil {
panic(err)
}

if err := os.WriteFile("/tmp/gopskit-test/fillr-out-values.yaml", content, 0600); err != nil {
panic(err)
}

// GIT

dir, err := os.Getwd()
if err != nil {
panic(err)
}

git, err := filesystem.RevParseGitRoot(dir)
if err != nil {
panic(err)
}

fmt.Printf("Found Git directory at: %s\n", git)

// SmallStep
res, err := tools.GenerateStepValues()
if err != nil {
log.Fatal(err)
}

mp, err := tools.AddSecretStepValues(res, util.GeneratePassphrase(util.WithLength(48)), os.Args[1])
if err != nil {
log.Fatal(err)
}

for k, v := range mp {
fmt.Printf("Key: %s - Value: %s\n", k, v)
}
```
