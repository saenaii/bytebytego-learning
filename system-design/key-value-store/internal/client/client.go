package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/bytebytego-learning/key-value-store/internal/store"
)

type Client struct {
	baseURL string
	http    *http.Client
}

func New(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		http: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (c *Client) Get(key string) (*store.Entry, error) {
	resp, err := c.http.Get(c.baseURL + "/kv/" + key)
	if err != nil {
		return nil, fmt.Errorf("get request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(body))
	}

	var entry store.Entry
	if err := json.NewDecoder(resp.Body).Decode(&entry); err != nil {
		return nil, fmt.Errorf("decode response failed: %w", err)
	}
	return &entry, nil
}

func (c *Client) Put(key, value string) error {
	body := map[string]string{"value": value}
	data, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("marshal request failed: %w", err)
	}

	req, err := http.NewRequest(http.MethodPut, c.baseURL+"/kv/"+key, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("create request failed: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("put request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(body))
	}
	return nil
}

func (c *Client) Health() error {
	resp, err := c.http.Get(c.baseURL + "/health")
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("health check failed: %d", resp.StatusCode)
	}
	return nil
}

func (c *Client) AddNode(nodeID, nodeAddr string) error {
	body := map[string]string{"action": "add", "node_id": nodeID, "node_addr": nodeAddr}
	data, _ := json.Marshal(body)
	resp, err := c.http.Post(c.baseURL+"/admin/node", "application/json", bytes.NewReader(data))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("add node failed: %d %s", resp.StatusCode, string(body))
	}
	return nil
}

func (c *Client) RemoveNode(nodeID string) error {
	body := map[string]string{"action": "remove", "node_id": nodeID}
	data, _ := json.Marshal(body)
	req, err := http.NewRequest(http.MethodPost, c.baseURL+"/admin/node", bytes.NewReader(data))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("remove node failed: %d %s", resp.StatusCode, string(body))
	}
	return nil
}
