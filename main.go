package main

import (
	"encoding/json"
	"net/http"
	"strings"
)

const (
	//API KEY http://maps.openweathermap.org/maps/2.0/weather
	openWeatherAPIKey = "ab6f47646b7415825239a8b7713cdc66"
	//API KEY FOR https://api.darksky.net/forecast/30799544d696520c3e1958bb860e26e7/37.8267,-122.4233
	darkSkyAPIKey = "30799544d696520c3e1958bb860e26e7"
	//OPEN WEATHER APP URL
	openWeatherURL = "http://api.openweathermap.org/data/2.5/weather?"
	//DARK SKY URL
	darkSkyURL = "https://api.darksky.net/forecast/"
)

type tempResponse struct {
	Name     string      `json:"city"`
	TimeZone string      `json:"timezone"`
	Weather  WeatherAttr `json:"wether"`
}

//WeatherAttr struct for different attributes of weather
type WeatherAttr struct {
	Summary  string      `json:"summary"`
	Humidity float64     `json:"humidity"`
	Pressure float64     `json:"pressure"`
	Temp     Temperature `json:"temperature"`
	Wind     WindAttr    `json:"wind"`
}

//Temperature struct for temperature in different units
type Temperature struct {
	Celcius   float64 `json:"celcius"`
	Farenheit float64 `json:"farenheit"`
	Kelvin    float64 `json:"kelvin"`
}

//WindAttr struct for different wind attributes
type WindAttr struct {
	Speed   float64 `json:"speed"`
	Gust    float64 `json:"gust"`
	Bearing float64 `json:"bearing"`
}

type openWeatherData struct {
	Name string `json:"name"`
	Main struct {
		Kelvin float64 `json:"temp"`
	} `json:"main"`
}

type darkWeatherData struct {
	Name      string `json:"timezone"`
	Currently struct {
		Summary     string  `json:"summary"`
		Celsius     float64 `json:"temperature"`
		Humidity    float64 `json:"humidity"`
		Pressure    float64 `json:"pressure"`
		WindSpeed   float64 `json:"windSpeed"`
		WindGust    float64 `json:"windGust"`
		WindBearing float64 `json:"windBearing"`
	} `json:"currently"`
}

func main() {
	http.HandleFunc("/", welcome)
	http.HandleFunc("/weather/", response)
	http.ListenAndServe(":8000", nil)
}

func welcome(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Welcome! \nGo to http://localhost:8000/weather/%City% e.g. Nairobi to find out the temperature\ne.g. /Nairobi"))
}

func response(w http.ResponseWriter, r *http.Request) {
	city := strings.SplitN(r.URL.Path, "/", 3)[2]
	data, error := query(city)
	if error != nil {
		http.Error(w, error.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset = utf-8")
	json.NewEncoder(w).Encode(data)
}

func query(city string) (tempResponse, error) {

	weatherAppResp, error1 := http.Get(openWeatherURL + "APPID=" + openWeatherAPIKey + "&q=" + city)
	darkSkyResp, error2 := http.Get(darkSkyURL + darkSkyAPIKey + "/-1.3032051,36.7073103")

	if error1 != nil {
		return tempResponse{}, error1
	} else if error2 != nil {
		return tempResponse{}, error2
	}

	defer weatherAppResp.Body.Close()
	defer darkSkyResp.Body.Close()

	var openData openWeatherData
	var darkData darkWeatherData

	if jsonError1 := json.NewDecoder(weatherAppResp.Body).Decode(&openData); jsonError1 != nil {
		return tempResponse{}, jsonError1
	} else if jsonError2 := json.NewDecoder(darkSkyResp.Body).Decode(&darkData); jsonError2 != nil {
		return tempResponse{}, jsonError1
	}

	tempCelcius := averageTempCelcius(openData.Main.Kelvin, darkData.Currently.Celsius)

	Temperature := Temperature{
		tempCelcius,
		convertTemp(tempCelcius, "celcius", "farenheit"),
		convertTemp(tempCelcius, "celcius", "kelvin"),
	}

	Wind := WindAttr{
		darkData.Currently.WindSpeed,
		darkData.Currently.WindGust,
		darkData.Currently.WindBearing,
	}

	Weather := WeatherAttr{darkData.Currently.Summary,
		darkData.Currently.Humidity,
		darkData.Currently.Pressure,
		Temperature, Wind,
	}

	return tempResponse{openData.Name, darkData.Name, Weather}, nil
}

func averageTempCelcius(temp1, temp2 float64) float64 {
	return ((temp1 - 273.0) + ((temp2 - 32.0) * (5.0 / 9.0))) / 2.0
}

func convertTemp(temp float64, from string, to string) float64 {
	if from == "celcius" && to == "farenheit" {
		return (temp * (9.0 / 5.0)) + 32.0
	} else if from == "celcius" && to == "kelvin" {
		return temp + 273.0
	}
	return 0.0
}
