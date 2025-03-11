package spider

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"time"
)

// Index represents different weather indices
type Index string

const (
	ChuanYiIndex   Index = "穿衣"
	WuranIndex     Index = "污染"
	ChuyouIndex    Index = "出游"
	XicheIndex     Index = "洗车"
	ShaiBeiZiIndex Index = "晒被子"
)

// IndexInfo contains status and explanation for a weather index
type IndexInfo struct {
	Status string
	Why    string
}

// Weather represents weather information for a specific day
type Weather struct {
	Date          time.Time
	IndexMap      map[Index]IndexInfo
	MinWeather    int
	MaxWeather    int
	Condition     string
	WindDirection string
	WindPower     string
	Humidity      string
	Date8         string // Date in format YYYY-MM-DD
	WeekDay       string
}

// GetWeatherDataOri fetches raw weather data from Baidu
func GetWeatherDataOri(province, city string) (string, error) {
	pc := url.QueryEscape(province + city + "天气")
	url1 := fmt.Sprintf("https://weathernew.pae.baidu.com/weathernew/pc?query=%s&srcid=4982&forecast=long_day_forecast", pc)
	
	req, err := http.NewRequest("GET", url1, nil)
	if err != nil {
		return "", fmt.Errorf("error creating request: %v", err)
	}

	req.Header.Set("DNT", "1")
	req.Header.Set("Referer", "https://www.baidu.com/link?url=LOoW7nfxfB2350fjBnQho9KK8Q8Ohrk3zjDnkt5-ji2dYVikhoZM0eMLAh4n9zX9JVGtbVCjEWTkgvmPficS0lutwN8YMcIveqCrmGMqUwHSQ7gheKSPqJa3LUg9_6OV3Qe9jEyVbPedGbd9sfZhn3Pa41CWbxXZCfPkOePFfq1xroUjxSIr0DBtAEjutPQSB0QXgbifwjJl7mkWQ5ZS_a&wd=&eqid=f3b700cf000029540000000465f3e3c1")
	req.Header.Set("Upgrade-Insecure-Requests", "1")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/122.0.0.0 Safari/537.36")
	req.Header.Set("sec-ch-ua", `"Chromium";v="122", "Not(A:Brand";v="24", "Google Chrome";v="122"`)
	req.Header.Set("sec-ch-ua-mobile", "?0")
	req.Header.Set("sec-ch-ua-platform", `"Windows"`)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("error sending request: %v", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error reading response body: %v", err)
	}

	return string(body), nil
}


// Get15DaysWeather parses the raw weather data and returns structured weather information for 15 days
func Get15DaysWeather(s string) ([]Weather, error) {
    // Regular expressions for extracting data
    zhiShuPattern := `data\["zhishu"\] *= *(\{.*\})`
    weather15DayTempPattern := `longDayForecast":{"info":(.*?)]`
    
    // Compile regex patterns
    zhiShuRe := regexp.MustCompile(zhiShuPattern)
    weather15DayTempRe := regexp.MustCompile(weather15DayTempPattern)
    
    // Find matches
    zhiShuMatch := zhiShuRe.FindStringSubmatch(s)
    weather15DayTempMatch := weather15DayTempRe.FindStringSubmatch(s)

    // Check if we have the temperature data
    if len(weather15DayTempMatch) < 2 {
        return nil, fmt.Errorf("error parsing origin string: missing required temperature data")
    }

    // Parse temperature data
    var dayTemList []DayTemperature
    // Complete the JSON array
    jsonStr := weather15DayTempMatch[1] + "]"
    
    if err := json.Unmarshal([]byte(jsonStr), &dayTemList); err != nil {
        return nil, fmt.Errorf("error parsing temperature data: %v", err)
    }

    // Parse life indices (only for today)
    var lifeIndex LifeIndex
    todayIndices := make(map[Index]IndexInfo)
    
    if len(zhiShuMatch) >= 2 {
        if err := json.Unmarshal([]byte(zhiShuMatch[1]), &lifeIndex); err != nil {
            fmt.Printf("Warning: error parsing zhishu json: %v\n", err)
        } else {
            for _, item := range lifeIndex.Item {
                todayIndices[Index(item.ItemName)] = IndexInfo{
                    Status: item.ItemTitle,
                    Why:    item.ItemDesc,
                }
            }
        }
    }

    // Create weather objects for each day
    var weatherList []Weather
    todayStr := time.Now().Format("2006-01-02")
    
    for _, day := range dayTemList {
        weather := Weather{
            Date8:     day.Date,
            IndexMap:  make(map[Index]IndexInfo),
        }

        // Parse date
        if parsedDate, err := time.Parse("2006-01-02", day.Date); err == nil {
            weather.Date = parsedDate
            weather.WeekDay = parsedDate.Weekday().String()
        }

        // Set condition from temperature data
        if day.WeatherDay != "" {
            weather.Condition = day.WeatherDay
        } else if day.WeatherNight != "" {
            weather.Condition = day.WeatherNight
        }

        // Set wind information
        if day.WindDirectionDay != "" {
            weather.WindDirection = day.WindDirectionDay
        } else if day.WindDirectionNight != "" {
            weather.WindDirection = day.WindDirectionNight
        }
        
        if day.WindPowerDay != "" {
            weather.WindPower = day.WindPowerDay
        } else if day.WindPowerNight != "" {
            weather.WindPower = day.WindPowerNight
        }

        // Add humidity if available
        if day.Humidity != nil && day.Humidity.Text != "" {
            weather.Humidity = day.Humidity.Text
        }

        // Extract temperature
        var minTemp, maxTemp int
        if _, err := fmt.Sscanf(day.TemperatureNight, "%d", &minTemp); err == nil {
            weather.MinWeather = minTemp
        }
        if _, err := fmt.Sscanf(day.TemperatureDay, "%d", &maxTemp); err == nil {
            weather.MaxWeather = maxTemp
        }

        // For today, add indices
        if day.Date == todayStr {
            weather.IndexMap = todayIndices
        }

        weatherList = append(weatherList, weather)
    }

    // Check if we have any data
    if len(weatherList) == 0 {
        return nil, fmt.Errorf("no weather data found")
    }

    return weatherList, nil
}


// GetTodayWeather is a convenience function that returns only today's weather
func GetTodayWeather(s string) (Weather, error) {
    weatherList, err := Get15DaysWeather(s)
    if err != nil {
        return Weather{}, err
    }
    
    todayStr := time.Now().Format("2006-01-02")
    for _, weather := range weatherList {
        if weather.Date8 == todayStr {
            return weather, nil
        }
    }
    
    // If we couldn't find today's weather specifically, return the first one
    if len(weatherList) > 0 {
        return weatherList[0], nil
    }
    
    return Weather{}, fmt.Errorf("today's weather data not found")
}


// Day15Weather represents daily weather information
type Day15Weather struct {
	FormatDate  string `json:"formatDate"`
	Date        string `json:"date"`
	FormatWeek  string `json:"formatWeek"`
	WeatherIcon string `json:"weatherIcon"`
	WeatherWind struct {
		WindDirectionDay   string `json:"windDirectionDay"`
		WindDirectionNight string `json:"windDirectionNight"`
		WindPowerDay       string `json:"windPowerDay"`
		WindPowerNight     string `json:"windPowerNight"`
	} `json:"weatherWind"`
	WeatherPm25 string `json:"weatherPm25"`
	WeatherText string `json:"weatherText"`
}

// LifeIndex represents various life indices related to weather
type LifeIndex struct {
	Url      string `json:"url"`
	Title    string `json:"title"`
	Desc     string `json:"desc"`
	OtherUrl string `json:"other_url"`
	Item     []struct {
		ItemName      string `json:"item_name"`
		ItemTitle     string `json:"item_title"`
		ItemIcon      string `json:"item_icon"`
		ItemIconWhite string `json:"item_icon_white"`
		ItemDesc      string `json:"item_desc"`
		ItemUrl       string `json:"item_url"`
		ItemOtherUrl  string `json:"item_other_url"`
	} `json:"item"`
	StrategyLog struct {
		RecommendZhishuSort []string      `json:"recommend_zhishu_sort"`
		UserAttr            []interface{} `json:"user_attr"`
		ObserveWeather      struct {
			BodytempInfo      string `json:"bodytemp_info"`
			WindDirection     string `json:"wind_direction"`
			Site              string `json:"site"`
			Weather           string `json:"weather"`
			DewTemperature    string `json:"dew_temperature"`
			PrecipitationType string `json:"precipitation_type"`
			WindDirectionNum  string `json:"wind_direction_num"`
			Temperature       string `json:"temperature"`
			WindPower         string `json:"wind_power"`
			F1HInfo           []struct {
				PrecipitationProbability string `json:"precipitation_probability"`
				Temperature              string `json:"temperature"`
				Hour                     string `json:"hour"`
				WindDirection            string `json:"wind_direction"`
				Uv                       string `json:"uv"`
				UvNum                    string `json:"uv_num"`
				WindPower                string `json:"wind_power"`
				Weather                  string `json:"weather"`
				WindPowerNum             string `json:"wind_power_num"`
				Precipitation            string `json:"precipitation"`
			} `json:"f1hInfo"`
			UpdateTime          string `json:"update_time"`
			PublishTime         string `json:"publish_time"`
			Visibility          string `json:"visibility"`
			Pressure            string `json:"pressure"`
			PrecMonitorTime     string `json:"prec_monitor_time"`
			Precipitation       string `json:"precipitation"`
			RealFeelTemperature string `json:"real_feel_temperature"`
			UvInfo              string `json:"uv_info"`
			Uv                  string `json:"uv"`
			Humidity            string `json:"humidity"`
			UvNum               string `json:"uv_num"`
			WindPowerNum        string `json:"wind_power_num"`
			F1HInfoNumBaidu     int    `json:"f1hInfo#num#baidu"`
			PsPm25              string `json:"ps_pm25"`
		} `json:"observe_weather"`
	} `json:"strategy_log"`
}

// DayTemperature represents temperature data for a specific day
type DayTemperature struct {
	Date                         string      `json:"date"`
	WeatherDay                   string      `json:"weather_day"`
	WeatherNight                 string      `json:"weather_night"`
	TemperatureDay               string      `json:"temperature_day"`
	TemperatureNight             string      `json:"temperature_night"`
	MoonPhase                    string      `json:"moon_phase,omitempty"`
	NextNewMoon                  string      `json:"next_new_moon,omitempty"`
	WindDirectionNight           string      `json:"wind_direction_night,omitempty"`
	Moonrise                     string      `json:"moonrise,omitempty"`
	WindDirectionDay             string      `json:"wind_direction_day,omitempty"`
	MoonPicNum                   string      `json:"moon_pic_num,omitempty"`
	Sunrisetime                  string      `json:"sunrisetime,omitempty"`
	WindPowerDay                 string      `json:"wind_power_day,omitempty"`
	WindPowerNight               string      `json:"wind_power_night,omitempty"`
	WeatherNightForBeijing       string      `json:"weather_night_for_beijing,omitempty"`
	WeatherDayForBeijing         string      `json:"weather_day_for_beijing,omitempty"`
	PrecipitationProbabilityNight string      `json:"precipitation_probability_night,omitempty"`
	Sunsettime                   string      `json:"sunsettime,omitempty"`
	Moonset                      string      `json:"moonset,omitempty"`
	PrecipitationProbabilityDay  string      `json:"precipitation_probability_day,omitempty"`
	PM25                         *PM25Info   `json:"pm25,omitempty"`
	Limitline                    *LimitInfo  `json:"limitline,omitempty"`
	Humidity                     *HumidityInfo `json:"humidity,omitempty"`
}

// PM25Info represents PM2.5 pollution information
type PM25Info struct {
	Listquality *QualityInfo `json:"listquality"`
	Listtitle   string       `json:"listtitle"`
}

// QualityInfo represents air quality information
type QualityInfo struct {
	Listkey    string `json:"listkey"`
	Listvalue  string `json:"listvalue"`
	Listaqival string `json:"listaqival"`
	Site       string `json:"site"`
}

// LimitInfo represents vehicle limitation information
type LimitInfo struct {
	Tip  string `json:"tip"`
	Text string `json:"text"`
}

// HumidityInfo represents humidity information
type HumidityInfo struct {
	Tip  string `json:"tip"`
	Text string `json:"text"`
}
