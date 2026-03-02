package buildsSitesClient

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/ByChanderZap/exile-tracker/utils"
)

// SiteInfo holds info for each build site
type SiteInfo struct {
	Label      string
	ID         string
	CodeOut    string
	PostURL    string
	PostFields string
	LinkURL    string
}

// SitesUrl provides static references to each supported site
var SitesUrl = struct {
	POBBin   SiteInfo
	PoeNinja SiteInfo
	Poedb    SiteInfo
}{
	POBBin: SiteInfo{
		Label:      "pobb.in",
		ID:         "POBBin",
		CodeOut:    "https://pobb.in/",
		PostURL:    "https://pobb.in/pob/",
		PostFields: "",
		LinkURL:    "pobb.in/",
	},
	PoeNinja: SiteInfo{
		Label:      "PoeNinja",
		ID:         "PoeNinja",
		CodeOut:    "",
		PostURL:    "https://poe.ninja/pob/api/api_post.php",
		PostFields: "api_paste_code=",
		LinkURL:    "poe.ninja/pob/",
	},
	Poedb: SiteInfo{
		Label:      "poedb.tw",
		ID:         "PoEDB",
		CodeOut:    "",
		PostURL:    "https://poedb.tw/pob/api/gen",
		PostFields: "",
		LinkURL:    "poedb.tw/pob/",
	},
}

var FallbackOrder = []SiteInfo{
	SitesUrl.PoeNinja,
	SitesUrl.POBBin,
	SitesUrl.Poedb,
}

var httpClient = &http.Client{
	Timeout: 10 * time.Second,
}

func UploadBuildWithFallback(buildCode string) (string, error) {
	log := utils.ChildLogger("builds-sites-client")
	var errs []string
	for _, site := range FallbackOrder {
		result, err := UploadBuild(buildCode, site)
		if err == nil {
			return result, nil
		}
		log.Warn().Err(err).Str("site", site.Label).Msg("Upload failed, trying next site")
		errs = append(errs, fmt.Sprintf("%s: %s", site.Label, err.Error()))
	}
	return "", fmt.Errorf("all upload sites failed: %s", strings.Join(errs, "; "))
}

func UploadBuild(buildCode string, site SiteInfo) (string, error) {
	if site.PostURL == "" {
		return "", fmt.Errorf("no post URL for site %s", site.Label)
	}
	postBody := site.PostFields + buildCode

	req, err := http.NewRequest("POST", site.PostURL, bytes.NewBufferString(postBody))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", "exile-tracker/0.0.1 (contact: neryt.alexander@gmail.com)")

	resp, err := httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode == 200 {
		return fmt.Sprintf("%s%s", site.CodeOut, string(body)), nil
	}

	return "", fmt.Errorf("upload failed: %s", string(body))
}
