package platform

// ref: https://pkg.go.dev/github.com/elastic/go-sysinfo@v1.14.1
// ref: https://stackoverflow.com/questions/44363911/detect-windows-version-in-go-to-figure-out-the-starup-folder

var (
	// ref: https://gist.github.com/eculver/d1338aa87e87890e05d4f61ed0a33d6e
	OSReleaseRegex string = `^(?<variable>[A-Za-z\d_-]+)=(?<startQuote>['"]?)(?<value>.[^"']*)(?<endQuote>['"]?)$`
)

type Info struct {
	version string
	name    string
}
