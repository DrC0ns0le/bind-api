package render

import (
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/DrC0ns0le/bind-api/rdb"
)

const (
	outputDir = "output"
)

var (
	// PTR errors
	ErrUnsupportedRecordType = errors.New("Unsupported record type:")
	ErrUnsupportedIPv4       = errors.New("Unsupported IPv4 address:")
	ErrUnsupportedIPv6       = errors.New("Unsupported IPv6 address:")
)

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

func createZones() ([]Zone, error) {
	var ZS []Zone
	rDNS := make(map[string][]Record)

	addReverseDNS := func(z rdb.Zone, r rdb.Record, rDNS *map[string][]Record, a string) {
		if _, ok := (*rDNS)[a]; !ok {
			(*rDNS)[a] = []Record{}
		}
		(*rDNS)[a] = append((*rDNS)[a], Record{
			Type:    r.Type,
			Host:    r.Host + "." + z.Name + ".",
			Content: r.Content,
			TTL:     r.TTL,
		})
	}

	zs, err := (&rdb.Zone{}).Get()
	if err != nil {
		return ZS, err
	}

	for _, z := range zs {
		rs, err := (&rdb.Record{}).Get(z.UUID)
		if err != nil {
			return ZS, err
		}

		var RS []Record
		for _, r := range rs {
			RS = append(RS, Record{
				Type: r.Type,
				Host: r.Host,
				Content: func() string {
					if r.Type == "CNAME" {
						return r.Content + "."
					}
					return r.Content
				}(),
				TTL: r.TTL,
			})

			// Create record for reverse lookup
			if r.AddPTR {
				if r.Type == "A" {
					parts := strings.Split(r.Content, ".")
					if parts[0] == "10" {
						addReverseDNS(z, r, &rDNS, "10.in-addr.arpa")
					} else if parts[0] == "192" && parts[1] == "168" {
						addReverseDNS(z, r, &rDNS, "168.192.in-addr.arpa")
					} else if parts[0] == "172" {
						val, err := strconv.Atoi(parts[1])
						if err != nil {
							return nil, err
						}
						if val >= 16 && val <= 31 {
							addReverseDNS(z, r, &rDNS, fmt.Sprintf("%s.172.in-addr.arpa", parts[1]))
						} else {
							return nil, fmt.Errorf("%w: %s not within RFC1918 range", ErrUnsupportedIPv4, r.Content)
						}
					} else {
						return nil, fmt.Errorf("%w: %s", ErrUnsupportedIPv4, r.Content)
					}
				} else if r.Type == "AAAA" {
					parts := strings.Split(r.Content, ":")
					if parts[0] == "fdac" {
						addReverseDNS(z, r, &rDNS, "d.f.ip6.arpa")
					} else if strings.Split(parts[0], "")[0] == "2" {
						addReverseDNS(z, r, &rDNS, "2.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.ip6.arpa")
					}
				}
			}
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

	for arpaZone, rs := range rDNS {
		var rRS []Record
		for _, r := range rs {
			var addr string
			if r.Type == "A" {
				// Generate PTR address for IPv4
				addr = fmt.Sprintf("%s.in-addr.arpa.", reverseIPv4(r.Content))
			} else if r.Type == "AAAA" {
				// Generate PTR address for IPv6
				addr = fmt.Sprintf("%s.ip6.arpa.", reverseIPv6(r.Content))
			}

			rRS = append(rRS, Record{
				Type:    "PTR",
				Host:    addr,
				Content: r.Host,
				TTL:     r.TTL,
			})
		}

		ZS = append(ZS, Zone{
			Name:    arpaZone,
			Records: rRS,
			SOA: SOA{
				// PrimaryNS:  _bd.Configs.PrimaryNS,
				// AdminEmail: _bd.Configs.AdminEmail,
				// Refresh:    _bd.Configs.Refresh,
				// Retry:      _bd.Configs.Retry,
				// Expire:     _bd.Configs.Expire,
				// Minimum:    _bd.Configs.Minimum,
				// TTL:        _bd.Configs.TTL,

				// Hardcoded for now
				PrimaryNS:  "ns.arpa.leejacksonz.com",
				AdminEmail: "admin.leejacksonz.com",
				Serial:     uint64(time.Now().Unix()),
				Refresh:    1800,
				Retry:      1800,
				Expire:     604800,
				Minimum:    1800,
				TTL:        3600,
			},
		})
	}

	return ZS, nil
}

func reverseIPv4(s string) string {
	parts := strings.Split(s, ".")
	for i, j := 0, len(parts)-1; i < j; i, j = i+1, j-1 {
		parts[i], parts[j] = parts[j], parts[i]
	}
	return strings.Join(parts, ".")
}

func reverseIPv6(s string) string {
	ip := net.ParseIP(s)
	var ipString strings.Builder
	for _, octet := range ip {
		ipString.WriteString(fmt.Sprintf("%02x", octet))
	}
	runes := []rune(ipString.String())
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	// Join with dots
	return strings.Join(strings.Split(string(runes), ""), ".")
}

// renderNamedZones renders the named zones based on the provided Zone slice.
//
// Parameters:
// - zones: a slice of Zone objects containing the zones to be rendered.
// Returns:
// - string: the path of the created configuration file.
// - error: an error if any occurred during the rendering process.
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

// renderZone renders a Zone object into a configuration file.
//
// It parses a template file located at "templates/bind-zone.tmpl" relative to the current file.
// It removes all files in the "outputDir" directory except for the "outputDir" directory itself,
// if they start with the same prefix as the Zone's name.
// It creates a file with the name "<zone.Name>.conf" in the "outputDir" directory and writes the
// rendered template into it.
//
// Parameters:
// - zone: the Zone object to be rendered.
//
// Returns:
// - string: the path of the created configuration file.
// - error: an error if any occurred during the rendering process.
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

func RenderZonesTemplate() error {

	// Create all zones
	zs, err := createZones()
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
	err = (&rdb.Record{}).CommitAll()
	if err != nil {
		return err
	}

	return nil
}
