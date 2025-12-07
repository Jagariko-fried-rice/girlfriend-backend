package external

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

// StableDiffusionClient はローカルのSD APIと通信するクライアントです
type StableDiffusionClient struct {
	APIURL        string
	OutputDir     string
	HttpClient    *http.Client
	StorageClient *SupabaseStorageClient // 追加: ストレージクライアント
}

// リクエストのJSON構造（txt2img用）
type sdRequest struct {
	Prompt         string  `json:"prompt"`
	NegativePrompt string  `json:"negative_prompt"`
	Steps          int     `json:"steps"`
	SamplerName    string  `json:"sampler_name"`
	CfgScale       float64 `json:"cfg_scale"`
	Width          int     `json:"width"`
	Height         int     `json:"height"`
}

// レスポンスのJSON構造
type sdResponse struct {
	Images []string `json:"images"` // Base64エンコードされた画像の配列
}

// ネガティブプロンプト（共通設定として定数化）
const defaultNegativePrompt = "(worst quality:1.4), (low quality:1.4), bad anatomy, extra limbs, mutation, text, watermark, logo, too sharp, monochrome, harsh shadows, bright light, (exposed skin:1.5), (cleavage:1.5), (breasts:1.5), nipples, (provocative:1.5), sexy, erotic, seductive, alluring, lewd, (swimsuit), (lingerie), (wet clothes), (transparent), (see-through), (open mouth), (tongue), heavy makeup, dirty, messy room, (adult content), (close-up), (child), (teenager), (school uniform), (student), (middle-aged), (wrinkles)"

func NewStableDiffusionClient(apiURL string, storageClient *SupabaseStorageClient) *StableDiffusionClient {
	// 保存先フォルダを作成
	outputDir := "output_images"
	if _, err := os.Stat(outputDir); os.IsNotExist(err) {
		os.Mkdir(outputDir, 0755)
	}

	return &StableDiffusionClient{
		APIURL:        apiURL,
		OutputDir:     outputDir,
		StorageClient: storageClient,
		HttpClient: &http.Client{
			Timeout: 5 * time.Minute, // 生成には時間がかかるので長めに設定
		},
	}
}

// GenerateImage: プロンプトを受け取り、画像を生成・保存して、そのパスを返します
func (c *StableDiffusionClient) GenerateImage(ctx context.Context, prompt string) (string, error) {
	// 1. リクエストボディの作成
	reqBody := sdRequest{
		Prompt:         prompt,
		NegativePrompt: defaultNegativePrompt,
		Steps:          20,
		SamplerName:    "Euler a",
		CfgScale:       7.0,
		Width:          512,
		Height:         512,
	}
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	// 2. APIリクエスト送信
	req, err := http.NewRequestWithContext(ctx, "POST", c.APIURL+"/sdapi/v1/txt2img", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	fmt.Printf("Stable Diffusionにリクエスト送信中... (Prompt: %s...)\n", prompt[:20])
	resp, err := c.HttpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to call SD API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("SD API returned error: %s", string(body))
	}

	// 3. レスポンスの解析 (Base64画像の取得)
	var sdResp sdResponse
	if err := json.NewDecoder(resp.Body).Decode(&sdResp); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}
	if len(sdResp.Images) == 0 {
		return "", fmt.Errorf("no images returned from SD API")
	}

	// 4. 画像をファイルとして保存
	// Base64デコード
	imgData, err := base64.StdEncoding.DecodeString(sdResp.Images[0])
	if err != nil {
		return "", fmt.Errorf("failed to decode base64 image: %w", err)
	}

	// ファイル名生成 (タイムスタンプ)
	fileName := fmt.Sprintf("generated_%d.png", time.Now().UnixNano())

	// Supabase Storageが設定されている場合はアップロード
	if c.StorageClient != nil {
		publicURL, err := c.StorageClient.UploadImage(imgData, fileName)
		if err != nil {
			return "", fmt.Errorf("failed to upload to supabase: %w", err)
		}
		fmt.Printf("Supabaseアップロード完了: %s\n", publicURL)
		return publicURL, nil
	}

	// 設定されていない場合はローカル保存 (開発用)
	filePath := filepath.Join(c.OutputDir, fileName)

	// 書き込み
	if err := os.WriteFile(filePath, imgData, 0644); err != nil {
		return "", fmt.Errorf("failed to save image file: %w", err)
	}

	fmt.Printf("画像保存完了: %s\n", filePath)

	// ローカル開発用として、APIサーバーからアクセスできるパスを返します
	return "/images/" + fileName, nil
}
