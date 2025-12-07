package external

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"time"
)

type SupabaseStorageClient struct {
	ProjectURL string
	APIKey     string
	BucketName string
	HttpClient *http.Client
}

func NewSupabaseStorageClient(projectURL, apiKey, bucketName string) *SupabaseStorageClient {
	return &SupabaseStorageClient{
		ProjectURL: projectURL,
		APIKey:     apiKey,
		BucketName: bucketName,
		HttpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

// UploadImage: 画像データをSupabase Storageにアップロードし、公開URLを返す
func (c *SupabaseStorageClient) UploadImage(imageData []byte, fileName string) (string, error) {
	// エンドポイント: POST /storage/v1/object/{bucket}/{path}
	url := fmt.Sprintf("%s/storage/v1/object/%s/%s", c.ProjectURL, c.BucketName, fileName)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(imageData))
	if err != nil {
		return "", fmt.Errorf("failed to create upload request: %w", err)
	}

	// ヘッダー設定
	req.Header.Set("Authorization", "Bearer "+c.APIKey)
	req.Header.Set("Content-Type", "image/png") // PNGとして保存

	resp, err := c.HttpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to upload image: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 && resp.StatusCode != 201 {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("supabase upload failed (status %d): %s", resp.StatusCode, string(body))
	}

	// 公開URLを生成して返す
	// 形式: {ProjectURL}/storage/v1/object/public/{BucketName}/{FileName}
	publicURL := fmt.Sprintf("%s/storage/v1/object/public/%s/%s", c.ProjectURL, c.BucketName, fileName)
	return publicURL, nil
}
