//go:build !js
// +build !js

//go:generate gopherjs build -o browser.js index.go

package main

func main() {}
