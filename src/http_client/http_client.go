package http_client

import (
	"bytes"
	"net"
	"net/http"
	"strings"
	"time"
)

const (
	jsonContent      = "application/json"
	xmlContent       = "application/xml"
	textPlain        = "text/plain"
	GET              = "GET"
	POST             = "POST"
	XmlContent       = "xml"
	JsonContent      = "json"
	TextPlainContent = "text"
)

var tr http.Transport
var cliente http.Client

var contenido map[string]string = map[string]string{
	"json": jsonContent,
	"xml":  xmlContent,
	"text": textPlain,
}

type Request struct {
	Mensaje        []byte
	Url            string
	Metodo         string
	Contenido      string
	Reintentos     int
	MaxReintentos  int
	TiempoDeEspera int
}

func init() {
	tr = http.Transport{
		MaxConnsPerHost:     250,
		MaxIdleConnsPerHost: 50,
		MaxIdleConns:        10,
		IdleConnTimeout:     10 * time.Second,
		DisableCompression:  true,
		Proxy:               nil,
	}
	cliente = http.Client{
		Transport: &tr,
		Timeout:   600 * time.Second,
	}
}
func NewRequest(mensaje []byte, url string, method string, content string) *Request {
	return &Request{
		Mensaje:        mensaje,
		Url:            url,
		Metodo:         strings.TrimSpace(strings.ToUpper(method)),
		Contenido:      contenido[content],
		Reintentos:     1,
		MaxReintentos:  2,
		TiempoDeEspera: 1,
	}
}
func SendRequest(request *Request) (*http.Response, error) {
	req, err := http.NewRequest(request.Metodo, request.Url, bytes.NewBuffer(request.Mensaje))
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-type", request.Contenido)
	req.Close = true
	response, err := cliente.Do(req)
	if err != nil {
		if timeOutErr, ok := err.(net.Error); ok && timeOutErr.Timeout() {
			return &http.Response{StatusCode: 599}, err
		}
		return response, err
	}
	return response, err
}

func SendRequestWithBasicAuth(request *Request, username string, password string) (*http.Response, error) {

	req, err := http.NewRequest(request.Metodo, request.Url, bytes.NewBuffer(request.Mensaje))
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(username, password)
	req.Header.Add("Content-type", request.Contenido)
	req.Close = true

	response, err := cliente.Do(req)

	if err != nil {
		if timeOutErr, ok := err.(net.Error); ok && timeOutErr.Timeout() {
			return &http.Response{StatusCode: 599}, err
		}
		return response, err
	}

	return response, err

}
