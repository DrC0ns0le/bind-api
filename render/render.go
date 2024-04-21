package render

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"text/template"
	"time"

	"github.com/DrC0ns0le/bind-api/rdb"
)

type SOA struct {
	PrimaryNS  string
	AdminEmail string
	Serial     int
	Refresh    int
	Retry      int
	Expire     int
	Minimum    int
	TTL        int
}

type Record struct {
	Type    string
	Host    string
	Content string
	TTL     int
}

type Zone struct {
	Name    string
	Records []Record
	SOA     SOA
}

func createZones(_bd *rdb.BindData) ([]Zone, error) {
	var ZS []Zone

	zs, err := _bd.Zones.Get()

	if err != nil {
		return ZS, err
	}

	for _, z := range zs {
		rs, err := _bd.Records.Get(z.UUID)

		if err != nil {
			return ZS, err
		}

		var RS []Record
		for _, r := range rs {
			RS = append(RS, Record{
				Type:    r.Type,
				Host:    r.Host,
				Content: r.Content,
				TTL:     r.TTL,
			})
		}

		Z := Zone{
			Name:    z.Name,
			Records: RS,
			SOA: SOA{
				PrimaryNS:  "ns.placeholder",
				AdminEmail: "webmaster",
				Serial:     int(time.Now().Unix()),
				Refresh:    3600,
				Retry:      600,
				Expire:     604800,
				Minimum:    1800,
				TTL:        3600,
			},
		}

		ZS = append(ZS, Z)
	}

	return ZS, nil
}

func renderZone(zone Zone) (string, error) {
	_, filePath, _, _ := runtime.Caller(0)
	templatePath := filepath.Dir(filePath) + "/templates/bind-zone.tmpl"
	t, err := template.New("bind-zone.tmpl").ParseFiles(templatePath)
	if err != nil {
		return "", errors.New("Failed to parse template: " + err.Error())
	}

	outputDir := "output"
	filename := fmt.Sprintf("%s-rendered_%d.conf", zone.Name, time.Now().Unix())
	path, err := filepath.Abs(filepath.Join(outputDir, filename))

	if err != nil {
		return "", errors.New("Failed to create file path: " + err.Error())
	}

	f, err := os.Create(path)
	if err != nil {
		return "", errors.New("Failed to create output file: " + err.Error())
	}
	defer f.Close()

	err = t.Execute(f, zone)
	if err != nil {
		return "", errors.New("Failed to render template: " + err.Error())
	}

	return path, nil
}

func RenderZonesTemplate(_bd *rdb.BindData) error {
	zs, err := createZones(_bd)
	if err != nil {
		return err
	}

	for _, z := range zs {
		_, err := renderZone(z)
		if err != nil {
			return err
		}
	}
	return nil
}
