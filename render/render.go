package render

import (
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

func CreateZones(_bd *rdb.BindData) ([]Zone, error) {
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
				PrimaryNS:  "ns.placeholder.",
				AdminEmail: "webmaster.",
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
