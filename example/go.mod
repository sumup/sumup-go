module github.com/sumup/sumup-go/example

go 1.17

require (
	github.com/sumup/sumup-go v0.0.0-20230919081147-7283f347e1b5
	golang.org/x/exp v0.0.0-20230905200255-921286631fa9
)

// Use version at HEAD, not the latest published.
replace github.com/sumup/sumup-go => ../
