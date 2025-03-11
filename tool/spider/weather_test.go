package spider

import (
	"testing"

	"github.com/intmian/mian_go_lib/tool/misc"
)

func TestGetWeather(t *testing.T) {
	s, err := GetWeatherDataOri("北京", "")
	if err != nil {
		t.Fatal(err)
	}
	s = misc.ReplaceUnicodeEscapes(s)
	weather, err := Get15DaysWeather(s)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(weather)
}
