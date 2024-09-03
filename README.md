# Go Ops Kit <img src="https://raw.githubusercontent.com/fmjstudios/artwork/0fbaea26cdaae204c9e6a03e5ec61d42d7b60cf7/projects/gopskit/icon/color/gopskit-icon-color.png" alt="GOpsKit Logo" align="right" width="225"/>

`GOpsKit` (Go Ops Kit) is an open-source [MIT][license]-licensed [Go][go]-based toolkit for working with [Kubernetes]
[kubernetes] Clusters `v1.26` and above. The project is built using Google's [Bazel][bazel] build system in
combination with their first-party [Gazelle][gazelle] `BUILD` file generator. The toolkit offers various
functionalities like setting up HashiCorp's [Vault][vault] with [`waltr`][waltr] or registering various
applications for SSO authentication with [Keycloak][keycloak] using [`ssolo`][ssolo].

## ✨ TL;DR

```shell
# build all projects at once
bazel build //...
```

## 📖 Overview

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

This project is owned and maintained by [FMJ Studios][org] refer to the [`AUTHORS`][authors] or [`CODEOWNERS`]
[owners] for more information. You may also use the linked contact details to reach out directly.

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
