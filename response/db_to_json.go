package trans

// Content : Comic + AdLink
type Content struct {
	Site         string   `json:"site"` //just used to generate CID (hash of site + title)
	Title        string   `json:"title"`
	CID          string   `json:"cid"`
	Introduction string   `json:"introduction"`
	Link         string   `json:"link"`
	Latest       string   `json:"latest"` //ex: story1-1
	LastUpdate   string   `json:"lastUpdate"`
	NextUpdate   string   `json:"nextUpdate,omitempty"`
	LatestRecord string   `json:"latestRecord"`
	Authors      []string `json:"author"`
	Publisher    string   `json:"publisher"`
	Genres       []string `json:"genre"`
	AdLink       []string `json:"adLink"`
}
