// Package services provides the ONVIF integration service
package services

import (
	"bytes"
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"sentinel-noc/internal/models"
)

// ONVIFService handles ONVIF protocol operations
type ONVIFService struct {
	httpClient *http.Client
}

// NewONVIFService creates a new ONVIF service
func NewONVIFService() *ONVIFService {
	return &ONVIFService{
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// ONVIFEnvelope represents a SOAP envelope
type ONVIFEnvelope struct {
	XMLName xml.Name    `xml:"Envelope"`
	Body    ONVIFBody   `xml:"Body"`
}

type ONVIFBody struct {
	Content []byte `xml:",innerxml"`
}

// GetCapabilitiesResponse represents ONVIF GetCapabilities response
type GetCapabilitiesResponse struct {
	Capabilities struct {
		Device struct {
			XAddr string `xml:"XAddr"`
		} `xml:"Device"`
		Events struct {
			XAddr string `xml:"XAddr"`
		} `xml:"Events"`
		PTZ struct {
			XAddr string `xml:"XAddr"`
		} `xml:"PTZ"`
		Media struct {
			XAddr string `xml:"XAddr"`
		} `xml:"Media"`
	} `xml:"Capabilities"`
}

// DeviceInfoResponse represents ONVIF GetDeviceInformation response
type DeviceInfoResponse struct {
	Manufacturer    string `xml:"Manufacturer"`
	Model           string `xml:"Model"`
	FirmwareVersion string `xml:"FirmwareVersion"`
	SerialNumber    string `xml:"SerialNumber"`
	HardwareId      string `xml:"HardwareId"`
}

// RelayOutputsResponse represents ONVIF GetRelayOutputs response
type RelayOutputsResponse struct {
	RelayOutputs []struct {
		Token      string `xml:"token,attr"`
		Properties struct {
			Mode      string `xml:"Mode"`
			IdleState string `xml:"IdleState"`
		} `xml:"Properties"`
	} `xml:"RelayOutputs"`
}

// DetectCapabilities probes a camera for ONVIF capabilities
func (s *ONVIFService) DetectCapabilities(ctx context.Context, ipAddress string, port int, username, password string) (*models.ONVIFCapabilities, error) {
	caps := &models.ONVIFCapabilities{
		Supported: false,
	}

	// Build ONVIF URL
	onvifURL := fmt.Sprintf("http://%s:%d/onvif/device_service", ipAddress, port)

	// Try GetCapabilities first
	getCapabilitiesXML := `<?xml version="1.0" encoding="UTF-8"?>
<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope">
  <s:Body xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xmlns:xsd="http://www.w3.org/2001/XMLSchema">
    <GetCapabilities xmlns="http://www.onvif.org/ver10/device/wsdl">
      <Category>All</Category>
    </GetCapabilities>
  </s:Body>
</s:Envelope>`

	resp, err := s.sendONVIFRequest(ctx, onvifURL, getCapabilitiesXML, username, password)
	if err != nil {
		return caps, fmt.Errorf("ONVIF not available: %w", err)
	}

	caps.Supported = true

	// Parse capabilities response
	if strings.Contains(string(resp), "PTZ") {
		caps.HasPTZ = true
	}
	if strings.Contains(string(resp), "Analytics") {
		caps.HasAnalytics = true
	}

	// Try to get relay outputs (alarm capability)
	relayOutputsXML := `<?xml version="1.0" encoding="UTF-8"?>
<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope">
  <s:Body xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xmlns:xsd="http://www.w3.org/2001/XMLSchema">
    <GetRelayOutputs xmlns="http://www.onvif.org/ver10/device/wsdl"/>
  </s:Body>
</s:Envelope>`

	relayResp, err := s.sendONVIFRequest(ctx, onvifURL, relayOutputsXML, username, password)
	if err == nil {
		// Count relay outputs
		relayCount := strings.Count(string(relayResp), "RelayOutput")
		if relayCount > 0 {
			caps.HasRelayOutputs = true
			caps.RelayCount = relayCount
		}
	}

	// Try to get audio outputs
	audioOutputsXML := `<?xml version="1.0" encoding="UTF-8"?>
<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope">
  <s:Body xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xmlns:xsd="http://www.w3.org/2001/XMLSchema">
    <GetAudioOutputs xmlns="http://www.onvif.org/ver10/media/wsdl"/>
  </s:Body>
</s:Envelope>`

	audioResp, err := s.sendONVIFRequest(ctx, onvifURL, audioOutputsXML, username, password)
	if err == nil && strings.Contains(string(audioResp), "AudioOutput") {
		caps.HasAudioOutputs = true
	}

	// Get profiles count
	profilesXML := `<?xml version="1.0" encoding="UTF-8"?>
<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope">
  <s:Body xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xmlns:xsd="http://www.w3.org/2001/XMLSchema">
    <GetProfiles xmlns="http://www.onvif.org/ver10/media/wsdl"/>
  </s:Body>
</s:Envelope>`

	profilesResp, err := s.sendONVIFRequest(ctx, onvifURL, profilesXML, username, password)
	if err == nil {
		caps.ProfilesCount = strings.Count(string(profilesResp), "Profiles")
	}

	return caps, nil
}

// TriggerAlarm activates a camera's relay output (alarm)
func (s *ONVIFService) TriggerAlarm(ctx context.Context, ipAddress string, port int, username, password string, relayToken string, durationSecs int) error {
	onvifURL := fmt.Sprintf("http://%s:%d/onvif/device_service", ipAddress, port)

	// If no relay token specified, try to get the first one
	if relayToken == "" {
		relayToken = "relay1" // Common default
	}

	// Set relay output state to active (triggers alarm)
	setRelayXML := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope">
  <s:Body xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xmlns:xsd="http://www.w3.org/2001/XMLSchema">
    <SetRelayOutputState xmlns="http://www.onvif.org/ver10/device/wsdl">
      <RelayOutputToken>%s</RelayOutputToken>
      <LogicalState>active</LogicalState>
    </SetRelayOutputState>
  </s:Body>
</s:Envelope>`, relayToken)

	_, err := s.sendONVIFRequest(ctx, onvifURL, setRelayXML, username, password)
	if err != nil {
		return fmt.Errorf("failed to trigger alarm: %w", err)
	}

	// If duration specified, schedule deactivation
	if durationSecs > 0 {
		go func() {
			time.Sleep(time.Duration(durationSecs) * time.Second)
			_ = s.DeactivateAlarm(context.Background(), ipAddress, port, username, password, relayToken)
		}()
	}

	return nil
}

// DeactivateAlarm deactivates a camera's relay output
func (s *ONVIFService) DeactivateAlarm(ctx context.Context, ipAddress string, port int, username, password string, relayToken string) error {
	onvifURL := fmt.Sprintf("http://%s:%d/onvif/device_service", ipAddress, port)

	if relayToken == "" {
		relayToken = "relay1"
	}

	setRelayXML := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope">
  <s:Body xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xmlns:xsd="http://www.w3.org/2001/XMLSchema">
    <SetRelayOutputState xmlns="http://www.onvif.org/ver10/device/wsdl">
      <RelayOutputToken>%s</RelayOutputToken>
      <LogicalState>inactive</LogicalState>
    </SetRelayOutputState>
  </s:Body>
</s:Envelope>`, relayToken)

	_, err := s.sendONVIFRequest(ctx, onvifURL, setRelayXML, username, password)
	return err
}

// TestConnection tests ONVIF connectivity
func (s *ONVIFService) TestConnection(ctx context.Context, ipAddress string, port int, username, password string) (bool, string, error) {
	onvifURL := fmt.Sprintf("http://%s:%d/onvif/device_service", ipAddress, port)

	// Try GetDeviceInformation
	getDeviceInfoXML := `<?xml version="1.0" encoding="UTF-8"?>
<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope">
  <s:Body xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xmlns:xsd="http://www.w3.org/2001/XMLSchema">
    <GetDeviceInformation xmlns="http://www.onvif.org/ver10/device/wsdl"/>
  </s:Body>
</s:Envelope>`

	resp, err := s.sendONVIFRequest(ctx, onvifURL, getDeviceInfoXML, username, password)
	if err != nil {
		return false, "", err
	}

	// Extract manufacturer/model info
	info := "ONVIF device"
	if strings.Contains(string(resp), "Manufacturer") {
		// Simple extraction
		start := strings.Index(string(resp), "<tds:Manufacturer>")
		if start == -1 {
			start = strings.Index(string(resp), "<Manufacturer>")
		}
		if start != -1 {
			end := strings.Index(string(resp)[start:], "</")
			if end != -1 {
				info = string(resp)[start:start+end]
				info = strings.TrimPrefix(info, "<tds:Manufacturer>")
				info = strings.TrimPrefix(info, "<Manufacturer>")
			}
		}
	}

	return true, info, nil
}

// sendONVIFRequest sends a SOAP request to an ONVIF device
func (s *ONVIFService) sendONVIFRequest(ctx context.Context, url, soapXML, username, password string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBufferString(soapXML))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/soap+xml; charset=utf-8")
	
	// Add basic auth if credentials provided
	if username != "" && password != "" {
		req.SetBasicAuth(username, password)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Check for SOAP faults
	if strings.Contains(string(body), "Fault") {
		return nil, fmt.Errorf("ONVIF error: %s", extractFaultMessage(body))
	}

	return body, nil
}

// extractFaultMessage extracts error message from SOAP fault
func extractFaultMessage(body []byte) string {
	bodyStr := string(body)
	
	// Try to find fault reason
	if idx := strings.Index(bodyStr, "<env:Reason>"); idx != -1 {
		end := strings.Index(bodyStr[idx:], "</env:Reason>")
		if end != -1 {
			return bodyStr[idx : idx+end]
		}
	}
	
	if idx := strings.Index(bodyStr, "<faultstring>"); idx != -1 {
		end := strings.Index(bodyStr[idx:], "</faultstring>")
		if end != -1 {
			return bodyStr[idx+13 : idx+end]
		}
	}
	
	return "unknown ONVIF error"
}
