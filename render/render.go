package render

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
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
	// Parse template
	_, filePath, _, _ := runtime.Caller(0)
	templatePath := filepath.Dir(filePath) + "/templates/bind-zone.tmpl"
	t, err := template.New("bind-zone.tmpl").ParseFiles(templatePath)
	if err != nil {
		return "", errors.New("Failed to parse template: " + err.Error())
	}

	outputDir := "output"

	// Remove everything in output folder except output folder
	if file, err := os.Stat(outputDir); err == nil {

		if !file.IsDir() {
			return "", errors.New("not a directory")
		}

		files, err := os.ReadDir(outputDir)
		if err != nil {
			return "", err
		}

		for _, f := range files {
			filePath := outputDir + "/" + f.Name()
			// only remove file with the same starting zone name
			if !strings.HasPrefix(f.Name(), zone.Name) {
				continue
			}
			if err = os.RemoveAll(filePath); err != nil {
				return "", err
			}
		}
	}

	// Render template
	filename := fmt.Sprintf("%s.conf", zone.Name)
	path, err := filepath.Abs(filepath.Join(outputDir, filename))

	if err != nil {
		return "", errors.New("Failed to create file path: " + err.Error())
	}

	// Create output file
	f, err := os.Create(path)
	if err != nil {
		return "", errors.New("Failed to create output file: " + err.Error())
	}
	defer f.Close()

	// Execute template
	err = t.Execute(f, zone)
	if err != nil {
		return "", errors.New("Failed to render template: " + err.Error())
	}

	return path, nil
}

func RenderZonesTemplate(_bd *rdb.BindData) error {

	// Create all zones
	zs, err := createZones(_bd)
	if err != nil {
		return err
	}

	// Render all zones
	for _, z := range zs {
		_, err := renderZone(z)
		if err != nil {
			return err
		}
	}

	// Commit all changes
	err = _bd.Records.CommitAll()
	if err != nil {
		return err
	}

	return nil
}

func PreviewZoneRender(_bd *rdb.BindData) (map[string]string, error) {

	// Create all zones
	zones, err := createZones(_bd)
	if err != nil {
		return nil, err
	}

	zoneOutputs := make(map[string]string)

	// Render all zones using a template and store the output in a map
	for _, zone := range zones {
		// Parse template
		_, filePath, _, _ := runtime.Caller(0)
		templatePath := filepath.Dir(filePath) + "/templates/bind-zone.tmpl"
		t, err := template.New("bind-zone.tmpl").ParseFiles(templatePath)
		if err != nil {
			return nil, errors.New("Failed to parse template: " + err.Error())
		}

		// Execute the template for each zone and store output in map
		var buf bytes.Buffer
		err = t.Execute(&buf, zone)
		if err != nil {
			return nil, errors.New("Failed to render template: " + err.Error())
		}

		// Store the rendered content for each zone in the output map
		zoneOutputs[zone.Name] = buf.String()
	}

	return zoneOutputs, nil
}
