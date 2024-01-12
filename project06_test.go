package main

import (
	"reflect"
	"testing"
)

func TestImages(t *testing.T) {
	tests := []struct {
		term string
		want map[string]float64
	}{
		{
			term: "cynthia",
			want: map[string]float64{
				"About UC Santa Cruz – UC Santa Cruz": 0.14912280701754385,
				"Research – UC Santa Cruz":            0.12686567164179105,
			},
		},
		{
			term: "monterey",
			want: map[string]float64{
				"2023 in Review – UC Santa Cruz":      0.05902777777777777,
				"About UC Santa Cruz – UC Santa Cruz": 0.09941520467836255,
				"Campus Destinations – UC Santa Cruz": 0.22666666666666666,
			},
		},
	}
	db := databases()
	defer db.Close()

	url := "https://www.ucsc.edu/wp-sitemap.xml"
	body, _ := getPolicy(url)
	parseRobotsTxt(body)
	crawlSQL(db, url)
	for _, test := range tests {
		stem := stemWords(test.term)
		_, resMap := imageSearch(db, stem, url)
		if !reflect.DeepEqual(test.want, resMap) {
			t.Errorf("Expected: %v, Got: %v", test.want, resMap)
		}
	}
}
