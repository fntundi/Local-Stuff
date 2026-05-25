package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"sentinel-noc/internal/repository"
	"sentinel-noc/internal/services"

	"github.com/gin-gonic/gin"
)

// StreamingHandler manages RTSP→HLS stream publishing to MediaMTX
type StreamingHandler struct {
	cameraRepo    *repository.CameraRepository
	cryptoService *services.CryptoService
	mediamtxURL   string
	httpClient    *http.Client
}

// NewStreamingHandler creates a new StreamingHandler
func NewStreamingHandler(
	cameraRepo *repository.CameraRepository,
	cryptoService *services.CryptoService,
	mediamtxURL string,
) *StreamingHandler {
	return &StreamingHandler{
		cameraRepo:    cameraRepo,
		cryptoService: cryptoService,
		mediamtxURL:   mediamtxURL,
		httpClient:    &http.Client{Timeout: 10 * time.Second},
	}
}

// mediamtxPathConfig is the MediaMTX path configuration for a camera stream
type mediamtxPathConfig struct {
	Source                     string `json:"source"`
	SourceOnDemand             bool   `json:"sourceOnDemand"`
	SourceOnDemandStartTimeout string `json:"sourceOnDemandStartTimeout"`
	SourceOnDemandCloseAfter   string `json:"sourceOnDemandCloseAfter"`
}

// StartStream starts HLS streaming for a camera via MediaMTX
func (h *StreamingHandler) StartStream(c *gin.Context) {
	cameraID := c.Param("id")

	camera, err := h.cameraRepo.FindByID(c.Request.Context(), cameraID)
	if err != nil || camera == nil {
		c.JSON(http.StatusNotFound, gin.H{"detail": "Camera not found"})
		return
	}

	username, err := h.cryptoService.Decrypt(camera.UsernameEncrypted)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "Failed to decrypt credentials"})
		return
	}
	password, err := h.cryptoService.Decrypt(camera.PasswordEncrypted)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "Failed to decrypt credentials"})
		return
	}

	rtspURL := fmt.Sprintf("rtsp://%s:%s@%s:%d%s",
		username, password, camera.IPAddress, camera.RTSPPort, camera.RTSPPath)

	pathName := fmt.Sprintf("cam-%s", cameraID)
	config := mediamtxPathConfig{
		Source:                     rtspURL,
		SourceOnDemand:             true,
		SourceOnDemandStartTimeout: "10s",
		SourceOnDemandCloseAfter:   "10s",
	}

	if err := h.configureMTXPath(pathName, config); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": fmt.Sprintf("Failed to configure stream: %v", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"camera_id":   cameraID,
		"stream_path": pathName,
		"hls_url":     fmt.Sprintf("/hls/%s/index.m3u8", pathName),
		"rtsp_url":    fmt.Sprintf("rtsp://localhost:8554/%s", pathName),
	})
}

// StopStream stops HLS streaming for a camera
func (h *StreamingHandler) StopStream(c *gin.Context) {
	cameraID := c.Param("id")
	pathName := fmt.Sprintf("cam-%s", cameraID)

	if err := h.deleteMTXPath(pathName); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": fmt.Sprintf("Failed to stop stream: %v", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Stream stopped"})
}

// GetStreamStatus returns the status of a camera's HLS stream
func (h *StreamingHandler) GetStreamStatus(c *gin.Context) {
	cameraID := c.Param("id")
	pathName := fmt.Sprintf("cam-%s", cameraID)

	status, err := h.getMTXPathStatus(pathName)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"camera_id": cameraID,
			"active":    false,
			"hls_url":   fmt.Sprintf("/hls/%s/index.m3u8", pathName),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"camera_id": cameraID,
		"active":    status,
		"hls_url":   fmt.Sprintf("/hls/%s/index.m3u8", pathName),
		"rtsp_url":  fmt.Sprintf("rtsp://localhost:8554/%s", pathName),
	})
}

func (h *StreamingHandler) configureMTXPath(pathName string, config mediamtxPathConfig) error {
	body, _ := json.Marshal(config)
	url := fmt.Sprintf("%s/v3/config/paths/add/%s", h.mediamtxURL, pathName)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := h.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("mediamtx not reachable: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusConflict {
		return h.patchMTXPath(pathName, config)
	}

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("mediamtx error %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

func (h *StreamingHandler) patchMTXPath(pathName string, config mediamtxPathConfig) error {
	body, _ := json.Marshal(config)
	url := fmt.Sprintf("%s/v3/config/paths/patch/%s", h.mediamtxURL, pathName)

	req, err := http.NewRequest("PATCH", url, bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := h.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

func (h *StreamingHandler) deleteMTXPath(pathName string) error {
	url := fmt.Sprintf("%s/v3/config/paths/delete/%s", h.mediamtxURL, pathName)
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return err
	}
	resp, err := h.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

func (h *StreamingHandler) getMTXPathStatus(pathName string) (bool, error) {
	url := fmt.Sprintf("%s/v3/paths/get/%s", h.mediamtxURL, pathName)
	resp, err := h.httpClient.Get(url)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == 404 {
		return false, nil
	}
	return resp.StatusCode == 200, nil
}
