package parser

import (
	"fmt"
	"io/ioutil"
	"sort"

	"log"
	"regexp"

	"os"

	"github.com/corsc/go-tools/package-coverage/utils"
)

// CoverageByPackage contains the result of parsing one or more package's coverage file
type coverageByPackage map[string]*coverage

// PrintCoverageSingle is the same as PrintCoverage only for 1 directory only
func PrintCoverageSingle(path string, matcher *regexp.Regexp, minCoverage int) bool {
	var fullPath string
	if path == "./" {
		fullPath = utils.GetCurrentDir()
	} else {
		fullPath = utils.GetCurrentDir() + path + "/"
	}
	fullPath += "profile.cov"

	return print([]string{fullPath}, matcher, minCoverage)
}

// PrintCoverage will attempt to calculate and print the coverage from the supplied coverage file
func PrintCoverage(basePath string, matcher *regexp.Regexp, minCoverage int) bool {
	paths, err := utils.FindAllCoverageFiles(basePath)
	if err != nil {
		log.Panicf("error file finding coverage files %s", err)
	}
	return print(paths, matcher, minCoverage)
}

func print(paths []string, matcher *regexp.Regexp, minCoverage int) bool {
	var contents string
	for _, path := range paths {
		if matcher.FindString(path) != "" {
			log.Printf("Printing of coverage for path '%s' skipped due to skipDir regex '%s'", path, matcher.String())
			continue
		}

		contents += getFileContents(path)
	}

	coverageData := calculateCoverage(contents)
	pkgs := getSortedPackages(coverageData)
	return printCoverage(pkgs, coverageData, float64(minCoverage))
}

// CalculateCoverage will calculate and return the coverage for a package or packages from the supplied coverage file contents
func calculateCoverage(contents string) coverageByPackage {
	coverageData := parseLines(contents)
	return coverageData
}

func getFileContents(filename string) string {
	contents, err := ioutil.ReadFile(filename)
	if err != nil {
		panic(err)
	}

	return string(contents)
}

func getSortedPackages(coverageData coverageByPackage) []string {
	output := []string{}

	for pkg := range coverageData {
		output = append(output, pkg)
	}

	sort.Strings(output)

	return output
}

func printCoverage(pkgs []string, coverageData coverageByPackage, minCoverage float64) bool {
	fmt.Printf("  %%		Statements	Package\n")

	coverageOk := true
	for _, pkg := range pkgs {
		cover := coverageData[pkg]
		covered, stmts := getStats(cover)

		if covered < minCoverage {
			fmt.Fprintf(os.Stderr, "\033[1;31m%2.2f		%5.0f		%s\033[0m\n", covered, stmts, pkg)
			coverageOk = false
		} else {
			fmt.Fprintf(os.Stdout, "%2.2f		%5.0f		%s\n", covered, stmts, pkg)
		}
	}
	fmt.Println()

	return coverageOk
}

func getStats(cover *coverage) (float64, float64) {
	stmts := float64(cover.selfStatements + cover.childStatements)
	stmtsCovered := float64(cover.selfCovered + cover.childCovered)

	covered := (stmtsCovered / stmts) * 100

	return covered, stmts
}
