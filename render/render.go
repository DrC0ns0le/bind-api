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

const outputDir = "output"

type SOA struct {
	PrimaryNS  string
	AdminEmail string
	Serial     uint64
	Refresh    uint16
	Retry      uint16
	Expire     uint32
	Minimum    uint16
	TTL        uint16
}

type Record struct {
	Type    string
	Host    string
	Content string
	TTL     uint16
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
				PrimaryNS:  z.PrimaryNS,
				AdminEmail: z.AdminEmail,
				Serial:     uint64(time.Now().Unix()),
				Refresh:    z.Refresh,
				Retry:      z.Retry,
				Expire:     z.Expire,
				Minimum:    z.Minimum,
				TTL:        z.TTL,
			},
		}

		ZS = append(ZS, Z)
	}

	return ZS, nil
}

func renderNamedZones(zones []Zone) (string, error) {
	const fileName = "named.conf.zones"
	const templateName = "bind-named-zones.tmpl"

	// Parse template
	_, filePath, _, _ := runtime.Caller(0)
	templatePath := filepath.Dir(filePath) + "/templates/" + templateName
	t, err := template.New(templateName).ParseFiles(templatePath)
	if err != nil {
		return "", errors.New("Failed to parse template: " + err.Error())
	}

	// Remove the file named.conf.zones if exists in output folder
	if _, err := os.Stat(outputDir + "/" + fileName); err == nil {
		if err = os.Remove(outputDir + "/" + fileName); err != nil {
			return "", err
		}
	}

	// Render template
	path, err := filepath.Abs(filepath.Join(outputDir, fileName))

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
	zs := struct {
		Zones []Zone
	}{
		Zones: zones}
	err = t.Execute(f, zs)
	if err != nil {
		return "", errors.New("Failed to render template: " + err.Error())
	}

	return path, nil
}

func renderZone(zone Zone) (string, error) {
	// Parse template
	_, filePath, _, _ := runtime.Caller(0)
	templatePath := filepath.Dir(filePath) + "/templates/bind-zone.tmpl"
	t, err := template.New("bind-zone.tmpl").ParseFiles(templatePath)
	if err != nil {
		return "", errors.New("Failed to parse template: " + err.Error())
	}

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

	// Render configs
	_, err = renderNamedZones(zs)
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

	// Render named.conf.zones
	fileName := "named.conf.zones"
	templateName := "bind-named-zones.tmpl"

	// Parse template
	_, filePath, _, _ := runtime.Caller(0)
	templatePath := filepath.Dir(filePath) + "/templates/" + templateName
	t, err := template.New(templateName).ParseFiles(templatePath)
	if err != nil {
		return nil, errors.New("Failed to parse template: " + err.Error())
	}

	// Execute the template for each zone and store output in map
	var buf bytes.Buffer
	err = t.Execute(&buf, zones)
	if err != nil {
		return nil, errors.New("Failed to render template: " + err.Error())
	}

	// Store the rendered content for each zone in the output map
	zoneOutputs[fileName] = buf.String()

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
