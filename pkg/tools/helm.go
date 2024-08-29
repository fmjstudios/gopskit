package tools

// HelmPlugin represents a Helm plugin required for gopskit to work
type HelmPlugin string

var (
	pluginDiff    HelmPlugin = "diff"
	pluginSecrets HelmPlugin = "secrets"

	// HelmPlugins is a list of Helm Plugins
	plugins    = []HelmPlugin{pluginDiff, pluginSecrets}
	pluginURLs = map[HelmPlugin]string{
		pluginDiff:    "https://github.com/databus23/helm-diff",
		pluginSecrets: "https://github.com/jkroepke/helm-secrets",
	}
)
