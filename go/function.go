package billing

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"time"

	"cloud.google.com/go/functions/metadata"
	"cloud.google.com/go/storage"
)

// GCSEvent is the payload of a GCS event. Please refer to the docs for
// additional information regarding GCS events.
type GCSEvent struct {
	Bucket         string    `json:"bucket"`
	Name           string    `json:"name"`
	Metageneration string    `json:"metageneration"`
	ResourceState  string    `json:"resourceState"`
	TimeCreated    time.Time `json:"timeCreated"`
	Updated        time.Time `json:"updated"`
}

// Billing represents GCP billing information
type Billing struct {
	AccountID     string `json:"accountId"`
	LineItemID    string `json:"lineItemId"`
	Description   string `json:"description"`
	StartTime     string `json:"startTime"`
	EndTime       string `json:"endTime"`
	ProjectNumber string `json:"projectNumber,omitempty"`
	ProjectID     string `json:"projectId,omitempty"`
	ProjectName   string `json:"projectName,omitempty"`
	Measurements  []struct {
		MeasurementID string `json:"measurementId"`
		Sum           string `json:"sum"`
		Unit          string `json:"unit"`
	} `json:"measurements"`
	Cost struct {
		Amount   string `json:"amount"`
		Currency string `json:"currency"`
	} `json:"cost"`
}

var (
	storageClient *storage.Client
	webhookURL    string
	regexB        *regexp.Regexp
)

const botName = "gcp-billing-bot"

func init() {
	var err error
	storageClient, err = storage.NewClient(context.Background())
	if err != nil {
		log.Fatalf("storage.NewClient: %v", err)
	}

	webhookURL = os.Getenv("WEBHOOK")
	regexB = regexp.MustCompile(`billing-(.*).json`)
}

// F sends biling information.
func F(ctx context.Context, e GCSEvent) error {
	meta, err := metadata.FromContext(ctx)
	if err != nil {
		return fmt.Errorf("metadata.FromContext: %v", err)
	}
	log.Printf("Event ID: %v\n", meta.EventID)
	log.Printf("Event type: %v\n", meta.EventType)
	log.Printf("Bucket: %v\n", e.Bucket)
	log.Printf("File: %v\n", e.Name)
	log.Printf("Metageneration: %v\n", e.Metageneration)
	log.Printf("Created: %v\n", e.TimeCreated)
	log.Printf("Updated: %v\n", e.Updated)

	rc, err := readFromGCS(ctx, e.Bucket, e.Name)
	if err != nil {
		return fmt.Errorf("readFromGCS: %v", err)
	}
	defer rc.Close()

	var b []Billing
	if err := json.NewDecoder(rc).Decode(&b); err != nil {
		return fmt.Errorf("json.NewDecoder: %v", err)
	}

	if len(b) == 0 {
		log.Println("billing information is empty")
		return nil
	}

	return webhook(webhookURL, buildMessage(e.Name, b))
}

func readFromGCS(ctx context.Context, bucket, name string) (io.ReadCloser, error) {
	obj := storageClient.Bucket(bucket).Object(name)
	return obj.NewReader(ctx)
}

func buildMessage(f string, bs []Billing) *Message {
	var fields []Field
	for _, b := range bs {
		f := Field{
			Title: fmt.Sprintf("%s: %s", b.ProjectID, b.Description),
			Value: fmt.Sprintf("%sドル（USD）", b.Cost.Amount),
		}
		fields = append(fields, f)
	}

	return &Message{
		Username: botName,
		Pretext:  fmt.Sprintf("%sの請求書", extractDate(f)),
		Color:    "#36a64f",
		Fields:   fields,
	}
}

func extractDate(f string) string {
	return regexB.FindAllStringSubmatch(f, -1)[0][1]
}

type Field struct {
	Title string `json:"title"`
	Value string `json:"value"`
}

type Message struct {
	Pretext  string  `json:"pretext"`
	Username string  `json:"username"`
	Color    string  `json:"color"`
	Fields   []Field `json:"fields"`
}

func webhook(url string, msg *Message) error {
	buf, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(
		"POST",
		url,
		bytes.NewBuffer(buf),
	)

	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Printf("resp.Body.Close(): %v", err)
		}
	}()
	return err
}
