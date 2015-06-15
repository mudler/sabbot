package packages

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"log"
	"strings"
)

type Package struct {
	name     string
	url      string
	flags    string
	diskSize string
	size     string
}

func Search(q string) ([]Package, string) {
	var packages []Package
	query := "https://packages.sabayon.org/quicksearch?q=" + q
	doc, err := goquery.NewDocument(query)
	if err != nil {
		log.Fatal(err)
	}

	doc.Find(".package-widget-meta-atom").Each(func(i int, s *goquery.Selection) {
		atom, _ := s.Find("a").Attr("href")
		url := "https://packages.sabayon.org/" + strings.TrimSpace(atom)
		packageDoc, _ := goquery.NewDocument(url)
		flags := strings.Join(strings.Fields(packageDoc.Find(".package-widget-meta-list-left-useflags dd").Text()), " ")
		diskSize := strings.TrimSpace(packageDoc.Find(".package-widget-meta-list-left-ondisksize dd").Text())
		size := strings.TrimSpace(packageDoc.Find(".package-widget-meta-list-left-size dd").Text())

		packages = append(packages, Package{name: strings.TrimSpace(s.Text()), url: url, flags: flags,
			size: size, diskSize: diskSize})
	})
	return packages, query
}

func ReverseDeps(s string) ([]Package, string) {

	search, _ := Search(s)
	query := search[0].url + "/reverse_dependencies"
	//Searching reverse deps of the first one
	var packages []Package // an empty list

	doc, err := goquery.NewDocument(query)
	if err != nil {
		log.Fatal(err)
	}

	doc.Find(".package-widget-show-deps-item").Each(func(i int, s *goquery.Selection) {
		url, _ := s.Find("a").Attr("href")
		packages = append(packages, Package{name: strings.TrimSpace(s.Text()), url: "https://packages.sabayon.org/" + strings.TrimSpace(url)})
	})
	return packages, query
}

func (p *Package) String() string {
	switch {
	case len(p.flags) != 0 && len(p.size) != 0 && len(p.diskSize) != 0:
		return fmt.Sprintf("%s %s , Flags: %s, Disk Size: %s, Size: %s", p.name, p.url, p.flags, p.diskSize, p.size)
	default:
		return fmt.Sprintf("%s %s", p.name, p.url)
	}
}
