package render

import (
	"bytes"
	"context"
	"errors"
	"path/filepath"
	"runtime"
	"text/template"
)

func PreviewZoneRender(ctx context.Context) (map[string]string, error) {

	// Create all zones
	zones, err := createZones(ctx)
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
	zs := struct {
		Zones []Zone
	}{
		Zones: zones}
	err = t.Execute(&buf, zs)
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
		zoneOutputs[zone.Name+".conf"] = buf.String()
	}

	return zoneOutputs, nil
}
