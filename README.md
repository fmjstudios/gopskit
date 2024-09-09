# Go Operations Toolkit <img src="https://github.com/fmjstudios/artwork/blob/5f7696fd09584e4d34d2ab0de3eb52b86037d924/projects/gopskit/icon/color/gopskit-icon-color.png?raw=true" alt="GOpsKit Logo" align="right" width="225"/>

`GOpsKit` (**Go** **Op**erations Tool**kit**) is an open-source [MIT][license]-licensed [Go][go]-based toolkit for working with [Kubernetes]
[kubernetes] Clusters `v1.26` and above. The project is built using Google's [Bazel][bazel] build system in
combination with their first-party [Gazelle][gazelle] `BUILD` file generator.

## 📖 Overview

The toolkit offers a plethora of functionalities like setting up HashiCorp's [Vault][vault] with [`waltr`][waltr], registering various
applications for SSO authentication with [Keycloak][keycloak] using [`ssolo`][ssolo]. Never write [Helmfile][helmfile] `values.yaml`
template files to manage applications on your cluster again. Instead generate them using [`fillr`][fillr]. Are you running your own custom
private Certificate Authority using [Smallstep's CA][smallstep_certificates]? Then you'd likely want to generate and manage PKI values
using [`steppa`][steppa]. The German KBA delivers data in a custom bespoke text-based format, which purely relies on columns to separate
data. _That ain't SQL..._ So let's swiftly generate some usable SQL import script using [`amtrac`][amtrac].

## ✨ TL;DR

```shell
# build all projects at once - requires Bazel at .bazelversion
bazel build //...
```

## 🛠️ Tools

Like most modern [Go][go] projects the various executables are located within the [cmd][cmd] directory. Here's a
quick-reference list as an overview:

- [`ssolo`][ssolo]: manage SSO authentication for various apps using Keycloak
- [`waltr`][waltr]: configure and manage [HashiCorp's Vault][vault]
- [`fillr`][fillr]: create Helmfile templates automatically
- [`steppa`][steppa]: generate and manage SmallStep PKI values
- [`amtrac`][amtrac]: generate SQL dumps from the German KBA's data files using Docker

### 🔃 Contributing

Refer to our [documentation for contributors][contributing] for contributing guidelines, commit message
formats and versioning tips.

### 📥 Maintainers

This project is owned and maintained by [FMJ Studios][org] refer to the [`AUTHORS`][authors] or [`CODEOWNERS`][owners]
for more information. You may also use the linked contact details to reach out directly.

### ©️ Copyright

- _Assets provided by:_ **[IconScout](https://iconscout.com)**
- _Sources provided by:_ **[FMJ Studios][org]** under the **[MIT License][license]**

<!-- INTERNAL REFERENCES -->

<!-- Project references -->

[cmd]: cmd
[ssolo]: cmd/ssolo
[waltr]: cmd/waltr
[fillr]: cmd/fillr
[steppa]: cmd/steppa
[amtrac]: cmd/amtrac

<!-- File references -->

[license]: LICENSE
[contributing]: docs/CONTRIBUTING.md
[authors]: .github/AUTHORS
[owners]: .github/CODEOWNERS

<!-- General links -->

[org]: https://github.com/fmjstudios
[kubernetes]: https://kubernetes.io
[vault]: https://vaultproject.io
[keycloak]: https://www.keycloak.org/
[go]: https://go.dev
[bazel]: https://bazel.build
[gazelle]: https://github.com/bazelbuild/bazel-gazelle
[helmfile]: https://github.com/helmfile/helmfile
[smallstep_certificates]: https://github.com/smallstep/certificates
