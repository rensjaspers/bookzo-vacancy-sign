package bookzo

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"
)

const endpoint = "https://bookzoapi.nl/api/json/reply/GetPossibleStays"
const fallbackOrigin = "https://www.hotelrasch.nl"
const clientName = "Elements"
const userAgent = "Mozilla/5.0 (Hotel Rasch Vacancy Board)"

type Client struct {
	apiKey string
	client *http.Client
	logger *slog.Logger
}

type possibleStaysResponse struct {
	Result []possibleStay `json:"Result"`
}

type possibleStay struct {
	Arrival string `json:"Arrival"`
}

type requestBody struct {
	AccommodationsToBook []occupancy `json:"AccommodationsToBook"`
	From                 string      `json:"From"`
	Until                string      `json:"Until"`
	ObjectIDs            []int       `json:"ObjectIds"`
}

type occupancy struct {
	NumberOfAdults  int `json:"NumberOfAdults"`
	NumberOfChilds  int `json:"NumberOfChilds"`
	NumberOfYouths  int `json:"NumberOfYouths"`
	NumberOfBabies  int `json:"NumberOfBabies"`
	NumberOfSeniors int `json:"NumberOfSeniors"`
	NumberOfPets    int `json:"NumberOfPets"`
}

func New(apiKey string, logger *slog.Logger) *Client {
	return &Client{apiKey: apiKey, client: newHTTPClient(), logger: logger}
}

func newHTTPClient() *http.Client {
	return &http.Client{Timeout: 12 * time.Second, Transport: newTransport()}
}

func newTransport() *http.Transport {
	return &http.Transport{
		MaxIdleConns:        4,
		MaxIdleConnsPerHost: 4,
		IdleConnTimeout:     90 * time.Second,
	}
}

func (c *Client) LookupFirstArrival(
	ctx context.Context,
	roomID int,
	from string,
	until string,
	targets map[string]struct{},
) (string, error) {
	body, err := marshalBody(roomID, from, until)
	if err != nil {
		return "", err
	}
	return c.lookup(ctx, body, targets, roomID)
}

func marshalBody(roomID int, from string, until string) ([]byte, error) {
	body := requestBody{AccommodationsToBook: []occupancy{defaultOccupancy()}, From: from, Until: until, ObjectIDs: []int{roomID}}
	return json.Marshal(body)
}

func defaultOccupancy() occupancy {
	return occupancy{NumberOfAdults: 1}
}

func (c *Client) lookup(
	ctx context.Context,
	body []byte,
	targets map[string]struct{},
	roomID int,
) (string, error) {
	request, err := c.newRequest(ctx, body)
	if err != nil {
		return "", err
	}
	return c.doLookup(request, targets, roomID)
}

func (c *Client) newRequest(ctx context.Context, body []byte) (*http.Request, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	setHeaders(request, c.apiKey)
	return request, nil
}

func setHeaders(request *http.Request, apiKey string) {
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Accept", "application/json")
	request.Header.Set("Origin", fallbackOrigin)
	request.Header.Set("Referer", fallbackOrigin+"/")
	request.Header.Set("User-Agent", userAgent)
	request.Header.Set("X-Client-Name", clientName)
	request.Header.Set("x-apikey", apiKey)
}

func (c *Client) doLookup(
	request *http.Request,
	targets map[string]struct{},
	roomID int,
) (string, error) {
	start := time.Now()
	response, err := c.client.Do(request)
	elapsed := time.Since(start)
	c.logHTTP(roomID, response, err, elapsed)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()
	return decodeArrival(response, targets)
}

func (c *Client) logHTTP(roomID int, response *http.Response, err error, elapsed time.Duration) {
	if c.logger == nil {
		return
	}
	if err != nil {
		c.logger.Debug("bookzo http", "objectId", roomID, "elapsed", elapsed, "err", err)
		return
	}
	c.logger.Debug("bookzo http", "objectId", roomID, "status", response.StatusCode, "elapsed", elapsed)
}

func decodeArrival(
	response *http.Response,
	targets map[string]struct{},
) (string, error) {
	if response.StatusCode != http.StatusOK {
		return "", fmt.Errorf("bookzo returned %s", response.Status)
	}
	var payload possibleStaysResponse
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		return "", err
	}
	return firstMatchingArrival(payload.Result, targets), nil
}

func firstMatchingArrival(rows []possibleStay, targets map[string]struct{}) string {
	first := ""
	for _, row := range rows {
		first = earlierDate(first, arrivalDate(row.Arrival, targets))
	}
	return first
}

func arrivalDate(value string, targets map[string]struct{}) string {
	if len(value) < 10 {
		return ""
	}
	day := value[:10]
	if _, ok := targets[day]; !ok {
		return ""
	}
	return day
}

func earlierDate(current string, candidate string) string {
	if candidate == "" || current != "" && current < candidate {
		return current
	}
	return candidate
}
