package apiservice

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type Request interface {
	Do(ctx context.Context, output any) error
}

type request struct {
	Method  Method
	Url     string
	Payload any
	clt     *http.Client
}

func (c *Client) GetRequest(method string, query string, payload any) Request {
	return &request{
		Method:  Method(method),
		Url:     c.address + query,
		Payload: payload,
		clt:     c.Client,
	}
}

func (r *request) Do(ctx context.Context, output any) error {
	body := new(bytes.Buffer)
	if r.Payload != nil {
		post, err := json.Marshal(r.Payload)
		if err != nil {
			return fmt.Errorf("marshal: %w", err)
		}
		body = bytes.NewBuffer(post)
	}

	ctx, cancel := context.WithTimeout(ctx, r.clt.Timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, r.Method.String(), r.Url, body)
	if err != nil {
		return fmt.Errorf("request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := r.clt.Do(req)
	if err != nil {
		return fmt.Errorf("do: %w", err)
	}
	defer resp.Body.Close()

	var response Response

	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return fmt.Errorf("decode: %w, (%+v)", err, resp)
	}

	if !response.Success {
		return fmt.Errorf("%s", response.Error)
	}

	err = json.Unmarshal([]byte(response.Return), &output)
	if err != nil {
		return fmt.Errorf("unmarshal: %w", err)
	}

	return nil
}
