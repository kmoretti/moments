package handler

import (
	"io"
	"net/http"
	"net/url"

	"github.com/labstack/echo/v4"
	"github.com/samber/do/v2"
)

type WeatherHandler struct {
	base BaseHandler
}

func NewWeatherHandler(injector do.Injector) *WeatherHandler {
	return &WeatherHandler{do.MustInvoke[BaseHandler](injector)}
}

func (h *WeatherHandler) GetWeather(c echo.Context) error {
	query := c.QueryParam("query")
	if query == "" {
		query = "北京"
	}

	apiUrl := "https://60s.081531.xyz/v2/weather?" + url.Values{"query": {query}}.Encode()
	resp, err := http.Get(apiUrl)
	if err != nil {
		return FailRespWithMsg(c, Fail, "获取天气数据失败")
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return FailRespWithMsg(c, Fail, "读取天气数据失败")
	}

	return c.JSONBlob(resp.StatusCode, body)
}
