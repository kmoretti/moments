package handler

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/kmoretti/moments/vo"
	"github.com/labstack/echo/v4"
	"github.com/samber/do/v2"
)

type WeatherHandler struct {
	base BaseHandler
}

func NewWeatherHandler(injector do.Injector) *WeatherHandler {
	return &WeatherHandler{do.MustInvoke[BaseHandler](injector)}
}

type ipGeoResp struct {
	IP        string  `json:"ip"`
	City      string  `json:"city"`
	Country   string  `json:"country"`
	Flag      string  `json:"flag"`
	Region    string  `json:"region"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type weatherAPIResp struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		Location struct {
			Province  string `json:"province"`
			City      string `json:"city"`
			Town      string `json:"town"`
			Formatted string `json:"formatted"`
		} `json:"location"`
		Realtime struct {
			Weather       string  `json:"weather"`
			Temperature   float64 `json:"temperature"`
			AQI           any     `json:"aqi"`
			WindDirection string  `json:"wind_direction"`
			WindSpeed     string  `json:"wind_speed"`
		} `json:"realtime"`
	} `json:"data"`
}

// GetCurrentWeather godoc
//
//	@Tags		Weather
//	@Summary	根据访问者 IP 获取当前位置天气
//	@Accept		json
//	@Produce	json
//	@Success	200	{object}	vo.WeatherCurrentVO
//	@Router		/api/sysConfig/weather/current [post]
func (h WeatherHandler) GetCurrentWeather(c echo.Context) error {
	client := &http.Client{Timeout: 8 * time.Second}
	ip, err := getVisitorIP(c)
	if err != nil {
		h.base.log.Warn().Err(err).Msg("resolve visitor ip failed")
		return FailRespWithMsg(c, Fail, "无法获取客户端IP")
	}

	geo, err := locateByIP(client, ip)
	if err != nil {
		h.base.log.Warn().Err(err).Msg("locate ip failed")
		return FailRespWithMsg(c, Fail, "获取位置信息失败")
	}

	weather, err := fetchWeather(client, geo)
	if err != nil {
		h.base.log.Warn().Err(err).Msg("fetch weather failed")
		return SuccessResp(c, vo.WeatherCurrentVO{
			Location:    pickLocation("", localizeGeoName(geo.City), localizeGeoName(geo.Region), localizeGeoName(geo.Country)),
			Weather:     "天气获取失败",
			Temperature: 0,
			AirQuality:  "--",
			Wind:        "",
		})
	}

	return SuccessResp(c, vo.WeatherCurrentVO{
		Location:    pickLocation(weather.Data.Location.Formatted, localizeGeoName(geo.City), localizeGeoName(geo.Region), localizeGeoName(geo.Country)),
		Weather:     weather.Data.Realtime.Weather,
		Temperature: weather.Data.Realtime.Temperature,
		AirQuality:  formatAQI(weather.Data.Realtime.AQI),
		Wind:        strings.TrimSpace(strings.Join([]string{weather.Data.Realtime.WindDirection, weather.Data.Realtime.WindSpeed}, " ")),
	})
}

func getVisitorIP(c echo.Context) (string, error) {
	candidates := []string{
		c.Request().Header.Get("x-client-ip"),
		c.Request().Header.Get("CF-Connecting-IP"),
		c.Request().Header.Get("X-Forwarded-For"),
		c.RealIP(),
	}
	for _, raw := range candidates {
		if raw == "" {
			continue
		}
		parts := strings.Split(raw, ",")
		ip := strings.TrimSpace(parts[0])
		parsed := net.ParseIP(ip)
		if parsed == nil {
			continue
		}
		if parsed.IsLoopback() || parsed.IsPrivate() {
			continue
		}
		return ip, nil
	}
	return "", fmt.Errorf("no public ip found")
}

func locateByIP(client *http.Client, ip string) (*ipGeoResp, error) {
	resp, err := client.Get("https://ip.081531.xyz/geo")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("ip geo api status %d", resp.StatusCode)
	}
	var data ipGeoResp
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}
	if data.IP == "" {
		data.IP = ip
	}
	return &data, nil
}

func fetchWeather(client *http.Client, geo *ipGeoResp) (*weatherAPIResp, error) {
	query := url.Values{}
	query.Set("query", pickQuery(geo))
	resp, err := client.Get("https://60s.081531.xyz/v2/weather?" + query.Encode())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("weather api status %d", resp.StatusCode)
	}
	var data weatherAPIResp
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}
	if data.Code != 200 {
		return nil, fmt.Errorf("weather api code %d", data.Code)
	}
	return &data, nil
}

func pickQuery(geo *ipGeoResp) string {
	if geo.City != "" {
		return geo.City
	}
	if geo.Region != "" {
		return geo.Region
	}
	return geo.Country
}

func pickLocation(preferred string, fallback ...string) string {
	if strings.TrimSpace(preferred) != "" {
		return strings.TrimSpace(preferred)
	}
	parts := make([]string, 0, len(fallback))
	for _, item := range fallback {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		parts = append(parts, item)
	}
	return strings.Join(parts, " ")
}

func formatAQI(aqi any) string {
	if aqi == nil {
		return "--"
	}
	return fmt.Sprintf("AQI %v", aqi)
}

func localizeGeoName(name string) string {
	trimmed := strings.TrimSpace(name)
	if trimmed == "" {
		return ""
	}
	mapping := map[string]string{
		"Singapore": "新加坡",
		"Hong Kong": "香港",
		"Tokyo": "东京",
		"Seoul": "首尔",
		"Osaka": "大阪",
		"Taipei": "台北",
		"California": "加利福尼亚",
		"Los Angeles": "洛杉矶",
		"New York": "纽约",
	}
	if localized, ok := mapping[trimmed]; ok {
		return localized
	}
	return trimmed
}
