package main

import (
	"fmt"
	"github.com/tidwall/gjson"
	"github.com/urfave/cli"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"sort"
	"strings"
)

type dependency struct {
	Group          string
	Name           string
	CurrentVersion string
	LatestVersion  string
}

var resolvedDependencies = make([]dependency, 0)
var unresolvedDependencies = make([]dependency, 0)

func main() {
	app := cli.NewApp()
	app.Name = "vleas"
	app.Usage = "be always up to date, extremely fast ;)"
	app.Version = "0.0.1"
	app.Author = "Nikola Stanković"
	app.Email = "nikola@stankovic.xyz"
	app.Description = "Vleas is an easy to use open source CLI for maintaining deps."

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "file, f",
			Usage: "Load deps from `FILE`",
		},
	}

	app.Commands = []cli.Command{
		{
			Name:    "check",
			Aliases: []string{"c"},
			Usage:   "check for new deps",
			Action: func(c *cli.Context) error {
				check(c.GlobalString("file"))
				return nil
			},
		},
		{
			Name:    "update",
			Aliases: []string{"u"},
			Usage:   "update all deps to latest version",
			Action: func(c *cli.Context) error {
				update(c.GlobalString("file"))
				return nil
			},
		},
	}

	sort.Sort(cli.FlagsByName(app.Flags))
	sort.Sort(cli.CommandsByName(app.Commands))

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func check(file string) {
	contentBytes, _ := ioutil.ReadFile(file)
	content := string(contentBytes)

	regex := regexp.MustCompile("(?P<group>[^\"$,;\\(\\)\\[\\]\\{\\}']+):(?P<name>[^\"$,;\\(\\)\\[\\]\\{\\}']+):(?P<version>[^\"$,;\\(\\)\\[\\]\\{\\}']+)")
	deps := regex.FindAllStringSubmatch(content, -1)

	for i := range deps {
		group := deps[i][1]
		name := deps[i][2]
		latestVersion := fetchLatestDeps(group, name)

		deps[i] = append(deps[i], latestVersion)

		if validDep(deps[i]) {
			resolvedDependencies = append(resolvedDependencies, dependency{
				Group:          deps[i][1],
				Name:           deps[i][2],
				CurrentVersion: deps[i][3],
				LatestVersion:  deps[i][4],
			})
		} else {
			if strings.EqualFold(deps[i][3], deps[i][4]) == false {
				unresolvedDependencies = append(unresolvedDependencies, dependency{
					Group:          deps[i][1],
					Name:           deps[i][2],
					CurrentVersion: deps[i][3],
					LatestVersion:  deps[i][4],
				})
			}
		}
	}
	resolvedDependencies = removeDuplicates(resolvedDependencies)

	if len(resolvedDependencies) > 0 {
		fmt.Printf("\nVleas found %d dependency update(s):\n\n", len(resolvedDependencies))
	} else {
		fmt.Printf("\nGreat! Your project is up to date :)")
	}

	for _, dep := range resolvedDependencies {
		fmt.Printf("group: %s name: %s version: %s --> %s\n", dep.Group, dep.Name, dep.CurrentVersion, dep.LatestVersion)
	}

	if len(unresolvedDependencies) > 0 {
		fmt.Printf("\nThe following dependencies have not been able to check:\n\n")
		for _, dep := range unresolvedDependencies {
			fmt.Printf("group: %s name: %s version: %s\n", dep.Group, dep.Name, dep.CurrentVersion)
		}
	}
}

func validDep(dep []string) bool {
	currentVersion := dep[3]
	latestVersion := dep[4]

	if strings.EqualFold(currentVersion, latestVersion) {
		return false
	}

	if latestVersion == "" {
		return false
	}

	return true
}

func update(file string) {
	fmt.Println("update deps from file: " + file)
}

func fetchLatestDeps(group, name string) string {
	url := "http://search.maven.org/solrsearch/select?q=g:%22#GROUP%22+AND+a:%22#NAME%22&"
	url = strings.Replace(url, "#GROUP", group, 1)
	url = strings.Replace(url, "#NAME", name, 1)

	resp, err := http.Get(url)
	if err != nil {
		log.Fatalln(err)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}

	return gjson.Get(string(body), "response.docs.0.latestVersion").String()
}

func removeDuplicates(elements []dependency) []dependency { // change string to int here if required
	// Use map to record duplicates as we find them.
	encountered := map[dependency]bool{} // change string to int here if required
	var result []dependency              // change string to int here if required

	for v := range elements {
		if encountered[elements[v]] == true {
			// Do not add duplicate.
		} else {
			// Record this element as an encountered element.
			encountered[elements[v]] = true
			// Append to result slice.
			result = append(result, elements[v])
		}
	}
	// Return the new slice.
	return result
}
