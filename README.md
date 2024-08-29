# GOpsKit <img src="https://raw.githubusercontent.com/fmjstudios/artwork/0fbaea26cdaae204c9e6a03e5ec61d42d7b60cf7/projects/gopskit/icon/color/gopskit-icon-color.png" alt="GOpsKit Logo" align="right" width="225"/>

`GOpsKit` (Go Ops Kit) is an open-source [MIT][license]-licensed [Go][go]-based toolkit for working with [Kubernetes]
[kubernetes] Clusters `v1.26` and above. The project is built using Google's [Bazel][bazel] build system in
combination with their first-party [Gazelle][gazelle] `BUILD` file generator. The toolkit offers various
functionalities like setting up HashiCorp's [Vault][vault] with [`vaultr`][vaultr] or registering various
applications for SSO authentication with [Keycloak][keycloak] using [`opsctl`][opsctl].

## ✨ TL;DR

```shell
# build all projects at once
bazel build //...
```

## 📖 Overview

Like most modern [Go][go] projects the various executables are located within the [cmd][cmd] directory. Here's a
quick-reference list as an overview:

- [`opsctl`][opsctl]: manage SSO authentication for various apps and more
- [`vault`][vaultr]: configure and manage [HashiCorp's Vault][vault]
- [`fillr`][fillr]: create Helmfile templates automatically

### 🔃 Contributing

Refer to our [documentation for contributors][contributing] for contributing guidelines, commit message
formats and versioning tips.

### 📥 Maintainers

This project is owned and maintained by [FMJ Studios][org] refer to the [`AUTHORS`][authors] or [`CODEOWNERS`]
[owners] for more information. You may also use the linked contact details to reach out directly.

<!-- INTERNAL REFERENCES -->

<!-- Project references -->

[cmd]: cmd
[opsctl]: cmd/opsctl
[vaultr]: cmd/vaultr
[fillr]: cmd/fillr

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
