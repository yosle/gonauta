package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

const (
	BaseURL           = "https://secure.etecsa.net:8443"
	HourRate          = 12.5
	NationalHourRate  = 2.5
	MaxTimeoutSeconds = 30
	IPCheckURL        = "http://ip-api.com/json/"
)

// Time representa el tiempo restante
type Time struct {
	Hours   int `json:"hours"`
	Minutes int `json:"minutes"`
	Seconds int `json:"seconds"`
}

// UserInfo contiene la información del usuario
type UserInfo struct {
	Status         string  `json:"status"`
	Credits        float64 `json:"credits"`
	ExpirationDate string  `json:"expiration_date"`
	AccessInfo     string  `json:"access_info"`
	RemainingTime  Time    `json:"remaining_time"`
}

// SessionData contiene los datos de la sesión
type SessionData struct {
	Username string `json:"username"`
	UUID     string `json:"uuid"`
}

// IPInfo contiene la información de geolocalización IP
type IPInfo struct {
	Status      string  `json:"status"`
	Country     string  `json:"country"`
	CountryCode string  `json:"countryCode"`
	Region      string  `json:"region"`
	RegionName  string  `json:"regionName"`
	City        string  `json:"city"`
	Zip         string  `json:"zip"`
	Lat         float64 `json:"lat"`
	Lon         float64 `json:"lon"`
	Timezone    string  `json:"timezone"`
	ISP         string  `json:"isp"`
	Org         string  `json:"org"`
	AS          string  `json:"as"`
	Query       string  `json:"query"`
}

// Client maneja las operaciones de Nauta
type Client struct {
	httpClient *http.Client
	cookieJar  *cookiejar.Jar
}

// NewClient crea una nueva instancia del cliente Nauta
func NewClient() (*Client, error) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, err
	}

	return &Client{
		httpClient: &http.Client{
			Jar:     jar,
			Timeout: MaxTimeoutSeconds * time.Second,
		},
		cookieJar: jar,
	}, nil
}

// checkConnection verifica la conectividad y detecta el uso de VPN
func checkConnection() (*IPInfo, error) {
	resp, err := http.Get(IPCheckURL)
	if err != nil {
		return nil, fmt.Errorf("no hay conexión a internet: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error leyendo respuesta: %w", err)
	}

	var ipInfo IPInfo
	if err := json.Unmarshal(body, &ipInfo); err != nil {
		return nil, fmt.Errorf("error parseando respuesta: %w", err)
	}

	if ipInfo.Status != "success" {
		return nil, errors.New("no se pudo verificar la ubicación")
	}

	return &ipInfo, nil
}

// getLoginParams extrae los parámetros ocultos del formulario de login
func (c *Client) getLoginParams(body string) (map[string]string, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(body))
	if err != nil {
		return nil, err
	}

	params := make(map[string]string)
	doc.Find("#formulario input[type='hidden']").Each(func(i int, s *goquery.Selection) {
		name, _ := s.Attr("name")
		value, _ := s.Attr("value")
		params[name] = value
	})

	return params, nil
}

// extractUUID extrae el UUID de la respuesta del login
func extractUUID(body string) (string, error) {
	re := regexp.MustCompile(`ATTRIBUTE_UUID=(\w*)&`)
	matches := re.FindStringSubmatch(body)
	if len(matches) < 2 {
		return "", errors.New("UUID no encontrado en la respuesta")
	}
	return matches[1], nil
}

// Login inicia sesión en Nauta
func (c *Client) Login(username, password string) (*SessionData, error) {
	// Verificar conectividad
	ipInfo, err := checkConnection()
	if err != nil {
		return nil, err
	}

	// Verificar si está conectado desde Cuba (posible VPN)
	if ipInfo.CountryCode != "CU" {
		fmt.Printf("\n⚠️  Conectado a través de VPN\n")
		fmt.Printf("País: %s\n", ipInfo.Country)
		fmt.Printf("ISP: %s\n\n", ipInfo.ISP)
	}

	// Obtener la página inicial
	resp, err := c.httpClient.Get(BaseURL)
	if err != nil {
		return nil, fmt.Errorf("error de conexión: %w. Comprueba que estás conectado a una WiFi de ETECSA", err)
	}
	defer resp.Body.Close()

	// Leer el body
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, err
	}
	body, _ := doc.Html()

	// Obtener parámetros del login
	loginParams, err := c.getLoginParams(body)
	if err != nil {
		return nil, err
	}

	// Preparar datos del formulario
	formData := url.Values{}
	for key, value := range loginParams {
		formData.Set(key, value)
	}
	formData.Set("username", username)
	formData.Set("password", password)

	// Hacer login
	resp, err = c.httpClient.PostForm(BaseURL+"/LoginServlet", formData)
	if err != nil {
		return nil, fmt.Errorf("error de conexión: %w", err)
	}
	defer resp.Body.Close()

	// Leer respuesta
	doc, err = goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, err
	}
	responseBody, _ := doc.Html()

	// Validar errores
	if strings.Contains(responseBody, "El nombre de usuario o contraseña son incorrectos.") {
		return nil, errors.New("el nombre de usuario o contraseña son incorrectos")
	}
	if strings.Contains(responseBody, "Su tarjeta no tiene saldo disponible") {
		return nil, fmt.Errorf("su cuenta %s no tiene saldo disponible", username)
	}
	if strings.Contains(responseBody, "El usuario ya está conectado.") {
		return nil, errors.New("su cuenta está siendo usada")
	}
	if strings.Contains(responseBody, "No se pudo autorizar al usuario") {
		return nil, errors.New("no se pudo autorizar al usuario")
	}

	// Extraer UUID
	uuid, err := extractUUID(responseBody)
	if err != nil {
		return nil, fmt.Errorf("no se ha podido obtener los datos de la sesión: %w", err)
	}

	return &SessionData{
		Username: username,
		UUID:     uuid,
	}, nil
}

// extractUserInfo extrae la información del usuario del HTML
func extractUserInfo(body string) (*UserInfo, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(body))
	if err != nil {
		return nil, err
	}

	statusText := strings.TrimSpace(doc.Find("#sessioninfo tr:nth-child(1) td:nth-child(2)").Text())
	creditsText := strings.TrimSpace(doc.Find("#sessioninfo tr:nth-child(2) td:nth-child(2)").Text())
	expirationDateText := strings.TrimSpace(doc.Find("#sessioninfo tr:nth-child(3) td:nth-child(2)").Text())
	accessInfoText := strings.TrimSpace(doc.Find("#sessioninfo tr:nth-child(4) td:nth-child(2)").Text())

	status := "Disabled"
	if statusText == "Activa" {
		status = "Active"
	}

	creditsStr := strings.Replace(creditsText, " CUP", "", 1)
	credits, err := strconv.ParseFloat(creditsStr, 64)
	if err != nil {
		return nil, fmt.Errorf("error parseando créditos: %w", err)
	}

	expirationDate := expirationDateText
	if expirationDateText == "No especificada" {
		expirationDate = "None"
	}

	accessInfo := accessInfoText
	if accessInfoText == "Acceso desde todas las áreas de Internet" {
		accessInfo = "All"
	}

	return &UserInfo{
		Status:         status,
		Credits:        credits,
		ExpirationDate: expirationDate,
		AccessInfo:     accessInfo,
	}, nil
}

// calculateRemainingTime calcula el tiempo restante basado en créditos
func calculateRemainingTime(credits, rate float64) Time {
	totalHours := credits / rate
	hours := int(totalHours)

	remainingFraction := totalHours - float64(hours)
	totalMinutes := remainingFraction * 60
	minutes := int(totalMinutes)

	remainingSeconds := totalMinutes - float64(minutes)
	seconds := int(remainingSeconds * 60)

	return Time{
		Hours:   hours,
		Minutes: minutes,
		Seconds: seconds,
	}
}

// GetUserInfo obtiene la información del usuario
func (c *Client) GetUserInfo(username, password string) (*UserInfo, error) {
	// Obtener la página inicial
	resp, err := c.httpClient.Get(BaseURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, err
	}
	body, _ := doc.Html()

	// Obtener parámetros del login
	loginParams, err := c.getLoginParams(body)
	if err != nil {
		return nil, err
	}

	// Preparar datos del formulario
	formData := url.Values{}
	for key, value := range loginParams {
		formData.Set(key, value)
	}
	formData.Set("username", username)
	formData.Set("password", password)

	// Consultar información del usuario
	resp, err = c.httpClient.PostForm(BaseURL+"/EtecsaQueryServlet", formData)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	doc, err = goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, err
	}
	responseBody, _ := doc.Html()

	userInfo, err := extractUserInfo(responseBody)
	if err != nil {
		return nil, err
	}

	// Determinar la tarifa según el tipo de cuenta
	rate := HourRate
	if strings.Contains(username, "@nauta.co.cu") {
		rate = NationalHourRate
	}

	userInfo.RemainingTime = calculateRemainingTime(userInfo.Credits, rate)

	return userInfo, nil
}

// Session representa una sesión activa de Nauta
type Session struct {
	Data   SessionData
	client *Client
}

// NewSession crea una nueva sesión
func NewSession(data SessionData, client *Client) *Session {
	return &Session{
		Data:   data,
		client: client,
	}
}

// parseTime parsea una cadena de tiempo en formato HH:MM:SS
func parseTime(value string) (Time, error) {
	re := regexp.MustCompile(`(\d+):([\d]{2}):([\d]{2})`)
	matches := re.FindStringSubmatch(value)
	if len(matches) < 4 {
		return Time{}, errors.New("formato de tiempo inválido")
	}

	hours, _ := strconv.Atoi(matches[1])
	minutes, _ := strconv.Atoi(matches[2])
	seconds, _ := strconv.Atoi(matches[3])

	return Time{
		Hours:   hours,
		Minutes: minutes,
		Seconds: seconds,
	}, nil
}

// GetRemainingTime obtiene el tiempo restante de la sesión
func (s *Session) GetRemainingTime() (*Time, error) {
	if s.Data.UUID == "" || s.Data.Username == "" {
		return nil, fmt.Errorf("sesión inválida: %+v", s.Data)
	}

	// Verificar conectividad
	ipInfo, err := checkConnection()
	if err != nil {
		return nil, err
	}

	// Bloquear si está conectado desde fuera de Cuba (VPN)
	if ipInfo.CountryCode != "CU" {
		return nil, fmt.Errorf("No se puede obtener el estado de la sesión cuando está conectado a través de VPN (País: %s, ISP: %s)", ipInfo.Country, ipInfo.ISP)
	}

	formData := url.Values{}
	formData.Set("op", "getLeftTime")
	formData.Set("ATTRIBUTE_UUID", s.Data.UUID)
	formData.Set("username", s.Data.Username)

	resp, err := s.client.httpClient.PostForm(BaseURL+"/EtecsaQueryServlet", formData)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, err
	}
	body := doc.Text()

	remainingTime, err := parseTime(body)
	if err != nil {
		return nil, err
	}

	return &remainingTime, nil
}

// Logout cierra la sesión
func (s *Session) Logout() error {
	// Verificar conectividad
	ipInfo, err := checkConnection()
	if err != nil {
		return err
	}

	// Verificar si está conectado desde Cuba (posible VPN)
	if ipInfo.CountryCode != "CU" {
		fmt.Printf("\n⚠️  Conectado a través de VPN\n")
		fmt.Printf("País: %s\n", ipInfo.Country)
		fmt.Printf("ISP: %s\n\n", ipInfo.ISP)
	}

	formData := url.Values{}
	formData.Set("ATTRIBUTE_UUID", s.Data.UUID)
	formData.Set("username", s.Data.Username)
	formData.Set("remove", "1")

	resp, err := s.client.httpClient.PostForm(BaseURL+"/LogoutServlet", formData)
	if err != nil {
		return fmt.Errorf("error al cerrar sesión: %w", err)
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return err
	}
	body := doc.Text()

	if strings.Contains(body, "logoutcallback('SUCCESS')") {
		return nil
	}

	return fmt.Errorf("fallo al cerrar sesión: %s", body)
}
