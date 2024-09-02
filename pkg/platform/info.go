package platform

var (
	OSReleaseRegex string = `^(?<variable>[A-Za-z\d_-]+)=(?<startQuote>['"]?)(?<value>.[^"']*)(?<endQuote>['"]?)$`
)

type Info struct {
	version string
	name    string
}
